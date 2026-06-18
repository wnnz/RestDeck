# RestDeck

RestDeck is a local-first, open-source API testing desktop app inspired by Postman. It has no account login, no cloud sync, and no team workspace dependency. The first version focuses on honest, implemented UI: if a capability is not implemented, it does not appear as an app navigation item or placeholder panel.

## Tech Stack

- Go + Wails v2
- Vue 3 + Vite
- Tailwind CSS v4
- PrimeVue-compatible unstyled/Volt-style code-owned UI
- SQLite for local workspace data
- goja for the supported Postman `pm` script subset

## Implemented in This Version

- Collections, requests, environments, globals, history, runner, and local settings screens.
- HTTP/REST request editing with method, URL, query params, headers, body, auth, pre-request script, tests, and timeout.
- Auth types: No Auth, API Key, Bearer Token, Basic, Digest header placeholder, OAuth 1 signature, and OAuth 2 bearer token.
- Response viewer with body, headers, cookies, test results, status, duration, and size.
- Local SQLite persistence under the OS user config directory.
- Local encrypted wrapping for secret environment values and sensitive auth fields.
- Postman Collection JSON import/export for common collection, folder, request, header, body, auth, and script fields.
- Dynamic variables such as `{{$guid}}`, `{{$timestamp}}`, `{{$isoTimestamp}}`, `{{$randomInt}}`, `{{$randomBoolean}}`, and `{{$randomEmail}}`.
- Common `pm` script subset: `pm.variables.get/set/replaceIn`, `pm.request`, `pm.response`, `pm.test`, and basic `expect(...)` assertions.
- WebSocket one-message debugging and SSE event-stream collection.

## Not Implemented Yet

These capabilities are intentionally absent from the app UI until they exist:

- gRPC
- Mock servers
- Monitors
- Flows
- Team/cloud workspaces
- AI features
- Full Postman sandbox compatibility
- Full Digest challenge-response handling
- OpenAPI import/export
- Request documentation side panel

See [docs/FEATURE_MATRIX.md](docs/FEATURE_MATRIX.md) for the current status.

## Development

Install Wails v2 if needed:

```powershell
$env:GOPROXY='https://goproxy.cn,direct'
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Install frontend dependencies:

```powershell
cd frontend
npm install
```

Run checks:

```powershell
go test ./...
cd frontend
npm run build
```

Build the desktop app:

```powershell
$env:Path="$env:USERPROFILE\go\bin;$env:Path"
wails build -clean
```

Run in development:

```powershell
$env:Path="$env:USERPROFILE\go\bin;$env:Path"
wails dev
```

## Product Direction

RestDeck aims to cover the common local API testing workflow first, then grow toward a competitive non-account feature set. The guiding rule is simple: the UI should stay useful, compact, and truthful.
