# Overview

tinyEmail is a wave2 fan-out declarative-HTTP migration. It reads tinyEmail subscribers, lists,
and campaigns through the tinyEmail REST API v1 (`GET https://api.tinyemail.com/v1/<resource>`).
This bundle is capability-parity migrated from `internal/connectors/tinyemail` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

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

## Write actions & risks

None. tinyEmail is read-only (`capabilities.write: false`, no `writes.json`), matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

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
