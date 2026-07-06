# Overview

Reads public Yahoo Finance chart prices and flattens them into OHLCV records. Read-only.

Readable streams: `prices`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.yahoofinanceapi.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://query1.finance.yahoo.com`; Yahoo Finance chart API
  base URL.
- `interval` (optional, string); The interval of between prices queried.
- `mode` (optional, string).
- `range` (optional, string); The range of prices to be queried.
- `symbol` (optional, string); default `AAPL`; Ticker symbol to query.

Default configuration values: `base_url=https://query1.finance.yahoo.com`, `symbol=AAPL`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks use a connector-managed request.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `prices`: GET connector-managed request path - records path `data`; incremental cursor
  `timestamp`; formatted as `rfc3339`; records at or before the lower bound are filtered
  client-side.

## Write actions & risks

This connector is read-only; no reverse-ETL write actions are declared.

## Known limits

- API coverage includes 1 stream-backed endpoint group(s).
- Client-side incremental filtering is used for: `prices`.
