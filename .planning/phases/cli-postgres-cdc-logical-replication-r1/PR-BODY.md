<!-- Issue-first PR body draft. Preserve dependency evidence when opening the stacked PR. -->

## Intent

Implement real PostgreSQL logical-replication CDC on top of the durable database
sync contract and the polling-watermark executor shape now merged to `main`.
This PR references the fail-closed CDC discovery work in #2986.

## Linked work

Refs #2986

Built on merged PRs #3880 (polling-watermark executor) and #3882 (durable
database sync contract).

## Conditional dependency evidence — `github.com/jackc/pglogrepl`

Recorded 2026-08-10 before updating the PR; the approval is limited to this
native PostgreSQL CDC implementation.

- **Maintenance / last release:** `go list -m` resolves `@latest` to
  `v0.0.0-20260401131349-e37c41485510`, released 2026-04-01, and the upstream
  `master` ref resolves to the same `e37c41485510` revision. It is an untagged
  pseudo-version rather than a stale release line; the exact current revision
  is pinned and the project remains maintained with the pgx family. Its package
  page identifies it as the PostgreSQL logical-replication client and links the
  maintained upstream repository.
- **Licence:** MIT (`LICENSE`, Copyright 2019 Jack Christensen; verified from
  the downloaded module and the Go package registry).
- **Known vulnerabilities:** Exact OSV queries on 2026-08-10 returned **0**
  advisories for `github.com/jackc/pglogrepl`
  `v0.0.0-20260401131349-e37c41485510`, new production dependency
  `github.com/jackc/pgio` `v1.0.0`, and the retained
  `github.com/jackc/pgx/v5` `v5.10.0`. There are no applicable known CVEs at
  the selected versions.
- **Transitive footprint:** Runtime imports are `pglogrepl`, new indirect
  `pgio` v1.0.0, and existing `pgx/v5` v5.10.0 (`pgconn`, `pgproto3`). The
  module-graph delta is exactly `pglogrepl`, `pgio`, and an existing test-only
  transitive upgrade of `github.com/rogpeppe/go-internal` from v1.11.0 to
  v1.12.0; no other production library was added. `testify` is already direct
  in this repository and remains only in pglogrepl's test graph.
- **Measured binary-size delta:** on the same Darwin/arm64 host with Go
  `go1.26.5`, `go build -trimpath -o <temporary>/pm ./cmd/pm` measured
  **145,079,490 bytes** at `origin/main` and **145,239,730 bytes** on this
  branch: **+160,240 bytes**.

Sources: https://pkg.go.dev/github.com/jackc/pglogrepl ;
https://osv.dev/ ;
https://www.postgresql.org/docs/current/logicaldecoding-explanation.html

## What changed

- Adds a native `pglogrepl` executor that derives a deterministic slot from the
  PostgreSQL system identity, database, and canonical schema-qualified stream.
- Reuses the durable sync checkpoint envelope for LSN recovery, validates source
  identity and generation before resume, and acknowledges PostgreSQL only after
  a durable downstream checkpoint commit.
- Validates that the publication includes the requested relation before slot
  creation; decodes and filters `pgoutput` relation/insert/update/delete and
  truncate messages while preserving transaction ordering.
- Makes slot cleanup an explicit, safe lifecycle operation that refuses unknown,
  active, or incompatible slots, and refuses an existing slot without the
  matching durable checkpoint rather than risking a WAL skip.
- Declares PostgreSQL CDC as implemented only through the exact native executor
  and updates the connector contract and documentation.

## Verification

- `go test -count=1 -timeout 20m ./internal/connectors ./internal/connectors/bundleregistry ./internal/connectors/engine`
- `go test -count=1 -timeout 20m ./internal/cli` (including #3964's
  unresolved connector-help exit-status coverage)
- `POLYMETRICS_INTEGRATION=1 go test -count=1 -timeout 20m ./internal/connectors/native/postgres`
- Live protocol conformance: `TestLogicalReplicationResumesAndCleansSlot` with
  an isolated **PostgreSQL 12.22** server configured with `wal_level=logical`,
  `max_replication_slots=8`, and `max_wal_senders=8` on a non-default port. It
  proves publication-membership refusal before slot creation, active-slot
  teardown refusal, selected-table insert/update/delete/truncate decoding,
  durable-LSN restart without duplicate prior records, rebootstrap refusal for
  an existing slot missing its checkpoint, and replication-slot absence after
  teardown. The test fails on an unreachable configured source and explicitly
  skips only when the integration environment is intentionally absent.

This body intentionally contains no connection string, password, or other
credential value.
