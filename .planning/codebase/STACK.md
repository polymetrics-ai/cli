# Technology Stack

**Analysis Date:** 2026-07-08
**Generated via:** Upstream `/gsd:map-codebase` workflow shape, issue #122 prompt.

## Languages

**Primary:**
- Go 1.25.4 module with pinned toolchain `go1.25.11` — CLI, connector runtime, conformance, certification, scheduling, vault, runtime probes.

**Secondary:**
- Shell — local runtime scripts and verification orchestration.
- TypeScript/JavaScript — website and release/documentation tooling under `website/` and build helpers.
- JSON / JSON Schema — connector definition bundles under `internal/connectors/defs/`.
- Markdown — CLI docs, architecture, migration, and active GSD planning artifacts.

## Runtime

**Environment:**
- Single Go binary `pm` built from `cmd/pm`.
- Optional runtime-backed execution uses local services via `scripts/runtime.sh` and Temporal dependencies.
- Local-first project state and warehouses live under `.polymetrics/` at runtime.

**Package Manager:**
- Go modules (`go.mod`, `go.sum`).
- Website uses Node tooling in `website/`.

## Frameworks and Core Libraries

**CLI and application:**
- Standard-library CLI architecture in `internal/cli` and `cmd/pm`.
- `internal/app` owns ETL, reverse ETL, warehouse/query, sync modes, and execution flows.

**Connector runtime:**
- `internal/connectors/engine` interprets JSON bundles.
- `internal/connectors/connsdk` provides low-level HTTP/auth/pagination/extract helpers.
- `internal/connectors/hooks` provides Tier 2 custom behavior.
- `internal/connectors/native` provides Tier 3 non-HTTP/custom connectors.

**Data/runtime dependencies:**
- `github.com/jackc/pgx/v5` for PostgreSQL.
- `github.com/marcboeker/go-duckdb` for optional DuckDB support.
- `github.com/redis/go-redis/v9` and Temporal SDK/API packages for runtime features.
- `github.com/stretchr/testify` for tests.

## Verification Tooling

- `make verify` runs formatting, tidy check, vet, tests, build, docs validation, smoke, lint, and `connectorgen validate`.
- `go test ./...`, `go vet ./...`, `go build ./cmd/pm` are primary local gates.
- `golangci-lint` covers declarative connector architecture packages.
- GitHub Actions workflows: verify, security, scorecard, release, website, PR issue guard.

## Connector Surface Technologies

Planning must not assume connectors are only REST/JSON APIs. Current repo evidence includes connector docs or defs mentioning:

- REST/HTTP JSON APIs — dominant connector shape.
- GraphQL — examples include `github`, `linear`, `monday`, `notion`, `plaid`, `stigg`.
- XML/SOAP/XML feeds — examples include `amazon-sqs`, `rss`, `tally-prime`, `workday`.
- CSV/NDJSON/report exports — examples include `amplitude`, `appsflyer`, `mixpanel`, `vercel`.
- Binary/download/upload/multipart capabilities — common across attachments, artifacts, archives, exports, documents, and media.
- Databases/CDC — `postgres`, `dynamodb`, and native connector directories.
- Queues/events/webhooks/audit logs — `amazon-sqs` plus many event/webhook resources.
- File/object storage style integrations — S3-like and signed-download flows appear in docs/defs.

---
*Stack analysis: 2026-07-08*
