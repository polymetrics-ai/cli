# Overview

ServiceNow is an IT service management platform. This bundle reads ServiceNow incident, user, and
group table rows through the ServiceNow Table API. It is read-only, matching legacy
`internal/connectors/service-now` exactly (`Capabilities{Write: false}`).

## Auth setup

Provide `base_url` (e.g. `https://your-instance.service-now.com`; required — legacy has no derived
default and errors when unset), a ServiceNow `username`, and a `password` secret. Credentials are
sent as HTTP Basic auth via `base.auth`'s `mode: basic`, identical to legacy's
`connsdk.Basic(username, password)`. `password` is never logged.

## Streams notes

All 3 streams (`incidents`, `users`, `groups`) share the same shape: `GET
/api/now/table/<table>`, records at the response body's `result` array (`records.path: "result"`),
primary key `["sys_id"]`, `x-cursor-field: updated_on` (matching legacy's `CursorFields:
["updated_on"]` on every stream). `projection: passthrough` is used because legacy's own `Read`
re-emits each decoded record verbatim through `connsdk.Harvest` (which itself calls `emit(rec)`
with no field filtering) — schema projection alone would silently drop any undeclared raw field, a
data-parity regression. The declared `schemas/*.json` properties are still a realistic, honest
field set for `records_match_schema`'s type-checking of the fields that ARE declared.

No stream declares an `incremental` block: legacy's own catalog declares `CursorFields:
["updated_on"]` for sync-mode eligibility, but its `Read` never actually sends a server-side
`sysparm_query`-style incremental filter for any of these 3 tables — matching that exactly, this
bundle declares `x-cursor-field: updated_on` on every schema (for `incremental_append_deduped`
eligibility) but no stream-level `incremental` block (no server-side filter param sent). Note
ServiceNow's `updated_on` wire format is `"2026-01-01 00:00:00"` (space separator, no `T`/timezone)
— not RFC3339 — matching legacy's own `fixtureUpdatedAt` constant verbatim; this bundle's schema
types the field as a plain string and does not attempt to reformat/validate it as a timestamp.

Pagination is `offset_limit` (`limit_param: sysparm_limit`, `offset_param: sysparm_offset`,
`page_size: 100`), identical to legacy's `connsdk.OffsetPaginator{LimitParam: "sysparm_limit",
OffsetParam: "sysparm_offset", PageSize: pageSize}` at legacy's own default `pageSize = 100`.
`page_size` is a fixed value on `streams.json`'s `base.pagination` block, not a per-stream query
template: the `offset_limit` paginator's own `Start()`/`Next()` always sets the `sysparm_limit`
query param itself, and the engine's query-merge (`mergeValues`) lets the paginator's own params
win over any same-keyed stream-level `query` entry — so a stream-level `sysparm_limit` template
would be silently dead code, never actually reaching the wire. See Known limits for the
config-surface consequence.

## Write actions & risks

None. ServiceNow is exposed read-only, matching legacy's `Capabilities{Write: false}` and its
`Write` method, which always returns `connectors.ErrUnsupportedOperation`.

## Known limits

- Full ServiceNow Table API surface (arbitrary custom tables, attachments, `sys_user_grmember`
  membership rows, single-record GET by `sys_id`, the Aggregate API, import sets, etc.) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 3 legacy-parity read streams are implemented.
- **`docs_url` is `"manual intervention needed"`.** This wave's connector manifest supplied no
  reachable ServiceNow API documentation URL for this connector. This did not block migration:
  legacy's own Go source (`internal/connectors/service-now/service_now.go`, read-only ground truth
  per `docs/migration/conventions.md`) is a complete, sufficient behavioral spec for a 3-table
  read-only Table API connector, and this bundle is derived directly from it rather than from
  ServiceNow's own docs. A reachable `docs_url` should be sourced and back-filled in a follow-up
  pass (ServiceNow's public Table API reference is normally at
  `https://docs.servicenow.com/bundle/.../table-api`, but no specific reachable link was verified
  for this migration).
- **`max_pages` is not a runtime config override.** Legacy accepts a `max_pages` config value
  (default `1`, meaning legacy itself only fetches a single page unless explicitly overridden) and
  threads it through to `connsdk.Harvest`'s page-count cap. The engine's declarative pagination has
  no per-read config-driven override mechanism for `PaginationSpec.MaxPages` — it is a fixed value
  baked into `streams.json`'s `base.pagination` block (the same "no runtime override" limitation
  `docs/migration/conventions.md` documents for searxng's `page_size`/`max_pages`). This bundle
  bakes in `max_pages: 1`, matching legacy's own default exactly; a caller wanting more pages cannot
  express it here. This is a documented config-surface narrowing, never a data-parity change for
  any read actually exercised at the default.
- **`page_size` is not a runtime config override either, and is not declared in `spec.json`.**
  Legacy accepts a `page_size` config value (clamped 1-10000,
  `positiveInt(..., 1, 10000, "page_size")`) and threads it into its paginator's `PageSize`. The
  engine's `offset_limit` paginator bakes its size param directly into the request from
  `PaginationSpec.PageSize` (`streams.json`'s `base.pagination.page_size: 100`, matching legacy's
  own default) — a stream-level `query` entry for the same key (`sysparm_limit`) is unconditionally
  overridden by the paginator itself (see Streams notes), so declaring `page_size` in `spec.json`
  with no way to actually wire it through would be worse than an absent key (F6,
  `docs/migration/conventions.md`). This is a documented config-surface narrowing, never a
  data-parity change for any read at the default page size.
- **Fixtures are single-page, matching the `max_pages: 1` cap.** Because `max_pages: 1` is baked in,
  a real read never issues a second request — a 2-page fixture would be unconsumed by design and
  would fail `conformance`'s `pagination_terminates` check (which asserts hits == the number of
  fixture pages). This mirrors the identical, already-accepted `searxng`/`sendowl` precedent — see
  `docs/migration/conventions.md` §4. `offset_limit` pagination's mechanics are still exercised
  generically by `connsdk`'s own `OffsetPaginator` unit tests outside this bundle.
