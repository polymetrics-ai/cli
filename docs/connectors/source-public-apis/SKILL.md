---
name: pm-source-public-apis
description: Public Apis connector knowledge and safe action guide.
---

# pm-source-public-apis

## Purpose

Public Apis catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/public-apis.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://github.com/public-apis/public-apis

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
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

- family: custom_go_port
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Public APIs directory: https://github.com/public-apis/public-apis

## Configuration

- No config schema fields were advertised in the catalog.

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
pm connectors inspect source-public-apis
```

### Inspect as JSON

```bash
pm connectors inspect source-public-apis --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Public APIs directory](https://github.com/public-apis/public-apis)
