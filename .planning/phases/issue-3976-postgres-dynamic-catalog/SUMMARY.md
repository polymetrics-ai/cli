---
phase: issue-3976-postgres-dynamic-catalog
issue: 3976
status: passed-with-explicit-follow-up-boundary
coverage:
  - id: D1
    description: The shipping PostgreSQL catalog path derives ordered base relations, columns, nullability, and primary/unique key membership from pg_catalog through the #4034 typed model.
    verification:
      - kind: unit
        ref: internal/connectors/native/postgres/typed_catalog_test.go
        status: pass
      - kind: other
        ref: go test -timeout 20m -count=1 ./internal/connectors/native/postgres ./internal/connectors/database ./internal/connectors/engine
        status: pass
    human_judgment: false
  - id: D2
    description: Supported PostgreSQL native details and logical types remain structured, while unsupported and unsafe shapes reject explicitly.
    verification:
      - kind: unit
        ref: TestPostgresColumnTypeRetainsSupportedNativeDetails and TestPostgresColumnTypeRejectsUnsupportedShapes
        status: pass
      - kind: other
        ref: go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres
        status: pass
    human_judgment: false
  - id: D3
    description: The legacy connectors.Catalog response is a one-way compatibility projection of the typed runtime catalog rather than a second static discovery model.
    verification:
      - kind: unit
        ref: TestLegacyStreamsFromTypedCatalogIsProjectionOnly
        status: pass
      - kind: other
        ref: make connector-boundary
        status: pass
    human_judgment: false
  - id: D4
    description: Two materially different live PostgreSQL schemas produce independently verified, different typed catalogs without code or connector schema changes.
    verification:
      - kind: integration
        ref: TestPostgresDynamicTypedCatalogUsesLiveMetadata with independent information_schema oracle
        status: pass
    human_judgment: false
---

# Summary — Issue #3976

Implemented dynamic typed PostgreSQL catalog discovery from `pg_catalog` for
the configured source database/schema. The native runtime builds #4034's
structured catalog/fingerprint before deriving its legacy compatibility stream
projection, preserving database/schema/relation identity, ordered columns,
nullability, keys, and supported type details.

## Delivery state

- RED commit: `db7e06d36`.
- GREEN implementation commit: `24d0055f5`.
- Integration synchronization: `0df3d5d4d` safely merges
  `integration/4015-mvp-flat-r1` at
  `fbd06e7d7c5c0632182e98cbb3a223ba25b19883`; no history was forced or
  discarded.
- Draft child PR: #4065, base `integration/4015-mvp-flat-r1`.

## Verification result

Automated local coverage and the opt-in live PostgreSQL source proof are
green. Through Docker's explicit Colima Unix socket, the dbtest harness
discovered seeded catalog metadata, returned all five full-read rows, and
returned only IDs 3–5 after cursor value 10. It also records current absent,
missing, nullable, and connection-level cursor-field behavior in
`traces/live-reads-green.txt`. The historical logical-replication test stays
intentionally skipped/fail-closed and is not claimed as live CDC coverage.

The evidence does not claim that the legacy scalar reader is #3858's
declaration-selected tuple/checkpoint executor. That execution boundary remains
an explicit follow-up rather than being hidden by a successful legacy read.

## Boundary audit

The #3976 diff changes only source catalog discovery, its tests, and generated
connector-description artifacts. Static fixture streams remain test-only.
Destination DDL/write/evolution (#3982), Parquet workset typing (#3980),
outbound delivery (#3983), and sync-mode execution (#3987) were audited but
left untouched.
