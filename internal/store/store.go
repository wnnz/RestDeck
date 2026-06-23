package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"restdeck/internal/domain"
)

type Store struct {
	db *sql.DB
}

var executablePath = os.Executable

func Open(ctx context.Context) (*Store, error) {
	dir, err := dataDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", filepath.Join(dir, "restdeck.db"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.seed(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func OpenInMemory(ctx context.Context) (*Store, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func dataDir() (string, error) {
	exe, err := executablePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "Data"), nil
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS collections (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS folders (
			id TEXT PRIMARY KEY,
			collection_id TEXT NOT NULL,
			parent_id TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(collection_id) REFERENCES collections(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS requests (
			id TEXT PRIMARY KEY,
			collection_id TEXT NOT NULL,
			parent_id TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			method TEXT NOT NULL,
			url TEXT NOT NULL,
			params_json TEXT NOT NULL,
				headers_json TEXT NOT NULL,
				body_mode TEXT NOT NULL,
				body TEXT NOT NULL,
				form_items_json TEXT NOT NULL DEFAULT '[]',
				auth_json TEXT NOT NULL,
				proxy_json TEXT NOT NULL DEFAULT '{}',
				pre_script TEXT NOT NULL,
				test_script TEXT NOT NULL,
			timeout_ms INTEGER NOT NULL DEFAULT 30000,
			sort_order INTEGER NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(collection_id) REFERENCES collections(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS environments (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			variables_json TEXT NOT NULL,
			is_active INTEGER NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS globals (
			id TEXT PRIMARY KEY,
			enabled INTEGER NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			description TEXT NOT NULL,
			secret INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS history (
			id TEXT PRIMARY KEY,
			request_id TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			method TEXT NOT NULL,
			url TEXT NOT NULL,
			status_code INTEGER NOT NULL DEFAULT 0,
			duration_ms INTEGER NOT NULL DEFAULT 0,
			request_json TEXT NOT NULL,
			response_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS runner_results (
			id TEXT PRIMARY KEY,
			collection_id TEXT NOT NULL,
			environment_id TEXT NOT NULL,
			name TEXT NOT NULL,
			iterations INTEGER NOT NULL,
			passed INTEGER NOT NULL,
			failed INTEGER NOT NULL,
			duration_ms INTEGER NOT NULL,
			items_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	if err := s.ensureColumn(ctx, "requests", "form_items_json", `TEXT NOT NULL DEFAULT '[]'`); err != nil {
		return err
	}
	if err := s.ensureColumn(ctx, "requests", "proxy_json", `TEXT NOT NULL DEFAULT '{}'`); err != nil {
		return err
	}
	return nil
}

func (s *Store) ensureColumn(ctx context.Context, table, column, definition string) error {
	rows, err := s.db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue interface{}
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `ALTER TABLE `+table+` ADD COLUMN `+column+` `+definition)
	return err
}

func (s *Store) seed(ctx context.Context) error {
	var collectionCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM collections`).Scan(&collectionCount); err != nil {
		return err
	}
	now := time.Now()
	if collectionCount == 0 {
		collection := domain.Collection{
			ID:          uuid.NewString(),
			Name:        "Getting Started",
			Description: "A small local collection for trying RestDeck.",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		request := domain.Request{
			ID:           uuid.NewString(),
			CollectionID: collection.ID,
			Name:         "GET httpbin anything",
			Method:       "GET",
			URL:          "https://httpbin.org/anything",
			Params:       []domain.KeyValue{{ID: uuid.NewString(), Enabled: true, Key: "source", Value: "restdeck"}},
			Headers:      []domain.KeyValue{{ID: uuid.NewString(), Enabled: true, Key: "Accept", Value: "application/json"}},
			BodyMode:     domain.BodyModeNone,
			Auth:         domain.AuthConfig{Type: domain.AuthTypeNone, Values: map[string]string{}},
			Proxy:        domain.ProxyConfig{Mode: "inherit"},
			TimeoutMs:    30000,
			UpdatedAt:    now,
		}
		if err := s.SaveCollection(ctx, collection); err != nil {
			return err
		}
		if err := s.SaveRequest(ctx, request); err != nil {
			return err
		}
	}

	var environmentCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM environments`).Scan(&environmentCount); err != nil {
		return err
	}
	if environmentCount == 0 {
		env := domain.Environment{
			ID:        uuid.NewString(),
			Name:      "Local",
			IsActive:  true,
			UpdatedAt: now,
			Variables: []domain.KeyValue{
				{ID: uuid.NewString(), Enabled: true, Key: "baseUrl", Value: "https://httpbin.org"},
				{ID: uuid.NewString(), Enabled: true, Key: "token", Value: "", Secret: true},
			},
		}
		if err := s.SaveEnvironment(ctx, env); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) State(ctx context.Context) (domain.WorkspaceState, error) {
	collections, err := s.ListCollections(ctx)
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	environments, err := s.ListEnvironments(ctx)
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	history, err := s.ListHistory(ctx, 80)
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	globals, err := s.ListGlobals(ctx)
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	active := ""
	for _, env := range environments {
		if env.IsActive {
			active = env.ID
			break
		}
	}
	return domain.WorkspaceState{
		Collections:         collections,
		Environments:        environments,
		History:             history,
		Globals:             globals,
		ActiveEnvironmentID: active,
		Settings:            settings,
	}, nil
}

func (s *Store) ListCollections(ctx context.Context) ([]domain.Collection, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description, created_at, updated_at FROM collections ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collections := []domain.Collection{}
	for rows.Next() {
		var c domain.Collection
		var createdAt, updatedAt string
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		c.CreatedAt = parseTime(createdAt)
		c.UpdatedAt = parseTime(updatedAt)
		collections = append(collections, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range collections {
		folders, err := s.listFolders(ctx, collections[i].ID)
		if err != nil {
			return nil, err
		}
		requests, err := s.listRequests(ctx, collections[i].ID)
		if err != nil {
			return nil, err
		}
		collections[i].Folders = folders
		collections[i].Requests = requests
	}
	return collections, nil
}

func (s *Store) SaveCollection(ctx context.Context, c domain.Collection) error {
	now := time.Now()
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	_, err := s.db.ExecContext(ctx, `INSERT INTO collections (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, description = excluded.description, updated_at = excluded.updated_at`,
		c.ID, c.Name, c.Description, formatTime(c.CreatedAt), formatTime(c.UpdatedAt))
	return err
}

func (s *Store) SaveFolder(ctx context.Context, f domain.Folder) error {
	if f.ID == "" {
		f.ID = uuid.NewString()
	}
	f.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx, `INSERT INTO folders (id, collection_id, parent_id, name, sort_order, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET collection_id = excluded.collection_id, parent_id = excluded.parent_id,
		name = excluded.name, sort_order = excluded.sort_order, updated_at = excluded.updated_at`,
		f.ID, f.CollectionID, f.ParentID, f.Name, f.SortOrder, formatTime(f.UpdatedAt))
	return err
}

func (s *Store) SaveRequest(ctx context.Context, r domain.Request) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	if r.Method == "" {
		r.Method = "GET"
	}
	if r.TimeoutMs <= 0 {
		r.TimeoutMs = 30000
	}
	if r.Auth.Type == "" {
		r.Auth = domain.AuthConfig{Type: domain.AuthTypeNone, Values: map[string]string{}}
	}
	r.Proxy = normalizeProxy(r.Proxy, "inherit")
	if r.BodyMode == domain.BodyModeForm {
		r.FormItems = normalizeFormItems(r.FormItems, r.Body)
		r.Body = formItemsToBody(r.FormItems)
	} else {
		r.FormItems = normalizeFormItems(r.FormItems, "")
	}
	r.UpdatedAt = time.Now()
	paramsJSON, err := marshal(r.Params)
	if err != nil {
		return err
	}
	headersJSON, err := marshal(r.Headers)
	if err != nil {
		return err
	}
	authJSON, err := marshal(r.Auth)
	if err != nil {
		return err
	}
	proxyJSON, err := marshal(r.Proxy)
	if err != nil {
		return err
	}
	formItemsJSON, err := marshal(r.FormItems)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO requests
		(id, collection_id, parent_id, name, method, url, params_json, headers_json, body_mode, body, form_items_json, auth_json,
		 proxy_json, pre_script, test_script, timeout_ms, sort_order, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET collection_id = excluded.collection_id, parent_id = excluded.parent_id,
		name = excluded.name, method = excluded.method, url = excluded.url, params_json = excluded.params_json,
		headers_json = excluded.headers_json, body_mode = excluded.body_mode, body = excluded.body,
		form_items_json = excluded.form_items_json,
		auth_json = excluded.auth_json, proxy_json = excluded.proxy_json, pre_script = excluded.pre_script, test_script = excluded.test_script,
		timeout_ms = excluded.timeout_ms, sort_order = excluded.sort_order, updated_at = excluded.updated_at`,
		r.ID, r.CollectionID, r.ParentID, r.Name, r.Method, r.URL, paramsJSON, headersJSON,
		string(r.BodyMode), r.Body, formItemsJSON, authJSON, proxyJSON, r.PreScript, r.TestScript, r.TimeoutMs, r.SortOrder, formatTime(r.UpdatedAt))
	return err
}

func (s *Store) DeleteRequest(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM requests WHERE id = ?`, id)
	return err
}

func (s *Store) DeleteCollection(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM collections WHERE id = ?`, id)
	return err
}

func (s *Store) SaveEnvironment(ctx context.Context, env domain.Environment) error {
	if env.ID == "" {
		env.ID = uuid.NewString()
	}
	env.UpdatedAt = time.Now()
	varsJSON, err := marshal(env.Variables)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if env.IsActive {
		if _, err := tx.ExecContext(ctx, `UPDATE environments SET is_active = 0`); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO environments (id, name, variables_json, is_active, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, variables_json = excluded.variables_json,
		is_active = excluded.is_active, updated_at = excluded.updated_at`,
		env.ID, env.Name, varsJSON, boolInt(env.IsActive), formatTime(env.UpdatedAt))
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) DeleteEnvironment(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	var total int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM environments`).Scan(&total); err != nil {
		return err
	}
	if total <= 1 {
		return fmt.Errorf("cannot delete the last environment")
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM environments WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("environment %s not found", id)
	}
	var activeCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM environments WHERE is_active = 1`).Scan(&activeCount); err != nil {
		return err
	}
	if activeCount == 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE environments SET is_active = 1 WHERE id = (SELECT id FROM environments ORDER BY name LIMIT 1)`); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) SetActiveEnvironment(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if _, err := tx.ExecContext(ctx, `UPDATE environments SET is_active = 0`); err != nil {
		return err
	}
	if id != "" {
		res, err := tx.ExecContext(ctx, `UPDATE environments SET is_active = 1 WHERE id = ?`, id)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return fmt.Errorf("environment %s not found", id)
		}
	}
	return tx.Commit()
}

func (s *Store) ListEnvironments(ctx context.Context) ([]domain.Environment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, variables_json, is_active, updated_at FROM environments ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	envs := []domain.Environment{}
	for rows.Next() {
		var env domain.Environment
		var variablesJSON, updatedAt string
		var active int
		if err := rows.Scan(&env.ID, &env.Name, &variablesJSON, &active, &updatedAt); err != nil {
			return nil, err
		}
		if err := unmarshal(variablesJSON, &env.Variables); err != nil {
			return nil, err
		}
		env.IsActive = active == 1
		env.UpdatedAt = parseTime(updatedAt)
		envs = append(envs, env)
	}
	return envs, rows.Err()
}

func (s *Store) ListGlobals(ctx context.Context) ([]domain.KeyValue, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, enabled, key, value, description, secret FROM globals ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.KeyValue{}
	for rows.Next() {
		var kv domain.KeyValue
		var enabled, secret int
		if err := rows.Scan(&kv.ID, &enabled, &kv.Key, &kv.Value, &kv.Description, &secret); err != nil {
			return nil, err
		}
		kv.Enabled = enabled == 1
		kv.Secret = secret == 1
		items = append(items, kv)
	}
	return items, rows.Err()
}

func (s *Store) GetSettings(ctx context.Context) (domain.Settings, error) {
	settings := defaultSettings()
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return domain.Settings{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return domain.Settings{}, err
		}
		switch key {
		case "language":
			settings.Language = value
		case "theme":
			settings.Theme = value
		case "defaultProxy":
			_ = json.Unmarshal([]byte(value), &settings.DefaultProxy)
		}
	}
	settings.DefaultProxy = normalizeProxy(settings.DefaultProxy, "none")
	return settings, rows.Err()
}

func (s *Store) SaveSettings(ctx context.Context, settings domain.Settings) error {
	if strings.TrimSpace(settings.Language) == "" {
		settings.Language = "zh-CN"
	}
	if settings.Theme != "dark" {
		settings.Theme = "light"
	}
	settings.DefaultProxy = normalizeProxy(settings.DefaultProxy, "none")
	proxyJSON, err := marshal(settings.DefaultProxy)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	for key, value := range map[string]string{
		"language":     settings.Language,
		"theme":        settings.Theme,
		"defaultProxy": proxyJSON,
	} {
		if _, err := tx.ExecContext(ctx, `INSERT INTO settings (key, value) VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) LatestHistoryForRequest(ctx context.Context, requestID string) (domain.HistoryItem, bool, error) {
	var item domain.HistoryItem
	var requestJSON, responseJSON, createdAt string
	err := s.db.QueryRowContext(ctx, `SELECT id, request_id, name, method, url, status_code, duration_ms,
		request_json, response_json, created_at FROM history WHERE request_id = ? ORDER BY created_at DESC LIMIT 1`, requestID).
		Scan(&item.ID, &item.RequestID, &item.Name, &item.Method, &item.URL, &item.StatusCode, &item.DurationMs, &requestJSON, &responseJSON, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.HistoryItem{}, false, nil
	}
	if err != nil {
		return domain.HistoryItem{}, false, err
	}
	if err := unmarshal(requestJSON, &item.Request); err != nil {
		return domain.HistoryItem{}, false, err
	}
	if err := unmarshal(responseJSON, &item.Response); err != nil {
		return domain.HistoryItem{}, false, err
	}
	item.CreatedAt = parseTime(createdAt)
	return item, true, nil
}

func (s *Store) SaveGlobals(ctx context.Context, globals []domain.KeyValue) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if _, err := tx.ExecContext(ctx, `DELETE FROM globals`); err != nil {
		return err
	}
	for _, kv := range globals {
		if kv.ID == "" {
			kv.ID = uuid.NewString()
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO globals (id, enabled, key, value, description, secret)
			VALUES (?, ?, ?, ?, ?, ?)`, kv.ID, boolInt(kv.Enabled), kv.Key, kv.Value, kv.Description, boolInt(kv.Secret)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) AddHistory(ctx context.Context, item domain.HistoryItem) error {
	if item.ID == "" {
		item.ID = uuid.NewString()
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	reqJSON, err := marshal(item.Request)
	if err != nil {
		return err
	}
	resJSON, err := marshal(item.Response)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO history
		(id, request_id, name, method, url, status_code, duration_ms, request_json, response_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.RequestID, item.Name, item.Method, item.URL, item.StatusCode, item.DurationMs,
		reqJSON, resJSON, formatTime(item.CreatedAt))
	return err
}

func (s *Store) ListHistory(ctx context.Context, limit int) ([]domain.HistoryItem, error) {
	if limit <= 0 {
		limit = 80
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, request_id, name, method, url, status_code, duration_ms,
		request_json, response_json, created_at FROM history ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.HistoryItem{}
	for rows.Next() {
		var item domain.HistoryItem
		var requestJSON, responseJSON, createdAt string
		if err := rows.Scan(&item.ID, &item.RequestID, &item.Name, &item.Method, &item.URL, &item.StatusCode,
			&item.DurationMs, &requestJSON, &responseJSON, &createdAt); err != nil {
			return nil, err
		}
		if err := unmarshal(requestJSON, &item.Request); err != nil {
			return nil, err
		}
		if err := unmarshal(responseJSON, &item.Response); err != nil {
			return nil, err
		}
		item.CreatedAt = parseTime(createdAt)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) AddRunnerResult(ctx context.Context, result domain.RunnerResult) error {
	if result.ID == "" {
		result.ID = uuid.NewString()
	}
	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}
	itemsJSON, err := marshal(result.Items)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO runner_results
		(id, collection_id, environment_id, name, iterations, passed, failed, duration_ms, items_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.ID, result.CollectionID, result.EnvironmentID, result.Name, result.Iterations, result.Passed,
		result.Failed, result.DurationMs, itemsJSON, formatTime(result.CreatedAt))
	return err
}

func (s *Store) listFolders(ctx context.Context, collectionID string) ([]domain.Folder, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, collection_id, parent_id, name, sort_order, updated_at
		FROM folders WHERE collection_id = ? ORDER BY sort_order, name`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	folders := []domain.Folder{}
	for rows.Next() {
		var f domain.Folder
		var updatedAt string
		if err := rows.Scan(&f.ID, &f.CollectionID, &f.ParentID, &f.Name, &f.SortOrder, &updatedAt); err != nil {
			return nil, err
		}
		f.UpdatedAt = parseTime(updatedAt)
		folders = append(folders, f)
	}
	return folders, rows.Err()
}

func (s *Store) listRequests(ctx context.Context, collectionID string) ([]domain.Request, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, collection_id, parent_id, name, method, url,
		params_json, headers_json, body_mode, body, form_items_json, auth_json, proxy_json, pre_script, test_script, timeout_ms,
		sort_order, updated_at FROM requests WHERE collection_id = ? ORDER BY sort_order, name`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := []domain.Request{}
	for rows.Next() {
		var r domain.Request
		var paramsJSON, headersJSON, formItemsJSON, authJSON, proxyJSON, bodyMode, updatedAt string
		if err := rows.Scan(&r.ID, &r.CollectionID, &r.ParentID, &r.Name, &r.Method, &r.URL,
			&paramsJSON, &headersJSON, &bodyMode, &r.Body, &formItemsJSON, &authJSON, &proxyJSON, &r.PreScript, &r.TestScript,
			&r.TimeoutMs, &r.SortOrder, &updatedAt); err != nil {
			return nil, err
		}
		if err := unmarshal(paramsJSON, &r.Params); err != nil {
			return nil, err
		}
		if err := unmarshal(headersJSON, &r.Headers); err != nil {
			return nil, err
		}
		if err := unmarshal(authJSON, &r.Auth); err != nil {
			return nil, err
		}
		if err := unmarshal(proxyJSON, &r.Proxy); err != nil {
			return nil, err
		}
		r.Proxy = normalizeProxy(r.Proxy, "inherit")
		if err := unmarshal(formItemsJSON, &r.FormItems); err != nil {
			return nil, err
		}
		r.BodyMode = domain.BodyMode(bodyMode)
		if r.BodyMode == domain.BodyModeForm {
			r.FormItems = normalizeFormItems(r.FormItems, r.Body)
			r.Body = formItemsToBody(r.FormItems)
		} else {
			r.FormItems = normalizeFormItems(r.FormItems, "")
		}
		r.UpdatedAt = parseTime(updatedAt)
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

func marshal(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshal(raw string, v interface{}) error {
	if raw == "" {
		raw = "null"
	}
	return json.Unmarshal([]byte(raw), v)
}

func normalizeFormItems(items []domain.FormItem, fallbackBody string) []domain.FormItem {
	if len(items) == 0 && fallbackBody != "" {
		items = formItemsFromBody(fallbackBody)
	}
	out := []domain.FormItem{}
	for _, item := range items {
		if item.ID == "" {
			item.ID = uuid.NewString()
		}
		if item.Type != "file" {
			item.Type = "text"
		}
		if item.Key == "" && item.Value == "" && item.FilePath == "" && item.Description == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func formItemsFromBody(raw string) []domain.FormItem {
	items := []domain.FormItem{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, _ := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		item := domain.FormItem{
			ID:      uuid.NewString(),
			Enabled: true,
			Key:     key,
			Type:    "text",
			Value:   value,
		}
		if strings.HasPrefix(value, "@") {
			item.Type = "file"
			item.Value = ""
			item.FilePath = strings.TrimPrefix(value, "@")
		}
		items = append(items, item)
	}
	return items
}

func formItemsToBody(items []domain.FormItem) string {
	lines := []string{}
	for _, item := range items {
		if item.Key == "" {
			continue
		}
		value := item.Value
		if item.Type == "file" {
			value = "@" + item.FilePath
		}
		lines = append(lines, item.Key+"="+value)
	}
	return strings.Join(lines, "\n")
}

func defaultSettings() domain.Settings {
	return domain.Settings{
		Language:     "zh-CN",
		Theme:        "light",
		DefaultProxy: domain.ProxyConfig{Mode: "none"},
	}
}

func normalizeProxy(proxy domain.ProxyConfig, fallbackMode string) domain.ProxyConfig {
	proxy.Mode = strings.TrimSpace(proxy.Mode)
	proxy.URL = strings.TrimSpace(proxy.URL)
	proxy.NoProxy = normalizeNoProxy(proxy.NoProxy)
	switch proxy.Mode {
	case "inherit", "none", "custom":
	default:
		proxy.Mode = fallbackMode
	}
	if proxy.Mode == "" {
		proxy.Mode = fallbackMode
	}
	if proxy.Mode != "custom" {
		proxy.URL = ""
		proxy.NoProxy = ""
	}
	return proxy
}

func normalizeNoProxy(raw string) string {
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
	return strings.Join(out, ",")
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(raw string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, raw)
	return t
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func rollback(tx *sql.Tx) {
	if tx != nil {
		_ = tx.Rollback()
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
