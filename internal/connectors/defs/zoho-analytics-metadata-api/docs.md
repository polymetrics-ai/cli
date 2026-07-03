# Overview

Zoho Analytics Metadata API is a Tier-2 (AuthHook) migration repairing the `AUTH_COMPLEX` quarantine
entry recorded in `docs/migration/quarantine.json` ("OAuth refresh-token exchange (hook needed)").
It reads Zoho Analytics workspace, view, and table metadata via the Zoho OAuth 2.0
**refresh-token grant** only — the 3-legged consent/acquisition dance is out of scope (the refresh
token arrives as a pre-issued secret; the credentials layer already owns acquisition/storage),
matching gmail's precedent (`internal/connectors/hooks/gmail/hooks.go`) and mirroring the sibling
zoho-bigin migration's identical hook shape. This bundle migrates
`internal/connectors/zoho-analytics-metadata-api` (the hand-written connector it replaces); the
legacy package stays registered and unchanged until wave6's registry flip. Read-only: legacy
`zoho_analytics_metadata_api.go`'s `Write` always returns `ErrUnsupportedOperation`, and this bundle
declares `capabilities.write: false` with no `writes.json` to match.

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

Three streams, all primary-keyed on `id`, none paginated (`base.pagination: {"type": "none"}`,
matching legacy: `zoho_analytics_metadata_api.go`'s `Read` issues exactly one request per stream,
with no page/offset/cursor query parameter anywhere):

- `workspaces` — `GET /workspaces`, records at `data`. A `computed_fields` entry renames the raw
  `createdTime` to the schema's `created_time` (a clean single-source rename, matching legacy's
  `mapWorkspace`'s `"created_time": item["createdTime"]`). Declared `projection: "passthrough"` —
  see Known limits for why `id`/`name` are not similarly derived.
- `views` — `GET /views`, records at `data`. Declared `projection: "passthrough"` — see Known
  limits.
- `tables` — `GET /tables`, records at `data`. Declared `projection: "passthrough"` for the same
  reason as `views` (legacy's `mapView` function is shared verbatim by both streams).

## Write actions & risks

None — Zoho Analytics Metadata API is read-only. `capabilities.write: false`, no `writes.json`
file, matching legacy's `ErrUnsupportedOperation` (`zoho_analytics_metadata_api.go:122-124`).

## Known limits

- **`workspaces`/`views`/`tables` streams do not reproduce legacy's multi-field id/name coalesce.**
  Legacy's `mapWorkspace` derives `id` as `first(item["workspaceId"], item["id"])` and `name` as
  `first(item["workspaceName"], item["name"])`; `mapView` (shared by both the `views` and `tables`
  streams) derives `id` as `first(item["viewId"], item["tableId"], item["id"])` and `name` as
  `first(item["viewName"], item["tableName"], item["name"])` — first non-empty value wins across
  differently-named raw fields. The engine's `computed_fields` dialect has no coalesce/
  fallback-across-multiple-source-fields primitive (`docs/migration/conventions.md` §3: every
  `computed_fields` entry is a single template resolved against one reference or literal, with only
  "skip if THIS entry's source is absent" tolerance, never "try field A, else field B, else field
  C"). Declaring only the first-priority field would silently drop records where legacy would have
  fallen back to a secondary field — an accepted-input emitted-DATA change, not cosmetic. All three
  streams are instead declared `projection: "passthrough"`: every raw field (`workspaceId`,
  `workspaceName`, `viewId`, `viewName`, `tableId`, `tableName`, and any other API field) survives
  verbatim, strictly more permissive than legacy (a downstream consumer can reproduce legacy's exact
  coalesce priority itself) and never drops data legacy would have emitted for any accepted input.
  This mirrors the identical, independently-documented deviation in the sibling zoho-bigin
  migration's docs.md. Classified ACCEPTABLE per `docs/migration/conventions.md` §5 (never
  drops/changes data for any legacy-accepted input, differs only in also exposing additional raw
  fields legacy's narrower projection discarded).
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
- **`TestConformance/zoho-analytics-metadata-api`'s dynamic (fixture-replay) checks are
  `skip_dynamic`'d** for the identical reason as gmail's bundle-level marker: this bundle's *sole*
  auth candidate is `mode: custom`, and conformance's synthetic config can never carry a real
  `https` `token_url` — the AuthHook's own https-only guard means no synthetic secret value can ever
  satisfy it, so every auth-resolving dynamic check would fail identically and uninformatively
  regardless of hook wiring. `hooks/zoho-analytics-metadata-api/hooks_test.go` is the authoritative
  substitute proof for the AuthHook's real OAuth2 refresh-grant behavior (form shape, caching/
  expiry, https enforcement, error paths, secret redaction) — the same gmail precedent this bundle's
  `metadata.json` `conformance.reason` names.
