---
name: pm-source-tiktok-marketing
description: TikTok Marketing connector knowledge and safe action guide.
---

# pm-source-tiktok-marketing

## Purpose

TikTok Marketing catalog connector for https://docs.airbyte.com/integrations/sources/tiktok-marketing. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-tiktok-marketing:5.1.1 (metadata only; not executed)

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

- Versioning docs: https://business-api.tiktok.com/portal/docs?id=1740029169927169
- Changelog: https://business-api.tiktok.com/portal/docs?id=1740029165513730
- TikTok Business API Documentation: https://business-api.tiktok.com/portal/docs?id=1740302848670722
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/tiktok-marketing

## Configuration

- attribution_window (integer): The attribution window in days.
- credentials (object): Authentication method
- end_date (string): The date until which you'd like to replicate data for all incremental streams, in the format YYYY-MM-DD. All data generated between start_date and this date will be replicated. ...
- include_deleted (boolean): Set to active if you want to include deleted data in report based streams and Ads, Ad Groups and Campaign streams.
- report_granularity (integer): The number of days per API request for daily report streams. Use the default 30 for most accounts. If syncs fail with TikTok API error 40067 ("query too large"), reduce this val...
- start_date (string): The Start Date in format: YYYY-MM-DD. Any data before this date will not be replicated. If this parameter is not set, all data will be replicated.
- secret fields: credentials.access_token, credentials.app_id, credentials.secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/tiktok-marketing

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-tiktok-marketing
```

### Inspect as JSON

```bash
pm connectors inspect source-tiktok-marketing --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [TikTok Marketing documentation](https://docs.airbyte.com/integrations/sources/tiktok-marketing)
