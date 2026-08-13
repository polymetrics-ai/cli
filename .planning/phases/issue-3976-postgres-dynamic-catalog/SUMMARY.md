---
phase: issue-3976-postgres-dynamic-catalog
issue: 3976
status: partial
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
        status: unknown
    human_judgment: true
    rationale: The live dbtest harness is implemented and invokes cleanly, but this workspace has no explicit direct local Podman endpoint or opt-in; a skipped test cannot establish the live claim.
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
- Parent synchronization: `25bda3e73` safely merges #4064's parent head
  `c2e013324`; no history was forced or discarded.
- Draft stacked child PR: #4065, base `feat/3972-postgres-parity`.

## Verification result

Automated local coverage is green. The optional live PostgreSQL proof is
deliberately `unknown` rather than passed: the tagged test skipped before
container startup because this environment provides neither
`POLYMETRICS_DATABASE_INTEGRATION=1` nor an explicit
`POLYMETRICS_PODMAN_ENDPOINT`. `UAT.md` records that limited result.

## Boundary audit

The #3976 diff changes only source catalog discovery, its tests, and generated
connector-description artifacts. Static fixture streams remain test-only.
Destination DDL/write/evolution (#3982), Parquet workset typing (#3980),
outbound delivery (#3983), and sync-mode execution (#3987) were audited but
left untouched.
