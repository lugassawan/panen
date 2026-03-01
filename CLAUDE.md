# Panen

Desktop decision engine for Indonesian Stock Exchange (IDX) investors.

## Tech Stack

- **Backend**: Go 1.26 with Wails v2 (native webview desktop app)
- **Frontend**: Svelte 5, TypeScript, Tailwind CSS v4
- **Build**: Vite 7, pnpm
- **Linting**: golangci-lint v2 with custom plugin (Go), Biome v2 (frontend)
- **Git hooks**: `.githooks/` with `core.hooksPath`
- **Tool versioning**: mise (Go, Node.js, pnpm, golangci-lint, Biome)

## Directory Layout

```
panen/
├── backend/app/       # Go application logic (Wails-bound structs)
├── frontend/src/      # Svelte 5 components and TypeScript
├── frontend/wailsjs/  # Auto-generated Wails bindings (gitignored)
├── tools/lint/        # Custom golangci-lint plugin (panenlint)
├── build/             # Build assets (app icon)
├── docs/plans/        # Design documents
├── main.go            # Wails entry point
└── wails.json         # Wails project config
```

## Commands

```sh
make setup             # Full project setup (wails CLI + deps + hooks)
make dev               # Start Wails dev server with HMR
make build             # Production build → build/bin/
make lint              # Build custom linter + run golangci-lint + Biome
make fmt               # Auto-format Go + frontend code
make test              # Run Go tests (app + lint analyzers)
make frontend-install  # Install frontend dependencies
```

## Conventions

- **Commits**: `type: description` only — no scopes, no `!` (enforced by `.githooks/commit-msg`)
- **Direct commits to main/master are blocked** by the pre-commit hook
- **Go**: Standard library style, `gofmt` formatting, tab indentation
- **Frontend**: 2-space indentation, double quotes, semicolons (Biome enforced)
- **Branches**: `feat/`, `fix/`, `chore/` prefixes

## Custom Linter (panenlint)

`tools/lint/` contains three analyzers built as a golangci-lint v2 module plugin:

- **maxparams**: forbids functions with >7 parameters
- **nolocalstruct**: forbids named struct declarations inside function bodies
- **nolateexport**: forbids exported standalone functions after unexported ones

`make lint` builds the custom binary (`custom-gcl`) via `.custom-gcl.yml` before running.

## Architecture

- `main.go` is a thin Wails bootstrap — all app logic lives in `backend/app/`
- Go methods on bound structs are auto-exposed to the frontend via `frontend/wailsjs/`
- Frontend is a standard Vite project; Wails proxies it during dev
- Git hooks live in `.githooks/` — other tools can inject blocks using `# BEGIN/END` markers
