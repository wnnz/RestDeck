package request

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"restdeck/internal/domain"
)

func TestSenderBuildsRequestAndRunsAgainstLocalServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "restdeck" {
			t.Fatalf("query param was not applied: %s", r.URL.RawQuery)
		}
		if r.Header.Get("Authorization") != "Bearer token-123" {
			t.Fatalf("bearer auth was not applied")
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer server.Close()

	req := domain.Request{
		Name:    "local",
		Method:  "GET",
		URL:     server.URL + "/anything",
		Params:  []domain.KeyValue{{Enabled: true, Key: "q", Value: "{{query}}"}},
		Headers: []domain.KeyValue{{Enabled: true, Key: "Accept", Value: "application/json"}},
		Auth:    domain.AuthConfig{Type: domain.AuthTypeBearer, Values: map[string]string{"token": "{{token}}"}},
	}
	env := domain.Environment{Variables: []domain.KeyValue{
		{Enabled: true, Key: "query", Value: "restdeck"},
		{Enabled: true, Key: "token", Value: "token-123"},
	}}

	res, err := NewSender().Send(t.Context(), req, env, nil)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	if res.SizeBytes == 0 {
		t.Fatalf("response body was empty")
	}
}

func TestSenderSendsMultipartFormWithFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "avatar.txt")
	if err := os.WriteFile(filePath, []byte("file-content"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("content type = %q", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		if got := r.FormValue("name"); got != "Ada" {
			t.Fatalf("name = %q", got)
		}
		file, header, err := r.FormFile("avatar")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		defer file.Close()
		if header.Filename != "avatar.txt" {
			t.Fatalf("filename = %q", header.Filename)
		}
		data, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("read file: %v", err)
		}
		if string(data) != "file-content" {
			t.Fatalf("file content = %q", data)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	req := domain.Request{
		Method:   "POST",
		URL:      server.URL,
		BodyMode: domain.BodyModeForm,
		FormItems: []domain.FormItem{
			{Enabled: true, Key: "name", Type: "text", Value: "Ada"},
			{Enabled: true, Key: "avatar", Type: "file", FilePath: filePath},
		},
	}
	res, err := NewSender().Send(t.Context(), req, domain.Environment{}, nil)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

func TestSenderReturnsErrorForMissingMultipartFile(t *testing.T) {
	req := domain.Request{
		Method:   "POST",
		URL:      "https://example.test/upload",
		BodyMode: domain.BodyModeForm,
		FormItems: []domain.FormItem{
			{Enabled: true, Key: "avatar", Type: "file", FilePath: filepath.Join(t.TempDir(), "missing.txt")},
		},
	}
	res, err := NewSender().Send(t.Context(), req, domain.Environment{}, nil)
	if err == nil {
		t.Fatal("expected missing file error")
	}
	if !strings.Contains(res.Error, "open form file") {
		t.Fatalf("error = %q", res.Error)
	}
}

func TestSenderSendsLegacyFormBodyAsMultipart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("content type = %q", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		if got := r.FormValue("role"); got != "admin" {
			t.Fatalf("role = %q", got)
		}
	}))
	defer server.Close()

	req := domain.Request{
		Method:   "POST",
		URL:      server.URL,
		BodyMode: domain.BodyModeForm,
		Body:     "role=admin",
	}
	res, err := NewSender().Send(t.Context(), req, domain.Environment{}, nil)
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
}

func TestSenderPreparesResolvedRequestPreview(t *testing.T) {
	sender := NewSender()
	sender.LoadCookies([]domain.Cookie{{Name: "sid", Value: "abc", Domain: "api.example.com", Path: "/", Secure: true}})
	req := domain.Request{
		Method:   "POST",
		URL:      "https://api.example.com/users/{{id}}",
		Params:   []domain.KeyValue{{Enabled: true, Key: "trace", Value: "{{trace}}"}},
		Headers:  []domain.KeyValue{{Enabled: true, Key: "X-Token", Value: "{{token}}"}},
		BodyMode: domain.BodyModeJSON,
		Body:     `{"name":"{{name}}"}`,
		Auth:     domain.AuthConfig{Type: domain.AuthTypeBearer, Values: map[string]string{"token": "{{token}}"}},
		Proxy:    domain.ProxyConfig{Mode: "inherit"},
	}
	variables := map[string]string{"id": "u-1", "trace": "t-1", "token": "secret", "name": "Ada"}

	preview, err := sender.PrepareRequest(t.Context(), req, variables, domain.ProxyConfig{Mode: "custom", URL: "http://127.0.0.1:7890", NoProxy: "localhost"})
	if err != nil {
		t.Fatalf("prepare request: %v", err)
	}
	if preview.URL != "https://api.example.com/users/u-1?trace=t-1" {
		t.Fatalf("url = %q", preview.URL)
	}
	if !preview.ProxyApplied || preview.ProxySource != "default" {
		t.Fatalf("proxy preview = %#v", preview)
	}
	if preview.Body.Text != `{"name":"Ada"}` {
		t.Fatalf("body preview = %q", preview.Body.Text)
	}
	if len(preview.Cookies) != 1 || preview.Cookies[0].Name != "sid" {
		t.Fatalf("cookies = %#v", preview.Cookies)
	}
	if !hasHeader(preview.Headers, "Authorization", "Bearer secret") || !hasHeader(preview.Headers, "X-Token", "secret") {
		t.Fatalf("headers = %#v", preview.Headers)
	}
}

func hasHeader(headers []domain.KeyValue, key, value string) bool {
	for _, header := range headers {
		if header.Key == key && header.Value == value {
			return true
		}
	}
	return false
}
