---
name: pm-source-sap-hana-enterprise
description: SAP HANA connector knowledge and safe action guide.
---

# pm-source-sap-hana-enterprise

## Purpose

SAP HANA catalog connector for https://docs.airbyte.com/integrations/enterprise-connectors/source-sap-hana. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-sap-hana-enterprise:0.0.18 (metadata only; not executed)

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

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/enterprise-connectors/source-sap-hana

## Configuration

- check_privileges (boolean): When this feature is enabled, during schema discovery the connector will query each table or view individually to check access privileges and inaccessible tables, views, or colu...
- checkpoint_target_interval_seconds (integer): How often (in seconds) a stream should checkpoint, when possible.
- concurrency (integer): Maximum number of concurrent queries to the database.
- cursor (object) required: Configures how data is extracted from the database.
- database (string): The name of the tenant database to connect to. This is required for multi-tenant SAP HANA systems. For single-tenant systems, this can be left empty.
- encryption (object) required: The encryption method with is used when communicating with the database.
- filters (array): Inclusion filters for table selection per schema. If no filters are specified for a schema, all tables in that schema will be synced.
- host (string) required: Hostname of the database.
- jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
- password (string) secret: The password associated with the username.
- port (integer) required: Port of the database. SAP recommends the following port numbers: 443 - Default listening port for SAP HANA Cloud client connections to the listener.
- schemas (array): The list of schemas to sync from. Defaults to user. Case sensitive.
- tunnel_method (object) required: Whether to initiate an SSH tunnel before connecting to the database, and if so, which kind of authentication to use.
- username (string) required: The username which is used to access the database.
- secret fields: encryption.ssl_certificate, password, tunnel_method.ssh_key, tunnel_method.tunnel_user_password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/enterprise-connectors/source-sap-hana

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-sap-hana-enterprise
```

### Inspect as JSON

```bash
pm connectors inspect source-sap-hana-enterprise --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [SAP HANA documentation](https://docs.airbyte.com/integrations/enterprise-connectors/source-sap-hana)
