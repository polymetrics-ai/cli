# Overview

Reads Salesloft people, accounts, cadences, users, and emails through the Salesloft REST API v2.

Readable streams: `people`, `accounts`, `cadences`, `users`, `emails`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.salesloft.com/docs/api.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); A Salesloft OAuth2 access token. Used directly as a
  Bearer token when no api_key is set and no refresh_token+client_id+client_secret triple is
  configured.
- `api_key` (optional, secret, string); Salesloft API key, sent as a Bearer token.
- `base_url` (optional, string); default `https://api.salesloft.com/v2`; format `uri`; Salesloft
  REST API v2 base URL override for tests or proxies.
- `client_id` (optional, secret, string); Salesloft OAuth2 client ID, used only in the
  refresh-token-grant token request form.
- `client_secret` (optional, secret, string); Salesloft OAuth2 client secret, used only in the
  refresh-token-grant token request form.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `refresh_token` (optional, secret, string); Long-lived Salesloft OAuth2 refresh token. Exchanged
  for a short-lived access token at token_url (grant_type=refresh_token) when
  client_id/client_secret are also configured and no api_key is set.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects updated at
  or after this time are read, on a fresh sync with no persisted cursor.
- `token_url` (optional, string); default `https://accounts.salesloft.com/oauth/token`; format
  `uri`; Salesloft OAuth2 token endpoint override. MUST be http(s) with a host; the hook fails
  closed on an invalid value to prevent exfiltrating the refresh token/client secret to an
  attacker-chosen endpoint.

Secret fields are redacted in logs and write previews: `access_token`, `api_key`, `client_id`,
`client_secret`, `refresh_token`.

Default configuration values: `base_url=https://api.salesloft.com/v2`, `max_pages=0`,
`page_size=100`, `token_url=https://accounts.salesloft.com/oauth/token`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key` when `{{ secrets.api_key }}`.
- Connector-specific authentication using `secrets.refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret` when `{{ secrets.refresh_token }}`.
- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `per_page`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from
`metadata.paging.next_page`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `people`: GET `/people` - records path `data`; query `per_page`=`{{ config.page_size }}`;
  `sort_by` from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent;
  `sort_direction` from template `{{ incremental.lower_bound | const:ASC }}`, omitted when absent;
  cursor pagination; cursor parameter `page`; next token from `metadata.paging.next_page`;
  incremental cursor `updated_at`; sent as `updated_at[gte]`; formatted as `rfc3339`; initial lower
  bound from `start_date`; computed output fields `account_id`, `owner_id`.
- `accounts`: GET `/accounts` - records path `data`; query `per_page`=`{{ config.page_size }}`;
  `sort_by` from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent;
  `sort_direction` from template `{{ incremental.lower_bound | const:ASC }}`, omitted when absent;
  cursor pagination; cursor parameter `page`; next token from `metadata.paging.next_page`;
  incremental cursor `updated_at`; sent as `updated_at[gte]`; formatted as `rfc3339`; initial lower
  bound from `start_date`; computed output fields `owner_id`.
- `cadences`: GET `/cadences` - records path `data`; query `per_page`=`{{ config.page_size }}`;
  `sort_by` from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent;
  `sort_direction` from template `{{ incremental.lower_bound | const:ASC }}`, omitted when absent;
  cursor pagination; cursor parameter `page`; next token from `metadata.paging.next_page`;
  incremental cursor `updated_at`; sent as `updated_at[gte]`; formatted as `rfc3339`; initial lower
  bound from `start_date`.
- `users`: GET `/users` - records path `data`; query `per_page`=`{{ config.page_size }}`; `sort_by`
  from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent;
  `sort_direction` from template `{{ incremental.lower_bound | const:ASC }}`, omitted when absent;
  cursor pagination; cursor parameter `page`; next token from `metadata.paging.next_page`;
  incremental cursor `updated_at`; sent as `updated_at[gte]`; formatted as `rfc3339`; initial lower
  bound from `start_date`.
- `emails`: GET `/emails` - records path `data`; query `per_page`=`{{ config.page_size }}`;
  `sort_by` from template `{{ incremental.lower_bound | const:updated_at }}`, omitted when absent;
  `sort_direction` from template `{{ incremental.lower_bound | const:ASC }}`, omitted when absent;
  cursor pagination; cursor parameter `page`; next token from `metadata.paging.next_page`;
  incremental cursor `updated_at`; sent as `updated_at[gte]`; formatted as `rfc3339`; initial lower
  bound from `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external Salesloft API read of people, accounts,
cadences, users, and email data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=6.
