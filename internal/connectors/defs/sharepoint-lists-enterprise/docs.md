# Overview

Reads and writes SharePoint lists and list items through Microsoft Graph.

Readable streams: `lists`, `list_items`.

Write actions: `create_list`, `update_list`, `create_list_item`, `update_list_item`.

Service API documentation:
https://learn.microsoft.com/en-us/graph/api/resources/list?view=graph-rest-1.0.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://graph.microsoft.com/v1.0`; format `uri`; Microsoft
  Graph API base URL override for tests or proxies.
- `client_id` (required, secret, string); Azure AD application (client) ID for OAuth2
  client-credentials. Never logged.
- `client_secret` (required, secret, string); Azure AD application client secret for OAuth2
  client-credentials. Never logged.
- `list_id` (optional, string); SharePoint list ID; required only for the list_items stream (the
  list whose items are read).
- `login_base_url` (optional, string); default `https://login.microsoftonline.com`; format `uri`;
  Azure AD login base URL override for tests or proxies; combined with tenant_id to form the OAuth2
  token endpoint.
- `mode` (optional, string).
- `site_id` (required, string); SharePoint site ID (or hostname:path pair) whose lists are read;
  used in every request path.
- `tenant_id` (required, string); Azure AD tenant ID (GUID or verified domain) used to build the
  OAuth2 token endpoint https://login.microsoftonline.com/<tenant_id>/oauth2/v2.0/token.
- `token_url` (optional, string); format `uri`; Full OAuth2 token endpoint override. When set, takes
  priority over the derived login_base_url/tenant_id endpoint (see docs.md Known limits for the
  dual-candidate auth mechanism used to express this).

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://graph.microsoft.com/v1.0`,
`login_base_url=https://login.microsoftonline.com`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret` when `{{ config.token_url }}`.
- OAuth 2.0 client credentials authentication using `config.login_base_url`, `config.tenant_id`,
  `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/sites/{{ config.site_id }}/lists`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `$skip`; limit parameter `$top`; page
size 100; maximum 1 page(s).

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `lists`: GET `/sites/{{ config.site_id }}/lists` - records path `value`; offset/limit pagination;
  offset parameter `$skip`; limit parameter `$top`; page size 100; maximum 1 page(s); incremental
  cursor `lastModifiedDateTime`; formatted as `rfc3339`; emits passthrough records.
- `list_items`: GET `/sites/{{ config.site_id }}/lists/{{ config.list_id }}/items` - records path
  `value`; offset/limit pagination; offset parameter `$skip`; limit parameter `$top`; page size 100;
  maximum 1 page(s); incremental cursor `lastModifiedDateTime`; formatted as `rfc3339`; emits
  passthrough records.

## Write actions & risks

Overall write risk: creates/updates SharePoint lists and list items (rows and their column values)
on the configured site via Microsoft Graph.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_list`: POST `/sites/{{ config.site_id }}/lists` - kind `create`; body type `json`;
  required record fields `displayName`; accepted fields `columns`, `description`, `displayName`,
  `list`; risk: creates a new SharePoint list (and any custom columns/template declared in the
  request) on the configured site; low-risk external mutation, no approval required.
- `update_list`: PATCH `/sites/{{ config.site_id }}/lists/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `description`,
  `displayName`, `id`; risk: mutates an existing list's display name/description by id; low-risk
  external mutation, no approval required.
- `create_list_item`: POST `/sites/{{ config.site_id }}/lists/{{ config.list_id }}/items` - kind
  `create`; body type `json`; required record fields `fields`; accepted fields `fields`; risk:
  creates a new item (row) in the configured list, with the submitted column values; low-risk
  external mutation, no approval required.
- `update_list_item`: PATCH `/sites/{{ config.site_id }}/lists/{{ config.list_id }}/items/{{
  record.id }}/fields` - kind `update`; body type `json`; path fields `id`; required record fields
  `id`; accepted fields `id`; risk: mutates an existing list item's column values by id, via the
  fields sub-resource (Graph's fieldValueSet update); only the submitted column names are changed,
  matching Graph's own partial-update semantics for this endpoint.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s), 4 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=3, duplicate_of=3, out_of_scope=8.
