# TDD ledger — Bitbucket Track A

| Stage | Intended evidence | Status |
| --- | --- | --- |
| Red | `go test ./internal/connectors/defs/bitbucket -run '^TestBitbucketSourceLaneMatrix' -count=1` failed before the matrix existed: `read sources/bitbucket-source-lane-matrix.json: no such file or directory`. | observed |
| Green | The same focused command passed after the source-lock-bound matrix was added; it validates all 297 lock IDs, all seven cells, exact source facts, exact 34-item crosswalk boundary, source-only lane rules, and the four explicit webhook gaps. | passed |
| Edge | In-memory malformed matrices with a deleted cell, deleted source row, or deleted boundary identity are rejected with named errors by the same passing focused test. | passed |
| Refactor | Reviewed names, error messages, deterministic ordering, and changed path scope; no runtime or generator refactor is permitted. | passed |

The negative cases are deliberate validation failures asserted by the focused unit test. They do not make a provider request, use credentials, or alter runtime behavior.

## Semantic-source repair — 2026-08-31

| Stage | Intended evidence | Status |
| --- | --- | --- |
| Red | `go test -timeout 20m ./internal/connectors/defs/bitbucket -run '^TestBitbucketSourceLaneMatrixSemanticSourceRules$' -count=1` failed because the old `paginated_` prefix selector returned no pagination refs for `searchTeam`, `searchAccount`, and `searchWorkspace`. | observed |
| Red | The same test then failed with source-semantic direct/write/reverse-ETL backlink drift for the existing method-based cells and an ETL count of 70 rather than the source-contract-derived 73. | observed |
| Green | The validator resolves successful-response schema refs against the immutable source contract and requires string `next` plus array `values`; it classifies direct read and mutations from the provider-summary action, then the matrix backs every applicable cell to that exact action/summary. | passed locally |
| Edge | A real POST list read remains a direct read, a real create remains a mutation, a real noncontinuable list remains non-ETL, a synthetic `values`-only response is non-ETL, and an ETL candidate remains non-sync absent webhook evidence. | passed locally |
| Refactor | Removed fixed POST-operation and `paginated_` selectors. The classifier is bounded and deterministic; an unknown provider-summary action fails validation rather than silently changing lane membership. | passed locally |

## Mapping-policy contract repair — 2026-08-31

| Stage | Intended evidence | Status |
| --- | --- | --- |
| Red | The matrix decoder began validating its `mapping_policy`; the existing method/schema-name wording failed `TestBitbucketSourceLaneMatrixRetainsEveryLockedOperationAndLane` with `mapping policy does not match source-semantic lane rules`. | observed |
| Green | Policy text now names provider-summary action semantics for direct/read-versus-mutation decisions and retained-source structural `next` + `values` continuation for ETL. No lane cell, count, source lock, or runtime claim changed. | passed locally |
| Edge | Table-driven mutations of each direct-read, write/reverse-ETL, and ETL policy field fail with a named mapping-policy error. | passed locally |
