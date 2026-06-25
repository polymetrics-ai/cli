---
name: pm-source-linkedin-ads
description: LinkedIn Ads connector knowledge and safe action guide.
---

# pm-source-linkedin-ads

## Purpose

LinkedIn Ads catalog connector for https://docs.airbyte.com/integrations/sources/linkedin-ads. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-linkedin-ads:5.6.9 (metadata only; not executed)

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

- Changelog: https://learn.microsoft.com/en-us/linkedin/marketing/integrations/recent-changes?view=li-lms-2024-10
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/linkedin-ads

## Configuration

- account_ids (array): Specify the account IDs to pull data from, separated by a space. Leave this field empty if you want to pull the data from all accounts accessible by the authenticated user. See ...
- ad_analytics_reports (array)
- credentials (object)
- lookback_window (integer): How far into the past to look for records. (in days)
- num_workers (integer): The number of workers to use for the connector. This is used to limit the number of concurrent requests to the LinkedIn Ads API. If not set, the default is 3 workers.
- start_date (string) required: UTC date in the format YYYY-MM-DD. Any data before this date will not be replicated.
- secret fields: credentials.access_token, credentials.client_id, credentials.client_secret, credentials.refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/linkedin-ads

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-linkedin-ads
```

### Inspect as JSON

```bash
pm connectors inspect source-linkedin-ads --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [LinkedIn Ads documentation](https://docs.airbyte.com/integrations/sources/linkedin-ads)
