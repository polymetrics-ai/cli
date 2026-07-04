# Overview

Reads CoinMarketCap's Pro API: global aggregate market metrics, id/slug/symbol-keyed cryptocurrency
detail lookups, price conversion, fear-and-greed index, and altcoin season index. Migrates
`internal/connectors/coinmarketcap` (the legacy hand-written connector, which stays registered and
unchanged until wave6's registry flip) at capability parity for its `global_metrics` stream, then
this Pass B full-surface expansion adds every other dialect-expressible (non-1-based-start/limit)
GET endpoint CoinMarketCap documents: 9 streams total. The connector's 4 originally-quarantined
paginated streams (`map`/`listings_latest`/`categories`/`fiat`) remain blocked — re-reviewed against
every dialect addition through the S4 engine mini-wave (`fan_out`, `keyed_object`, `start_page: 0`,
oauth2 `extra_params`, date-only incremental bounds) and again for this Pass B pass; none apply to
the pagination gap these 4 streams (and every sibling `start`/`limit`-paginated list endpoint the
wider API surface documents) share. The API is read-only; this bundle exposes no write actions.

## Auth setup

CoinMarketCap authenticates via the `X-CMC_PRO_API_KEY` header. `spec.json` requires `api_key`
(`x-secret: true`); `streams.json`'s `base.auth` declares a single unconditional `api_key_header`
candidate (matching legacy, which always requires this secret — `coinmarketcapSecret` hard-errors
when empty, no anonymous/public mode exists for this API). `base_url` defaults to
`https://pro-api.coinmarketcap.com`, matching legacy's `coinmarketcapDefaultBaseURL`.

## Streams notes

`global_metrics` reads `GET /v1/global-metrics/quotes/latest`, not paginated
(`endpoint.paginated: false` in legacy's routing table — `readSingle` issues exactly one request),
not incremental (legacy publishes no `CursorFields`). CoinMarketCap wraps every payload in
`{status:{...}, data:...}`; `data` is a single object here, so `records.path: "data"` (via
`connsdk.RecordsAt`'s object-becomes-one-record behavior) yields exactly one record per read,
matching legacy's `readSingle` + `globalMetricsRecord` field set (`active_cryptocurrencies`,
`total_cryptocurrencies`, `active_market_pairs`, `active_exchanges`, `total_exchanges`,
`btc_dominance`, `eth_dominance`, `last_updated`, `quote`). `x-primary-key` is
`active_cryptocurrencies`, matching legacy's own declared (not-really-unique, single-record-stream)
`PrimaryKey` choice verbatim — not "fixed" here, since this bundle targets parity with legacy's
existing behavior, not a design improvement. It also sends an optional `convert` query param
(`config.convert`, defaulting `"USD"`) via the opt-in optional-query dialect.

The 8 Pass B streams added this pass, all genuinely non-paginated (parameterless single-object
responses, or id/slug/symbol-keyed lookups CoinMarketCap itself never paginates):

- `global_metrics_quotes_historical` (`GET /v1/global-metrics/quotes/historical`): records at
  `data.quotes` (a bare array under the object envelope). Optional `time_start`/`time_end`/`count`/
  `interval`/`convert` config-driven passthrough filters, all `omit_when_absent`. Deliberately NOT
  modeled as `incremental` — the endpoint's own `count`/`interval` snapshot-window semantics don't
  map cleanly onto "records newer than the last-seen cursor" without risking a silent behavior
  guess; full-refresh only, `x-primary-key: timestamp` with no `x-cursor-field` declared.
- `cryptocurrency_info` (`GET /v2/cryptocurrency/info`) and `cryptocurrency_quotes_latest`
  (`GET /v3/cryptocurrency/quotes/latest`): both take a required `id` query param
  (`config.cryptocurrency_ids`, comma-separated CMC ids) and return `data` as a JSON object KEYED BY
  the requested id (`{"1": {...}, "1027": {...}}`, confirmed against CoinMarketCap's own documented
  response shape for bundling-capable endpoints) — modeled with `records.keyed_object: true`,
  `key_field: cmc_id` so each exploded record carries its own source map key alongside the API's own
  `id` field.
- `price_conversion` (`GET /v2/tools/price-conversion`): single-object `data` response; requires
  `amount` (`config.price_conversion_amount`) plus exactly one of `id`/`symbol`
  (`config.price_conversion_id`/`config.price_conversion_symbol`, both `omit_when_absent` — the
  caller is responsible for setting exactly one, matching CoinMarketCap's own "one of id or symbol
  required" rule; the engine has no `anyOf`-equivalent to enforce this at the config layer, so this
  is documented operator guidance, not a validated constraint).
- `fear_and_greed_latest` (`GET /v3/fear-and-greed/latest`), `altcoin_season_index_latest`
  (`GET /v1/altcoin-season-index/latest`), `key_info` (`GET /v1/key/info`): all three are
  parameterless, single-object `data` responses.
- `altcoin_season_index_historical` (`GET /v1/altcoin-season-index/historical`): `data` is an
  object with a `timeframe` field plus a `points` array — `records.path: "data.points"`. Optional
  `timeframe` config (`config.altcoin_season_timeframe`, default `"7d"`, matching CoinMarketCap's
  own documented default).

## Write actions & risks

None. CoinMarketCap is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`. Every documented CoinMarketCap Pro
API endpoint is itself read-only (GET); the Pro API has no write/mutation surface at all.

## Known limits

- **Blocked: `map`, `listings_latest`, `categories`, `fiat` streams (`ENGINE_GAP`, unchanged in
  substance from the original quarantine finding, re-verified against every new dialect addition).**
  Legacy's `harvest` (`coinmarketcap.go`) pages these 4 endpoints with CoinMarketCap's documented
  1-based `start`/`limit` convention: `start := 1` on the first request, then `start += pageSize` on
  each subsequent page (i.e. `1`, `1+pageSize`, `1+2*pageSize`, ...), stopping when a page returns
  fewer than `pageSize` records. Neither dialect pagination type expresses this:
  - `page_number` (the type the new `start_page: 0`/`start_page: 1` flexibility applies to) sends
    its `page` value **verbatim** as the query param and advances by exactly **+1** per page
    (`pageNumberPaginator.Next`: `p.page++`) — even with `start_page: 1` declared, it would send
    `start=1`, `start=2`, `start=3`, ... never `start=1`, `start=101`, `start=201`.
  - `offset_limit` (`connsdk.OffsetPaginator`) advances by `+pageSize` per page (the stride CMC
    needs) but its `Start()` is hardcoded to begin at offset `0` — `PaginationSpec` exposes no
    start-offset override field for this type at all (`StartPage` is read only by the
    `page_number` case in `newPaginator`'s switch; `offset_limit`'s case never consults it), so it
    would send `start=0`, `start=0+pageSize`, ... — off by exactly one record's worth of the true
    start on every page, silently skipping the record CMC returns first for the real first page's
    `start=1`. This is a real accepted-input record-data change (an off-by-`1` window shift on
    every page, not just the first), not a cosmetic deviation.
  - Re-checked every S3/S4 mini-wave addition (`fan_out`, `keyed_object`, `extra_params`,
    date-only `parseLowerBoundTime`, typed `computed_fields`) — none add a configurable start-offset
    to `offset_limit` or a configurable per-page stride to `page_number`; the underlying gap (no
    pagination type combines "arbitrary 1-based start" with "advance by page size, not by 1") is
    unchanged from the original quarantine finding, even though the exact wording differs because
    this pass checked it against the newly available primitives specifically.
  - Would need either a new `offset_limit`-family field (an explicit `start_offset`, mirroring
    `page_number`'s now-pointer-typed `start_page`) or a Tier-2 `StreamHook`/hand-rolled paginator;
    filed as `ENGINE_GAP` rather than escalated to a hooks package, since a hook here would exist
    for pagination alone with no other Tier-2 trigger in this connector.
- **Not covered this pass (deliberately narrowed, not silently dropped — see `api_surface.json`
  for the per-endpoint reason)**: every remaining `start`/`limit`-paginated list endpoint across
  Cryptocurrency/Exchange/Global-Metrics/Content categories shares the identical `ENGINE_GAP` above
  (market-pairs/latest, trending/*, airdrops, exchange map/listings/market-pairs,
  fear-and-greed/historical, content/latest). A second, narrower group of id/slug/symbol-keyed
  detail endpoints in the SAME families as the newly covered streams
  (`quotes/historical`/`ohlcv/latest`/`ohlcv/historical`/`price-performance-stats/latest`,
  `category`/`airdrop` singular lookups, every `exchange/*` detail endpoint) is dialect-expressible
  in principle but not implemented this pass — each is a distinct response shape not yet reviewed,
  or (for the `exchange/*` family) depends on an id list sourced from the very
  `start`/`limit`-paginated `exchange/map` this connector cannot page through. DEX Data, Derivatives,
  CMC Index, and Content/Community (cursor-by-score pagination, a different shape from the
  `start`/`limit` gap) are distinct product surfaces out of scope for this connector's core
  cryptocurrency/exchange/global-metrics streams. The 21-endpoint Deprecated bucket (Flipside
  Crypto FCAS, legacy blockchain/statistics) is excluded as `deprecated`, not a coverage gap.
