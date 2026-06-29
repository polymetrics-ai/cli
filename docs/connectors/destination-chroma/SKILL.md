---
name: pm-destination-chroma
description: Chroma connector knowledge and safe action guide.
---

# pm-destination-chroma

## Purpose

Chroma catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/chroma.svg
- source: official
- review_status: official_verified
- review_url: https://docs.trychroma.com/docs/overview/getting-started

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

- Chroma documentation: https://docs.trychroma.com/docs/overview/getting-started

## Configuration

- embedding (object) required: Embedding configuration
- indexing (object) required: Indexing configuration
- omit_raw_text (boolean): Do not store the text that gets embedded along with the vector and the metadata in the destination. If set to true, only the vector and the metadata will be stored - in this cas...
- processing (object) required
- secret fields: embedding.api_key, embedding.cohere_key, embedding.openai_key, indexing.auth_method.password

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
pm connectors inspect destination-chroma
```

### Inspect as JSON

```bash
pm connectors inspect destination-chroma --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Chroma documentation](https://docs.trychroma.com/docs/overview/getting-started)
