# Verification — issue #4329, r2

## Status

Planning checkpoint complete. Production edits and green evidence are pending.

## Planned commands

- `go test -count=1 -timeout 20m ./cmd/connectorgen -run '<focused r2 tests>'`
- `go test -count=1 -timeout 20m ./internal/connectors/engine`
- `go test -count=1 -timeout 20m ./internal/connectors/commandrunner`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync --check`
- `go build ./cmd/pm`
- credential-free built-binary probe for every command newly implemented by this PR; expected outcome is the credential boundary, never a provider request
- applicable serial `make verify` gates, `go vet`, generated docs checks, and `git diff --check`

No user command is newly implemented by this planned shared artifact rule; the
binary-probe entry will be recorded as not applicable unless materialization
introduces commands on this branch.
