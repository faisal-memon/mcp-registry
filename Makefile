GO ?= go
GOFMT ?= gofmt
CURL ?= curl

include .versions

TOOLS_DIR := .tools
GOLANGCI_LINT_DIR := $(TOOLS_DIR)/golangci-lint/$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT_BIN := $(GOLANGCI_LINT_DIR)/golangci-lint

.PHONY: lint lint-fix test tidy govulncheck

lint: $(GOLANGCI_LINT_BIN)
	@test -z "$$($(GOFMT) -l .)"
	@packages="$$($(GO) list ./... 2>/dev/null)"; \
	if test -n "$$packages"; then \
		$(GO) vet ./...; \
		$(GOLANGCI_LINT_BIN) run ./...; \
	fi

lint-fix: $(GOLANGCI_LINT_BIN)
	@$(GOLANGCI_LINT_BIN) run --fix ./...

test:
	@packages="$$($(GO) list ./... 2>/dev/null)"; \
	if test -n "$$packages"; then \
		$(GO) test ./...; \
	fi

tidy:
	@$(GO) mod tidy

govulncheck:
	@packages="$$($(GO) list ./... 2>/dev/null)"; \
	if test -n "$$packages"; then \
		$(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...; \
	fi

$(GOLANGCI_LINT_BIN):
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@rm -rf "$(dir $(GOLANGCI_LINT_DIR))"
	@mkdir -p "$(GOLANGCI_LINT_DIR)"
	@$(CURL) -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$(GOLANGCI_LINT_DIR)" "$(GOLANGCI_LINT_VERSION)"
