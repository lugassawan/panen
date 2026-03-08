# Contributing to Panen

Thank you for your interest in contributing to Panen, a desktop decision engine for Indonesian Stock Exchange (IDX) investors. This guide will help you get up and running quickly.

## Getting Started

### Prerequisites

Install [mise](https://mise.jdx.dev/) to manage all required tool versions automatically:

```sh
# mise handles Go, Node.js, pnpm, golangci-lint, Biome, and golines
mise install
```

Pinned versions (from `mise.toml`):

| Tool | Version |
|------|---------|
| Go | 1.26 |
| Node.js | 25 |
| pnpm | 10 |
| golangci-lint | 2 |
| Biome | 2 |

### Setup

```sh
git clone git@github.com:lugassawan/panen.git
cd panen
mise install       # Install pinned tool versions
make setup         # Install Wails CLI, frontend deps, configure git hooks
```

`make setup` does three things:
1. Installs the Wails v2 CLI (`wails`)
2. Runs `pnpm install` in `frontend/`
3. Sets `core.hooksPath` to `.githooks/` for commit enforcement

### Running the App

```sh
make dev    # Start Wails dev server with hot module replacement
```

This opens the native desktop window and starts two servers:
- `http://localhost:34115` -- full app with Wails runtime (use this for testing backend features)
- `http://localhost:5173` -- frontend only (layout, theming, interactions)

## Development Workflow

### Branches

Create a branch from `main` with a type prefix:

```
feat/add-portfolio-view
fix/price-calculation
chore/update-dependencies
docs/add-api-guide
```

Direct commits to `main` are blocked by the pre-commit hook.

### Commits

Use the format `type: description` (enforced by the commit-msg hook):

| Type | When to use | Example |
|------|-------------|---------|
| `feat` | New user-facing feature | `feat: add portfolio summary page` |
| `fix` | Bug fix | `fix: correct dividend yield calculation` |
| `refactor` | Code restructuring, no behavior change | `refactor: extract price service` |
| `test` | Adding or updating tests | `test: add screener filter tests` |
| `docs` | Documentation changes | `docs: add API design document` |
| `chore` | Maintenance, deps, config | `chore: update Go dependencies` |
| `style` | Formatting only (no logic) | `style: fix indentation in app.css` |
| `perf` | Performance improvement | `perf: cache stock price lookups` |
| `build` | Build system changes | `build: upgrade Vite to v8` |
| `ci` | CI/CD changes | `ci: add release workflow` |
| `revert` | Revert a previous commit | `revert: revert portfolio page changes` |

No scopes or `!` breaking change markers. Split changes into logical commits -- separate infra, core logic, tests, and wiring.

### Pull Requests

- Title uses `type: description` format (same as commits)
- One logical change per PR
- Run `make lint` and `make test` before opening
- Fill in the PR template: issue link, summary, test plan

## Code Style

### Go

- Standard library style, formatted with `gofmt` (tab indentation)
- Line length capped at 120 characters (`golines`)
- Linted with golangci-lint v2 plus the custom `panenlint` plugin (see below)

### Frontend

- Formatted and linted with [Biome](https://biomejs.dev/) v2
- 2-space indentation, double quotes, semicolons
- No SvelteKit APIs (`$lib`, `$app/environment`) -- use relative imports
- `localStorage` only for UI preferences (theme, locale); all other state via the Go backend

### Formatting and Linting

```sh
make fmt    # Auto-format Go + frontend code
make lint   # Build custom linter + run golangci-lint + Biome check
```

The pre-commit hook runs formatters and linters automatically on staged files.

### Custom Linter (panenlint)

The project includes five custom analyzers in `tools/lint/`, built as a golangci-lint v2 module plugin:

| Analyzer | Rule |
|----------|------|
| **funcname** | No underscores in function names |
| **maxparams** | Max 7 function parameters |
| **nolateconst** | No `const`/`var` declarations after function declarations |
| **nolocalstruct** | No named struct declarations inside function bodies |
| **nolateexport** | No exported functions after unexported ones |

`make lint` builds the custom binary (`custom-gcl`) before running. Always fix the root cause of lint warnings rather than suppressing them. Use `//nolint` or `// biome-ignore` only as a last resort, with a justification comment.

## Testing

### Quick Reference

```sh
make test              # All unit + integration tests (fast)
make test-go           # Go tests only
make test-frontend     # Frontend unit tests only
make test-integration  # Frontend integration tests
make test-e2e          # E2E tests (requires browser)
make coverage          # Generate coverage reports → coverage/
```

### Go Tests

- Standard library `testing` package -- no testify
- Table-driven tests with `t.Run()` subtests
- Test files: `*_test.go` alongside source

Example (from `backend/usecase/brokerage_test.go`):

```go
func TestBrokerageServiceCreateNegativeFee(t *testing.T) {
	svc := NewBrokerageService(newMockBrokerageRepo(), newMockPortfolioRepo(), nil)

	for _, tt := range negativeFeeTests() {
		t.Run(tt.name, func(t *testing.T) {
			acct := &brokerage.Account{
				ID: shared.NewID(), BrokerName: "Broker",
				BuyFeePct: tt.buyFee, SellFeePct: tt.sellFee, SellTaxPct: tt.sellTax,
			}
			err := svc.Create(context.Background(), acct)
			if !errors.Is(err, ErrInvalidFee) {
				t.Errorf("Create() error = %v, want ErrInvalidFee", err)
			}
		})
	}
}
```

### Frontend Unit Tests

- Vitest + `@testing-library/svelte` + jsdom
- Test files: `*.test.ts` in `frontend/src/`
- Watch mode: `cd frontend && pnpm run test:unit:watch`

### Frontend Integration Tests

- Same stack, separate config (`vitest-integration.config.ts`)
- Test files: `*.integration.test.ts` in `frontend/src/`

### E2E Tests

```sh
make playwright-install  # One-time: install Chromium
make test-e2e
```

- Playwright with Chromium, auto-starts Vite dev server
- Test files: `frontend/e2e/*.spec.ts`

### Mocking Wails Bindings

The auto-generated `frontend/wailsjs/` bindings are gitignored and unavailable in tests. Mock them inline:

```ts
vi.mock("../wailsjs/go/backend/App", () => ({
  GetPortfolios: vi.fn(() => Promise.resolve([])),
}));
```

## Adding a New Language (i18n)

Translation files live in `frontend/src/i18n/`. Currently supported: English (`en.json`) and Indonesian (`id.json`).

### Steps

1. Copy `en.json` to a new file (e.g., `ja.json`)
2. Translate all user-facing strings. The file uses nested keys:

```json
{
  "common": {
    "save": "Save",
    "cancel": "Cancel"
  },
  "nav": {
    "dashboard": "Dashboard",
    "portfolio": "Portfolio"
  }
}
```

3. Keep financial terms in English (e.g., "EPS", "PBV", "Graham Number", "DCA") -- they are industry-standard abbreviations
4. Register the new locale in `frontend/src/i18n/locale.svelte.ts`:
   - Add the import for your JSON file
   - Add the locale key to the `messages` record
   - Update the `Locale` type in `frontend/src/i18n/types.ts`
5. Add a language option in the settings UI (`frontend/src/i18n/en.json` under `settings.language`)
6. Test with the locale switcher in Settings

### Interpolation

Strings with dynamic values use `{placeholder}` syntax:

```json
"tickerAdded": "{ticker} added to watchlist"
```

The `t()` helper handles substitution: `t("common.tickerAdded", { ticker: "BBCA" })`.

## Updating Broker Fees

Broker fee defaults are stored in `configs/brokers.json` (embedded at build time, live-refreshed from GitHub):

```json
{
  "code": "XC",
  "name": "Ajaib Sekuritas",
  "buyFeePct": 0.15,
  "sellFeePct": 0.15,
  "sellTaxPct": 0.10,
  "notes": "Commission-free for certain promo periods"
}
```

| Field | Description |
|-------|-------------|
| `code` | IDX member code (2-letter) |
| `name` | Official broker name |
| `buyFeePct` | Buy commission as percentage (0.15 = 0.15%) |
| `sellFeePct` | Sell commission as percentage |
| `sellTaxPct` | Final income tax on sell (PPh) |
| `notes` | Optional notes (promo info, caveats) |

### To update or add a broker

1. Edit `configs/brokers.json` -- add a new entry or update fee values
2. Keep entries sorted by `code` for readability
3. Submit a PR with type `chore` (e.g., `chore: update Ajaib broker fees`)

Changes to `configs/` on `main` are automatically picked up by running instances within 24 hours via the live config system.

## Creating a Custom Data Provider

Panen uses a pluggable provider system. All stock data flows through the `stock.DataProvider` interface defined in `backend/domain/stock/provider.go`:

```go
type DataProvider interface {
    Source() string
    FetchPrice(ctx context.Context, ticker string) (*PriceResult, error)
    FetchFinancials(ctx context.Context, ticker string) (*FinancialResult, error)
    FetchPriceHistory(ctx context.Context, ticker string) ([]PricePoint, error)
    FetchDividendHistory(ctx context.Context, ticker string) ([]dividend.DividendEvent, error)
}
```

### Provider Skeleton

```go
package myprovider

import (
    "context"

    "github.com/lugassawan/panen/backend/domain/dividend"
    "github.com/lugassawan/panen/backend/domain/stock"
)

type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) Source() string { return "myprovider" }

func (p *Provider) FetchPrice(ctx context.Context, ticker string) (*stock.PriceResult, error) {
    // Fetch current price, 52-week high/low from your data source.
    return &stock.PriceResult{
        Price:      0,
        High52Week: 0,
        Low52Week:  0,
    }, nil
}

func (p *Provider) FetchFinancials(ctx context.Context, ticker string) (*stock.FinancialResult, error) {
    // Fetch EPS, BVPS, ROE, DER, PBV, PER, DividendYield, PayoutRatio.
    return &stock.FinancialResult{}, nil
}

func (p *Provider) FetchPriceHistory(ctx context.Context, ticker string) ([]stock.PricePoint, error) {
    // Return daily OHLCV data.
    return nil, nil
}

func (p *Provider) FetchDividendHistory(ctx context.Context, ticker string) ([]dividend.DividendEvent, error) {
    // Return historical dividend events.
    return nil, nil
}
```

### Registration

Providers are registered in `backend/app.go` with a priority (lower number = tried first):

```go
registry.Register(myprovider.New(), 20)
```

The registry tries each enabled provider in priority order and falls back to the next on failure. Provider health is monitored via Settings > Data Providers.

### Testing

Write tests using mock data -- no network calls in unit tests. See `backend/infra/provider/registry_test.go` for examples of testing the fallback behavior.

## Architecture Overview

```
presenter  -->  usecase  -->  domain  <--  infra
(handlers)     (services)   (entities)    (database, scraper, platform)
```

### Clean Architecture Layers

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **Domain** | `backend/domain/` | Entities, value objects, repository interfaces. No external dependencies. |
| **Use Case** | `backend/usecase/` | Application services: orchestration, validation, business rules. |
| **Presenter** | `backend/presenter/` | DTOs, converters, Wails-bound handler methods. |
| **Infra** | `backend/infra/` | Database repos, scrapers, platform integrations, logging. |

Dependencies point inward: infra and presenter depend on domain, never the reverse.

### Wails Binding Pattern

- `main.go` bootstraps Wails with `backend/app.go` as the composition root
- `App` embeds per-domain handler structs from `presenter/`; Go method promotion exposes them to the frontend
- Auto-generated TypeScript bindings in `frontend/wailsjs/` (gitignored) bridge Go methods to the frontend

### Database Scan Helpers

Repos use generic scan helpers (`queryRow`, `queryAll`) from `backend/infra/database/scan.go`. Each repo provides a `scanFn` that maps SQL columns to a domain entity:

```go
func queryRow[T any](ctx context.Context, db *sql.DB, query string, scanFn scanFunc[T], args ...any) (T, error)
func queryAll[T any](ctx context.Context, db *sql.DB, query string, scanFn scanFunc[T], args ...any) ([]T, error)
```

This keeps SQL queries and struct mapping co-located in each repository file.

## Project Structure

```
backend/              Go backend
  app.go              Composition root (App struct, Startup, Shutdown)
  presenter/          Per-domain handlers, DTOs, converters
  domain/             Entities, value objects, repository interfaces
  usecase/            Application services (orchestration + validation)
  infra/              Database, scraper, platform, applog implementations
frontend/src/         Svelte 5 app
  pages/              Page components
  lib/components/     Reusable UI components
  lib/stores/         Svelte 5 rune-based stores (theme, mode, sync, etc.)
  i18n/               Internationalization (en/id translations)
configs/              Embedded config files (brokers, indices, sectors)
tools/lint/           Custom golangci-lint analyzers
docs/                 Documentation (design-system.md)
scripts/              Release and install scripts
```

## Questions?

Open a [GitHub issue](https://github.com/lugassawan/panen/issues) for bugs, feature requests, or questions about the codebase.
