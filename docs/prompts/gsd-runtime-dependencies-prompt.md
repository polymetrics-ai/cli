# Prompt: Runtime Dependencies Phase

Use this prompt to implement the PostgreSQL, DragonflyDB, and Temporal integration phase after the local MVP.

```text
Act as an expert Go engineer and architect.

Goal:
Implement the Polymetrics runtime-dependencies phase using the GSD universal programming loop.

Phase:
runtime-dependencies

Repository:
Polymetrics Go CLI monolith.

Context files:
- .planning/PROJECT.md
- .planning/ROADMAP.md
- .planning/STATE.md
- .planning/config.json
- docs/architecture/repo-profile.json
- POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md
- docs/architecture/runtime-dependencies.md
- deploy/compose/polymetrics-runtime.yml
- scripts/runtime.sh
- scripts/setup-runtime-macos.sh
- scripts/setup-runtime-linux.sh

Execution rules:
1. Use the GSD universal programming loop.
2. Use strict TDD for behavior changes.
3. Prefer Podman first. If Podman is unavailable, fall back to Docker.
4. Keep runtime service setup separate from application logic.
5. Do not store plaintext credentials in repository files.
6. Do not leak credentials to logs or JSON output.
7. Keep ETL/reverse ETL agent operations typed and approval-gated.
8. Add new Go module dependencies only after explicit approval.

Approved target runtime services:
- PostgreSQL for durable metadata, run ledgers, catalog snapshots, approvals, and optional SQL source/destination connector tests.
- DragonflyDB for Redis-compatible ephemeral queues, leases, rate limits, batch pointers, and short-lived coordination state.
- Temporal for durable ETL and reverse ETL workflow orchestration once the local in-process MVP is ready to graduate.

Proposed Go dependencies after approval:
- github.com/jackc/pgx/v5 for PostgreSQL.
- github.com/redis/go-redis/v9 for DragonflyDB's Redis-compatible API.
- go.temporal.io/sdk for Temporal workflows/workers.

Expected implementation sequence:
1. Verify local runtime scripts and Compose config.
2. Add configuration structs and validation for PostgreSQL, DragonflyDB, and Temporal endpoints.
3. Add health-check commands under the CLI, e.g. pm runtime doctor --json.
4. Add PostgreSQL store behind the existing app store boundary.
5. Add Dragonfly-backed coordination package behind a small interface.
6. Add Temporal workflow package behind the existing ETL/reverse ETL use case boundary.
7. Keep local JSON mode available as a development fallback.
8. Add integration tests gated by POLYMETRICS_INTEGRATION=1 so unit tests remain fast.
9. Update docs and phase artifacts.

Verification:
- make verify
- scripts/runtime.sh doctor
- scripts/runtime.sh up
- scripts/runtime.sh ps
- scripts/runtime.sh down
- POLYMETRICS_INTEGRATION=1 go test ./...

Deliverables:
- Code changes
- Tests
- CLI docs
- Runtime docs
- Updated planning artifacts
```

