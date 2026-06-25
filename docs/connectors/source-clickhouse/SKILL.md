---
name: pm-source-clickhouse
description: ClickHouse connector knowledge and safe action guide.
---

# pm-source-clickhouse

## Purpose

ClickHouse catalog connector for https://docs.airbyte.com/integrations/sources/clickhouse. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-clickhouse:0.3.0 (metadata only; not executed)

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

- ClickHouse HTTP interface: https://clickhouse.com/docs/en/interfaces/http
- ClickHouse SQL reference: https://clickhouse.com/docs/en/sql-reference
- ClickHouse authentication: https://clickhouse.com/docs/en/operations/access-rights
- Changelog: https://clickhouse.com/docs/whats-new/changelog
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/clickhouse

## Configuration

- database (string) required: The name of the database.
- host (string) required: The host endpoint of the Clickhouse cluster.
- jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (Eg. key1=value1&key2=value2&key...
- password (string) secret: The password associated with this username.
- port (integer) required: The port of the database.
- ssl (boolean): Encrypt data using SSL.
- tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
- username (string) required: The username which is used to access the database.
- secret fields: password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/clickhouse

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-clickhouse
```

### Inspect as JSON

```bash
pm connectors inspect source-clickhouse --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [ClickHouse documentation](https://docs.airbyte.com/integrations/sources/clickhouse)
