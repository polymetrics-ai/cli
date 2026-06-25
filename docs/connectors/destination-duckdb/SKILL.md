---
name: pm-destination-duckdb
description: DuckDB connector knowledge and safe action guide.
---

# pm-destination-duckdb

## Purpose

DuckDB catalog connector for https://docs.airbyte.com/integrations/destinations/duckdb. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: beta
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-duckdb:0.6.0 (metadata only; not executed)

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

- family: destination_writer
- priority_wave: 2
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- DuckDB documentation: https://duckdb.org/docs/
- SQL reference: https://duckdb.org/docs/sql/introduction
- MotherDuck Version Lifecycle Schedules: https://motherduck.com/docs/troubleshooting/version-lifecycle-schedules/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/duckdb

## Configuration

- destination_path (string) required: Path to the .duckdb file, or the text 'md:' to connect to MotherDuck. The file will be placed inside that local mount. For more information check out our <a href="https://docs.a...
- motherduck_api_key (string) secret: API key to use for authentication to a MotherDuck database.
- schema (string): Database schema name, default for duckdb is 'main'.
- secret fields: motherduck_api_key

## Sync Modes

- supported sync modes: append, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/duckdb

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-duckdb
```

### Inspect as JSON

```bash
pm connectors inspect destination-duckdb --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [DuckDB documentation](https://docs.airbyte.com/integrations/destinations/duckdb)
