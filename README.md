# Polymetrics AI pm CLI MVP

This repository contains a working Go-only `pm` CLI monolith prototype for Polymetrics AI.

It implements a dependency-free local vertical slice:

- project initialization
- embedded help/man-style docs
- encrypted local credential vault
- built-in `sample`, `github`, `file`, `warehouse`, and `outbox` connectors
- connection and catalog management
- ETL from a source connector into a local JSONL warehouse
- query over local warehouse tables
- reverse ETL plan, preview, approval, and execution into a local outbox
- agent-oriented JSON output

## Build

```bash
make verify
```

This runs `gofmt`, `go vet`, `go test ./...`, builds `./pm`, and executes an end-to-end smoke flow.

## Install

Install `pm` onto your PATH so it can be run from any directory:

```bash
make install
pm help
```

By default this installs to `~/.local/bin/pm`. Override the destination when needed:

```bash
make install INSTALL_DIR=/opt/homebrew/bin
```

Remove the installed binary with:

```bash
make uninstall
```

## Quick Start

```bash
go build ./cmd/pm

ROOT=$(mktemp -d)
export PM_SAMPLE_TOKEN=sample-token

./pm init --root "$ROOT"
./pm credentials add sample-local --connector sample --from-env token=PM_SAMPLE_TOKEN --root "$ROOT"
./pm credentials add warehouse-local --connector warehouse --config path="$ROOT/.polymetrics/warehouse" --root "$ROOT"
./pm credentials add outbox-local --connector outbox --config path="$ROOT/.polymetrics/outbox" --root "$ROOT"

./pm connections create sample_to_warehouse \
  --source sample:sample-local \
  --destination warehouse:warehouse-local \
  --stream customers \
  --primary-key id \
  --cursor updated_at \
  --table sample_customers \
  --root "$ROOT"

./pm catalog refresh --connection sample_to_warehouse --root "$ROOT"
./pm etl run --connection sample_to_warehouse --stream customers --root "$ROOT" --json
./pm query run --table sample_customers --limit 3 --root "$ROOT" --json
```

GitHub ETL can read public repositories without a token. Use `GITHUB_TOKEN` for
private repositories or a higher GitHub API rate limit:

```bash
go build ./cmd/pm

ROOT=$(mktemp -d)
./pm init --root "$ROOT"

# Public repository, no secret required.
./pm credentials add github-public \
  --connector github \
  --config repository=octocat/Hello-World \
  --root "$ROOT"

# Authenticated token alternatives. The token can be a classic PAT,
# fine-grained PAT, OAuth token, GitHub Actions GITHUB_TOKEN, or a
# pre-generated installation token.
# export GITHUB_TOKEN=...
# ./pm credentials add github-private \
#   --connector github \
#   --config repository=OWNER/REPO \
#   --from-env token=GITHUB_TOKEN \
#   --root "$ROOT"

# GitHub App installation alternative.
# ./pm credentials add github-app \
#   --connector github \
#   --config repository=OWNER/REPO \
#   --config auth_type=github_app \
#   --config app_id=12345 \
#   --config installation_id=67890 \
#   --value-stdin private_key \
#   --root "$ROOT" < app-private-key.pem

./pm credentials add warehouse-local \
  --connector warehouse \
  --config path="$ROOT/.polymetrics/warehouse" \
  --root "$ROOT"

./pm connections create github_to_warehouse \
  --source github:github-public \
  --destination warehouse:warehouse-local \
  --stream issues \
  --primary-key id \
  --cursor updated_at \
  --table github_issues \
  --root "$ROOT"

./pm catalog refresh --connection github_to_warehouse --root "$ROOT" --json
./pm etl run --connection github_to_warehouse --stream issues --root "$ROOT" --json
./pm query run --table github_issues --limit 5 --root "$ROOT" --json
```

The GitHub connector defaults to one page for safe local runs. To exhaust a
stream, set `--config max_pages=0` on the credential or `--source-config
max_pages=0` on a connection. `max_pages=all` and `max_pages=unlimited` are
accepted aliases.

For large streams, use bounded ETL batches:

```bash
./pm etl run --connection github_to_warehouse --stream pull_requests --batch-size 100 --root "$ROOT" --json
```

Reverse ETL is preview/approval based:

```bash
./pm reverse plan customers_to_outbox \
  --source-table sample_customers \
  --destination outbox:outbox-local \
  --map id:external_id \
  --map name:full_name \
  --map email:email \
  --root "$ROOT"

./pm reverse run <plan-id> --approve <approval-token> --root "$ROOT" --json
```

GitHub can also be a reverse ETL destination for approved issue and pull request
mutations. Writes require a token and an explicit action:

```bash
export GITHUB_TOKEN=...

./pm credentials add github-write \
  --connector github \
  --config repository=OWNER/REPO \
  --from-env token=GITHUB_TOKEN \
  --root "$ROOT"

./pm reverse plan prs_to_github \
  --source-table github_pr_candidates \
  --destination github:github-write \
  --action create_pull_request \
  --map title:title \
  --map body:body \
  --map head:head \
  --map base:base \
  --map reviewers:reviewers \
  --root "$ROOT"

./pm reverse preview <plan-id> --root "$ROOT" --json
./pm reverse run <plan-id> --approve <approval-token> --root "$ROOT" --json
```

Supported GitHub reverse ETL actions include `create_issue`, `update_issue`,
`comment_issue`, `close_issue`, `create_pull_request`, `update_pull_request`,
`close_pull_request`, `request_reviewers`, and `merge_pull_request`.

See `docs/connectors/github-etl-reverse-etl.md` for the complete GitHub auth,
ETL stream, sync mode, and reverse ETL action reference.

## Docs

```bash
./pm help
./pm help credentials
./pm man reverse
./pm docs generate --dir docs/cli
./pm skills generate --dir docs/skills --json
```

## Runtime Dependencies

The next integration phase uses local PostgreSQL, DragonflyDB, and Temporal services.

Runtime preference is Podman first, Docker fallback:

1. `podman compose`
2. `podman-compose`
3. `docker compose`
4. `docker-compose`

macOS:

```bash
scripts/setup-runtime-macos.sh
scripts/runtime.sh up
```

Linux:

```bash
scripts/setup-runtime-linux.sh
scripts/runtime.sh up
```

Useful commands:

```bash
make runtime-doctor
make runtime-up
make runtime-down
make runtime-reset
```

Service endpoints:

```text
PostgreSQL: localhost:15433
DragonflyDB: localhost:6379
Temporal gRPC: localhost:7233
Temporal UI: http://localhost:8080
```

Runtime details are in [docs/runtime/SETUP.md](docs/runtime/SETUP.md).

PostgreSQL and DragonflyDB now default to GHCR images:

```text
ghcr.io/enterprisedb/postgresql:16
ghcr.io/dragonflydb/dragonfly:latest
```

Temporal image paths are configurable with `TEMPORAL_SERVER_IMAGE`, `TEMPORAL_ADMINTOOLS_IMAGE`, and `TEMPORAL_UI_IMAGE`. The official Temporal images were not available under verified GHCR names for the configured tags, so the default Temporal images remain upstream `temporalio/*` unless a trusted GHCR mirror is configured.

## Performance Comparison

The CLI can compare the dependency-free path with the dependency-backed path.

Dependency-free mode:

```bash
make perf-free
```

Runtime-backed mode requires PostgreSQL, DragonflyDB, and Temporal to be healthy:

```bash
make runtime-up
make perf-runtime
make runtime-down
```

Dependency-free means the CLI uses only local files:

- JSON state
- AES-GCM encrypted vault files
- JSONL warehouse and outbox
- in-process ETL and reverse ETL execution

Runtime-backed means the same local ETL loop runs while also using:

- PostgreSQL for run-ledger persistence
- DragonflyDB for lease coordination
- Temporal health checks as the durable workflow target

## Current Limitations

This is a working MVP, not the final production architecture. The PRD calls for SQLite, DuckDB, OS keychain integration, Parquet batches, and real SaaS/database connectors. Those are intentionally deferred because they require dependency and integration decisions.

The runtime Compose stack is local development infrastructure. The Go client dependencies for PostgreSQL, DragonflyDB, and Temporal are now present, but full workflow migration to Temporal is still incremental.
