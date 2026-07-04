# Overview

Zoho Bigin is a Tier-2 (AuthHook) migration repairing the `AUTH_COMPLEX` quarantine entry recorded
in `docs/migration/quarantine.json` ("Legacy performs an OAuth2 refresh_token grant token exchange
... before every read/check. The engine's declarative auth dialect only supports the oauth2 [client
credentials grant]"). It reads and writes Zoho Bigin CRM data via the Zoho OAuth 2.0
**refresh-token grant** only — the 3-legged consent/acquisition dance is out of scope (the refresh
token arrives as a pre-issued secret; the credentials layer already owns acquisition/storage),
matching gmail's precedent (`internal/connectors/hooks/gmail/hooks.go`). This bundle migrates
`internal/connectors/zoho-bigin` (the hand-written connector it replaces); the legacy package stays
registered and unchanged until wave6's registry flip.

**Pass B full-surface expansion** (this revision): legacy `zoho_bigin.go` only ever read 3 streams
(`pipelines`, generic `records`, `fields`) and was read-only (`Write` always returned
`ErrUnsupportedOperation`). This bundle now covers the full documented Zoho Bigin API v2 REST
surface reachable from the standard refresh-token grant: 12 read streams (`pipelines`, `records`,
`fields`, `contacts`, `companies`, `products`, `tasks`, `events`, `calls`, `notes`, `users`, `tags`,
`modules`) and 6 write actions (`create_record`/`update_record`/`upsert_record`/`delete_record` on
the generic module surface, plus `create_note`/`delete_note`). `capabilities.write` is now `true`;
see `api_surface.json` for the full endpoint-by-endpoint coverage/exclusion ledger (binary
attachment/photo upload-download, the separate Bulk/Notifications/COQL API families, and several
settings/admin-shaped endpoints are excluded with a specific real-vocabulary reason each — no
blanket "Pass B" bucket).

## Auth setup

Provide three secrets: `client_id`, `client_secret`, and `client_refresh_token` (long-lived; never
logged) — all three are `required` in `spec.json`, matching legacy's `requireOAuth` check (unlike
gmail, Zoho Bigin's legacy connector treats `client_secret` as mandatory, not optional).
`hooks/zoho-bigin/hooks.go` implements `AuthHook`, copying gmail's hook pattern
(`docs/migration/conventions.md` §1's Tier-2 table: token-exchange auth) adapted for zoho-bigin's
own required-field shape: it POSTs `grant_type=refresh_token` + `client_id` + `client_secret` +
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
mirrors legacy's `validateURL` (`zoho_bigin.go:233-241`) but tightens it to https-only in the hook
(legacy's `validateURL` also accepted plain `http`) — see Known limits.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook": "zoho-bigin",
...}` — legacy has no alternate auth path (no static API key, no public/no-auth fallback), so there
is no `when`-gated bypass to declare.

## Streams notes

All 12 streams are primary-keyed on `id`. `base.pagination` is now `{"type": "page_number",
"page_param": "page", "size_param": "per_page", "start_page": 1, "page_size": 200}` — Zoho Bigin's
real documented `get-records`/`get-related-records`/`get-notes`/`get-users` shape (`page`/
`per_page`, 200 being both the default AND max `per_page` value; the engine's short-page stop
signal terminates pagination the same way `more_records: false` would). This supersedes the
pre-Pass-B `{"type": "none"}` declaration: legacy's own hand-written `Read` never paginated (a
single request per stream), which was a **legacy limitation**, not a real single-page API — Pass B
brings the 3 pre-existing streams (`pipelines`/`records`/`fields`) in line with the real,
fully-paginated API surface, matching every sibling module-list stream added this pass.
`fields`/`tags`/`modules` keep `{"type": "none"}` (stream-level override) since Zoho Bigin's own
metadata/settings endpoints are not documented as paginated.

- `pipelines` — `GET /Pipelines`, records at `data`. Schema projection (default mode) matches
  legacy's `mapPipeline` exactly: `id`, `name`, `display_value`, no rename needed (raw field names
  already match the schema).
- `records` — `GET /{{ config.module_name }}` (defaults to `Deals`, matching legacy's
  `zoho_bigin.go:103-107` fallback when `module_name` is unset), records at `data`. Declared
  `projection: "passthrough"` rather than schema projection — see Known limits for why.
- `fields` — `GET /settings/fields`, records at `fields`. Declared `projection: "passthrough"` for
  the same reason as `records` — see Known limits.
- `contacts` — `GET /Contacts`, records at `data`, `passthrough` (same module-family coalesce
  concern as `records`/`fields` — see Known limits).
- `companies` — `GET /Accounts` (Bigin's "Companies" module; the wire API name remains `Accounts`,
  matching Zoho's own CRM-family naming), records at `data`, `passthrough`.
- `products` — `GET /Products`, records at `data`, `passthrough`.
- `tasks` — `GET /Tasks`, records at `data`, `passthrough`.
- `events` — `GET /Events`, records at `data`, `passthrough`.
- `calls` — `GET /Calls`, records at `data`, `passthrough`.
- `notes` — `GET /Notes` (the module-scoped "every note across every module" listing, not the
  narrower per-parent-record `GET /{module}/{id}/Notes` related-list variant — see `api_surface.json`
  for why the related-list form is out of scope), records at `data`, `passthrough`.
- `users` — `GET /users?type=AllUsers`, records at the top-level `users` key (not `data` — a
  distinct envelope shape from every module-record stream), `passthrough`.
- `tags` — `GET /settings/tags?module={{ config.module_name }}`, records at `tags`, `passthrough`,
  unpaginated. Reuses `module_name` (the same config key `records` uses) to scope which module's
  tag set is returned, rather than declaring a second module-selector config key.
- `modules` — `GET /settings/modules`, records at `modules`, `passthrough`, unpaginated. Module
  metadata (which modules exist, their capability flags) rather than module records.

## Write actions & risks

Zoho Bigin is now writable (`capabilities.write: true`). All 6 actions operate through the real
Bigin v2 write endpoints, which wrap every payload in a top-level `data` **array** (not the
single-object envelope some other JSON:API-style bundles use) — expressed declaratively by
`record_schema` requiring a `data` property of `"type": "array"`; the engine's default JSON body
construction (`docs/migration/conventions.md` §3) copies whatever value the caller supplies for
`record["data"]` verbatim into the request body, so a caller-supplied JSON array serializes exactly
as Zoho's wire format expects with no hook needed.

- `create_record` — `POST /{{ config.module_name }}` with `{"data": [...], "trigger": [...]}`.
  Inserts 1-100 new records in the configured module. Low-risk external mutation, approval required.
- `update_record` — `PUT /{{ config.module_name }}` with `{"data": [{"id": ..., ...}, ...]}`.
  Updates 1-100 existing records by id; omitted fields are left unmodified server-side. Approval
  required.
- `upsert_record` — `POST /{{ config.module_name }}/upsert` with `{"data": [...],
  "duplicate_check_fields": [...]}`. Inserts or overwrites based on Zoho's own duplicate-detection
  logic (system/user-defined unique fields when `duplicate_check_fields` is omitted). Approval
  required.
- `delete_record` — `DELETE /{{ config.module_name }}/{{ record.id }}`. Single-record delete (the
  real API's bulk `?ids=a,b,c` form is not exposed — see Known limits); `missing_ok_status: [404]`
  makes a delete-of-already-deleted idempotent. Destructive confirmation required.
- `create_note` — `POST /{{ config.module_name }}/{{ record.parent_id }}/Notes` with
  `{"data": [{"Note_Title": ..., "Note_Content": ...}, ...]}`. Attaches 1-100 notes to an existing
  record. Low-risk, no approval required.
- `delete_note` — `DELETE /{{ config.module_name }}/{{ record.parent_id }}/Notes/{{ record.id }}`.
  Single-note delete, idempotent on 404. Destructive confirmation required.

## Known limits

- **`records` and `fields` streams do not reproduce legacy's multi-field name/id coalesce.**
  Legacy's `mapRecord` derives its `name` output field as `first(item["name"], item["Deal_Name"],
  item["display_value"])` (first non-empty value wins across three differently-shaped raw fields,
  since different Zoho Bigin modules use different display-name conventions), and `mapField`
  derives `id` as `first(item["id"], item["api_name"])`. The engine's `computed_fields` dialect has
  no coalesce/fallback-across-multiple-source-fields primitive (`docs/migration/conventions.md` §3:
  every `computed_fields` entry is a single template resolved against one reference or literal, with
  only "skip if THIS entry's source is absent" tolerance, never "try field A, else field B, else field
  C"). Declaring only the first-priority field (e.g. `name` alone) would silently drop records where
  legacy would have fallen back to `Deal_Name`/`display_value`/`api_name` — an accepted-input
  emitted-DATA change, not cosmetic. Both streams are instead declared `projection: "passthrough"`:
  every raw field (`id`, `name`, `Deal_Name`, `display_value`, `api_name`, `display_label`, and any
  other module-specific field) survives verbatim, strictly more permissive than legacy (a downstream
  consumer can reproduce legacy's exact coalesce priority itself, or read the specific field it
  needs) and never drops data legacy would have emitted for any accepted input. This is documented
  here per `docs/migration/conventions.md` §5's parity-deviation ledger convention; classified
  ACCEPTABLE (never drops/changes data for any legacy-accepted input, differs only in also exposing
  additional raw fields legacy's narrower projection discarded). The 8 Pass-B-added module-shaped
  streams (`contacts`/`companies`/`products`/`tasks`/`events`/`calls`/`notes`/`users`/`tags`/
  `modules`) are `passthrough` for the identical reason — every one of them shares Zoho Bigin's same
  cross-module display-name/lookup-object field-shape variability, and none of them have a prior
  legacy implementation to diverge from (this is fresh Pass B surface, not a migration deviation).
- **Pass B's `page_number` pagination on `pipelines`/`records` is a real-API-surface correction,
  not a parity deviation.** Legacy's hand-written `Read` issued exactly one request per stream with
  no page parameter — a genuine limitation of the legacy implementation (it never paged past the
  first ~200 records of any module), not a documented single-page API contract. Per
  `docs/migration/conventions.md` §8 rule 3 ("live config must reproduce legacy defaults, never
  inherit fixture conveniences") and the Pass B full-surface-expansion brief's mandate to implement
  every practical GET list/detail resource, this bundle now paginates through Zoho Bigin's real,
  documented `page`/`per_page`/`more_records` shape. This is a genuine functional improvement over
  legacy (it now returns ALL records instead of only the first page), never a narrowing — recorded
  here for traceability, not as an ACCEPTABLE/ENGINE_GAP ledger entry, since no accepted-input
  behavior is lost.
- **Bulk multi-id delete (`DELETE /{module}?ids=a,b,c`) is not exposed** — only the single-record
  `DELETE /{module}/{id}` path-parameterized form (`delete_record`). The write dialect's
  `WriteAction` has no per-record batching/query-param construction primitive for a write action (only
  `path`/`path_fields`, resolved once per record in the write loop — see
  `docs/migration/conventions.md` §3's write body construction section); expressing the bulk form
  would require collecting multiple records' ids into one request, which the engine's one-record-per-
  write-call loop (`write.go`) does not support for any bundle. Single-record delete is strictly
  correct or a subset of the bulk form for any accepted input (same eventual per-record outcome, more
  requests) — ACCEPTABLE.
- **`record_schema`'s array-typed `data` fields have no `minItems`/`maxItems` cardinality
  enforcement.** The engine's draft-07 subset (`internal/connectors/engine/schema.go`) supports only
  `type`/`required`/`properties`/`items`/`enum`/`pattern`/`minProperties`/`additionalProperties` plus
  the `x-*` extensions — `minItems`/`maxItems` are unknown keywords and hard-fail
  `CompileSchema`/`write_schemas_valid`. Zoho Bigin's real 1-100-records-per-call cap is documented
  in each write action's `data` field `description` and enforced server-side (a call with 0 or >100
  records fails at the API, not at this bundle's dry-run validation layer). ACCEPTABLE: never
  silently accepts an out-of-range call as if it were valid data, just defers the cardinality check
  to the same place legacy always would have (there is no legacy write path to compare against here
  at all — Bigin was read-only until this pass).
- **`users`/`tags`/`modules` are assumed unpaginated (`pagination: {"type": "none"}`).** Zoho
  Bigin's public docs do not document `page`/`per_page` support for `/settings/tags` or
  `/settings/modules` at all (both are small, org-scoped configuration lists — realistically well
  under any page-size threshold for any real Bigin organization); `/users` DOES document `page`/
  `per_page`, but its response envelope's records key is `users`, not `data`, and no other stream in
  this bundle shares that shape to safely reuse — `users` is declared paginated
  (inherits `base.pagination`) since its docs explicitly support it, while `tags`/`modules` override
  to `none` since theirs don't document it. If a real organization ever exceeds a single page of
  tags/modules, this would silently truncate — SCHEMA_AMBIGUOUS-adjacent but not filed as a blocker
  since no public documentation describes pagination for either endpoint to implement against.
- **`token_url` https-only enforcement is stricter than legacy's `validateURL`** (which accepted
  plain `http` too, `zoho_bigin.go:233-241`): the hook only accepts `https://` overrides. Never
  stricter for any *production* Zoho OAuth endpoint, which is always https; strictly safer for the
  one new SSRF-adjacent secret-bearing surface this migration adds. See the parity-deviation ledger
  in `docs/migration/conventions.md` §5.
- **`data_center` is not modeled as a config key.** Legacy's own test fixtures set a `data_center`
  config value, but `zoho_bigin.go` never reads it anywhere (dead config in legacy itself, not just
  in this migration) — `base_url` is the sole, already-correct override mechanism for a
  region-specific data center (e.g. `https://www.zohoapis.eu/bigin/v2`). Not declared in `spec.json`
  per `docs/migration/conventions.md` F6 (a spec property with no wired template is dead config).
- **`TestConformance/zoho-bigin`'s dynamic (fixture-replay) checks are `skip_dynamic`'d** for the
  identical reason as gmail's bundle-level marker: this bundle's *sole* auth candidate is `mode:
  custom`, and conformance's synthetic config can never carry a real `https` `token_url` — the
  AuthHook's own https-only guard means no synthetic secret value can ever satisfy it, so every
  auth-resolving dynamic check would fail identically and uninformatively regardless of hook wiring.
  `hooks/zoho-bigin/hooks_test.go` is the authoritative substitute proof for the AuthHook's real
  OAuth2 refresh-grant behavior (form shape, caching/expiry, https enforcement, error paths, secret
  redaction) — the same gmail precedent this bundle's `metadata.json` `conformance.reason` names.
