GO ?= go
CORE_DOWNLOADER := scripts/download-core-library.sh
BOT_BINARY := build/huginn-bot-api

.PHONY: all core bot test format clean

all: core bot

core:
	HUGINN_CORE_OUTPUT_DIR=build $(CORE_DOWNLOADER)

bot:
	mkdir -p build
	CGO_ENABLED=1 $(GO) build -o $(BOT_BINARY) ./cmd/huginn-bot-api

test:
	CGO_ENABLED=1 $(GO) test ./...

format:
	$(GO) fmt ./...

clean:
	rm -rf build
