# Overview

Reads Alpha Vantage daily, weekly, monthly, and intraday OHLCV time series plus the latest global
quote for a configured stock symbol.

Readable streams: `time_series_daily`, `time_series_weekly`, `time_series_monthly`,
`time_series_intraday`, `global_quote`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.alphavantage.co/documentation/.

## Auth setup

Connection fields:

- `adjusted` (optional, string); Whether to return adjusted data. Only applicable to intraday
  endpoints.
- `api_key` (required, secret, string); API Key.
- `base_url` (optional, string).
- `interval` (optional, string); Time-series data point interval. Required for intraday endpoints.
- `mode` (optional, string).
- `outputsize` (optional, string); Whether to return full or compact data (the last 100 data
  points).
- `symbol` (required, string); Stock symbol (with exchange code).

Secret fields are redacted in logs and write previews: `api_key`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `time_series_daily`: GET connector-managed request path - records path `data`; incremental cursor
  `date`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `time_series_weekly`: GET connector-managed request path - records path `data`; incremental cursor
  `date`; formatted as `rfc3339`; records at or before the lower bound are filtered client-side.
- `time_series_monthly`: GET connector-managed request path - records path `data`; incremental
  cursor `date`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `time_series_intraday`: GET connector-managed request path - records path `data`; incremental
  cursor `date`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.
- `global_quote`: GET connector-managed request path - records path `data`; incremental cursor
  `latest_trading_day`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `time_series_daily`, `time_series_weekly`,
  `time_series_monthly`, `time_series_intraday`, `global_quote`.
