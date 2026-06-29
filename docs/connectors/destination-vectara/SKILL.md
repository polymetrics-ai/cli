---
name: pm-destination-vectara
description: Vectara connector knowledge and safe action guide.
---

# pm-destination-vectara

## Purpose

Vectara catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/vectara.svg
- source: official
- review_status: official_verified
- review_url: https://docs.vectara.com/docs/

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

- Vectara documentation: https://docs.vectara.com/docs/

## Configuration

- corpus_name (string) required: The Name of Corpus to load data into
- customer_id (string) required: Your customer id as it is in the authenticaion url
- metadata_fields (array): List of fields in the record that should be stored as metadata. The field list is applied to all streams in the same way and non-existing fields are ignored. If none are defined...
- oauth2 (object) required: OAuth2.0 credentials used to authenticate admin actions (creating/deleting corpora)
- parallelize (boolean): Parallelize indexing into Vectara with multiple threads
- text_fields (array): List of fields in the record that should be in the section of the document. The field list is applied to all streams in the same way and non-existing fields are ignored. If none...
- title_field (string): A field that will be used to populate the `title` of each document. The field list is applied to all streams in the same way and non-existing fields are ignored. If none are def...
- secret fields: oauth2.client_secret

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
pm connectors inspect destination-vectara
```

### Inspect as JSON

```bash
pm connectors inspect destination-vectara --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Vectara documentation](https://docs.vectara.com/docs/)
