package request

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/proxy"

	"restdeck/internal/domain"
)

func EffectiveProxy(requestProxy, defaultProxy domain.ProxyConfig) (domain.ProxyConfig, error) {
	requestProxy = normalizeProxy(requestProxy, "inherit")
	defaultProxy = normalizeProxy(defaultProxy, "none")
	effective := requestProxy
	if effective.Mode == "inherit" {
		effective = defaultProxy
	}
	if effective.Mode != "custom" {
		return domain.ProxyConfig{Mode: "none"}, nil
	}
	if effective.URL == "" {
		return domain.ProxyConfig{}, fmt.Errorf("proxy URL is required")
	}
	parsed, err := url.Parse(effective.URL)
	if err != nil {
		return domain.ProxyConfig{}, err
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "socks5":
		return domain.ProxyConfig{Mode: "custom", URL: effective.URL}, nil
	default:
		return domain.ProxyConfig{}, fmt.Errorf("unsupported proxy scheme %q", parsed.Scheme)
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
		return nil, fmt.Errorf("unsupported proxy scheme %q", parsed.Scheme)
	}
	return transport, nil
}

func normalizeProxy(proxyConfig domain.ProxyConfig, fallbackMode string) domain.ProxyConfig {
	proxyConfig.Mode = strings.TrimSpace(proxyConfig.Mode)
	proxyConfig.URL = strings.TrimSpace(proxyConfig.URL)
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
	}
	return proxyConfig
}
