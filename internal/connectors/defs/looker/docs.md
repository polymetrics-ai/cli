# Overview

Reads Looker users, groups, folders, looks, and dashboards through the Looker API 4.0.

Readable streams: `users`, `groups`, `folders`, `looks`, `dashboards`.

This connector is read-only; no write actions are declared.

Service API documentation: https://cloud.google.com/looker/docs/reference/looker-api/latest.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Looker API access token. When set, used directly as a
  Bearer token; takes priority over client_id/client_secret when both are configured.
- `base_url` (required, string); format `uri`; Looker API base URL, including the API version path
  (for example https://company.looker.com/api/4.0).
- `client_id` (optional, secret, string); Looker API3 client ID. Used with client_secret to log in
  and obtain a Bearer token when access_token is not set.
- `client_secret` (optional, secret, string); Looker API3 client secret. Used with client_id to log
  in and obtain a Bearer token when access_token is not set.
- `mode` (optional, string).
- `token_url` (optional, string); format `uri`; Looker login endpoint override for tests or proxies.
  Defaults to <base_url>/login.

Secret fields are redacted in logs and write previews: `access_token`, `client_id`, `client_secret`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret` when `{{ config.token_url }}`.
- OAuth 2.0 client credentials authentication using `config.base_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `users`: GET `/users` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `groups`: GET `/groups` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `folders`: GET `/folders` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `looks`: GET `/looks` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `dashboards`: GET `/dashboards` - records at response root; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Looker API read of users, groups, folders,
looks, and dashboards.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, non_data_endpoint=1, out_of_scope=8.
