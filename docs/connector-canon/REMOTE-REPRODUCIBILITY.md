# Remote reproducibility

A clean clone can reproduce vNext authoring, deterministic rendering, JSON-only
runtime discovery, fake-server protocol behavior, and DuckDB saved-flow proofs
without provider credentials.

## Required environment

- Go and the toolchain selected by `go.mod`;
- CGO and a working C toolchain for DuckDB/Parquet;
- Git and standard POSIX shell utilities;
- no provider account, provider database, or mutable documentation fetch for
  deterministic checks.

## Clean-clone checks

```bash
go run ./cmd/agentcontractgen check
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -run '^(TestVNextSourceLock.*|TestVNextGenerationPublisher.*|TestRunLockRenderPublishesOnlyClosedGeneration)$' -count=1
go test ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
go test ./internal/cli -count=1
go build ./cmd/pm
```

The runtime-inventory test proves source locks and authoring evidence are not
embedded. The in-memory reference-render and temporary-root publication tests
prove deterministic closed-generation construction, journal recovery, exact
`CURRENT` selection, and lease-safe pruning without materializing the
checked-in corpus. Local fake servers prove request encoding and
credential-boundary reachability. Temporary DuckDB paths prove ETL,
reverse ETL, and configured sync transport without an external provider.

## External provider tests

Live provider tests are separate, optional operational evidence. They require
explicit authorization, scoped non-interactive credentials, bounded test data,
and cleanup. Their presence or absence never changes runtime admission. A
deterministic build must not fetch current provider documentation.

## Genuine infrastructure requirements

Native database and CDC live tests may require an explicitly configured local
database/container environment. Absence of that environment is reported as a
test precondition, not converted into connector runtime state. A missing shared
encoder, executor, or warehouse protocol is a Foundation Atlas gap and requires
approval before implementation.
