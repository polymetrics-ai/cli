# Overview

Reads Plaid institutions and category metadata through read-only POST endpoints.

Readable streams: `institutions`, `categories`.

This connector is read-only; no write actions are declared.

Service API documentation: https://plaid.com/docs/api/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://production.plaid.com`; format `uri`; Plaid API
  base URL override (e.g. https://sandbox.plaid.com or https://development.plaid.com for
  non-production environments).
- `client_id` (required, secret, string); Plaid client_id. Sent in the JSON body of every request
  (Plaid's own convention, never a header); never logged.
- `country_codes` (optional, string); default `US`; Comma-separated ISO-3166-1 alpha-2 country codes
  for the institutions.get request body (e.g. US,CA).
- `max_pages` (optional, string); default `3`; Hard cap on pages read per sync; 0, "all", or
  "unlimited" means unbounded.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-500), sent as the
  institutions.get request body's count field.
- `secret` (required, secret, string); Plaid secret. Sent in the JSON body of every request; never
  logged.

Secret fields are redacted in logs and write previews: `client_id`, `secret`.

Default configuration values: `base_url=https://production.plaid.com`, `country_codes=US`,
`max_pages=3`, `page_size=100`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call POST `categories/get`.

## Streams notes

Default pagination: single request; no pagination.

- `institutions`: POST `institutions/get` - records path `institutions`; computed output fields
  `country_codes`.
- `categories`: POST `categories/get` - records path `categories`; computed output fields
  `hierarchy`.

## Write actions & risks

This connector is read-only. Read behavior: external Plaid API read of institution/category
metadata.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
