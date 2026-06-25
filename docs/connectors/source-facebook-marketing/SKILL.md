---
name: pm-source-facebook-marketing
description: Facebook Marketing connector knowledge and safe action guide.
---

# pm-source-facebook-marketing

## Purpose

Facebook Marketing catalog connector for https://docs.airbyte.com/integrations/sources/facebook-marketing. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-facebook-marketing:6.0.1 (metadata only; not executed)

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
- etl_operations: catalog, check, read_incremental, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, docs_skill, read_fixture, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- API Upgrade Tool: https://developers.facebook.com/tools/api_versioning/600551260845577/
- Changelog: https://developers.facebook.com/docs/marketing-api/marketing-api-changelog
- Graph API Changelog: https://developers.facebook.com/docs/graph-api/changelog
- 2026 Out-Of-Cycle Changes: https://developers.facebook.com/documentation/ads-commerce/marketing-api/out-of-cycle-changes/occ-2026
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/facebook-marketing

## Configuration

- access_token (string) secret: The value of the generated access token. From your App’s Dashboard, click on "Marketing API" then "Tools". Select permissions <b>ads_management, ads_read, read_insights, busin...
- account_ids (array) required: The Facebook Ad account ID(s) to pull data from. The Ad account ID number is in the account dropdown menu or in your browser's address bar of your <a href="https://adsmanager.fa...
- action_breakdowns_allow_empty (boolean): Allows action_breakdowns to be an empty list
- ad_statuses (array): Select the statuses you want to be loaded in the stream. If no specific statuses are selected, the API's default behavior applies, and some statuses may be filtered out.
- adset_statuses (array): Select the statuses you want to be loaded in the stream. If no specific statuses are selected, the API's default behavior applies, and some statuses may be filtered out.
- campaign_statuses (array): Select the statuses you want to be loaded in the stream. If no specific statuses are selected, the API's default behavior applies, and some statuses may be filtered out.
- client_id (string) secret: The Client Id for your OAuth app
- client_secret (string) secret: The Client Secret for your OAuth app
- credentials (object) required: Credentials for connecting to the Facebook Marketing API
- custom_insights (array): A list which contains ad statistics entries, each entry must have a name and can contains fields, breakdowns or action_breakdowns. Click on "add" to fill this field.
- default_ads_insights_action_breakdowns (array): Action breakdowns for the Built-in Ads Insights stream that will be used in the request. You can override default values or remove them to make it empty if needed.
- end_date (string): The date until which you'd like to replicate data for all incremental streams, in the format YYYY-MM-DDT00:00:00Z. All data generated between the start date and this end date wi...
- fetch_thumbnail_images (boolean): Set to active if you want to fetch the thumbnail_url and store the result in thumbnail_data_url for each Ad Creative.
- include_incrementality (boolean): If enabled, the incrementality attribution window will be included in the action attribution windows for all built-in insight streams. This allows you to retrieve incrementality...
- insights_job_timeout (integer): Insights Job Timeout establishes the maximum amount of time (in minutes) of waiting for the report job to complete. When timeout is reached the job is considered failed and we a...
- insights_lookback_window (integer): The attribution window. Facebook freezes insight data 28 days after it was generated, which means that all data from the past 28 days may have changed since we last emitted it, ...
- page_size (integer): Page size used when sending requests to Facebook API to specify number of records per page when response has pagination. Most users do not need to set this field unless they spe...
- start_date (string): The date from which you'd like to replicate data for all incremental streams, in the format YYYY-MM-DDT00:00:00Z. If not set then all data will be replicated for usual streams a...
- secret fields: access_token, client_id, client_secret, credentials.access_token, credentials.client_id, credentials.client_secret

## Sync Modes

- supported sync modes: full_refresh, incremental
- supports incremental: true

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/facebook-marketing

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-facebook-marketing
```

### Inspect as JSON

```bash
pm connectors inspect source-facebook-marketing --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Facebook Marketing documentation](https://docs.airbyte.com/integrations/sources/facebook-marketing)
