# Overview

CoinAPI is a read-only market-data source (the entire documented REST surface is GET-only; there is
no write/mutation capability in the API at all). This bundle started as a migration of
`internal/connectors/coin-api` (the hand-written legacy connector, which stays registered and
unchanged until wave6's registry flip) at parity, then this Pass B pass expanded it to the practical
current documented surface after reviewing the live CoinAPI Market Data REST API OpenAPI spec (see
`api_surface.json`). 9 streams total: 3 metadata reference streams (`symbols`, `exchanges`,
`assets`), 2 symbol-scoped historical streams (`ohlcv_historical_data`, `trades_historical_data`),
and 4 new in this pass (`exchange_rates`, `quotes_current`, `orderbook_current`,
`metrics_listing`).

## Auth setup

Provide a CoinAPI key via the `api_key` secret; it is sent as the `X-CoinAPI-Key` request header
(`streams.json` `base.auth`'s `api_key_header` mode) and never logged, matching legacy's
`connsdk.APIKeyHeader(coinAPIAuthHeader, secret, "")` exactly.

`base_url` optionally overrides the API host (for tests/proxies). When unset, `environment`
(`production` default, or `sandbox`) selects between `https://rest.coinapi.io` and
`https://rest-sandbox.coinapi.io` — legacy's own `coinAPIBaseURL` fallback logic. Only the fixed
production/sandbox pair is representable via the engine's plain `{{ config.base_url }}` +
`"default"` materialization mechanism; the `environment`-conditional derivation itself is left to
operator convention (set `base_url` explicitly for sandbox), documented under Known limits below.

## Streams notes

`exchanges` and `assets` are full-refresh reference lists: `GET /v1/<resource>` with no query
parameters, records read from the top-level JSON array (`records.path: ""`), no pagination
(`type: none`) — matching legacy's `readMetadata`, which issues exactly one request per stream.

**`symbols`'s endpoint was corrected, not just extended, in this pass.** Legacy (and this bundle,
before this pass) read the bare `GET /v1/symbols` (an all-exchanges symbol listing with no query
parameters). CoinAPI's current, live OpenAPI spec
(`https://raw.githubusercontent.com/api-bricks/api-bricks-sdk/master/coinapi/market-data-api-rest/spec/openapi.yaml`,
reviewed 2026-07-04) no longer documents that bare endpoint at all — symbol listing is now
exchange-scoped only, via `GET /v1/symbols/{exchange_id}/active`. This bundle now requires a new
`exchange_id` config value (e.g. `BITSTAMP`) and reads `/v1/symbols/{{ config.exchange_id }}/active`
instead. This is filed as a correctness fix rather than a parity deviation: the OLD endpoint may
still work in practice (an unauthenticated probe against `https://rest.coinapi.io/v1/symbols`
returns a 401 auth-gate response rather than a 404, so its live routing status could not be
confirmed without a real API key), but it is no longer part of CoinAPI's documented, supported
surface, and Pass B's mandate is to target the real current documented API. `symbol_id`/
`exchange_id` filter query params on the new endpoint (`filter_symbol_id`/`filter_asset_id`) are not
wired — this bundle always lists every active symbol for the configured exchange, matching the
scope (if not the exact endpoint) of legacy's original all-exchanges listing.

`ohlcv_historical_data` and `trades_historical_data` are symbol-scoped historical series requiring
`symbol_id` (and, for OHLCV, `period`) in config. Both send `limit` (default `100`, materialized
from `spec.json`'s default exactly like legacy's `coinAPIDefaultLimit` fallback) and an optional
`time_end` (`config.end_date`, `omit_when_absent: true` — sent only when configured, matching
legacy's conditional `base.Set("time_end", timeEnd)`). `ohlcv_historical_data` additionally sends
`period_id` (`config.period`, required — legacy hard-errors when absent for this stream; this bundle
relies on the same absent-value-propagates-as-empty-string engine behavior, so an unset `period`
sends an empty `period_id` rather than erroring, a minor parity note — see Known limits).

Pagination is `type: cursor` with `last_record_field` (not `token_path`): the next page's
`time_start` is read from the LAST record's own time-cursor field
(`time_period_start`/`time_exchange`) rather than a separate token in the response envelope — this
is exactly legacy's `readTimeseries` loop (`timeStart = lastCursor`, where `lastCursor` comes from
`stringField(item, endpoint.cursorField)` on the last emitted record of the page). No `stop_path` is
declared: like legacy (`len(records) < limit || lastCursor == "" || lastCursor == timeStart`), the
engine's `lastRecordCursor` paginator stops when a page is empty (`recordCount == 0`) or the last
record has no usable cursor value; the engine does not special-case a *short-but-nonempty* page the
way legacy's `len(records) < limit` check does, so a short final page triggers one additional
request that returns empty before the read terminates — never emits duplicate or incorrect data,
just one harmless extra round trip (documented parity deviation, see Known limits).

Incremental reads use `time_period_start`/`time_exchange` as the cursor field and send `time_start`
(`param_format: rfc3339`, verbatim passthrough — CoinAPI's ISO-8601 wire format needs no
reformatting, matching legacy's raw string `time_start`/`start_date` passthrough) computed from the
persisted state cursor or, on a fresh sync, from `start_date`. `symbol_id` (and `period` for OHLCV)
are stamped onto every emitted record via `computed_fields` referencing `config.*` (never present on
the raw wire record) — matching legacy's explicit `rec["symbol_id"] = symbolID` /
`rec["period_id"] = period` assignments in `readTimeseries`.

**New Pass B streams** (verified against the live OpenAPI spec's documented response schemas):

- **`exchange_rates`** (`GET /v1/exchangerate/{{ config.asset_id_base }}`, records at `rates`):
  current exchange rates from a configured base asset (e.g. `BTC`) to every quote asset CoinAPI
  tracks. The envelope's own `asset_id_base` field is at the response ROOT, not on each `rates[]`
  item, so `computed_fields` stamps `config.asset_id_base` onto every emitted record (the same
  config-stamping pattern the historical streams already use for `symbol_id`). Primary key is the
  `(asset_id_base, asset_id_quote)` pair. Not paginated — one request returns every quote asset's
  current rate.
- **`quotes_current`** (`GET /v1/quotes/{{ config.symbol_id }}/current`, `records.path: "."` — the
  response body IS the record, a single object with no array envelope): the current best bid/ask
  for a configured symbol.
- **`orderbook_current`** (`GET /v1/orderbooks/{{ config.symbol_id }}/current`, `records.path:
  "."`): the current order book snapshot (`asks`/`bids` arrays of price/size levels, passed through
  as raw JSON arrays — no per-level schema is declared since level shape/depth is exchange- and
  request-dependent) for a configured symbol.
- **`metrics_listing`** (`GET /v1/metrics/listing`, records at the response root): the catalog of
  every metric ID CoinAPI supports (used to discover what's available from the `/v1/metrics/*`
  current/historical-value endpoints, which are themselves out of scope for this pass — see Known
  limits).

## Write actions & risks

None. CoinAPI is a read-only market-data API; `capabilities.write` is `false` and this bundle ships
no `writes.json`, matching legacy's `Write` stub (`connectors.ErrUnsupportedOperation`). This is not
a scope narrowing — the full documented CoinAPI REST surface (50 method+path operations reviewed;
see `api_surface.json`) is exclusively GET.

## Known limits

- **`environment`-based sandbox/production base URL selection is not fully modeled.** Legacy derives
  `base_url` from `environment` in code (`coinAPIBaseURL`: `production` → `https://rest.coinapi.io`,
  `sandbox` → `https://rest-sandbox.coinapi.io`, only when no explicit `base_url` override is set).
  The engine's `spec.json` `"default"` materialization only fills a fixed literal, not a value
  derived from another config field, so this bundle declares `environment` (documented, with its own
  enum and default) alongside `base_url` for operator guidance, but `streams.json`'s `base.url`
  templates `{{ config.base_url }}` directly with no default — an operator must set `base_url`
  explicitly (to either CoinAPI host) rather than relying on `environment` alone to select it. This
  is a config-surface narrowing versus legacy's in-code derivation, not a behavior change for any
  input that explicitly sets `base_url`.
- **A short-but-nonempty final page triggers one extra, empty-page request** before pagination
  stops, versus legacy's `len(records) < limit` short-page check which stops immediately on the
  final page itself. This never changes emitted record data (the extra request returns zero
  records) — see Streams notes above.
- **`ohlcv_historical_data`'s `period` config is required for correct behavior but not enforced at
  validate time.** Legacy hard-errors when `period` is unset for this stream
  (`"coin-api config period is required for ohlcv_historical_data"`); this bundle's declarative
  `period_id` query param has no equivalent required-at-read-time check beyond the engine's
  standard unresolved-key error path. `symbol_id` is documented as required in `spec.json`'s
  description for both historical streams for the same reason (the engine's draft-07 dialect has no
  per-stream conditional `required[]`).
- Full CoinAPI API surface remains out of scope beyond the 9 streams above; see
  `api_surface.json`'s `excluded` entries (50 method+path operations reviewed against the live
  OpenAPI spec, each with a specific real reason — mostly `duplicate_of` latest/current variants of
  already-covered resources, per-entity metrics values requiring iteration over the metrics catalog,
  order-book depth/history/v3-envelope variants, options data, and blockchain-chain metadata).
- **`orderbook_current`'s `asks`/`bids` are opaque arrays, not per-level-typed schemas.** CoinAPI's
  own OpenAPI spec declares these fields with no `items` schema at all (untyped), so this bundle
  matches that by declaring them `["array", "null"]` rather than guessing a `[price, size]` tuple
  shape the spec itself doesn't commit to.
- CoinAPI's `/v1/metrics/*` current/historical VALUE endpoints (as opposed to the covered
  `metrics_listing` catalog) require selecting a specific `metric_id` from a large, evolving
  catalog per asset/exchange/symbol/chain; implementing them would need either a `fan_out`-style
  iteration over `metrics_listing`'s own output (the dialect's `fan_out.ids_from.request` shape
  lists ids via ONE preliminary paginated GET, which doesn't fit "iterate over a large catalog and
  issue N further-filtered requests per selected metric") or per-metric connector configuration;
  out of scope for this pass.
