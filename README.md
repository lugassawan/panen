<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="build/assets/logo-dark.svg" />
    <source media="(prefers-color-scheme: light)" srcset="build/assets/logo-light.svg" />
    <img src="build/assets/logo-light.svg" alt="Panen" height="80" />
  </picture>
</p>

<p align="center">
  Desktop decision engine for Indonesian Stock Exchange (IDX) investors.
  <br />
  Built with Go, Wails, Svelte 5, and Tailwind CSS.
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg" alt="License: Apache 2.0" /></a>
</p>

---

## Overview

Panen helps IDX investors make informed decisions with clarity and conviction. It runs as a native desktop app powered by [Wails](https://wails.io), combining a Go backend with a modern Svelte 5 frontend.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.26, Wails v2 |
| Frontend | Svelte 5, TypeScript, Tailwind CSS v4 |
| Icons | Lucide (lucide-svelte) |
| Fonts | Plus Jakarta Sans, DM Sans, DM Mono (self-hosted WOFF2) |
| Build | Vite 7, pnpm |
| Linting | golangci-lint v2 (custom plugin), Biome v2 |
| Tool versioning | mise |

## Getting Started

### Prerequisites

- [mise](https://mise.jdx.dev) for tool versioning (Go, Node.js, pnpm, etc.)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Setup

```sh
make setup    # Install Wails CLI, dependencies, and git hooks
make dev      # Start development server with HMR
```

### Key Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start Wails dev server with hot module reload |
| `make build` | Production build to `build/bin/` |
| `make lint` | Run Go (golangci-lint + custom plugin) and frontend (Biome) linters |
| `make fmt` | Auto-format Go and frontend code |
| `make test` | Run all tests (Go + frontend) |
| `make coverage` | Generate coverage reports |

## Project Structure

```
panen/
├── backend/
│   ├── app.go         # Composition root (App struct, Startup, Shutdown)
│   ├── presenter/     # Per-domain handlers, DTOs, converters
│   ├── domain/        # Entities, value objects, repository interfaces
│   ├── usecase/       # Application services (orchestration + validation)
│   └── infra/         # Database, scraper, platform implementations
├── frontend/src/
│   ├── lib/components/  # Reusable UI primitives (Button, Input, Select, etc.)
│   ├── components/      # Shared domain components (ConfirmDialog, Sidebar)
│   ├── pages/           # Page components organized by domain
│   ├── assets/fonts/    # Self-hosted WOFF2 font files
│   └── ...              # Stores, types, utilities
├── tools/lint/        # Custom golangci-lint plugin (panenlint)
├── build/assets/      # Brand SVG assets
├── main.go            # Wails entry point
└── wails.json         # Wails project config
```

## License

Licensed under the [Apache License, Version 2.0](LICENSE).

## Trademark Notice

"Panen" and the Panen logo are trademarks of Lugas Septiawan. Use of these trademarks in modified versions of this software requires prior written permission.
