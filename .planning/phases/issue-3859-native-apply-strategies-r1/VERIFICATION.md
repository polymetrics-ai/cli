# Verification — #3859 native database apply strategies

## Planned checks

- [x] Focused engine/database/PostgreSQL/synctransport unit tests prove every
  added refusal has zero side effects and every success has a durable state.
- [x] Race test for each changed concurrent registry/session package.
- [x] Explicit Docker/Colima PostgreSQL dbtest run proves every strategy by
  resulting durable rows, not status alone.
- [x] `go vet` on changed packages; `gofmt`; `go build ./cmd/pm`.
- [x] All individual repository non-suite gates listed in `PLAN.md`.
- [x] `git diff --check` and final scope review confirm no #4125/#4136/#4090
  change, no public generic write surface, and no credential material.
- [x] CLI/docs/website parity reviewed and recorded as not applicable: the
  the final diff adds a user-visible surface.

## Result

Passed locally on 2026-08-15. The live PostgreSQL test observes all six
native apply strategies rather than only their completion: append and replace
row sets, upsert/dedupe source-order fences, explicit tombstone removal,
physical-absence retention, and history validity-window closure. The rejected
oversized polling page is also re-read from PostgreSQL to prove zero target
row mutation.

Commands: scoped connector/synctransport tests; race tests for engine,
database, and PostgreSQL; explicit Docker/Colima `databaseintegration` test;
`go vet`; `go build ./cmd/pm`; `tidy-check`, `lint` (0 issues after shared-lock
wait), `docs-check`, `smoke-no-build`, `agent-contract-check`,
`connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`,
and `release-workflow-check`.
