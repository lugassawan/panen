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
make test   # Run all tests
```

### Go

- Formatted with `gofmt` (tab indentation)
- Linted with golangci-lint v2 + custom `panenlint` plugin
- Custom rules: max 7 function params, no local struct declarations, exported functions before unexported

### Frontend

- Formatted and linted with [Biome](https://biomejs.dev/) v2 (managed by mise, not a npm devDependency)
- 2-space indentation, double quotes, semicolons

## Git Workflow

### Branches

Create a branch from `main` with a type prefix:

```
feat/add-portfolio-view
fix/price-calculation
chore/update-dependencies
```

### Commits

Conventional Commits format, strictly `type: description`:

```
feat: add portfolio summary page
fix: correct dividend yield calculation
chore: update Go dependencies
docs: add API design document
refactor: extract price service
test: add screener filter tests
```

No scopes or breaking change markers. Direct commits to `main` are blocked.

### Pull Requests

- One logical change per PR
- Ensure `make lint` and `make test` pass before opening
- Fill in the PR template (issue link, summary, test plan)

## Project Structure

```
backend/app/       Go application logic (Wails-bound methods)
frontend/src/      Svelte 5 components and TypeScript
tools/lint/        Custom golangci-lint analyzers
build/             Build assets
docs/plans/        Design documents
```
