---
name: pm-source-delighted
description: Delighted connector knowledge and safe action guide.
---

# pm-source-delighted

## Purpose

Delighted catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/delighted.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://delighted.com/docs/api

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: beta
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
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

- family: declarative_http_source
- priority_wave: 2
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Delighted API reference: https://delighted.com/docs/api
- Delighted authentication: https://delighted.com/docs/api#authentication
- Delighted rate limits: https://delighted.com/docs/api#rate-limiting

## Configuration

- api_key (string) required secret: A Delighted API key.
- since (string) required: The date from which you'd like to replicate the data
- secret fields: api_key

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
pm connectors inspect source-delighted
```

### Inspect as JSON

```bash
pm connectors inspect source-delighted --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Delighted API reference](https://delighted.com/docs/api)
- [Delighted authentication](https://delighted.com/docs/api#authentication)
- [Delighted rate limits](https://delighted.com/docs/api#rate-limiting)
