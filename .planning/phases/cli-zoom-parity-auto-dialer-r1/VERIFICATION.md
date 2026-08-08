# Verification Checklist — Zoom Auto Dialer parity, R1

## Lifecycle

- [x] GSD command provenance was resolved with `scripts/gsd sources`; the inline manual-GSD
  fallback and required skills are recorded in `PLAN.md`.
- [x] The live Auto Dialer artifact URL, retrieval date, HTTP result, byte count, SHA-256, and
  exact sixteen-operation audit were recorded before RED.
- [x] Test-only RED was captured verbatim before production declaration changes and pushed in
  `6c92cde24`.
- [x] No engine foundation was needed: existing fixed typed path/body contracts, status-only
  handling, approval lifecycle, and redaction cover the full category.
- [x] Connector declarations, fixtures, generated output, docs, website catalog, inline
  verify-work, and manual code-review evidence are complete.

## Source parity

- [x] All sixteen live-artifact endpoints map to exactly one executable command: eight direct
  reads and eight direct writes. The method/path reconciliation set is complete.
- [x] `surface-reconcile --check --notes-contains provider_module=auto-dialer` is clean after
  reconciling exactly sixteen rows.
- [x] The global endpoint-ledger delta contains eight new `rest_read` entries under `zoom`; the
  SHA-256 hash of all non-Zoom entries remains
  `7170208799055da56fb04ddcdc642fec68495e138808a41904e7f37cf8837691`.
- [x] Zero Zoom rows are `unsafe_or_disallowed`, and no Auto Dialer row remains locally blocked.
- [x] No response-only paging, date, token, or guessed query value is exposed as a request flag.
- [x] The four documented no-content mutations use status-only success; both DELETE actions require
  destructive confirmation.
- [x] Calls, call lists, prospects, contacts, companies, communications, transcripts, reports,
  identifiers, and generic token values are redacted in fixture-verified preview/result paths.

## Runtime/docs checks

- [x] Focused Zoom, engine, commandrunner, connectorgen, CLI, and vet checks pass; `make lint`
  reports no issues.
- [x] A fresh binary passes `pm help zoom`, bare `pm zoom`, bare `pm zoom auto-dialer`, and every
  exact Auto Dialer command `--help`; all sixteen routes are reachable.
- [x] Isolated fixtures prove exact fixed method/path/auth/body/status, no invented query/paging
  input, output redaction, plan/preview/approval execution, and destructive deletion confirmation.
- [x] Surface sync/reconciliation/validation, docs validation, generated website data twice,
  website typecheck/lint, endpoint-ledger scope checks, and scoped Make gates pass.

## verify-work evidence

```text
go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
go test -count=1 -timeout 20m ./internal/connectors/commandrunner ./internal/connectors/engine ./cmd/connectorgen
go test -count=1 -timeout 20m ./internal/cli
go vet ./internal/connectors/commandrunner ./internal/connectors/engine ./internal/connectors/defs/zoom ./cmd/connectorgen ./internal/cli
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen surface-reconcile --check --notes-contains provider_module=auto-dialer
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
make tidy-check
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
make lint
cd website && pnpm run gen:website-data && pnpm run gen:website-data
cd website && pnpm run typecheck && pnpm run lint
go build -o .tmp/pm ./cmd/pm
```

All commands passed. Website lint emitted 13 pre-existing warnings and no errors; none is in the
Zoom-generated catalog files. Full-repository `go test ./...` and `make verify` remain CI-owned per
`AGENTS.md` because their normal runtime exceeds the per-command worker window.

## Manual code-review evidence

The canonical runtime cannot register this provider-category phase and the parent contract forbids
spawning roles, so `code-review` was performed inline after resolving its GSD source. Review
confirmed that each command binds one audited method/path, input remains fixed and typed, mutations
require plan/no-network preview/approval, status-only actions do not invent bodies, and no generic
HTTP, shell, header, URL, or JSON escape hatch exists. Review also corrected the batch-update
ID-only no-op case with `minProperties: 2` and expanded source-specific response redaction. Final
inspection found no blocking issue; generated non-Zoom scope hashes and connector-boundary remain
clean.
