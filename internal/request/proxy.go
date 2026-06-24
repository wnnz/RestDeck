package request

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"golang.org/x/net/proxy"

	"restdeck/internal/domain"
)

func EffectiveProxy(requestProxy, defaultProxy domain.ProxyConfig) (domain.ProxyConfig, error) {
	return EffectiveProxyForURL(requestProxy, defaultProxy, "")
}

func EffectiveProxyForURL(requestProxy, defaultProxy domain.ProxyConfig, rawURL string) (domain.ProxyConfig, error) {
	effective, _, _, err := ResolveProxyForURL(requestProxy, defaultProxy, rawURL)
	return effective, err
}

func ResolveProxyForURL(requestProxy, defaultProxy domain.ProxyConfig, rawURL string) (domain.ProxyConfig, string, bool, error) {
	requestProxy = normalizeProxy(requestProxy, "inherit")
	requestProxy.NoProxy = ""
	defaultProxy = normalizeProxy(defaultProxy, "none")
	effective := requestProxy
	source := "request"
	if effective.Mode == "inherit" {
		effective = defaultProxy
		source = "default"
	}
	if effective.Mode != "custom" {
		return domain.ProxyConfig{Mode: "none"}, source, false, nil
	}
	if proxyExcluded(rawURL, effective.NoProxy) {
		return domain.ProxyConfig{Mode: "none"}, source, true, nil
	}
	if effective.URL == "" {
		return domain.ProxyConfig{}, source, false, fmt.Errorf("代理地址不能为空")
	}
	parsed, err := url.Parse(effective.URL)
	if err != nil {
		return domain.ProxyConfig{}, source, false, err
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "socks5":
		return domain.ProxyConfig{Mode: "custom", URL: effective.URL, NoProxy: effective.NoProxy}, source, false, nil
	default:
		return domain.ProxyConfig{}, source, false, fmt.Errorf("不支持的代理协议 %q", parsed.Scheme)
	}
}

func HTTPTransportForProxy(effective domain.ProxyConfig) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if effective.Mode != "custom" || effective.URL == "" {
		return transport, nil
	}
	parsed, err := url.Parse(effective.URL)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		transport.Proxy = http.ProxyURL(parsed)
	case "socks5":
		dialer, err := proxy.FromURL(parsed, proxy.Direct)
		if err != nil {
			return nil, err
		}
		contextDialer, ok := dialer.(proxy.ContextDialer)
		if ok {
			transport.DialContext = contextDialer.DialContext
		} else {
			transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
				type result struct {
					conn net.Conn
					err  error
				}
				ch := make(chan result, 1)
				go func() {
					conn, err := dialer.Dial(network, address)
					ch <- result{conn: conn, err: err}
				}()
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case result := <-ch:
					return result.conn, result.err
				}
			}
		}
	default:
		return nil, fmt.Errorf("不支持的代理协议 %q", parsed.Scheme)
	}
	return transport, nil
}

func normalizeProxy(proxyConfig domain.ProxyConfig, fallbackMode string) domain.ProxyConfig {
	proxyConfig.Mode = strings.TrimSpace(proxyConfig.Mode)
	proxyConfig.URL = strings.TrimSpace(proxyConfig.URL)
	proxyConfig.NoProxy = normalizeNoProxy(proxyConfig.NoProxy)
	switch proxyConfig.Mode {
	case "inherit", "none", "custom":
	default:
		proxyConfig.Mode = fallbackMode
	}
	if proxyConfig.Mode == "" {
		proxyConfig.Mode = fallbackMode
	}
	if proxyConfig.Mode != "custom" {
		proxyConfig.URL = ""
		proxyConfig.NoProxy = ""
	}
	return proxyConfig
}

func proxyExcluded(rawURL, rules string) bool {
	if strings.TrimSpace(rawURL) == "" || strings.TrimSpace(rules) == "" {
		return false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return false
	}
	for _, rule := range splitNoProxy(rules) {
		rule = strings.ToLower(strings.TrimSpace(rule))
		rule = strings.TrimPrefix(rule, "http://")
		rule = strings.TrimPrefix(rule, "https://")
		rule = strings.TrimPrefix(rule, "ws://")
		rule = strings.TrimPrefix(rule, "wss://")
		if strings.Contains(rule, "://") {
			if parsedRule, err := url.Parse(rule); err == nil {
				rule = parsedRule.Hostname()
			}
		}
		if strings.Contains(rule, ":") {
			ruleHost, _, err := net.SplitHostPort(rule)
			if err == nil {
				rule = ruleHost
			}
		}
		rule = strings.Trim(rule, "[]")
		if rule == "" {
			continue
		}
		if rule == "*" || rule == host {
			return true
		}
		if strings.HasPrefix(rule, ".") && (host == strings.TrimPrefix(rule, ".") || strings.HasSuffix(host, rule)) {
			return true
		}
		if strings.HasPrefix(rule, "*.") && strings.HasSuffix(host, strings.TrimPrefix(rule, "*")) {
			return true
		}
		if matched, _ := filepath.Match(rule, host); matched {
			return true
		}
	}
	return false
}

func normalizeNoProxy(raw string) string {
	return strings.Join(splitNoProxy(raw), ",")
}

func splitNoProxy(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	out := []string{}
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		out = append(out, part)
	}
	return out
}
