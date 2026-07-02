# Overview

Mantle (`api.heymantle.com`) is a billing/usage platform for Shopify apps. This bundle migrates
`internal/connectors/mantle` (the hand-written legacy connector) to a declarative Tier-1 bundle at
capability parity: 2 read-only streams (`customers`, `subscriptions`), no writes.

## Auth setup

Provide a Mantle API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

Both streams (`customers`, `subscriptions`) share the same shape: `GET` against the Mantle list
endpoint, records at their own top-level selector (`customers`/`subscriptions`), primary key
`["id"]`. Pagination follows Mantle's own `cursor`/`hasNextPage` convention
(`pagination.type: cursor` with `token_path: cursor` and `stop_path: hasNextPage`): the next page's
`cursor` query param is read from the response body's own `cursor` field, and pagination stops when
`hasNextPage` is falsy — matching legacy `harvest`'s
`hasNext != "true" || strings.TrimSpace(nextCursor) == ""` stop condition (the token-path
paginator's own absent/empty-token stop check covers the second half of that OR; `stop_path`
covers the first).

Every request sends `take=500` (matches legacy's default `page_size`) via each stream's static
`query: {"take": "500"}`; legacy's config-driven `page_size`/`max_pages` overrides are not
reproduced — the engine's `cursor` pagination type takes no page-size parameter for its own
stop-condition bookkeeping (unlike `page_number`, whose `PageSize` field doubles as a stop
threshold, `cursor`'s stop signal is entirely `token_path`/`stop_path`-driven), so this is a
static-`query`-value deviation only: the STOP condition itself is unaffected by whatever `take`
value is declared. `max_pages` has no engine equivalent exposed to this bundle beyond the
`PaginationSpec.MaxPages` static hard-cap field, which this bundle leaves unset (unbounded),
matching legacy's own default (`mantleMaxPages` returns 0/unbounded unless a caller sets
`max_pages`).

Neither stream declares an `incremental` block: legacy's `mantleStreams()` publishes
`CursorFields` (`updatedAt`/`createdAt`) for catalog/schema purposes, but `harvest` never sends any
server-side date filter — every read is a full pull, matching this bundle's omission of a
`request_param`/`start_config_key` on either stream exactly (`x-cursor-field` is still declared on
both schemas for downstream dedup/ordering, matching legacy's published `CursorFields`).

Legacy enforces no client-side rate limiting, so this bundle declares no `streams.json`
`base.rate_limit` either, matching that (lack of) behavior exactly.

## Write actions & risks

None. Mantle is a read-only source in both legacy and this bundle (`capabilities.write: false`) —
legacy's own `Write` method is an unconditional `ErrUnsupportedOperation` stub.

## Known limits

- `page_size`/`max_pages` are not exposed as config: `streams.json`'s static `query: {"take": "500"}`
  and the absence of a declared `PaginationSpec.MaxPages` reproduce legacy's DEFAULT values exactly,
  but neither is runtime-overridable in this bundle the way legacy's `mantlePageSize`/
  `mantleMaxPages` config parsing allowed. This never changes emitted record DATA for any input
  legacy itself would accept at its own defaults; it narrows configurability only.
- The full Mantle API surface (usage events, plans) is out of scope for this wave; see
  `api_surface.json`'s `excluded` entries.
