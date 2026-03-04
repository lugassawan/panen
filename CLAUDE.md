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
├── docs/                # Documentation
│   └── design-system.md # Design system reference (tokens, components, patterns)
├── scripts/             # Release and install scripts
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
make release-check     # Validate VERSION against wails.json productVersion
```

## Conventions

- **Commits**: `type: description` — valid types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `style`, `perf`, `build`, `ci`, `revert` (enforced by `.githooks/commit-msg`; no scopes, no `!`)
- **Direct commits to main/master are blocked** by the pre-commit hook
- **Go**: Standard library style, `gofmt` formatting, tab indentation
- **Frontend**: 2-space indentation, double quotes, semicolons (Biome enforced)
- **Branches**: `feat/`, `fix/`, `chore/` prefixes
- **Worktrees**: Only use git worktrees when running parallel agents on independent tasks — never for single sequential work or when targeting main/master (pre-commit hook blocks direct commits)
- **Code review**: Prefer running code review before creating PRs (e.g., via available code-review skills or agents)
- **PRs**: Title uses `type: description` (same types as commits); body follows `.github/pull_request_template.md`

## Custom Linter (panenlint)

`tools/lint/` contains five analyzers built as a golangci-lint v2 module plugin:

- **funcname**: forbids underscores in function names
- **maxparams**: forbids functions with >7 parameters
- **nolateconst**: forbids package-level const/var declarations after function declarations
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

### Manual Testing

- Start the app with `make dev` — this launches the native Wails window and the Vite dev server at `http://localhost:5173`
- Use **Claude Chrome extension** tools (`mcp__claude-in-chrome__*`) for visual/UI verification; fall back to **Playwright MCP** tools (`mcp__plugin_playwright_playwright__*`) if the Chrome extension is unavailable
- Wails runtime bindings are unavailable in browser — backend-dependent features (stock lookup, portfolio actions) won't work; use for layout, navigation, theming, and interaction verification
- Include manual verification steps in PR test plans when changes affect UI

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

## Design System

Full reference: `docs/design-system.md`

### Tokens & Theming

- All design tokens live in `frontend/src/app.css` via Tailwind v4 `@theme` — no `tailwind.config.js`
- Two brand palettes: **green** (Value Mode, primary accent) and **gold** (Dividend Mode)
- Semantic colors (`bg-primary`, `text-secondary`, `border-default`) adapt to light/dark via CSS variables in `:root` / `.dark`
- Theme store (`lib/stores/theme.svelte.ts`): light/dark/system, persists to `localStorage`, toggles `.dark` class on `<html>`
- Mode store (`lib/stores/mode.svelte.ts`): value/dividend, switches accent colors globally

### Typography

- `font-display` (Plus Jakarta Sans): headings, tickers, brand
- `font-body` (DM Sans): body text, labels — set on `html` as default
- `font-mono` (DM Mono): **always** use for financial numbers (prices, percentages, ratios)

### Components

Reusable components in `frontend/src/lib/components/`: Button, Badge, Alert, ThemeToggle, ModeTabs, StockCard

### Key Rules

- Use semantic color tokens (`bg-bg-elevated`, `text-text-secondary`, `border-border-default`), not raw palette colors for themed surfaces
- Use `text-profit` / `text-loss` for financial gain/loss colors (adapts to theme)
- Use `focus-ring` on all interactive elements
- Use `transition-fast` (120ms) for hover/focus, `transition-normal` (200ms) for dropdowns
- Three-layer depth: `bg-secondary` (sidebar) → `bg-primary` (canvas) → `bg-elevated` (cards)
- Sidebar width: `w-sidebar` (220px)
- No SvelteKit APIs (`$lib`, `$app/environment`) — use relative imports
- `localStorage` only for UI preferences (theme); all other state via Go backend

## Architecture

- `main.go` is a thin Wails bootstrap — `backend/app.go` is the composition root
- `App` embeds per-domain handler structs from `backend/presenter/`; Go promotes their methods for Wails binding
- Go methods on bound structs are auto-exposed to the frontend via `frontend/wailsjs/`
- Frontend is a standard Vite project; Wails proxies it during dev
- Git hooks live in `.githooks/` — other tools can inject blocks using `# BEGIN/END` markers

## Release

### Workflow

1. Update `wails.json` `info.productVersion` to the new version (e.g. `0.2.0`)
2. Commit: `chore: bump version to 0.2.0`
3. Run: `scripts/release.sh 0.2.0` (validates version, creates tag, pushes)
4. CI builds all platforms, creates GitHub Release with archives + checksums

```sh
scripts/release.sh 0.2.0    # Manual version
scripts/release.sh --auto    # Auto-detect from conventional commits
make release-check VERSION=0.2.0  # Local validation only (no tag/push)
```

### Archive Formats

| Platform | Archive | Contents |
|----------|---------|----------|
| macOS | `panen-darwin-universal.zip` | `panen.app/` bundle |
| Linux | `panen-linux-amd64.tar.gz` | Binary + `.desktop` + icon |
| Windows | `panen-windows-amd64.zip` | `panen.exe` |

Release builds use `garble` + Wails `-obfuscated` for Go binary obfuscation.

### Install Script

```sh
# Install latest release (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/lugassawan/panen/main/scripts/install.sh | sh

# Install specific version
PANEN_VERSION=v0.2.0 curl -fsSL https://raw.githubusercontent.com/lugassawan/panen/main/scripts/install.sh | sh
```

Install locations (no sudo required):
- **macOS**: `~/Applications/panen.app`
- **Linux**: `~/.local/bin/panen` + `.desktop` + icon
