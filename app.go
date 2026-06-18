package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"restdeck/internal/domain"
	"restdeck/internal/realtime"
	reqsvc "restdeck/internal/request"
	"restdeck/internal/secrets"
	"restdeck/internal/store"
)

type App struct {
	ctx      context.Context
	store    *store.Store
	vault    *secrets.Vault
	sender   *reqsvc.Sender
	scripts  *reqsvc.ScriptRuntime
	runner   *reqsvc.Runner
	realtime *realtime.Service
}

func NewApp() *App {
	sender := reqsvc.NewSender()
	scripts := reqsvc.NewScriptRuntime()
	return &App{
		sender:   sender,
		scripts:  scripts,
		runner:   reqsvc.NewRunner(sender, scripts),
		realtime: realtime.NewService(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	vault, err := secrets.Open()
	if err == nil {
		a.vault = vault
	}
	s, err := store.Open(ctx)
	if err != nil {
		panic(err)
	}
	a.store = s
}

func (a *App) shutdown(ctx context.Context) {
	if a.store != nil {
		_ = a.store.Close()
	}
}

func (a *App) GetState() (domain.WorkspaceState, error) {
	return a.store.State(a.ctx)
}

func (a *App) CreateCollection(name string) (domain.WorkspaceState, error) {
	c := domain.Collection{
		ID:        uuid.NewString(),
		Name:      fallback(name, "New Collection"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := a.store.SaveCollection(a.ctx, c); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) SaveRequest(r domain.Request) (domain.WorkspaceState, error) {
	r = a.prepareSecrets(r, true)
	if err := a.store.SaveRequest(a.ctx, r); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) DeleteRequest(id string) (domain.WorkspaceState, error) {
	if err := a.store.DeleteRequest(a.ctx, id); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) SaveEnvironment(env domain.Environment) (domain.WorkspaceState, error) {
	env = a.prepareEnvironmentSecrets(env, true)
	if err := a.store.SaveEnvironment(a.ctx, env); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) SetActiveEnvironment(id string) (domain.WorkspaceState, error) {
	if err := a.store.SetActiveEnvironment(a.ctx, id); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) SaveGlobals(globals []domain.KeyValue) (domain.WorkspaceState, error) {
	if err := a.store.SaveGlobals(a.ctx, globals); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) SendRequest(r domain.Request, environmentID string, globals []domain.KeyValue) (domain.Response, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return domain.Response{}, err
	}
	env := findEnvironment(state.Environments, environmentID)
	r = a.prepareSecrets(r, false)
	env = a.prepareEnvironmentSecrets(env, false)
	variables := reqsvc.NewResolver(env, globals).Values()
	preResults := a.scripts.RunPreRequest(a.ctx, r.PreScript, r, variables)
	response, sendErr := a.sender.SendWithVariables(a.ctx, r, variables)
	tests := a.scripts.RunTests(a.ctx, r.TestScript, r, response, variables)
	if len(preResults) > 0 {
		tests = append(preResults, tests...)
	}
	response.TestResults = tests
	_ = a.store.AddHistory(a.ctx, domain.HistoryItem{
		RequestID:  r.ID,
		Name:       r.Name,
		Method:     r.Method,
		URL:        r.URL,
		StatusCode: response.StatusCode,
		DurationMs: response.DurationMs,
		Request:    r,
		Response:   response,
		CreatedAt:  time.Now(),
	})
	if sendErr != nil {
		return response, nil
	}
	return response, nil
}

func (a *App) ImportPostmanCollection(raw string) (domain.WorkspaceState, error) {
	collection, err := reqsvc.ImportPostman(raw)
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	if err := a.store.SaveCollection(a.ctx, collection); err != nil {
		return domain.WorkspaceState{}, err
	}
	for _, folder := range collection.Folders {
		if err := a.store.SaveFolder(a.ctx, folder); err != nil {
			return domain.WorkspaceState{}, err
		}
	}
	for _, request := range collection.Requests {
		if err := a.store.SaveRequest(a.ctx, request); err != nil {
			return domain.WorkspaceState{}, err
		}
	}
	return a.GetState()
}

func (a *App) ExportPostmanCollection(collectionID string) (string, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return "", err
	}
	for _, c := range state.Collections {
		if c.ID == collectionID {
			return reqsvc.ExportPostman(c)
		}
	}
	return "", fmt.Errorf("collection %s not found", collectionID)
}

func (a *App) RunCollection(collectionID, environmentID string, iterations int) (domain.RunnerResult, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return domain.RunnerResult{}, err
	}
	var collection domain.Collection
	found := false
	for _, c := range state.Collections {
		if c.ID == collectionID {
			collection = c
			found = true
			break
		}
	}
	if !found {
		return domain.RunnerResult{}, fmt.Errorf("collection %s not found", collectionID)
	}
	env := findEnvironment(state.Environments, environmentID)
	env = a.prepareEnvironmentSecrets(env, false)
	result := a.runner.Run(a.ctx, collection, env, state.Globals, iterations)
	if err := a.store.AddRunnerResult(a.ctx, result); err != nil {
		return domain.RunnerResult{}, err
	}
	return result, nil
}

func (a *App) TestWebSocket(input realtime.WebSocketRequest, environmentID string, globals []domain.KeyValue) realtime.WebSocketResult {
	env := findEnvironmentFromApp(a, environmentID)
	env = a.prepareEnvironmentSecrets(env, false)
	return a.realtime.TestWebSocket(a.ctx, input, env, globals)
}

func (a *App) TestSSE(input realtime.SSERequest, environmentID string, globals []domain.KeyValue) realtime.SSEResult {
	env := findEnvironmentFromApp(a, environmentID)
	env = a.prepareEnvironmentSecrets(env, false)
	return a.realtime.TestSSE(a.ctx, input, env, globals)
}

func (a *App) FormatBody(contentType, body string) string {
	return reqsvc.FormatBody(contentType, body)
}

func findEnvironmentFromApp(a *App, id string) domain.Environment {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return domain.Environment{}
	}
	return findEnvironment(state.Environments, id)
}

func (a *App) prepareSecrets(r domain.Request, seal bool) domain.Request {
	if a.vault == nil || r.Auth.Values == nil {
		return r
	}
	for _, key := range []string{"password", "token", "accessToken", "consumerSecret", "tokenSecret", "value"} {
		if value, ok := r.Auth.Values[key]; ok {
			if seal {
				if enc, err := a.vault.Seal(value); err == nil {
					r.Auth.Values[key] = enc
				}
			} else if dec, err := a.vault.Open(value); err == nil {
				r.Auth.Values[key] = dec
			}
		}
	}
	return r
}

func (a *App) prepareEnvironmentSecrets(env domain.Environment, seal bool) domain.Environment {
	if a.vault == nil {
		return env
	}
	for i, kv := range env.Variables {
		if !kv.Secret {
			continue
		}
		if seal {
			if enc, err := a.vault.Seal(kv.Value); err == nil {
				env.Variables[i].Value = enc
			}
		} else if dec, err := a.vault.Open(kv.Value); err == nil {
			env.Variables[i].Value = dec
		}
	}
	return env
}

func findEnvironment(environments []domain.Environment, id string) domain.Environment {
	for _, env := range environments {
		if env.ID == id {
			return env
		}
	}
	for _, env := range environments {
		if env.IsActive {
			return env
		}
	}
	return domain.Environment{}
}

func fallback(value, def string) string {
	if value == "" {
		return def
	}
	return value
}
