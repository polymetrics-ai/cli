---
name: pm-source-freshdesk
description: Freshdesk connector knowledge and safe action guide.
---

# pm-source-freshdesk

## Purpose

Freshdesk catalog connector for https://docs.airbyte.com/integrations/sources/freshdesk. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-freshdesk:3.2.20 (metadata only; not executed)

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Changelog: https://developers.freshdesk.com/api/#change_log
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/freshdesk

## Configuration

- api_key (string) required secret: Freshdesk API Key. See the <a href="https://docs.airbyte.com/integrations/sources/freshdesk">docs</a> for more information on how to obtain this key.
- domain (string) required: Freshdesk domain
- lookback_window_in_days (integer): Number of days for lookback window for the stream Satisfaction Ratings
- num_workers (integer): Number of concurrent threads for syncing. Higher values can speed up syncs but may increase API rate limit usage. Adjust based on your Freshdesk API plan.
- rate_limit_plan (object): Rate Limit Plan for API Budget
- requests_per_minute (integer): The number of requests per minute that this source allowed to use. There is a rate limit of 50 requests per minute per app per account.
- start_date (string): UTC date and time. Any data created after this date will be replicated. If this parameter is not set, all data will be replicated.
- subscription_tier (string): Your API subscription tier (affects rate limits)
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/freshdesk

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-freshdesk
```

### Inspect as JSON

```bash
pm connectors inspect source-freshdesk --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Freshdesk documentation](https://docs.airbyte.com/integrations/sources/freshdesk)
