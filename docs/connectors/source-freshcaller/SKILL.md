---
name: pm-source-freshcaller
description: Freshcaller connector knowledge and safe action guide.
---

# pm-source-freshcaller

## Purpose

Freshcaller catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/freshcaller.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.freshcaller.com/api/

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

- Freshcaller API reference: https://developers.freshcaller.com/api/
- Freshcaller authentication: https://developers.freshcaller.com/api/#authentication
- Freshworks Status: https://status.freshworks.com/

## Configuration

- api_key (string) required secret: Freshcaller API Key. See the docs for more information on how to obtain this key.
- domain (string) required: Used to construct Base URL for the Freshcaller APIs
- requests_per_minute (integer): The number of requests per minute that this source allowed to use. There is a rate limit of 50 requests per minute per app per account.
- start_date (string): UTC date and time. Any data created after this date will be replicated.
- sync_lag_minutes (integer): Lag in minutes for each sync, i.e., at time T, data for the time range [prev_sync_time, T-30] will be fetched
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
pm connectors inspect source-freshcaller
```

### Inspect as JSON

```bash
pm connectors inspect source-freshcaller --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Freshcaller API reference](https://developers.freshcaller.com/api/)
- [Freshcaller authentication](https://developers.freshcaller.com/api/#authentication)
- [Freshworks Status](https://status.freshworks.com/)
