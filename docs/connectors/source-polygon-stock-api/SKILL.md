---
name: pm-source-polygon-stock-api
description: Polygon Stock API connector knowledge and safe action guide.
---

# pm-source-polygon-stock-api

## Purpose

Polygon Stock API catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/polygon.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://polygon.io/docs/stocks/getting-started

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

- Polygon.io API reference: https://polygon.io/docs/stocks/getting-started
- Polygon.io rate limits: https://polygon.io/pricing

## Configuration

- adjusted (string): Determines whether or not the results are adjusted for splits. By default, results are adjusted and set to true. Set this to false to get results that are NOT adjusted for splits.
- apiKey (string) required secret: Your API ACCESS Key
- end_date (string) required: The target date for the aggregate window.
- limit (integer): The target date for the aggregate window.
- multiplier (integer) required: The size of the timespan multiplier.
- sort (string): Sort the results by timestamp. asc will return results in ascending order (oldest at the top), desc will return results in descending order (newest at the top).
- start_date (string) required: The beginning date for the aggregate window.
- stocksTicker (string) required: The exchange symbol that this item is traded under.
- timespan (string) required: The size of the time window.
- secret fields: apiKey

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
pm connectors inspect source-polygon-stock-api
```

### Inspect as JSON

```bash
pm connectors inspect source-polygon-stock-api --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Polygon.io API reference](https://polygon.io/docs/stocks/getting-started)
- [Polygon.io rate limits](https://polygon.io/pricing)
