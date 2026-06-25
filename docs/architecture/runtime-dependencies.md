# Runtime Dependencies Architecture

Status: Draft
Date: 2026-06-24

## Purpose

The current Go CLI MVP is dependency-free and file-backed. The next production hardening phase should introduce PostgreSQL, DragonflyDB, and Temporal without weakening the CLI and agent safety model.

This document defines how those dependencies should be introduced.

## Runtime Roles

### PostgreSQL

Use PostgreSQL for durable control-plane data:

- credentials metadata, not plaintext secrets
- connector definitions and catalog snapshots
- connections and stream sync configuration
- ETL runs, reverse ETL plans, approvals, and audit events
- cursor checkpoints and run ledgers
- integration-test source/destination tables

Go client:

- `github.com/jackc/pgx/v5 v5.7.6`
- Latest `pgx` requires a newer Go toolchain than the local Go 1.23.4 environment, so this repo currently pins a compatible version.

Boundary:

- Add `internal/store/postgres` behind the app store interface.
- Keep the existing JSON store as a local fallback until migrations are stable.
- Use parameterized SQL only.
- Use append-only audit tables for mutation records.

### DragonflyDB

Use DragonflyDB as the Redis-compatible coordination layer:

- short-lived leases
- rate-limit counters
- workflow hints and batch pointers
- retry coordination
- ephemeral agent action locks
- cache for catalog and schema hints where recomputation is cheap

Go client:

- `github.com/redis/go-redis/v9 v9.12.1`
- Latest `go-redis` requires a newer Go toolchain than the local Go 1.23.4 environment, so this repo currently pins a compatible version.

Boundary:

- Add `internal/coordination` with a minimal interface.
- Do not store durable run truth in DragonflyDB.
- Never store plaintext credentials, raw row payloads, or approval tokens in DragonflyDB.

### Temporal

Use Temporal when ETL and reverse ETL need durable orchestration beyond the local in-process runner:

- long-running extraction
- checkpointed transforms
- retries and backoff
- reverse ETL approval waits
- cancellation and resume
- per-connector worker isolation

Go dependency:

- `go.temporal.io/sdk v1.37.0`
- Latest Temporal SDK requires a newer Go toolchain than the local Go 1.23.4 environment, so this repo currently pins a compatible version.

Boundary:

- Add workflows under `internal/workflows`.
- Add side-effecting activities under `internal/activities`.
- Keep workflow inputs bounded and serializable.
- Do not put secrets, raw row payloads, or large batches in workflow history.
- Store large batch payloads in object storage, PostgreSQL references, or local spool files depending on deployment mode.

## Local Runtime

Local runtime services are defined in:

- `deploy/compose/polymetrics-runtime.yml`
- `deploy/compose/.env.example`
- `deploy/temporal/dynamicconfig/development-sql.yaml`
- `deploy/temporal/scripts/setup-postgres.sh`
- `deploy/temporal/scripts/create-namespace.sh`

Control scripts:

- `scripts/setup-runtime-macos.sh`
- `scripts/setup-runtime-linux.sh`
- `scripts/runtime.sh`

Runtime selection rule:

1. Prefer Podman if installed.
2. Use `podman compose` or `podman-compose`.
3. If Podman is unavailable, use Docker Compose.

Image registry rule:

- PostgreSQL defaults to `ghcr.io/enterprisedb/postgresql:16`.
- DragonflyDB defaults to `ghcr.io/dragonflydb/dragonfly:latest`.
- Temporal image names are configurable through `TEMPORAL_SERVER_IMAGE`, `TEMPORAL_ADMINTOOLS_IMAGE`, and `TEMPORAL_UI_IMAGE`.
- Official Temporal GHCR images were not available under verified `ghcr.io/temporalio/*` names for the configured tags, so the default Temporal images remain the upstream `temporalio/*` references until a trusted GHCR mirror is selected.

## Default Local Endpoints

```text
PostgreSQL: localhost:15433
DragonflyDB: localhost:6379
Temporal gRPC: localhost:7233
Temporal UI: http://localhost:8080
```

## Security Notes

- The checked-in `.env.example` is for local development only.
- Real deployments must set strong passwords through environment variables or a secret manager.
- Temporal workflow histories must never include plaintext credentials.
- Agent JSON output must include references and status only, not secrets.

## Integration Test Strategy

Default unit tests remain dependency-free:

```bash
go test ./...
```

Runtime integration tests should be explicitly gated:

```bash
scripts/runtime.sh up
POLYMETRICS_INTEGRATION=1 go test ./...
scripts/runtime.sh down
```

## Implemented MVP Integration

Implemented in this repository:

- `pm runtime doctor --json`
- `pm perf compare --json`
- `pm perf compare --runtime --json`
- PostgreSQL run-ledger append path
- DragonflyDB lease coordination path
- Temporal health check path

Still pending:

- Full PostgreSQL replacement for JSON state.
- Full Temporal workflow execution for ETL and reverse ETL.
- Dragonfly-backed rate limits and batch pointer coordination in the production runner.

## Production Direction

The local Compose runtime is for development and CI-like local checks. Production should run:

- managed PostgreSQL or operator-managed PostgreSQL
- DragonflyDB with explicit persistence/replication decisions
- Temporal Server through Kubernetes/Helm or a managed Temporal option
