# Overview

Reads global financial data (dividends, stock splits, historical candlesticks, and commodity prices)
from the Finnworlds REST API.

Readable streams: `dividends`, `stock_splits`, `historical_candlestick`, `commodities`.

This connector is read-only; no write actions are declared.

Service API documentation: https://finnworlds.com/documentation/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.finnworlds.com/api/v1`; format `uri`;
  Finnworlds API base URL override for tests or proxies.
- `commodities` (optional, string); Comma-separated list of commodity names to fan out over for the
  commodities stream. Required for that stream (no auto-discovery fallback).
- `key` (required, secret, string); Finnworlds API key, sent as the 'key' query parameter on every
  request. Never logged.
- `tickers` (optional, string); Comma-separated list of stock ticker symbols to fan out over for the
  dividends, stock_splits, and historical_candlestick streams. Required for those three streams (no
  auto-discovery fallback); the commodities stream is unaffected.

Secret fields are redacted in logs and write previews: `key`.

Default configuration values: `base_url=https://api.finnworlds.com/api/v1`.

Authentication behavior:

- API key authentication in query parameter `key` using `secrets.key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/commodities`.

## Streams notes

Default pagination: single request; no pagination.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `dividends`: GET `/dividends` - records path `result.output.dividends`; incremental cursor `date`;
  formatted as `rfc3339`; fan-out; ids from config field `tickers`; id sent as query parameter
  `ticker`; stamps `ticker`.
- `stock_splits`: GET `/stocksplits` - records path `result.output.stocksplits`; incremental cursor
  `date`; formatted as `rfc3339`; fan-out; ids from config field `tickers`; id sent as query
  parameter `ticker`; stamps `ticker`.
- `historical_candlestick`: GET `/historicalcandlestick` - records path `result.output`; incremental
  cursor `date`; formatted as `rfc3339`; fan-out; ids from config field `tickers`; id sent as query
  parameter `ticker`; stamps `ticker`.
- `commodities`: GET `/commodities` - records path `result.output`; incremental cursor `datetime`;
  formatted as `rfc3339`; fan-out; ids from config field `commodities`; id sent as query parameter
  `commodity_name`; stamps `commodity_name`.

## Write actions & risks

This connector is read-only. Read behavior: external Finnworlds API read of global financial/market
data for the configured tickers/commodities.

## Known limits

- API coverage includes 4 stream-backed endpoint group(s).
