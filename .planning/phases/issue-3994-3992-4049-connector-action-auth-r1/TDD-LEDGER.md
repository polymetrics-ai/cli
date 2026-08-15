# TDD ledger

| ID | Guarantee | RED command/evidence | GREEN proof | Status |
| --- | --- | --- | --- | --- |
| R1 | prepared execution has a distinct payload-bound identity | focused app tests require new types/fields and payload drift identity inequality | pending | planned |
| R2 | one durable single-consume grant per firing | focused app/schedule tests require exactly one winner across replay/race/reopen | pending | planned |
| R3 | refusals do not consume or dispatch | cancellation/revocation/expiry/binding tests require typed error plus zero events/checkpoint and marker absence | pending | planned |
| R4 | partial/ambiguous write cannot auto-replay | connector failure after consume followed by reopen/replay requires zero second write | pending | planned |
| R5 | terminal schedule state precedes cleanup | schedule ordering recorder requires durable success/park before lock removal | pending | planned |
| R6 | unavailable shared coordination has the named SDK error/code | connsdk/engine/GitHub-hook tests refer to absent symbols and assert zero sends | pending | planned |
| R7 | production entry point reaches the components | fresh-binary test requires safe prepared/grant receipt fields produced through `cmd/pm` | pending | planned |

## Planned RED commands

```sh
go test -count=1 -timeout 20m ./internal/app ./internal/schedule ./internal/cli \
  -run 'Test(AuthorizedFlowActionPrepared|AuthorizedFlowActionGrant|ScheduleFirePerUseGrant|PMBinaryExecutesScheduled)'
go test -count=1 -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine/... ./internal/connectors/hooks/github/... \
  -run 'Test(RateBudgetRefusal|RequireShared|GitHubWriteHookCreateLabel)'
```

Actual failures will be recorded before production edits.

