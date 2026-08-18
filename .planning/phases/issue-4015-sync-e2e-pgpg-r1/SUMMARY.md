# Issue 4015 Sync Pipeline E2E — Summary

## Outcome

The PostgreSQL → PostgreSQL control proved the real source → warehouse → managed target path for 1,001 rows and independently matched sample `id=1001` as `sequence=10010, label="event-1001"`.

A second `incremental_upsert` run skipped acknowledged rows (`records_read=0`, `records_loaded=0`), left the target at 1,001 rows, and preserved the named sample. That matches the declared incremental polling contract and corrects a stale test expectation that demanded a full replay.

The same live harness then failed CDC process-death recovery with `invalid_checkpoint: PostgreSQL polling checkpoint mechanism is not resumable`. PostgreSQL → PostgreSQL is therefore reported `broken-with-evidence` overall for the release gate. Product code was not changed.

The three GitHub routes were not attempted after the control finding; no GitHub credential was read and no GitHub object was created. All task-owned PostgreSQL containers, anonymous volumes, capacity probes, and generated image tags were independently proven absent.

## Changed files

- Added a direct target content assertion and correct no-change incremental expectation to `internal/cli/postgres_transport_binary_integration_test.go`.
- Added durable GSD context, plan, TDD, verification, summary, and review evidence for the live run.

## Delivery status

Ready for a direct PR to `integration/4015-mvp-flat-r1` with `Refs #4015`, preserving the recovery finding for release-gate disposition.
