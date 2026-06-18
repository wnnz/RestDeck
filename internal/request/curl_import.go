package request

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"restdeck/internal/domain"
)

func ImportCurl(raw, collectionID string, sortOrder int) (domain.Request, error) {
	tokens, err := lexCurlCommand(raw)
	if err != nil {
		return domain.Request{}, err
	}
	if len(tokens) == 0 || !isCurlExecutable(tokens[0]) {
		return domain.Request{}, fmt.Errorf("expected a curl command")
	}

	method := ""
	requestURL := ""
	headers := []domain.KeyValue{}
	bodyParts := []string{}
	bodyModeHint := domain.BodyModeNone

	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		switch {
		case token == "--":
			continue
		case token == "-X" || token == "--request":
			value, ok := nextCurlValue(tokens, &i, token)
			if !ok {
				return domain.Request{}, fmt.Errorf("%s requires a method", token)
			}
			method = strings.ToUpper(strings.TrimSpace(value))
		case strings.HasPrefix(token, "-X") && len(token) > 2:
			method = strings.ToUpper(strings.TrimSpace(token[2:]))
		case strings.HasPrefix(token, "--request="):
			method = strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(token, "--request=")))
		case token == "-H" || token == "--header":
			value, ok := nextCurlValue(tokens, &i, token)
			if !ok {
				return domain.Request{}, fmt.Errorf("%s requires a header", token)
			}
			appendCurlHeader(&headers, value)
		case strings.HasPrefix(token, "-H") && len(token) > 2:
			appendCurlHeader(&headers, token[2:])
		case strings.HasPrefix(token, "--header="):
			appendCurlHeader(&headers, strings.TrimPrefix(token, "--header="))
		case token == "-b" || token == "--cookie":
			value, ok := nextCurlValue(tokens, &i, token)
			if !ok {
				return domain.Request{}, fmt.Errorf("%s requires a cookie value", token)
			}
			appendSyntheticHeader(&headers, "Cookie", value)
		case strings.HasPrefix(token, "--cookie="):
			appendSyntheticHeader(&headers, "Cookie", strings.TrimPrefix(token, "--cookie="))
		case token == "-A" || token == "--user-agent":
			value, ok := nextCurlValue(tokens, &i, token)
			if !ok {
				return domain.Request{}, fmt.Errorf("%s requires a user agent", token)
			}
			appendSyntheticHeader(&headers, "User-Agent", value)
		case strings.HasPrefix(token, "--user-agent="):
			appendSyntheticHeader(&headers, "User-Agent", strings.TrimPrefix(token, "--user-agent="))
		case token == "-e" || token == "--referer":
			value, ok := nextCurlValue(tokens, &i, token)
			if !ok {
				return domain.Request{}, fmt.Errorf("%s requires a referer", token)
			}
			appendSyntheticHeader(&headers, "Referer", value)
		case strings.HasPrefix(token, "--referer="):
			appendSyntheticHeader(&headers, "Referer", strings.TrimPrefix(token, "--referer="))
		case token == "--url":
			value, ok := nextCurlValue(tokens, &i, token)
			if !ok {
				return domain.Request{}, fmt.Errorf("--url requires a URL")
			}
			requestURL = value
		case strings.HasPrefix(token, "--url="):
			requestURL = strings.TrimPrefix(token, "--url=")
		case token == "-I" || token == "--head":
			method = "HEAD"
		case isCurlDataToken(token):
			value, ok := curlAttachedValue(token)
			if !ok {
				value, ok = nextCurlValue(tokens, &i, token)
			}
			if !ok {
				return domain.Request{}, fmt.Errorf("%s requires a body value", token)
			}
			bodyParts = append(bodyParts, value)
			if token == "-F" || strings.HasPrefix(token, "-F") || token == "--form" || strings.HasPrefix(token, "--form=") {
				bodyModeHint = domain.BodyModeForm
			} else if token == "--data-urlencode" || strings.HasPrefix(token, "--data-urlencode=") {
				bodyModeHint = domain.BodyModeURLEncoded
			}
		case strings.HasPrefix(token, "-"):
			if curlFlagConsumesValue(token) {
				i++
			}
		default:
			if requestURL == "" {
				requestURL = token
			}
		}
	}

	requestURL = strings.TrimSpace(requestURL)
	if requestURL == "" {
		return domain.Request{}, fmt.Errorf("curl import requires a URL")
	}
	if method == "" {
		if len(bodyParts) > 0 {
			method = "POST"
		} else {
			method = "GET"
		}
	}

	body := strings.Join(bodyParts, "&")
	bodyMode := domain.BodyModeNone
	if bodyModeHint == domain.BodyModeForm {
		bodyMode = domain.BodyModeForm
		body = strings.Join(bodyParts, "\n")
	} else if body != "" {
		bodyMode = detectFetchBodyMode(contentType(headers), body)
		if bodyModeHint == domain.BodyModeURLEncoded && bodyMode == domain.BodyModeRaw {
			bodyMode = domain.BodyModeURLEncoded
		}
		if bodyMode == domain.BodyModeURLEncoded {
			body = urlEncodedBodyLines(body)
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

func lexCurlCommand(raw string) ([]string, error) {
	raw = normalizeCurlContinuations(raw)
	tokens := []string{}
	var token strings.Builder
	inToken := false
	var quote byte

	flush := func() {
		if inToken {
			tokens = append(tokens, token.String())
			token.Reset()
			inToken = false
		}
	}

	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if quote != 0 {
			if c == quote {
				quote = 0
				inToken = true
				continue
			}
			if quote != '\'' && (c == '\\' || c == '^') && i+1 < len(raw) {
				i++
				token.WriteByte(raw[i])
				inToken = true
				continue
			}
			token.WriteByte(c)
			inToken = true
			continue
		}

		if isShellSpace(c) {
			flush()
			continue
		}
		if isQuote(c) {
			quote = c
			inToken = true
			continue
		}
		if (c == '\\' || c == '^') && i+1 < len(raw) {
			i++
			token.WriteByte(raw[i])
			inToken = true
			continue
		}
		token.WriteByte(c)
		inToken = true
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quoted string in curl command")
	}
	flush()
	return tokens, nil
}

func normalizeCurlContinuations(raw string) string {
	replacer := strings.NewReplacer(
		"\\\r\n", " ",
		"\\\n", " ",
		"^\r\n", " ",
		"^\n", " ",
	)
	return strings.TrimSpace(replacer.Replace(raw))
}

func isCurlExecutable(token string) bool {
	lower := strings.ToLower(strings.TrimSpace(token))
	return lower == "curl" || lower == "curl.exe"
}

func nextCurlValue(tokens []string, index *int, flag string) (string, bool) {
	next := *index + 1
	if next >= len(tokens) {
		return "", false
	}
	*index = next
	return tokens[next], true
}

func appendCurlHeader(headers *[]domain.KeyValue, raw string) {
	key, value, ok := strings.Cut(raw, ":")
	key = strings.TrimSpace(key)
	if !ok || key == "" {
		return
	}
	appendSyntheticHeader(headers, key, strings.TrimSpace(value))
}

func appendSyntheticHeader(headers *[]domain.KeyValue, key, value string) {
	if strings.TrimSpace(key) == "" {
		return
	}
	*headers = append(*headers, domain.KeyValue{
		ID:      uuid.NewString(),
		Enabled: true,
		Key:     strings.TrimSpace(key),
		Value:   strings.TrimSpace(value),
	})
}

func isCurlDataToken(token string) bool {
	if token == "-d" || token == "--data" || token == "--data-raw" || token == "--data-binary" || token == "--data-ascii" || token == "--data-urlencode" || token == "-F" || token == "--form" {
		return true
	}
	return strings.HasPrefix(token, "-d") ||
		strings.HasPrefix(token, "--data=") ||
		strings.HasPrefix(token, "--data-raw=") ||
		strings.HasPrefix(token, "--data-binary=") ||
		strings.HasPrefix(token, "--data-ascii=") ||
		strings.HasPrefix(token, "--data-urlencode=") ||
		strings.HasPrefix(token, "-F") ||
		strings.HasPrefix(token, "--form=")
}

func curlAttachedValue(token string) (string, bool) {
	switch {
	case strings.HasPrefix(token, "-d") && len(token) > 2:
		return token[2:], true
	case strings.HasPrefix(token, "-F") && len(token) > 2:
		return token[2:], true
	case strings.HasPrefix(token, "--data="):
		return strings.TrimPrefix(token, "--data="), true
	case strings.HasPrefix(token, "--data-raw="):
		return strings.TrimPrefix(token, "--data-raw="), true
	case strings.HasPrefix(token, "--data-binary="):
		return strings.TrimPrefix(token, "--data-binary="), true
	case strings.HasPrefix(token, "--data-ascii="):
		return strings.TrimPrefix(token, "--data-ascii="), true
	case strings.HasPrefix(token, "--data-urlencode="):
		return strings.TrimPrefix(token, "--data-urlencode="), true
	case strings.HasPrefix(token, "--form="):
		return strings.TrimPrefix(token, "--form="), true
	default:
		return "", false
	}
}

func curlFlagConsumesValue(token string) bool {
	switch token {
	case "-u", "--user", "--connect-timeout", "--max-time", "--proxy", "--proxy-user", "--resolve", "--cacert", "--cert", "--key", "--request-target", "--output", "-o":
		return true
	default:
		return false
	}
}

func isShellSpace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r' || c == '\t'
}
