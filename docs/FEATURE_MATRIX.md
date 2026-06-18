# RestDeck Feature Matrix

Status legend:

- Done: implemented and visible in the app.
- Partial: implemented for common cases, with known gaps.
- Planned: not visible in the app yet.
- Excluded: intentionally outside the no-account/local-first scope.

| Area | Status | Notes |
| --- | --- | --- |
| HTTP requests | Done | Method, URL, params, headers, body, timeout, send, response metadata. |
| Collections | Done | Local collections and request persistence. |
| Environments | Done | Active environment and local variables. |
| Globals | Done | Local global variables. |
| History | Done | Recent sends stored locally in SQLite. |
| Runner | Done | Runs the selected collection once or by iteration from backend API. |
| Response viewer | Done | Body, headers, cookies, test results, status, duration, size. |
| Postman Collection import/export | Partial | Common v2.1 collection, folder, request, body, auth, and script fields. |
| Pre-request scripts | Partial | Common `pm.variables` subset. |
| Test scripts | Partial | `pm.test`, `pm.response`, `pm.request`, variables, and basic `expect`. |
| Dynamic variables | Partial | Common random/time variables. |
| API Key auth | Done | Header or query param. |
| Bearer auth | Done | Authorization bearer token. |
| Basic auth | Done | Standard HTTP basic auth. |
| Digest auth | Partial | UI and stored config exist; full challenge-response flow is not complete. |
| OAuth 1 | Partial | HMAC-SHA1 signing for common requests. |
| OAuth 2 | Partial | Existing access token bearer flow only. |
| Secret storage | Partial | Local AES-GCM wrapper stored under user config. OS keychain integration is planned. |
| WebSocket | Partial | Connects, sends one text message, and reads replies. Persistent sessions are planned. |
| SSE | Partial | Connects to an event stream and collects a bounded number of events. |
| gRPC | Planned | Later milestone. |
| OpenAPI import/export | Planned | Later milestone. |
| Request documentation panel | Planned | Do not show in UI until generated docs exist. |
| Mock servers | Planned | Hidden until implemented. |
| Monitors | Planned | Hidden until implemented. |
| Flows | Planned | Hidden until implemented. |
| Team/cloud workspaces | Excluded | No account login or cloud dependency. |
| AI features | Excluded | Not part of the local-first v1. |
