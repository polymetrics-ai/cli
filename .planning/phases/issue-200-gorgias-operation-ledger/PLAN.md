# Plan: Gorgias Operation Ledger

Parent issue: #196  
Sub-issue: #200  
Parent PR: https://github.com/polymetrics-ai/cli/pull/229  
Parent branch: `feat/196-gorgias-cli-parity`  
Sub-issue branch: `feat/200-gorgias-operation-ledger`  
Connector: `gorgias`

## GSD command path

- Adapter health: `scripts/gsd doctor`.
- Planning prompt: `scripts/gsd prompt plan-phase "Issue #200 Gorgias operation ledger: account for all 114 official Gorgias operations as implemented stream/direct_read/write or blocked operation categories before production edits; preserve safety gates, no credentials, no raw generic write tools"`.
- Programming-loop prompt remains unavailable in this repo-local command set (`programming-loop` is not listed by `scripts/gsd list --json`); manual GSD universal runtime fallback is active and recorded here.

## Required skills loaded

- `gsd-core`
- `caveman`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-lint`

## Objective

Replace the partial Gorgias API surface with an operation ledger that accounts for every public operation captured from official Gorgias docs: executable rows only for existing streams, blocked metadata rows for planned direct reads, binary reads, reverse-ETL/admin/destructive operations, product-scope or auth/internal operations, and no legacy `excluded` rows.

## Allowed files for this slice

- `internal/connectors/defs/gorgias/api_surface.json`
- `cmd/connectorgen/gorgias_api_surface_test.go` for static ledger regression coverage
- `.planning/phases/issue-200-gorgias-operation-ledger/**`
- `.planning/phases/issue-196-gorgias-cli-parity/**` only for parent orchestration state updates after verification/PR work

## Out of scope

- Implementing new streams or schemas.
- Implementing direct-read or binary executors.
- Implementing write actions.
- Runtime help renderer, docs website, or generated CLI docs; #198 owns help/doc rendering.
- Credentialed Gorgias checks or live API calls.

## Current baseline

- Official captured operations: 114 (`GET:46`, `POST:23`, `PUT:27`, `DELETE:18`) from `.planning/phases/issue-197-gorgias-cli-surface-metadata/OFFICIAL-OPERATIONS.json`; the #197 capture reconciled this with the older parent-issue method taxonomy.
- Current `api_surface.json`: 11 rows, legacy `excluded` classifiers, no `operation_ledger_version`.
- Existing executable streams: `tickets`, `customers`, `messages`, `satisfaction_surveys`.
- Existing writes: none.
- Existing implemented direct reads: none.

## Implementation slices

### Slice 1 — red ledger regression

1. Add a focused static Go test for Gorgias API surface ledger invariants:
   - `operation_ledger_version == 1`.
   - exactly 114 endpoint rows.
   - official method counts match `GET:59`, `POST:17`, `PUT:22`, `DELETE:16`.
   - no legacy `excluded` rows.
2. Run the focused test and record the expected failure.

### Slice 2 — generate complete ledger

1. Load official operations from the #197 source capture.
2. Rewrite `internal/connectors/defs/gorgias/api_surface.json` using connector-relative paths (without `/api`).
3. Preserve `covered_by.stream` for the four existing implemented streams only.
4. Classify every other operation as blocked metadata with `operation` fields:
   - `direct_read` for planned bounded GET/detail/list reads.
   - `binary_read` for file/recording download/upload payload endpoints.
   - `sensitive_reverse_etl`, `admin_reverse_etl`, or `destructive_action` for mutation/admin/destructive operations.
   - `local_workflow`, `duplicate`, `deprecated`, or `disallowed` only when justified by docs/source capture.
5. Ensure all blocked operation rows include non-empty reason and source URL or notes.

### Slice 3 — validation and handoff

1. Run focused ledger test and connector validation.
2. Run relevant connector/conformance tests.
3. Update `SUMMARY.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and parent orchestration state.
4. Commit, push branch, and open a stacked PR to the parent branch if checks pass.

## TDD strategy

- Red: focused Go test fails against current 11-row legacy surface.
- Green: operation ledger has 114 rows, ledger mode enabled, no legacy exclusions, validation clean.
- Refactor: normalize reasons/source URLs and JSON formatting without changing executable runtime behavior.

## Verification checklist

Focused:

- [ ] `go test ./cmd/connectorgen -run GorgiasAPISurfaceOperationLedger`
- [ ] `jq empty internal/connectors/defs/gorgias/api_surface.json .planning/phases/issue-197-gorgias-cli-surface-metadata/OFFICIAL-OPERATIONS.json`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/gorgias'`
- [ ] `git diff --check`

Broader before sub-PR handoff:

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## CLI help/docs/website parity

This slice is metadata-only operation accounting. It intentionally does not make planned Gorgias commands executable and does not change runtime help rendering. #198 owns help/docs/website parity once metadata is available.

## Safety gates

- No secrets requested, printed, summarized, or stored.
- No credentialed connector checks.
- No reverse ETL execution.
- No destructive/admin external actions.
- No new dependencies.
- No generic shell, generic HTTP write, generic SQL write, direct_write, raw_api, or raw mutation escape hatches.
