.PHONY: build test lint run clean

VERSION ?= dev

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/lethe ./cmd/lethe

test:
	go test ./...

lint:
	go vet ./...

run:
	go run ./cmd/lethe --stdio

clean:
	rm -rf bin/
