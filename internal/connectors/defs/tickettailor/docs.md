# Overview

Ticket Tailor is a wave2 fan-out declarative-HTTP migration. It reads Ticket Tailor events,
orders, and issued tickets through the Ticket Tailor REST API v1
(`GET https://api.tickettailor.com/v1/<resource>`). This bundle is capability-parity migrated
from `internal/connectors/tickettailor` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Ticket Tailor API key via the `api_key` secret; it is sent as the username of HTTP
Basic auth with an empty password (`Authorization: Basic base64(<api_key>:)`), matching legacy's
`connsdk.Basic(key, "")` (`tickettailor.go:107`) exactly, and is never logged. `base_url` defaults
to `https://api.tickettailor.com/v1`, matching legacy's `defaultBaseURL` fallback.

## Streams notes

All 3 streams (`events`, `orders`, `issued_tickets`) share `page_number` pagination: `page` +
`limit` query params, `page_size: 100` (legacy's own `defaultPageSize`), records at the `data`
envelope key, primary key `["id"]`. A page shorter than 100 records signals the last page,
matching legacy's `connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage:
1, PageSize: size}` contract exactly (`tickettailor.go:87-88`). The `check` request mirrors
legacy's own check call (`GET /events?page=1&limit=1`, `tickettailor.go:47`).

All 3 streams declare `"projection": "passthrough"` (conventions.md §8 rule 1). Legacy's `Read`
calls `connsdk.Harvest(..., func(rec connsdk.Record) error { return emit(connectors.Record(rec)) })`
(`tickettailor.go:88`) — every record `connsdk.Harvest` extracts from the raw `data` envelope is
emitted completely unfiltered; `streamSpecs[...].fields` (the 5-ish field list per stream, e.g.
`id`/`name`/`start_date`/`end_date`/`status` for `events`) is Catalog-only decoration
(`tickettailor.go:116-129`, consumed solely by `Catalog()`'s `connectors.Stream` construction),
never applied to the emitted record itself. Default `"schema"` projection mode would silently drop
every real Ticket Tailor field not named in each stream's declared schema properties (the live API
returns considerably more per-object detail than legacy's catalog list names, e.g. `events`'
`description`/`venue`/`currency`/`timezone`/`ticket_types`/`url`), an undocumented silent
data-shape change relative to legacy's raw passthrough. Each schema still declares the fields
legacy's own catalog names (for `x-primary-key` typing and `records_match_schema` coverage), but
passthrough mode means any other real field Ticket Tailor returns still survives unfiltered,
matching legacy exactly.

## Write actions & risks

None. Ticket Tailor is read-only (`capabilities.write: false`, no `writes.json`), matching
legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable per the engine dialect.** Legacy exposes
  both as config-driven overrides (`pageSize`/`maxPages`, `tickettailor.go:166-187`). The engine's
  `page_number` paginator's `page_size` is a fixed value baked into `streams.json`'s
  `base.pagination` block, and `MaxPages` has no per-request config-driven override mechanism
  either (conventions.md §3/§"read-only, no-auth variant": no runtime config-driven page-size or
  max-pages override mechanism exists for any pagination type). This bundle bakes in legacy's own
  default (`page_size: 100`, unbounded `max_pages`), matching legacy's out-of-the-box behavior for
  every caller that never overrode either value; a caller that previously set a non-default
  `page_size`/`max_pages` would see a documented, out-of-scope config surface narrowing here.
  Neither property is declared in `spec.json` since no template in this bundle consumes it
  (conventions.md F6: a declared-but-unwireable key is worse than an absent one).
