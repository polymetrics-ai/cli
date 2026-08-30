# TDD ledger — #4410 Sentry source-to-seven-lane matrix

| Stage | Evidence | Result |
| --- | --- | --- |
| Red | `go test ./internal/connectors/defs/sentry -run TestSentrySourceLaneMatrix -count=1` after adding the test but before creating the matrix. | Expected failure: all nine focused tests failed only because `sources/sentry-source-lane-matrix.json` was absent. |
| Green | Materialize source-lock-bound matrix and validate all source rows/cells/facts/backlinks. | Pass: `TestSentrySourceLaneMatrixContract` validates all 223 rows and 1,561 cells. |
| Edge | Mutate decoded matrix in memory for hidden IDs, duplicate IDs, invalid backlinks, missing source facts, pageable/extractable ETL, mutation reverse-ETL, and count mismatch. | Pass: all adversarial tests reject the altered in-memory matrix. |
| Refactor | Normalize citations/reasons, add source URL to each cited cell, tighten backlink-gap coverage, and gofmt. | Pass: every cell records pinned URL and source location; all facts and reasons are source-only. |
