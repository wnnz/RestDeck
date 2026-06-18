package request

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
