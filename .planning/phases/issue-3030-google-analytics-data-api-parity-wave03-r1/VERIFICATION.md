# Verification checklist: Google Analytics Data API 24-operation parity

## Inventory and safety

- [x] Official v1beta and v1alpha discovery/reference re-audited on 2026-08-05 (revision `20260803`).
- [x] Counting policy applied: 24 semantic operations, 20 reads and 4 writes; v1alpha `getMetadata` and `runReport` deduplicated against v1beta.
- [ ] `api_surface.json` ledger-only commit names both artifacts, revision, retrieval date, and exact operation classification.
- [ ] No live provider calls, credentials, writes, or certification evidence introduced.

## Implementation and local gates

- [ ] New fixed v1alpha GET direct reads have red/green native tests and sanitized fixtures.
- [ ] Every 24-operation row has a correct executable mapping or evidence-backed #2985/reverse-ETL dependency.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/google-analytics-data-api`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/google-analytics-data-api' -count=1`
- [ ] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- [ ] `go build ./cmd/pm`
- [ ] `make connector-boundary`
- [ ] `make verify`
- [ ] `git diff --check`

## CLI, documentation, and delivery

- [ ] `pm help connectors`, `pm connectors`, and the GA connector help behavior are checked where generated surface changes apply.
- [ ] GA MANUAL/SKILL, all-connectors catalog, website generated data, and golden transcripts state the final inventory and fixture-only status.
- [ ] Captain-policy addendum is idempotently appended to #3030–#3037 using final actual counts only.
- [ ] GSD/manual-fallback, required skills, TDD, help/docs/website parity, review, no-mistakes, and CI-green evidence appear in the PR body.
