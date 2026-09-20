GO ?= go
CORE_DIR := third_party/huginn-messenger
CORE_LIBRARY := $(abspath build/libhuginn_messenger.so)
BOT_BINARY := build/huginn-bot-api

.PHONY: all core-update core bot test test-native docker-build docker-up format clean

all: core bot

core-update:
	@if [ -e "$(CORE_DIR)/.git" ] && [ -n "$$(git -C "$(CORE_DIR)" status --porcelain)" ]; then \
		echo "Go submodule has local changes; commit or stash them before building." >&2; \
		exit 1; \
	fi
	git submodule sync --recursive -- "$(CORE_DIR)"
	git -c "submodule.$(CORE_DIR).branch=main" submodule update --init --recursive --remote --checkout -- "$(CORE_DIR)"
	@echo "Huginn core main: $$(git -C "$(CORE_DIR)" rev-parse HEAD)"

core: core-update
	mkdir -p build
	cd $(CORE_DIR) && CGO_ENABLED=1 $(GO) build -ldflags='-checklinkname=0' \
		-buildmode=c-shared -o "$(CORE_LIBRARY)" .

bot:
	mkdir -p build
	CGO_ENABLED=1 $(GO) build -o $(BOT_BINARY) ./cmd/huginn-bot-api

test:
	CGO_ENABLED=1 $(GO) test ./...

test-native: core
	HUGINN_CORE_TEST_LIBRARY="$(CORE_LIBRARY)" CGO_ENABLED=1 $(GO) test ./...

docker-build: core-update
	docker build -t huginn-bot-api .

docker-up: core-update
	docker compose up --build

format:
	$(GO) fmt ./...

clean:
	rm -rf build
