# Overview

Reads Taboola campaigns through the Backstage API. Read-only.

Readable streams: `campaigns`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.taboola.com/backstage-api/reference.

## Auth setup

Connection fields:

- `account_id` (required, string); Taboola account ID; sent as a path segment on the campaigns
  endpoint.
- `base_url` (optional, string); default `https://backstage.taboola.com`; format `uri`; Taboola
  Backstage API base URL override for tests or proxies. Also used to derive the OAuth2 token
  endpoint (<base_url>/backstage/oauth/token).
- `client_id` (required, secret, string); Taboola Backstage OAuth2 client ID. Never logged.
- `client_secret` (required, secret, string); Taboola Backstage OAuth2 client secret. Never logged.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://backstage.taboola.com`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.base_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/backstage/api/1.0/{{ config.account_id }}/campaigns` with query
`page`=`1`; `page_size`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100.

- `campaigns`: GET `/backstage/api/1.0/{{ config.account_id }}/campaigns` - records path `results`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100.

## Write actions & risks

This connector is read-only. Read behavior: external Taboola Backstage API read of campaign data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 1 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
