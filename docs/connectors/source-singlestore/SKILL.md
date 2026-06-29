---
name: pm-source-singlestore
description: SingleStore connector knowledge and safe action guide.
---

# pm-source-singlestore

## Purpose

SingleStore catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/singlestore.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.singlestore.com/

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

- SingleStore documentation: https://docs.singlestore.com/

## Configuration

- database (string) required: Name of the database.
- host (string) required: Hostname of the database.
- jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
- password (string) secret: Password associated with the username.
- port (integer) required: Port of the database.
- replication_method (string) required: Replication method to use for extracting data from the database. STANDARD replication requires no setup on the DB side but will not be able to represent deletions incrementally.
- ssl_mode (object): SSL connection modes.
- username (string) required: Username to use to access the database.
- secret fields: password, ssl_mode.ca_certificate, ssl_mode.client_certificate, ssl_mode.client_key, ssl_mode.client_key_password

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
pm connectors inspect source-singlestore
```

### Inspect as JSON

```bash
pm connectors inspect source-singlestore --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [SingleStore documentation](https://docs.singlestore.com/)
