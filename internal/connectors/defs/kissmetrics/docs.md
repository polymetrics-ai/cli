# Overview

Kissmetrics reads products, reports, events, and properties through the Kissmetrics query API
(`https://query.kissmetrics.io/v3`) using HTTP Basic authentication. This bundle migrates
`internal/connectors/kissmetrics` (the hand-written connector) to a declarative defs bundle at
capability parity; the legacy package stays registered and unchanged until wave6's registry flip.
Kissmetrics exposes a read-only query API, so `capabilities.write` is `false` and this bundle ships
no `writes.json`.

## Auth setup

Provide a Kissmetrics API key/username via the `username` config value and its matching secret via
the `password` secret; both flow only into HTTP Basic auth (`Authorization: Basic
base64(username:password)`) and the password is never logged, matching legacy's
`connsdk.Basic(username, password)` wiring exactly. `base_url` defaults to
`https://query.kissmetrics.io/v3` (legacy's `kissmetricsDefaultBaseURL`).

## Streams notes

The top-level `products` stream lists every product (account) visible to the authenticated user at
`GET /products`. `reports`, `events`, and `properties` are nested under a product partition
(`GET /products/{product_id}/<resource>`) and therefore require the `product_id` config value — this
matches legacy's `streamPath` helper exactly, including using the identical path escaping
(`url.PathEscape` in legacy; the engine's `InterpolatePath` urlencodes path segments by default).
Attempting to read a nested stream without `product_id` configured hard-errors, matching legacy's own
`kissmetrics stream %q requires config product_id` error.

All 4 streams share records at the response body's `data` array and `offset_limit` pagination
(`limit_param: limit`, `offset_param: offset`, `page_size: 50`, matching legacy's
`kissmetricsDefaultPageSize`/`OffsetPaginator`); the engine stops once a page returns fewer than
`page_size` records. `max_pages` defaults to `0` (unlimited), matching legacy's `kissmetricsMaxPages`
default. No stream declares an `incremental` block or `x-cursor-field`: legacy's
`kissmetricsStreams()` catalog publishes no `CursorFields` for any stream (the API supports full
refresh only), so this bundle matches that exactly.

## Write actions & risks

None. Kissmetrics is a read-only query source for pm; `capabilities.write` is `false` and no
`writes.json` is shipped, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`).

## Known limits

- **`product_id` is a single shared config value, not per-stream.** Legacy's own design has the same
  limitation: one connector configuration can only scope `reports`/`events`/`properties` reads to one
  product at a time. Reading multiple products' nested resources requires separate connector
  configurations, identical to legacy.
- Full Kissmetrics API surface (report result data, ad-hoc queries) is out of scope for this wave;
  see `api_surface.json`'s `excluded: {category: out_of_scope}` entries. Only the 4 legacy-parity
  read streams are implemented.
