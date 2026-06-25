---
name: pm-destination-milvus
description: Milvus connector knowledge and safe action guide.
---

# pm-destination-milvus

## Purpose

Milvus catalog connector for https://docs.airbyte.com/integrations/destinations/milvus. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: destination
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: destination_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/destination-milvus:0.0.58 (metadata only; not executed)

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

- Milvus documentation: https://milvus.io/docs
- Authentication: https://milvus.io/docs/authenticate.md
- Release notes: https://milvus.io/docs/release_notes.md
- Airbyte connector documentation: https://docs.airbyte.com/integrations/destinations/milvus

## Configuration

- embedding (object) required: Embedding configuration
- indexing (object) required: Indexing configuration
- omit_raw_text (boolean): Do not store the text that gets embedded along with the vector and the metadata in the destination. If set to true, only the vector and the metadata will be stored - in this cas...
- processing (object) required
- secret fields: embedding.api_key, embedding.cohere_key, embedding.openai_key, indexing.auth.password, indexing.auth.token

## Sync Modes

- supported sync modes: append, append_dedup, overwrite
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/destinations/milvus

## Commands

### Inspect catalog entry

```bash
pm connectors inspect destination-milvus
```

### Inspect as JSON

```bash
pm connectors inspect destination-milvus --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Milvus documentation](https://docs.airbyte.com/integrations/destinations/milvus)
