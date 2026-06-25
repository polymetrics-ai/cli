---
name: pm-destination-mysql
description: MySQL connector knowledge and safe action guide.
---

# pm-destination-mysql

## Purpose

MySQL catalog connector for https://docs.airbyte.com/integrations/destinations/mysql. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-mysql:1.1.1 (metadata only; not executed)

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

- MySQL documentation: https://dev.mysql.com/doc/
- SQL statement syntax: https://dev.mysql.com/doc/refman/8.0/en/sql-statements.html
- Access control and account management: https://dev.mysql.com/doc/refman/8.0/en/access-control.html
- GRANT statement: https://dev.mysql.com/doc/refman/8.0/en/grant.html
- MySQL Release Notes: https://dev.mysql.com/doc/relnotes/mysql/en/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/mysql

## Configuration

- database (string) required: Name of the database.
- disable_type_dedupe (boolean): Disable Writing Final Tables. WARNING! The data format in _airbyte_data is likely stable but there are no guarantees that other metadata columns will remain the same in future v...
- host (string) required: Hostname of the database.
- jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
- password (string) secret: Password associated with the username.
- port (integer) required: Port of the database.
- raw_data_schema (string): The database to write raw tables into
- ssl (boolean): Encrypt data using SSL.
- tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
- username (string) required: Username to use to access the database.
- secret fields: password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

## Sync Modes

- supported sync modes: append, append_dedup, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/mysql

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-mysql
```

### Inspect as JSON

```bash
pm connectors inspect destination-mysql --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [MySQL documentation](https://docs.airbyte.com/integrations/destinations/mysql)
