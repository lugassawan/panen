# Contributing to Panen

## Prerequisites

- [mise](https://mise.jdx.dev/) — manages Go, Node.js, pnpm, golangci-lint, and Biome

## Setup

```sh
git clone git@github.com:lugassawan/panen.git
cd panen
mise install       # Install pinned tool versions
make setup         # Install Wails CLI, frontend deps, git hooks
```

## Development

```sh
make dev    # Start dev server with hot reload
make build  # Production build
```

## Code Quality

```sh
make lint   # Run all linters (Go + frontend)
make fmt    # Auto-format all code
make test   # Run all unit + integration tests
```

### Go

- Formatted with `gofmt` (tab indentation)
- Linted with golangci-lint v2 + custom `panenlint` plugin
- Custom rules: max 7 function params, no local struct declarations, exported functions before unexported

### Frontend

- Formatted and linted with [Biome](https://biomejs.dev/) v2 (managed by mise, not a npm devDependency)
- 2-space indentation, double quotes, semicolons

## Testing

Run all fast tests (unit + integration) with:

```sh
make test
```

### Go Tests

```sh
make test-go    # Run Go tests (app + lint analyzers)
```

- Standard library `testing` package (no testify)
- Table-driven tests with `t.Run()` subtests
- Test files: `*_test.go` alongside source

### Frontend Unit Tests

```sh
make test-frontend    # Run once
cd frontend && pnpm run test:unit:watch  # Watch mode
```

- Vitest + `@testing-library/svelte` + jsdom
- Test files: `*.test.ts` in `frontend/src/`
- Mock Wails bindings with `vi.mock()` (see CLAUDE.md for pattern)

### Frontend Integration Tests

```sh
make test-integration
```

- Same stack as unit tests, separate config (`vitest-integration.config.ts`)
- Test files: `*.integration.test.ts` in `frontend/src/`
- Longer timeout (10s) for tests involving multiple component interactions

### E2E Tests

```sh
make playwright-install  # One-time: install Chromium
make test-e2e            # Run E2E tests
```

- Playwright with Chromium, auto-starts Vite dev server
- Test files: `frontend/e2e/*.spec.ts`
- Not included in `make test` (requires browser, slower)

### Coverage

```sh
make coverage            # Go + frontend coverage reports
make coverage-go         # Go only → coverage/go/
make coverage-frontend   # Frontend only → coverage/frontend/
```

## Git Workflow

### Branches

Create a branch from `main` with a type prefix:

```
feat/add-portfolio-view
fix/price-calculation
chore/update-dependencies
```

### Commits

Conventional Commits format, strictly `type: description`. The commit-msg hook enforces these types:

| Type | When to use | Example |
|------|-------------|---------|
| `feat` | New user-facing feature or capability | `feat: add portfolio summary page` |
| `fix` | Bug correction | `fix: correct dividend yield calculation` |
| `chore` | Maintenance, dependency updates, config tweaks | `chore: update Go dependencies` |
| `docs` | Documentation only (README, CLAUDE.md, comments) | `docs: add API design document` |
| `refactor` | Code restructuring with no behavior change | `refactor: extract price service` |
| `test` | Adding or updating tests only | `test: add screener filter tests` |
| `style` | Formatting, whitespace, semicolons (no logic change) | `style: fix indentation in app.css` |
| `perf` | Performance improvement | `perf: cache stock price lookups` |
| `build` | Build system or tooling changes (Wails, Vite, Makefile) | `build: upgrade Vite to v8` |
| `ci` | CI/CD pipeline changes (GitHub Actions, workflows) | `ci: add release workflow` |
| `revert` | Reverting a previous commit | `revert: revert portfolio page changes` |

No scopes or breaking change markers. Direct commits to `main` are blocked.

### Pull Requests

- Title uses `type: description` format (same types as commits above)
- One logical change per PR
- Ensure `make lint` and `make test` pass before opening
- Fill in the PR template (issue link, summary, test plan)

### Deferred Tasks

When planning or reviewing surfaces improvement ideas not implemented immediately, create GitHub issues to track them. These issues do NOT belong in the PR's `## Issue` section — that section is reserved for issues the PR directly closes/fixes/resolves.

## Project Structure

```
backend/              Go backend (presenter, domain, usecase, infra layers)
frontend/src/         Svelte 5 app (pages/, lib/components/, i18n/, stores, utils)
tools/lint/           Custom golangci-lint analyzers
build/                Build assets
docs/                 Documentation (design-system.md)
```
