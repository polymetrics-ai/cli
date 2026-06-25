---
name: pm-source-shopify
description: Shopify connector knowledge and safe action guide.
---

# pm-source-shopify

## Purpose

Shopify catalog connector for https://docs.airbyte.com/integrations/sources/shopify. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: generally_available
- support level: certified

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: native_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-shopify:3.5.1 (metadata only; not executed)

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

- Shopify Admin API: https://shopify.dev/docs/api/admin-rest
- Shopify authentication: https://shopify.dev/docs/apps/auth
- Developer changelog: https://shopify.dev/changelog
- Shopify API changelog: https://shopify.dev/docs/api/release-notes
- Shopify API versioning and deprecation policy: https://shopify.dev/docs/api/usage/versioning
- Shopify rate limits: https://shopify.dev/docs/api/usage/rate-limits
- Shopify Status: https://www.shopifystatus.com/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/shopify

## Configuration

- bulk_window_in_days (integer): Defines what would be a date range per single BULK Job
- credentials (object): The authorization method to use to retrieve data from Shopify
- fetch_transactions_user_id (boolean): Defines which API type (REST/BULK) to use to fetch `Transactions` data. If you are a `Shopify Plus` user, leave the default value to speed up the fetch.
- fulfillment_orders_include_closed (boolean): If enabled, the `Fulfillment Orders` stream includes closed fulfillment orders. Shopify excludes closed orders by default.
- job_checkpoint_interval (integer): The threshold, after which the single BULK Job should be checkpointed (min: 15k, max: 1M)
- job_product_variants_include_pres_prices (boolean): If enabled, the `Product Variants` stream attempts to include `Presentment prices` field (may affect the performance).
- job_termination_threshold (integer): The max time in seconds, after which the single BULK Job should be `CANCELED` and retried. The bigger the value the longer the BULK Job is allowed to run.
- lookback_window_in_days (integer): If set to a positive number, during each incremental sync the connector will re-fetch records from the past N days before the saved state. This helps capture records that may ha...
- shop (string) required: The name of your Shopify store found in the URL. For example, if your URL is https://NAME.myshopify.com, then the name is 'NAME'. You may also paste the full myshopify URL (e.g....
- start_date (string): The date you would like to replicate data from. Format: YYYY-MM-DD. Any data before this date will not be replicated.
- secret fields: credentials.access_token, credentials.api_password, credentials.client_id, credentials.client_secret

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/shopify

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-shopify
```

### Inspect as JSON

```bash
pm connectors inspect source-shopify --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Shopify documentation](https://docs.airbyte.com/integrations/sources/shopify)
