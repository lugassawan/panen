.PHONY: dev build lint fmt frontend-install setup custom-gcl \
	test test-unit test-go test-frontend test-integration test-e2e \
	coverage coverage-go coverage-frontend playwright-install

dev:
	wails dev -tags dev

build:
	wails build

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

setup:
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	cd frontend && pnpm install
	git config core.hooksPath .githooks
