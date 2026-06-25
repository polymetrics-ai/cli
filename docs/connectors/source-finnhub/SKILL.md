---
name: pm-source-finnhub
description: Finnhub connector knowledge and safe action guide.
---

# pm-source-finnhub

## Purpose

Finnhub catalog connector for https://docs.airbyte.com/integrations/sources/finnhub. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-finnhub:0.0.51 (metadata only; not executed)

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

- Finnhub API documentation: https://finnhub.io/docs/api
- Finnhub authentication: https://finnhub.io/docs/api/authentication
- Finnhub rate limits: https://finnhub.io/pricing
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/finnhub

## Configuration

- api_key (string) required secret: The API key to use for authentication
- exchange (string) required: More info: https://finnhub.io/docs/api/stock-symbols
- market_news_category (string) required: This parameter can be 1 of the following values general, forex, crypto, merger.
- start_date_2 (string) required
- symbols (array) required
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/finnhub

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-finnhub
```

### Inspect as JSON

```bash
pm connectors inspect source-finnhub --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Finnhub documentation](https://docs.airbyte.com/integrations/sources/finnhub)
