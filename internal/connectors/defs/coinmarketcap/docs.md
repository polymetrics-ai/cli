# Overview

Reads CoinMarketCap's latest aggregate global cryptocurrency market metrics from the Pro API
(`GET /v1/global-metrics/quotes/latest`), migrating `internal/connectors/coinmarketcap` (the legacy
hand-written connector, which stays registered and unchanged until wave6's registry flip) at
capability parity for its `global_metrics` stream only. This bundle is an **unblock re-review**, not
a fresh migration: the connector was previously quarantined entirely (`ENGINE_GAP`,
`docs/migration/quarantine.json`) for its 4 paginated streams' 1-based `start`/`limit` pagination
shape. Re-reviewed against the new dialect additions (`fan_out`, `keyed_object`, `start_page: 0`,
oauth2 `extra_params`, date-only incremental bounds) — none of those apply to the pagination gap
these 4 streams share; they remain blocked. The API is read-only; this bundle exposes no write
actions.

## Auth setup

CoinMarketCap authenticates via the `X-CMC_PRO_API_KEY` header. `spec.json` requires `api_key`
(`x-secret: true`); `streams.json`'s `base.auth` declares a single unconditional `api_key_header`
candidate (matching legacy, which always requires this secret — `coinmarketcapSecret` hard-errors
when empty, no anonymous/public mode exists for this API). `base_url` defaults to
`https://pro-api.coinmarketcap.com`, matching legacy's `coinmarketcapDefaultBaseURL`.

## Streams notes

Only `global_metrics` is implemented. It reads `GET /v1/global-metrics/quotes/latest`, not
paginated (`endpoint.paginated: false` in legacy's routing table — `readSingle` issues exactly one
request), not incremental (legacy publishes no `CursorFields`). CoinMarketCap wraps every payload
in `{status:{...}, data:...}`; `data` is a single object here, so `records.path: "data"` (via
`connsdk.RecordsAt`'s object-becomes-one-record behavior) yields exactly one record per read,
matching legacy's `readSingle` + `globalMetricsRecord` field set (`active_cryptocurrencies`,
`total_cryptocurrencies`, `active_market_pairs`, `active_exchanges`, `total_exchanges`,
`btc_dominance`, `eth_dominance`, `last_updated`, `quote`). `x-primary-key` is
`active_cryptocurrencies`, matching legacy's own declared (not-really-unique, single-record-stream)
`PrimaryKey` choice verbatim — not "fixed" here, since this bundle targets parity with legacy's
existing behavior, not a design improvement.

## Write actions & risks

None. CoinMarketCap is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's
`Write` always returning `connectors.ErrUnsupportedOperation`.

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
- Full CoinMarketCap API surface (historical quotes, OHLCV, exchange listings/info, DEX/on-chain
  endpoints) is out of scope until Pass B; see `api_surface.json`.
