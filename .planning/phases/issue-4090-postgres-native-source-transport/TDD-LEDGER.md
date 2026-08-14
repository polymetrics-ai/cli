# TDD ledger — Issue #4090

| ID | RED failure | GREEN contract |
| --- | --- | --- |
| R1 | PostgreSQL has no declared transport source or preflight rejects it. | `Definition()` declares exactly the PostgreSQL `native_database` source and preflight resolves it. |
| R2 | A wrong-family definition is able to reach source I/O. | Family mismatch rejects before the instrumented source can open/query. |
| R3 | A declared but unregistered executor is able to reach source I/O. | Exact missing registration rejects before source invocation. |
| R4 | No descriptor reaches an implicit `Read()` fallback. | Absence rejects before source invocation; no capability bit substitutes for it. |
| R5 | Full modes can emit an unbounded, unordered, loosely typed record set. | Both full modes emit pages bounded by request and definition caps, with catalog-derived projection and stable ordering. |
| R6 | Output identity/checkpoint can vary or omit schema/source binding. | A valid candidate carries the exact source identity, typed schema fingerprint/barrier, generation, and deterministic dedupe boundary. |
| R7 | A fake proves only exit status. | A real PostgreSQL 16.10 dbtest logs the emitted rows plus identity/schema/checkpoint values. |

## RED command

```sh
go test -timeout 20m -count=1 ./internal/connectors/native/postgres -run 'TestPostgres.*Transport'
```

The exact RED output is retained in `traces/source-transport-red.txt` before
the production executor is added.

**Captured RED:** `go test -timeout 20m -count=1
./internal/connectors/native/postgres -run 'TestPostgres.*Transport'` failed
with `PostgreSQL definition has no declared source transport`.
