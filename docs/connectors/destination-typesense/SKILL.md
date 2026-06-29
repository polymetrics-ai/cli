---
name: pm-destination-typesense
description: Typesense connector knowledge and safe action guide.
---

# pm-destination-typesense

## Purpose

Typesense catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/typesense.svg
- source: official
- review_status: official_verified
- review_url: https://typesense.org/docs/

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
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
- priority_wave: 3
- etl_operations: catalog, check, write_append, write_dedup, write_overwrite
- reverse_etl_operations: none until native write conformance passes
- conformance: approval_policy, batch_write, catalog, check, dedup_write, docs_skill, idempotency, overwrite_write, secret_redaction, spec, write_fixture

## Official Application Documentation

- Typesense documentation: https://typesense.org/docs/

## Configuration

- api_key (string) required: Typesense API Key
- batch_size (integer): How many documents should be imported together. Default 1000
- host (string) required: Hostname of the Typesense instance without protocol. Accept multiple hosts separated by comma.
- path (string): Path of the Typesense instance. Default is none
- port (string): Port of the Typesense instance. Ex: 8108, 80, 443. Default is 8108
- protocol (string): Protocol of the Typesense instance. Ex: http or https. Default is https

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
pm connectors inspect destination-typesense
```

### Inspect as JSON

```bash
pm connectors inspect destination-typesense --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Typesense documentation](https://typesense.org/docs/)
