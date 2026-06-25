# Verification

Phase: native-connector-port-program

## Automated Checks

- `go test ./internal/connectors -run TestNativePort` passed.
- `go test ./internal/cli -run TestConnectorPortPlan` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `go build -o pm ./cmd/pm` passed.
- `./pm docs generate --dir docs/cli --connectors-dir docs/connectors` passed.
- `./pm docs validate --connectors-dir docs/connectors` passed.
- `make verify` passed, including the sample ETL and reverse ETL smoke flow.
- `make install` installed `pm` to `/Users/karthiksivadas/.local/bin/pm`.

## CLI Spot Checks

- `pm connectors port-plan --all --json` returned `total=647`.
- Native port family summary:
  - `native_saas=1`
  - `declarative_http_source=503`
  - `database_cdc_source=7`
  - `database_source=18`
  - `file_object_source=11`
  - `destination_writer=56`
  - `custom_go_port=51`
- Wave summary:
  - `wave_0=1`
  - `wave_1=60`
  - `wave_2=29`
  - `wave_3=557`
- `pm connectors port-plan source-postgres` renders `postgres_logical_replication`, `wal_level=logical`, and `lsn` state.
- `pm connectors port-plan source-mysql --json` renders `mysql_binlog`, `binlog_format=ROW`, and `gtid_or_binlog_position`.
- `pm connectors port-plan source-mongodb-v2 --json` renders `mongodb_change_streams`, `replica set or sharded cluster`, and `resume_token`.

## GSD Helper Warning

`programming-loop.mjs verify --phase native-connector-port-program --execute` still reports missing inferred commands and `git diff --check` failure because this workspace is not a git worktree and the helper does not detect the Makefile checks. The explicit local verification above is authoritative.
