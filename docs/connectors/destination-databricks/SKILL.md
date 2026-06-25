---
name: pm-destination-databricks
description: Databricks Lakehouse connector knowledge and safe action guide.
---

# pm-destination-databricks

## Purpose

Databricks Lakehouse catalog connector for https://docs.airbyte.com/integrations/destinations/databricks. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-databricks:3.3.8 (metadata only; not executed)

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

- SQL reference: https://docs.databricks.com/sql/language-manual/index.html
- Authentication: https://docs.databricks.com/dev-tools/auth.html
- Access control: https://docs.databricks.com/security/access-control/index.html
- Release notes: https://docs.databricks.com/release-notes/index.html
- Databricks Status: https://status.databricks.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/databricks

## Configuration

- accept_terms (boolean) required: You must agree to the Databricks JDBC Driver <a href="https://databricks.com/jdbc-odbc-driver-license">Terms & Conditions</a> to use this connector.
- authentication (object) required: Authentication mechanism for Staging files and running queries
- database (string) required: The name of the unity catalog for the database
- hostname (string) required: Databricks Cluster Server Hostname.
- http_path (string) required: Databricks Cluster HTTP Path.
- port (string): Databricks Cluster Port.
- purge_staging_data (boolean): Default to 'true'. Switch it to 'false' for debugging purpose.
- raw_schema_override (string): The schema to write raw tables into (default: airbyte_internal)
- schema (string): The default schema tables are written. If not specified otherwise, the "default" will be used.
- secret fields: authentication.personal_access_token, authentication.secret

## Sync Modes

- supported sync modes: append, append_dedup, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/databricks

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-databricks
```

### Inspect as JSON

```bash
pm connectors inspect destination-databricks --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Databricks Lakehouse documentation](https://docs.airbyte.com/integrations/destinations/databricks)
