# Overview

Reads SharePoint/Microsoft Lists, list items, columns, and content types from a site through the
Microsoft Graph API using an OAuth2 client-credentials grant. Read-only.

Readable streams: `lists`, `list_items`, `columns`, `content_types`.

This connector is read-only; no write actions are declared.

Service API documentation: https://learn.microsoft.com/en-us/graph/api/resources/list.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://graph.microsoft.com/v1.0`; format `uri`; Microsoft
  Graph REST API base URL override for tests or proxies. Defaults to
  https://graph.microsoft.com/v1.0.
- `client_id` (optional, secret, string); Azure AD application (client) ID for OAuth2
  client-credentials. Never logged.
- `client_secret` (optional, secret, string); Azure AD application client secret for OAuth2
  client-credentials. Never logged.
- `list_id` (optional, string); SharePoint list ID; required only for the
  list_items/columns/content_types streams (the list whose items/columns/content types are read).
- `login_base_url` (optional, string); default `https://login.microsoftonline.com`; format `uri`;
  Azure AD login base URL override for tests or proxies; combined with tenant_id to derive the
  OAuth2 token endpoint.
- `max_pages` (optional, string); Permissive parse: empty, "all", or "unlimited" means unbounded.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records requested per @odata.nextLink page (1-200),
  sent as the $top query parameter on the first request of each stream.
- `scope` (optional, string); default `https://graph.microsoft.com/.default`; OAuth2
  client-credentials scope. Defaults to the static Graph application-permission scope.
- `site_id` (required, string); SharePoint site ID (or hostname:path pair) whose lists are read;
  used in every request path.
- `tenant_id` (optional, secret, string); Azure AD tenant ID (GUID or verified domain) used to
  derive the OAuth2 token endpoint.
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

Connection checks call GET `/sites/{{ config.site_id }}/lists`.

## Streams notes

Default pagination: single request; no pagination.

- `lists`: GET `/sites/{{ config.site_id }}/lists` - records path `value`.
- `list_items`: GET `/sites/{{ config.site_id }}/lists/{{ config.list_id }}/items` - records path
  `value`.
- `columns`: GET `/sites/{{ config.site_id }}/lists/{{ config.list_id }}/columns` - records path
  `value`.
- `content_types`: GET `/sites/{{ config.site_id }}/lists/{{ config.list_id }}/contentTypes` -
  records path `value`.

## Write actions & risks

This connector is read-only. Read behavior: external Microsoft Graph API read of a SharePoint site's
lists/list items/columns/content types.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3, requires_elevated_scope=1.
