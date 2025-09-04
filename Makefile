GO ?= go
HAS_GO := $(shell hash $(GO) > /dev/null 2>&1 && echo yes)
TAG ?= $(shell git describe --tags --abbrev=0 HEAD)
DATE_FMT = +"%Y-%m-%dT%H:%M:%S%z"
BUILD_DATE ?= $(shell date "$(DATE_FMT)")

TEST_TAGS ?=
GOTESTFLAGS ?=
GO_TEST_PACKAGES ?= $(filter-out $(shell $(GO) list ./tests/...),$(shell $(GO) list ./...))
GO_INTEGRATION_TEST_PACKAGES ?= $(shell $(GO) list ./tests/...)

ifeq ($(HAS_GO), yes)
	GOPATH ?= $(shell $(GO) env GOPATH)
	export PATH := $(GOPATH)/bin:$(PATH)
	CGO_CFLAGS ?= $(shell $(GO) env CGO_CFLAGS)
endif

ifeq ($(IS_WINDOWS),yes)
	GOFLAGS := -v -buildmode=exe
	EXECUTABLE_EXT := .exe
endif

.PHONY: go-check
go-check:
	$(eval MIN_GO_VERSION_STR := $(shell grep -Eo '^go\s+[0-9]+\.[0-9]+' go.mod | cut -d' ' -f2))
	$(eval MIN_GO_VERSION := $(shell printf "%03d%03d" $(shell echo '$(MIN_GO_VERSION_STR)' | tr '.' ' ')))
	$(eval GO_VERSION := $(shell printf "%03d%03d" $(shell $(GO) version | grep -Eo '[0-9]+\.[0-9]+' | tr '.' ' ');))
	@if [ "$(GO_VERSION)" -lt "$(MIN_GO_VERSION)" ]; then \
		echo "GreekMilkBot requires Go $(MIN_GO_VERSION_STR) or greater to build. You can get it at https://go.dev/dl/"; \
		exit 1; \
	fi


.PHONY: tidy
tidy:
	$(eval MIN_GO_VERSION := $(shell grep -Eo '^go\s+[0-9]+\.[0-9.]+' go.mod | cut -d' ' -f2))
	$(GO) mod tidy -compat=$(MIN_GO_VERSION)

.PHONY: test
test: unit-test integration-test

.PHONY: unit-test
unit-test:
	@echo "Running unit-test $(GOTESTFLAGS) -tags '$(TEST_TAGS)'..."
	@$(GO) test $(GOTESTFLAGS) -tags='$(TEST_TAGS)' $(GO_TEST_PACKAGES)

.PHONY: integration-test
integration-test:
	@echo "Running integration-test with $(GOTESTFLAGS) -tags '$(TEST_TAGS)'..."
	@$(GO) test $(GOTESTFLAGS) -tags='$(TEST_TAGS)' $(GO_INTEGRATION_TEST_PACKAGES)


.PHONY: fmt
fmt:
	@(test -f "$(GOPATH)/bin/gofumpt$(EXECUTABLE_EXT)" || $(GO) install mvdan.cc/gofumpt@latest) && \
	"$(GOPATH)/bin/gofumpt$(EXECUTABLE_EXT)" -l -w .