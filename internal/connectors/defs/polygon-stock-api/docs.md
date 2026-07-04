# Overview

Polygon Stock API is a read-only declarative-HTTP bundle migrated from
`internal/connectors/polygon-stock-api` (the hand-written legacy connector, which stays registered
and unchanged until wave6's registry flip). It reads Polygon.io stock tickers, dividends, and
splits through the Polygon.io reference REST API.

## Auth setup

Provide a Polygon.io API key via the `api_key` secret. It is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(key)` construction exactly.

## Streams notes

All 3 streams (`tickers`, `dividends`, `splits`) share the same shape: `GET` against a Polygon
reference list endpoint, records at `results`, `limit=<page_size>` (default `100`, matching
legacy's `defaultPageSize`) sent on the first request. Each stream's optional passthrough query
filters (`ticker`/`sort`/`order` on all three; `market`/`locale`/`type`/`active` on `tickers` only;
`ex_dividend_date` on `dividends`; `execution_date` on `splits`) use `streams.json`'s optional-query
object dialect (`omit_when_absent: true`) so an unset config value is left off the request
entirely, matching legacy's `for _, key := range [...] { if v := cfg.Config[key]; v != "" {
q.Set(key, v) } }` loop. The `tickers` stream's `active` filter additionally carries `"default":
"true"` — legacy always sets `active=true` on the tickers endpoint specifically when the config
value is unset (`if endpoint.path == "v3/reference/tickers" && q.Get("active") == "" { q.Set
("active", "true") }`), a stream-specific default this bundle reproduces exactly (the other two
streams have no such default).

Pagination follows Polygon's `next_url` absolute-URL convention (`pagination.type: next_url`,
`next_url_path: next_url`) — the engine's built-in same-host SSRF guard and loop-guard (same URL
requested twice) reproduce legacy's `harvest`-style loop, which followed `next_url` verbatim until
it was empty. `primary_key` is `["ticker"]` for `tickers` (Polygon's own natural key) and `["id"]`
for `dividends`/`splits` (Polygon's own opaque record identifier, confirmed present on every real
dividend/split object per Polygon's API reference — legacy's `first(item, "id", "cash_amount")`/
`first(item, "id", "execution_date")` fallback for a missing `id` is defensive dead code for a
shape the real API never actually returns; this bundle's plain schema projection of `id` is
equivalent for every real response).

`dividends`/`splits` declare `x-cursor-field` (`ex_dividend_date`/`execution_date` respectively)
for schema/manifest parity with legacy's `CursorFields`, but — matching legacy exactly — no
`incremental` block is declared on either stream: legacy's `harvest` never sends a server-side
date-range filter param for either endpoint (the `ex_dividend_date`/`execution_date` config values
are plain passthrough query filters, not incremental-cursor-driven), so adding one here would be
new, legacy-diverging behavior, not a straight port.

## Write actions & risks

None. Polygon Stock API is a read-only source in both legacy and this bundle
(`capabilities.write: false`, no `writes.json`).

## Known limits

- Only the 3 legacy-parity streams (`tickers`, `dividends`, `splits`) are implemented; the full
  Polygon.io stocks surface (ticker details, financials, aggregates/bars, snapshots, exchanges,
  market status) is out of scope for this wave — see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "not implemented in this bundle"}` entries.
- `max_pages` (legacy default `3`, config-overridable including `0`/`all`/`unlimited` for
  unbounded) is NOT exposed as a config property here: the `next_url` paginator's `MaxPages` is a
  static `int` field on `streams.json`'s `pagination` block, with no runtime config-driven override
  mechanism at all (same F6 lesson as searxng's `max_pages`/`page_size` — a `spec.json` property no
  template anywhere in the bundle consumes should not be declared). `pagination.max_pages` is fixed
  at `3` here, matching legacy's own default constant exactly; a config-time override of this value
  is out of scope until the engine dialect grows one. `page_size`, by contrast, IS config-driven
  (sent via each stream's own `query.limit`, independent of the paginator).
- Legacy's `pageSize`/`maxPages` numeric-range validation (`page_size` clamped 1-1000, `max_pages`
  clamped to a configured ceiling) is not reproduced: the engine dialect has no numeric-range
  validation for `spec.json` string-typed config values. An out-of-range `page_size` is sent to
  Polygon as-is rather than rejected client-side or clamped; Polygon's own API validates the `limit`
  query param server-side.
- Per `conventions.md` §4's sanctioned `next_url` exception, every stream ships a single-page
  fixture (the replay server's own absolute URL cannot be embedded in a static fixture ahead of
  time) — `pagination_terminates` exercises the `tickers` stream's 1-page fixture and confirms
  exactly one request is issued and consumed, proving the loop terminates on a null `next_url`
  rather than looping. No live `paritytest/polygon-stock-api` 2-page test exists for this migration
  wave; the single-page fixture is the Pass-A-scope proof here.
- The engine's `next_url` paginator stops purely on an absent/empty `next_url` value; legacy also
  independently stops early on a short page (`len(records) < pageSize`). Real Polygon responses
  only set `next_url` on a genuinely-incomplete final page, so this never diverges for any real API
  response — documented here because it is a structural difference in the stop condition, not
  because it is reachable behavior.
