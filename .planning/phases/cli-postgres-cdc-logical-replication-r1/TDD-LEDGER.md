# TDD ledger — PostgreSQL logical-replication CDC containment

The original live protocol rows are retained as historical evidence only. The
current public contract is `planned` until the staged PostgreSQL 14+ design is
implemented.

| ID | Red behaviour to prove | Green implementation | Status |
| --- | --- | --- | --- |
| PCDC-1 | An incomplete native reader must not advertise CDC. | `ChangefeedExecutorDescriptor` and bundle stay `planned` until staged streaming exists. | Green: fail-closed capability tests |
| PCDC-2 | A scalar `lsn` state, missing durable committer, source mismatch, or uncommitted envelope cannot resume. | Consume only `synccontract.CheckpointEnvelope` and `ValidateResume`. | green |
| PCDC-3 | Slot naming is source-bound; incompatible, active, or wrong-plugin slots are refused; teardown removes the slot. | Add derived-name discovery/create/reuse/drop lifecycle. | green |
| PCDC-4 | A transaction must not publish or acknowledge at commit burst without a streamed whole-transaction receipt. | Red: unbounded at-commit accumulator review finding. Green: executor rejects before replication while stage is absent. | green — containment |
| PCDC-5 | A live PostgreSQL test must not falsely claim current conformance while the executor is planned. | Integration test skips before source access; historical PostgreSQL 12.22 result is marked historical. | green — containment |
| PCDC-6 | Bundle metadata/docs/inspect/capability reflect the runnable executor. | `cdc=false`, `changefeed.status=planned`, and CLI discovery tests reject PostgreSQL CDC. | green |
| PCDC-7 | A multi-table publication cannot leak another relation; a missing selected relation, an active slot, an uncheckpointed existing slot, or a protocol-shape mismatch is refused. Full-table truncation must remain an explicit selected-stream change. | Filter `pgoutput` by canonical stream, validate publication membership before slot creation, preserve a checkpoint's native barrier, refuse unsafe slot reuse, and map truncate events. | green — focused unit tests and live PostgreSQL 12.22 conformance |
| PCDC-8 | Valid non-ASCII text must survive decode and durable JSON representation byte-exact; malformed bytes must not reach a checkpoint boundary. | Red: invalid client/session or tuple bytes could normalize silently. Green: force/verify UTF-8, reject invalid tuple/origin/type payloads, and byte-round-trip non-ASCII text. | green |
| PCDC-9 | Replica identity modes that the decoder cannot map must never silently misattribute updates/deletes. | Red: `FULL` and `USING INDEX` were silently accepted. Green: admit DEFAULT only with a primary key; reject FULL, USING INDEX, NOTHING, and unknown modes with named errors. | green |
| PCDC-10 | A planned/at-commit reader must not be promoted on the strength of unit decoding alone. | Require PostgreSQL 14+, pgoutput protocol v2 streaming, a bounded crash-recoverable transaction stage, and a durable whole-transaction downstream receipt before source progress. | blocked — requires shared committed-transaction receipt foundation |
| PCDC-11 | A real public GitHub read and a PostgreSQL CDC test cannot be conflated while `Postgres.Write` is unsupported. | Prove GitHub public-read limiter behavior separately; load the same public dataset directly into isolated Docker/Colima PostgreSQL, then prove ordered insert/update/delete, Unicode, restart/resume, and slot teardown under `POLYMETRICS_INTEGRATION=1`. | split: GitHub green; the #4083 `dbtest` foundation now supplies the documented Docker/Colima route, while CDC remains blocked on PCDC-10's committed-transaction receipt seam |
| PCDC-12 | A quota breach, abort, or teardown cannot publish partial changes or leave a WAL-pinning slot. | `StreamAbort` discards the private stage, `TransactionStageLimitExceeded` acknowledges no LSN, and explicit teardown/rebootstrap reports slot health and removes the derived slot. | blocked — depends on PCDC-10 committed-transaction receipt seam |

No test fixture, error assertion, log, test name, or planning artifact may contain a connection
string, password, or other credential value.
