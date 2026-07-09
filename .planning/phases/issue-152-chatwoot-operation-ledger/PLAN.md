# Plan: Chatwoot Operation Ledger

Issue: #152
Parent issue: #148
Parent PR: https://github.com/polymetrics-ai/cli/pull/223
Branch: `feat/152-chatwoot-operation-ledger`
Base: `feat/148-chatwoot-cli-parity`

## Required skills used

- gsd-core
- golang-how-to
- golang-testing
- golang-documentation
- golang-security
- golang-safety

## Goal

Lock Chatwoot's official Swagger operation accounting with connector tests and explicit policy notes so every official operation remains classified as implemented stream/write coverage or blocked-by-default operation metadata.

## Scope

- Add a Chatwoot operation-ledger metrics test for total operations, unique paths, methods, covered rows, blocked operation rows, model counts, risk counts, and status counts.
- Assert ledger-mode safety invariants: exactly one of `covered_by` or `operation`, no legacy `excluded`, every operation blocked by default, reason present, duplicate rows identify `duplicate_of`, sensitive/admin/destructive/disallowed rows have `source_url` or notes.
- Tighten binary/multipart policy wording for the profile update row where the official operation includes multipart avatar/profile fields.
- Preserve #149/#150 executable coverage; do not add runtime command execution.

## Non-goals

- Do not add `pm chatwoot ...` command dispatch.
- Do not implement direct reads, binary transfer, or new reverse-ETL writes.
- Do not call live Chatwoot APIs or use credentials.
- Do not add dependencies.
- Do not expose raw generic HTTP/write, shell, or SQL-write tools.

## TDD plan

1. Red: add `TestChatwootAPISurfaceOperationLedgerMetrics` expecting the official 144 operations / 89 paths and the current model/method/risk/status distribution.
2. Green: update metadata wording only if the test exposes ambiguous binary/multipart policy gaps.
3. Refactor: share existing GitHub helper types/functions where possible; keep Chatwoot expectations explicit.

## Verification checklist

```bash
go test ./cmd/connectorgen -run ChatwootAPISurfaceOperationLedgerMetrics -count=1
go test ./cmd/connectorgen -run 'GitHubAPISurface|ChatwootAPISurface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/conformance -run 'TestConformance/chatwoot' -count=1
git diff --check
```

Full handoff before PR/integration:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

## Safety

Operation-ledger rows are metadata only and must remain blocked by default. Binary/file and destructive/admin surfaces stay blocked until later slices add bounded policies, typed schemas, previews, approval text, and confirmations.

## Manual GSD fallback

`scripts/gsd prompt quick --validate ...` generated the quick-task prompt. `scripts/gsd prompt programming-loop ...` is not registered in this adapter, so this phase uses the manual programming loop with explicit plan, red test, green implementation, refactor, verification, and summary artifacts.
