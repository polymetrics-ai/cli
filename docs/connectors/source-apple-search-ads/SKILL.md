---
name: pm-source-apple-search-ads
description: Apple Ads connector knowledge and safe action guide.
---

# pm-source-apple-search-ads

## Purpose

Apple Ads catalog connector for https://docs.airbyte.com/integrations/sources/apple-search-ads. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-apple-search-ads:1.1.8 (metadata only; not executed)

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

- API reference: https://developer.apple.com/documentation/apple_search_ads
- Authentication: https://developer.apple.com/documentation/apple_search_ads/implementing_oauth_for_the_apple_search_ads_api
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/apple-search-ads

## Configuration

- backoff_factor (integer): This factor factor determines the delay increase factor between retryable failures. Valid values are integers between 1 and 20.
- client_id (string) required secret: A user identifier for the token request. See <a href="https://developer.apple.com/documentation/apple_search_ads/implementing_oauth_for_the_apple_search_ads_api">here</a>
- client_secret (string) required secret: A string that authenticates the user’s setup request. See <a href="https://developer.apple.com/documentation/apple_search_ads/implementing_oauth_for_the_apple_search_ads_api">...
- end_date (string): Data is retrieved until that date (included)
- lookback_window (integer): Apple Search Ads uses a 30-day attribution window. However, you may consider smaller values in order to shorten sync durations, at the cost of missing late data attributions.
- num_workers (integer): The number of concurrent workers for syncing data. Increase this value to speed up syncs for accounts with many campaigns and ad groups, at the cost of higher API usage. Valid v...
- org_id (integer) required: The identifier of the organization that owns the campaign. Your Org Id is the same as your account in the Apple Search Ads UI.
- start_date (string) required: Start getting data from that date.
- timezone (string) required: The timezone for the reporting data. Use 'ORTZ' for Organization Time Zone or 'UTC' for Coordinated Universal Time. Default is UTC.
- token_refresh_endpoint (string) required: Token Refresh Endpoint. You should override the default value in scenarios where it's required to proxy requests to Apple's token endpoint
- secret fields: client_id, client_secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/apple-search-ads

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-apple-search-ads
```

### Inspect as JSON

```bash
pm connectors inspect source-apple-search-ads --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Apple Ads documentation](https://docs.airbyte.com/integrations/sources/apple-search-ads)
