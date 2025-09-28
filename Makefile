# Makefile (taruh di /var/www/bea-cukai-app/bea-cukai-backend/Makefile)

SHELL := /bin/bash

APP_NAME   := bea-cukai-backend
MAIN_PKG   := ./         # main.go ada di root project
BUILD_DIR  := bin
BIN        := $(BUILD_DIR)/$(APP_NAME)

GO        ?= go
GOFLAGS   :=
LDFLAGS   := -s -w \
	-X 'main.version=$(shell git describe --tags --always 2>/dev/null || echo dev)' \
	-X 'main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo nogit)' \
	-X 'main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'

# File ENV. Ubah jika kamu pakai lokasi lain (mis. .env di project)
ENV_FILE  ?= ./env

.PHONY: all deps build build-linux-amd64 build-linux-arm64 run fmt vet test verify clean print-port

all: build

deps:
	$(GO) mod tidy
	$(GO) mod verify

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o $(BIN) $(MAIN_PKG)

# Cross-compile bila perlu
build-linux-amd64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o $(BIN)-linux-amd64 $(MAIN_PKG)

build-linux-arm64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o $(BIN)-linux-arm64 $(MAIN_PKG)

# Run lokal; akan source ENV_FILE kalau ada.
# Default port fallback ke 8787 bila PORT tidak diset di env.
run: build
	@echo "Using ENV_FILE=$(ENV_FILE)"
	@set -a; [ -f "$(ENV_FILE)" ] && . "$(ENV_FILE)"; set +a; \
	echo "Running $(BIN) on PORT=$${PORT:-8787}"; \
	PORT=$${PORT:-8787} $(BIN)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

verify: fmt vet test

clean:
	rm -rf $(BUILD_DIR)

print-port:
	@set -a; [ -f "$(ENV_FILE)" ] && . "$(ENV_FILE)"; set +a; \
	echo PORT=$${PORT:-8787}
