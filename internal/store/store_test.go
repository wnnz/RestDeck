package store

import (
	"os"
	"path/filepath"
	"strings"
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

func TestStoreDeletesActiveEnvironmentAndSelectsNext(t *testing.T) {
	s, err := OpenInMemory(t.Context())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	if err := s.SaveEnvironment(t.Context(), domain.Environment{ID: "e1", Name: "One", IsActive: true}); err != nil {
		t.Fatalf("save e1: %v", err)
	}
	if err := s.SaveEnvironment(t.Context(), domain.Environment{ID: "e2", Name: "Two"}); err != nil {
		t.Fatalf("save e2: %v", err)
	}
	if err := s.DeleteEnvironment(t.Context(), "e1"); err != nil {
		t.Fatalf("delete active environment: %v", err)
	}
	state, err := s.State(t.Context())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.ActiveEnvironmentID != "e2" {
		t.Fatalf("active environment = %q", state.ActiveEnvironmentID)
	}
}

func TestStoreRejectsDeletingLastEnvironment(t *testing.T) {
	s, err := OpenInMemory(t.Context())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	if err := s.SaveEnvironment(t.Context(), domain.Environment{ID: "e1", Name: "One", IsActive: true}); err != nil {
		t.Fatalf("save e1: %v", err)
	}
	err = s.DeleteEnvironment(t.Context(), "e1")
	if err == nil || !strings.Contains(err.Error(), "last environment") {
		t.Fatalf("delete last environment err = %v", err)
	}
}

func TestStorePersistsSettingsAndRequestProxy(t *testing.T) {
	s, err := OpenInMemory(t.Context())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	settings := domain.Settings{Language: "zh-CN", Theme: "dark", DefaultProxy: domain.ProxyConfig{Mode: "custom", URL: "http://127.0.0.1:7890"}}
	if err := s.SaveSettings(t.Context(), settings); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	got, err := s.GetSettings(t.Context())
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if got.DefaultProxy.URL != settings.DefaultProxy.URL || got.Theme != "dark" {
		t.Fatalf("settings = %#v", got)
	}

	collection := domain.Collection{ID: "c1", Name: "Demo", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.SaveCollection(t.Context(), collection); err != nil {
		t.Fatalf("save collection: %v", err)
	}
	req := domain.Request{ID: "r1", CollectionID: "c1", Name: "Ping", Method: "GET", URL: "https://example.com", Proxy: domain.ProxyConfig{Mode: "none"}}
	if err := s.SaveRequest(t.Context(), req); err != nil {
		t.Fatalf("save request: %v", err)
	}
	state, err := s.State(t.Context())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if state.Collections[0].Requests[0].Proxy.Mode != "none" {
		t.Fatalf("request proxy = %#v", state.Collections[0].Requests[0].Proxy)
	}
}

func TestDataDirUsesExecutableDataFolderAndMigratesLegacyDB(t *testing.T) {
	oldExecutablePath := executablePath
	oldUserConfigDir := userConfigDir
	defer func() {
		executablePath = oldExecutablePath
		userConfigDir = oldUserConfigDir
	}()

	root := t.TempDir()
	exeDir := filepath.Join(root, "app")
	configDir := filepath.Join(root, "config")
	if err := os.MkdirAll(filepath.Join(configDir, "RestDeck"), 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	executablePath = func() (string, error) { return filepath.Join(exeDir, "RestDeck.exe"), nil }
	userConfigDir = func() (string, error) { return configDir, nil }
	dir, err := dataDir()
	if err != nil {
		t.Fatalf("data dir: %v", err)
	}
	if dir != filepath.Join(exeDir, "Data") {
		t.Fatalf("data dir = %q", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	legacy := filepath.Join(configDir, "RestDeck", "restdeck.db")
	if err := os.WriteFile(legacy, []byte("legacy-db"), 0o644); err != nil {
		t.Fatalf("write legacy: %v", err)
	}
	if err := migrateLegacyDatabase(dir); err != nil {
		t.Fatalf("migrate legacy: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "restdeck.db"))
	if err != nil {
		t.Fatalf("read migrated: %v", err)
	}
	if string(data) != "legacy-db" {
		t.Fatalf("migrated data = %q", data)
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

func TestStoreDeletesEmptyCollection(t *testing.T) {
	s, err := OpenInMemory(t.Context())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	collection := domain.Collection{ID: "c1", Name: "Empty", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.SaveCollection(t.Context(), collection); err != nil {
		t.Fatalf("save collection: %v", err)
	}
	if err := s.DeleteCollection(t.Context(), "c1"); err != nil {
		t.Fatalf("delete collection: %v", err)
	}

	state, err := s.State(t.Context())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if len(state.Collections) != 0 {
		t.Fatalf("collections length = %d", len(state.Collections))
	}
}

func TestStoreDeletesCollectionCascade(t *testing.T) {
	s, err := OpenInMemory(t.Context())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	collection := domain.Collection{ID: "c1", Name: "Demo", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.SaveCollection(t.Context(), collection); err != nil {
		t.Fatalf("save collection: %v", err)
	}
	if err := s.SaveFolder(t.Context(), domain.Folder{ID: "f1", CollectionID: "c1", Name: "Folder"}); err != nil {
		t.Fatalf("save folder: %v", err)
	}
	req := domain.Request{
		ID:           "r1",
		CollectionID: "c1",
		ParentID:     "f1",
		Name:         "Ping",
		Method:       "GET",
		URL:          "https://example.com",
		BodyMode:     domain.BodyModeNone,
		Auth:         domain.AuthConfig{Type: domain.AuthTypeNone, Values: map[string]string{}},
	}
	if err := s.SaveRequest(t.Context(), req); err != nil {
		t.Fatalf("save request: %v", err)
	}

	if err := s.DeleteCollection(t.Context(), "c1"); err != nil {
		t.Fatalf("delete collection: %v", err)
	}
	state, err := s.State(t.Context())
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	if len(state.Collections) != 0 {
		t.Fatalf("collections length = %d", len(state.Collections))
	}
}
