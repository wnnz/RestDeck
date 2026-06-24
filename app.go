package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

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

func (a *App) SaveCollection(c domain.Collection) (domain.WorkspaceState, error) {
	c.Name = fallback(strings.TrimSpace(c.Name), "Collection")
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

func (a *App) SelectFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择文件",
	})
}

func (a *App) DeleteRequest(id string) (domain.WorkspaceState, error) {
	if err := a.store.DeleteRequest(a.ctx, id); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) SaveTextFile(title, defaultFilename, content string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           fallback(title, "保存文件"),
		DefaultFilename: fallback(defaultFilename, "restdeck.txt"),
	})
	if err != nil || path == "" {
		return path, err
	}
	if filepath.Ext(path) == "" && filepath.Ext(defaultFilename) != "" {
		path += filepath.Ext(defaultFilename)
	}
	return path, os.WriteFile(path, []byte(content), 0o644)
}

func (a *App) DeleteCollection(collectionID string) (domain.WorkspaceState, error) {
	if err := a.store.DeleteCollection(a.ctx, collectionID); err != nil {
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

func (a *App) CreateEnvironment(name string) (domain.WorkspaceState, error) {
	env := domain.Environment{
		ID:        uuid.NewString(),
		Name:      fallback(strings.TrimSpace(name), "New Environment"),
		IsActive:  true,
		UpdatedAt: time.Now(),
		Variables: []domain.KeyValue{},
	}
	if err := a.store.SaveEnvironment(a.ctx, env); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) DeleteEnvironment(id string) (domain.WorkspaceState, error) {
	if err := a.store.DeleteEnvironment(a.ctx, id); err != nil {
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

func (a *App) SaveSettings(settings domain.Settings) (domain.WorkspaceState, error) {
	if err := a.store.SaveSettings(a.ctx, settings); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) DeleteCookie(cookie domain.Cookie) (domain.WorkspaceState, error) {
	if err := a.store.DeleteCookie(a.ctx, cookie); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) ClearCookies() (domain.WorkspaceState, error) {
	if err := a.store.ClearCookies(a.ctx); err != nil {
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
	a.sender.LoadCookies(state.Cookies)
	resolver := a.newResolver(env, globals, state.Settings.DefaultProxy)
	variables, err := resolver.ValuesWithError()
	if err != nil {
		return domain.Response{Error: err.Error()}, nil
	}
	preResults := a.scripts.RunPreRequest(a.ctx, r.PreScript, r, variables)
	response, sendErr := a.sender.SendWithVariablesAndProxy(a.ctx, r, variables, state.Settings.DefaultProxy)
	if len(response.Cookies) > 0 {
		_ = a.store.SaveCookies(a.ctx, response.Cookies)
	}
	if historyCookies := a.sender.CookiesForURL(response.RequestedURL); len(historyCookies) > 0 {
		_ = a.store.SaveCookies(a.ctx, historyCookies)
	}
	tests := a.scripts.RunTests(a.ctx, r.TestScript, r, response, variables)
	if len(preResults) > 0 {
		tests = append(preResults, tests...)
	}
	response.TestResults = tests
	historyURL := strings.TrimSpace(response.RequestedURL)
	if historyURL == "" {
		historyURL = r.URL
	}
	_ = a.store.AddHistory(a.ctx, domain.HistoryItem{
		RequestID:  r.ID,
		Name:       r.Name,
		Method:     r.Method,
		URL:        historyURL,
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

func (a *App) PreviewRequest(r domain.Request, environmentID string, globals []domain.KeyValue) (domain.PreparedRequest, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return domain.PreparedRequest{}, err
	}
	env := findEnvironment(state.Environments, environmentID)
	r = a.prepareSecrets(r, false)
	env = a.prepareEnvironmentSecrets(env, false)
	a.sender.LoadCookies(state.Cookies)
	resolver := a.newResolver(env, globals, state.Settings.DefaultProxy)
	variables, err := resolver.ValuesWithError()
	if err != nil {
		return domain.PreparedRequest{VariableErrors: []string{err.Error()}, Error: err.Error()}, nil
	}
	preview, err := a.sender.PrepareRequest(a.ctx, r, variables, state.Settings.DefaultProxy)
	if err != nil {
		return preview, nil
	}
	return preview, nil
}

func (a *App) DebugVariables(environmentID string, globals []domain.KeyValue) (domain.VariableDebugReport, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return domain.VariableDebugReport{}, err
	}
	env := findEnvironment(state.Environments, environmentID)
	env = a.prepareEnvironmentSecrets(env, false)
	report := domain.VariableDebugReport{}
	addVariablesToDebugReport(&report, "global", globals, a.newResolver(env, globals, state.Settings.DefaultProxy))
	addVariablesToDebugReport(&report, "environment", env.Variables, a.newResolver(env, globals, state.Settings.DefaultProxy))
	return report, nil
}

func (a *App) DebugRequestVariables(r domain.Request, environmentID string, globals []domain.KeyValue) (domain.VariableDebugReport, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return domain.VariableDebugReport{}, err
	}
	env := findEnvironment(state.Environments, environmentID)
	env = a.prepareEnvironmentSecrets(env, false)
	resolver := a.newResolver(env, globals, state.Settings.DefaultProxy)
	report := domain.VariableDebugReport{}
	for _, name := range reqsvc.RequestVariableNames(r) {
		item := domain.VariableDebugItem{Name: name, Source: "request", Type: "reference", Raw: "{{" + name + "}}"}
		value, err := resolver.ResolveVariableForDebug(name)
		if err != nil {
			item.Error = err.Error()
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %s", name, err.Error()))
		} else {
			item.Value = value
			item.Resolved = true
		}
		report.Variables = append(report.Variables, item)
	}
	addVariablesToDebugReport(&report, "global", globals, resolver)
	addVariablesToDebugReport(&report, "environment", env.Variables, resolver)
	return report, nil
}

func (a *App) TestVariable(variable domain.KeyValue, environmentID string, globals []domain.KeyValue) (string, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return "", err
	}
	env := findEnvironment(state.Environments, environmentID)
	env = a.prepareEnvironmentSecrets(env, false)
	variable.Enabled = true
	if strings.TrimSpace(variable.Key) == "" {
		variable.Key = "__preview"
	}
	env.Variables = append(env.Variables, variable)
	return a.newResolver(env, globals, state.Settings.DefaultProxy).ResolveVariableForDebug(variable.Key)
}

func (a *App) SaveRunnerResult(result domain.RunnerResult) (domain.WorkspaceState, error) {
	if err := a.store.AddRunnerResult(a.ctx, result); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
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

func (a *App) ImportOpenAPICollection(raw string) (domain.WorkspaceState, error) {
	return a.ImportOpenAPICollectionWithOptions(raw, domain.OpenAPIImportOptions{})
}

func (a *App) ImportOpenAPICollectionWithOptions(raw string, options domain.OpenAPIImportOptions) (domain.WorkspaceState, error) {
	collection, err := reqsvc.ImportOpenAPIWithOptions(raw, options)
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	if err := a.store.SaveCollection(a.ctx, collection); err != nil {
		return domain.WorkspaceState{}, err
	}
	for _, request := range collection.Requests {
		if err := a.store.SaveRequest(a.ctx, request); err != nil {
			return domain.WorkspaceState{}, err
		}
	}
	return a.GetState()
}

func (a *App) InspectOpenAPI(raw string) (domain.OpenAPIInfo, error) {
	return reqsvc.InspectOpenAPI(raw)
}

func (a *App) ImportFetchRequest(rawFetch, collectionID string) (domain.WorkspaceState, error) {
	collection, err := a.ensureImportCollection(collectionID)
	if err != nil {
		return domain.WorkspaceState{}, err
	}

	request, err := reqsvc.ImportFetch(rawFetch, collection.ID, len(collection.Requests))
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	if err := a.store.SaveRequest(a.ctx, request); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) ImportCurlRequest(rawCurl, collectionID string) (domain.WorkspaceState, error) {
	collection, err := a.ensureImportCollection(collectionID)
	if err != nil {
		return domain.WorkspaceState{}, err
	}

	request, err := reqsvc.ImportCurl(rawCurl, collection.ID, len(collection.Requests))
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	if err := a.store.SaveRequest(a.ctx, request); err != nil {
		return domain.WorkspaceState{}, err
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

func (a *App) ExportOpenAPICollection(collectionID string) (string, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return "", err
	}
	for _, c := range state.Collections {
		if c.ID == collectionID {
			return reqsvc.ExportOpenAPI(c)
		}
	}
	return "", fmt.Errorf("collection %s not found", collectionID)
}

func (a *App) ExportHARCollection(collectionID string) (string, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return "", err
	}
	for _, c := range state.Collections {
		if c.ID == collectionID {
			return reqsvc.ExportHAR(c)
		}
	}
	return "", fmt.Errorf("collection %s not found", collectionID)
}

func (a *App) ImportHARCollection(raw string) (domain.WorkspaceState, error) {
	collection, err := reqsvc.ImportHAR(raw)
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	if err := a.store.SaveCollection(a.ctx, collection); err != nil {
		return domain.WorkspaceState{}, err
	}
	for _, request := range collection.Requests {
		if err := a.store.SaveRequest(a.ctx, request); err != nil {
			return domain.WorkspaceState{}, err
		}
	}
	return a.GetState()
}

func (a *App) ExportPostmanRequest(r domain.Request, collectionName string) (string, error) {
	r = a.prepareSecrets(r, false)
	r.ParentID = ""
	return reqsvc.ExportPostman(domain.Collection{
		ID:       r.CollectionID,
		Name:     fallback(r.Name, collectionName),
		Requests: []domain.Request{r},
	})
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
	result := a.runner.RunWithProxy(a.ctx, collection, env, state.Globals, iterations, state.Settings.DefaultProxy)
	if err := a.store.AddRunnerResult(a.ctx, result); err != nil {
		return domain.RunnerResult{}, err
	}
	return result, nil
}

func (a *App) TestWebSocket(input realtime.WebSocketRequest, environmentID string, globals []domain.KeyValue) realtime.WebSocketResult {
	env := findEnvironmentFromApp(a, environmentID)
	env = a.prepareEnvironmentSecrets(env, false)
	settings, _ := a.store.GetSettings(a.ctx)
	return a.realtime.TestWebSocket(a.ctx, input, env, globals, settings.DefaultProxy)
}

func (a *App) TestSSE(input realtime.SSERequest, environmentID string, globals []domain.KeyValue) realtime.SSEResult {
	env := findEnvironmentFromApp(a, environmentID)
	env = a.prepareEnvironmentSecrets(env, false)
	settings, _ := a.store.GetSettings(a.ctx)
	return a.realtime.TestSSE(a.ctx, input, env, globals, settings.DefaultProxy)
}

func (a *App) FormatBody(contentType, body string) string {
	return reqsvc.FormatBody(contentType, body)
}

func (a *App) QueryJSONPath(body, path string) (string, error) {
	value, ok, err := reqsvc.ExtractJSONPath(body, path)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("JSONPath %s 未找到", path)
	}
	return value, nil
}

func (a *App) CreateResponseVariable(environmentID, key, requestID, jsonPath, fallbackValue string) (domain.WorkspaceState, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return domain.WorkspaceState{}, err
	}
	env := findEnvironment(state.Environments, environmentID)
	if env.ID == "" {
		return domain.WorkspaceState{}, fmt.Errorf("environment %s not found", environmentID)
	}
	env = a.prepareEnvironmentSecrets(env, false)
	env.Variables = append(env.Variables, domain.KeyValue{
		ID:               uuid.NewString(),
		Enabled:          true,
		Key:              fallback(strings.TrimSpace(key), "responseValue"),
		ValueType:        "responseJsonPath",
		SourceRequestID:  requestID,
		JSONPath:         fallback(strings.TrimSpace(jsonPath), "$."),
		ResponseStrategy: "latestHistory",
		FallbackValue:    fallbackValue,
	})
	env = a.prepareEnvironmentSecrets(env, true)
	if err := a.store.SaveEnvironment(a.ctx, env); err != nil {
		return domain.WorkspaceState{}, err
	}
	return a.GetState()
}

func (a *App) ensureImportCollection(collectionID string) (domain.Collection, error) {
	state, err := a.store.State(a.ctx)
	if err != nil {
		return domain.Collection{}, err
	}

	collection, found := findCollection(state.Collections, collectionID)
	if !found && collectionID == "" && len(state.Collections) > 0 {
		collection = state.Collections[0]
		found = true
	}
	if found {
		return collection, nil
	}

	now := time.Now()
	collection = domain.Collection{
		ID:        uuid.NewString(),
		Name:      "Imported Requests",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := a.store.SaveCollection(a.ctx, collection); err != nil {
		return domain.Collection{}, err
	}
	return collection, nil
}

func (a *App) newResolver(env domain.Environment, globals []domain.KeyValue, defaultProxy domain.ProxyConfig) *reqsvc.Resolver {
	return reqsvc.NewResolverWithOptions(env, globals, reqsvc.ResolverOptions{
		Context: a.ctx,
		HistoryLookup: func(ctx context.Context, requestID string) (domain.HistoryItem, bool, error) {
			return a.store.LatestHistoryForRequest(ctx, requestID)
		},
		ResponseRefresh: func(ctx context.Context, requestID string, variables map[string]string) (domain.HistoryItem, error) {
			return a.refreshResponseVariable(ctx, requestID, variables, defaultProxy)
		},
	})
}

func addVariablesToDebugReport(report *domain.VariableDebugReport, source string, variables []domain.KeyValue, resolver *reqsvc.Resolver) {
	for _, kv := range variables {
		if !kv.Enabled || strings.TrimSpace(kv.Key) == "" {
			continue
		}
		item := domain.VariableDebugItem{
			Name:   kv.Key,
			Source: source,
			Type:   fallback(kv.ValueType, "static"),
			Raw:    kv.Value,
		}
		value, err := resolver.ResolveVariableForDebug(kv.Key)
		if err != nil {
			item.Error = err.Error()
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %s", kv.Key, err.Error()))
		} else {
			item.Value = value
			item.Resolved = true
		}
		report.Variables = append(report.Variables, item)
	}
}

func (a *App) refreshResponseVariable(ctx context.Context, requestID string, variables map[string]string, defaultProxy domain.ProxyConfig) (domain.HistoryItem, error) {
	state, err := a.store.State(ctx)
	if err != nil {
		return domain.HistoryItem{}, err
	}
	req, ok := findRequest(state.Collections, requestID)
	if !ok {
		return domain.HistoryItem{}, fmt.Errorf("request %s not found", requestID)
	}
	req = a.prepareSecrets(req, false)
	response, _ := a.sender.SendWithVariablesAndProxy(ctx, req, variables, defaultProxy)
	historyURL := strings.TrimSpace(response.RequestedURL)
	if historyURL == "" {
		historyURL = req.URL
	}
	item := domain.HistoryItem{
		RequestID:  req.ID,
		Name:       req.Name,
		Method:     req.Method,
		URL:        historyURL,
		StatusCode: response.StatusCode,
		DurationMs: response.DurationMs,
		Request:    req,
		Response:   response,
		CreatedAt:  time.Now(),
	}
	if err := a.store.AddHistory(ctx, item); err != nil {
		return domain.HistoryItem{}, err
	}
	return item, nil
}

func findCollection(collections []domain.Collection, id string) (domain.Collection, bool) {
	for _, collection := range collections {
		if collection.ID == id {
			return collection, true
		}
	}
	return domain.Collection{}, false
}

func findRequest(collections []domain.Collection, id string) (domain.Request, bool) {
	for _, collection := range collections {
		for _, request := range collection.Requests {
			if request.ID == id {
				return request, true
			}
		}
	}
	return domain.Request{}, false
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
