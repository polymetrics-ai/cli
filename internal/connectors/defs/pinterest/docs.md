# Overview

Reads Pinterest ad accounts, boards, campaigns, ad groups, and audiences through the Pinterest API
v5 (OAuth2 refresh-token auth). Read-only.

Readable streams: `ad_accounts`, `boards`, `campaigns`, `ad_groups`, `audiences`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.pinterest.com/docs/api/v5/.

## Auth setup

Connection fields:

- `account_id` (optional, string); Pinterest ad account ID. Required only for the account-scoped
  streams (campaigns, ad_groups, audiences), whose resource path substitutes it; not required for
  ad_accounts/boards.
- `base_url` (optional, string); default `https://api.pinterest.com/v5`; format `uri`; Pinterest API
  v5 base URL override for tests or proxies.
- `client_id` (required, secret, string); Pinterest OAuth 2.0 client ID for the refresh-token grant.
  Sent only as HTTP Basic auth on the token request; never logged.
- `client_secret` (required, secret, string); Pinterest OAuth 2.0 client secret. Sent only as HTTP
  Basic auth on the token request; never logged.
- `mode` (optional, string).
- `page_size` (optional, string); Records per page (1-250).
- `refresh_token` (required, secret, string); Long-lived Pinterest OAuth 2.0 refresh token.
  Exchanged for a short-lived access token at token_url; never logged. The 3-legged
  consent/acquisition dance is out of scope for this connector (credentials layer already owns it).
- `token_url` (optional, string); default `https://api.pinterest.com/v5/oauth/token`; format `uri`;
  Pinterest OAuth 2.0 token endpoint override. MUST be http(s) with a host; the hook fails closed on
  an invalid value to prevent exfiltrating the refresh token/client secret to an attacker-chosen
  endpoint.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`,
`refresh_token`.

Default configuration values: `base_url=https://api.pinterest.com/v5`,
`token_url=https://api.pinterest.com/v5/oauth/token`.

Authentication behavior:

- Connector-specific authentication using `secrets.refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/ad_accounts` with query `page_size`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `bookmark`; next token from `bookmark`.

- `ad_accounts`: GET `/ad_accounts` - records path `items`; query `page_size` from template `{{
  config.page_size }}`, omitted when absent; cursor pagination; cursor parameter `bookmark`; next
  token from `bookmark`.
- `boards`: GET `/boards` - records path `items`; query `page_size` from template `{{
  config.page_size }}`, omitted when absent; cursor pagination; cursor parameter `bookmark`; next
  token from `bookmark`.
- `campaigns`: GET `/ad_accounts/{{ config.account_id }}/campaigns` - records path `items`; query
  `page_size` from template `{{ config.page_size }}`, omitted when absent; cursor pagination; cursor
  parameter `bookmark`; next token from `bookmark`.
- `ad_groups`: GET `/ad_accounts/{{ config.account_id }}/ad_groups` - records path `items`; query
  `page_size` from template `{{ config.page_size }}`, omitted when absent; cursor pagination; cursor
  parameter `bookmark`; next token from `bookmark`.
- `audiences`: GET `/ad_accounts/{{ config.account_id }}/audiences` - records path `items`; query
  `page_size` from template `{{ config.page_size }}`, omitted when absent; cursor pagination; cursor
  parameter `bookmark`; next token from `bookmark`.

## Write actions & risks

This connector is read-only. Read behavior: external Pinterest API read of ad account, board,
campaign, ad group, and audience data.

## Known limits

- Batch defaults: read_page_size=25.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=6.
