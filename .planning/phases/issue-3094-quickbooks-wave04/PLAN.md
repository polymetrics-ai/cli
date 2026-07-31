# QuickBooks Wave 04 Parity Plan

Parent issue: #3094
Child issues: #3095, #3096, #3097, #3098, #3099, #3100, #3101
Branch: `fm/cli-quickbooks-parity-wave04-r1`

## GSD mode

Repo-local GSD adapter health was checked with `scripts/gsd doctor` and `scripts/gsd list`. The AGENTS.md-required `scripts/gsd prompt programming-loop ...` command is not registered in this checkout (`scripts/gsd: unknown GSD command: programming-loop`), so this slice uses the documented manual-GSD fallback while preserving the programming-loop lifecycle: plan, red tests, implementation, focused gates, broader gates, clean commit.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-cli`
- `golang-documentation`
- `golang-context`
- `golang-concurrency`
- `context-mode`

## Scope

- Re-audit the official QuickBooks sources named in #3094-#3101.
- Replace the existing coarse QuickBooks `api_surface.json` with an operation-ledger v1 file carrying the r2 official lane totals: 43 ETL/read, 60 reverse-ETL write, 32 direct/provider-query/search, 25 binary/file/report, 1 CDC/changefeed, total 161.
- Keep executable coverage truthful: the current five hook-backed read streams remain the only executable operations unless the existing connector-local contract can safely support more without shared-runtime changes.
- Add definition-owned CLI/certification metadata and update connector docs/generated surfaces.
- Keep all non-executable QuickBooks official operations blocked by default with source evidence and no generic raw query/HTTP/file/shell escapes.
- Preserve no-live-provider-call behavior and no credential requests.

## Non-goals / hard gates

- No live QuickBooks or OAuth calls.
- No credential collection or secret literals.
- No shared runtime behavior changes for operation executors, binary transfer, CDC, generic provider query, or reverse ETL.
- No certification claim; fixture-only remains uncertified.
- No push, PR, merge, or `/no-mistakes` invocation in this worker lane.

## Implementation slices

1. Red: add a focused QuickBooks API-surface ledger metrics test expecting operation-ledger v1, 161 rows, zero legacy exclusions, and r2 lane counts.
2. Green: regenerate `internal/connectors/defs/quickbooks/api_surface.json` from the official source inventory, preserving current five stream-covered rows and marking all other rows blocked/planned.
3. Green: add `cli_surface.json` and `certification.json` metadata; update `metadata.json`, `streams.json`, and connector docs to describe the complete ledger and fixture-only status.
4. Safety fix: keep `realm_id` as numeric non-secret config in the spec/declarative paths, while accepting legacy secret fallback in the connector-local hook with redacted hook errors.
5. Regenerate connector manuals, skill docs, website connector data, and golden/generated surfaces touched by the connector docs commands.
6. Stop after a clean local commit per the resumed-turn instruction; no push, PR, no-mistakes, live-provider, or credential work.

## Verification checklist

Required by the brief:

```bash
tmp=$(mktemp -d); cp -R internal/connectors/defs/quickbooks "$tmp/quickbooks"; go run ./cmd/connectorgen validate "$tmp"; rm -rf "$tmp"
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./internal/connectors/conformance -run 'TestConformance/quickbooks' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go build ./cmd/pm
make connector-boundary
make verify
git diff --check
```

Additional focused checks:

```bash
go test ./cmd/connectorgen -run TestQuickBooksAPISurfaceOperationLedgerMetrics -count=1
go test ./internal/connectors/hooks/quickbooks -count=1
```
