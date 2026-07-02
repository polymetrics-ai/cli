# Overview

SendOwl is a digital-product delivery and e-commerce platform. This bundle reads SendOwl orders,
products, and subscriptions through the SendOwl API. It is read-only, matching legacy
`internal/connectors/sendowl` exactly (`Capabilities{Write: false}`).

## Auth setup

Provide a SendOwl API key/secret pair via the `username` config value (API key) and the `password`
secret (API secret). They are sent as HTTP Basic credentials via `base.auth`'s `mode: basic`,
identical to legacy's `connsdk.Basic(username, password)`. The password is never logged.

## Streams notes

All 3 streams (`orders`, `products`, `subscriptions`) share the same shape: `GET` against the
SendOwl list endpoint, records at the top-level JSON array (`records.path: "."`), primary key
`["id"]`. `projection: passthrough` is used on every stream because legacy's own `Read` re-emits
each decoded record verbatim through `connsdk.Harvest` (which itself calls `emit(rec)` with no
field filtering — see `connsdk/paginate.go`'s `Harvest`) — schema projection alone would silently
drop any undeclared raw field, a data-parity regression. The declared `schemas/*.json` properties
are still a realistic, honest field set for `records_match_schema`'s type-checking of the fields
that ARE declared.

Only `orders` declares `x-cursor-field: created_at`, matching legacy's `CursorFields:
["created_at"]` (declared only on the `orders` stream in legacy's `streams()` catalog function);
`products` and `subscriptions` declare no cursor field, matching legacy exactly. No stream declares
an `incremental` block — legacy never applies a server-side incremental filter for any of these 3
streams.

Pagination is `page_number` (`page_param: page`, `size_param: per_page`, `start_page: 1`,
`page_size: 100`), identical to legacy's `connsdk.PageNumberPaginator{PageParam: "page", SizeParam:
"per_page", StartPage: 1, PageSize: pageSize}` at legacy's own default `pageSize = 100`. `page_size`
is a fixed value on `streams.json`'s `base.pagination` block, not a per-stream query template: the
`page_number` paginator's own `Start()`/`Next()` always sets the `per_page` query param itself
(`connsdk.PageNumberPaginator.pageQuery`), and the engine's query-merge (`mergeValues`) lets the
paginator's own params win over any same-keyed stream-level `query` entry — so a stream-level
`per_page` template would be silently dead code, never actually reaching the wire. See Known limits
for the config-surface consequence.

## Write actions & risks

None. SendOwl is exposed read-only, matching legacy's `Capabilities{Write: false}` and its `Write`
method, which always returns `connectors.ErrUnsupportedOperation`.

## Known limits

- Full SendOwl API surface (discounts, packages, license keys, affiliates, webhooks, etc.) is out
  of scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries. Only the 3 legacy-parity read streams are implemented.
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
  Legacy accepts a `page_size` config value (clamped 1-500,
  `positiveInt(..., 1, 500, "page_size")`) and threads it into its paginator's `PageSize`. The
  engine's `page_number` paginator bakes its size param directly into the request from
  `PaginationSpec.PageSize` (`streams.json`'s `base.pagination.page_size: 100`, matching legacy's
  own default) — a stream-level `query` entry for the same key (`per_page`) is unconditionally
  overridden by the paginator itself (see Streams notes), so declaring `page_size` in `spec.json`
  with no way to actually wire it through would be worse than an absent key (F6,
  `docs/migration/conventions.md`). This is a documented config-surface narrowing, never a
  data-parity change for any read at the default page size.
- **Fixtures are single-page, matching the `max_pages: 1` cap.** Because `max_pages: 1` is baked in
  (see above), a real read never issues a second request — a 2-page fixture would be unconsumed by
  design and would fail `conformance`'s `pagination_terminates` check (which asserts hits == the
  number of fixture pages). This mirrors the identical, already-accepted `searxng` precedent (also
  `max_pages: 1`, also single-page fixtures per stream) — see `docs/migration/conventions.md` §4.
  `page_number` pagination's mechanics are still exercised generically by `connsdk`'s own
  `PageNumberPaginator` unit tests outside this bundle.
