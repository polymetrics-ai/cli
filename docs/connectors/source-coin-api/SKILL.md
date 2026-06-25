---
name: pm-source-coin-api
description: Coin API connector knowledge and safe action guide.
---

# pm-source-coin-api

## Purpose

Coin API catalog connector for https://docs.airbyte.com/integrations/sources/coin-api. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-coin-api:0.3.22 (metadata only; not executed)

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

- CoinAPI documentation: https://docs.coinapi.io/
- CoinAPI authentication: https://docs.coinapi.io/#authentication
- CoinAPI rate limits: https://docs.coinapi.io/#rate-limits
- CoinAPI Status: https://status.coinapi.io/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/coin-api

## Configuration

- api_key (string) required secret: API Key
- end_date (string): The end date in ISO 8601 format. If not supplied, data will be returned from the start date to the current time, or when the count of result elements reaches its limit.
- environment (string) required: The environment to use. Either sandbox or production.
- limit (integer): The maximum number of elements to return. If not supplied, the default is 100. For numbers larger than 100, each 100 items is counted as one request for pricing purposes. Maximu...
- period (string) required: The period to use. See the documentation for a list. https://docs.coinapi.io/#list-all-periods-get
- start_date (string) required: The start date in ISO 8601 format.
- symbol_id (string) required: The symbol ID to use. See the documentation for a list. https://docs.coinapi.io/#list-all-symbols-get
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/coin-api

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-coin-api
```

### Inspect as JSON

```bash
pm connectors inspect source-coin-api --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Coin API documentation](https://docs.airbyte.com/integrations/sources/coin-api)
