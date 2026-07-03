# Overview

Aviationstack is a read-only declarative-HTTP bundle migrated from
`internal/connectors/aviationstack` (the hand-written legacy connector, which stays registered and
unchanged until wave6's registry flip). It reads flight records and aviation reference data
(airlines, airports, airplanes, countries) through the aviationstack REST API.

## Auth setup

Provide an aviationstack API access key via the `access_key` secret. It is sent as the
`access_key` query parameter (`{"mode": "api_key_query", "param": "access_key", "value": "{{
secrets.access_key }}"}`), matching legacy's `connsdk.APIKeyQuery("access_key", key)` exactly.

## Streams notes

All 5 streams (`flights`, `airlines`, `airports`, `airplanes`, `countries`) share the identical
`{pagination:{...}, data:[...]}` envelope and `offset_limit` pagination (`limit`/`offset` query
params, short-page stop) — matching legacy's `harvest` loop, which advanced `offset` by
`page_size` until a short page (or the reported `pagination.total`) was reached. `flights` is the
only nested-object stream: its raw `departure`/`arrival`/`airline`/`flight` sub-objects are
flattened into top-level fields (`departure_airport`, `airline_name`, `flight_iata`, etc.) via
`computed_fields` bare `{{ record.<nested.path> }}` references — a computed field whose source
path is absent (nil parent) on a given record is silently skipped for that record, matching
legacy's own `nestedField` nil-safe helper. The 4 reference-data streams
(`airlines`/`airports`/`airplanes`/`countries`) need no `computed_fields` at all: every kept field
name already matches the raw API key, so plain schema projection suffices.

`flights` has a composite primary key (`["flight_date", "flight_iata"]`, no singular `id` field at
all) and cursors on `flight_date`; the 4 reference-data streams key on a stable string `id` and
declare no incremental cursor (legacy exposes no cursor field for them either).

Legacy's `page_size` (config-overridable, default 100, max 100) and `max_pages` config values are
**not wired into this bundle's `spec.json`**: the engine's `offset_limit` paginator
(`PaginationSpec.PageSize`/`MaxPages`) reads a static JSON int declared once in `streams.json`'s
`base.pagination` block, with no per-request template/config-driven override mechanism at all
(confirmed by `internal/connectors/engine/parity_searxng_test.go`'s own comment: "`PaginationSpec.
MaxPages` is a static int with no template support") — this is the exact same engine-shape gap
searxng's golden already documented and resolved by simply not declaring the dead config (F6,
`docs/migration/conventions.md` searxng worked example). `streams.json`'s `base.pagination.page_size`
is set to legacy's real production default, `100` (legacy: `defaultPageSize = 100`,
`maxPageSize = 100`) — this is the actual value a live deployment's paginator sends; it is not a
fixture convenience. All 5 streams, including `airlines`, use this same base pagination block
end-to-end (no stream-level override) — `airlines` previously declared a stream-level
`page_size: 2` override that leaked a fixture-sized page size into live config; that override has
been removed so `airlines` reads legacy's true 100-record page size like every other stream. Its
2-page conformance fixture (`fixtures/streams/airlines/{page_1,page_2}.json`) is sized to match:
page 1 returns a full 100-record page (so the paginator continues), page 2 returns the remainder.

## Write actions & risks

None. Aviationstack is a read-only data source in both legacy and this bundle
(`capabilities.write: false`, no `writes.json`).

## Known limits

- `page_size`/`max_pages` config-driven overrides are not implemented (see Streams notes above) —
  an `ENGINE_GAP`-class limitation of the `offset_limit` paginator's static (non-templated)
  `PageSize`/`MaxPages` fields, not a scope-narrowing choice specific to this connector. A live
  read always requests pages of the bundle's fixed page size; there is no way to request larger or
  smaller pages, or a hard page-count cap, without an engine change.
- Only the 5 legacy-parity reference/flight streams are implemented; premium-tier endpoints
  (`/v1/timetable`, `/v1/historical`) and additional reference resources (`/v1/routes`,
  `/v1/cities`, `/v1/taxes`) are out of scope — see `api_surface.json`'s `excluded` entries.
- **Check request omits legacy's `limit=1` query param (`ENGINE_GAP`).** Legacy's `Check` sends
  `GET /countries?limit=1` (a bounded read of the small countries reference list, used only to
  confirm auth/connectivity without pulling a full page). This bundle's `streams.json` `base.check`
  is `{"method": "GET", "path": "/countries"}` with no `limit` param, because the engine's check
  dispatch (`internal/connectors/engine/bundle.go`'s `RequestSpec`, used only by `check`) is a
  method+path descriptor with **no query-parameter field at all** —
  `internal/connectors/engine/read.go`'s `Check()` calls `rt.Requester.Do(ctx, method, checkPath,
  nil, nil)`, always passing a `nil` query. No bundle in this repo declares a check query param
  (stripe's check is `{"method": "GET", "path": "/customers"}`, searxng's is `{"method": "GET",
  "path": "/search"}`) because the dialect has nowhere to put one; adding this would require
  extending `RequestSpec` with a `query` field and threading it through `Check()` — a genuine
  engine gap, not a per-connector fix. Impact is minimal and non-data-emitting: the check still
  hits the identical `/countries` endpoint and proves auth/connectivity the same way; the only
  difference is legacy's check pulls at most 1 record instead of a full page.
