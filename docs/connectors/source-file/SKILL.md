---
name: pm-source-file
description: File (CSV, JSON, Excel, Feather, Parquet) connector knowledge and safe action guide.
---

# pm-source-file

## Purpose

File (CSV, JSON, Excel, Feather, Parquet) catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/file.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: bounded_streaming, catalog, check, docs_skill, format_detection, path_safety, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- manual intervention needed

## Configuration

- dataset_name (string) required: The Name of the final table to replicate this file into (should include letters, numbers dash and underscores only).
- format (string) required: The Format of the file which should be replicated (Warning: some formats may be experimental, please refer to the docs).
- provider (object) required: The storage Provider or Location of the file(s) which should be replicated.
- reader_options (string): This should be a string in JSON format. It depends on the chosen file format to provide additional options and tune its behavior.
- url (string) required: The URL path to access the file which should be replicated.
- secret fields: provider.aws_secret_access_key, provider.password, provider.sas_token, provider.service_account_json, provider.shared_key

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
pm connectors inspect source-file
```

### Inspect as JSON

```bash
pm connectors inspect source-file --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.
