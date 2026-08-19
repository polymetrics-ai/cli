# TDD ledger — issue #4291

| Slice | Red | Green | Refactor / evidence |
| --- | --- | --- | --- |
| Batch 6 map | `test ! -f` for each batch-6 source lock and disposition ledger: pass before implementation, proving the acceptance artifact does not exist | Run the real ledger invariant checker against the ten generated connector directories | Inspect source URLs, one-to-one endpoint identity, documented DELETE count, six-class count, and reason vocabulary |
| Batch 7 map | `test ! -f` for each batch-7 source lock and disposition ledger: pass before implementation, proving the acceptance artifact does not exist | Run the same real checker against the remaining ten directories | Repeat provenance and coverage inspection; run connector generator validation and surface drift checks |
| Transport truthfulness | Before implementation, no issue-local ledger can evidence the corrected reverse-ETL state | Checker requires typed write actions to remain enabled `direct_write` while their `declaration.reverse_etl` attribute carries `generic-typed-destination-executor`, `internal/app/issue_label_warehouse_transport.go:85-95`, and the prescribed minimal change | Confirm no `transport_binding` or connector action is invented |
| Reachability truthfulness | The provisional map enabled existing streams even where no API-surface command or typed action bound the source operation | Checker requires every `enabled` row to carry `covered_by.direct_read`, `covered_by.direct_reads`, or `covered_by.write`; streams without such a binding are `declaration-pending` | Report per connector documented / enabled / commands / writes / deletes / ENABLED% |

The map is a declarative inventory artifact, not a change to runtime behaviour. The red/green checks therefore assert the observable artifact state that is absent before the map and present after it.
