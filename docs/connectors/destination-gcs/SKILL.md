---
name: pm-destination-gcs
description: Google Cloud Storage (GCS) connector knowledge and safe action guide.
---

# pm-destination-gcs

## Purpose

Google Cloud Storage (GCS) catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/googlecloudstorage.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://cloud.google.com/storage/docs

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: beta
- support level: community

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
- priority_wave: 2
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- Cloud Storage documentation: https://cloud.google.com/storage/docs
- Service account authentication: https://cloud.google.com/iam/docs/service-accounts
- Access control: https://cloud.google.com/storage/docs/access-control
- Google Cloud Release Notes: https://cloud.google.com/release-notes
- Quotas and limits: https://cloud.google.com/storage/quotas
- Google Cloud Status: https://status.cloud.google.com/

## Configuration

- credential (object) required: An HMAC key is a type of credential and can be associated with a service account or a user account in Cloud Storage. Read more <a href="https://cloud.google.com/storage/docs/aut...
- format (object) required: Output data format. One of the following formats must be selected - <a href="https://cloud.google.com/bigquery/docs/loading-data-cloud-storage-avro#advantages_of_avro">AVRO</a> ...
- gcs_bucket_name (string) required: You can find the bucket name in the App Engine Admin console Application Settings page, under the label Google Cloud Storage Bucket. Read more <a href="https://cloud.google.com/...
- gcs_bucket_path (string) required: GCS Bucket Path string Subdirectory under the above bucket to sync the data into.
- gcs_bucket_region (string): Select a Region of the GCS Bucket. Read more <a href="https://cloud.google.com/storage/docs/locations">here</a>.
- secret fields: credential.hmac_key_access_id, credential.hmac_key_secret

## Sync Modes

- supported sync modes: append, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-gcs
```

### Inspect as JSON

```bash
pm connectors inspect destination-gcs --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Cloud Storage documentation](https://cloud.google.com/storage/docs)
- [Service account authentication](https://cloud.google.com/iam/docs/service-accounts)
- [Access control](https://cloud.google.com/storage/docs/access-control)
- [Google Cloud Release Notes](https://cloud.google.com/release-notes)
- [Quotas and limits](https://cloud.google.com/storage/quotas)
- [Google Cloud Status](https://status.cloud.google.com/)
