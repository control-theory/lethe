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

# Cross-compile for all target platforms
build-all: clean
	GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-darwin-arm64      ./cmd/lethe
	GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-darwin-amd64      ./cmd/lethe
	GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-linux-amd64       ./cmd/lethe
	GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/lethe-windows-amd64.exe ./cmd/lethe

# Build MCPB bundle
bundle: build-all
	mkdir -p bundle/server
	cp bin/lethe-darwin-arm64      bundle/server/
	cp bin/lethe-darwin-amd64      bundle/server/
	cp bin/lethe-linux-amd64       bundle/server/
	cp bin/lethe-windows-amd64.exe bundle/server/
	@echo "Bundle ready in bundle/. Run 'mcpb pack bundle/' to create .mcpb file."
