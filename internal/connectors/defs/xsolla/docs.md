# Overview

Xsolla is a read-only declarative bundle migrated from `internal/connectors/xsolla` (the
hand-written legacy connector, which stays registered and unchanged until wave6's registry flip).
It reads Xsolla merchant projects, and per-project orders and transactions, through the Xsolla
Merchant API v2.

## Auth setup

Provide the Xsolla merchant ID and merchant API key as the `merchant_id`/`api_key` secrets. Both
are sent as HTTP Basic auth (`merchant_id` as username, `api_key` as password), matching legacy's
`connsdk.Basic(merchantID, apiKey)`. Neither is ever logged.

## Streams notes

`projects` (`GET /projects`) has no path parameters. `orders` (`GET
/projects/{project_id}/orders`) and `transactions` (`GET /projects/{project_id}/transactions`) are
scoped to the `project_id` config value, substituted into the path (urlencoded by default per
`InterpolatePath`); `project_id` is required for those two streams (an unset value is a runtime
interpolation error, matching legacy's own "project_id is required for this stream" check). All
three streams share the identical shape: `GET`, records at `items`, primary key `["id"]`, cursor
field `updated_at`. Pagination is `page_number` (`page`/`limit` query params, `page_size: 100`,
1-based), matching legacy's `connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit",
StartPage: 1, PageSize: pageSize}` with legacy's default `pageSize` of 100.

## Write actions & risks

None — this connector is read-only (`capabilities.write: false`), matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation` unconditionally.

## Known limits

- Legacy's runtime-configurable `page_size` (bounded 1-500, default 100) and `max_pages` config
  keys are **not modeled** in this bundle's `spec.json`: `engine.PaginationSpec.PageSize`/`MaxPages`
  are plain (non-templated) JSON integers set once in `streams.json`, with no mechanism to bind a
  runtime `config.*` value into them (F6, `docs/migration/conventions.md` — a declared-but-unwireable
  spec property is worse than an absent one). `page_size` is fixed at the bundle level to legacy's
  own default (100), so behavior for any caller that never overrode legacy's default is unchanged;
  a caller that previously set a non-default `page_size`/`max_pages` loses that override. This
  mirrors the accepted `float`/`stripe` goldens' identical page_size-is-static precedent.
  `max_pages` is likewise not enforced (unbounded reads; `PaginationSpec.MaxPages` is left unset).
- Full Xsolla merchant API surface (promotions, pricing, webhooks, agent-order stats) is out of
  scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `fixtures/streams/projects/{page_1,page_2}.json` is the required 2-page pagination fixture
  (page 1 returns 100 records to trigger a next page per `page_number`'s short-page stop rule;
  page 2 returns 1 record and stops). `orders`/`transactions` ship single-page fixtures scoped to
  conformance's synthetic `project_id` value (`synthetic-conformance-value`).
