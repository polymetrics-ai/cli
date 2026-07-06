# Overview

Reads Zoho Analytics workspace/view/table/organization/folder/query-table/datasource metadata and
triggers datasource/view data syncs, via the Zoho OAuth 2.0 refresh-token grant.

Readable streams: `workspaces`, `views`, `tables`, `organizations`, `recent_views`,
`shared_workspaces`, `shared_dashboards`, `folders`, `query_tables`, `datasources`.

Write actions: `sync_datasource`, `refetch_view_data`.

Service API documentation: https://www.zoho.com/analytics/api/v2/metadata-api.html.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://analyticsapi.zoho.com/restapi/v2`; format `uri`;
  Zoho Analytics Metadata API base URL override for tests or region-specific data centers (e.g.
  https://analyticsapi.zoho.eu/restapi/v2).
- `client_id` (required, secret, string); Zoho OAuth 2.0 client ID for the refresh-token grant. Used
  only in the token-request form; never logged.
- `client_secret` (required, secret, string); Zoho OAuth 2.0 client secret. Used only in the
  token-request form; never logged.
- `mode` (optional, string).
- `org_id` (optional, string); Optional Zoho Analytics organization ID; sent as the ZANALYTICS-ORGID
  header on every request when set.
- `refresh_token` (required, secret, string); Long-lived Zoho OAuth 2.0 refresh token. Exchanged for
  a short-lived access token at token_url; never logged. The 3-legged consent/acquisition dance is
  out of scope (credentials layer already owns it).
- `token_url` (optional, string); default `https://accounts.zoho.com/oauth/v2/token`; format `uri`;
  Zoho OAuth 2.0 token endpoint override. MUST be https in production; the hook fails closed on a
  non-https or unparseable value to prevent exfiltrating the refresh token to an attacker-chosen
  endpoint.
- `workspace_id` (optional, string); Reading/writing those without this set fails with an
  unresolved-config-key error naming workspace_id; every other stream/action does not require it.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`,
`refresh_token`.

Default configuration values: `base_url=https://analyticsapi.zoho.com/restapi/v2`,
`token_url=https://accounts.zoho.com/oauth/v2/token`.

Authentication behavior:

- Connector-specific authentication using `secrets.refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/workspaces`.

## Streams notes

Default pagination: single request; no pagination.

- `workspaces`: GET `/workspaces` - records path `data`; computed output fields `created_time`,
  `id`, `name`.
- `views`: GET `/views` - records path `data`; computed output fields `id`, `name`.
- `tables`: GET `/tables` - records path `data`; computed output fields `id`, `name`.
- `organizations`: GET `/orgs` - records path `data.orgs`; emits passthrough records.
- `recent_views`: GET `/recentviews` - records path `data.views`; emits passthrough records.
- `shared_workspaces`: GET `/workspaces/shared` - records path `data.workspaces`; emits passthrough
  records.
- `shared_dashboards`: GET `/dashboards/shared` - records path `data.views`; emits passthrough
  records.
- `folders`: GET `/workspaces/{{ config.workspace_id }}/folders` - records path `data.folders`;
  emits passthrough records.
- `query_tables`: GET `/workspaces/{{ config.workspace_id }}/querytables` - records path
  `data.queryTables`; emits passthrough records.
- `datasources`: GET `/workspaces/{{ config.workspace_id }}/datasources` - records path
  `data.dataSources`; emits passthrough records.

## Write actions & risks

Overall write risk: triggers an asynchronous datasource or view data sync/refetch in Zoho Analytics;
does not create, modify, or delete any Zoho Analytics workspace/view/table/data record itself --
only re-pulls from an already-configured external datasource.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `sync_datasource`: POST `/workspaces/{{ record.workspace_id }}/datasources/{{ record.datasource_id
  }}/sync` - kind `update`; body type `none`; path fields `workspace_id`, `datasource_id`; required
  record fields `workspace_id`, `datasource_id`; accepted fields `datasource_id`, `workspace_id`;
  risk: triggers an asynchronous data sync for one datasource in a workspace; low-risk (re-fetches
  data from the connected source, does not itself mutate any Zoho Analytics record). The documented
  optional CONFIG query parameter, which can carry a datasource's own username/password credential
  for the sync, is NOT supported by this action (see docs.md Known limits) -- only the no-CONFIG
  invocation shown in Zoho's own sample request is modeled.
- `refetch_view_data`: POST `/workspaces/{{ record.workspace_id }}/views/{{ record.view_id }}/sync`
  - kind `update`; body type `none`; path fields `workspace_id`, `view_id`; required record fields
  `workspace_id`, `view_id`; accepted fields `view_id`, `workspace_id`; risk: triggers an
  asynchronous data refetch for one view from its available datasource; low-risk (re-fetches, does
  not itself mutate any Zoho Analytics record). Same CONFIG-credential limitation as sync_datasource
  -- see docs.md Known limits.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, out_of_scope=9.
