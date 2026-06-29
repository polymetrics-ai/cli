---
name: pm-source-bing-ads
description: Bing Ads connector knowledge and safe action guide.
---

# pm-source-bing-ads

## Purpose

Bing Ads catalog connector. Native implementation status: planned_native_port.

## Icon

- asset: icons/bingads.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://learn.microsoft.com/en-us/advertising/guides/release-notes

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

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
- priority_wave: 1
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- Bing Ads API Release Notes: https://learn.microsoft.com/en-us/advertising/guides/release-notes
- Release notes: https://learn.microsoft.com/en-us/advertising/guides/release-notes?view=bingads-13

## Configuration

- account_names (array): Predicates that will be used to sync data by specific accounts.
- auth_method (string)
- client_id (string) required secret: The Client ID of your Microsoft Advertising developer application.
- client_secret (string) secret: The Client Secret of your Microsoft Advertising developer application.
- custom_reports (array): You can add your Custom Bing Ads report by creating one.
- developer_token (string) required secret: Developer token associated with user. See more info <a href="https://docs.microsoft.com/en-us/advertising/guides/get-started?view=bingads-13#get-developer-token"> in the docs</a>.
- lookback_window (integer): Also known as attribution or conversion window. How far into the past to look for records (in days). If your conversion window has an hours/minutes granularity, round it up to t...
- num_workers (integer): The number of worker threads to use for the sync. Increase this to speed up syncs for accounts with many reports. The default should work for most use cases.
- refresh_token (string) required secret: Refresh Token to renew the expired Access Token.
- reports_start_date (string): The start date from which to begin replicating report data. Any data generated before this date will not be replicated in reports. This is a UTC date in YYYY-MM-DD format. If no...
- tenant_id (string) secret: The Tenant ID of your Microsoft Advertising developer application. Set this to "common" unless you know you need a different value.
- secret fields: client_id, client_secret, developer_token, refresh_token, tenant_id

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
pm connectors inspect source-bing-ads
```

### Inspect as JSON

```bash
pm connectors inspect source-bing-ads --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Bing Ads API Release Notes](https://learn.microsoft.com/en-us/advertising/guides/release-notes)
- [Release notes](https://learn.microsoft.com/en-us/advertising/guides/release-notes?view=bingads-13)
