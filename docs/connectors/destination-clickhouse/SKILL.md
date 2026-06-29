---
name: pm-destination-clickhouse
description: ClickHouse connector knowledge and safe action guide.
---

# pm-destination-clickhouse

## Purpose

ClickHouse catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/clickhouse.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://clickhouse.com/docs

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
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

- family: destination_writer
- priority_wave: 1
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- ClickHouse documentation: https://clickhouse.com/docs
- SQL reference: https://clickhouse.com/docs/en/sql-reference
- User authentication: https://clickhouse.com/docs/en/operations/access-rights
- Changelog: https://clickhouse.com/docs/whats-new/changelog

## Configuration

- database (string) required: Name of the database.
- enable_json (boolean): Use the JSON type for Object fields. If disabled, the JSON will be converted to a string.
- host (string) required: Hostname of the database.
- password (string) required secret: Password associated with the username.
- port (string) required: HTTP port of the database. Default(s) HTTP: 8123 — HTTPS: 8443
- protocol (string) required: Protocol for the database connection string.
- record_window_size (integer): Warning: Tuning this parameter can impact the performances. The maximum number of records that should be written to a batch. The batch size limit is still limited to 70 Mb
- tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
- username (string) required: Username to use to access the database.
- secret fields: password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

## Sync Modes

- supported sync modes: append, append_dedup, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-clickhouse
```

### Inspect as JSON

```bash
pm connectors inspect destination-clickhouse --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [ClickHouse documentation](https://clickhouse.com/docs)
- [SQL reference](https://clickhouse.com/docs/en/sql-reference)
- [User authentication](https://clickhouse.com/docs/en/operations/access-rights)
- [Changelog](https://clickhouse.com/docs/whats-new/changelog)
