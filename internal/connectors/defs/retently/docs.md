# Overview

Reads Retently customers, survey responses, surveys, and campaigns through the REST API.

Readable streams: `customers`, `responses`, `surveys`, `campaigns`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.retently.com/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Never logged.
- `base_url` (optional, string); default `https://app.retently.com/api/v2`; format `uri`; Retently
  API base URL override for tests or proxies.
- `campaign_id` (optional, string); Optional passthrough filter: scope the responses stream to a
  single campaign id.
- `created_after` (optional, string); Optional passthrough filter: only return responses created at
  or after this value.
- `email` (optional, string); Optional passthrough filter: scope the customers stream to a single
  customer email.
- `updated_after` (optional, string); Optional passthrough filter: only return
  customers/surveys/campaigns updated at or after this value.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.retently.com/api/v2`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `api_key=` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

- `customers`: GET `/customers` - records path `data`; query `campaign_id` from template `{{
  config.campaign_id }}`, omitted when absent; `created_after` from template `{{
  config.created_after }}`, omitted when absent; `email` from template `{{ config.email }}`, omitted
  when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `stream`.
- `responses`: GET `/responses` - records path `data`; query `campaign_id` from template `{{
  config.campaign_id }}`, omitted when absent; `created_after` from template `{{
  config.created_after }}`, omitted when absent; `email` from template `{{ config.email }}`, omitted
  when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `stream`.
- `surveys`: GET `/surveys` - records path `data`; query `campaign_id` from template `{{
  config.campaign_id }}`, omitted when absent; `created_after` from template `{{
  config.created_after }}`, omitted when absent; `email` from template `{{ config.email }}`, omitted
  when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `stream`.
- `campaigns`: GET `/campaigns` - records path `data`; query `campaign_id` from template `{{
  config.campaign_id }}`, omitted when absent; `created_after` from template `{{
  config.created_after }}`, omitted when absent; `email` from template `{{ config.email }}`, omitted
  when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `stream`.

## Write actions & risks

This connector is read-only. Read behavior: external Retently API read of customer and NPS/CSAT
survey response data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, out_of_scope=6.
