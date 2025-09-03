GO ?= go
HAS_GO := $(shell hash $(GO) > /dev/null 2>&1 && echo yes)
TAG ?= $(shell git describe --tags --abbrev=0 HEAD)
DATE_FMT = +"%Y-%m-%dT%H:%M:%S%z"
BUILD_DATE ?= $(shell date "$(DATE_FMT)")

ifeq ($(HAS_GO), yes)
	GOPATH ?= $(shell $(GO) env GOPATH)
	export PATH := $(GOPATH)/bin:$(PATH)
	CGO_CFLAGS ?= $(shell $(GO) env CGO_CFLAGS)
endif



ifeq ($(IS_WINDOWS),yes)
	GOFLAGS := -v -buildmode=exe
	EXECUTABLE_EXT := .exe
endif

.PHONY: fmt
fmt:
	@(test -f "$(GOPATH)/bin/gofumpt$(EXECUTABLE_EXT)" || $(GO) install mvdan.cc/gofumpt@latest) && \
	"$(GOPATH)/bin/gofumpt$(EXECUTABLE_EXT)" -l -w .