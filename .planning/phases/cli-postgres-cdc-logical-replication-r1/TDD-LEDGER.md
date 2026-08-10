# TDD ledger — PostgreSQL logical-replication CDC (historical)

These rows record pre-containment implementation work. They are retained for
provenance and do not claim a current executable CDC path.

| ID | Red behaviour to prove | Green implementation | Status |
| --- | --- | --- | --- |
| PCDC-1 | A native reader alone cannot advertise CDC; a matching logical-replication executor can. | Implement `ChangefeedExecutorDescriptor` matching `changefeed.json`. | historical pass; current status planned |
| PCDC-2 | A scalar `lsn` state, missing durable committer, source mismatch, or uncommitted envelope cannot resume. | Consume only `synccontract.CheckpointEnvelope` and `ValidateResume`. | historical pass; retained code is non-executable |
| PCDC-3 | Slot naming is source-bound; incompatible, active, or wrong-plugin slots are refused; teardown removes the slot. | Add derived-name discovery/create/reuse/drop lifecycle. | historical pass; no current slot operation |
| PCDC-4 | Records are delivered before commit, then an LSN is durably committed before server acknowledgement; failure replays. | Decode `pgoutput` transactions and acknowledge only after durable checkpoint success. | historical pass; no current source acknowledgement |
| PCDC-5 | A live PostgreSQL with `wal_level=logical` delivers insert/update/delete; restart receives the next transaction once; deferred teardown leaves no slot. | Real protocol integration test, environment-gated. | historical pass — PostgreSQL 12.22; test now skipped |
| PCDC-6 | Bundle metadata/docs/inspect/capability reflect exactly the working executor and source evidence. | Update PostgreSQL bundle and focused CLI discovery tests. | historical pass; current capability is false |

No test fixture, error assertion, log, test name, or planning artifact may contain a connection
string, password, or other credential value.
