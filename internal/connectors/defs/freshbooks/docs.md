# Overview

Reads FreshBooks clients, invoices, expenses, payments, and items through the FreshBooks accounting
REST API.

Readable streams: `clients`, `invoices`, `expenses`, `payments`, `items`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.freshbooks.com/api/start.

## Auth setup

Connection fields:

- `account_id` (required, string); FreshBooks account id; every accounting list endpoint is scoped
  under /accounting/account/{account_id}/.
- `base_url` (optional, string); default `https://api.freshbooks.com`; format `uri`; FreshBooks API
  base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `oauth_access_token` (required, secret, string); FreshBooks OAuth2 access token. Sent as
  Authorization: Bearer <oauth_access_token>; never logged. The refresh token / client id / secret
  used for the OAuth dance itself are not consumed directly by this connector.
- `page_size` (optional, string); default `100`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `oauth_access_token`.

Default configuration values: `base_url=https://api.freshbooks.com`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.oauth_access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounting/account/{{ config.account_id }}/users/clients` with query
`per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

- `clients`: GET `/accounting/account/{{ config.account_id }}/users/clients` - records path
  `response.result.clients`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100.
- `invoices`: GET `/accounting/account/{{ config.account_id }}/invoices/invoices` - records path
  `response.result.invoices`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100.
- `expenses`: GET `/accounting/account/{{ config.account_id }}/expenses/expenses` - records path
  `response.result.expenses`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100.
- `payments`: GET `/accounting/account/{{ config.account_id }}/payments/payments` - records path
  `response.result.payments`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100.
- `items`: GET `/accounting/account/{{ config.account_id }}/items/items` - records path
  `response.result.items`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external FreshBooks API read of accounting data
(clients, invoices, expenses, payments, items).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
