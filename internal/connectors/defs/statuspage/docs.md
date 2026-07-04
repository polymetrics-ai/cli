# Overview

Statuspage is a Tier-1 declarative-HTTP connector reading Statuspage pages, components, incidents,
subscribers, component groups, metrics, metrics providers, page access groups/users, and incident
templates through the Statuspage API (`https://api.statuspage.io/v1/...`). This bundle was Pass-B
full-surface expanded against the real published OpenAPI 3.0.0 specification embedded in
`https://developer.statuspage.io/`'s Redoc page (`__redoc_state.spec.data` in the page's inline
`<script>` — the site renders client-side from this JSON blob rather than publishing a
separately-fetchable spec URL) — see `api_surface.json` for the per-endpoint disposition of all
112 documented method+path pairs. It originally targeted capability parity with
`internal/connectors/statuspage` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Statuspage API key via the `api_key` secret; it is sent as the `Authorization` header
with an `OAuth ` prefix (`Authorization: OAuth <api_key>`), matching legacy's
`connsdk.APIKeyHeader("Authorization", key, "OAuth ")` (`statuspage.go:119`) and the real API's
documented `api_key` security scheme (`type: apiKey, in: header, name: Authorization`). Never
logged. `base_url` defaults to `https://api.statuspage.io/v1` and may be overridden for
tests/proxies.

## Streams notes

10 GET list streams. `pages` is a top-level `GET /pages` list with no page scoping; records are the
response body's top-level JSON array (`records.path: "."`). Every other stream is scoped to one
Statuspage page via the required `page_id` config value, substituted into each stream's
`/pages/{{ config.page_id }}/...` path template (urlencoded by `InterpolatePath`'s per-segment
default, matching legacy's own `url.PathEscape(pageID)` in `resolveResource`); an absent `page_id`
hard-errors on both sides (legacy: `"statuspage stream requires config page_id for path %q"`;
engine: an unresolved `config.page_id` path-template key). All streams' records are top-level JSON
arrays (`records.path: "."`), matching legacy's `recordsPath: "."` for the original 4 and the real
API's documented response shape for every new stream EXCEPT `metrics` (see below).

- `pages`, `components`, `incidents`, `subscribers` — the original 4 legacy-parity streams; schemas
  intentionally retain the minimal fields emitted by legacy's `copyRecord` mappers. Widening these
  schema-mode streams to the real API's full objects would emit fields legacy dropped.
- `component_groups` (`/component-groups`), `metrics_providers` (`/metrics_providers`),
  `page_access_groups` (`/page_access_groups`), `page_access_users` (`/page_access_users`),
  `incident_templates` (`/incident_templates`) — new Pass-B streams, all real top-level JSON array
  responses.
- `metrics` (`/metrics`) — new Pass-B stream. The real OpenAPI spec documents this endpoint's
  response schema as a bare `Metric` object (not an array), which is very likely a spec-authoring
  error: the endpoint's own summary/description is "Get a list of metrics", it accepts `page`/
  `per_page` query parameters exactly like every other paginated list endpoint, and every sibling
  `.../metrics_providers` collection uses the standard array-of-object shape. This bundle treats it
  as a real list endpoint (`records.path: "."`, `page_number` pagination inherited from `base`),
  matching the documented query parameters and description rather than the almost-certainly-wrong
  singular response schema.

`incidents` declares `incremental.cursor_field: created_at`, matching legacy's own
`CursorFields: []string{"created_at"}` declaration; neither this bundle nor legacy sends a
server-side lower-bound filter or performs client-side filtering for this stream (legacy's `Read`
performs no incremental filtering at all) — this bundle matches that exactly (no `request_param`/
`client_filtered` declared), not introducing new filtering under the guise of a migration. No other
stream declares a cursor field (full refresh only), matching both legacy and the absence of any
documented `updated_after`-style filter param on any other list endpoint.

Pagination is page-number (`pagination.type: page_number`, `page_param: page`, `size_param:
per_page`, `start_page: 1`, `page_size: 100`), identical to legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize:
pageSize}` with legacy's default `pageSize` of 100, and confirmed against the real API's documented
`page`/`per_page` query parameters on every list endpoint covered by a stream here.

## Write actions & risks

**None — blocked (`ENGINE_GAP`).** `capabilities.write` remains `false` and this bundle ships no
`writes.json`, unchanged from the pre-Pass-B state, but for a different reason now that the real API
surface has been reviewed: every Statuspage mutation endpoint (`POST/PATCH/PUT/DELETE` on
components, incidents, subscribers, component groups, metrics, metrics providers, page access
groups/users, incident templates, and more) requires the request body to be wrapped in a Rails-style
top-level params envelope keyed by the singular resource name — e.g. `POST
/pages/{page_id}/incidents` expects `{"incident": {"name": "...", "status": "..."}}`, not the
incident's fields sent flat. Confirmed against the real OpenAPI spec's request-body schemas for
every mutating endpoint reviewed (`postPagesPageIdIncidents`, `postPagesPageIdComponents`,
`postPagesPageIdSubscribers`, `postPagesPageIdComponentGroups`, etc. all wrap their payload under a
single named key).

The engine's declarative write dialect (`internal/connectors/engine/write.go`) has no mechanism to
express this: `body_type: "json"`'s `buildJSONBody` (default body construction) and `body_type:
"none"`'s `buildBodyFieldsPayload` (allow-list construction) both always produce a FLAT top-level
JSON object from the record's fields — there is no `writes.json` field to additionally nest that
object under a fixed literal key. Sending a flat body to the real API would not silently succeed
with wrong data (which would be a silent parity break); Statuspage's Rails backend would reject an
un-enveloped body as a validation/parameter error, so this is a correctness blocker, not a stylistic
gap.

Per this pass's constraints, no new hook package may be created for a bundle (statuspage has none
today), so a `WriteHook`-based single-line envelope wrap — the natural Tier-2 fix — is not available
in this pass. **This is filed as an `ENGINE_GAP` blocker**: the two viable fixes are (1) a Tier-1
dialect extension adding an optional `writes.json` per-action field (e.g. `body_envelope: "incident"`)
that wraps the constructed body under that key before sending, which would need no hook and would
very likely unblock every other Rails-conventioned API in this migration program hitting the same
shape (recurs across every mutation endpoint on this API alone — 40+ instances — meeting the §6
recurrence threshold on its own), or (2) an orchestrator-approved new `hooks/statuspage/` package
implementing `WriteHook` to wrap the body. Every mutation endpoint's `api_surface.json` entry is
`excluded: {category: "out_of_scope", reason: "ENGINE_GAP ..."}`, naming this exact blocker.

## Known limits

- **All mutations are blocked by the params-envelope `ENGINE_GAP`** — see Write actions & risks
  above. This supersedes the prior "Pass B capability expansion" placeholder reason; the real reason
  is now documented per-endpoint in `api_surface.json`.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`statuspage.go`'s `pageSize`/`maxPages`, bounded 1-100 / a non-negative integer or
  `all`/`unlimited`). The engine's `page_number` paginator has no config-driven page-size or
  max-pages knob (`PaginationSpec.PageSize`/`MaxPages` are static bundle JSON, never templated), so
  this bundle uses legacy's own default (`page_size: 100`) as a fixed bundle value and does not
  declare `page_size`/`max_pages` in `spec.json` at all (a declared-but-unwireable config key is
  worse than an absent one, per `docs/migration/conventions.md` F6). Pagination is unbounded by
  default (reads every page until a short page), matching legacy's own default of `maxPages=0`
  (unbounded) when `max_pages` is unset.
- **Single-record GET endpoints are excluded as `duplicate_of`** their already-covered list stream
  (e.g. `GET /pages/{page_id}/incidents/{incident_id}` duplicates the `incidents` stream) rather than
  modeled as separate streams — see `api_surface.json`.
- **Uptime analytics, subscriber count/histogram aggregates, postmortem document sub-resources, and
  the embeddable status-widget config are out of scope** (`out_of_scope`/`non_data_endpoint`) — none
  are catalog/config data the connector's stream model targets.
- **Organization-level user/permission management is excluded as `requires_elevated_scope`** — it
  needs an organization-admin-scoped API key, a different credential class than the page-scoped
  Console API key `spec.json` documents.
- **Bulk metric-datapoint ingestion is excluded as `binary_payload`** — it accepts a timeseries
  array payload, not a single catalog record.
- **Subscriber notification/postmortem-publish actions are excluded as `destructive_admin`** —
  even once the params-envelope gap is closed, these trigger real external notifications
  (emails/SMS/tweets to subscribers) rather than being plain record CRUD.
- Full API-surface disposition (every one of the 112 documented Statuspage API v1 method+path pairs)
  is recorded in `api_surface.json`.
