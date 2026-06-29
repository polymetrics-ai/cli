---
name: pm-source-fauna
description: Fauna connector knowledge and safe action guide.
---

# pm-source-fauna

## Purpose

Fauna catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/fauna.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.fauna.com/fauna/current/api/

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.

## Runtime Capabilities

- metadata=true
- check=false
- catalog=false
- read=false
- write=false
- query=false
- etl=false
- reverse_etl=false
- unsupported_reason: Native Go port is planned but not enabled; only catalog metadata is available.

## Native Port Plan

- family: database_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

## Official Application Documentation

- Fauna API reference: https://docs.fauna.com/fauna/current/api/
- Fauna authentication: https://docs.fauna.com/fauna/current/security/
- Fauna Status: https://status.fauna.com/

## Configuration

- collection (object): Settings for the Fauna Collection.
- domain (string) required: Domain of Fauna to query. Defaults db.fauna.com. See <a href=https://docs.fauna.com/fauna/current/learn/understanding/region_groups#how-to-use-region-groups>the docs</a>.
- port (integer) required: Endpoint port.
- scheme (string) required: URL scheme.
- secret (string) required secret: Fauna secret, used when authenticating with the database.
- secret fields: secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-fauna
```

### Inspect as JSON

```bash
pm connectors inspect source-fauna --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Fauna API reference](https://docs.fauna.com/fauna/current/api/)
- [Fauna authentication](https://docs.fauna.com/fauna/current/security/)
- [Fauna Status](https://status.fauna.com/)
