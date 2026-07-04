# Overview

BlueTally is a wave2 fan-out declarative-HTTP migration, expanded to full API-surface coverage in
Pass B. It reads every documented BlueTally IT asset management resource (assets, employees,
licenses, maintenances, accessories, components, consumables, categories, departments,
depreciations, locations, manufacturers, products, statuses, suppliers, audits, activity, tenants)
through the BlueTally REST API (`GET https://app.bluetallyapp.com/api/v1/<resource>`). This bundle
is migrated from `internal/connectors/bluetally` (the hand-written connector); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a BlueTally API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(secret)`
(`bluetally.go:275`). Never logged. `base_url` defaults to `https://app.bluetallyapp.com` and may
be overridden for tests/proxies.

## Streams notes

All eighteen streams (`assets`, `employees`, `licenses`, `maintenances`, `accessories`,
`components`, `consumables`, `categories`, `departments`, `depreciations`, `locations`,
`manufacturers`, `products`, `statuses`, `suppliers`, `audits`, `activity`, `tenants`) read
`/api/v1/<resource>` list endpoints, each returning a bare JSON array at the response root
(`records.path: ""`), matching legacy's root-path `connsdk.RecordsAt(resp.Body, "")` convention,
except `tenants` (see below). Pagination is `offset_limit` (`limit`/`offset` query params,
`page_size: 50` — legacy's own `bluetallyDefaultPageSize`) for every paginated stream; the next
page's `offset` advances by the page size and the engine stops on a short/empty page, identical to
legacy's own `len(records) < pageSize` stop rule and `offset = page * pageSize` request
construction. `tenants` overrides `pagination: none` (`GET /api/v1/tenants` returns a single
`{"tenants": [...]}` object, not a paginated list — `records.path: "tenants"` selects the nested
array) since BlueTally's own reference documents this endpoint with no `limit`/`offset`/`sort`
parameters at all (multi-tenancy accounts only; the array is expected to stay small).

The original 5 streams (`assets`, `employees`, `licenses`, `maintenances`, `accessories`) declare
`updated_at` as `x-cursor-field` for manifest-surface parity with legacy's
`CursorFields: []string{"updated_at"}`, but — like legacy — none of them actually filter
server-side: legacy's `harvest` never sends any lower-bound query parameter (BlueTally's list API
supports only offset/limit pagination, no time-based filter), so every read is a full sync
regardless of a cursor's value. This bundle declares no `incremental` block on any of those
streams, matching legacy's behavior exactly. The 13 new Pass B streams follow the identical
shape (`updated_at`/`created_at` present on every resource per the connector's published OpenAPI
spec) except `activity` (cursor field `timestamp`, the only real time-ordered field on that
resource) and `tenants` (no timestamp field at all, so no `x-cursor-field`).

## Write actions & risks

None. `capabilities.write` is `false` and this bundle ships no `writes.json`. BlueTally's full
create/update/delete/check-in/check-out surface (`api_surface.json`) was researched during this
Pass B pass and found to be **entirely query-string-parameterized** on every mutation endpoint
(verified directly against the connector's published OpenAPI spec, embedded at
`developer.bluetally.com/branches/1.0/apis/bluetally-api.json`: every POST/PUT operation declares
its parameters `"in": "query"`, none declare a `requestBody`). The engine's write dialect
(`engine/bundle.go` `WriteAction.BodyType`: `json`/`form`/`none`) only constructs a request BODY —
there is no mechanism to place write-action fields on the query string — so every BlueTally
mutation is an `ENGINE_GAP` blocker (`docs/migration/conventions.md` §6): a `body_type: json` or
`form` action would silently send an empty or wrong-shaped request while the real parameters
BlueTally expects sit unset in the query string, diverging from the documented contract rather
than reproducing it. This is reported, not worked around; no `writes.json` is shipped and
`capabilities.write` stays `false`. Should the engine gain a query-string write-body mechanism in
a future mini-wave, every BlueTally resource's create/update/delete (and the 9 check-in/check-out
actions) can be revisited from `api_surface.json`'s per-endpoint `ENGINE_GAP` reasons.

## Known limits

- **Every BlueTally mutation endpoint is an ENGINE_GAP, not a scope choice.** See "Write actions &
  risks" above; `api_surface.json` documents each excluded POST/PUT/DELETE endpoint individually.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`
  (1-100, default 50) and `max_pages` (0/all/unlimited default) as config-driven overrides
  (`bluetallyPageSize`/`bluetallyMaxPages`, `bluetally.go:314-342`). The engine's `offset_limit`
  paginator's `page_size` is a fixed value baked into `streams.json`'s `base.pagination` block at
  bundle-author time (`PaginationSpec.PageSize` is a plain int, never `config.*`-templated), and
  `MaxPages` is likewise a fixed bundle-time int — matching the identical, already-documented
  searxng/bitly precedent (`docs/migration/conventions.md`). This bundle bakes in legacy's own
  default (`page_size: 50`); `max_pages` is left unbounded (0/omitted), matching legacy's own
  default (empty `max_pages` config = unlimited). Neither is declared in `spec.json` (F6,
  REVIEW.md: a declared-but-unwireable config key is worse than an absent one).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`) stamps a `previous_cursor` field (echoing
  `req.State["cursor"]` when set) onto every fixture-mode record (`bluetally.go:243-247`), and
  synthesizes a superset of fields shared loosely across all five original streams rather than
  each stream's own real wire shape. This is not part of the LIVE record shape; this bundle's
  schemas and fixtures instead use each stream's real per-resource wire shape, matching the
  connector's published OpenAPI spec (original 5 streams also match each stream's own
  `bluetally*Record`/`bluetally*Fields` mapping in `internal/connectors/bluetally/streams.go`).
- **`GET /{resource}/{id}` single-record detail endpoints are not modeled as separate streams.**
  Every resource's list endpoint (already a stream) returns the identical fields as its per-id
  detail endpoint; the detail endpoint is `duplicate_of` surface for a full-sync connector, not a
  distinct capability (`api_surface.json`).
- **Nested sub-objects (`custom_fields`, `deployed_to`, `seats`, `checked_out_to*`, id-array
  cross-references like `assets`/`licenses`/`products` on parent resources) are typed
  `["array", "null"]`/passed through as raw JSON in schemas rather than flattened into
  sub-streams.** These are genuinely nested, variable-shape data (a `custom_fields` entry is an
  arbitrary `{field_name: value}` map, not a fixed schema) with no natural single-parent-key
  fan-out shape distinct from the resource that already carries them; flattening them further is
  out of scope for this pass.
