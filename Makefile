# pio-scaffold Go refactor

BINARY = pio-scaffold
MODULE = github.com/runtime-terror404/pio-scaffold

.PHONY: build test lint cover install clean

build:
	go build -o $(BINARY) ./cmd/cli

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

cover:
	go test ./... -race -coverprofile=cover.out
	go tool cover -func cover.out | tail -1

install: build
	install -Dm755 $(BINARY) $(HOME)/.local/bin/$(BINARY)

clean:
	rm -f $(BINARY) cover.out
