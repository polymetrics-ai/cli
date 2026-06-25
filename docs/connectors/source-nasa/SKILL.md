---
name: pm-source-nasa
description: Nasa connector knowledge and safe action guide.
---

# pm-source-nasa

## Purpose

Nasa catalog connector for https://docs.airbyte.com/integrations/sources/nasa. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-nasa:0.3.32 (metadata only; not executed)

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

- NASA APIs: https://api.nasa.gov/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/nasa

## Configuration

- api_key (string) required secret: API access key used to retrieve data from the NASA APOD API.
- concept_tags (boolean): Indicates whether concept tags should be returned with the rest of the response. The concept tags are not necessarily included in the explanation, but rather derived from common...
- count (integer): A positive integer, no greater than 100. If this is specified then `count` randomly chosen images will be returned in a JSON array. Cannot be used in conjunction with `date` or ...
- end_date (string): Indicates that end of a date range. If `start_date` is specified without an `end_date` then `end_date` defaults to the current date.
- start_date (string): Indicates the start of a date range. All images in the range from `start_date` to `end_date` will be returned in a JSON array. Must be after 1995-06-16, the first day an APOD pi...
- thumbs (boolean): Indicates whether the API should return a thumbnail image URL for video files. If set to True, the API returns URL of video thumbnail. If an APOD is not a video, this parameter ...
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/nasa

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-nasa
```

### Inspect as JSON

```bash
pm connectors inspect source-nasa --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Nasa documentation](https://docs.airbyte.com/integrations/sources/nasa)
