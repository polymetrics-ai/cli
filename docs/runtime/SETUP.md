# Runtime Setup

This project can run local PostgreSQL, DragonflyDB, and Temporal services for the next integration phase.

The scripts prefer Podman first. Docker is used only when Podman is unavailable.

## macOS

```bash
scripts/setup-runtime-macos.sh
scripts/runtime.sh up
```

## Linux

```bash
scripts/setup-runtime-linux.sh
scripts/runtime.sh up
```

## Runtime Commands

```bash
scripts/runtime.sh doctor
scripts/runtime.sh up
scripts/runtime.sh ps
scripts/runtime.sh logs
scripts/runtime.sh down
scripts/runtime.sh reset
```

## Services

```text
PostgreSQL: localhost:15433
DragonflyDB: localhost:6379
Temporal gRPC: localhost:7233
Temporal UI: http://localhost:8080
```

## Image Registries

Defaults use GHCR where verified images exist:

```text
PostgreSQL: ghcr.io/enterprisedb/postgresql:16
DragonflyDB: ghcr.io/dragonflydb/dragonfly:latest
```

Temporal's official `server`, `admin-tools`, and `ui` images were not available under verified `ghcr.io/temporalio/*` names for the configured tags, so the defaults remain:

```text
Temporal Server: temporalio/server:1.31.0
Temporal Admin Tools: temporalio/admin-tools:1.31.0
Temporal UI: temporalio/ui:2.49.1
```

If you publish or choose a trusted GHCR mirror, override these in `deploy/compose/.env`:

```text
TEMPORAL_SERVER_IMAGE=ghcr.io/<org>/<temporal-server>:<tag>
TEMPORAL_ADMINTOOLS_IMAGE=ghcr.io/<org>/<temporal-admin-tools>:<tag>
TEMPORAL_UI_IMAGE=ghcr.io/<org>/<temporal-ui>:<tag>
```

## Configuration

The first runtime command copies:

```text
deploy/compose/.env.example -> deploy/compose/.env
```

Edit `deploy/compose/.env` for local overrides. Do not commit real secrets.

## Podman Fallback Rule

Runtime selection order:

1. `podman compose`
2. `podman-compose`
3. `docker compose`
4. `docker-compose`

On macOS, Podman requires a running machine. The setup script creates and starts:

```text
polymetrics-runtime
```

## Production Notes

This Compose runtime is for development and local integration tests only.

Production should use managed or operator-managed services:

- PostgreSQL with backups and migration discipline.
- DragonflyDB with explicit persistence and memory policy.
- Temporal Server through Kubernetes/Helm or a managed Temporal option.
