---
name: pm-source-postgres
description: Postgres connector knowledge and safe action guide.
---

# pm-source-postgres

## Purpose

Postgres catalog connector for https://docs.airbyte.com/integrations/sources/postgres. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-postgres:3.8.1 (metadata only; not executed)

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

- family: database_cdc_source
- priority_wave: 1
- etl_operations: catalog, check, read_cdc, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- cdc_modes: snapshot, postgres_logical_replication
- cdc_state_fields: lsn, slot_name, publication_name, snapshot_completed
- conformance: catalog, cdc_checkpoint, cdc_setup_validation, check, delete_semantics, docs_skill, ordering, read_fixture, secret_redaction, snapshot_consistency, spec, state_checkpoint

## Official Application Documentation

- PostgreSQL documentation: https://www.postgresql.org/docs/
- PostgreSQL authentication: https://www.postgresql.org/docs/current/auth-methods.html
- Release Notes: https://www.postgresql.org/docs/release/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/postgres

## Configuration

- check_privileges (boolean): When this feature is enabled, during schema discovery the connector will query each table or view individually to check access privileges and inaccessible tables, views, or colu...
- checkpoint_target_interval_seconds (integer): How often (in seconds) a stream should checkpoint, when possible.
- database (string) required: The name of the database to connect to.
- entra_client_id (string): If using Entra service principal, the application ID of the service principal
- entra_service_principal_auth (boolean): Interpret password as a client secret for a Microsoft Entra service principal
- entra_tenant_id (string): If using Entra service principal, the ID of the tenant
- host (string) required: Hostname of the database.
- jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
- max_db_connections (integer): Maximum number of concurrent queries to the database. Leave empty to let Airbyte optimize performance.
- password (string) secret: The password associated with the username.
- port (integer) required: Port of the database. Defaults to 5432.
- replication_method (object) required: Configures how data is extracted from the database.
- schemas (array): The list of schemas to sync from. Case sensitive. Empty means all schemas.
- ssl_mode (object): The encryption method which is used when communicating with the database.
- tunnel_method (object) required: Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
- username (string) required: The username which is used to access the database.
- secret fields: password, ssl_mode.ca_certificate, ssl_mode.client_certificate, ssl_mode.client_key, ssl_mode.client_key_password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/postgres

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-postgres
```

### Inspect as JSON

```bash
pm connectors inspect source-postgres --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Postgres documentation](https://docs.airbyte.com/integrations/sources/postgres)
