# Overview

Reads Twelve Data time series, quotes, stocks, and forex pair reference data.

Readable streams: `time_series`, `quote`, `stocks`, `forex_pairs`.

This connector is read-only; no write actions are declared.

Service API documentation: https://twelvedata.com/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Twelve Data API key, sent as the apikey query parameter on
  every request. Never logged.
- `base_url` (optional, string); default `https://api.twelvedata.com`; format `uri`; Twelve Data API
  base URL override for tests or proxies.
- `interval` (optional, string); default `1day`; Time series bar interval (e.g. 1min, 1day) for the
  time_series stream.
- `output_size` (optional, string); default `100`; Number of time series bars to request per read
  (1-5000).
- `symbol` (optional, string); default `AAPL`; Ticker symbol the time_series/quote streams are
  scoped to.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.twelvedata.com`, `interval=1day`,
`output_size=100`, `symbol=AAPL`.

Authentication behavior:

- API key authentication in query parameter `apikey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/quote` with query `symbol`=`AAPL`.

## Streams notes

Default pagination: single request; no pagination.

- `time_series`: GET `/time_series` - records path `values`; query `interval`=`{{ config.interval
  }}`; `outputsize`=`{{ config.output_size }}`; `symbol`=`{{ config.symbol }}`; computed output
  fields `symbol`.
- `quote`: GET `/quote` - records path `.`; query `symbol`=`{{ config.symbol }}`; computed output
  fields `symbol`.
- `stocks`: GET `/stocks` - records path `data`.
- `forex_pairs`: GET `/forex_pairs` - records path `data`.

## Write actions & risks

This connector is read-only. Read behavior: external Twelve Data API read of market time series,
quote, and reference data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
