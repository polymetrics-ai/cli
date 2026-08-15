# Verification — Issue #3982 PostgreSQL managed-table write driver

## GSD/TDD evidence

- Manual inline lifecycle is required by the canonical no-delegation worker
  contract: `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review` prompts were resolved with the project
  adapter. The issue is not a numbered roadmap phase, so this phase directory
  is the durable evidence location.
- Red: `traces/mapping-provisioning-red.txt` proves the shared provisioning
  plan initially could not carry `MappingContractV1`.
- Red: `traces/postgres-first-create-red.txt` proves the live PostgreSQL first
  create was unavailable before mapped DDL.
- Red: `traces/postgres-write-session-red.txt` proves the PostgreSQL driver did
  not initially satisfy `DatabaseWriteDriver`.
- Green: the dbtest test now reads created OIDs/control rows/typed target rows,
  five-mode data outcomes, explicit deletion, rollback retention, and unknown
  commit outcome from a real PostgreSQL connection.

## Final commands and observed result

| Command | Result |
| --- | --- |
| `go test -timeout 20m -count=1 ./internal/connectors/native/postgres/... ./internal/connectors/database/... ./internal/connectors/engine` | pass |
| `go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/connectors/database ./internal/connectors/engine` | pass |
| `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres` | pass; all tagged PostgreSQL tests completed in 13.316s |
| `go vet ./internal/connectors/native/postgres/... ./internal/connectors/database/... ./internal/connectors/engine` | pass |
| `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build` | pass |
| `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check` | pass |

## Acceptance evidence

| Criterion | Observable proof |
| --- | --- |
| First create/reassert | Live test reads one private owner/control row, PostgreSQL namespace/relation OIDs, and mapped `text`/`bigint` columns; repeat assertion returns the same relation OID. |
| Ownership/identity refusal | Live ownerless, foreign, tampered, permission, OID replacement, and schema drift cases re-read target state and assert no mutation. |
| Five modes and composite keys | Live rows are read after full append, incremental append, dedupe, composite-key upsert, and full overwrite. |
| Explicit delete only | Live ordinary upsert leaves the absent `tenant-b/id=1` row; a typed tombstone then removes exactly that row. |
| Atomicity/durability/unknown | Live PostgreSQL re-reads the durable receipt after successful sessions; statement failure, cancellation, and a real connection close after apply retain the prior receipt. The close reports shared unknown outcome with exactly one begin. |
| Capability fence | Unit tests retain `write=false` and `Connector.Write` unsupported while the private definition admits only the five driver modes. |
