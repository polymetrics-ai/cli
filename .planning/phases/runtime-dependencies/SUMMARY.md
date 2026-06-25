# Phase Summary: runtime-dependencies

Implemented the first dependency-backed runtime layer while preserving the dependency-free local MVP.

## Added Runtime Infrastructure

- `deploy/compose/polymetrics-runtime.yml`
- `deploy/compose/.env.example`
- `deploy/temporal/dynamicconfig/development-sql.yaml`
- `deploy/temporal/scripts/setup-postgres.sh`
- `deploy/temporal/scripts/create-namespace.sh`
- `scripts/runtime.sh`
- `scripts/setup-runtime-macos.sh`
- `scripts/setup-runtime-linux.sh`

## Added Go Dependencies

- `github.com/jackc/pgx/v5 v5.7.6`
- `github.com/redis/go-redis/v9 v9.12.1`
- `go.temporal.io/sdk v1.37.0`

Pinned versions are compatible with the local Go 1.23.4 toolchain. Latest versions require newer Go releases.

## Implemented Application Layer

- `internal/runtimecheck`
- `internal/ledger`
- `internal/coordination`
- `internal/perf`
- `poly runtime doctor`
- `poly perf compare`
- `poly perf compare --runtime`
- `poly etl run --runtime`

## Performance Result

Dependency-free sample:

- 50 ETL loops
- 150 records
- 38.28725ms total
- 765.745us average per loop
- 3917.75 records/sec

Runtime-backed sample:

- Not measured successfully because PostgreSQL and Temporal were unavailable.
- Runtime stack startup failed during image pulls due DNS resolution failure for Docker CloudFront.

