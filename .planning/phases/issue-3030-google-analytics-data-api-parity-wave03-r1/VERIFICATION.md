# Verification checklist: Google Analytics Data API 24-operation parity

## Inventory and safety

- [x] Official v1beta and v1alpha discovery/reference re-audited on 2026-08-05 (revision `20260803`).
- [x] Counting policy applied: 24 semantic operations, 20 reads and 4 writes; v1alpha `getMetadata` and `runReport` deduplicated against v1beta.
- [x] `api_surface.json` ledger-only commit `7c550c075` names both artifacts, revision, retrieval date, and exact operation classification.
- [x] No provider data calls, credentials, writes, or certification evidence introduced; test evidence is fixtures and local `httptest` only.

## Implementation and local gates

- [x] New fixed v1alpha GET direct reads have red/green native tests and sanitized fixtures.
- [x] Every 24-operation row has a correct executable mapping or evidence-backed #2985/reverse-ETL dependency.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/google-analytics-data-api`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/google-analytics-data-api' -count=1`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [ ] `make verify`
- [x] `git diff --check`

## CLI, documentation, and delivery

- [x] `pm help connectors`, `pm connectors`, and GA connector help behavior are covered by CLI golden tests; targeted command invocations will be repeated in final delivery verification.
- [x] GA MANUAL/SKILL, all-connectors catalog, website generated data, and golden transcripts state the final inventory and fixture-only status.
- [ ] Captain-policy addendum is idempotently appended to #3030–#3037 using final actual counts only.
- [ ] GSD/manual-fallback, required skills, TDD, help/docs/website parity, review, no-mistakes, and CI-green evidence appear in the PR body.
