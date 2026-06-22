package request

import (
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
}
