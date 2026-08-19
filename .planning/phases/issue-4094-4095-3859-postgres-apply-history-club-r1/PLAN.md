# Plan — PostgreSQL apply/history club

## Goal

Make the registered database polling adapter construct the required sealed
PostgreSQL history route, so the existing history and CDC delete machinery is
actually reachable through #3859's shared apply boundary.

## TDD slices

1. **Red — adapter planning.** Add a focused test that passes an admitted
   `incremental_dedupe_history` page through `DatabasePollingApplyExecutor`.
   Assert the current implementation returns an error wrapping
   `ErrDatabaseWritePlanInvalid` and that write/session/ledger mutation is zero.
2. **Green — definition-bound route.** Extend `DatabasePollingApplyConfig`
   with the immutable source database definition, validate it at construction,
   and seal `{Source: source.Driver(), Destination: target.Driver()}` only for
   history plans. Keep all other mode plans unchanged.
3. **Red/green — route integrity.** Add table-driven source/destination route
   tests through the adapter. Require `DatabaseWriteHistoryRouteError` with the
   exact source/destination/both reason and zero preview/session/ledger I/O.
4. **Live — observable history.** Route the first history record and the
   CDC-derived delete through the adapter in the tagged PostgreSQL test. Query
   exact validity/current state and read durable acknowledgement/receipt state;
   prove replay and restart do not regress it.
5. **Regression and review.** Run focused package/race/static checks, the full
   requested tagged suite, individual repository gates, `verify-work`, and
   `code-review`. Record #4158 separately if its named base failure remains.

## Scope guards

- Production ownership: `internal/connectors/engine/database_polling_apply.go`
  only unless the red test demonstrates an unavoidable adjacent contract gap.
- Test ownership: the matching engine test and existing PostgreSQL managed
  target integration test.
- Do not alter the history SQL, CDC decoder/mapping, shared rate limits, live
  control-assertion behavior, CLI/docs/website surfaces, dependencies, or any
  unrelated connector.
- All identifiers and route authority come from existing closed definitions.
  No raw SQL, relation expression, arbitrary HTTP, shell, or public write API.

## Verification commands

```sh
go test -timeout 20m -count=1 ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres
go test -race -timeout 20m -count=1 ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m ./...
go vet ./internal/connectors/engine ./internal/connectors/database ./internal/connectors/native/postgres
go build ./cmd/pm
```

Run applicable `make verify` gates individually per `AGENTS.md`: `tidy-check`,
`lint`, `docs-check`, `smoke-no-build`, `agent-contract-check`,
`connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`, and
`release-workflow-check`.

## Checkpoints

1. Commit the combined delivery header, plan, TDD ledger, and red evidence.
2. Commit the smallest green adapter implementation and focused proofs.
3. Commit final live/verification/review evidence, push only the working branch,
   open the explicit-base PR, and report the API-observed base.

## Execution deviation

The source definition already belongs to the exact registered source stored in
`ResolvedPollingWatermark`. Adding it to the target configuration would have
created a second, independently supplied source identity. The green slice
therefore uses an optional definition-provider interface on that resolved
source and checks its driver ID against the descriptor source identity before
sealing the full source/destination driver pair. This is narrower than the
planned configuration field, keeps non-history paths unchanged, and preserves
definition ownership without touching the #4093 transport registry.
