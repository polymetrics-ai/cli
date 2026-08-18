# TDD ledger — PostgreSQL apply/history club

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Adapter plans history | The admitted history page currently returns `database polling target plan: database write plan is invalid`; no target row or receipt is created. | The same adapter call returns a durable acknowledgement and the target query observes the new history row. |
| R2 | Route comes from definitions | A history adapter without a valid source definition, or with a non-PostgreSQL source/destination pair, cannot provide the required sealed route. | Loaded source/destination definitions produce the exact route; invalid route cells return typed reasons before write/session/ledger counters change. |
| R3 | Existing validity/replay semantics remain intact | The adapter cannot reach the existing close/insert logic, so no adapter-driven validity state exists to inspect. | Real PostgreSQL rows show one closed predecessor and one current successor; late/equal replay leaves the row set unchanged. |
| R4 | CDC delete closes through the adapter | The adapter cannot plan the history page carrying the CDC-derived tombstone. | The tombstone leaves the historical row stored but closes `_valid_to` and clears `_is_current`. |
| R5 | Receipt-before-ack survives restart | The adapter currently returns no acknowledgement because planning fails. | Acknowledgement is returned only after the persisted receipt; a fresh driver/ledger/executor reads the durable evidence and unchanged rows. |
| R6 | Existing forbidden CDC routes remain non-I/O | Unsupported route fakes must not be contacted. | Existing preflight/route tests return typed refusals with zero read/send/session/ledger counters. |

## Red command

```sh
go test -timeout 20m -count=1 ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres -run 'Test.*(DatabasePolling.*History|History.*Adapter)'
```

**Red:** The focused live PostgreSQL test reached the registered database
polling adapter and failed before its first write with
`database polling target plan: database history route source and destination
do not match the declared managed-target driver`. The returned acknowledgement
was zero, so the adapter could not create an observable history row or durable
receipt. Exact output is retained in `traces/red-adapter-history-plan.txt`.

## Green commands

The focused, race, live, and static commands are listed in `PLAN.md` and will
be recorded here with observable results during execution.

**Green — adapter history:** `TestPostgresManagedTargetIncrementalDedupeHistoryLive`
passed after the registered source definition was sealed with the destination
definition. The test applied v1, v2, a restart replay, and a CDC-derived delete
through `DatabasePollingApplyExecutor`; after each call it required a changed
durable delivery ID, and its PostgreSQL queries proved the validity-window and
soft-delete state. See `traces/green-adapter-history-live.txt`.

**Green — route/CDC regressions:** the non-PostgreSQL history route matrix kept
its typed reason and zero begin/batch/commit/rollback/ledger mutation checks;
the tombstone mapping/CDC tests passed; and the live workset scenario read back
the retained omission and explicit delete effect. See
`traces/green-route-cdc-focused.txt`.
