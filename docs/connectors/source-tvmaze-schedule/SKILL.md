---
name: pm-source-tvmaze-schedule
description: TVMaze Schedule connector knowledge and safe action guide.
---

# pm-source-tvmaze-schedule

## Purpose

TVMaze Schedule catalog connector for https://docs.airbyte.com/integrations/sources/tvmaze-schedule. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-tvmaze-schedule:0.2.25 (metadata only; not executed)

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

- TVmaze API documentation: https://www.tvmaze.com/api
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/tvmaze-schedule

## Configuration

- domestic_schedule_country_code (string) required: Country code for domestic TV schedule retrieval.
- end_date (string): End date for TV schedule retrieval. May be in the future. Optional.
- start_date (string) required: Start date for TV schedule retrieval. May be in the future.
- web_schedule_country_code (string): ISO 3166-1 country code for web TV schedule retrieval. Leave blank for all countries plus global web channels (e.g. Netflix). Alternatively, set to 'global' for just global web ...

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/tvmaze-schedule

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-tvmaze-schedule
```

### Inspect as JSON

```bash
pm connectors inspect source-tvmaze-schedule --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [TVMaze Schedule documentation](https://docs.airbyte.com/integrations/sources/tvmaze-schedule)
