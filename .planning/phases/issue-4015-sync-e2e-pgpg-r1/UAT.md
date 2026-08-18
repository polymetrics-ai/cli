# Issue 4015 Sync Pipeline E2E — UAT

| Test | Result | Evidence |
| --- | --- | --- |
| Real PostgreSQL source reaches real PostgreSQL destination | pass | Direct `pgx` target query returned exactly 1,001 rows after the `pm` process exited. |
| Named source/destination sample equality | pass | Source formula yields `1001/10010/event-1001`; direct target query returned the same triple. |
| Bounded batching | pass | The 1,001-row source ran with `--batch-size 1000`, requiring more than one source page. |
| Second-run incremental behavior | pass | Second `incremental_upsert` run read/loaded `0/0`; direct target count and sample stayed unchanged. |
| Process-death recovery | fail | Restarted CDC child returned `invalid_checkpoint: PostgreSQL polling checkpoint mechanism is not resumable` before the new row reached the destination. |
| Task-owned fixture removal | pass | Docker events recorded volume/container destruction; exact run-image tags and containers were independently absent after cleanup. |

## UAT verdict

`gaps_found`. The basic control and no-change incremental path are live-proven, while CDC restart recovery is a release-gate finding.
