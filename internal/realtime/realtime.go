package realtime

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"restdeck/internal/domain"
	reqsvc "restdeck/internal/request"
)

type WebSocketRequest struct {
	URL       string             `json:"url"`
	Message   string             `json:"message"`
	Headers   []domain.KeyValue  `json:"headers"`
	Proxy     domain.ProxyConfig `json:"proxy"`
	TimeoutMs int                `json:"timeoutMs"`
}

type WebSocketResult struct {
	Connected  bool     `json:"connected"`
	Sent       string   `json:"sent"`
	Received   []string `json:"received"`
	DurationMs int64    `json:"durationMs"`
	Error      string   `json:"error"`
}

type SSERequest struct {
	URL       string             `json:"url"`
	Headers   []domain.KeyValue  `json:"headers"`
	Proxy     domain.ProxyConfig `json:"proxy"`
	TimeoutMs int                `json:"timeoutMs"`
	MaxEvents int                `json:"maxEvents"`
}

type SSEResult struct {
	StatusCode int      `json:"statusCode"`
	Events     []string `json:"events"`
	DurationMs int64    `json:"durationMs"`
	Error      string   `json:"error"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) TestWebSocket(ctx context.Context, input WebSocketRequest, env domain.Environment, globals []domain.KeyValue, defaultProxy domain.ProxyConfig) WebSocketResult {
	timeout := time.Duration(defaultInt(input.TimeoutMs, 10000)) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resolver := reqsvc.NewResolver(env, globals)
	start := time.Now()
	headers := http.Header{}
	for _, header := range input.Headers {
		if header.Enabled && header.Key != "" {
			headers.Set(resolver.Resolve(header.Key), resolver.Resolve(header.Value))
		}
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: timeout,
		TLSClientConfig:  &tls.Config{MinVersion: tls.VersionTLS12},
	}
	effectiveProxy, err := reqsvc.EffectiveProxy(input.Proxy, defaultProxy)
	if err != nil {
		return WebSocketResult{DurationMs: time.Since(start).Milliseconds(), Error: err.Error()}
	}
	transport, err := reqsvc.HTTPTransportForProxy(effectiveProxy)
	if err != nil {
		return WebSocketResult{DurationMs: time.Since(start).Milliseconds(), Error: err.Error()}
	}
	dialer.Proxy = transport.Proxy
	dialer.NetDialContext = transport.DialContext
	conn, _, err := dialer.DialContext(ctx, resolver.Resolve(input.URL), headers)
	if err != nil {
		return WebSocketResult{DurationMs: time.Since(start).Milliseconds(), Error: err.Error()}
	}
	defer conn.Close()

	result := WebSocketResult{Connected: true, Sent: resolver.Resolve(input.Message)}
	if result.Sent != "" {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(result.Sent)); err != nil {
			result.Error = err.Error()
			result.DurationMs = time.Since(start).Milliseconds()
			return result
		}
	}
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	for len(result.Received) < 5 {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if len(result.Received) == 0 {
				result.Error = err.Error()
			}
			break
		}
		result.Received = append(result.Received, string(data))
		if result.Sent != "" {
			break
		}
	}
	result.DurationMs = time.Since(start).Milliseconds()
	return result
}

func (s *Service) TestSSE(ctx context.Context, input SSERequest, env domain.Environment, globals []domain.KeyValue, defaultProxy domain.ProxyConfig) SSEResult {
	timeout := time.Duration(defaultInt(input.TimeoutMs, 10000)) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resolver := reqsvc.NewResolver(env, globals)
	start := time.Now()
	rawURL := resolver.Resolve(input.URL)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return SSEResult{DurationMs: time.Since(start).Milliseconds(), Error: err.Error()}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return SSEResult{DurationMs: time.Since(start).Milliseconds(), Error: err.Error()}
	}
	req.Header.Set("Accept", "text/event-stream")
	for _, header := range input.Headers {
		if header.Enabled && header.Key != "" {
			req.Header.Set(resolver.Resolve(header.Key), resolver.Resolve(header.Value))
		}
	}
	effectiveProxy, err := reqsvc.EffectiveProxy(input.Proxy, defaultProxy)
	if err != nil {
		return SSEResult{DurationMs: time.Since(start).Milliseconds(), Error: err.Error()}
	}
	transport, err := reqsvc.HTTPTransportForProxy(effectiveProxy)
	if err != nil {
		return SSEResult{DurationMs: time.Since(start).Milliseconds(), Error: err.Error()}
	}
	client := http.Client{Timeout: timeout, Transport: transport}
	res, err := client.Do(req)
	if err != nil {
		return SSEResult{DurationMs: time.Since(start).Milliseconds(), Error: err.Error()}
	}
	defer res.Body.Close()

	events, err := readSSEEvents(ctx, res.Body, defaultInt(input.MaxEvents, 5))
	result := SSEResult{
		StatusCode: res.StatusCode,
		Events:     events,
		DurationMs: time.Since(start).Milliseconds(),
	}
	if err != nil && len(events) == 0 {
		result.Error = err.Error()
	}
	return result
}

func readSSEEvents(ctx context.Context, body io.Reader, maxEvents int) ([]string, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	events := []string{}
	current := []string{}
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return events, ctx.Err()
		default:
		}
		line := scanner.Text()
		if line == "" {
			if len(current) > 0 {
				events = append(events, strings.Join(current, "\n"))
				current = nil
				if len(events) >= maxEvents {
					return events, nil
				}
			}
			continue
		}
		if strings.HasPrefix(line, "data:") || strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "id:") {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		events = append(events, strings.Join(current, "\n"))
	}
	if err := scanner.Err(); err != nil {
		return events, err
	}
	if len(events) == 0 {
		return events, fmt.Errorf("no SSE events received")
	}
	return events, nil
}

func defaultInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
