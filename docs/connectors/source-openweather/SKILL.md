---
name: pm-source-openweather
description: Openweather connector knowledge and safe action guide.
---

# pm-source-openweather

## Purpose

Openweather catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/openweather.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://openweathermap.org/api

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

- OpenWeather API documentation: https://openweathermap.org/api

## Configuration

- appid (string) required secret: API KEY
- lang (string): You can use lang parameter to get the output in your language. The contents of the description field will be translated. See <a href="https://openweathermap.org/api/one-call-api...
- lat (string) required: Latitude, decimal (-90; 90). If you need the geocoder to automatic convert city names and zip-codes to geo coordinates and the other way around, please use the OpenWeather Geoco...
- lon (string) required: Longitude, decimal (-180; 180). If you need the geocoder to automatic convert city names and zip-codes to geo coordinates and the other way around, please use the OpenWeather Ge...
- only_current (boolean): True for particular day
- units (string): Units of measurement. standard, metric and imperial units are available. If you do not use the units parameter, standard units will be applied by default.
- secret fields: appid

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
pm connectors inspect source-openweather
```

### Inspect as JSON

```bash
pm connectors inspect source-openweather --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [OpenWeather API documentation](https://openweathermap.org/api)
