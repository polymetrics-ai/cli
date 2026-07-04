# Overview

Zoho Analytics Metadata API is a Tier-2 (AuthHook) migration repairing the `AUTH_COMPLEX` quarantine
entry recorded in `docs/migration/quarantine.json` ("OAuth refresh-token exchange (hook needed)").
It reads Zoho Analytics workspace, view, and table metadata via the Zoho OAuth 2.0
**refresh-token grant** only — the 3-legged consent/acquisition dance is out of scope (the refresh
token arrives as a pre-issued secret; the credentials layer already owns acquisition/storage),
matching gmail's precedent (`internal/connectors/hooks/gmail/hooks.go`) and mirroring the sibling
zoho-bigin migration's identical hook shape. This bundle migrates
`internal/connectors/zoho-analytics-metadata-api` (the hand-written connector it replaces); the
legacy package stays registered and unchanged until wave6's registry flip.

**Pass B full-surface expansion** (`api_surface.json`) added every remaining GET endpoint in Zoho
Analytics' **Metadata API** category specifically — `organizations`, `recent_views`,
`shared_workspaces`, `shared_dashboards`, `folders`, `query_tables`, `datasources` — plus the 2 POST
data-sync-trigger mutations that category documents (`sync_datasource`, `refetch_view_data`), so
`capabilities.write` is now `true`. The connector's name/`docs_url` deliberately scope it to the
Metadata API category alone; Zoho Analytics' separate Workspace Management/Modeling/Bulk/
Data-Manipulation API categories (workspace/view/table/column CRUD, row-level data import/export)
are out of scope — see `api_surface.json`'s per-endpoint `out_of_scope` reasons.

## Auth setup

Provide three secrets: `client_id`, `client_secret`, and `refresh_token` (long-lived; never logged)
— all three are `required` in `spec.json`, matching legacy's `requireOAuth` check.
`hooks/zoho-analytics-metadata-api/hooks.go` implements `AuthHook`, copying gmail's hook pattern
(`docs/migration/conventions.md` §1's Tier-2 table: token-exchange auth) and mirroring the sibling
zoho-bigin hook: it POSTs `grant_type=refresh_token` + `client_id` + `client_secret` +
`refresh_token` to `token_url` (default `https://accounts.zoho.com/oauth/v2/token`,
config-overridable), caches the resulting access token until 60 seconds before its declared expiry,
and sets `Authorization: Zoho-oauthtoken <access_token>` on every request (Zoho's own header scheme
— legacy's `refreshToken` decodes `access_token` from the JSON response and the read path applies it
via `connsdk.Bearer`, which legacy itself sends as a plain `Bearer <token>` header; this bundle
instead uses Zoho's documented `Zoho-oauthtoken` scheme directly in the hook, since a custom
AuthHook is not constrained to `connsdk.Bearer`'s prefix — this is a stricter-correctness match to
Zoho's own published API contract, not a deviation from any legacy-observable behavior since legacy
only replayed the raw access token string it received).

`token_url` MUST resolve to an `https://` URL: the hook fails closed on a non-https or unparseable
override rather than sending the refresh token/client secret to an attacker-chosen endpoint. This
mirrors legacy's `validateURL` (`zoho_analytics_metadata_api.go:226-234`) but tightens it to
https-only in the hook (legacy's `validateURL` also accepted plain `http`) — see Known limits.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook":
"zoho-analytics-metadata-api", ...}` — legacy has no alternate auth path (no static API key, no
public/no-auth fallback), so there is no `when`-gated bypass to declare.

Optional `org_id` config sends a `ZANALYTICS-ORGID` header on every request when set, matching
legacy's `zoho_analytics_metadata_api.go:101-104` conditional header (declared but not in
`required[]` — omitted entirely when unset, per `docs/migration/conventions.md` §3's conditional
header rule, matching legacy's `if orgID != "" { headers[...] = orgID }` guard exactly).

## Streams notes

Nine streams total, primary-keyed on whichever id field their own real API response uses (`id` for
the 3 legacy streams; `orgId`/`viewId`/`workspaceId`/`folderId`/`datasourceId` for the 6 new ones —
each schema's `x-primary-key` names the field that stream's own response actually returns, per
`docs/migration/conventions.md`'s schema-as-projection rule). None are paginated
(`base.pagination: {"type": "none"}` at the base level; none of the 9 streams override it) — every
metadata-api.html endpoint researched for this Pass B pass returns its full result set in one
response with no page/cursor/next-link field of any kind (confirmed by reading each endpoint's own
documented sample response).

The original 3 legacy-parity streams reproduce the hand-written connector's mapper output:

- `workspaces` — `GET /workspaces`, records at `data`. `computed_fields` maps `id` from
  `workspaceId` falling back to `id`, `name` from `workspaceName` falling back to `name`, and
  `created_time` from `createdTime`, matching legacy's `mapWorkspace`.
- `views` — `GET /views`, records at `data`. `computed_fields` maps `id` from the first non-empty
  value of `viewId`, `tableId`, or `id`, and `name` from `viewName`, `tableName`, or `name`,
  matching legacy's shared `mapView`.
- `tables` — `GET /tables`, records at `data`. Uses the same `mapView`-equivalent mapping as
  `views`, since the legacy Go implementation shared the mapper for both streams.

The 6 new Pass-B streams model the REAL Zoho Analytics Metadata API response envelopes (each
endpoint's own documented sample response was read individually rather than assumed, since the
envelope key differs per endpoint — `data.orgs`, `data.views`, `data.workspaces`, `data.folders`,
`data.queryTables`, `data.dataSources` are all genuinely different keys, not a uniform `data.data`
shape):

- `organizations` — `GET /orgs`, records at `data.orgs`. Top-level, no workspace scoping.
- `recent_views` — `GET /recentviews`, records at `data.views`. Top-level, no workspace scoping.
- `shared_workspaces` — `GET /workspaces/shared`, records at `data.workspaces`.
- `shared_dashboards` — `GET /dashboards/shared`, records at `data.views` (Zoho's own docs use the
  same `views` envelope key for both regular views and dashboards — a dashboard is itself a `view`
  object with `viewType: "Dashboard"`).
- `folders` — `GET /workspaces/{{ config.workspace_id }}/folders`, records at `data.folders`.
  **Requires `config.workspace_id`** — see Known limits.
- `query_tables` — `GET /workspaces/{{ config.workspace_id }}/querytables`, records at
  `data.queryTables`. **Requires `config.workspace_id`**.
- `datasources` — `GET /workspaces/{{ config.workspace_id }}/datasources`, records at
  `data.dataSources`. **Requires `config.workspace_id`**. Each record's `tableDetails` field is a
  nested array of per-table sync-status objects, preserved as an opaque JSON array
  (`type: ["array","null"]`) rather than fanned out into separate records — there is no legacy
  behavior to match (this is entirely new coverage) and fanning it out would require a `fan_out`
  spec keyed on a per-datasource table list this connector has no other reason to enumerate
  up-front.

## Write actions & risks

Two write actions, both triggering an asynchronous Zoho-side data sync rather than mutating any
Zoho Analytics object — `capabilities.write` is `true`:

- `sync_datasource` — `POST /workspaces/{workspace_id}/datasources/{datasource_id}/sync`,
  `body_type: none`, `path_fields: ["workspace_id", "datasource_id"]`. Initiates a data sync for one
  datasource. Low risk: re-fetches from an already-configured external source, never creates/
  modifies/deletes a Zoho Analytics record.
- `refetch_view_data` — `POST /workspaces/{workspace_id}/views/{view_id}/sync`, `body_type: none`,
  `path_fields: ["workspace_id", "view_id"]`. Initiates a data refetch for one view. Same low-risk
  profile as `sync_datasource`.

Both endpoints also document an optional `CONFIG` query parameter that can carry the target
datasource's OWN `userName`/`password` credential (and, for `sync_datasource` only, a
`syncIntervalId` when multiple sync schedules are configured) to authenticate the sync — this
bundle does NOT model that parameter; see Known limits.

## Known limits

- **`token_url` https-only enforcement is stricter than legacy's `validateURL`** (which accepted
  plain `http` too, `zoho_analytics_metadata_api.go:226-234`): the hook only accepts `https://`
  overrides. Never stricter for any *production* Zoho OAuth endpoint, which is always https;
  strictly safer for the one new SSRF-adjacent secret-bearing surface this migration adds. See the
  parity-deviation ledger in `docs/migration/conventions.md` §5.
- **`data_center` is not modeled as a config key.** Legacy's own test fixtures set a `data_center`
  config value, but `zoho_analytics_metadata_api.go` never reads it anywhere (dead config in legacy
  itself, not just in this migration) — `base_url` is the sole, already-correct override mechanism
  for a region-specific data center (e.g. `https://analyticsapi.zoho.eu/restapi/v2`). Not declared
  in `spec.json` per `docs/migration/conventions.md` F6 (a spec property with no wired template is
  dead config).
- **`folders`/`query_tables`/`datasources` streams and both write actions require `config.workspace_id`
  set.** These are the only genuinely workspace-scoped paths in the real Metadata API this bundle
  models (unlike `workspaces`/`views`/`tables`, whose flat legacy-parity paths return every
  accessible object with no workspace segment at all). `workspace_id` is declared optional in
  `spec.json` (not `required[]`, since the 3 legacy streams and `organizations`/`recent_views`/
  `shared_workspaces`/`shared_dashboards` never need it) — reading/writing one of the
  workspace-scoped items without it configured hard-errors with an unresolved-config-key message
  naming `workspace_id` (path interpolation has no absent-key tolerance, per
  `docs/migration/conventions.md` §3), which is the honest, loud failure mode rather than a silent
  wrong-URL request.
- **`sync_datasource`/`refetch_view_data` do not model the optional `CONFIG` query parameter.** Both
  endpoints document an optional `CONFIG` JSONObject query parameter that can carry the target
  datasource's own `userName`/`password` (and, for `sync_datasource`, a `syncIntervalId` when
  multiple sync schedules exist) to authenticate the triggered sync. This is out of scope for two
  reasons: (1) the parameter's shape is itself a second, datasource-scoped credential pair with no
  natural home in this connector's `spec.json` (it varies per datasource, not per connection), and
  (2) Zoho's own sample request for both endpoints omits `CONFIG` entirely, confirming the
  no-parameter invocation is the documented common case. Only that no-`CONFIG` invocation is
  modeled; a datasource that specifically requires re-authentication on every sync cannot be synced
  through this action today.
- **`meta_details` (`GET /metadetails`) and view-dependents (`GET
  /workspaces/{workspace-id}/views/{view-id}/dependents`) are excluded, not migrated** — see
  `api_surface.json` for the specific per-endpoint reasons (an opaque target-type-dependent `CONFIG`
  shape for the former; a view-id fan-out this connector has no other reason to collect for the
  latter).
- **`TestConformance/zoho-analytics-metadata-api`'s dynamic (fixture-replay) checks are
  `skip_dynamic`'d** for the identical reason as gmail's bundle-level marker: this bundle's *sole*
  auth candidate is `mode: custom`, and conformance's synthetic config can never carry a real
  `https` `token_url` — the AuthHook's own https-only guard means no synthetic secret value can ever
  satisfy it, so every auth-resolving dynamic check would fail identically and uninformatively
  regardless of hook wiring. `hooks/zoho-analytics-metadata-api/hooks_test.go` is the authoritative
  substitute proof for the AuthHook's real OAuth2 refresh-grant behavior (form shape, caching/
  expiry, https enforcement, error paths, secret redaction) — the same gmail precedent this bundle's
  `metadata.json` `conformance.reason` names.
