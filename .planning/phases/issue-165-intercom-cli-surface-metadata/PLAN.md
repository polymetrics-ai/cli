# Plan: Intercom CLI Surface Metadata

Parent issue: #164
Sub-issue: #165
Parent branch: `feat/164-intercom-cli-parity`
Sub-issue branch target: `feat/165-intercom-cli-surface-metadata` → `feat/164-intercom-cli-parity` when stacked PR flow is available

## GSD Command Path

- Adapter preflight passed: `scripts/gsd doctor`, `scripts/gsd verify-pi`, `scripts/gsd list --json`.
- Requested programming-loop command is not registered in this checkout: `scripts/gsd prompt programming-loop init --phase issue-164-intercom-cli-parity --dry-run` returned `unknown GSD command: programming-loop`.
- Fallback command prompt generated: `scripts/gsd prompt quick --full "issue #164 Intercom CLI parity parent planning and issue #165 CLI surface metadata"`.
- Manual GSD fallback: plan → red validation/test → implement → green verification → refactor/summary.

## Required Skills Loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.

## Objective

Refresh Intercom connector metadata from the official Intercom OpenAPI 2.14 source and add CLI surface inventory that maps provider/product operations into safe `pm intercom` app intents without enabling raw generic writes or credentialed checks.

## Scope

Expected production edits:

- `internal/connectors/defs/intercom/metadata.json`
- `internal/connectors/defs/intercom/api_surface.json`
- `internal/connectors/defs/intercom/cli_surface.json`
- `internal/connectors/defs/intercom/docs.md`
- `cmd/connectorgen/intercom_api_surface_test.go` or equivalent data-contract test

No live credentials, no write execution, no new dependencies, and no runtime command dispatch changes in this slice.

## Source Baseline

Official OpenAPI source:

```text
https://raw.githubusercontent.com/intercom/Intercom-OpenAPI/main/descriptions/2.14/api.intercom.io.yaml
```

Expected official operation metrics:

- Total operations: 149
- Paths: 105
- Methods: GET 67, PUT 16, POST 47, DELETE 19
- Product taxonomy from issue prompt: 35 ETL stream candidates, 24 direct-read candidates, 8 binary/file candidates, 82 reverse-ETL write candidates.

## Slice Boundaries

This slice is metadata-only:

- Add a complete official `api_surface.json` inventory using operation-ledger mode.
- Keep existing 5 streams covered by their current stream names.
- Mark non-implemented operations as blocked-by-default metadata rows (`direct_read`, `binary_read`, `sensitive_reverse_etl`, `admin_reverse_etl`, `destructive_action`, `duplicate`, `deprecated`, or `disallowed`) for later #168-#171 implementation/refinement.
- Add `cli_surface.json` for implemented stream-backed commands and planned safe app intents.
- Do not add `writes.json` or direct-read dispatch in #165.

## TDD Plan

1. Add a failing Intercom API surface metrics test before refreshing `api_surface.json`.
2. Red command: `go test ./cmd/connectorgen -run TestIntercomAPISurfaceOperationLedgerMetrics -count=1` should fail against the current 10-entry `api_surface.json`.
3. Refresh Intercom metadata and CLI surface files.
4. Green commands:
   - `go test ./cmd/connectorgen -run TestIntercomAPISurfaceOperationLedgerMetrics -count=1`
   - `go run ./cmd/connectorgen validate internal/connectors/defs`
   - focused bundle load validation if needed.
5. Refactor only formatting/data generation issues; do not weaken safety validation.

## Verification Checklist

```bash
jq empty internal/connectors/defs/intercom/api_surface.json internal/connectors/defs/intercom/cli_surface.json .planning/phases/issue-165-intercom-cli-surface-metadata/RUN-STATE.json
go test ./cmd/connectorgen -run TestIntercomAPISurfaceOperationLedgerMetrics -count=1
go test ./internal/connectors/engine -run 'Intercom|CLISurface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/intercom' -count=1
```

Broader parent gates (`go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`) run before parent handoff or after a coherent green slice if runtime allows.

## CLI Help / Docs / Website Parity

Applies as metadata input, but this slice does not create runtime `pm intercom` dispatch. Exemptions to record:

- Runtime help (`pm help intercom`, `pm intercom`, `pm intercom --help`) deferred to #166 help renderer.
- `docs/cli/**`, website docs, generated help/manual artifacts deferred to #166 unless generated metadata checks require an update.
- `docs.md` connector docs updated in this slice to avoid overclaiming implemented writes/direct reads.

## Human Gates / Safety

- No credentials.
- No secrets in prompts, files, fixtures, or logs.
- No live connector checks.
- No write execution.
- No new dependencies.
- No generic shell, generic HTTP write, generic SQL write, unrestricted raw API, or credential tools.
- Destructive/admin/sensitive Intercom operations stay blocked-by-default until #171 adds typed reverse-ETL gates.
