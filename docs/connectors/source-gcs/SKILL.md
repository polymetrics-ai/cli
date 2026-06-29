---
name: pm-source-gcs
description: Google Cloud Storage (GCS) connector knowledge and safe action guide.
---

# pm-source-gcs

## Purpose

Google Cloud Storage (GCS) catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/gcs.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://cloud.google.com/storage/docs

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: file_go
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

- family: file_object_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: bounded_streaming, catalog, check, docs_skill, format_detection, path_safety, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Google Cloud Storage documentation: https://cloud.google.com/storage/docs
- GCS authentication: https://cloud.google.com/storage/docs/authentication
- Google Cloud Status: https://status.cloud.google.com/

## Configuration

- bucket (string) required: Name of the GCS bucket where the file(s) exist.
- credentials (object) required: Credentials for connecting to the Google Cloud Storage API
- delivery_method (object)
- sanitize_signed_urls (boolean): When enabled, removes credential-bearing query parameters from signed URLs in the _ab_source_file_url record field. Only relevant for Service Account authentication.
- start_date (string): UTC date and time in the format 2017-01-25T00:00:00.000000Z. Any file modified before this date will not be replicated.
- streams (array) required: manual intervention needed
- secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token, credentials.service_account, streams[].format.processing.api_key

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
pm connectors inspect source-gcs
```

### Inspect as JSON

```bash
pm connectors inspect source-gcs --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Google Cloud Storage documentation](https://cloud.google.com/storage/docs)
- [GCS authentication](https://cloud.google.com/storage/docs/authentication)
- [Google Cloud Status](https://status.cloud.google.com/)
