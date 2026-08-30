# TDD ledger — GitLab Track A

| Stage | Intended evidence | Status |
| --- | --- | --- |
| Red | Focused GitLab source-matrix test fails while the matrix is absent. | passed — `read sources/gitlab-source-lane-matrix.json: no such file or directory` |
| Green | Exact source-lock, supplemental binary evidence, crosswalk boundary, seven-lane, paging, and mutation reconciliation pass. | passed — focused package test, 1,754 rows / 12,278 cells |
| Edge | Deliberate hidden-row, missing-cell, mutation-pair, boundary, and invalid-disposition variants fail. | passed — in-memory variants asserted in the focused test |
| Refactor | Deterministic matrix ordering and scoped changed paths are reviewed. | passed — deterministic source order, source snapshot identity, and Track A-only path review recorded in `REVIEW.md` |

No test invokes GitLab, uses credentials, or changes provider/runtime state.

## Semantic repair continuation — 2026-08-31

| Stage | Intended evidence | Status |
| --- | --- | --- |
| Red | The pre-repair matrix is evaluated with retained request-and-response continuation facts rather than a GET/name heuristic. It must fail because 255 request-control-only cells were previously claimed as ETL and bounded HEAD/POST reads were excluded. | passed — the focused real-matrix run reported `matrix source pagination reconciliation drift: computed states=map[not_documented_by_locked_operation:1493 request_controls_without_response_continuation:257 request_response_continuation_candidate:2]` before the corrected matrix was written. |
| Green | Semantic direct reads use retained successful response evidence: source-documented bounded HEAD metadata and query/search POSTs become `mapped_unproven` direct-read cells; semantic POSTs are not direct writes. | passed — three HEAD and thirteen POST rows are dynamically selected from the locked real matrix, source-lock backlinks are checked, and the final total is 763 direct-read candidates. |
| Bad | A source POST that does not meet the semantic read rule cannot be promoted to `direct_read`; a semantic POST read cannot be fabricated as `direct_write`/`reverse_etl`. | passed — both in-memory mutations fail `validateGitLabSourceLaneMatrix` on real source-derived expected cells. |
| Edge | A true continuation maps ETL, while a page/per-page request with no retained response continuation remains explicit non-ETL. | passed — two request/response pairs are `mapped_unproven`; 257 partial rows are `not_applicable`, including at least one page/per-page case. |
| Refactor | JSON rewrite retains only semantic changes, no fixed operation IDs, no source-lock/crosswalk/runtime/shared-code edits, and no pagination-derived sync. | passed — changed paths are GitLab matrix/test and #4384 evidence only; three pre-existing webhook registrations remain the only sync candidates. |

No continuation test invokes GitLab, uses credentials, or changes provider/runtime state.
