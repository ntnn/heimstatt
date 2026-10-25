# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: The heimstatt Authors.

GO ?= go
GO_TEST_FLAGS ?= -race
WHAT ?= ./...

.PHONY: check check-fix
check:                    headers     apis-lint     lint     apis-test test
check-fix: apis-tidy tidy headers-fix apis-lint-fix lint-fix check

include Makefile.mindl

.PHONY: go.work
go.work:
	test -f go.work || go work init .
	go work use ./apis

.PHONY: headers
headers:
	$(GO) run ./hack/spdx-headers $(SPDX_HEADERS_FLAGS)

.PHONY: headers-fix
headers-fix: override SPDX_HEADERS_FLAGS := $(SPDX_HEADERS_FLAGS) -fix
headers-fix: headers

.PHONY: apis-tidy
apis-tidy:
	cd apis && $(GO) mod tidy -v

.PHONY: apis-lint
apis-lint: $(GOLANGCI_LINT) go.work
	cd apis && $(GOLANGCI_LINT) run $(GOLANGCI_LINT_FLAGS) $(WHAT)

.PHONY: apis-lint-fix
apis-lint-fix: override GOLANGCI_LINT_FLAGS := $(GOLANGCI_LINT_FLAGS) --fix
apis-lint-fix: apis-lint

.PHONY: tidy
tidy:
	$(GO) mod tidy -v

.PHONY: lint
lint: $(GOLANGCI_LINT) go.work
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_FLAGS) $(WHAT)

.PHONY: lint-fix
lint-fix: override GOLANGCI_LINT_FLAGS := $(GOLANGCI_LINT_FLAGS) --fix
lint-fix: lint

.PHONY: apis-test
apis-test: go.work
	cd apis && $(GO) test $(GO_TEST_FLAGS) $(WHAT)

.PHONY: test
test: go.work
	$(GO) test $(GO_TEST_FLAGS) $(WHAT)
