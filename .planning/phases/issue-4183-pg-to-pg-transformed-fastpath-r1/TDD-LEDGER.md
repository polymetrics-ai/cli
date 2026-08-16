# TDD ledger — issue #4183

| ID | Class | Behavior and observable assertion | Red | Green |
| --- | --- | --- | --- | --- |
| R1 | Happy/live regression | Two-plus-page `full_overwrite` publishes rows from every page, not only the last. | Pending | Pending |
| R1b | Edge/live | Empty/single/multi-page fixtures preserve the exact expected final row set. | Pending | Pending |
| R2 | Happy | A neutral typed batch becomes a versioned segment and manifest entry under byte-credit bounds. | Pending | Pending |
| R2b | Bad | An over-credit/corrupt-manifest input returns its typed refusal before a bulk applier call. | Pending | Pending |
| R2c | Edge | Zero rows, a full 512 MiB boundary, and one byte over retain bounded accounting. | Pending | Pending |
| R3 | Happy | A normalized projection/type/expression/filter plan has stable canonical bytes and hash. | Pending | Pending |
| R3b | Bad | Invalid expression/type/hash drift fails pre-I/O with exact typed reason. | Pending | Pending |
| R3c | Edge | Nulls, unicode/whitespace names, replayed normalized plan, and minimum/maximum bounds stay deterministic. | Pending | Pending |
| R4 | Happy/live | Binary `CopyFrom` shadow publish produces transformed target rows, one receipt, then checkpoint. | Pending | Pending |
| R4b | Bad/live | A non-binary/row-apply route is refused before target mutation. | Pending | Pending |
| R4c | Edge/live | Post-receipt/pre-checkpoint failure resumes by receipt without duplicate apply. | Pending | Pending |
| R5 | Happy | Success persists source/transform/Parquet/COPY/publish/checkpoint phase data before cleanup. | Pending | Pending |
| R5b | Bad | Invalid authorization/plan/checkpoint rejects before extractor or applier I/O. | Pending | Pending |
| R5c | Edge | Cancellation, stale checkpoint, unit timeout/retry, constrained credits, and interleaved receipt replay are durable and bounded. | Pending | Pending |
| R6 | Happy/live binary | The tagged binary fixture proves exact transformed count/aggregates and reports MB/s + MiB/s from measured logical bytes. | Pending | Pending |
| R6b | Bad | Missing performance opt-in visibly skips; invalid direct container endpoint fails rather than downgrading. | Pending | Pending |
| R6c | Edge | Identity and transformed runs on the same source report transformation tax without scoring identity as the feature result. | Pending | Pending |

