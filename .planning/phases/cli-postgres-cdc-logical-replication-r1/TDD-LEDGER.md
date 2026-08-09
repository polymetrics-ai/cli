# TDD ledger — PostgreSQL logical-replication CDC

| ID | Red behaviour to prove | Green implementation | Status |
| --- | --- | --- | --- |
| PCDC-1 | A native reader alone cannot advertise CDC; a matching logical-replication executor can. | Implement `ChangefeedExecutorDescriptor` matching `changefeed.json`. | green |
| PCDC-2 | A scalar `lsn` state, missing durable committer, source mismatch, or uncommitted envelope cannot resume. | Consume only `synccontract.CheckpointEnvelope` and `ValidateResume`. | green |
| PCDC-3 | Slot naming is source-bound; incompatible, active, or wrong-plugin slots are refused; teardown removes the slot. | Add derived-name discovery/create/reuse/drop lifecycle. | green |
| PCDC-4 | Records are delivered before commit, then an LSN is durably committed before server acknowledgement; failure replays. | Decode `pgoutput` transactions and acknowledge only after durable checkpoint success. | green |
| PCDC-5 | A live PostgreSQL with `wal_level=logical` delivers insert/update/delete; restart receives the next transaction once; deferred teardown leaves no slot. | Real protocol integration test, environment-gated. | green — PostgreSQL 12.22 |
| PCDC-6 | Bundle metadata/docs/inspect/capability reflect exactly the working executor and source evidence. | Update PostgreSQL bundle and focused CLI discovery tests. | green |
| PCDC-7 | A multi-table publication cannot leak another relation; a missing selected relation, an active slot, an uncheckpointed existing slot, or a protocol-shape mismatch is refused. Full-table truncation must remain an explicit selected-stream change. | Filter `pgoutput` by canonical stream, validate publication membership before slot creation, preserve a checkpoint's native barrier, refuse unsafe slot reuse, and map truncate events. | green — focused unit tests and live PostgreSQL 12.9 conformance |

No test fixture, error assertion, log, test name, or planning artifact may contain a connection
string, password, or other credential value.
