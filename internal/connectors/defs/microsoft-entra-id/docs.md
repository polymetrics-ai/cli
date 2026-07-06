# Overview

Reads Microsoft Entra ID (Azure AD) directory objects - users, groups, applications, service
principals, and directory roles - from the Microsoft Graph API using an OAuth2 client-credentials
grant. Read-only.

Readable streams: `users`, `groups`, `applications`, `serviceprincipals`, `directoryroles`.

This connector is read-only; no write actions are declared.

Service API documentation: https://learn.microsoft.com/en-us/graph/api/overview.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://graph.microsoft.com/v1.0`; format `uri`; Microsoft
  Graph REST API base URL override for tests or proxies. Defaults to
  https://graph.microsoft.com/v1.0.
- `client_id` (optional, secret, string); Microsoft Entra ID (Azure AD) application (client) ID used
  for the OAuth2 client-credentials grant. Never logged.
- `client_secret` (optional, secret, string); Microsoft Entra ID application client secret, sent
  only in the OAuth2 client-credentials token-exchange request; never logged.
- `login_base_url` (optional, string); default `https://login.microsoftonline.com`; format `uri`;
  Azure AD login base URL override for tests or proxies; combined with tenant_id to derive the
  OAuth2 token endpoint.
- `max_pages` (optional, string); Permissive parse: empty, "all", or "unlimited" means unbounded.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records requested per @odata.nextLink page (1-999),
  sent as the $top query parameter on the first request of each stream.
- `scope` (optional, string); default `https://graph.microsoft.com/.default`; OAuth2
  client-credentials scope. Defaults to the static Graph application-permission scope.
- `tenant_id` (optional, secret, string); Microsoft Entra ID tenant ID (GUID or verified domain),
  used to derive the per-tenant token endpoint.
- `token_url` (optional, string); format `uri`; Full OAuth2 token endpoint override. When set, takes
  priority over the derived login_base_url/tenant_id endpoint (see docs.md Known limits for the
  dual-candidate auth mechanism used to express this).

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`, `tenant_id`.

Default configuration values: `base_url=https://graph.microsoft.com/v1.0`,
`login_base_url=https://login.microsoftonline.com`, `page_size=100`,
`scope=https://graph.microsoft.com/.default`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`, `config.scope` when `{{ config.token_url }}`.
- OAuth 2.0 client credentials authentication using `config.login_base_url`, `secrets.tenant_id`,
  `secrets.client_id`, `secrets.client_secret`, `config.scope`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `$top`=`1`.

## Streams notes

Default pagination: single request; no pagination.

- `users`: GET `/users` - records path `value`.
- `groups`: GET `/groups` - records path `value`.
- `applications`: GET `/applications` - records path `value`.
- `serviceprincipals`: GET `/servicePrincipals` - records path `value`.
- `directoryroles`: GET `/directoryRoles` - records path `value`.

## Write actions & risks

This connector is read-only. Read behavior: external Microsoft Graph API read of tenant directory
(users/groups/applications/service principals/directory roles) data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2, requires_elevated_scope=2.
