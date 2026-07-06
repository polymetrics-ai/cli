# Overview

Reads CoinMarketCap Pro API global market metrics, id/slug/symbol-keyed cryptocurrency detail and
quote lookups, price conversion, fear-and-greed index, and altcoin season index. Read-only.

Readable streams: `global_metrics`, `global_metrics_quotes_historical`, `cryptocurrency_info`,
`cryptocurrency_quotes_latest`, `price_conversion`, `fear_and_greed_latest`,
`altcoin_season_index_latest`, `altcoin_season_index_historical`, `key_info`.

This connector is read-only; no write actions are declared.

Service API documentation: https://coinmarketcap.com/api/documentation/v1/.

## Auth setup

Connection fields:

- `altcoin_season_timeframe` (optional, string); default `7d`; Timeframe for the
  altcoin_season_index_historical stream: one of 7d, 30d, 90d.
- `api_key` (required, secret, string); CoinMarketCap Pro API key, sent on the X-CMC_PRO_API_KEY
  header. Never logged.
- `base_url` (optional, string); default `https://pro-api.coinmarketcap.com`; format `uri`;
  CoinMarketCap Pro API base URL override for tests or proxies.
- `convert` (optional, string); default `USD`; Comma-separated fiat/crypto symbols CoinMarketCap
  converts quote values into (the `convert` query param on quote-bearing endpoints). Defaults to
  USD, matching CoinMarketCap's own documented default.
- `cryptocurrency_ids` (optional, string); Comma-separated CoinMarketCap cryptocurrency IDs (e.g.
  "1,1027"), used as the `id` query param for the id/slug/symbol-keyed detail streams
  (cryptocurrency_info, cryptocurrency_quotes_latest). Required for those two streams only; leave
  unset if only reading other streams.
- `historical_count` (optional, string); Optional result count (1-10000, CoinMarketCap default 10)
  for the global_metrics_quotes_historical stream's `count` query param.
- `historical_interval` (optional, string); Optional interval (e.g. "1d") for the
  global_metrics_quotes_historical stream's `interval` query param.
- `historical_time_end` (optional, string); Optional Unix-or-ISO-8601 end timestamp for the
  global_metrics_quotes_historical stream's `time_end` query param.
- `historical_time_start` (optional, string); Optional Unix-or-ISO-8601 start timestamp for the
  global_metrics_quotes_historical stream's `time_start` query param.
- `mode` (optional, string).
- `price_conversion_amount` (optional, string); Amount to convert for the price_conversion stream's
  `amount` query param (required only when reading that stream).
- `price_conversion_id` (optional, string); CoinMarketCap cryptocurrency ID to convert FROM, for the
  price_conversion stream's `id` query param (use this OR price_conversion_symbol, not both).
- `price_conversion_symbol` (optional, string); Cryptocurrency symbol to convert FROM, for the
  price_conversion stream's `symbol` query param (use this OR price_conversion_id, not both).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `altcoin_season_timeframe=7d`,
`base_url=https://pro-api.coinmarketcap.com`, `convert=USD`.

Authentication behavior:

- API key authentication in `X-CMC_PRO_API_KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/fiat/map` with query `limit`=`1`; `start`=`1`.

## Streams notes

Default pagination: single request; no pagination.

- `global_metrics`: GET `/v1/global-metrics/quotes/latest` - records path `data`; query `convert`
  from template `{{ config.convert }}`, omitted when absent.
- `global_metrics_quotes_historical`: GET `/v1/global-metrics/quotes/historical` - records path
  `data.quotes`; query `convert` from template `{{ config.convert }}`, omitted when absent; `count`
  from template `{{ config.historical_count }}`, omitted when absent; `interval` from template `{{
  config.historical_interval }}`, omitted when absent; `time_end` from template `{{
  config.historical_time_end }}`, omitted when absent; `time_start` from template `{{
  config.historical_time_start }}`, omitted when absent.
- `cryptocurrency_info`: GET `/v2/cryptocurrency/info` - records path `data`; flattens keyed
  objects; key field `cmc_id`; query `id`=`{{ config.cryptocurrency_ids }}`.
- `cryptocurrency_quotes_latest`: GET `/v3/cryptocurrency/quotes/latest` - records path `data`;
  flattens keyed objects; key field `cmc_id`; query `convert` from template `{{ config.convert }}`,
  omitted when absent; `id`=`{{ config.cryptocurrency_ids }}`.
- `price_conversion`: GET `/v2/tools/price-conversion` - records path `data`; query `amount`=`{{
  config.price_conversion_amount }}`; `convert` from template `{{ config.convert }}`, omitted when
  absent; `id` from template `{{ config.price_conversion_id }}`, omitted when absent; `symbol` from
  template `{{ config.price_conversion_symbol }}`, omitted when absent.
- `fear_and_greed_latest`: GET `/v3/fear-and-greed/latest` - records path `data`.
- `altcoin_season_index_latest`: GET `/v1/altcoin-season-index/latest` - records path `data`.
- `altcoin_season_index_historical`: GET `/v1/altcoin-season-index/historical` - records path
  `data.points`; query `timeframe` from template `{{ config.altcoin_season_timeframe }}`, default
  `7d`.
- `key_info`: GET `/v1/key/info` - records path `data`.

## Write actions & risks

This connector is read-only. Read behavior: external CoinMarketCap Pro API read of aggregate global
market metrics.

## Known limits

- Batch defaults: read_page_size=1.
- API coverage includes 9 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=3, duplicate_of=1, non_data_endpoint=1, out_of_scope=42.
