# Overview

Reads Microsoft Teams users, groups, channels, and device-usage reports through the Microsoft Graph
REST API using an OAuth2 client-credentials grant. Read-only.

Readable streams: `users`, `groups`, `channels`, `team_device_usage_report`.

This connector is read-only; no write actions are declared.

Service API documentation: https://learn.microsoft.com/en-us/graph/api/overview.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://graph.microsoft.com/v1.0`; format `uri`; Microsoft
  Graph REST API base URL override for tests or proxies. Defaults to
  https://graph.microsoft.com/v1.0.
- `client_id` (required, string); Microsoft Entra ID (Azure AD) application (client) ID used for the
  OAuth2 client-credentials grant.
- `client_secret` (optional, secret, string); Microsoft Entra ID application client secret, sent
  only in the OAuth2 client-credentials token-exchange request; never logged.
- `login_base_url` (optional, string); default `https://login.microsoftonline.com`; format `uri`.
- `max_pages` (optional, string); Permissive parse: empty, "all", "unlimited", or any
  non-positive-integer string means unbounded.
- `period` (optional, string); default `D7`; Teams device usage report aggregation period: D7, D30,
  D90, or D180.
- `scope` (optional, string); default `https://graph.microsoft.com/.default`; OAuth2
  client-credentials scope. Defaults to the static Graph application-permission scope.
- `tenant_id` (optional, secret, string); Microsoft Entra ID tenant ID (GUID or verified domain),
  used to derive the per-tenant token endpoint.
- `token_url` (optional, string); format `uri`; Full OAuth2 token endpoint override. When set, takes
  priority over the derived login_base_url/tenant_id endpoint (dual-candidate auth pattern; see
  docs.md Known limits and the sharepoint-lists-enterprise golden).

Secret fields are redacted in logs and write previews: `client_secret`, `tenant_id`.

Default configuration values: `base_url=https://graph.microsoft.com/v1.0`,
`login_base_url=https://login.microsoftonline.com`, `period=D7`,
`scope=https://graph.microsoft.com/.default`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `config.client_id`,
  `secrets.client_secret`, `config.scope` when `{{ config.token_url }}`.
- OAuth 2.0 client credentials authentication using `config.login_base_url`, `secrets.tenant_id`,
  `config.client_id`, `secrets.client_secret`, `config.scope`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/organization`.

## Streams notes

Default pagination: single request; no pagination.

- `users`: GET `/users` - records path `value`.
- `groups`: GET `/groups` - records path `value`.
- `channels`: GET `/teams/getAllChannels` - records path `value`.
- `team_device_usage_report`: GET `/reports/getTeamsDeviceUsageUserDetail` - records path `value`;
  query `$format`=`application/json`; `period` from template `{{ config.period }}`, default `D7`.

## Write actions & risks

This connector is read-only. Read behavior: external Microsoft Graph API read of tenant
users/groups/channels/device-usage data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=2, requires_elevated_scope=1.
