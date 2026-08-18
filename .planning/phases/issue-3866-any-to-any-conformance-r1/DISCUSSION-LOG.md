# Issue #3866 discussion record

`scripts/gsd prompt discuss-phase 3866` was resolved inline because this task's issue and delivery contract already fix every product decision and prohibit role spawning.

| Decision | Source | Result |
| --- | --- | --- |
| Evidence boundary | #3866 boundary and launch brief | Fakes and contract fixtures only; no production connector registration, provider/database I/O, generic protocol executor, certification matrix, or capability flag. |
| Matrix axis | #3866 and merged PR #4195 | Exercise four source/destination **families**, not four production connector routes. State the non-overlap in the PR. |
| Dependency status | `integration/4015-mvp-flat-r1` history | Treat #3864, #4035/#4131, #3865/#4138, #3867/#4152, and #4072/#4197 as delivered behavior despite stale open issue state. |
| Mode truth | base commits #4187 and #4188 | Cover executable `incremental_dedupe_history`; do not make it fail to match stale prose. |
| Failure proof | #3866 acceptance | Deliberately make a schema-valid family binding wrong after compilation, capture the named red result, restore, and capture green. |
