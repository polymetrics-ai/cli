# TDD ledger — #3977 pgoutput v2 CDC

| ID | Red behaviour | Green implementation | Status |
| --- | --- | --- | --- |
| PGC2-1 | Planned `ReadCDC` cannot satisfy v2 staged delivery. | PostgreSQL-private v2 executor starts only after preflight and a durable stage is opened. | Green: `green-v2-machine.txt`. |
| PGC2-2 | A streamed DML segment is invisible before `StreamCommit`. | Stage serializes chunks privately and releases them only through `CommitTransaction`. | Green: stream and non-stream boundary tests in `green-v2-machine.txt`. |
| PGC2-3 | `StreamAbort` must create no event, checkpoint, receipt, or LSN acknowledgement. | Abort removes the active private stage and leaves durable position unchanged. | Green: synthetic zero-counter assertion plus live rollback/LSN assertion. |
| PGC2-4 | Checkpoint/LSN acknowledgement before a durable whole-transaction receipt would lose data. | Receiver consumes all chunks; stage persists receipt; checkpoint then standby status follow in that exact order. | Green: synthetic strict `emit -> receipt -> checkpoint -> ack` assertion. |
| PGC2-5 | A live test must not skip when explicitly opted in with a valid direct local runtime. | dbtest PostgreSQL 14+ container asserts streaming, restart, lag/slot lifecycle, and teardown. | Green: `live-postgres-dbtest.txt`. |
| PGC2-6 | CDC cannot be advertised if the bundle and descriptor do not exactly match executable v2 behavior. | Promote the three runtime/declarative capability rows only after PGC2-5 passes. | Green: descriptor/bundle/capability tests plus `connectorgen validate` and `surface-sync --check`. |
