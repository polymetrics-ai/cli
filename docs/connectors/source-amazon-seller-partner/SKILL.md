---
name: pm-source-amazon-seller-partner
description: Amazon Seller Partner connector knowledge and safe action guide.
---

# pm-source-amazon-seller-partner

## Purpose

Amazon Seller Partner catalog connector for https://docs.airbyte.com/integrations/sources/amazon-seller-partner. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-amazon-seller-partner:5.7.10 (metadata only; not executed)

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

- SP-API documentation: https://developer-docs.amazon.com/sp-api/
- SP-API Deprecation Schedule: https://developer-docs.amazon.com/sp-api/docs/sp-api-deprecations
- SP-API Release Notes: https://developer-docs.amazon.com/sp-api/docs/sp-api-release-notes
- Authorization: https://developer-docs.amazon.com/sp-api/docs/authorizing-selling-partner-api-applications
- Usage plans and rate limits: https://developer-docs.amazon.com/sp-api/docs/usage-plans-and-rate-limits-in-the-sp-api
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/amazon-seller-partner

## Configuration

- account_type (string) required: Type of the Account you're going to authorize the Airbyte application by
- app_id (string) secret: Your Amazon Application ID.
- auth_type (string)
- aws_environment (string) required: Select the AWS Environment.
- creation_requester_429_max_retries (integer): Maximum number of retry attempts when the report creation API returns HTTP 429 (Too Many Requests). Each retry uses exponential backoff based on the x-amzn-RateLimit-Limit heade...
- failed_retry_wait_time_in_seconds (integer): Time in seconds to wait before retrying a report that returned FATAL status. Amazon enforces per-report-type cooldowns after generating a report. Near-real-time FBA reports have...
- financial_events_max_results_per_page (integer): The maximum number of results to return per page for the ListFinancialEvents stream. If the response exceeds the maximum number of transactions or 10 MB, the API returns an Inva...
- financial_events_step (string): The time window size for fetching financial events data in chunks for the ListFinancialEvents and ListFinancialEventGroups streams. Options include hourly (1H, 6H, 12H) and dail...
- include_pii (boolean): When enabled, the connector requests a Restricted Data Token (RDT) to access PII fields such as BuyerInfo and ShippingAddress in the Orders and OrderItems streams. Your Amazon S...
- lwa_app_id (string) required secret: Your Login with Amazon Client ID.
- lwa_client_secret (string) required secret: Your Login with Amazon Client Secret.
- max_async_job_count (integer): The maximum number of concurrent asynchronous job requests that can be active at a time.
- max_done_report_age_hours (integer): When the connector finds an existing completed (DONE) report matching the same date range and marketplace, it can reuse that report instead of creating a new one. This setting c...
- num_workers (integer): The number of workers to use for the connector when syncing concurrently.
- period_in_days (integer): For syncs spanning a large date range, this option is used to request data in a smaller fixed window to improve sync reliability. This time window can be configured granularly b...
- refresh_token (string) required secret: The Refresh Token obtained via OAuth flow authorization.
- region (string) required: Select the AWS Region.
- replication_end_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Any data after this date will not be replicated.
- replication_start_date (string): UTC date and time in the format 2017-01-25T00:00:00Z. Any data before this date will not be replicated. If start date is not provided or older than 2 years ago from today, the d...
- report_options_list (array): Additional information passed to reports. This varies by report type.
- report_stream_lookback_window_in_hours (integer): Number of hours to re-fetch for incremental report streams on each sync. Increase this value if Amazon continues updating report data after the previous sync has completed. Dupl...
- sales_and_traffic_report_asin_granularity (string): The level of ASIN granularity for the Sales and Traffic report streams. PARENT returns data aggregated at the parent ASIN level. CHILD returns data at the child ASIN level with ...
- stop_sync_on_rate_limit (boolean): Only applies to report streams. When enabled, the source stops retrying immediately once the rate limit retry budget is exhausted and fails with an actionable configuration erro...
- wait_to_avoid_fatal_errors (boolean): Deprecated - this option is no longer functional and will be removed in a future version. Rate limiting is now handled automatically by the connector.
- secret fields: app_id, lwa_app_id, lwa_client_secret, refresh_token

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/amazon-seller-partner

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-amazon-seller-partner
```

### Inspect as JSON

```bash
pm connectors inspect source-amazon-seller-partner --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Amazon Seller Partner documentation](https://docs.airbyte.com/integrations/sources/amazon-seller-partner)
