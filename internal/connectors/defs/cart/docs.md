# Overview

Cart.com reads orders, customers, products, and inventory through a read-only REST API against a
per-store `base_url`. This bundle migrates `internal/connectors/cart` (the hand-written connector)
to a declarative defs bundle at capability parity; the legacy package stays registered and
unchanged until wave6's registry flip. It is a pure Tier-1 declarative migration — legacy is a
connsdk-HTTP-based connector with no custom auth/stream/write hooks and no protocol-native
(SQL/queue/SDK) surface, so no `hooks/cart/` package or `native/cart/` component split is needed.

**Tier/label note**: the migration catalog entry for this connector is tagged native/destination,
but that label does not match the legacy ground truth read during this migration.
`internal/connectors/cart/cart.go`'s `Write` unconditionally returns
`connectors.ErrUnsupportedOperation` (`Capabilities.Write: false`), and there is no writer, no
protocol-native connection/reader/cataloger split, and no non-HTTP protocol anywhere in the legacy
package — it is a plain read-only `connsdk.Requester`-based HTTP client, structurally identical to
`braintree`. This bundle is therefore a Tier-1 declarative bundle (JSON + docs only), not a Tier-3
native package and not a destination with `writes.json`; there is no destination capability in
legacy to preserve.

## Auth setup

Provide `base_url` and `access_token`. `access_token` is sent as a Bearer token
(`Authorization: Bearer <access_token>`), matching legacy's `connsdk.Bearer(token)`, and is never
logged.

`base_url` is **required** in this bundle (`spec.json`'s `required: ["base_url", "access_token"]`).
Legacy instead accepts EITHER `base_url` directly OR a `store_name` value it derives a URL from
(`"https://" + store_name + "/api/v1"`) when `base_url` is unset; this bundle does not reproduce
that derivation (identical shape to chargebee's `site`→`base_url` narrowing already ledgered in
`docs/migration/conventions.md`) — see Known limits for why and the config-surface change this
implies for an operator migrating a legacy `store_name`-only config.

## Streams notes

Four streams:

- `orders` — `GET /orders`, records at `data`.
- `customers` — `GET /customers`, records at `data`.
- `products` — `GET /products`, records at `data`.
- `inventory` — `GET /inventory`, records at `data`.

All four send `page_size` (default `100`, matching legacy's `defaultPageSize`) as a static
per-stream query value templated from `config.page_size`. Pagination is `cursor` with `token_path:
meta.next_page` (`cursor_param: page`) — the next page's `page` value is read verbatim from the
current page's response body at `meta.next_page`, matching legacy's `readPaged` exactly (`next,
err := connsdk.StringAt(resp.Body, "meta.next_page")`; an absent/empty value stops pagination).
Legacy always sends an explicit `page=1` on the FIRST request too; this bundle's `cursor`+
`token_path` paginator sends no `page` param at all on the first request (only from the second page
onward, once a `meta.next_page` token is found) — Cart.com's documented list endpoints default an
absent `page` to page 1, so this never changes which records are returned, only whether the literal
string `page=1` appears on the wire for the first request (see Known limits).

Every stream declares `projection: "passthrough"`: legacy's `readPaged` emits
`connectors.Record(rec)` — a verbatim cast of the raw decoded JSON object, not a field-built
mapping — so schema-mode projection (which would drop every field not explicitly declared) would
silently narrow legacy's real output. `schemas/*.json` declare the fields legacy's own `Catalog()`
and test fixtures name explicitly (`id`, `order_number`, `updated_at`) as the verified baseline;
passthrough mode means any additional real Cart.com API field survives in the emitted record
regardless of whether it is declared in the schema.

None of the streams are incremental: legacy's `Catalog()` sets no `CursorFields` on any of
them (despite exposing an `updated_at` timestamp field), and `Read`/`readPaged` performs no
time-based filtering at all (full extraction every sync). This bundle therefore declares no
`incremental` block and no `x-cursor-field`, matching legacy's real behavior rather than its
timestamp-shaped-but-unused field.

## Write actions & risks

None. This connector is `capabilities.write: false`; no `writes.json` is shipped, matching
legacy's `Write` always returning `connectors.ErrUnsupportedOperation`. See the Tier/label note
above — legacy has no write or native-destination surface to migrate.

## Known limits

- Shipments, coupons, and mutation surfaces remain excluded from this bundle's reviewed surface.
  `inventory` is modeled as a passthrough collection stream; order creation and product update stay
  excluded because the legacy connector exposes no reverse-ETL contract or approved write schema.
- **`store_name` config key dropped; `base_url` is now required.** Legacy derives the API host
  from a `store_name` config value (`"https://" + store_name + "/api/v1"`) when `base_url` is
  unset. The engine's spec-default materialization only fills in a LITERAL per-key default — it
  cannot express "derive `base_url` from `store_name`", a cross-key template (the same class of gap
  chargebee's `site`/`base_url` pair and sentry's `hostname` hit). Per
  `docs/migration/conventions.md`'s guidance for this exact shape, this bundle drops `store_name`
  entirely and requires `base_url` instead: an operator migrating a legacy `store_name`-only config
  must now supply the fully-formed `https://{store_name}/api/v1` URL as `base_url`. This is a
  documented config-surface narrowing (every legacy-accepted `store_name` value has an
  operator-reachable `base_url` equivalent; no request/data change once configured), not a data-shape
  regression.
- **First-page request omits the literal `page=1` query param.** Legacy's `readPaged` always sends
  `page` starting at `"1"` explicitly; this bundle's `cursor`+`token_path` paginator sends no `page`
  param at all on the first request (only subsequent pages, once a `meta.next_page` token is found
  in the response body). Cart.com's documented list endpoints default an absent `page` to page 1,
  so this is a wire-shape difference only, never a difference in which records are returned for any
  input legacy itself would accept (`ACCEPTABLE`, matching this repo's parity-deviation meta-rule).
- **`max_pages` config override is not modeled.** Legacy exposes `max_pages` (default 100) as a
  config-driven hard cap on request count. The engine's `cursor`+`token_path` paginator has no
  config-driven request-count-cap knob — `PaginationSpec.MaxPages` is a fixed JSON integer declared
  once in `streams.json`, not a per-request-templatable value — so `max_pages` is not declared in
  `spec.json` (a declared-but-unwireable key is worse than an absent one, per this repo's
  dead-config rule). This bundle relies on the token-absence stop signal alone (matching Cart.com's
  real pagination termination for any bounded result set); an operator needing a hard page-count cap
  for a specific sync can express it at the orchestration layer instead.
