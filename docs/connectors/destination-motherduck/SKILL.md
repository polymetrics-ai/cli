---
name: pm-destination-motherduck
description: MotherDuck connector knowledge and safe action guide.
---

# pm-destination-motherduck

## Purpose

MotherDuck catalog connector for https://docs.airbyte.com/integrations/destinations/motherduck. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-motherduck:0.2.3 (metadata only; not executed)

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
- priority_wave: 3
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- MotherDuck documentation: https://motherduck.com/docs
- Authentication: https://motherduck.com/docs/key-tasks/authenticating-and-connecting-to-motherduck/
- MotherDuck Version Lifecycle Schedules: https://motherduck.com/docs/troubleshooting/version-lifecycle-schedules/
- MotherDuck Status: https://status.motherduck.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/motherduck

## Configuration

- destination_path (string): Path to a .duckdb file or 'md:<DATABASE_NAME>' to connect to a MotherDuck database. If 'md:' is specified without a database name, the default MotherDuck database name ('my_db')...
- motherduck_api_key (string) required secret: API access token to use for authentication to a MotherDuck database.
- schema (string): Database schema name, defaults to 'main' if not specified.
- secret fields: motherduck_api_key

## Sync Modes

- supported sync modes: append, append_dedup, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/motherduck

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-motherduck
```

### Inspect as JSON

```bash
pm connectors inspect destination-motherduck --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [MotherDuck documentation](https://docs.airbyte.com/integrations/destinations/motherduck)
