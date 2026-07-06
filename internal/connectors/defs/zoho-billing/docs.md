# Overview

Reads Zoho Billing customers, subscriptions, and invoices through the Zoho Billing REST API.

Readable streams: `customers`, `subscriptions`, `invoices`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.zoho.com/billing/api/v1/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Zoho OAuth access token. Sent as 'Authorization:
  Zoho-oauthtoken <access_token>'; never logged.
- `base_url` (optional, string); default `https://www.zohoapis.com/billing/v1`; format `uri`; Zoho
  Billing API base URL override for tests or proxies.
- `organization_id` (optional, string); Optional Zoho Billing organization ID; sent as the
  organization_id query parameter on every request.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://www.zohoapis.com/billing/v1`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Zoho-oauthtoken` using
  `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 200.

- `customers`: GET `/customers` - records path `customers`; query `organization_id` from template
  `{{ config.organization_id }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`.
- `subscriptions`: GET `/subscriptions` - records path `subscriptions`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`.
- `invoices`: GET `/invoices` - records path `invoices`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Zoho Billing API read of customer and billing
data.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=7.
