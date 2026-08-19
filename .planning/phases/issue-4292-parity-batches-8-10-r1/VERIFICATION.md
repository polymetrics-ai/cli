# Issue #4292 — verification checklist

## Planned checks

- [ ] Per-batch red map-integrity assertion fails before its source artifacts
  are added, then passes after the complete batch map is present.
- [ ] JSON integrity assertion: source lock, crosswalk, and disposition IDs
  agree exactly; every row has one primary class and class totals agree; typed
  write actions are `direct_write`; every direct-write row carries the locked
  generic-destination foundation gap only in its reverse-ETL eligibility
  attribute.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/<connector> --json`
  for each changed connector.
- [ ] `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`.
- [ ] Targeted Go tests identified from the validation/generator code, with
  `-timeout 20m`.
- [ ] Repository generated-file/snapshot checks applicable to changed bundle
  metadata.
- [ ] `go run ./cmd/connectorgen boundary . --json` in detached capture,
  polling to a recorded exit result.
- [ ] Standard review of the final changed files and an automated review route
  recorded in the PR body.

## Results

### Batch 8 source artifacts

- PASS — pre-generation presence assertion failed as expected: all thirty
  Batch 8 source artifact paths were absent.
- PASS — `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/generate-parity-maps.mjs 8`
  generated ten connector triplets. The attempted browser retrieval of
  TestRail's official rendered reference was intercepted by Cloudflare; per
  the approved fallback its source lock, crosswalk, and disposition are
  explicitly skipped as `no-public-api-description`, with no fake SHA-256,
  bytes, or operation inventory.
- PASS — `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/verify-parity-maps.mjs 8`
  passed for brex, zoho-books, testrail, amplitude, posthog, metabase, dbt,
  looker, mode, and dremio. Its first execution correctly caught duplicate
  source identities in the Zoho Books inventory; a deterministic
  `api_surface` index suffix preserves every operation and the rerun passed.
- PASS — `go run ./cmd/connectorgen validate internal/connectors/defs/<connector> --json`
  for each Batch 8 connector: each result reported `connectors_checked: 1`,
  `findings: null`, and `warnings: null`.
- PASS — `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`:
  `552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)`.

Final scoped Go tests, repository generated checks, boundary capture, and
review run after batches 9 and 10 are complete.
