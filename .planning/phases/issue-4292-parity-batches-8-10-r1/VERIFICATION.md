# Issue #4292 — verification checklist

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

## Completed generated/bundle checks

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
  — `552 connector(s) checked, 0 findings`.
- [x] `go run ./cmd/connectorgen surface-sync --check`
  — `552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected
  across 0 connector(s)`.
- [x] `git diff --check`.
- [x] `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/conformance ./internal/connectors/commandrunner`.
- [x] `go test -timeout 20m ./internal/cli`.
- [x] `go build ./cmd/pm` and `go vet ./...`.
- [x] `make tidy-check`, `make lint`, `make docs-check-no-build`,
  `make smoke-no-build`, `make agent-contract-check`, and
  `make release-workflow-check`.
- [x] `go run ./cmd/connectorgen boundary . --json`.

## Pending final gates

- [ ] Rebase to current `main` (including PR #4297), regenerate/rerun the
  declaration integrity checks as needed, then repeat the relevant final
  checks.
- [ ] Review, commit, push, and open the direct PR.
