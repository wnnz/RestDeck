package request

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"restdeck/internal/domain"
)

type Sender struct {
	client *http.Client
}

func NewSender() *Sender {
	jar, _ := cookiejar.New(nil)
	return &Sender{
		client: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Sender) LoadCookies(cookies []domain.Cookie) {
	if s == nil || s.client == nil || s.client.Jar == nil {
		return
	}
	for _, cookie := range cookies {
		host := strings.TrimSpace(cookie.Domain)
		if host == "" || cookie.Name == "" {
			continue
		}
		host = strings.TrimPrefix(host, ".")
		scheme := "http"
		if cookie.Secure {
			scheme = "https"
		}
		u := &url.URL{Scheme: scheme, Host: host, Path: "/"}
		path := cookie.Path
		if path == "" {
			path = "/"
		}
		s.client.Jar.SetCookies(u, []*http.Cookie{{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     path,
			Domain:   cookie.Domain,
			Expires:  cookie.Expires,
			HttpOnly: cookie.HTTPOnly,
			Secure:   cookie.Secure,
		}})
	}
}

func (s *Sender) CookiesForURL(rawURL string) []domain.Cookie {
	if s == nil || s.client == nil || s.client.Jar == nil {
		return nil
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return nil
	}
	out := []domain.Cookie{}
	for _, cookie := range s.client.Jar.Cookies(u) {
		out = append(out, domain.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   u.Hostname(),
			Path:     fallbackPath(cookie.Path),
			Expires:  cookie.Expires,
			HTTPOnly: cookie.HttpOnly,
			Secure:   cookie.Secure,
		})
	}
	return out
}

func (s *Sender) Send(ctx context.Context, req domain.Request, env domain.Environment, globals []domain.KeyValue) (domain.Response, error) {
	resolver := NewResolver(env, globals)
	variables, err := resolver.ValuesWithError()
	if err != nil {
		return domain.Response{Error: err.Error()}, err
	}
	return s.SendWithVariables(ctx, req, variables)
}

func (s *Sender) SendWithVariables(ctx context.Context, req domain.Request, variables map[string]string) (domain.Response, error) {
	return s.SendWithVariablesAndProxy(ctx, req, variables, domain.ProxyConfig{Mode: "none"})
}

func (s *Sender) PrepareRequest(ctx context.Context, req domain.Request, variables map[string]string, defaultProxy domain.ProxyConfig) (domain.PreparedRequest, error) {
	resolver := NewResolver(domain.Environment{}, nil)
	resolver.values = variables
	_, prepared, _, err := s.buildPreparedRequest(ctx, req, resolver, defaultProxy)
	return prepared, err
}

func (s *Sender) SendWithVariablesAndProxy(ctx context.Context, req domain.Request, variables map[string]string, defaultProxy domain.ProxyConfig) (domain.Response, error) {
	resolver := NewResolver(domain.Environment{}, nil)
	resolver.values = variables
	httpReq, prepared, effectiveProxy, err := s.buildPreparedRequest(ctx, req, resolver, defaultProxy)
	if err != nil {
		return domain.Response{Error: err.Error(), RequestedURL: prepared.URL, Request: prepared}, err
	}
	transport, err := HTTPTransportForProxy(effectiveProxy)
	if err != nil {
		prepared.Error = err.Error()
		return domain.Response{Error: err.Error(), RequestedURL: prepared.URL, Request: prepared}, err
	}
	timeout := req.TimeoutMs
	if timeout <= 0 {
		timeout = 30000
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()
	httpReq = httpReq.WithContext(ctx)

	start := time.Now()
	client := &http.Client{
		Jar:       s.client.Jar,
		Timeout:   time.Duration(timeout) * time.Millisecond,
		Transport: transport,
	}
	httpRes, err := client.Do(httpReq)
	duration := time.Since(start)
	if err != nil {
		return domain.Response{
			DurationMs:   duration.Milliseconds(),
			Error:        err.Error(),
			RequestedURL: httpReq.URL.String(),
			Request:      prepared,
		}, err
	}
	defer httpRes.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(httpRes.Body, 20*1024*1024))
	if readErr != nil {
		return domain.Response{Error: readErr.Error(), RequestedURL: httpReq.URL.String(), Request: prepared}, readErr
	}

	headers := make([]domain.KeyValue, 0, len(httpRes.Header))
	for key, values := range httpRes.Header {
		headers = append(headers, domain.KeyValue{Enabled: true, Key: key, Value: strings.Join(values, ", ")})
	}
	sort.Slice(headers, func(i, j int) bool { return headers[i].Key < headers[j].Key })

	cookies := []domain.Cookie{}
	for _, cookie := range httpRes.Cookies() {
		path := cookie.Path
		if path == "" {
			path = "/"
		}
		domainName := cookie.Domain
		if domainName == "" {
			domainName = httpReq.URL.Hostname()
		}
		cookies = append(cookies, domain.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   domainName,
			Path:     path,
			Expires:  cookie.Expires,
			HTTPOnly: cookie.HttpOnly,
			Secure:   cookie.Secure,
		})
	}

	return domain.Response{
		StatusCode:   httpRes.StatusCode,
		Status:       httpRes.Status,
		DurationMs:   duration.Milliseconds(),
		SizeBytes:    int64(len(body)),
		Headers:      headers,
		Cookies:      cookies,
		Body:         string(body),
		ContentType:  httpRes.Header.Get("Content-Type"),
		RequestedURL: httpReq.URL.String(),
		Request:      prepared,
	}, nil
}

func (s *Sender) buildPreparedRequest(ctx context.Context, req domain.Request, resolver *Resolver, defaultProxy domain.ProxyConfig) (*http.Request, domain.PreparedRequest, domain.ProxyConfig, error) {
	httpReq, preparedBody, err := buildHTTPRequestWithPrepared(ctx, req, resolver)
	prepared := domain.PreparedRequest{Body: preparedBody}
	if err != nil {
		prepared.Error = err.Error()
		return nil, prepared, domain.ProxyConfig{Mode: "none"}, err
	}
	effectiveProxy, proxySource, proxyExcluded, err := ResolveProxyForURL(req.Proxy, defaultProxy, httpReq.URL.String())
	prepared = domain.PreparedRequest{
		Method:        httpReq.Method,
		URL:           httpReq.URL.String(),
		Headers:       headersFromRequest(httpReq),
		Cookies:       s.CookiesForURL(httpReq.URL.String()),
		Body:          preparedBody,
		Proxy:         effectiveProxy,
		ProxyApplied:  effectiveProxy.Mode == "custom",
		ProxyExcluded: proxyExcluded,
		ProxySource:   proxySource,
	}
	if err != nil {
		prepared.Error = err.Error()
		return httpReq, prepared, domain.ProxyConfig{Mode: "none"}, err
	}
	return httpReq, prepared, effectiveProxy, nil
}

func buildHTTPRequest(ctx context.Context, req domain.Request, resolver *Resolver) (*http.Request, error) {
	httpReq, _, err := buildHTTPRequestWithPrepared(ctx, req, resolver)
	return httpReq, err
}

func buildHTTPRequestWithPrepared(ctx context.Context, req domain.Request, resolver *Resolver) (*http.Request, domain.PreparedBody, error) {
	rawURL := strings.TrimSpace(resolver.Resolve(req.URL))
	if rawURL == "" {
		return nil, domain.PreparedBody{}, fmt.Errorf("request URL is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, domain.PreparedBody{}, err
	}
	query := parsed.Query()
	for _, param := range req.Params {
		if param.Enabled && param.Key != "" {
			query.Set(resolver.Resolve(param.Key), resolver.Resolve(param.Value))
		}
	}
	parsed.RawQuery = query.Encode()

	var body io.Reader
	var bodyPreview domain.PreparedBody
	multipartContentType := ""
	switch req.BodyMode {
	case domain.BodyModeJSON, domain.BodyModeRaw:
		bodyText := resolver.Resolve(req.Body)
		body = strings.NewReader(bodyText)
		bodyPreview = preparedTextBody(req.BodyMode, "", bodyText)
	case domain.BodyModeForm:
		var err error
		body, multipartContentType, bodyPreview, err = buildMultipartBody(req, resolver)
		if err != nil {
			return nil, bodyPreview, err
		}
	case domain.BodyModeURLEncoded:
		values := url.Values{}
		for _, row := range parseBodyRows(req.Body) {
			if row.Enabled && row.Key != "" {
				values.Set(resolver.Resolve(row.Key), resolver.Resolve(row.Value))
			}
		}
		bodyText := values.Encode()
		body = strings.NewReader(bodyText)
		bodyPreview = preparedTextBody(req.BodyMode, "application/x-www-form-urlencoded", bodyText)
	default:
		body = nil
		bodyPreview = domain.PreparedBody{Mode: req.BodyMode}
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, parsed.String(), body)
	if err != nil {
		return nil, bodyPreview, err
	}

	for _, header := range req.Headers {
		if header.Enabled && header.Key != "" {
			httpReq.Header.Set(resolver.Resolve(header.Key), resolver.Resolve(header.Value))
		}
	}
	if req.BodyMode == domain.BodyModeJSON && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if req.BodyMode == domain.BodyModeForm && multipartContentType != "" {
		httpReq.Header.Set("Content-Type", multipartContentType)
	}
	if req.BodyMode == domain.BodyModeURLEncoded && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	applyAuth(httpReq, req.Auth, resolver)
	bodyPreview.ContentType = httpReq.Header.Get("Content-Type")
	return httpReq, bodyPreview, nil
}

func preparedTextBody(mode domain.BodyMode, contentType, text string) domain.PreparedBody {
	const maxPreviewBytes = 64 * 1024
	preview := text
	truncated := false
	if len(preview) > maxPreviewBytes {
		preview = preview[:maxPreviewBytes]
		truncated = true
	}
	return domain.PreparedBody{
		Mode:        mode,
		ContentType: contentType,
		Text:        preview,
		SizeBytes:   int64(len(text)),
		Truncated:   truncated,
	}
}

func headersFromRequest(req *http.Request) []domain.KeyValue {
	headers := make([]domain.KeyValue, 0, len(req.Header))
	for key, values := range req.Header {
		headers = append(headers, domain.KeyValue{Enabled: true, Key: key, Value: strings.Join(values, ", ")})
	}
	sort.Slice(headers, func(i, j int) bool { return headers[i].Key < headers[j].Key })
	return headers
}

func buildMultipartBody(req domain.Request, resolver *Resolver) (io.Reader, string, domain.PreparedBody, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	previewItems := []domain.FormItem{}
	previewLines := []string{}
	for _, item := range normalizeFormItems(req.FormItems, req.Body) {
		if !item.Enabled || item.Key == "" {
			continue
		}
		key := resolver.Resolve(item.Key)
		if key == "" {
			continue
		}
		previewItem := item
		previewItem.Key = key
		if item.Type == "file" {
			path := strings.TrimSpace(resolver.Resolve(item.FilePath))
			if path == "" {
				return nil, "", domain.PreparedBody{Mode: domain.BodyModeForm}, fmt.Errorf("form file path for %q is required", item.Key)
			}
			file, err := os.Open(path)
			if err != nil {
				return nil, "", domain.PreparedBody{Mode: domain.BodyModeForm}, fmt.Errorf("open form file %q: %w", path, err)
			}
			info, statErr := file.Stat()
			part, err := writer.CreateFormFile(key, filepath.Base(path))
			if err != nil {
				_ = file.Close()
				return nil, "", domain.PreparedBody{Mode: domain.BodyModeForm}, err
			}
			if _, err := io.Copy(part, file); err != nil {
				_ = file.Close()
				return nil, "", domain.PreparedBody{Mode: domain.BodyModeForm}, err
			}
			if err := file.Close(); err != nil {
				return nil, "", domain.PreparedBody{Mode: domain.BodyModeForm}, err
			}
			previewItem.FilePath = path
			previewItem.Value = ""
			sizeText := ""
			if statErr == nil {
				sizeText = fmt.Sprintf(", %d bytes", info.Size())
			}
			previewLines = append(previewLines, fmt.Sprintf("%s=@%s (%s%s)", key, path, filepath.Base(path), sizeText))
			previewItems = append(previewItems, previewItem)
			continue
		}
		value := resolver.Resolve(item.Value)
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", domain.PreparedBody{Mode: domain.BodyModeForm}, err
		}
		previewItem.Value = value
		previewItem.FilePath = ""
		previewLines = append(previewLines, key+"="+value)
		previewItems = append(previewItems, previewItem)
	}
	if err := writer.Close(); err != nil {
		return nil, "", domain.PreparedBody{Mode: domain.BodyModeForm}, err
	}
	contentType := writer.FormDataContentType()
	preview := preparedTextBody(domain.BodyModeForm, contentType, strings.Join(previewLines, "\n"))
	preview.FormItems = previewItems
	preview.SizeBytes = int64(buf.Len())
	return &buf, contentType, preview, nil
}

func normalizeFormItems(items []domain.FormItem, fallbackBody string) []domain.FormItem {
	if len(items) == 0 && fallbackBody != "" {
		items = formItemsFromBody(fallbackBody)
	}
	out := []domain.FormItem{}
	for _, item := range items {
		if item.Type != "file" {
			item.Type = "text"
		}
		out = append(out, item)
	}
	return out
}

func formItemsFromBody(raw string) []domain.FormItem {
	items := []domain.FormItem{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, _ := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		item := domain.FormItem{Enabled: true, Key: key, Type: "text", Value: value}
		if strings.HasPrefix(value, "@") {
			item.Type = "file"
			item.Value = ""
			item.FilePath = strings.TrimPrefix(value, "@")
		}
		items = append(items, item)
	}
	return items
}

func parseBodyRows(raw string) []domain.KeyValue {
	lines := strings.Split(raw, "\n")
	out := []domain.KeyValue{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		out = append(out, domain.KeyValue{Enabled: true, Key: parts[0], Value: value})
	}
	return out
}

func applyAuth(req *http.Request, auth domain.AuthConfig, resolver *Resolver) {
	values := auth.Values
	if values == nil {
		values = map[string]string{}
	}
	switch auth.Type {
	case domain.AuthTypeAPIKey:
		key := resolver.Resolve(values["key"])
		value := resolver.Resolve(values["value"])
		if key == "" {
			return
		}
		if values["in"] == "query" {
			q := req.URL.Query()
			q.Set(key, value)
			req.URL.RawQuery = q.Encode()
			return
		}
		req.Header.Set(key, value)
	case domain.AuthTypeBearer:
		token := resolver.Resolve(values["token"])
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case domain.AuthTypeBasic:
		req.SetBasicAuth(resolver.Resolve(values["username"]), resolver.Resolve(values["password"]))
	case domain.AuthTypeOAuth2:
		token := resolver.Resolve(values["accessToken"])
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case domain.AuthTypeOAuth1:
		applyOAuth1(req, values, resolver)
	case domain.AuthTypeDigest:
		username := resolver.Resolve(values["username"])
		password := resolver.Resolve(values["password"])
		if username != "" || password != "" {
			req.Header.Set("Authorization", "Digest username=\""+escapeHeader(username)+"\", password=\""+escapeHeader(password)+"\"")
		}
	}
}

func applyOAuth1(req *http.Request, values map[string]string, resolver *Resolver) {
	consumerKey := resolver.Resolve(values["consumerKey"])
	consumerSecret := resolver.Resolve(values["consumerSecret"])
	token := resolver.Resolve(values["token"])
	tokenSecret := resolver.Resolve(values["tokenSecret"])
	if consumerKey == "" {
		return
	}
	nonce := randomNonce()
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	params := map[string]string{
		"oauth_consumer_key":     consumerKey,
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        timestamp,
		"oauth_version":          "1.0",
	}
	if token != "" {
		params["oauth_token"] = token
	}
	signature := oauth1Signature(req, params, consumerSecret, tokenSecret)
	params["oauth_signature"] = signature
	parts := make([]string, 0, len(params))
	for key, value := range params {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, percent(key), percent(value)))
	}
	sort.Strings(parts)
	req.Header.Set("Authorization", "OAuth "+strings.Join(parts, ", "))
}

func oauth1Signature(req *http.Request, oauthParams map[string]string, consumerSecret, tokenSecret string) string {
	params := map[string]string{}
	for key, values := range req.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	for key, value := range oauthParams {
		params[key] = value
	}
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	paramParts := make([]string, 0, len(keys))
	for _, key := range keys {
		paramParts = append(paramParts, percent(key)+"="+percent(params[key]))
	}
	baseURL := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	base := strings.ToUpper(req.Method) + "&" + percent(baseURL) + "&" + percent(strings.Join(paramParts, "&"))
	signingKey := percent(consumerSecret) + "&" + percent(tokenSecret)
	mac := hmac.New(sha1.New, []byte(signingKey))
	_, _ = mac.Write([]byte(base))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func percent(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}

func randomNonce() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}

func escapeHeader(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}

func fallbackPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return "/"
	}
	return path
}

func FormatBody(contentType, body string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err == nil && strings.Contains(mediaType, "json") {
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(body), "", "  "); err == nil {
			return buf.String()
		}
	}
	return body
}
