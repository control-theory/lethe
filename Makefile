.PHONY: build test lint run clean build-all bundle

VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags="$(LDFLAGS)" -o bin/lethe ./cmd/lethe

test:
	go test ./...

lint:
	go vet ./...

run:
	go run ./cmd/lethe --stdio

clean:
	rm -rf bin/ bundle/server/

# Cross-compile for all target platforms (preserves bin/lethe)
build-all:
	GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-darwin-arm64      ./cmd/lethe
	GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-darwin-amd64      ./cmd/lethe
	GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-linux-amd64       ./cmd/lethe
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-windows-amd64.exe ./cmd/lethe

# Build MCPB bundle for the current platform.
# MCPB entry_point is "server/lethe" — one binary per bundle.
# To ship all platforms, build and pack separately per target.
bundle: build
	rm -rf bundle/server/
	mkdir -p bundle/server
	cp bin/lethe bundle/server/lethe
	@echo "Bundle ready. Run 'mcpb pack bundle/' to create .mcpb file."

# Convenience: bundle a specific cross-compiled target
bundle-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bundle/server/lethe ./cmd/lethe
	@echo "Bundle ready (darwin-arm64). Run 'mcpb pack bundle/'"

bundle-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bundle/server/lethe ./cmd/lethe
	@echo "Bundle ready (darwin-amd64). Run 'mcpb pack bundle/'"

bundle-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bundle/server/lethe ./cmd/lethe
	@echo "Bundle ready (linux-amd64). Run 'mcpb pack bundle/'"
