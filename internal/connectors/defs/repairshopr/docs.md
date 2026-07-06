# Overview

Reads RepairShopr customers, tickets, invoices, estimates, and assets through the REST API.

Readable streams: `customers`, `tickets`, `invoices`, `estimates`, `assets`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api-docs.repairshopr.com/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); RepairShopr API token, sent as a Bearer token
  (Authorization: Bearer <api_token>). Never logged.
- `base_url` (required, string); format `uri`; RepairShopr API base URL, e.g.
  https://<subdomain>.repairshopr.com/api/v1.
- `created_after` (optional, string); Optional passthrough filter: only return records created at or
  after this value.
- `query` (optional, string); Optional free-text search query passed through to RepairShopr's list
  endpoints.
- `updated_after` (optional, string); Optional passthrough filter: only return records updated at or
  after this value.

Secret fields are redacted in logs and write previews: `api_token`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers` with query `page`=`1`; `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

- `customers`: GET `/customers` - records path `customers`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted
  when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `stream`; emits passthrough records.
- `tickets`: GET `/tickets` - records path `tickets`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted
  when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `stream`; emits passthrough records.
- `invoices`: GET `/invoices` - records path `invoices`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted
  when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `stream`; emits passthrough records.
- `estimates`: GET `/estimates` - records path `estimates`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted
  when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `stream`; emits passthrough records.
- `assets`: GET `/customer_assets` - records path `assets`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `query` from template `{{ config.query }}`, omitted
  when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `stream`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external RepairShopr API read of customer and
shop-management data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
