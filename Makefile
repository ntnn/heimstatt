# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: The heimstatt Authors.

GO ?= go

include Makefile.mindl

.PHONY: check check-fix
check: headers lint
check-fix: headers-fix lint-fix

.PHONY: headers
headers:
	$(GO) run ./hack/spdx-headers $(SPDX_HEADERS_FLAGS)

.PHONY: headers-fix
headers-fix: override SPDX_HEADERS_FLAGS := $(SPDX_HEADERS_FLAGS) -fix
headers-fix: headers

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_FLAGS) ./...

.PHONY: lint-fix
lint-fix: override GOLANGCI_LINT_FLAGS := $(GOLANGCI_LINT_FLAGS) --fix
lint-fix: lint
