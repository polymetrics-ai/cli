# TDD ledger — #3979 PostgreSQL gap-free bootstrap

| ID | Red behaviour | Green implementation | Status |
| --- | --- | --- | --- |
| PGB-1 | A snapshot taken before a logical-slot boundary can omit a concurrent committed change. | Create the slot with an exported snapshot and import that exact snapshot into the bounded source read. | Green — live. |
| PGB-2 | Delivering/acknowledging a barrier before snapshot materialization is durable could discard required WAL after a crash. | Require the snapshot receiver success, then persist the initial checkpoint, then send the barrier standby status. | Green — live snapshot/checkpoint failure injection. |
| PGB-3 | Reusing an uncheckpointed existing slot silently picks an unknown snapshot/change boundary. | Preserve typed `RequireRebootstrap`; no slot recreation or cursor fallback occurs. | Green — live explicit rebootstrap refusal. |
| PGB-4 | A change committed during the blocked snapshot can be lost or create duplicate final keys. | Replay pgoutput-v2 strictly from the exported slot consistent point and apply snapshot/change records through a keyed oracle. | Green — live concurrent mutation. |
| PGB-5 | A multi-row transaction can appear partially, and deletes can be erased rather than represented. | Reuse committed transaction staging and emit the existing delete CDC event as an explicit receiver tombstone. | Green — live transaction and inherited v2 receipt proof. |
| PGB-6 | Invalid source/timeline/publication/schema state could make an old slot look resumable. | Validate source and generation plus bootstrap relation/schema bindings before snapshot delivery or replication start. | Green — focused drift matrix. |
| PGB-7 | A nominal integration test can pass without exercising the handover. | PostgreSQL dbtest blocks the snapshot receiver, commits a concurrent mutation, and asserts final rows, tombstones, checkpoint, LSN acknowledgement, and slot teardown. | Green — live. |

## Red:

`go test -timeout 20m ./internal/connectors/native/postgres -run '^TestPostgresBootstrapCheckpointBindsBarrierAndSchemaFingerprint$'` initially failed to compile because the PostgreSQL source had no bootstrap metadata or checkpoint decoder. The exact failure is retained in `traces/red-bootstrap-checkpoint.txt`.

## Green:

- `TestPostgresBootstrapCheckpointBindsBarrierAndSchemaFingerprint` passes, including schema, publication, timeline, and system drift rebootstrap outcomes.
- `TestPostgresGapFreeBootstrapContainerHarness` blocks snapshot receipt while a real update/delete/insert commits, then proves the combined keyed state equals the live source.
- `TestPostgresBootstrapSnapshotFailureRequiresExplicitRebootstrap` injects both snapshot-receiver and initial-checkpoint persistence failures, proving no durable checkpoint or streaming starts and a retained slot refuses implicit reuse.
