# Overview

Reads Klaus (Zendesk QA) users and rating categories through the Klaus public REST API.

Readable streams: `users`, `categories`.

This connector is read-only; no write actions are declared.

Service API documentation: https://help.klausapp.com/en/articles/2911907-klaus-api.

## Auth setup

Connection fields:

- `account` (required, string); Klaus account id (integer). Required for every stream;
  account-scoped paths are /account/{account}/<resource>.
- `api_key` (required, secret, string); Klaus API key. Used only for Bearer auth; never logged.
- `base_url` (optional, string); default `https://kibbles.klausapp.com/api/v2`; format `uri`; Klaus
  API base URL override for tests or proxies.
- `mode` (optional, string).
- `workspace` (optional, string); Klaus workspace id (integer). Required only for the categories
  stream, which is workspace-scoped (/account/{account}/workspace/{workspace}/categories).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://kibbles.klausapp.com/api/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/account/{{ config.account }}/users`.

## Streams notes

Default pagination: single request; no pagination.

- `users`: GET `/account/{{ config.account }}/users` - records path `users`.
- `categories`: GET `/account/{{ config.account }}/workspace/{{ config.workspace }}/categories` -
  records path `categories`.

## Write actions & risks

This connector is read-only. Read behavior: external Klaus API read of user and quality-review
configuration data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
