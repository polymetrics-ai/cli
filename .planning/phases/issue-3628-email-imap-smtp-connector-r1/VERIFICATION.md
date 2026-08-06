# VERIFICATION — issue #3628 Email IMAP/SMTP connector

Status: scoped implementation validation complete; no-mistakes/CI/PR gates remain for firstmate.

## Dependency and binary evidence

- Approved module: `github.com/emersion/go-imap/v2@v2.0.0-beta.8`; `go mod verify` passed.
- Pre-dependency `pm` build: 93,216,322 bytes.
- Final `go build ./cmd/pm` measurement: 94,072,354 bytes.
- Delta: +856,032 bytes (about 0.82 MiB). This exact measured value belongs in the PR body.

## Local validation evidence

- `go test ./internal/connectors/native/email ./internal/connectors/native/nativeset ./internal/connectors/bundleregistry ./internal/app ./internal/cli -count=1` passed.
- `TestMailboxListHonorsRequestedLimit`, the full Email native suite, the Email credential-constraint test, and the CLI help/manual golden tests were rerun after the inline review fix and passed.
- `go test ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passed: 551 connectors, 0 findings.
- `go run ./cmd/connectorgen surface-sync --check` passed: 551 connectors, 0 corrections.
- `go run ./cmd/connectorgen boundary . --json` passed.
- `go vet` on the changed connector, registry, app, and CLI packages passed; `make lint` passed with 0 issues.
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors` passed; website generator-script tests passed.
- `go test ./internal/cli -run '^TestGoldenTranscripts$' -count=1` passed after the supported golden-fixture regeneration for the new dynamic connector and connector manual.
- `go run ./cmd/pm email --help`, and each of `mailboxes list` and `message send` with `--help`, rendered their declared command surfaces.
- `go run ./cmd/pm connectors list --all --json | jq '.connectors | length'` returned `555`.
- `git diff --check`, `go run ./cmd/agentcontractgen check`, and the release-workflow assertion passed.

## Scope statement

Development used only in-memory/local IMAP and SMTP protocol doubles and temporary attachment
files. No live mailbox, credential, or externally visible SMTP delivery was used. The local SMTP
double receives one message only in the approval-gate test; preview and attachment-drift tests
prove zero dispatch. Full `go test ./...` and `make verify` are intentionally left to CI/no-mistakes
because they exceed this worker's per-command timeout; `make smoke-no-build` was not run because it
creates an external temporary project and is unrelated to this native protocol connector.

## PR delivery limitation

- Email message reads and full refresh are blocked pending #3810 because
  `internal/app/app.go:350-376` does not validate catalog sync modes and
  `internal/app/app.go:543-551` forwards persisted cursor state. Both become available when #3810 lands.
- Sparse UID continuation is blocked pending #3810 because `internal/app/types.go:40-47` lacks
  scan-continuation state and `internal/app/local_warehouse.go:246-256` persists only an emitted
  cursor. It becomes available when #3810 lands.
