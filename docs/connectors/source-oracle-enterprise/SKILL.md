---
name: pm-source-oracle-enterprise
description: Oracle connector knowledge and safe action guide.
---

# pm-source-oracle-enterprise

## Purpose

Oracle catalog connector for https://docs.airbyte.com/integrations/enterprise-connectors/source-oracle-enterprise. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-oracle-enterprise:0.1.4 (metadata only; not executed)

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
- cdc_modes: snapshot, oracle_logminer_or_xstream
- cdc_state_fields: scn, snapshot_completed
- conformance: catalog, cdc_checkpoint, cdc_setup_validation, check, delete_semantics, docs_skill, ordering, read_fixture, secret_redaction, snapshot_consistency, spec, state_checkpoint

## Official Application Documentation

- No upstream application documentation URL was listed in the imported connector registry.
- Airbyte connector documentation: https://docs.airbyte.com/integrations/enterprise-connectors/source-oracle-enterprise

## Configuration

- additionalProperties (object)
- check_privileges (boolean): When this feature is enabled, during schema discovery the connector will query each table or view individually to check access privileges and inaccessible tables, views, or colu...
- checkpoint_target_interval_seconds (integer): How often (in seconds) a stream should checkpoint, when possible.
- concurrency (integer): Maximum number of concurrent queries to the database.
- connection_data (object) required: The scheme by which to establish a database connection.
- cursor (object) required: Configures how data is extracted from the database.
- encryption (object) required: The encryption method with is used when communicating with the database.
- host (string) required: Hostname of the database.
- jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
- max_db_connections (integer): Maximum number of concurrent queries to the database. Leave empty to let Airbyte optimize performance.
- password (string) secret: The password associated with the username.
- port (integer) required: Port of the database. Oracle Corporations recommends the following port numbers: 1521 - Default listening port for client connections to the listener. 2484 - Recommended and off...
- schemas (array): The list of schemas to sync from. Defaults to user. Case sensitive.
- table_filters (array): Inclusion filters for table selection per schema. If no filters are specified for a schema, all tables in that schema will be synced.
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

- https://docs.airbyte.com/integrations/enterprise-connectors/source-oracle-enterprise

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-oracle-enterprise
```

### Inspect as JSON

```bash
pm connectors inspect source-oracle-enterprise --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Oracle documentation](https://docs.airbyte.com/integrations/enterprise-connectors/source-oracle-enterprise)
