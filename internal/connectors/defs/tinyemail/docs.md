# Overview

tinyEmail reads tinyEmail subscribers, lists, and campaigns, and writes a subscriber create/upsert
action, through the tinyEmail API. This bundle originated as a capability-parity migration from
`internal/connectors/tinyemail` (the hand-written connector it migrates; the legacy package stays
registered and unchanged until wave6's registry flip) and was then expanded (Pass B) to tinyEmail's
own published API reference. That reference (docs-api.tinyemail.com, a readme.io-hosted site)
declares EXACTLY ONE endpoint in the entire product: `POST /segment/customer` — now implemented as
`create_subscriber`. tinyEmail's real public API is genuinely this narrow (Enterprise-only,
request-access, marketing-automation-import-focused), not under-researched here; see
`api_surface.json`'s scope note for the full sourcing detail, including why the pre-existing
`subscribers`/`lists`/`campaigns` GET streams (undocumented in tinyEmail's own reference, but kept
unchanged for parity) coexist with the one genuinely documented endpoint.

## Auth setup

Provide a tinyEmail API key via the `api_key` secret; it is sent as the `X-API-Key` request
header with no prefix, matching legacy's `connsdk.APIKeyHeader("X-API-Key", key, "")`
(`tinyemail.go:107`) exactly, and is never logged. `base_url` defaults to
`https://api.tinyemail.com/v1`, matching legacy's `defaultBaseURL` fallback.

## Streams notes

All 3 streams (`subscribers`, `lists`, `campaigns`) share `page_number` pagination: `page` +
`limit` query params, `page_size: 100` (legacy's own `defaultPageSize`), records at the `data`
envelope key, primary key `["id"]`. A page shorter than 100 records signals the last page,
matching legacy's `connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage:
1, PageSize: size}` contract exactly (`tinyemail.go:87-88`). The `check` request mirrors legacy's
own check call (`GET /subscribers?page=1&limit=1`, `tinyemail.go:47`).

Every stream uses `projection: "passthrough"`: legacy's `Read` forwards each `connsdk.Harvest`
record verbatim via `emit(connectors.Record(rec))` with no field-building step
(`tinyemail.go:88`), so schema-mode projection would silently drop any API field not in the
declared list. The schemas document legacy's own known field surface (legacy's `fields(...)`
declarations, `tinyemail.go:116-118`) but do not constrain what is actually emitted.

## Write actions & risks

- **`create_subscriber`** (`POST /segment/customer`) creates or upserts a subscriber (customer)
  record into the caller's tinyEmail account, optionally assigning it to a named audience segment
  (`segmentName`) and setting its subscription `status` (`Subscribed`/`Unsubscribed`). Uses the
  same `X-API-Key` auth as every read stream. Low-risk external mutation (a marketing-list
  import/upsert, not a destructive action); no approval required beyond the standard reverse-ETL
  plan-approval gate.

**Deviation**: tinyEmail's own reference documents `segmentName` as a QUERY parameter on this
endpoint, not a body field; the engine's write dialect has no declarative query-parameter
mechanism for write actions (only `path`/`body` — see `internal/connectors/engine/bundle.go`'s
`WriteAction`, which has no `query` field at all, unlike `stream.Query` for reads). `segmentName`
is sent as an ordinary JSON body field instead, alongside every other documented body parameter.
This is the only way to express it declaratively at all; if tinyEmail's real endpoint strictly
requires it as a query string parameter and rejects/ignores a body-provided value, `segmentName`
assignment would silently fail (the customer would still be imported, just outside the intended
segment) — flagged here rather than assumed to work, since it could not be verified against a
live account. Every other documented field is a body-only parameter with no such ambiguity.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable per the engine dialect.** Legacy exposes
  both as config-driven overrides (`pageSize`/`maxPages`, `tinyemail.go:166-187`). The engine's
  `page_number` paginator's `page_size` is a fixed value baked into `streams.json`'s
  `base.pagination` block, and `MaxPages` has no per-request config-driven override mechanism
  either (conventions.md §3/§"read-only, no-auth variant": no runtime config-driven page-size or
  max-pages override mechanism exists for any pagination type). This bundle bakes in legacy's own
  default (`page_size: 100`, unbounded `max_pages`), matching legacy's out-of-the-box behavior for
  every caller that never overrode either value; a caller that previously set a non-default
  `page_size`/`max_pages` would see a documented, out-of-scope config surface narrowing here.
  Neither property is declared in `spec.json` since no template in this bundle consumes it
  (conventions.md F6: a declared-but-unwireable key is worse than an absent one).
- **`subscribers`/`lists`/`campaigns` GET streams are undocumented in tinyEmail's own current API
  reference.** They are kept exactly as this bundle originally migrated them (parity-locked); see
  `api_surface.json`'s scope note for the full sourcing discussion. No further GET resources were
  added beyond these 3, since tinyEmail's own reference documents no GET endpoint at all.
- **The `segmentName` write parameter's query-vs-body placement is unverified** — see the
  deviation note under "Write actions & risks" above.
