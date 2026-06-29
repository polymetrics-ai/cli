---
name: pm-source-weatherstack
description: Weatherstack connector knowledge and safe action guide.
---

# pm-source-weatherstack

## Purpose

Weatherstack catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/weatherstack.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://weatherstack.com/documentation

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

- Weatherstack API documentation: https://weatherstack.com/documentation

## Configuration

- access_key (string) required secret: API access key used to retrieve data from the Weatherstack API.(https://weatherstack.com/product)
- historical_date (string) required: This is required for enabling the Historical date API with format- (YYYY-MM-DD). * Note, only supported by paid accounts
- query (string) required: A location to query such as city, IP, latitudeLongitude, or zipcode. Multiple locations with semicolon seperated if using a professional plan or higher. For more info- (https://...
- secret fields: access_key

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
pm connectors inspect source-weatherstack
```

### Inspect as JSON

```bash
pm connectors inspect source-weatherstack --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Weatherstack API documentation](https://weatherstack.com/documentation)
