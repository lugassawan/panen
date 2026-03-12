.PHONY: dev build build-darwin build-linux build-windows build-all \
	lint fmt frontend-install setup init custom-gcl \
	test test-unit test-go test-frontend test-integration test-e2e \
	coverage coverage-go coverage-frontend playwright-install \
	release-check

VERSION := $(shell jq -r '.info.productVersion' wails.json)
LDFLAGS := -X github.com/lugassawan/panen/backend.version=$(VERSION)

dev:
	wails dev -tags dev

build:
	wails build -ldflags "$(LDFLAGS)"

build-darwin:
	wails build -clean -platform darwin/universal -ldflags "$(LDFLAGS)"

build-linux:
	@if [ "$$(uname)" != "Linux" ]; then \
		echo "Error: Linux builds require a Linux host (GTK headers needed)." >&2; \
		exit 1; \
	fi
	wails build -clean -platform linux/amd64 -ldflags "$(LDFLAGS)"

build-windows:
	wails build -clean -platform windows/amd64 -ldflags "$(LDFLAGS)"

# build-all builds for platforms that support cross-compilation from the current host.
# Linux is excluded because Wails requires native GTK headers (use CI or a Linux host).
build-all: build-darwin build-windows

custom-gcl:
	golangci-lint custom

lint: custom-gcl
	./custom-gcl run ./...
	cd frontend && biome check .

fmt:
	gofmt -w .
	golines -w --max-len=120 .
	cd frontend && biome format --write .

# Test targets — `make test` runs unit + integration (fast); E2E is separate
test: test-unit test-integration

test-unit: test-go test-frontend

test-go:
	go test ./...
	cd tools/lint && go test ./...

test-frontend:
	cd frontend && pnpm run test:unit

test-integration:
	cd frontend && pnpm run test:integration

test-e2e:
	cd frontend && pnpm run test:e2e

# Coverage targets
coverage: coverage-go coverage-frontend

coverage-go:
	mkdir -p coverage/go
	go test -coverprofile=coverage/go/coverage.out ./...
	go tool cover -html=coverage/go/coverage.out -o coverage/go/coverage.html
	# Note: tools/lint/ is a separate Go module; its coverage is not included here.

coverage-frontend:
	cd frontend && pnpm run coverage

# Playwright browser install (run once before E2E tests)
playwright-install:
	cd frontend && pnpm exec playwright install --with-deps chromium

frontend-install:
	cd frontend && pnpm install

init:
	cd frontend && pnpm install
	mkdir -p frontend/dist
	git config core.hooksPath .githooks

setup:
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	cd frontend && pnpm install
	git config core.hooksPath .githooks

release-check:
	@TAG_VERSION="$(VERSION)"; \
	WAILS_VERSION=$$(jq -r '.info.productVersion' wails.json); \
	if [ -z "$$TAG_VERSION" ]; then echo "Usage: make release-check VERSION=0.2.0"; exit 1; fi; \
	if [ "$$TAG_VERSION" != "$$WAILS_VERSION" ]; then \
		echo "ERROR: VERSION ($$TAG_VERSION) != wails.json productVersion ($$WAILS_VERSION)"; exit 1; \
	fi; \
	echo "Version $$WAILS_VERSION OK"
