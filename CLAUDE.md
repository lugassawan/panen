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
├── backend/
│   ├── app.go           # Composition root (App struct, Startup, Shutdown)
│   ├── presenter/       # Per-domain handlers, DTOs, converters
│   ├── domain/          # Entities, value objects, repository interfaces
│   ├── usecase/         # Application services (orchestration + validation)
│   └── infra/           # Database, scraper, platform implementations
├── frontend/src/        # Svelte 5 components and TypeScript
├── frontend/wailsjs/    # Auto-generated Wails bindings (gitignored)
├── tools/lint/          # Custom golangci-lint plugin (panenlint)
├── build/               # Build assets (app icon)
├── docs/plans/          # Design documents
├── main.go              # Wails entry point
└── wails.json           # Wails project config
```

## Commands

```sh
make setup             # Full project setup (wails CLI + deps + hooks)
make dev               # Start Wails dev server with HMR
make build             # Production build → build/bin/
make lint              # Build custom linter + run golangci-lint + Biome
make fmt               # Auto-format Go + frontend code
make test              # Run all unit + integration tests (Go + frontend)
make test-go           # Go tests only (app + lint analyzers)
make test-frontend     # Frontend unit tests only (Vitest)
make test-integration  # Frontend integration tests only (Vitest)
make test-e2e          # Frontend E2E tests (Playwright, requires browser)
make coverage          # Generate coverage reports (Go + frontend)
make playwright-install # Install Chromium for E2E tests (run once)
make frontend-install  # Install frontend dependencies
```

## Conventions

- **Commits**: `type: description` only — no scopes, no `!` (enforced by `.githooks/commit-msg`)
- **Direct commits to main/master are blocked** by the pre-commit hook
- **Go**: Standard library style, `gofmt` formatting, tab indentation
- **Frontend**: 2-space indentation, double quotes, semicolons (Biome enforced)
- **Branches**: `feat/`, `fix/`, `chore/` prefixes
- **Worktrees**: Only use git worktrees when running parallel agents on independent tasks — never for single sequential work or when targeting main/master (pre-commit hook blocks direct commits)
- **Code review**: Prefer running code review before creating PRs (e.g., via available code-review skills or agents)

## Custom Linter (panenlint)

`tools/lint/` contains three analyzers built as a golangci-lint v2 module plugin:

- **maxparams**: forbids functions with >7 parameters
- **nolocalstruct**: forbids named struct declarations inside function bodies
- **nolateexport**: forbids exported standalone functions after unexported ones

`make lint` builds the custom binary (`custom-gcl`) via `.custom-gcl.yml` before running.

## Testing

### Backend (Go)

- Standard library `testing` package — no testify
- Table-driven tests with `t.Run()` subtests
- Test files: `*_test.go` alongside source files
- Run: `make test-go` or `go test ./...`

### Frontend Unit/Integration (Vitest)

- **Unit tests**: `*.test.ts` in `frontend/src/` — run with `make test-frontend`
- **Integration tests**: `*.integration.test.ts` — run with `make test-integration`
- Uses `@testing-library/svelte` + `jsdom` environment
- Separate Vitest configs: `vitest.config.ts` (unit) and `vitest-integration.config.ts` (integration)

### Frontend E2E (Playwright)

- Tests in `frontend/e2e/*.spec.ts` — run with `make test-e2e`
- Chromium only, auto-starts Vite dev server
- Requires `make playwright-install` before first run

### Wails Mock Pattern

Auto-generated `wailsjs/` bindings are gitignored and unavailable in tests. Mock them inline:

```ts
vi.mock("../wailsjs/go/backend/App", () => ({
  Greet: vi.fn((name: string) => Promise.resolve(`Hello ${name}!`)),
}));
```

### Coverage

- Go: `make coverage-go` → `coverage/go/`
- Frontend: `make coverage-frontend` → `coverage/frontend/`
- Both: `make coverage`

## Architecture

- `main.go` is a thin Wails bootstrap — `backend/app.go` is the composition root
- `App` embeds per-domain handler structs from `backend/presenter/`; Go promotes their methods for Wails binding
- Go methods on bound structs are auto-exposed to the frontend via `frontend/wailsjs/`
- Frontend is a standard Vite project; Wails proxies it during dev
- Git hooks live in `.githooks/` — other tools can inject blocks using `# BEGIN/END` markers
