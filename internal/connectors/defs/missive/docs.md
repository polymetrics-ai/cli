# Overview

Reads Missive contacts, contact groups, users, teams, and shared labels through the Missive REST
API.

Readable streams: `contacts`, `contact_groups`, `users`, `teams`, `shared_labels`.

This connector is read-only; no write actions are declared.

Service API documentation: https://missiveapp.com/help/api-documentation.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Missive API token. Sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://public.missiveapp.com/v1`; format `uri`; Missive
  API base URL override for tests or proxies.
- `kind` (optional, string); Optional contact_groups filter: 'group' or 'organization'. When unset,
  the contact_groups stream returns both kinds (Missive's own default).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://public.missiveapp.com/v1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 50.

- `contacts`: GET `/contacts` - records path `contacts`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50.
- `contact_groups`: GET `/contact_groups` - records path `contact_groups`; query `kind` from
  template `{{ config.kind }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50.
- `users`: GET `/users` - records path `users`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 50.
- `teams`: GET `/teams` - records path `teams`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 50.
- `shared_labels`: GET `/shared_labels` - records path `shared_labels`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 50.

## Write actions & risks

This connector is read-only. Read behavior: external Missive API read of contact, user, team, and
label data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
