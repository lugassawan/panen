SHELL := /bin/bash
export PATH := $(shell mise bin-paths 2>/dev/null | tr '\n' ':'):$(PATH)

.PHONY: dev build lint fmt frontend-install setup custom-gcl test

dev:
	wails dev

build:
	wails build

custom-gcl:
	golangci-lint custom

lint: custom-gcl
	./custom-gcl run ./...
	cd frontend && biome check .

fmt:
	gofmt -w .
	cd frontend && biome format --write .

test:
	go test ./...
	cd tools/lint && go test ./...

frontend-install:
	cd frontend && pnpm install

setup:
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	cd frontend && pnpm install
	git config core.hooksPath .githooks
