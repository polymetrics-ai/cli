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

Recorded 2026-08-06 before adding the module; the approval is limited to this
native PostgreSQL CDC implementation.

- **Maintenance:** Go module proxy resolves `@latest` to
  `v0.0.0-20260401131349-e37c41485510`, published 2026-04-01. It is an
  untagged pseudo-version, so the exact version is pinned; the project remains
  maintained as part of Jack Christensen's active pgx family. Its package page
  identifies it as the PostgreSQL logical-replication client and links the
  maintained upstream repository.
- **Licence:** MIT (`LICENSE`, Copyright 2019 Jack Christensen; verified from
  the downloaded module and the Go package registry).
- **Known vulnerabilities:** OSV query on 2026-08-06 returned no advisories
  for `github.com/jackc/pglogrepl` or its only new production dependency,
  `github.com/jackc/pgio`. The existing `github.com/jackc/pgx/v5` has
  GO-2026-4771 / CVE-2026-33815 before v5.9.0; this repository already pins
  v5.10.0, so the introduced use is not in the affected range.
- **Production transitive footprint:** `pglogrepl` imports `pgio` v1.0.0 and
  `pgx/v5` (`pgconn`, `pgproto3`). `pgx/v5` v5.10.0 and `testify` are already
  direct project dependencies; `pgio` v1.0.0 is the only new production module
  and has no module dependencies. `testify` is present only in pglogrepl's
  own test graph and is already direct here.
- **Binary-size baseline:** `go build -trimpath -o <temporary>/pm ./cmd/pm`
  before the addition was **93,577,602 bytes**. The post-addition build is
  **93,740,082 bytes**, a measured increase of **162,480 bytes**.

Sources: https://pkg.go.dev/github.com/jackc/pglogrepl ;
https://pkg.go.dev/vuln/GO-2026-4771 ;
https://www.postgresql.org/docs/current/logicaldecoding-explanation.html

## What changed

- Adds a native `pglogrepl` executor that derives a deterministic slot from the
  PostgreSQL system identity, database, and canonical schema-qualified stream.
- Reuses the durable sync checkpoint envelope for LSN recovery, validates source
  identity and generation before resume, and acknowledges PostgreSQL only after
  a durable downstream checkpoint commit.
- Decodes `pgoutput` relation/insert/update/delete messages, preserves
  transaction ordering, and makes slot cleanup an explicit, safe lifecycle
  operation that refuses unknown, active, or incompatible slots.
- Declares PostgreSQL CDC as implemented only through the exact native executor
  and updates the connector contract and documentation.

## Verification

- `go test -count=1 ./internal/connectors/native/postgres ./internal/connectors/engine ./internal/cli`
- Live protocol conformance: `TestLogicalReplicationResumesAndCleansSlot` with
  an isolated **PostgreSQL 12.22** server configured with `wal_level=logical`,
  `max_replication_slots=4`, and `max_wal_senders=4` on a non-default port. It
  proves insert/update/delete decoding, durable-LSN resume to the next
  transaction, and replication-slot absence after teardown. The test fails on
  an unreachable configured source and explicitly skips only when the
  integration environment is intentionally absent.

This body intentionally contains no connection string, password, or other
credential value.
