# Contributing to RestDeck

RestDeck is local-first and no-login by design. Contributions should keep the app compact, honest, and useful for daily API testing.

## UI Rules

- Do not add navigation items, tabs, buttons, or empty panels for features that are not implemented.
- Keep the workspace dense and practical: no landing page, marketing hero, decorative cards, or cloud/team prompts.
- Prefer code-owned components styled with Tailwind over heavy theme overrides.
- Use icon buttons with tooltips for compact commands, and text buttons only for clear actions like Send, Save, Import, and Export.

## Development Checks

Run these before sending a change:

```powershell
go test ./...
cd frontend
npm run build
```

For desktop verification:

```powershell
$env:Path="$env:USERPROFILE\go\bin;$env:Path"
wails build -clean
```

## Feature Policy

When adding a feature:

1. Implement the backend behavior.
2. Add or update tests.
3. Add the UI only after the behavior works.
4. Update `docs/FEATURE_MATRIX.md`.

Account login, team collaboration, and cloud sync are out of scope unless the project direction changes explicitly.
