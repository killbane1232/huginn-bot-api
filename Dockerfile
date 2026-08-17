# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS builder
WORKDIR /src

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl && \
    rm -rf /var/lib/apt/lists/*
COPY scripts/download-core-library.sh ./scripts/download-core-library.sh
ARG TARGETARCH
ARG HUGINN_CORE_VERSION
ARG HUGINN_CORE_RELEASE_BASE
ARG GOPROXY
RUN HUGINN_CORE_ARCH="$TARGETARCH" \
    HUGINN_CORE_VERSION="$HUGINN_CORE_VERSION" \
    HUGINN_CORE_RELEASE_BASE="$HUGINN_CORE_RELEASE_BASE" \
    GOPROXY="$GOPROXY" \
    HUGINN_CORE_OUTPUT_DIR=/out \
    ./scripts/download-core-library.sh

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -o /out/huginn-bot-api ./cmd/huginn-bot-api

FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates gosu && \
    rm -rf /var/lib/apt/lists/* && \
    useradd --create-home --uid 10001 --user-group bot && \
    install -d -o bot -g bot /app/data/uploads /app/lib
COPY --from=builder /out/huginn-bot-api /app/huginn-bot-api
COPY --from=builder /out/libhuginn_messenger.so /app/lib/libhuginn_messenger.so
COPY --chmod=755 docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
WORKDIR /app
ENV BOT_API_ADDR=:8081 \
    HUGINN_CORE_LIBRARY=/app/lib/libhuginn_messenger.so \
    HUGINN_DB_PATH=/app/data/huginn.db \
    BOT_API_UPLOAD_DIR=/app/data/uploads
EXPOSE 8081
VOLUME ["/app/data"]
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["/app/huginn-bot-api"]
