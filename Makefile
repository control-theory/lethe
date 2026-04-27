PREFIX ?= /usr/local

.PHONY: build test lint run clean install uninstall build-all bundle

VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

# Code signing identity. Use "-" for ad-hoc (local dev), or set to your
# Developer ID for distribution: CODESIGN_IDENTITY="Developer ID Application: ..."
CODESIGN_IDENTITY ?= -

# Sign a macOS binary. No-ops on non-darwin or for non-darwin targets.
define sign
	@if [ "$$(uname)" = "Darwin" ] && file $(1) | grep -q "Mach-O"; then \
		codesign --force --options runtime --sign "$(CODESIGN_IDENTITY)" --timestamp $(1) && \
		echo "signed $(1)"; \
	fi
endef

build:
	go build -ldflags="$(LDFLAGS)" -o bin/lethe ./cmd/lethe
	$(call sign,bin/lethe)

test:
	go test ./...

lint:
	go vet ./...

run:
	go run ./cmd/lethe --stdio

install: build
	install -d $(PREFIX)/bin
	install -m 755 bin/lethe $(PREFIX)/bin/lethe
	$(call sign,$(PREFIX)/bin/lethe)

uninstall:
	rm -f $(PREFIX)/bin/lethe

clean:
	rm -rf bin/ bundle/server/ *.mcpb

# Cross-compile for all target platforms
build-all:
	GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-darwin-arm64      ./cmd/lethe
	GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-darwin-amd64      ./cmd/lethe
	GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-linux-amd64       ./cmd/lethe
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-windows-amd64.exe ./cmd/lethe
	$(call sign,bin/lethe-darwin-arm64)
	$(call sign,bin/lethe-darwin-amd64)

# Build MCPB bundle for the current platform, pack into .mcpb
bundle: build
	rm -rf bundle/server/ *.mcpb
	mkdir -p bundle/server
	cp bin/lethe bundle/server/lethe
	$(call sign,bundle/server/lethe)
	mcpb pack bundle/

# Convenience: bundle a specific cross-compiled target
bundle-darwin-arm64:
	mkdir -p bundle/server
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bundle/server/lethe ./cmd/lethe
	$(call sign,bundle/server/lethe)
	@echo "Bundle ready (darwin-arm64). Run 'mcpb pack bundle/'"

bundle-darwin-amd64:
	mkdir -p bundle/server
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bundle/server/lethe ./cmd/lethe
	$(call sign,bundle/server/lethe)
	@echo "Bundle ready (darwin-amd64). Run 'mcpb pack bundle/'"

bundle-linux-amd64:
	mkdir -p bundle/server
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bundle/server/lethe ./cmd/lethe
	@echo "Bundle ready (linux-amd64). Run 'mcpb pack bundle/'"
