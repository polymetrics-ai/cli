---
name: pm-destination-bigquery
description: BigQuery connector knowledge and safe action guide.
---

# pm-destination-bigquery

## Purpose

BigQuery catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/bigquery.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://cloud.google.com/bigquery/docs/release-notes

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
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

- family: destination_writer
- priority_wave: 1
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- Standard SQL reference: https://cloud.google.com/bigquery/docs/reference/standard-sql
- Service account authentication: https://cloud.google.com/iam/docs/service-accounts
- Access control and permissions: https://cloud.google.com/bigquery/docs/access-control
- Release notes: https://cloud.google.com/bigquery/docs/release-notes
- Quotas and limits: https://cloud.google.com/bigquery/quotas
- Google Cloud Status: https://status.cloud.google.com/

## Configuration

- cdc_deletion_mode (string): Whether to execute CDC deletions as hard deletes (i.e. propagate source deletions to the destination), or soft deletes (i.e. leave a tombstone record in the destination). Defaul...
- credentials_json (string) secret: manual intervention needed
- dataset_id (string) required: The default BigQuery Dataset ID that tables are replicated to if the source does not specify a namespace. Read more <a href="https://cloud.google.com/bigquery/docs/datasets#crea...
- dataset_location (string) required: The location of the dataset. Warning: Changes made after creation will not be applied. Read more <a href="https://cloud.google.com/bigquery/docs/locations">here</a>.
- disable_type_dedupe (boolean): Write the legacy "raw tables" format, to enable backwards compatibility with older versions of this connector.
- loading_method (object): The way data will be uploaded to BigQuery.
- project_id (string) required: The GCP project ID for the project containing the target BigQuery dataset. Read more <a href="https://cloud.google.com/resource-manager/docs/creating-managing-projects#identifyi...
- raw_data_dataset (string): manual intervention needed
- secret fields: credentials_json, loading_method.credential.hmac_key_access_id, loading_method.credential.hmac_key_secret

## Sync Modes

- supported sync modes: append, append_dedup, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-bigquery
```

### Inspect as JSON

```bash
pm connectors inspect destination-bigquery --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Standard SQL reference](https://cloud.google.com/bigquery/docs/reference/standard-sql)
- [Service account authentication](https://cloud.google.com/iam/docs/service-accounts)
- [Access control and permissions](https://cloud.google.com/bigquery/docs/access-control)
- [Release notes](https://cloud.google.com/bigquery/docs/release-notes)
- [Quotas and limits](https://cloud.google.com/bigquery/quotas)
- [Google Cloud Status](https://status.cloud.google.com/)
