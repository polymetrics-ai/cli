# Overview

Reads Recurly accounts, subscriptions, invoices, transactions, and plans through the Recurly v3 REST
API.

Readable streams: `accounts`, `subscriptions`, `invoices`, `transactions`, `plans`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.recurly.com/api/v2021-02-25/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Recurly private API key, sent as the HTTP Basic username
  with an empty password (Authorization: Basic base64(api_key:)). Never logged.
- `base_url` (optional, string); default `https://v3.recurly.com`; format `uri`; Recurly API base
  URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://v3.recurly.com`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts` with query `limit`=`1`.

## Streams notes

Default pagination: follows RFC 5988 Link headers with rel=next.

- `accounts`: GET `/accounts` - records path `data`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; computed output fields `code`, `created_at`, `email`, `id`, `state`,
  `updated_at`.
- `subscriptions`: GET `/subscriptions` - records path `data`; query `limit`=`200`; follows RFC 5988
  Link headers with rel=next; computed output fields `account_id`, `created_at`, `id`, `plan_id`,
  `state`, `updated_at`.
- `invoices`: GET `/invoices` - records path `data`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; computed output fields `account_id`, `created_at`, `id`, `state`, `total`.
- `transactions`: GET `/transactions` - records path `data`; query `limit`=`200`; follows RFC 5988
  Link headers with rel=next; computed output fields `account_id`, `amount`, `created_at`, `id`,
  `status`.
- `plans`: GET `/plans` - records path `data`; query `limit`=`200`; follows RFC 5988 Link headers
  with rel=next; computed output fields `code`, `id`, `name`, `state`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Recurly API read of subscription billing data.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, non_data_endpoint=1, out_of_scope=5.
