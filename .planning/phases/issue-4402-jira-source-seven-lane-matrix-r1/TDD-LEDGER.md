# TDD ledger — Jira Track A

| Stage | Intended evidence | Status |
| --- | --- | --- |
| Red | `go test ./internal/connectors/defs/jira -run '^TestJiraSourceLaneMatrixRetainsEveryLockedOperationAndLane$' -count=1` failed before the matrix existed with `read sources/jira-source-lane-matrix.json: no such file or directory`. | observed |
| Green | The same command passes after adding the source-lock-bound matrix. It validates all 617 source IDs, all seven cells, exact source facts, exact crosswalk reconciliation, all 95 `maxResults` candidates, the three stream backlinks, the four binary uploads, the three binary downloads, and the one webhook gap. | passed |
| Edge | In-memory matrices with a deleted lane cell, a hidden source row, a drifted legacy `streams.json` backlink, or an `implemented` direct-read disposition are rejected by named validator errors. | passed |
| Refactor | Names, deterministic ordering, error messages, and changed-path scope are reviewed. No runtime or generator refactor is permitted. | passed |

The negative cases are deliberate local validation failures asserted by the focused test. They make no provider request, use no credentials, and do not alter runtime behavior.
