# Verification — reverse-ETL API-surface derivation r1

## Checklist

- [ ] New invariant demonstrably fails on the pre-fix defect and passes after regeneration.
- [ ] Focused test proves generic endpoint-summary derivation and preserves alias-like human summaries.
- [ ] GitHub has 214 recovered API-surface references and exactly 14 intentional empty aliases, named in the result.
- [ ] Generated sweep status buckets remain `1466 + 25 + 50 + 29 + 1 = 1571`, with zero delta per bucket.
- [ ] Surface, certification sweep, website docs/data, and tracked-skills generators are byte-stable on their second relevant pass.
- [ ] `go test -timeout 20m ./cmd/connectorgen`, `go vet ./...`, `go build ./cmd/pm`, and applicable individual repository gates pass or an exact external/base blocker is recorded.
- [ ] Inline `verify-work` and code review are complete and recorded.

## Commands and results

Pending implementation.
