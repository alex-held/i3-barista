.DEFAULT_GOAL := help

GOLANGCI_LINT_VERSION ?= v1.19.1
TEST_FLAGS ?= -race
PKG_BASE ?= $(shell go list .)
PKGS ?= $(shell go list ./... | grep -v /vendor/)

.PHONY: help
help:
	@grep -E '^[a-zA-Z0-9-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "[32m%-10s[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build i3-barista
	go build -ldflags "-s -w" -o i3-barista

.PHONY: install
install: build clean ## install i3-barista
	cp i3-barista $(GOPATH)/bin

.PHONY: clean
clean:
	rm -f $(GOPATH)/bin/i3-barista

.PHONY: test
test: ## run tests
	go test $(TEST_FLAGS) $(PKGS)

.PHONY: vet
vet: ## run go vet
	go vet $(PKGS)

.PHONY: coverage
coverage: ## generate code coverage
	go test $(TEST_FLAGS) -covermode=atomic -coverprofile=coverage.txt $(PKGS)
	go tool cover -func=coverage.txt

.PHONY: lint
lint: ## run golangci-lint
	command -v golangci-lint > /dev/null 2>&1 || \
	  curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	golangci-lint run
