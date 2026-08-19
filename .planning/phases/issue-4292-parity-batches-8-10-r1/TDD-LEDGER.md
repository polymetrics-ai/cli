# Issue #4292 — TDD ledger

## Red

- The captain's `SOURCE-LOCK-DEFECT.md` established the initial red state:
  maps used landing-page pins, omitted `counts.total`, bounded their result by
  `api_surface.json`, and reported self-referential `declared_percent`.
- The initial integrity assertion correctly rejected the old Batch 8 source
  locks as incomplete source evidence. It also caught the pre-fix
  per-method-count implementation error while the new source extractor and
  cross-artifact verifier were being introduced.
- A source map that put a typed write under `reverse_etl` is red: write
  endpoints are `direct_write`; reverse-ETL is only the destination-executor
  eligibility attribute.

## Green

- Added `extract-source-operations.go`, which pins every provider document's
  exact byte count and SHA-256, then extracts REST method/path operations from
  complete OpenAPI/Swagger documents or complete official rendered references.
- Regenerated all mapped `api_surface.json` files from that source inventory,
  retaining old bindings only when they resolve to a pinned operation. The
  generated report records prior and refreshed counts for every connector.
- `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/verify-parity-maps.mjs 8`
  passed for all Batch 8 connectors.
- `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/verify-parity-maps.mjs 9`
  passed for all Batch 9 connectors.
- `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/verify-parity-maps.mjs 10`
  passed for all Batch 10 connectors.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passed with
  `552 connector(s) checked, 0 findings`.
- `go run ./cmd/connectorgen surface-sync --check` passed with 552 scanned and
  zero fields changed.

## Refactor / review

- The source lock carries `counts.total`, per-method and per-kind counts,
  source-document pins, and coverage basis. Dispositions use
  `operations_found`; no generated summary has `declared_percent`.
- TestRail, Eventbrite, and Greenhouse are explicitly skipped with browser
  evidence and `counts.total: null`; Adobe Commerce is explicitly
  `dynamic-instance-dependent` with a pinned official rendered reference and
  no fabricated total.
- Every `direct_write` has the locked
  `generic-typed-destination-executor` reverse-ETL eligibility gap; no row is
  primarily classed `reverse_etl` and no `transport_binding` action exists.
