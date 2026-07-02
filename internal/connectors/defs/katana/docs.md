# Overview

Katana reads Katana MRP (Cloud Inventory) products, materials, variants, sales orders, and
customers through the Katana REST API (`https://api.katanamrp.com/v1`). This bundle migrates
`internal/connectors/katana` (the hand-written connector) to a declarative defs bundle at capability
parity; the legacy package stays registered and unchanged until wave6's registry flip. Katana is a
read-only source (the upstream API supports full sync only for pm), so `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Auth setup

Provide a Katana API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(secret)`
wiring exactly. `base_url` defaults to `https://api.katanamrp.com/v1` (legacy's
`katanaDefaultBaseURL`) and may be overridden for tests or proxies.

## Streams notes

All 5 streams (`products`, `materials`, `variants`, `sales_orders`, `customers`) share the identical
shape: `GET` against the Katana list endpoint, records at the top-level `data` array, primary key
`["id"]`. Pagination is `page_number` (`page_param: page`, `size_param: limit`, `start_page: 1`,
`page_size: 50`, matching legacy's `katanaDefaultPageSize`/`PageNumberPaginator`); the engine stops
once a page returns fewer than `page_size` records, exactly like legacy's short-page stop rule.
`max_pages` defaults to `0` (unlimited), matching legacy's `katanaMaxPages` default.

Every Katana object exposes ISO-8601 `created_at`/`updated_at` timestamps; `updated_at` is declared
as `x-cursor-field` in every schema for manifest-surface parity with legacy's published
`CursorFields: []string{"updated_at"}` catalog entries. However, **legacy's `Read` never actually
applies an incremental filter** — there is no `incrementalLowerBound`/`request_param` wiring anywhere
in `katana.go`'s `Read`, only a full-page-by-page `Harvest` call every sync. This bundle matches that
exact behavior: no `incremental` block is declared on any stream, so every read is a full stream scan,
identical to legacy's real (not merely published) behavior.

## Write actions & risks

None. Katana is a read-only source for pm; `capabilities.write` is `false` and no `writes.json` is
shipped, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **Published cursor fields are not enforced.** As noted above, legacy's `katanaStreams()` catalog
  declares `CursorFields: []string{"updated_at"}` on every stream, but `Read` performs a full sync
  regardless — no request parameter or client-side filter derives from it. This bundle reproduces
  that exact gap (schema declares `x-cursor-field: updated_at`, no `incremental` block), rather than
  inventing a filter legacy itself never applied.
- Full Katana API surface (purchase orders, stocktakes, webhooks, sales-order writes) is out of scope
  for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries. Only the 5
  legacy-parity read streams are implemented.
