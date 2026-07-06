# Overview

Reads Polygon.io stock tickers, dividends, and splits through the Polygon.io reference REST API.

Readable streams: `tickers`, `dividends`, `splits`.

This connector is read-only; no write actions are declared.

Service API documentation: https://polygon.io/docs/stocks/getting-started.

## Auth setup

Connection fields:

- `active` (optional, string).
- `api_key` (required, secret, string); Polygon.io API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.polygon.io`; format `uri`; Polygon.io API base
  URL override for tests or proxies.
- `ex_dividend_date` (optional, string); Optional ex-dividend date filter, applied to the dividends
  stream.
- `execution_date` (optional, string); Optional split execution date filter, applied to the splits
  stream.
- `locale` (optional, string); Optional locale filter (e.g. us, global), applied to the tickers
  stream.
- `market` (optional, string); Optional market filter (e.g. stocks, otc), applied to the tickers
  stream.
- `mode` (optional, string).
- `order` (optional, string); Optional sort order (asc/desc), applied to all three streams.
- `page_size` (optional, string); default `100`; Records per page (1-1000), sent as the 'limit'
  query param on the first request.
- `sort` (optional, string); Optional sort field, applied to all three streams.
- `ticker` (optional, string); Optional ticker symbol filter, applied to all three streams.
- `type` (optional, string); Optional ticker type filter (e.g. CS, ETF), applied to the tickers
  stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.polygon.io`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v3/reference/tickers` with query `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next_url`; next URLs
stay on the configured API host; maximum 3 page(s).

- `tickers`: GET `/v3/reference/tickers` - records path `results`; query `active` from template `{{
  config.active }}`, default `true`; `limit`=`{{ config.page_size }}`; `locale` from template `{{
  config.locale }}`, omitted when absent; `market` from template `{{ config.market }}`, omitted when
  absent; `order` from template `{{ config.order }}`, omitted when absent; `sort` from template `{{
  config.sort }}`, omitted when absent; `ticker` from template `{{ config.ticker }}`, omitted when
  absent; `type` from template `{{ config.type }}`, omitted when absent; follows a next-page URL
  from the response body; URL path `next_url`; next URLs stay on the configured API host; maximum 3
  page(s).
- `dividends`: GET `/v3/reference/dividends` - records path `results`; query `ex_dividend_date` from
  template `{{ config.ex_dividend_date }}`, omitted when absent; `limit`=`{{ config.page_size }}`;
  `order` from template `{{ config.order }}`, omitted when absent; `sort` from template `{{
  config.sort }}`, omitted when absent; `ticker` from template `{{ config.ticker }}`, omitted when
  absent; follows a next-page URL from the response body; URL path `next_url`; next URLs stay on the
  configured API host; maximum 3 page(s).
- `splits`: GET `/v3/reference/splits` - records path `results`; query `execution_date` from
  template `{{ config.execution_date }}`, omitted when absent; `limit`=`{{ config.page_size }}`;
  `order` from template `{{ config.order }}`, omitted when absent; `sort` from template `{{
  config.sort }}`, omitted when absent; `ticker` from template `{{ config.ticker }}`, omitted when
  absent; follows a next-page URL from the response body; URL path `next_url`; next URLs stay on the
  configured API host; maximum 3 page(s).

## Write actions & risks

This connector is read-only. Read behavior: external Polygon.io API read of stock reference data
(tickers, dividends, splits).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=6.
