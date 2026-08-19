# Summary — PostgreSQL transport surface truthfulness

PostgreSQL now advertises the exact five modes that production preflight can resolve between its declared source and managed destination: `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`, and `incremental_dedupe`.

- Preserved the previously restored executable `full_overwrite` route.
- Removed only source-only `incremental_dedupe_history`; it has no shipped destination route.
- Regenerated the PostgreSQL certification shard through its scoped generator and regenerated the connector catalog through `pm docs generate`.
- Added production-composition happy, bad, and edge regression tests. The happy test observes three fixture records from a full-overwrite read; the bad test asserts the exact pre-I/O refusal; the edge test asserts the exact finite intersection.

GSD lifecycle prompts were resolved and executed inline under the documented single-worker fallback. Required review/verification artifacts are in this phase directory.
