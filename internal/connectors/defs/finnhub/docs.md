# Overview

Reads Finnhub stock symbols, market news, per-symbol company profiles, and per-symbol analyst
recommendation trends through the Finnhub REST API.

Readable streams: `stock_symbols`, `market_news`, `company_profile`, `stock_recommendations`.

This connector is read-only; no write actions are declared.

Service API documentation: https://finnhub.io/docs/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Finnhub API key. Sent only as the X-Finnhub-Token header;
  never logged.
- `base_url` (optional, string); default `https://finnhub.io/api/v1`; format `uri`; Finnhub API base
  URL override for tests or proxies.
- `exchange` (optional, string); default `US`; Exchange code for the stock_symbols stream (default
  US).
- `market_news_category` (optional, string); default `general`; News category for the market_news
  stream (default general).
- `mode` (optional, string).
- `symbols` (optional, string); Comma-, whitespace-, or semicolon-separated stock ticker symbols to
  fan out over for the company_profile and stock_recommendations streams (one request per symbol).
  Required for those two streams only; stock_symbols/market_news do not reference it.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://finnhub.io/api/v1`, `exchange=US`,
`market_news_category=general`.

Authentication behavior:

- API key authentication in `X-Finnhub-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/stock/symbol` with query `exchange`=`US`.

## Streams notes

Default pagination: single request; no pagination.

- `stock_symbols`: GET `/stock/symbol` - records at response root; query `exchange`=`{{
  config.exchange }}`.
- `market_news`: GET `/news` - records at response root; query `category`=`{{
  config.market_news_category }}`; computed output fields `symbol`.
- `company_profile`: GET `/stock/profile2` - records at response root; fan-out; ids from config
  field `symbols`; id sent as query parameter `symbol`; stamps `ticker`.
- `stock_recommendations`: GET `/stock/recommendation` - records at response root; fan-out; ids from
  config field `symbols`; id sent as query parameter `symbol`; stamps `symbol`.

## Write actions & risks

This connector is read-only. Read behavior: external Finnhub API read of market data.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
