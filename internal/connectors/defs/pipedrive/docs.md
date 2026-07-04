# Overview

Pipedrive originated as a read-only declarative migration of `internal/connectors/pipedrive`
(legacy Go connector), and is expanded in Pass B. It reads the 6 legacy-parity resources (deals,
persons, organizations, activities, products, users) plus 14 new v1-native resources (notes, leads,
lead labels/sources/fields, saved filters, custom activity types, webhooks, legacy teams, roles,
and deal/person/organization/product field metadata), and writes 17 practical mutations (lead/note/
filter/activity-type/lead-label create-update-delete, plus webhook create/delete), through
Pipedrive's REST API v1. The 6 original streams stay capability-parity with legacy; legacy stays
registered and unchanged until wave6's registry flip.

**Important Pass B research finding**: Pipedrive has migrated full CRUD (list/get/create/update/
delete) for Deals, Persons, Organizations, Products, Activities, Pipelines, and Stages to API v2 —
the v1 OpenAPI document no longer declares plain `/deals`, `/persons`, `/organizations`, `/products`,
`/activities`, `/pipelines`, or `/stages` endpoints at all (only each resource's v1-exclusive
sub-resources — changelog, files, followers, merge, mailMessages, permittedUsers — remain under
v1), and the live v1 docs pages for "Get all deals"/"Add a new deal" etc. now redirect to
`/api/v2/deals`. This bundle's 6 original streams are UNCHANGED and keep reading the v1 list
endpoints (Pipedrive still serves them for backward compatibility) — this is a deliberate,
parity-preserving choice: migrating `base_url`/these streams to v2 is a real API-version change with
its own auth/pagination/schema implications, out of scope for a Pass B capability EXPANSION (which
adds coverage without altering existing accepted-input behavior). See Known limits.

## Auth setup

Provide a Pipedrive API token via the `api_token` secret. It is sent as the documented `api_token`
query-string parameter on every request (`auth.mode: api_key_query`), matching legacy's
`connsdk.APIKeyQuery("api_token", key)` exactly — legacy accepts either an `api_token` or (as a
fallback) an `api_key` secret name (`firstNonEmpty(secret(cfg,"api_token"), secret(cfg,"api_key"))`);
this bundle declares only `api_token` since the `api_key` fallback name has no separate documented
meaning and every legacy call site that actually authenticates supplies `api_token`. Never logged.
Every Pass-B-added stream and write action reuses this identical `api_token` query-param auth — none
of the new v1 resources requires a different credential shape.

## Streams notes

All 6 streams (`deals`, `persons`, `organizations`, `activities`, `products`, `users`) pass through
the RAW record with no field-filtering (`"projection": "passthrough"`) — legacy's `Read` emits
`connectors.Record(rec)` directly from the decoded `data` array with no `mapRecord`-style shaping
function. Pagination follows Pipedrive's `additional_data.pagination.next_start` cursor convention
(`pagination.type: cursor` with `token_path: additional_data.pagination.next_start`, `cursor_param:
start`): the next page's `start` query value is read straight from the previous response body, and
pagination stops when `next_start` is absent/null/empty — identical to legacy's
`strings.TrimSpace(next) == ""` stop check. No `stop_path` is declared: legacy's own stop condition
looks only at `next_start`, never at `more_items_in_collection`, so adding a `stop_path` gate on that
field would be new behavior, not a port.

`deals`, `persons`, and `organizations` send `since_timestamp` (an optional-query entry,
`omit_when_absent: true`) when `replication_start_date` is configured; `activities` sends the same
value as `since` (a different param name — matches legacy's per-endpoint `sinceParam` map:
`deals`/`persons`/`organizations` → `since_timestamp`, `activities` → `since`, `products`/`users` →
`""`, i.e. no since param wired at all for those two, matching this bundle's `products`/`users`
streams declaring no such query key). Legacy resolves the since value from `req.State["cursor"]`
first, falling back to `req.Config.Config["replication_start_date"]`; this bundle does not attempt to
express the state-cursor half of that fallback as a stream-level `incremental` block because none of
Pipedrive's since params here is wired as this dialect's `incremental.request_param` (no per-stream
`x-cursor-field`/state-cursor round-trip is declared) — see Known limits.

The 14 Pass-B-added streams split into three pagination shapes. `notes` reuses the base
`cursor`+`token_path: additional_data.pagination.next_start` shape exactly like the 6 legacy
streams. `leads`/`deal_fields`/`person_fields`/`organization_fields`/`product_fields`/`lead_fields`/
`roles` declare a stream-level `pagination: {"type": "offset_limit", "offset_param": "start",
"limit_param": "limit"}` override: Pipedrive's own OpenAPI document shows these endpoints' 
`additional_data` carries only `start`/`limit`/`more_items_in_collection` (`leads`) or no
pagination metadata at all (the field-metadata endpoints), never a `pagination.next_start` cursor
token — a genuinely different (short-page-stop, not token-stop) pagination convention from the base
default, requiring the override. `filters`/`activity_types`/`legacy_teams`/`webhooks`/`lead_labels`/
`lead_sources`/`currencies` declare `pagination: {"type": "none"}`: Pipedrive's spec declares no
`start`/`limit` query parameters at all for these endpoints, and each returns its full collection in
one response. All 14 new streams share the 6 legacy streams' `"projection": "passthrough"` shape.

`leads`' `id` is a UUID string (Pipedrive's Lead resource, unlike Deals/Persons/Organizations, is
keyed by UUID, not an auto-increment integer) — `schemas/leads.json`'s `x-primary-key` field is
typed `"string"` accordingly, and `lead_labels`' `id` is likewise a UUID string. `lead_sources` has
no `id`/numeric key of any kind in Pipedrive's own schema — only a bare `name` string per entry — so
`schemas/lead_sources.json` declares `x-primary-key: ["name"]`.

## Write actions & risks

17 write actions: `create_lead`/`update_lead`/`delete_lead`, `create_note`/`update_note`/
`delete_note`, `create_filter`/`update_filter`/`delete_filter`,
`create_activity_type`/`update_activity_type`/`delete_activity_type`,
`create_lead_label`/`update_lead_label`/`delete_lead_label`, and `create_webhook`/`delete_webhook`.
`update_lead` and `update_lead_label` use `method: "PATCH"` (Pipedrive's own documented partial-patch
verb for these two resources specifically); every other update action uses `PUT` (Pipedrive's
documented full-replacement verb for notes/filters/activityTypes). None of these actions declares
`missing_ok_status`: Pipedrive's OpenAPI document only formally documents a `200` response for each
delete operation used here (leads/leadLabels additionally document a `404`, but this bundle does not
invent idempotent-delete tolerance beyond what the other new streams' delete operations document).

Per `metadata.json`'s `risk.approval`: `create_lead`/`create_note`/`create_filter`/
`create_activity_type`/`create_lead_label`/`create_webhook` require no approval (low-risk,
non-destructive); every `update_*`/`delete_*` action requires approval.

## Known limits

- Legacy's `since`/`since_timestamp` value can come from EITHER the sync's persisted
  `state["cursor"]` OR the static `replication_start_date` config value (state takes precedence).
  This bundle only wires the static config half (`config.replication_start_date`) via the
  optional-query dialect; the state-cursor half is not modeled as a stream `incremental` block
  because no stream here declares `x-cursor-field`/`incremental.cursor_field` — the raw
  `update_time`/`add_time` fields legacy tracks are ordinary passthrough fields, not a declared
  incremental cursor tied to a `request_param`. This means every read is effectively a
  config-anchored (not truly resumable, state-cursor-advancing) read from wave2's bundle's
  perspective. Out of scope for wave2 fan-out; a future capability-expansion pass can promote these
  streams to full `incremental` blocks once the since-param-vs-state-cursor precedence is worth
  modeling declaratively.
- `page_size`/`max_pages` config knobs (legacy's client-side page-size and page-count bounds,
  1-500 and 0/all/unlimited respectively) have no bundle-level equivalent; `limit=100` is a fixed
  per-stream `query` value (matching legacy's `defaultPageSize`) and pagination relies solely on the
  cursor paginator's own token-based stop signal (no `MaxPages` cap declared), matching legacy's own
  unbounded default. Out of scope for wave2 fan-out (Pass B).
- The 2-page conformance fixture lives on the `deals` stream (`fixtures/streams/deals/page_1.json`
  /`page_2.json`); `persons`/`organizations`/`activities`/`products`/`users` ship single-page
  fixtures only, since they share the identical `cursor`+`token_path` pagination shape already proven
  by `deals`'s fixture pair. The new `offset_limit`-paginated streams (`leads` and the field-metadata
  streams) and `none`-paginated streams likewise ship single-page fixtures only — `pagination_
  terminates` (the dynamic check requiring a 2-page proof) only ever exercises the bundle's
  first-declared eligible stream (`deals`), so proving each additional pagination shape a second time
  is not required by conformance, and a realistic single short page is the honest representation for
  every other stream.
- **Deals/Persons/Organizations/Products/Activities/Pipelines/Stages full CRUD (create/update/
  delete, and their v2-only list/get) is NOT modeled — Pipedrive moved this to API v2.** This
  bundle's `base_url` targets v1 throughout; the 6 original streams' v1 GET-list reads are
  unaffected (Pipedrive still serves them), but v1 no longer documents create/update/delete for these
  7 resources at all (verified against both the v1 OpenAPI document and the live v1 docs pages, which
  now redirect to `/api/v2/<resource>`). Adding v2-shaped writes against a v1-configured bundle would
  require a second `base_url`/auth surface (v2 uses the same `api_token` query auth, but a different
  host path prefix and different pagination/response envelope shape entirely — `data`-wrapped without
  the `additional_data.pagination` convention). This is a real, scoped-out capability gap, not a
  silent omission: a future pass could add a `v2_base_url` config property and a second `auth`/
  pagination convention specifically for these 7 resources' writes, but that is a deliberate
  architecture decision beyond a Pass B expansion's "add coverage, preserve existing behavior"
  mandate.
- **Pipelines/Stages have no v1-native top-level list endpoint at all** (only
  `/pipelines/{id}/deals`, `/pipelines/{id}/conversion_statistics`, `/pipelines/{id}/
  movement_statistics`, and `/stages/{id}/deals` sub-resources remain under v1) — since the
  Deal/Organization/Person/Activity/Product streams above are already out of v2 scope for this pass,
  these narrower per-pipeline/per-stage sub-resources are excluded alongside them (see
  `api_surface.json`).
- **Pipedrive's Projects add-on (projects/tasks/project boards/phases/templates) is entirely out of
  scope.** This is a distinct, not-universally-enabled paid feature area with its own
  `cursor`+`limit` pagination convention (different from every other resource in this bundle); Pass B
  prioritized the universally-available CRM/lead/configuration resources instead.
- **`organizationRelationships` requires a mandatory `org_id` query parameter** — it can only be
  listed per-organization, not as a standalone top-level collection. Modeling this would require the
  engine's `fan_out` dialect (one preliminary request per organization id, driven by the
  `organizations` stream), adding real complexity for a narrow, rarely-synced resource; deprioritized
  for this pass.
- **Schema-authoring mutations for deal/person/organization/product field DEFINITIONS (create/
  update/delete a custom field) are not modeled** — the corresponding `*_fields` streams in this
  bundle are read-only reference/reporting data (what fields exist and their metadata), not a
  business-object create/update/delete surface; authoring a NEW custom field is an account-schema
  change with materially different risk than an ordinary record mutation, out of scope for this pass.
- **`http_auth_password` is intentionally omitted from `schemas/webhooks.json`.** Pipedrive's
  webhook resource can carry HTTP Basic auth credentials for the receiving endpoint; while the field
  is not marked `x-secret` at the source (Pipedrive documents it as a plain response field), emitting
  a credential-shaped value into synced record data is the same category of risk `x-secret`
  discipline exists to prevent, so this bundle drops it from the projected schema rather than
  surface it verbatim.
