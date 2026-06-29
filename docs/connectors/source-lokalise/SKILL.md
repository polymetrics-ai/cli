---
name: pm-source-lokalise
description: Lokalise connector knowledge and safe action guide.
---

# pm-source-lokalise

## Purpose

Lokalise catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/lokalise.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.lokalise.com/reference/api-introduction

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
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
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Lokalise API reference: https://developers.lokalise.com/reference/api-introduction
- Lokalise authentication: https://developers.lokalise.com/reference/api-authentication
- Lokalise rate limits: https://developers.lokalise.com/reference/api-rate-limits

## Configuration

- api_key (string) required secret: Lokalise API Key with read-access. Available at Profile settings > API tokens. See <a href="https://docs.lokalise.com/en/articles/1929556-api-tokens">here</a>.
- project_id (string) required: Lokalise project ID. Available at Project Settings > General.
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
pm connectors inspect source-lokalise
```

### Inspect as JSON

```bash
pm connectors inspect source-lokalise --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Lokalise API reference](https://developers.lokalise.com/reference/api-introduction)
- [Lokalise authentication](https://developers.lokalise.com/reference/api-authentication)
- [Lokalise rate limits](https://developers.lokalise.com/reference/api-rate-limits)
