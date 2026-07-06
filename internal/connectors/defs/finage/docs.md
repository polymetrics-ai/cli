# Overview

Reads Finage US market data: most active stocks, top gainers and losers, sector performance,
delisted companies, and per-symbol market news via the Finage REST API.

Readable streams: `most_active_us_stocks`, `most_gainers`, `most_losers`, `sector_performance`,
`delisted_companies`, `market_news`, `earnings_calendar`, `ipo_calendar`.

This connector is read-only; no write actions are declared.

Service API documentation: https://finage.co.uk/docs/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Finage API key. Sent only as the apikey query parameter;
  never logged.
- `base_url` (optional, string); default `https://api.finage.co.uk`; format `uri`; Finage API base
  URL override for tests or proxies.
- `calendar_from` (optional, string); format `date`; Start date (YYYY-MM-DD) for the
  earnings_calendar and ipo_calendar streams' required from/to date-range window. Required for those
  two streams only; the other streams do not reference it.
- `calendar_to` (optional, string); format `date`; End date (YYYY-MM-DD) for the earnings_calendar
  and ipo_calendar streams' required from/to date-range window. Required for those two streams only;
  the other streams do not reference it.
- `mode` (optional, string).
- `symbols` (optional, string); Comma- or whitespace-separated stock ticker symbols to fan out over
  for the market_news stream (one GET /news/market/{symbol} request per symbol). Required for that
  stream only; the other streams do not reference it.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.finage.co.uk`.

Authentication behavior:

- API key authentication in query parameter `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/fnd/market-information/us/sector-performance`.

## Streams notes

Default pagination: single request; no pagination.

- `most_active_us_stocks`: GET `/fnd/market-information/us/most-actives` - records at response root.
- `most_gainers`: GET `/fnd/market-information/us/most-gainers` - records at response root.
- `most_losers`: GET `/fnd/market-information/us/most-losers` - records at response root.
- `sector_performance`: GET `/fnd/market-information/us/sector-performance` - records at response
  root.
- `delisted_companies`: GET `/fnd/delisted-companies/` - records at response root; query
  `limit`=`1000`; `period`=`annual`.
- `market_news`: GET `/news/market/{{ fanout.id }}` - records path `news`; query `limit`=`30`;
  fan-out; ids from config field `symbols`; id inserted into the request path; stamps `symbol`.
- `earnings_calendar`: GET `/fnd/earning-calendar` - records at response root; query `from`=`{{
  config.calendar_from }}`; `to`=`{{ config.calendar_to }}`.
- `ipo_calendar`: GET `/fnd/ipo-calendar` - records at response root; query `from`=`{{
  config.calendar_from }}`; `to`=`{{ config.calendar_to }}`.

## Write actions & risks

This connector is read-only. Read behavior: external Finage API read of market data.

## Known limits

- API coverage includes 8 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=1, non_data_endpoint=1, out_of_scope=14.
