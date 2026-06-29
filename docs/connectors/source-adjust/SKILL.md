---
name: pm-source-adjust
description: Adjust connector knowledge and safe action guide.
---

# pm-source-adjust

## Purpose

Adjust catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/adjust.svg
- source: official
- review_status: official_verified
- review_url: https://dev.adjust.com/en/api/rs-api/

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

- Adjust documentation: https://dev.adjust.com/en/api/rs-api/

## Configuration

- additional_metrics (array): Metrics names that are not pre-defined, such as cohort metrics or app specific metrics.
- api_token (string) required secret: Adjust API key, see https://help.adjust.com/en/article/report-service-api-authentication
- dimensions (array) required: Dimensions allow a user to break down metrics into groups using one or several parameters. For example, the number of installs by date, country and network. See https://help.adj...
- ingest_start (string) required: Data ingest start date.
- metrics (array) required: Select at least one metric to query.
- until_today (boolean): Syncs data up until today. Useful when running daily incremental syncs, and duplicates are not desired.
- secret fields: api_token

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
pm connectors inspect source-adjust
```

### Inspect as JSON

```bash
pm connectors inspect source-adjust --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Adjust documentation](https://dev.adjust.com/en/api/rs-api/)
