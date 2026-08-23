Refs #4316

## Summary

- Add a closed, declaration-owned named operation-route resolver shared by direct reads/writes,
  binary downloads/uploads, ETL, and reverse ETL.
- Fail unresolved, conflicting, or version-mismatched routes before provider I/O with a
  source-traced missing-foundation diagnostic.
- Enable and generate the five Help Scout Mailbox API v3 direct reads without changing the stored
  `/v2` connection default.

## GSD/TDD evidence

- Inline/manual GSD fallback: the project Pi adapter has no compatible isolated worker runtime;
  `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts
  were resolved and the artifacts record the fallback.
- Red: the Help Scout real-definition test failed for all five v3 operation IDs with zero provider
  requests. Green: it now captures each exact `/v3` request against a configured `/v2` fixture base.
- Required skills: golang-how-to, golang-cli, golang-testing, golang-error-handling,
  golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces,
  golang-context, golang-concurrency, golang-documentation.

## Verification

- `go test -timeout 20m ./internal/connectors/engine`
- `go test -timeout 20m ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1`
- `go run ./cmd/connectorgen validate`
- `go run ./cmd/connectorgen surface-sync --check`
- `npm --prefix website run gen:website-data`
- `go run ./cmd/agentcontractgen check`
- `go vet ./internal/connectors/engine ./internal/connectors/commandrunner`
- `go build ./cmd/pm`
- `git diff --check`

CLI parity: generated Help Scout command surfaces and website connector data are included. The
source-specific v3 endpoints have no new generic command, HTTP, URL, or credential surface.

Base verified: API-reported PR base must be `main`; branch ancestor `fed381e13` was checked before
implementation.
