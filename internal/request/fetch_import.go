package request

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"

	"restdeck/internal/domain"
)

type jsProperty struct {
	Key   string
	Value string
}

func ImportFetch(raw, collectionID string, sortOrder int) (domain.Request, error) {
	argsRaw, err := extractFetchArguments(raw)
	if err != nil {
		return domain.Request{}, err
	}
	args := splitTopLevel(argsRaw, ',')
	if len(args) == 0 {
		return domain.Request{}, fmt.Errorf("expected a browser fetch(url, init) snippet")
	}

	requestURL, ok := parseJSStringLiteral(args[0])
	if !ok || strings.TrimSpace(requestURL) == "" {
		return domain.Request{}, fmt.Errorf("fetch import only supports a static string URL")
	}

	var props []jsProperty
	if len(args) > 1 && strings.TrimSpace(args[1]) != "" {
		props, err = parseObjectProperties(args[1])
		if err != nil {
			return domain.Request{}, fmt.Errorf("parse fetch init: %w", err)
		}
	}

	method := "GET"
	if rawMethod, found := getProperty(props, "method"); found {
		if parsed, ok := parseJSStringLiteral(rawMethod); ok && strings.TrimSpace(parsed) != "" {
			method = strings.ToUpper(strings.TrimSpace(parsed))
		}
	}

	headers := []domain.KeyValue{}
	if rawHeaders, found := getProperty(props, "headers"); found {
		headers = parseFetchHeaders(rawHeaders)
	}

	bodyMode := domain.BodyModeNone
	body := ""
	if rawBody, found := getProperty(props, "body"); found {
		body, ok = parseFetchBody(rawBody)
		if ok {
			bodyMode = detectFetchBodyMode(contentType(headers), body)
			if bodyMode == domain.BodyModeURLEncoded {
				body = urlEncodedBodyLines(body)
			}
		}
	}

	return domain.Request{
		ID:           uuid.NewString(),
		CollectionID: collectionID,
		Name:         fetchRequestName(method, requestURL),
		Method:       method,
		URL:          requestURL,
		Params:       []domain.KeyValue{},
		Headers:      headers,
		BodyMode:     bodyMode,
		Body:         body,
		Auth:         domain.AuthConfig{Type: domain.AuthTypeNone, Values: map[string]string{}},
		TimeoutMs:    30000,
		SortOrder:    sortOrder,
		UpdatedAt:    time.Now(),
	}, nil
}

func extractFetchArguments(raw string) (string, error) {
	open := findFetchOpenParen(raw)
	if open < 0 {
		return "", fmt.Errorf("expected a browser fetch(...) snippet")
	}
	close, err := findMatching(raw, open, '(', ')')
	if err != nil {
		return "", err
	}
	return raw[open+1 : close], nil
}

func findFetchOpenParen(raw string) int {
	for offset := 0; offset < len(raw); {
		idx := strings.Index(raw[offset:], "fetch")
		if idx < 0 {
			return -1
		}
		idx += offset
		beforeOK := idx == 0 || !isIdentifierChar(raw[idx-1])
		after := idx + len("fetch")
		afterOK := after >= len(raw) || !isIdentifierChar(raw[after])
		if beforeOK && afterOK {
			i := skipWhitespaceAndComments(raw, after)
			if i < len(raw) && raw[i] == '(' {
				return i
			}
		}
		offset = idx + len("fetch")
	}
	return -1
}

func parseFetchHeaders(raw string) []domain.KeyValue {
	raw = unwrapHeadersConstructor(raw)
	headers := []domain.KeyValue{}

	if strings.HasPrefix(strings.TrimSpace(raw), "{") {
		props, err := parseObjectProperties(raw)
		if err != nil {
			return headers
		}
		for _, prop := range props {
			value, ok := parsePrimitiveString(prop.Value)
			if !ok || strings.TrimSpace(prop.Key) == "" {
				continue
			}
			headers = append(headers, domain.KeyValue{
				ID:      uuid.NewString(),
				Enabled: true,
				Key:     prop.Key,
				Value:   value,
			})
		}
		return headers
	}

	for _, pair := range parseHeaderPairs(raw) {
		headers = append(headers, domain.KeyValue{
			ID:      uuid.NewString(),
			Enabled: true,
			Key:     pair.Key,
			Value:   pair.Value,
		})
	}
	return headers
}

func unwrapHeadersConstructor(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "new Headers") && !strings.HasPrefix(trimmed, "Headers") {
		return raw
	}
	open := strings.IndexByte(trimmed, '(')
	if open < 0 {
		return raw
	}
	close, err := findMatching(trimmed, open, '(', ')')
	if err != nil {
		return raw
	}
	args := splitTopLevel(trimmed[open+1:close], ',')
	if len(args) == 0 {
		return raw
	}
	return args[0]
}

func parseHeaderPairs(raw string) []jsProperty {
	inner, ok := arrayInner(raw)
	if !ok {
		return nil
	}
	pairs := []jsProperty{}
	for _, item := range splitTopLevel(inner, ',') {
		itemInner, ok := arrayInner(item)
		if !ok {
			continue
		}
		parts := splitTopLevel(itemInner, ',')
		if len(parts) < 2 {
			continue
		}
		key, keyOK := parseJSStringLiteral(parts[0])
		value, valueOK := parsePrimitiveString(parts[1])
		if keyOK && valueOK && strings.TrimSpace(key) != "" {
			pairs = append(pairs, jsProperty{Key: key, Value: value})
		}
	}
	return pairs
}

func parseFetchBody(raw string) (string, bool) {
	if parsed, ok := parseJSStringLiteral(raw); ok {
		return parsed, true
	}

	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "JSON.stringify") {
		open := strings.IndexByte(trimmed, '(')
		if open < 0 {
			return "", false
		}
		close, err := findMatching(trimmed, open, '(', ')')
		if err != nil {
			return "", false
		}
		args := splitTopLevel(trimmed[open+1:close], ',')
		if len(args) == 0 {
			return "", false
		}
		value := strings.TrimSpace(args[0])
		if json.Valid([]byte(value)) {
			return value, true
		}
		if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") {
			return value, true
		}
	}

	return "", false
}

func detectFetchBodyMode(ct, body string) domain.BodyMode {
	trimmed := strings.TrimSpace(body)
	lowerCT := strings.ToLower(ct)
	if strings.Contains(lowerCT, "application/x-www-form-urlencoded") {
		return domain.BodyModeURLEncoded
	}
	if strings.Contains(lowerCT, "application/json") || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return domain.BodyModeJSON
	}
	if trimmed == "" {
		return domain.BodyModeNone
	}
	return domain.BodyModeRaw
}

func contentType(headers []domain.KeyValue) string {
	for _, header := range headers {
		if strings.EqualFold(header.Key, "content-type") {
			return header.Value
		}
	}
	return ""
}

func urlEncodedBodyLines(body string) string {
	if strings.TrimSpace(body) == "" {
		return ""
	}
	lines := []string{}
	for _, part := range strings.Split(body, "&") {
		if part == "" {
			continue
		}
		key, value, _ := strings.Cut(part, "=")
		decodedKey, err := url.QueryUnescape(key)
		if err == nil {
			key = decodedKey
		}
		decodedValue, err := url.QueryUnescape(value)
		if err == nil {
			value = decodedValue
		}
		lines = append(lines, key+"="+value)
	}
	return strings.Join(lines, "\n")
}

func fetchRequestName(method, rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err == nil && parsed.Host != "" {
		path := parsed.EscapedPath()
		if path == "" {
			path = "/"
		}
		return method + " " + parsed.Host + path
	}
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		trimmed = "request"
	}
	return method + " " + trimmed
}

func parseObjectProperties(raw string) ([]jsProperty, error) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "{") {
		return nil, fmt.Errorf("expected object literal")
	}
	close, err := findMatching(trimmed, 0, '{', '}')
	if err != nil {
		return nil, err
	}
	inner := trimmed[1:close]
	props := []jsProperty{}
	for i := 0; i < len(inner); {
		i = skipWhitespaceAndComments(inner, i)
		if i >= len(inner) {
			break
		}

		key := ""
		if isQuote(inner[i]) {
			var err error
			key, i, err = readJSString(inner, i)
			if err != nil {
				return nil, err
			}
		} else {
			start := i
			for i < len(inner) && inner[i] != ':' && inner[i] != ',' {
				i++
			}
			key = strings.TrimSpace(inner[start:i])
		}

		i = skipWhitespaceAndComments(inner, i)
		if i >= len(inner) || inner[i] != ':' {
			return nil, fmt.Errorf("expected ':' after object key %q", key)
		}
		i++
		i = skipWhitespaceAndComments(inner, i)
		start := i
		i = scanValueEnd(inner, i)
		props = append(props, jsProperty{Key: key, Value: strings.TrimSpace(inner[start:i])})

		i = skipWhitespaceAndComments(inner, i)
		if i < len(inner) && inner[i] == ',' {
			i++
		}
	}
	return props, nil
}

func getProperty(props []jsProperty, key string) (string, bool) {
	for _, prop := range props {
		if prop.Key == key {
			return prop.Value, true
		}
	}
	return "", false
}

func parsePrimitiveString(raw string) (string, bool) {
	if parsed, ok := parseJSStringLiteral(raw); ok {
		return parsed, true
	}
	trimmed := strings.TrimSpace(raw)
	lower := strings.ToLower(trimmed)
	if trimmed == "" || lower == "undefined" || lower == "null" {
		return "", false
	}
	if lower == "true" || lower == "false" {
		return lower, true
	}
	if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return trimmed, true
	}
	return "", false
}

func parseJSStringLiteral(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || !isQuote(trimmed[0]) {
		return "", false
	}
	value, next, err := readJSString(trimmed, 0)
	if err != nil {
		return "", false
	}
	rest := strings.TrimSpace(trimmed[next:])
	if rest != "" {
		return "", false
	}
	return value, true
}

func readJSString(raw string, start int) (string, int, error) {
	if start >= len(raw) || !isQuote(raw[start]) {
		return "", start, fmt.Errorf("expected string literal")
	}
	quote := raw[start]
	var b strings.Builder
	for i := start + 1; i < len(raw); i++ {
		c := raw[i]
		if quote == '`' && c == '$' && i+1 < len(raw) && raw[i+1] == '{' {
			return "", start, fmt.Errorf("template interpolation is not supported")
		}
		if c == quote {
			return b.String(), i + 1, nil
		}
		if c != '\\' {
			b.WriteByte(c)
			continue
		}
		if i+1 >= len(raw) {
			return "", start, fmt.Errorf("unfinished escape sequence")
		}
		i++
		escaped := raw[i]
		switch escaped {
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'v':
			b.WriteByte('\v')
		case '0':
			b.WriteByte(0)
		case '\n':
		case '\r':
			if i+1 < len(raw) && raw[i+1] == '\n' {
				i++
			}
		case 'x':
			if i+2 >= len(raw) {
				return "", start, fmt.Errorf("invalid hex escape")
			}
			value, err := strconv.ParseInt(raw[i+1:i+3], 16, 32)
			if err != nil {
				return "", start, fmt.Errorf("invalid hex escape")
			}
			b.WriteRune(rune(value))
			i += 2
		case 'u':
			value, next, err := readUnicodeEscape(raw, i+1)
			if err != nil {
				return "", start, err
			}
			b.WriteRune(value)
			i = next - 1
		default:
			b.WriteByte(escaped)
		}
	}
	return "", start, fmt.Errorf("unterminated string literal")
}

func readUnicodeEscape(raw string, start int) (rune, int, error) {
	if start < len(raw) && raw[start] == '{' {
		end := strings.IndexByte(raw[start+1:], '}')
		if end < 0 {
			return 0, start, fmt.Errorf("invalid unicode escape")
		}
		end += start + 1
		value, err := strconv.ParseInt(raw[start+1:end], 16, 32)
		if err != nil {
			return 0, start, fmt.Errorf("invalid unicode escape")
		}
		return rune(value), end + 1, nil
	}
	if start+4 > len(raw) {
		return 0, start, fmt.Errorf("invalid unicode escape")
	}
	value, err := strconv.ParseInt(raw[start:start+4], 16, 32)
	if err != nil {
		return 0, start, fmt.Errorf("invalid unicode escape")
	}
	return rune(value), start + 4, nil
}

func arrayInner(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "[") {
		return "", false
	}
	close, err := findMatching(trimmed, 0, '[', ']')
	if err != nil {
		return "", false
	}
	if strings.TrimSpace(trimmed[close+1:]) != "" {
		return "", false
	}
	return trimmed[1:close], true
}

func splitTopLevel(raw string, sep byte) []string {
	parts := []string{}
	start := 0
	depthParen := 0
	depthBrace := 0
	depthBracket := 0
	for i := 0; i < len(raw); i++ {
		if next, ok := skipComment(raw, i); ok {
			i = next - 1
			continue
		}
		if isQuote(raw[i]) {
			next, err := skipJSString(raw, i)
			if err != nil {
				break
			}
			i = next - 1
			continue
		}
		switch raw[i] {
		case '(':
			depthParen++
		case ')':
			if depthParen > 0 {
				depthParen--
			}
		case '{':
			depthBrace++
		case '}':
			if depthBrace > 0 {
				depthBrace--
			}
		case '[':
			depthBracket++
		case ']':
			if depthBracket > 0 {
				depthBracket--
			}
		case sep:
			if depthParen == 0 && depthBrace == 0 && depthBracket == 0 {
				parts = append(parts, strings.TrimSpace(raw[start:i]))
				start = i + 1
			}
		}
	}
	parts = append(parts, strings.TrimSpace(raw[start:]))
	return parts
}

func scanValueEnd(raw string, start int) int {
	depthParen := 0
	depthBrace := 0
	depthBracket := 0
	for i := start; i < len(raw); i++ {
		if next, ok := skipComment(raw, i); ok {
			i = next - 1
			continue
		}
		if isQuote(raw[i]) {
			next, err := skipJSString(raw, i)
			if err != nil {
				return len(raw)
			}
			i = next - 1
			continue
		}
		switch raw[i] {
		case '(':
			depthParen++
		case ')':
			if depthParen > 0 {
				depthParen--
			}
		case '{':
			depthBrace++
		case '}':
			if depthBrace > 0 {
				depthBrace--
			}
		case '[':
			depthBracket++
		case ']':
			if depthBracket > 0 {
				depthBracket--
			}
		case ',':
			if depthParen == 0 && depthBrace == 0 && depthBracket == 0 {
				return i
			}
		}
	}
	return len(raw)
}

func findMatching(raw string, open int, openCh, closeCh byte) (int, error) {
	if open >= len(raw) || raw[open] != openCh {
		return -1, fmt.Errorf("expected %q", openCh)
	}
	depth := 0
	for i := open; i < len(raw); i++ {
		if next, ok := skipComment(raw, i); ok {
			i = next - 1
			continue
		}
		if isQuote(raw[i]) {
			next, err := skipJSString(raw, i)
			if err != nil {
				return -1, err
			}
			i = next - 1
			continue
		}
		if raw[i] == openCh {
			depth++
			continue
		}
		if raw[i] == closeCh {
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return -1, fmt.Errorf("unterminated %q", openCh)
}

func skipJSString(raw string, start int) (int, error) {
	_, next, err := readJSString(raw, start)
	return next, err
}

func skipWhitespaceAndComments(raw string, start int) int {
	for start < len(raw) {
		for start < len(raw) && unicode.IsSpace(rune(raw[start])) {
			start++
		}
		if next, ok := skipComment(raw, start); ok {
			start = next
			continue
		}
		return start
	}
	return start
}

func skipComment(raw string, start int) (int, bool) {
	if start+1 >= len(raw) || raw[start] != '/' {
		return start, false
	}
	switch raw[start+1] {
	case '/':
		end := strings.IndexByte(raw[start+2:], '\n')
		if end < 0 {
			return len(raw), true
		}
		return start + 2 + end + 1, true
	case '*':
		end := strings.Index(raw[start+2:], "*/")
		if end < 0 {
			return len(raw), true
		}
		return start + 2 + end + 2, true
	default:
		return start, false
	}
}

func isIdentifierChar(c byte) bool {
	return c == '$' || c == '_' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
}

func isQuote(c byte) bool {
	return c == '\'' || c == '"' || c == '`'
}
