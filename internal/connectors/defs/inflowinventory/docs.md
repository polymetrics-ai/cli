# Overview

inFlow Inventory is a read-only declarative-HTTP migration (wave2 fan-out) of
`internal/connectors/inflowinventory` (legacy Go package `inflowinventory`). It reads inFlow
products, customers, vendors, sales orders, and categories through the inFlow cloud REST API
(`https://cloudapi.inflowinventory.com`). This bundle targets capability parity with the legacy
connector; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an inFlow API key via the `api_key` secret and the inFlow company id via the required
`companyid` config value (inFlow embeds it as the first path segment of every resource request,
e.g. `/<companyid>/products`). `api_key` is sent as the RAW value of the `Authorization` header
(no `Bearer` prefix — inFlow's own convention, matching legacy's `connsdk.APIKeyHeader
("Authorization", secret, "")`), and every request also carries the static
`Accept: application/json;version=2024-03-12` header legacy sends. `base_url` defaults to
`https://cloudapi.inflowinventory.com` and may be overridden for tests or proxies.

## Streams notes

All 5 streams (`products`, `customers`, `vendors`, `sales_orders`, `categories`) share the same
shape: `GET /<companyid>/<resource>`, a bare top-level JSON array response (`records.path: ""`),
and inFlow's `count`/`after` cursor pagination (`pagination.type: cursor` with
`last_record_field` set to that stream's own id field — `productId`/`customerId`/`vendorId`/
`salesOrderId`/`categoryId` — and `cursor_param: after`), matching legacy's `harvest` loop: the
next page's `after` value is the id of the LAST record on the current page. Every request sends
`count` (default `100`, configurable via the `page_size` spec property, matching legacy's
`inflowDefaultPageSize`/`inflowMaxPageSize` bounds of 1-100).

Legacy's stop condition is "a page shorter than the requested count, OR no advanceable last id"
(`len(records) < pageSize || lastID == ""`); the engine's `cursor`+`last_record_field` paginator
(no `stop_path` declared, matching the `agilecrm` golden shape) stops only on a fully EMPTY page or
an absent/blank last-record id. This means a stream whose true final page happens to be exactly
`page_size` records long issues one additional (empty-result) request before stopping — an extra
round-trip, never a change to which records are emitted, so it is not a parity deviation under the
ledger's meta-rule (no emitted-record DATA changes for any input legacy would accept).

Primary key is each stream's own id field. `products` declares `x-cursor-field:
lastModifiedDateTime` for manifest-surface parity with legacy's `CursorFields`; legacy never
actually filters or advances any inFlow stream by a server-side incremental parameter (the API has
no such filter), so no `incremental` block is declared on any stream here either, matching legacy's
own full-refresh-only behavior. `customers`/`vendors`/`sales_orders`/`categories` have no cursor
field, matching legacy's `CursorFields` being unset (nil) for those four streams.

## Write actions & risks

None. inFlow Inventory is read-only in this bundle (`capabilities.write: false`); legacy also
rejects every write with `connectors.ErrUnsupportedOperation`. No `writes.json` is shipped.

## Known limits

- Full inFlow API surface (purchase orders, stock adjustments, locations, taxes, teams, webhooks)
  is out of scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope,
  reason: "Pass B capability expansion"}` entries. Only the 5 legacy-parity streams are
  implemented.
- `companyid`'s legacy validation (rejecting `/`, `?`, `#`, and `..`) is not separately re-declared
  in `spec.json`; the engine's own path-interpolation guard (urlencoded-by-default per-segment
  insertion, plus the `..`-traversal rejection documented in `docs/migration/conventions.md` §3)
  provides an equivalent safety net for this path-embedded config value.
- `page_size`'s legacy validation range (1-100) is not enforced by this bundle's `spec.json` (no
  numeric bounds check is wired for a string-typed config field here); an out-of-range value is
  passed through to the API as-is rather than rejected client-side the way legacy's
  `inflowPageSize()` helper does. This is a narrowing of legacy's input validation, not a change to
  any successfully-processed request's shape.
- `max_pages` is not declared in `spec.json`. `PaginationSpec.MaxPages` (the engine's only
  request-count cap) is a static JSON integer with no `{{ }}` template support, so a runtime
  `config.max_pages` value can never actually bound anything — declaring the config property
  anyway would be dead config a bundle author cannot wire to any real behavior (F6,
  `docs/migration/conventions.md`). Legacy's own `max_pages` config (0/`all`/`unlimited` = no cap)
  is reproduced here as the engine's default unbounded behavior (`MaxPages` omitted).
