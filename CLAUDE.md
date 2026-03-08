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
│   └── infra/           # Database, scraper, provider, backup, platform, applog, etc.
├── frontend/src/        # Svelte 5 components and TypeScript
│   └── i18n/            # Internationalization (en/id translations)
├── frontend/wailsjs/    # Auto-generated Wails bindings (gitignored)
├── tools/lint/          # Custom golangci-lint plugin (panenlint)
├── build/               # Build assets (app icon)
├── configs/             # Embedded config files (brokers, indices, sectors)
├── docs/                # Documentation (design system, user guides)
├── .github/workflows/   # CI pipeline (test, release)
├── scripts/             # Release and install scripts
├── CONTRIBUTING.md      # Contributor guide
├── main.go              # Wails entry point
└── wails.json           # Wails project config
```

## Commands

```sh
make init              # Lightweight setup for worktrees (skip wails CLI)
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
- **Commit splitting**: Split changes into logical commits — separate infra/config, core logic, tests, and wiring. Never bundle unrelated changes into a single commit.
- **Direct commits to main/master are blocked** by the pre-commit hook
- **Go**: Standard library style, `gofmt` formatting, tab indentation
- **Frontend**: 2-space indentation, double quotes, semicolons (Biome enforced)
- **Branches**: `feat/`, `fix/`, `chore/` prefixes
- **Worktrees**: Do NOT default to worktrees. Only use git worktrees when running parallel agents working on independent tasks simultaneously — a single agent on one task should use a regular branch, never a worktree. Never target main/master (pre-commit hook blocks direct commits). Use `rimba add <task>` to create worktrees — this runs `make init` automatically via `post_create` hook. Clean up with `rimba remove <task>`. `make setup` installs rimba hooks for auto-cleanup of merged worktrees on `git pull`
- **Lint warnings**: Always fix the root cause before considering suppression. Refactor code, extract helpers, or restructure queries to satisfy the linter. Only use `//nolint` or `// biome-ignore` as a last resort when a fix is genuinely impossible, and always include a justification comment explaining why.
- **Code review**: Run code review before creating PRs (e.g., via available code-review skills or agents) unless one was already performed in the current session. Address reviewer feedback to maintain code quality.
- **PRs**: Title uses `type: description` (same types as commits); body follows `.github/pull_request_template.md`
- **Deferred tasks**: When planning surfaces improvement ideas not implemented immediately, create GitHub issues to track them. These issues do NOT belong in the PR's `## Issue` section — that section is reserved for issues the PR directly closes/fixes/resolves.

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
- **Pool**: `threads` (not `vmThreads` — vmThreads `vi.mock()` closures break on Linux due to VM context isolation, even with identical Node.js versions)
- **CI sharding**: Tests are split across 3 parallel CI jobs via `vitest --shard` for faster feedback
- **No hardcoded locale strings in tests**: `Intl.NumberFormat` output differs across platforms (e.g., `id-ID` currency: `"Rp 9.250"` on macOS vs `"IDR 9,250"` on Ubuntu). Use `Intl.NumberFormat` helpers to compute expected values.

### Frontend E2E (Playwright)

- Tests in `frontend/e2e/*.spec.ts` — run with `make test-e2e`
- Chromium only, auto-starts Vite dev server
- Requires `make playwright-install` before first run

### Manual Testing

- Start the app with `make dev` — this launches the native Wails window, the Vite dev server at `http://localhost:5173`, and the Wails dev server at `http://localhost:34115`
- **Use `http://localhost:34115` for full testing** — Wails runtime is injected, so backend-dependent features (stock lookup, portfolio actions) work in the browser
- Use `http://localhost:5173` only for frontend-only checks (layout, theming, interactions) — Wails runtime bindings are **not** available on this port
- Use **Claude Chrome extension** tools (`mcp__claude-in-chrome__*`) for visual/UI verification; fall back to **Playwright MCP** tools (`mcp__plugin_playwright_playwright__*`) if the Chrome extension is unavailable
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
- Additional stores: `sync.svelte.ts`, `command-palette.svelte.ts`, `toast.svelte.ts`, `alerts.svelte.ts`

### Typography

- `font-display` (Plus Jakarta Sans): headings, tickers, brand
- `font-body` (DM Sans): body text, labels — set on `html` as default
- `font-mono` (DM Mono): **always** use for financial numbers (prices, percentages, ratios)

### Components

Reusable components in `frontend/src/lib/components/`: Alert, Badge, BrokerPicker, Button, CommandPalette, ConfirmDialog, DataTimestamp, EmptyState, Input, LoadingState, Modal, ModeTabs, SearchableSelect, Select, SkeletonCard, SkeletonLine, SkeletonTable, SortableHeader, StockCard, SyncIndicator, ThemeToggle, Toast, ToastContainer, Tooltip, UpdateDialog

### Key Rules

- Use semantic color tokens (`bg-bg-elevated`, `text-text-secondary`, `border-border-default`), not raw palette colors for themed surfaces
- Use `text-profit` / `text-loss` for financial gain/loss colors (adapts to theme)
- Use `focus-ring` on all interactive elements
- Use `transition-fast` (120ms) for hover/focus, `transition-normal` (200ms) for dropdowns
- Three-layer depth: `bg-secondary` (sidebar) → `bg-primary` (canvas) → `bg-elevated` (cards)
- Sidebar width: `w-sidebar` (220px)
- No SvelteKit APIs (`$lib`, `$app/environment`) — use relative imports
- `localStorage` only for UI preferences (theme, locale); all other state via Go backend

## Internationalization

- Translation files: `en.json`, `id.json` in `frontend/src/i18n/`
- Svelte 5 runes-based locale store (`locale.svelte.ts`)
- `t()` helper from `i18n/index.ts` for user-facing strings
- Persists to `localStorage`, detects system language

## Architecture

- `main.go` is a thin Wails bootstrap — `backend/app.go` is the composition root
- `App` embeds per-domain handler structs from `backend/presenter/`; Go promotes their methods for Wails binding
- Go methods on bound structs are auto-exposed to the frontend via `frontend/wailsjs/`
- Frontend is a standard Vite project; Wails proxies it during dev
- Database repos use generic scan helpers (`queryRow`, `queryAll`) in `backend/infra/database/scan.go` — each repo provides a `scanFn` that maps columns to a domain entity, keeping SQL and struct mapping co-located
- Git hooks live in `.githooks/` — other tools can inject blocks using `# BEGIN/END` markers

### Data Provider System

- `backend/domain/stock/provider.go` defines `DataProvider` interface — `FetchPrice`, `FetchFinancials`, `FetchPriceHistory`, `FetchDividendHistory`, `HealthCheck`
- `backend/domain/provider/status.go` defines `Registry` interface and health status types (`StatusHealthy`, `StatusDegraded`, `StatusDown`, `StatusUnknown`)
- `backend/infra/provider/registry.go` implements `Registry` with priority-based fallback — tries providers in order, falls back on failure
- Presenters depend on domain interfaces (`domainProvider.Registry`), not infra implementations
- Providers: Yahoo Finance (primary), IDX (secondary). New providers implement `stock.DataProvider`

### Backup and Export/Import

- `backend/infra/backup/` — daily auto-backup, pre-destructive backup, manual backup, cleanup with retention
- `backend/usecase/export.go` — creates zip archive (SQLite DB + `meta.json` with SHA-256 checksum)
- `backend/usecase/import.go` — validates checksum, `PRAGMA quick_check`, atomic file replacement (temp file + `os.Rename`), zip bomb protection via `io.LimitReader`

## CI Pipeline

- `.github/workflows/test.yml` — runs on push and PRs: lint (`make lint`), Go tests (`make test-go`), frontend unit + integration tests
- `.github/workflows/release.yml` — triggered by `v*` tags: cross-platform build (macOS universal, Linux amd64, Windows amd64), creates GitHub Release with archives + checksums
- Uses `jdx/mise-action@v3` for tool version management in CI — pin tool versions to minor (e.g., `node = "25.6"` not `"25"`) so CI and local resolve the same version
- Frontend tests are sharded across 3 parallel CI jobs for faster feedback
- Pure-Go SQLite (`modernc.org/sqlite`) — no CGO dependency, `CGO_ENABLED=0` works
- Release notes are auto-generated from conventional commits between tags — no manual CHANGELOG needed

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

Release builds use Wails production mode with `CGO_ENABLED=0` (pure-Go SQLite).

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
