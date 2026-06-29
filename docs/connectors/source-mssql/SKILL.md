---
name: pm-source-mssql
description: Microsoft SQL Server (MSSQL) connector knowledge and safe action guide.
---

# pm-source-mssql

## Purpose

Microsoft SQL Server (MSSQL) catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: certified

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

- family: database_cdc_source
- priority_wave: 3
- etl_operations: catalog, check, read_cdc, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- cdc_modes: snapshot, sql_server_cdc
- cdc_state_fields: lsn, capture_instance, snapshot_completed
- conformance: catalog, cdc_checkpoint, cdc_setup_validation, check, delete_semantics, docs_skill, ordering, read_fixture, secret_redaction, snapshot_consistency, spec, state_checkpoint

## Official Application Documentation

- SQL Server documentation: https://learn.microsoft.com/en-us/sql/sql-server/
- SQL Server authentication: https://learn.microsoft.com/en-us/sql/relational-databases/security/choose-an-authentication-mode
- SQL Server 2022 release notes: https://learn.microsoft.com/en-us/sql/sql-server/sql-server-2022-release-notes

## Configuration

- additionalProperties (object)
- check_privileges (boolean): When this feature is enabled, during schema discovery the connector will query each table or view individually to check access privileges and inaccessible tables, views, or colu...
- checkpoint_target_interval_seconds (integer): How often (in seconds) a stream should checkpoint, when possible.
- client_id (string): Application (client) ID of a Microsoft Entra ID service principal. When provided together with Client Secret, Entra ID authentication is used instead of username and password.
- client_secret (string) secret: Client secret for the Microsoft Entra ID service principal. When provided together with Client ID, Entra ID authentication is used instead of username and password.
- concurrency (integer): Maximum number of concurrent queries to the database.
- database (string) required: The name of the database.
- host (string) required: The hostname of the database.
- jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
- password (string) secret: The password associated with the username. Not required if Microsoft Entra ID authentication is configured below.
- port (integer) required: The port of the database.
- replication_method (object) required: Configures how data is extracted from the database.
- schemas (array): The list of schemas to sync from. If not specified, all schemas will be discovered. Case sensitive.
- ssl_mode (object): The encryption method which is used when communicating with the database.
- tenant_id (string): Optional Microsoft Entra tenant ID. If omitted, the driver uses the tenant inferred from the service principal.
- tunnel_method (object): Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
- username (string): The username which is used to access the database. Not required if Microsoft Entra ID authentication is configured below.
- secret fields: client_secret, password, ssl_mode.certificate, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

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
pm connectors inspect source-mssql
```

### Inspect as JSON

```bash
pm connectors inspect source-mssql --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [SQL Server documentation](https://learn.microsoft.com/en-us/sql/sql-server/)
- [SQL Server authentication](https://learn.microsoft.com/en-us/sql/relational-databases/security/choose-an-authentication-mode)
- [SQL Server 2022 release notes](https://learn.microsoft.com/en-us/sql/sql-server/sql-server-2022-release-notes)
