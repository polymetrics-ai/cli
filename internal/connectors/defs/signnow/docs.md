# Overview

Reads signNow documents, templates, and users through the signNow REST API.

Readable streams: `documents`, `templates`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.signnow.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); signNow OAuth access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `base_url` (optional, string); default `https://api.signnow.com`; format `uri`; signNow API base
  URL override for tests or proxies.
- `page_size` (optional, string); default `50`; Records per page (1-500), sent as the limit query
  parameter on every request.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.signnow.com`, `page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/document` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page_token`; next token from `next`.

- `documents`: GET `/document` - records path `data`; query `limit`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page_token`; next token from `next`; computed output fields `name`,
  `updated_at`.
- `templates`: GET `/template` - records path `data`; query `limit`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page_token`; next token from `next`; computed output fields `name`,
  `updated_at`.
- `users`: GET `/user` - records path `data`; query `limit`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `page_token`; next token from `next`; computed output fields
  `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external signNow API read of document, template, and
user data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
