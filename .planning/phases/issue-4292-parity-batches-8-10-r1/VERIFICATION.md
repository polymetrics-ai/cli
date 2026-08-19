# Issue #4292 — verification checklist

## Reconciliation status — 2026-08-20

- [x] Existing PR #4301 retargeted through the GitHub API to
  `fm/cli-reverse-etl-destination-r1`; API read-back reported that exact base.
- [x] Foundation SHA `c6f03c937c1f4e516d339b48e8c2143726179fdf` merged as
  `16c047eaf` without history rewriting.
- [x] CI Verify failure root-caused to Lever source-map regeneration, then
  reproduced locally.
- [x] Lever focused repair: `go test -timeout 20m ./cmd/connectorgen -run
  '^TestLeverHiringAPISurfaceOperationLedger$'` and Batch 9 source integrity
  check pass.
- [ ] Generate and validate the 30-row seven-surface ledger and human summary.
- [ ] Add only faithful connector-local source/destination declarations and
  per-increment static evidence; do not claim generic App/CLI dispatch before
  the foundation integration is merged.
- [ ] Before final push: fetch and merge the latest foundation branch, prove
  its exact SHA is an ancestor, and exercise the installed App/CLI dispatch
  route without credentials.

## Source-first map checks

- [x] Red: initial maps rejected because their locks lacked complete provider
  source evidence, total counts, and independent coverage basis.
- [x] Green: Batch 8 integrity map check passes.
- [x] Green: Batch 9 integrity map check passes.
- [x] Green: Batch 10 integrity map check passes.
- [x] Every direct write is primary class `direct_write`; reverse ETL is only
  the `generic-typed-destination-executor` eligibility attribute.
- [x] All 30 old/new counts and source bases are recorded in
  `SOURCE-SURFACE-REPORT.md`.
- [x] Red/green: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/audit-reverse-etl-readiness.mjs --check`
  confirms direct-write action coverage without claiming a destination before
  #4303 (26 mapped connectors; 6,311 writes; 1,419 action-backed; 4,892 pending
  connector-local typed-action authoring; 4 unavailable/dynamic sources), and
  confirms all 1,419 action-backed source/action bindings in
  `REVERSE-ETL-TYPED-ACTION-INVENTORY.json` remain `not-declared` under the
  uniform #4303 foundation gap.

## Completed generated/bundle checks

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
  — `552 connector(s) checked, 0 findings`.
- [x] `go run ./cmd/connectorgen surface-sync --check`
  — `552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected
  across 0 connector(s)`.
- [x] After the reverse-ETL readiness artifacts: reran
  `go run ./cmd/connectorgen validate internal/connectors/defs` (552 checked,
  0 findings) and `go run ./cmd/connectorgen surface-sync --check` (exit 0;
  no surface drift reported).
- [x] `git diff --check`.
- [x] `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/conformance ./internal/connectors/commandrunner`.
- [x] `go test -timeout 20m ./internal/cli`.
- [x] `go build ./cmd/pm` and `go vet ./...`.
- [x] `make tidy-check`, `make lint`, `make docs-check-no-build`,
  `make smoke-no-build`, `make agent-contract-check`, and
  `make release-workflow-check`.
- [x] `go run ./cmd/connectorgen boundary . --json`.

## Pending final gates

- [x] Rebased to `origin/main` at `51dd6d468` (including PR #4297), then
  reran all three declaration integrity checks, connector validation,
  `surface-sync --check`, targeted Go tests, `go build`, lint, and the agent
  contract check. No map retains the superseded HEAD/pagination engine-gap IDs.
- [ ] Review, commit, push, and open the direct PR.
