# TDD Ledger

This is a certification/test-only phase. Product behavior and defects are explicitly out of scope, so Red/Green tracks the missing executable live proof and the test assertions that make a no-op fail.

## PostgreSQL → GitHub

- Red: No run-specific authorized live proof connects a real PostgreSQL row, owned warehouse artifact, typed GitHub reverse action, independent destination read-back, second run, and 404 cleanup.
- Green: Live run observed exact GitHub label `pm-cert-db-api-e10940f636b8` through a separate HTTP client, `incremental_upsert` replay `0/0`, one destination object after replay, and deletion to HTTP 404.
- Refactor: Test-local helper extraction only; no connector/runtime behavior change.

## GitHub → PostgreSQL

- Red: Existing live proof reads another repository and does not satisfy this task's fixture scope or named-sample cleanup chain.
- Green: First delivery read and loaded all 10 independently counted GitHub labels; pgx read exactly 10 PostgreSQL rows and matched `pm-cert-db-api-e10940f636b8` by content.
- Finding: The standing-authorized `full_overwrite` replay reported `0/0` and replaced the target with zero rows. The expected second complete replacement was `10/10`; the named sample disappeared. The route is therefore `broken-with-evidence` even though its first delivery passed.
- Refactor: Test-local helper extraction only.

## GitHub → GitHub

- Red: Existing flow coverage uses a provider double or a retained external proof target, not the authorized fixture repository with deletable objects and independent 404 cleanup.
- Green: The exact-`since` source returned only task-created issue comment `5328121289`; the warehouse held exactly one row; approved `update_issue_comment` changed its body to `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`; independent GET and list reads found that one exact comment. Scheduled replay preserved the same body and singular ID.
- Refactor: Reuse the same approved job for the schedule slice; do not add direct provider hops.

## Flow and schedule

- Red: No run-specific installed scheduler payload currently proves a cross-system approved job fired, changed/read the destination, and persisted terminal schedule state.
- Green: The exact task-local installed crontab payload ran the persisted two-step flow. The sync read one comment, the action updated one comment, schedule inspection recorded terminal success and a non-empty prepared execution identity, and independent GitHub read-back matched comment `5328121289`.
- Refactor: None beyond test fixture cleanup.
