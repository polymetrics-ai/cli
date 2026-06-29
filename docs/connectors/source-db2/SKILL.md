---
name: pm-source-db2
description: IBM Db2 connector knowledge and safe action guide.
---

# pm-source-db2

## Purpose

IBM Db2 catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/db2.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.ibm.com/docs/en/db2/11.5?topic=reference-sql

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

- IBM Db2 SQL reference: https://www.ibm.com/docs/en/db2/11.5?topic=reference-sql
- IBM Db2 authentication: https://www.ibm.com/docs/en/db2/11.5?topic=security-authentication

## Configuration

- db (string) required: Name of the database.
- encryption (object) required: Encryption method to use when communicating with the database
- host (string) required: Host of the Db2.
- jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
- password (string) required secret: Password associated with the username.
- port (integer) required: Port of the database.
- username (string) required: Username to use to access the database.
- secret fields: encryption.key_store_password, encryption.ssl_certificate, password

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
pm connectors inspect source-db2
```

### Inspect as JSON

```bash
pm connectors inspect source-db2 --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [IBM Db2 SQL reference](https://www.ibm.com/docs/en/db2/11.5?topic=reference-sql)
- [IBM Db2 authentication](https://www.ibm.com/docs/en/db2/11.5?topic=security-authentication)
