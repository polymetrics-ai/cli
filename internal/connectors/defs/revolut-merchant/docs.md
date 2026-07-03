# Overview

Revolut Merchant is a wave2 fan-out declarative-HTTP migration. It reads Revolut Merchant orders,
customers, settlements, and payment links through the Merchant API
(`GET https://merchant.revolut.com/api/1.0/...`). This bundle targets capability parity with
`internal/connectors/revolut-merchant` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip. Read-only
(`capabilities.write` is `false`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`).

## Auth setup

Provide a Revolut Merchant API secret key via the `api_key` secret. It is sent as an
`Authorization: Bearer <api_key>` header, matching legacy's `connsdk.Bearer(key)`
(`revolut_merchant.go:170`). Every request also carries a fixed `Revolut-Api-Version: 2023-09-01`
header, matching legacy's `DefaultHeaders: map[string]string{"Revolut-Api-Version": "2023-09-01"}`
— declared as a static-literal value in `streams.json`'s `base.headers` (no template markers, so
it is a fixed constant, not config-driven). `base_url` defaults to
`https://merchant.revolut.com/api/1.0` and may be overridden for tests/proxies.

## Streams notes

All 4 streams (`orders`, `customers`, `settlements`, `payment_links`) share the identical shape:
`GET` against the Merchant API's list endpoint (`/orders`, `/customers`, `/settlements`,
`/payment-links` — note the endpoint's own hyphenated path vs. the stream's snake_case name,
matching legacy's `endpoints["payment_links"] = {"payment-links", ...}` exactly), records at the
response body's root array (`records.path: ""`) — legacy's own `recordsPath` for every endpoint is
`""` (`revolut_merchant.go:106-111`), so this reproduces the exact primary-candidate behavior;
legacy's `recordsAt` fallback list would only try `"data"`/`"items"`/etc. for a differently-shaped
response, which the root-array envelope never triggers. Pagination is `page_number`
(`page`/`limit`, `page_size: 100`), stopping on a short page exactly as legacy's
`connsdk.PageNumberPaginator` does.

Legacy applies four passthrough filters (`from_created_date`, `to_created_date`, `customer_id`,
`state`) identically to every stream's request (`revolut_merchant.go:87-92`'s loop iterates a
fixed key list regardless of which stream is being read) — this bundle reproduces that exact
blanket behavior via the identical four `omit_when_absent` query entries declared on EACH of the
four streams' own `query` block (`HTTPBase` has no `query` field in the engine dialect, so this is
per-stream duplicated rather than a single shared declaration), sent only when the corresponding
config value is set, matching legacy's own `strings.TrimSpace(...) != ""` gate. `computed_fields`
stamps a static `stream` marker on every record, matching legacy's `mapRecord`'s
`out["stream"] = stream`.

`created_at` is declared as `x-cursor-field` on every schema, matching legacy's own
`CursorFields: []string{"created_at"}` Catalog declarations for all 4 streams. No `incremental`
block is declared: legacy's `Read` never reads a persisted sync cursor back into
`from_created_date`/`to_created_date` (`harvest` reads only `req.Config.Config[key]`, never
`req.State["cursor"]`) — it always resends the exact same raw config value on every sync, with no
forward advancement.

## Write actions & risks

None. Legacy `revolut_merchant.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`page_size` bounded 1-100, default 100; `max_pages` 0/all/unlimited for unbounded).
  The engine's `page_number` paginator reads `PaginationSpec.PageSize`/`MaxPages` as static
  bundle-authored integers, not config templates — there is no mechanism to wire a `spec.json`
  property into either field. This bundle sends `page_size: 100` (legacy's own default) as a
  static value in `streams.json`'s `base.pagination` block; neither `page_size` nor `max_pages` is
  declared in `spec.json` (F6: dead config is worse than absent config). Pagination is otherwise
  unbounded (matches legacy's `max_pages: 0` = unlimited default) other than the short-page stop
  signal.
- **Legacy's `id` fallback (`uuid`/`email`/`reference`) is not modeled.** Legacy's `mapRecord`
  falls back to a record's `uuid`, `email`, or `reference` field when `id` is absent. Every Revolut
  Merchant resource this bundle reads always carries an `id` in its real wire shape (legacy's own
  `Catalog`/`PrimaryKey` declarations assume `id` unconditionally for all 4 streams), so this
  fallback is defensive dead code against the real API — not exercised by any input legacy itself
  would realistically receive. Documented here for completeness, not implemented via a hook.
- The full Revolut Merchant API surface (order create/capture/cancel/refund, customer
  create/update, webhook management, payout endpoints) is out of scope for this wave; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
