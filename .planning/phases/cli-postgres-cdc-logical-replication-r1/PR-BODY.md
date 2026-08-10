<!-- Issue-first PR body draft. Preserve dependency evidence when opening the stacked PR. -->

## Intent

Preserve the approved PostgreSQL logical-replication design and its dependency
evidence while `change_capture` remains deliberately planned and
non-executable. The executor rejects before opening a source connection; this
PR does not claim current end-to-end CDC, source-LSN progress, or replication-
slot lifecycle execution. This PR references the fail-closed CDC discovery
work in #2986.

## Linked work

Refs #2986

Built on merged PRs #3880 (polling-watermark executor) and #3882 (durable
database sync contract).

## Conditional dependency evidence — `github.com/jackc/pglogrepl`

Recorded 2026-08-10 before updating the PR; the approval is limited to the
native PostgreSQL CDC path and is not a general dependency licence.

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

## Current containment and retained implementation material

- Keeps `pglogrepl` limited to the native PostgreSQL CDC implementation material;
  it is not used by another connector or as a general-purpose dependency.
- Keeps the bundle, metadata, and runtime capability fail-closed: PostgreSQL is
  absent from the CDC catalogue and `ReadCDC` returns a named unsupported error
  before a source connection, slot creation/reuse, source acknowledgement, or
  checkpoint advance.
- Forces and verifies UTF-8 at the replication boundary; decoder regressions
  prove a non-ASCII payload round-trips byte-exact and malformed text is rejected
  before durable handling.
- Admits only `REPLICA IDENTITY DEFAULT` with a primary key in the retained
  protocol path. `FULL`, `USING INDEX`, `NOTHING`, and unknown modes return named
  errors rather than silently misidentify updates or deletes.
- The next executable phase must use PostgreSQL 14+ `proto_version 2` with
  `streaming=on`, a bounded crash-recoverable transaction stage, `StreamAbort`
  discard, a whole-transaction durable receipt at `StreamCommit`, and source LSN
  acknowledgement only after that receipt. There is no cursor/timestamp fallback.
  Slot-health observability plus explicit connector-owned teardown/rebootstrap
  are mandatory for that phase.

## Verification

- Current containment: focused PostgreSQL and CLI tests prove that the capability
  is false, inspect reports a planned reason, and `ReadCDC` is unsupported before
  any source interaction. No live source is contacted.
- Historical, pre-containment protocol evidence: `TestHistoricalLogicalReplicationResumesAndCleansSlot` with
  an isolated **PostgreSQL 12.22** server configured with `wal_level=logical`,
  `max_replication_slots=8`, and `max_wal_senders=8` on a non-default port. It
  previously proved publication-membership refusal before slot creation, active-slot
  teardown refusal, selected-table insert/update/delete/truncate decoding,
  durable-LSN restart without duplicate prior records, rebootstrap refusal for
  an existing slot missing its checkpoint, and replication-slot absence after
  teardown. It is now skipped while CDC is planned, so it is not current
  end-to-end conformance evidence.

This body intentionally contains no connection string, password, or other
credential value.
