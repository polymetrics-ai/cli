# TDD Ledger: GitLab Full Operation Parity Follow-up (#78)

## 2026-07-09

- GSD prompt: `scripts/gsd prompt execute-phase issue-78-gitlab-full-operation-parity --tdd`
- Programming-loop fallback: `scripts/gsd prompt programming-loop ...` unavailable; manual universal GSD loop used.
- Required skills: gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-documentation.

## Red

- Added `TestGitLabAPISurfaceFullOperationParityMetrics` and `TestGitLabFullOperationParityCommandAndWriteMetrics`.
- Red run: `go test ./cmd/connectorgen -run TestGitLab.*FullOperationParity -count=1` failed because GitLab had only 8 covered endpoints/implemented commands, not the GitHub-style full parity scaffold.

## Green

- Added `scripts/gen-gitlab-parity.py`, modeled after `scripts/gen-github-parity.py`.
- Generated GitLab full parity scaffold:
  - 1,142 covered endpoint rows (1,144 official OpenAPI operations + `/users` compatibility row minus 3 deprecated operations);
  - 1,142 implemented command-surface commands;
  - 637 typed reverse-ETL write actions;
  - 497 operation-backed read/binary/HEAD commands;
  - 4 runnable ETL stream commands and 4 runnable bounded direct-read commands retained.
- Updated validators to allow read-only `HEAD` coverage for operation-backed direct-read metadata while runtime direct-read dispatch remains GET-only unless an operation executor exists.

## Verification Evidence

- `go test ./cmd/connectorgen -run 'TestGitLab.*FullOperationParity|TestGitHubAPISurfaceOperationLedgerMetrics' -count=1` ✅
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` ✅ (`connectors_checked=547`, `findings=0`)
- `go test ./internal/connectors/conformance -run 'TestConformance/gitlab' -count=1` ✅
- `go test ./internal/cli ./internal/connectors/commandrunner ./internal/connectors/engine -run 'GitLab|Operation|DirectRead|Write' -count=1` ✅
- `go vet ./...` ✅
- `go test ./...` ✅
- `go build ./cmd/pm` ✅
- `make verify` ✅

## Leaf help follow-up

- Red: `go test ./internal/cli -run 'TestGitLabCommandSurfaceLeafHelp' -count=1` failed because `pm gitlab project list --help` and write/direct-read leaf help attempted project/credential execution, while operation-backed help returned a blocked executor error.
- Green: added metadata-only connector leaf help rendering before preflight/project loading. Representative stream, direct-read, operation-backed, reverse-ETL write, and JSON help cases now pass.
