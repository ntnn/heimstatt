# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: The heimstatt Authors.

GO ?= go
GO_TEST_FLAGS ?= -race
WHAT ?= ./...

.PHONY: check check-fix
check:                    headers              apis-lint     lint     apis-test test
check-fix: apis-tidy tidy headers-fix generate apis-lint-fix lint-fix check

include Makefile.mindl

export PATH := $(TOOLS_DIR):$(PATH)

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

.PHONY: generate
generate: $(CONTROLLER_GEN) go.work
	$(CONTROLLER_GEN) object paths=./apis/...
	$(CONTROLLER_GEN) crd output:crd:dir=config/deploy/crds paths=./apis/deploy/...

.PHONY: apis-test
apis-test: go.work
	cd apis && $(GO) test $(GO_TEST_FLAGS) $(WHAT)

.PHONY: test
test: go.work
	$(GO) test $(GO_TEST_FLAGS) $(WHAT)

TILT_CLUSTER ?= heimstatt
TILT_KUBECONFIG ?= tilt.kubeconfig.yaml
TILT_UP := up

.PHONY: tilt-up
tilt-up: $(TILT) go.work
	kind get clusters | grep -q $(TILT_CLUSTER) || kind create cluster --config ./config/tilt/kind.yaml --name $(TILT_CLUSTER)
	kind export kubeconfig --name $(TILT_CLUSTER) --kubeconfig $(TILT_KUBECONFIG)
	echo KUBECONFIG="$(TILT_KUBECONFIG)" $(TILT) $(TILT_UP)

.PHONY: tilt-ci
tilt-ci: override TILT_UP := ci
tilt-ci: tilt-up
