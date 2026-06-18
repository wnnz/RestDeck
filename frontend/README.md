# RestDeck Frontend

The frontend is a Vue 3 + Vite workspace embedded by Wails.

## Commands

```powershell
npm install
npm run build
npm run dev
```

The production desktop app should be built from the repository root with Wails:

```powershell
wails build -clean
```

## UI Direction

The UI follows a compact API-client workbench style: narrow icon navigation, collection sidebar, request editor, response viewer, and local-only utility panels. Do not add UI for features that are not implemented in the backend.
