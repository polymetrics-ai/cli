# Self-review — Issue #4421 Vercel semantic mapping repair

## Reviewed invariants

- The repair branch starts at the exact frozen review target `e4b8a4da8795c30ea4e6fd948bd98cd116b8d043`.
- Changed provider artifacts are restricted to Vercel’s matrix and connector-local test; evidence records live only in this phase directory.
- The source lock and crosswalk have zero diff.
- No source-operation ID table or blanket POST/HEAD acceptance was added. Non-GET reads require retained successful-response and semantic source facts.
- `readSessionFile` needs both binary-capable response media and a binary schema. Structured JSON media is rejected even if it carries a `format: binary` schema.
- `writeSessionFiles` needs both an explicit documented content type and binary payload wording. Bare gzip header text is rejected.
- Existing ordinary mutation POST coverage remains negative for direct reads and positive for direct writes.
- ETL and sync-transport selection logic and statuses are unchanged.

## Review outcome

No scoped blocker found in self-review. A **fresh independent review is still required** before any integration; no PR or merge is created by this repair task.
