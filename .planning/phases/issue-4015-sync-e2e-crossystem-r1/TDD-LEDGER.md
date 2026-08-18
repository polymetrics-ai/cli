# TDD Ledger

This is a certification/test-only phase. Product behavior and defects are explicitly out of scope, so Red/Green tracks the missing executable live proof and the test assertions that make a no-op fail.

## PostgreSQL → GitHub

- Red: No run-specific authorized live proof connects a real PostgreSQL row, owned warehouse artifact, typed GitHub reverse action, independent destination read-back, second run, and 404 cleanup.
- Green: Pending live execution. The test must observe exact GitHub label content through a separate HTTP client, `incremental_upsert` skip behavior, one destination object after replay, and deletion to 404.
- Refactor: Test-local helper extraction only; no connector/runtime behavior change.

## GitHub → PostgreSQL

- Red: Existing live proof reads another repository and does not satisfy this task's fixture scope or named-sample cleanup chain.
- Green: Pending live execution. The independent GitHub source count must equal a separately queried PostgreSQL managed-table count; a named `pm-cert-` row must match by content before and after a second `full_overwrite` run.
- Refactor: Test-local helper extraction only.

## GitHub → GitHub

- Red: Existing flow coverage uses a provider double or a retained external proof target, not the authorized fixture repository with deletable objects and independent 404 cleanup.
- Green: Pending live execution. A uniquely prefixed release must be extracted through the warehouse, changed through a typed approved action, independently read, and remain singular/content-stable on scheduled replay.
- Refactor: Reuse the same approved job for the schedule slice; do not add direct provider hops.

## Flow and schedule

- Red: No run-specific installed scheduler payload currently proves a cross-system approved job fired, changed/read the destination, and persisted terminal schedule state.
- Green: Pending live execution. The exact installed crontab payload must run the persisted flow, return a prepared execution identity, and match the independently read destination and durable schedule inspection.
- Refactor: None beyond test fixture cleanup.

