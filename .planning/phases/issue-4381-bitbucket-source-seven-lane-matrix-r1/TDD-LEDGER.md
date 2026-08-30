# TDD ledger — Bitbucket Track A

| Stage | Intended evidence | Status |
| --- | --- | --- |
| Red | `go test ./internal/connectors/defs/bitbucket -run '^TestBitbucketSourceLaneMatrix' -count=1` failed before the matrix existed: `read sources/bitbucket-source-lane-matrix.json: no such file or directory`. | observed |
| Green | The same focused command passed after the source-lock-bound matrix was added; it validates all 297 lock IDs, all seven cells, exact source facts, exact 34-item crosswalk boundary, source-only lane rules, and the four explicit webhook gaps. | passed |
| Edge | In-memory malformed matrices with a deleted cell, deleted source row, or deleted boundary identity are rejected with named errors by the same passing focused test. | passed |
| Refactor | Reviewed names, error messages, deterministic ordering, and changed path scope; no runtime or generator refactor is permitted. | passed |

The negative cases are deliberate validation failures asserted by the focused unit test. They do not make a provider request, use credentials, or alter runtime behavior.
