# Overview

Reads Marketstack exchanges, tickers, end-of-day prices, splits, and dividends through the
Marketstack REST API.

Readable streams: `exchanges`, `tickers`, `eod`, `splits`, `dividends`.

This connector is read-only; no write actions are declared.

Service API documentation: https://marketstack.com/documentation.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Marketstack API access key, sent as the access_key query
  parameter. Never logged.
- `base_url` (optional, string); default `https://api.marketstack.com/v1`; format `uri`; Marketstack
  API base URL override for tests or proxies.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound truncated to YYYY-MM-DD;
  sent as date_from on eod/splits/dividends streams.
- `symbols` (optional, string); Comma-separated ticker symbols; sent as the symbols query param on
  eod/splits/dividends streams.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.marketstack.com/v1`.

Authentication behavior:

- API key authentication in query parameter `access_key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/exchanges`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `exchanges`: GET `/exchanges` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; computed output fields `currency_code`,
  `currency_name`, `currency_symbol`, `timezone`, `timezone_abbr`.
- `tickers`: GET `/tickers` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; computed output fields `stock_exchange_mic`,
  `stock_exchange_name`.
- `eod`: GET `/eod` - records path `data`; query `symbols` from template `{{ config.symbols }}`,
  omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
  page size 100; incremental cursor `date`; sent as `date_from`; formatted as YYYY-MM-DD date;
  initial lower bound from `start_date`.
- `splits`: GET `/splits` - records path `data`; query `symbols` from template `{{ config.symbols
  }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter
  `limit`; page size 100; incremental cursor `date`; sent as `date_from`; formatted as YYYY-MM-DD
  date; initial lower bound from `start_date`.
- `dividends`: GET `/dividends` - records path `data`; query `symbols` from template `{{
  config.symbols }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit
  parameter `limit`; page size 100; incremental cursor `date`; sent as `date_from`; formatted as
  YYYY-MM-DD date; initial lower bound from `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external Marketstack API read of financial market data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
