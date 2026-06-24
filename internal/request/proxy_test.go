package request

import (
	"strings"
	"testing"

	"restdeck/internal/domain"
)

func TestEffectiveProxyResolvesModes(t *testing.T) {
	defaultProxy := domain.ProxyConfig{Mode: "custom", URL: "socks5://127.0.0.1:10808"}
	got, err := EffectiveProxy(domain.ProxyConfig{Mode: "inherit"}, defaultProxy)
	if err != nil {
		t.Fatalf("inherit proxy: %v", err)
	}
	if got.URL != defaultProxy.URL {
		t.Fatalf("proxy URL = %q", got.URL)
	}
	got, err = EffectiveProxy(domain.ProxyConfig{Mode: "none"}, defaultProxy)
	if err != nil {
		t.Fatalf("none proxy: %v", err)
	}
	if got.Mode != "none" || got.URL != "" {
		t.Fatalf("none proxy = %#v", got)
	}
	got, err = EffectiveProxy(domain.ProxyConfig{Mode: "custom", URL: "http://127.0.0.1:7890"}, defaultProxy)
	if err != nil {
		t.Fatalf("custom proxy: %v", err)
	}
	if got.URL != "http://127.0.0.1:7890" {
		t.Fatalf("custom proxy = %#v", got)
	}
}

func TestEffectiveProxyRejectsUnsupportedScheme(t *testing.T) {
	_, err := EffectiveProxy(domain.ProxyConfig{Mode: "custom", URL: "ftp://127.0.0.1:21"}, domain.ProxyConfig{Mode: "none"})
	if err == nil {
		t.Fatal("expected unsupported proxy scheme")
	}
	if !strings.Contains(err.Error(), "不支持的代理协议") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestEffectiveProxyRequiresCustomProxyURL(t *testing.T) {
	_, err := EffectiveProxy(domain.ProxyConfig{Mode: "custom"}, domain.ProxyConfig{Mode: "none"})
	if err == nil {
		t.Fatal("expected required proxy URL")
	}
	if err.Error() != "代理地址不能为空" {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestEffectiveProxyForURLBypassesNoProxyHosts(t *testing.T) {
	defaultProxy := domain.ProxyConfig{Mode: "custom", URL: "http://127.0.0.1:7890", NoProxy: "localhost, 127.0.0.1, .internal"}
	cases := []string{
		"http://localhost:8080/api",
		"http://127.0.0.1:8080/api",
		"https://api.internal/users",
	}
	for _, rawURL := range cases {
		got, err := EffectiveProxyForURL(domain.ProxyConfig{Mode: "inherit"}, defaultProxy, rawURL)
		if err != nil {
			t.Fatalf("effective proxy for %s: %v", rawURL, err)
		}
		if got.Mode != "none" {
			t.Fatalf("effective proxy for %s = %#v", rawURL, got)
		}
	}
}

func TestEffectiveProxyForURLUsesProxyWhenNoProxyMisses(t *testing.T) {
	defaultProxy := domain.ProxyConfig{Mode: "custom", URL: "http://127.0.0.1:7890", NoProxy: "localhost,127.0.0.1"}
	got, err := EffectiveProxyForURL(domain.ProxyConfig{Mode: "inherit"}, defaultProxy, "https://example.com")
	if err != nil {
		t.Fatalf("effective proxy: %v", err)
	}
	if got.Mode != "custom" || got.URL != defaultProxy.URL {
		t.Fatalf("effective proxy = %#v", got)
	}
}

func TestEffectiveProxyForURLIgnoresRequestNoProxy(t *testing.T) {
	requestProxy := domain.ProxyConfig{Mode: "custom", URL: "http://127.0.0.1:7890", NoProxy: "localhost,127.0.0.1"}
	got, err := EffectiveProxyForURL(requestProxy, domain.ProxyConfig{Mode: "none"}, "http://localhost:8080/api")
	if err != nil {
		t.Fatalf("effective proxy: %v", err)
	}
	if got.Mode != "custom" || got.URL != requestProxy.URL || got.NoProxy != "" {
		t.Fatalf("effective proxy = %#v", got)
	}
}
