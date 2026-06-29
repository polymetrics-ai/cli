---
name: pm-source-recharge
description: Recharge connector knowledge and safe action guide.
---

# pm-source-recharge

## Purpose

Recharge catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/recharge.svg
- source: official
- review_status: official_verified
- review_url: https://docs.getrecharge.com/

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Recharge documentation: https://docs.getrecharge.com/

## Configuration

- access_token (string) required secret: manual intervention needed
- lookback_window_days (integer): Specifies how many days of historical data should be reloaded each time the recharge connector runs.
- start_date (string) required: The date from which you'd like to replicate data for Recharge API, in the format YYYY-MM-DDT00:00:00Z. Any data before this date will not be replicated.
- use_orders_deprecated_api (boolean): Define whether or not the `Orders` stream should use the deprecated `2021-01` API version, or use `2021-11`, otherwise.
- secret fields: access_token

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
pm connectors inspect source-recharge
```

### Inspect as JSON

```bash
pm connectors inspect source-recharge --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Recharge documentation](https://docs.getrecharge.com/)
