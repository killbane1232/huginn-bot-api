# AGENTS.md

This is an independent Go repository for the Huginn Bot API. The workspace root
is not a Git repository. Run Git commands in this directory.

Huginn Messenger is a versioned binary dependency declared in
`core-library.version`, not a Git submodule. `make core` downloads the matching
GitHub Release archive, verifies it against the published `SHA256SUMS`, and
extracts its exported C ABI shared library. Do not commit generated shared
libraries. Keep the `internal/core` adapter in sync with the dependency's
`bridge.go`.

Before submitting Go changes, run `go fmt ./...`, `go test ./...`, and
`git diff --check`. For C ABI or packaging changes, also run `make all`.

Never log API tokens, private keys, message contents, TURN credentials, or paths
that may contain secrets. Runtime databases, uploads, `.env`, binaries, and
shared libraries must remain untracked.
