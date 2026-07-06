# Overview

Reads Lob addresses, postcards, letters, checks, and bank accounts through the Lob print & mail REST
API.

Readable streams: `addresses`, `postcards`, `letters`, `checks`, `bank_accounts`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.lob.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Lob API key, sent as the HTTP Basic auth username with a
  blank password. Never logged.
- `base_url` (optional, string); default `https://api.lob.com/v1`; format `uri`; Lob API base URL
  override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.lob.com/v1`, `max_pages=0`, `page_size=50`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/addresses` with query `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next_url`; next URLs
stay on the configured API host.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `addresses`: GET `/addresses` - records path `data`; query `limit`=`{{ config.page_size }}`;
  follows a next-page URL from the response body; URL path `next_url`; next URLs stay on the
  configured API host; incremental cursor `date_created`; formatted as `rfc3339`.
- `postcards`: GET `/postcards` - records path `data`; query `limit`=`{{ config.page_size }}`;
  follows a next-page URL from the response body; URL path `next_url`; next URLs stay on the
  configured API host; incremental cursor `date_created`; formatted as `rfc3339`.
- `letters`: GET `/letters` - records path `data`; query `limit`=`{{ config.page_size }}`; follows a
  next-page URL from the response body; URL path `next_url`; next URLs stay on the configured API
  host; incremental cursor `date_created`; formatted as `rfc3339`.
- `checks`: GET `/checks` - records path `data`; query `limit`=`{{ config.page_size }}`; follows a
  next-page URL from the response body; URL path `next_url`; next URLs stay on the configured API
  host; incremental cursor `date_created`; formatted as `rfc3339`.
- `bank_accounts`: GET `/bank_accounts` - records path `data`; query `limit`=`{{ config.page_size
  }}`; follows a next-page URL from the response body; URL path `next_url`; next URLs stay on the
  configured API host; incremental cursor `date_created`; formatted as `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external Lob API read of address book, mailpiece, and
bank account data.

## Known limits

- Published rate limit metadata: requests_per_minute=150.
- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=7.
