# Plan: Zendesk Advanced Query / Binary Engine

Parent issue: #156
Sub-issue: #162
Branch: `feat/162-zendesk-advanced-query-binary-engine`
Stack base: current Zendesk executable/help stack; prior lanes remain review-gated.

## GSD Command Path

- GSD health already verified with `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json`.
- Programming loop prompt remains unavailable: `scripts/gsd prompt programming-loop ...` returns `unknown GSD command: programming-loop`.
- Manual GSD/TDD fallback active per AGENTS.md.

## Required Skills Loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.

## Objective

Make Zendesk binary/file-like GET operations explicitly executable as bounded metadata reads instead of leaving them as a planned placeholder.

## Scope

- Add a safe `binary_manifest` direct-read output policy that never emits binary body bytes.
- Convert all Zendesk `binary_read`/`binary_download` ledger entries into typed CLI direct-read commands with path flags, API coverage, and metadata-only output.
- Update validation schema/enums, tests, docs, and operation ledger notes.
- Preserve bounded reads (`MaxDirectReadBytes`) and reject absolute URLs/path traversal via existing direct-read path resolution.

## Non-goals

- Do not write downloaded files to disk.
- Do not expose generic HTTP download/raw request commands.
- Do not run credentialed Zendesk checks.

## TDD Plan

1. Add red tests for a `binary_manifest` direct read that returns metadata without body bytes.
2. Add red bundle test asserting no Zendesk binary candidate remains planned.
3. Implement binary policy and regenerate Zendesk binary command coverage.
4. Run targeted and full gates.

## Verification

Targeted:

```bash
go test ./internal/connectors/engine -run 'DirectRead.*Binary|BundleLoadEmbeddedZendeskCLISurface' -count=1
go test ./internal/connectors/commandrunner -run 'DirectRead.*Binary|Zendesk' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
./pm zendesk binary show-attachment --help
```

Full:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```
