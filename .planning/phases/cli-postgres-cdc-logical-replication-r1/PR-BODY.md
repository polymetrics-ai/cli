<!-- Issue-first PR body draft. Preserve dependency evidence when opening the stacked PR. -->

## Intent

Preserve the approved native PostgreSQL logical-replication CDC design and
dependency evidence while change capture remains deliberately planned and
non-executable. `ReadCDC` fails closed before opening a source connection;
this PR does not claim current end-to-end CDC, source-LSN progress, or slot
lifecycle execution. This PR references the fail-closed CDC discovery work in
#2986.

## Linked work

Refs #2986

Built on merged PRs #3880 (polling-watermark executor) and #3882 (durable
database sync contract).

## Conditional dependency evidence — `github.com/jackc/pglogrepl`

Recorded 2026-08-06 before adding the module; the approval is limited to the
planned native PostgreSQL CDC path and is not an assertion that it is currently
executable.

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

## Current containment and retained implementation material

- Keeps `pglogrepl` limited to the native PostgreSQL CDC path, while the
  connector and bundle declare change capture planned rather than implemented.
- Retains the prior source-bound slot, checkpoint, and `pgoutput` implementation
  material for the future streamed-staging design, but `ReadCDC` rejects before
  a source connection, checkpoint advance, or replication-slot operation.
- Keeps connector documentation and capability projections aligned with the
  planned, non-executable state.

## Verification

- Current containment: focused PostgreSQL capability tests confirm that
  `ReadCDC` returns the fail-closed unsupported result and the connector does
  not advertise CDC. No live source is contacted.
- Historical, pre-containment protocol evidence:
  `TestHistoricalLogicalReplicationResumesAndCleansSlot` previously ran against
  an isolated **PostgreSQL 12.22** server configured with `wal_level=logical`,
  `max_replication_slots=4`, and `max_wal_senders=4` on a non-default port. It
  recorded insert/update/delete decoding, durable-LSN resume, and slot cleanup.
  The preserved test is now skipped while change capture is planned, so this is
  not current conformance or evidence of source progress.

This body intentionally contains no connection string, password, or other
credential value.
