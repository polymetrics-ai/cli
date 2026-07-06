# Overview

Reads CoinAPI market data: symbols, exchanges, assets, exchange rates, current quotes, current order
book, the metrics catalog, and historical OHLCV and trades for a configured symbol via the CoinAPI
REST API.

Readable streams: `symbols`, `exchanges`, `assets`, `ohlcv_historical_data`,
`trades_historical_data`, `exchange_rates`, `quotes_current`, `orderbook_current`,
`metrics_listing`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.coinapi.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); CoinAPI key, sent as the X-CoinAPI-Key request header. Used
  only for auth; never logged.
- `asset_id_base` (optional, string); CoinAPI base asset identifier (e.g. BTC), required for the
  exchange_rates stream.
- `base_url` (optional, string); format `uri`; CoinAPI base URL override for tests or proxies. When
  unset, environment selects the production or sandbox CoinAPI host.
- `end_date` (optional, string); format `date-time`; Optional ISO-8601 upper bound (time_end) for
  historical streams.
- `environment` (optional, string); default `production`; allowed values `production`, `sandbox`;
  Selects the CoinAPI host when base_url is not set: production (https://rest.coinapi.io) or sandbox
  (https://rest-sandbox.coinapi.io).
- `limit` (optional, string); default `100`; Records per page for historical streams (1-100000).
- `period` (optional, string); CoinAPI OHLCV period identifier (e.g. 1DAY), required for the
  ohlcv_historical_data stream.
- `start_date` (optional, string); format `date-time`; ISO-8601 lower bound for historical streams;
  only records at or after this time are read on a fresh sync.
- `symbol_id` (optional, string); CoinAPI symbol identifier (e.g. BITSTAMP_SPOT_BTC_USD), required
  for the ohlcv_historical_data, trades_historical_data, quotes_current, and orderbook_current
  streams.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `environment=production`, `limit=100`.

Authentication behavior:

- API key authentication in `X-CoinAPI-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/exchanges`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `ohlcv_historical_data`, `trades_historical_data`; none: `symbols`,
`exchanges`, `assets`, `exchange_rates`, `quotes_current`, `orderbook_current`, `metrics_listing`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `symbols`: GET `/v1/symbols` - records at response root.
- `exchanges`: GET `/v1/exchanges` - records at response root.
- `assets`: GET `/v1/assets` - records at response root.
- `ohlcv_historical_data`: GET `/v1/ohlcv/{{ config.symbol_id }}/history` - records at response
  root; query `limit`=`{{ config.limit }}`; `period_id`=`{{ config.period }}`; `time_end` from
  template `{{ config.end_date }}`, omitted when absent; cursor pagination; cursor parameter
  `time_start`; next cursor from last record field `time_period_start`; incremental cursor
  `time_period_start`; sent as `time_start`; formatted as `rfc3339`; initial lower bound from
  `start_date`; computed output fields `period_id`, `symbol_id`.
- `trades_historical_data`: GET `/v1/trades/{{ config.symbol_id }}/history` - records at response
  root; query `limit`=`{{ config.limit }}`; `time_end` from template `{{ config.end_date }}`,
  omitted when absent; cursor pagination; cursor parameter `time_start`; next cursor from last
  record field `time_exchange`; incremental cursor `time_exchange`; sent as `time_start`; formatted
  as `rfc3339`; initial lower bound from `start_date`; computed output fields `symbol_id`.
- `exchange_rates`: GET `/v1/exchangerate/{{ config.asset_id_base }}` - records path `rates`;
  computed output fields `asset_id_base`, `asset_id_quote`, `rate`, `time`.
- `quotes_current`: GET `/v1/quotes/{{ config.symbol_id }}/current` - records path `.`.
- `orderbook_current`: GET `/v1/orderbooks/{{ config.symbol_id }}/current` - records path `.`.
- `metrics_listing`: GET `/v1/metrics/listing` - records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external CoinAPI REST API read of public market data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 9 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=6, non_data_endpoint=4, out_of_scope=32.
