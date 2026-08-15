GO ?= go
CORE_DIR := third_party/huginn-messenger
CORE_LIBRARY := build/libhuginn_messenger.so
BOT_BINARY := build/huginn-bot-api

.PHONY: all core bot test format clean

all: core bot

core:
	mkdir -p build
	cd $(CORE_DIR) && $(GO) build -ldflags='-checklinkname=0' -buildmode=c-shared -o ../../$(CORE_LIBRARY) .

bot:
	mkdir -p build
	CGO_ENABLED=1 $(GO) build -o $(BOT_BINARY) ./cmd/huginn-bot-api

test:
	CGO_ENABLED=1 $(GO) test ./...

format:
	$(GO) fmt ./...

clean:
	rm -rf build
