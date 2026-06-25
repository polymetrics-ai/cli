---
name: pm-source-redshift
description: Redshift connector knowledge and safe action guide.
---

# pm-source-redshift

## Purpose

Redshift catalog connector for https://docs.airbyte.com/integrations/sources/redshift. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-redshift:0.5.5 (metadata only; not executed)

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

- Amazon Redshift documentation: https://docs.aws.amazon.com/redshift/
- Redshift authentication: https://docs.aws.amazon.com/redshift/latest/mgmt/generating-user-credentials.html
- AWS Status: https://health.aws.amazon.com/health/status
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/redshift

## Configuration

- database (string) required: Name of the database.
- host (string) required: Host Endpoint of the Redshift Cluster (must include the cluster-id, region and end with .redshift.amazonaws.com).
- jdbc_url_params (string): Additional properties to pass to the JDBC URL string when connecting to the database formatted as 'key=value' pairs separated by the symbol '&'. (example: key1=value1&key2=value...
- password (string) required secret: Password associated with the username.
- port (integer) required: Port of the database.
- schemas (array): The list of schemas to sync from. Specify one or more explicitly or keep empty to process all schemas. Schema names are case sensitive.
- username (string) required: Username to use to access the database.
- secret fields: password

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/redshift

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-redshift
```

### Inspect as JSON

```bash
pm connectors inspect source-redshift --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Redshift documentation](https://docs.airbyte.com/integrations/sources/redshift)
