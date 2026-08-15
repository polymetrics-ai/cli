# TDD ledger

| ID | Guarantee | RED command/evidence | GREEN proof | Status |
| --- | --- | --- | --- | --- |
| R1 | prepared execution has a distinct payload-bound identity | `go test ... ./internal/app -run 'TestAuthorizedFlowAction...'` failed on absent `PrepareAuthorizedFlowAction`, `ExecutePreparedFlowAction`, identity and firing fields | pending | red |
| R2 | one durable single-consume grant per firing | the same focused app compile failed on absent `ExecutionGrantConsumedError` and execution boundary | pending | red |
| R3 | refusals do not consume or dispatch | cancellation/revocation/tamper tests compile against absent execution-grant refusal contract and assert unchanged write, receipt, and marker counts | pending | red |
| R4 | partial/ambiguous write cannot auto-replay | write-failure/reopen test compiles against absent durable consumption boundary | pending | red |
| R5 | terminal schedule state precedes cleanup | schedule ordering recorder requires durable success/park before lock removal | pending | planned |
| R6 | unavailable shared coordination has the named SDK error/code | focused connsdk/engine/GitHub-hook run failed on absent `RateBudgetRefusalError`, `RateBudgetRefusalCode`, and `RateBudgetRefusalSharedCoordinatorUnavailable`; hook test retains its zero-send assertion | pending | red |
| R7 | production entry point reaches the components | fresh-binary test requires safe prepared/grant receipt fields produced through `cmd/pm` | pending | planned |

## Planned RED commands

```sh
go test -count=1 -timeout 20m ./internal/app ./internal/schedule ./internal/cli \
  -run 'Test(AuthorizedFlowActionPrepared|AuthorizedFlowActionGrant|ScheduleFirePerUseGrant|PMBinaryExecutesScheduled)'
go test -count=1 -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine/... ./internal/connectors/hooks/github/... \
  -run 'Test(RateBudgetRefusal|RequireShared|GitHubWriteHookCreateLabel)'
```

## RED evidence captured

2026-08-15, before production edits:

```text
internal/app/flow_action_test.go:218:21: a.PrepareAuthorizedFlowAction undefined
internal/app/flow_action_test.go:279:20: undefined: app.ExecutionGrantConsumedError
internal/connectors/connsdk/rate_budget_refusal_test.go:12:18: undefined: connsdk.RateBudgetRefusalError
internal/connectors/connsdk/rate_budget_refusal_test.go:13:19: undefined: connsdk.RateBudgetRefusalSharedCoordinatorUnavailable
```

The full focused command also compiled the engine and GitHub hook assertions against
the missing named SDK contract; no unrelated suite was run.
