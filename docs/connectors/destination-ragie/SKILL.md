---
name: pm-destination-ragie
description: Ragie connector knowledge and safe action guide.
---

# pm-destination-ragie

## Purpose

Ragie catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/pm-warehouse.svg
- source: polymetrics
- review_status: polymetrics

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

- Ragie documentation: https://docs.ragie.ai/docs/getting-started

## Configuration

- api_key (string) required secret: API Key for Ragie.ai.
- api_url (string): URL for the Ragie API. Defaults to https://api.ragie.ai
- content_fields (array): (Optional) List of fields from the record to use as the main document content. If empty, the entire record is used. Use dot notation for nested fields (e.g., 'user.profile').
- document_name_field (anyOf): (Optional) Field from the record to use as the document name. If empty or field not found, a name is auto-generated.
- external_id_field (anyOf): (Optional) Field from the record to use as the unique 'external_id' for Ragie documents.
- metadata_fields (array): (Optional) List of fields from the record to store as metadata. If empty, no record fields are added as metadata. Use dot notation.
- metadata_static (anyOf): (Optional) Static key-value pairs as a JSON object string to add to every document's metadata.
- partition (anyOf): (Optional) Name of the partition (index/dataset) to write data into. Must be lowercase alphanumeric with '-' or '_'. If empty, uses default.
- processing_mode (string): Processing mode for ingestion ('fast' or 'hi-res').
- secret fields: api_key

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
pm connectors inspect destination-ragie
```

### Inspect as JSON

```bash
pm connectors inspect destination-ragie --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Ragie documentation](https://docs.ragie.ai/docs/getting-started)
