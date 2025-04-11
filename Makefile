SHELL := /bin/bash
GOMODCACHE := $(GOPATH)/pkg/mod
GOCACHE := $(GOPATH)/.cache/go-build

.DEFAULT_GOAL := all

.PHONY: all
all: ## build pipeline
all: mod gen build spell lint test

.PHONY: precommit
precommit: ## validate the branch before commit
precommit: all vuln

.PHONY: ci
ci: ## CI build pipeline
ci: precommit diff

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## remove files created during build pipeline
	rm -rf dist
	rm -f coverage.*
	rm -f '"$(shell go env GOCACHE)/../golangci-lint"'
	GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) go clean -i -cache -testcache -modcache -fuzzcache -x

.PHONY: mod
mod: ## go mod tidy
	GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) go mod tidy

.PHONY: gen
gen: ## go generate
	GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) go generate ./...

.PHONY: build
build: ## go build
	GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) go build -o ./cmd/secure-shell/secure-shell ./cmd/secure-shell/main.go

.PHONY: spell
spell: ## misspell
	GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) go tool misspell -error -locale=US -w **.md

.PHONY: lint
lint: ## golangci-lint
	GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) GOLANGCI_LINT_CACHE=$(GOPATH)/.cache/golangci-lint go tool golangci-lint run --fix

.PHONY: vuln
vuln: ## govulncheck
	GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) go tool govulncheck ./...

.PHONY: test
test: ## go test
	GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./... ./...
	GOPATH=$(GOPATH) GOMODCACHE=$(GOMODCACHE) GOCACHE=$(GOCACHE) go tool cover -html=coverage.out -o coverage.html

.PHONY: diff
diff: ## git diff
	git diff --exit-code
	RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi
