# AGENTS.md

This is an independent Go repository for the Huginn Bot API. The workspace root
is not a Git repository. Run Git commands in this directory.

Huginn Messenger is the Git submodule `third_party/huginn-messenger`, tracking
`main`. `make core` and `make all` refresh it with
`git submodule update --init --recursive --remote --checkout` and compile its
C ABI shared library. Local core edits stop the update; do not discard them.
Use `make docker-build` or `make docker-up` to refresh main before creating the
Docker build context. For direct Docker commands, run `make core-update` first.
CI resolves main once per run and builds both images from the tested revision.
Do not commit generated shared libraries. Keep the `internal/core` adapter in
sync with the submodule's `bridge.go`.

Before submitting Go changes, run `go fmt ./...`, `go test ./...`, and
`git diff --check`. For C ABI or packaging changes, also run `make all`.
Run `make test-native` to test the adapter with a freshly compiled core.

Never log API tokens, private keys, message contents, TURN credentials, or paths
that may contain secrets. Runtime databases, uploads, `.env`, binaries, and
shared libraries must remain untracked.
