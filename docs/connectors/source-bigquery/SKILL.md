---
name: pm-source-bigquery
description: BigQuery connector knowledge and safe action guide.
---

# pm-source-bigquery

## Purpose

BigQuery catalog connector for https://docs.airbyte.com/integrations/sources/bigquery. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-bigquery:0.4.5 (metadata only; not executed)

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
- etl_operations: catalog, check, read_incremental, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

## Official Application Documentation

- BigQuery REST API: https://cloud.google.com/bigquery/docs/reference/rest
- Authentication: https://cloud.google.com/bigquery/docs/authentication
- Release notes: https://cloud.google.com/bigquery/docs/release-notes
- Quotas and limits: https://cloud.google.com/bigquery/quotas
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/bigquery

## Configuration

- credentials_json (string) required secret: The contents of your Service Account Key JSON file. See the <a href="https://docs.airbyte.com/integrations/sources/bigquery#setup-the-bigquery-source-in-airbyte">docs</a> for mo...
- dataset_id (string): The dataset ID to search for tables and views. If you are only loading data from one dataset, setting this option could result in much faster schema discovery.
- project_id (string) required: The GCP project ID for the project containing the target BigQuery dataset.
- secret fields: credentials_json

## Sync Modes

- supported sync modes: full_refresh, incremental
- supports incremental: true

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/bigquery

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-bigquery
```

### Inspect as JSON

```bash
pm connectors inspect source-bigquery --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [BigQuery documentation](https://docs.airbyte.com/integrations/sources/bigquery)
