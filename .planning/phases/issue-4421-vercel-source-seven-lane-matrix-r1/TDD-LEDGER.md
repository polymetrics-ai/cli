# TDD ledger — Vercel Track A

| Stage | Intended evidence | Status |
| --- | --- | --- |
| Red | `go test ./internal/connectors/defs/vercel -run '^TestVercelSourceLaneMatrixRetainsEveryLockedOperationAndLane$' -count=1` failed before the matrix existed: `read sources/vercel-source-lane-matrix.json: no such file or directory`. | observed |
| Green | The same command passes after the matrix exists. It validates 400 source IDs, seven cells each, exact source facts, all 22 boundary-only identities, 22 selected paging candidates, four legacy stream backlinks, three binary-upload candidates, and the single webhook gap. | passed |
| Edge | In-memory variants with a missing lane, hidden source row, changed legacy ETL backlink, dropped boundary record, or `implemented` disposition are rejected by named errors. | passed |
| Refactor | Deterministic ordering, error messages, and changed-path scope reviewed. No runtime or generator refactor was made. | passed |

The negative cases are deliberate local validator failures. They issue no provider request, use no credentials, and alter no runtime behavior.
