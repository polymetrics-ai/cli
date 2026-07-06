# Overview

Reads SendGrid Marketing Campaigns lists, segments, and contacts, plus suppression bounces, through
the SendGrid v3 REST API. Read-only.

Readable streams: `lists`, `segments`, `contacts`, `suppression_bounces`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://docs.sendgrid.com/api-reference/how-to-use-the-sendgrid-v3-api/authentication.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); SendGrid API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.sendgrid.com/v3`; format `uri`; SendGrid API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.sendgrid.com/v3`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/marketing/lists` with query `page_size`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `_metadata.next`; next
URLs stay on the configured API host.

Pagination by stream: next_url: `lists`, `segments`, `contacts`; offset_limit:
`suppression_bounces`.

- `lists`: GET `/marketing/lists` - records path `result`; query `page_size`=`100`; follows a
  next-page URL from the response body; URL path `_metadata.next`; next URLs stay on the configured
  API host.
- `segments`: GET `/marketing/segments/2.0` - records path `results`; query `page_size`=`100`;
  follows a next-page URL from the response body; URL path `_metadata.next`; next URLs stay on the
  configured API host.
- `contacts`: GET `/marketing/contacts` - records path `result`; query `page_size`=`100`; follows a
  next-page URL from the response body; URL path `_metadata.next`; next URLs stay on the configured
  API host.
- `suppression_bounces`: GET `/suppression/bounces` - records path `.`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external SendGrid API read of marketing list, segment,
contact, and suppression-bounce data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=4.
