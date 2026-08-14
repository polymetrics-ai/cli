# TDD ledger — #3979 PostgreSQL gap-free bootstrap

| ID | Red behaviour | Green implementation | Status |
| --- | --- | --- | --- |
| PGB-1 | A snapshot taken before a logical-slot boundary can omit a concurrent committed change. | Create the slot with an exported snapshot and import that exact snapshot into the bounded source read. | Planned. |
| PGB-2 | Delivering/acknowledging a barrier before snapshot materialization is durable could discard required WAL after a crash. | Require the snapshot receiver success, then persist the initial checkpoint, then send the barrier standby status. | Planned. |
| PGB-3 | Reusing an uncheckpointed existing slot silently picks an unknown snapshot/change boundary. | Preserve typed `RequireRebootstrap`; no slot recreation or cursor fallback occurs. | Planned. |
| PGB-4 | A change committed during the blocked snapshot can be lost or create duplicate final keys. | Replay pgoutput-v2 strictly from the exported slot consistent point and apply snapshot/change records through a keyed oracle. | Planned. |
| PGB-5 | A multi-row transaction can appear partially, and deletes can be erased rather than represented. | Reuse committed transaction staging and emit the existing delete CDC event as an explicit receiver tombstone. | Planned. |
| PGB-6 | Invalid source/timeline/publication/schema state could make an old slot look resumable. | Validate source and generation plus bootstrap relation/schema bindings before snapshot delivery or replication start. | Planned. |
| PGB-7 | A nominal integration test can pass without exercising the handover. | PostgreSQL dbtest blocks the snapshot receiver, commits a concurrent mutation, and asserts final rows, tombstones, checkpoint, LSN acknowledgement, and slot teardown. | Planned. |
