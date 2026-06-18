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
	"net/http"
	"net/http/cookiejar"
	"net/url"
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

func (s *Sender) Send(ctx context.Context, req domain.Request, env domain.Environment, globals []domain.KeyValue) (domain.Response, error) {
	resolver := NewResolver(env, globals)
	return s.SendWithVariables(ctx, req, resolver.Values())
}

func (s *Sender) SendWithVariables(ctx context.Context, req domain.Request, variables map[string]string) (domain.Response, error) {
	resolver := NewResolver(domain.Environment{}, nil)
	resolver.values = variables
	httpReq, err := buildHTTPRequest(ctx, req, resolver)
	if err != nil {
		return domain.Response{Error: err.Error()}, err
	}
	timeout := req.TimeoutMs
	if timeout <= 0 {
		timeout = 30000
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()
	httpReq = httpReq.WithContext(ctx)

	start := time.Now()
	httpRes, err := s.client.Do(httpReq)
	duration := time.Since(start)
	if err != nil {
		return domain.Response{
			DurationMs:   duration.Milliseconds(),
			Error:        err.Error(),
			RequestedURL: httpReq.URL.String(),
		}, err
	}
	defer httpRes.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(httpRes.Body, 20*1024*1024))
	if readErr != nil {
		return domain.Response{Error: readErr.Error(), RequestedURL: httpReq.URL.String()}, readErr
	}

	headers := make([]domain.KeyValue, 0, len(httpRes.Header))
	for key, values := range httpRes.Header {
		headers = append(headers, domain.KeyValue{Enabled: true, Key: key, Value: strings.Join(values, ", ")})
	}
	sort.Slice(headers, func(i, j int) bool { return headers[i].Key < headers[j].Key })

	cookies := []domain.Cookie{}
	for _, cookie := range httpRes.Cookies() {
		cookies = append(cookies, domain.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
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
	}, nil
}

func buildHTTPRequest(ctx context.Context, req domain.Request, resolver *Resolver) (*http.Request, error) {
	rawURL := strings.TrimSpace(resolver.Resolve(req.URL))
	if rawURL == "" {
		return nil, fmt.Errorf("request URL is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	query := parsed.Query()
	for _, param := range req.Params {
		if param.Enabled && param.Key != "" {
			query.Set(resolver.Resolve(param.Key), resolver.Resolve(param.Value))
		}
	}
	parsed.RawQuery = query.Encode()

	var body io.Reader
	switch req.BodyMode {
	case domain.BodyModeJSON, domain.BodyModeRaw:
		body = strings.NewReader(resolver.Resolve(req.Body))
	case domain.BodyModeForm, domain.BodyModeURLEncoded:
		values := url.Values{}
		for _, row := range parseBodyRows(req.Body) {
			if row.Enabled && row.Key != "" {
				values.Set(resolver.Resolve(row.Key), resolver.Resolve(row.Value))
			}
		}
		body = strings.NewReader(values.Encode())
	default:
		body = nil
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, parsed.String(), body)
	if err != nil {
		return nil, err
	}

	for _, header := range req.Headers {
		if header.Enabled && header.Key != "" {
			httpReq.Header.Set(resolver.Resolve(header.Key), resolver.Resolve(header.Value))
		}
	}
	if req.BodyMode == domain.BodyModeJSON && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if req.BodyMode == domain.BodyModeURLEncoded && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	applyAuth(httpReq, req.Auth, resolver)
	return httpReq, nil
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
