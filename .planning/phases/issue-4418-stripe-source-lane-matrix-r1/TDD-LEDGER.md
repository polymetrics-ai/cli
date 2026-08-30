# TDD ledger — #4418 Stripe source-to-seven-lane matrix

| Stage | Evidence | Expected result |
| --- | --- | --- |
| Red | Add `TestStripeSourceLaneMatrixContract` before the matrix exists. | `go test ./internal/connectors/defs/stripe -run TestStripeSourceLaneMatrix -count=1` failed in all eight focused tests only because `sources/stripe-source-lane-matrix.json` was absent. |
| Green | Materialize source-lock-bound matrix and run focused test. | `go test ./internal/connectors/defs/stripe -run TestStripeSourceLaneMatrix -count=1` passed: every source row, cell, fact, disposition, and backlink validates. |
| Edge | Mutate decoded matrix in-memory in focused tests. | Hidden source row, invalid backlink, missing paging ETL/sync state, missing mutation reverse-ETL state, and source-count mismatch each fail with a specific diagnostic. |
| Refactor | Run `gofmt`; keep all validation test-local. | No production/runtime package is modified. |

The validator is intentionally local to the Stripe definition test. It consumes existing source lock and artifacts but neither imports nor changes shared execution/admission behavior.
