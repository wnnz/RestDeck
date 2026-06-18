package store

import (
	"testing"
	"time"

	"restdeck/internal/domain"
)

func TestStorePersistsWorkspaceState(t *testing.T) {
	s, err := OpenInMemory(t.Context())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	collection := domain.Collection{ID: "c1", Name: "Demo", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.SaveCollection(t.Context(), collection); err != nil {
		t.Fatalf("save collection: %v", err)
	}
	req := domain.Request{
		ID:           "r1",
		CollectionID: "c1",
		Name:         "Ping",
		Method:       "GET",
		URL:          "https://example.com",
		BodyMode:     domain.BodyModeNone,
		Auth:         domain.AuthConfig{Type: domain.AuthTypeNone, Values: map[string]string{}},
	}
	if err := s.SaveRequest(t.Context(), req); err != nil {
		t.Fatalf("save request: %v", err)
	}
	env := domain.Environment{ID: "e1", Name: "Local", IsActive: true, Variables: []domain.KeyValue{{Enabled: true, Key: "baseUrl", Value: "https://example.com"}}}
	if err := s.SaveEnvironment(t.Context(), env); err != nil {
		t.Fatalf("save environment: %v", err)
	}
	if err := s.AddHistory(t.Context(), domain.HistoryItem{ID: "h1", Name: "Ping", Method: "GET", URL: "https://example.com", Request: req, Response: domain.Response{StatusCode: 200}}); err != nil {
		t.Fatalf("add history: %v", err)
	}

	state, err := s.State(t.Context())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if len(state.Collections) != 1 || len(state.Collections[0].Requests) != 1 {
		t.Fatalf("unexpected collections: %#v", state.Collections)
	}
	if state.ActiveEnvironmentID != "e1" {
		t.Fatalf("active environment = %q", state.ActiveEnvironmentID)
	}
	if len(state.History) != 1 {
		t.Fatalf("history length = %d", len(state.History))
	}
}

func TestStoreRenamesCollection(t *testing.T) {
	s, err := OpenInMemory(t.Context())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	collection := domain.Collection{ID: "c1", Name: "Before", Description: "old", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.SaveCollection(t.Context(), collection); err != nil {
		t.Fatalf("save collection: %v", err)
	}

	collection.Name = "After"
	collection.Description = "new"
	if err := s.SaveCollection(t.Context(), collection); err != nil {
		t.Fatalf("rename collection: %v", err)
	}

	state, err := s.State(t.Context())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if len(state.Collections) != 1 {
		t.Fatalf("collections length = %d", len(state.Collections))
	}
	if state.Collections[0].Name != "After" || state.Collections[0].Description != "new" {
		t.Fatalf("collection was not renamed: %#v", state.Collections[0])
	}
}
