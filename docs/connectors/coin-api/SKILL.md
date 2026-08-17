---
name: pm-coin-api
description: Coin API connector knowledge and safe action guide.
---

# pm-coin-api

## Purpose

Reads CoinAPI market data: symbols, exchanges, assets, exchange rates, current quotes, current order book, the metrics catalog, and historical OHLCV and trades for a configured symbol via the CoinAPI REST API.

## Icon

- asset: icons/coinapi.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.coinapi.io/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- asset_id_base
- base_url
- end_date
- environment
- limit
- period
- start_date
- symbol_id
- api_key (secret)

## ETL Streams

- symbols:
  - primary key: symbol_id
  - fields: asset_id_base(), asset_id_quote(), data_end(), data_start(), exchange_id(), symbol_id(), symbol_type()
- exchanges:
  - primary key: exchange_id
  - fields: data_quote_end(), data_quote_start(), data_symbols_count(), exchange_id(), name(), website()
- assets:
  - primary key: asset_id
  - fields: asset_id(), data_end(), data_start(), name(), price_usd(), type_is_crypto()
- ohlcv_historical_data:
  - primary key: symbol_id, time_period_start
  - cursor: time_period_start
  - fields: period_id(), price_close(), price_high(), price_low(), price_open(), symbol_id(), time_close(), time_open(), time_period_end(), time_period_start(), trades_count(), volume_traded()
- trades_historical_data:
  - primary key: symbol_id, uuid
  - cursor: time_exchange
  - fields: price(), size(), symbol_id(), taker_side(), time_coinapi(), time_exchange(), uuid()
- exchange_rates:
  - primary key: asset_id_base, asset_id_quote
  - fields: asset_id_base(), asset_id_quote(), rate(), time()
- quotes_current:
  - primary key: symbol_id
  - fields: ask_price(), ask_size(), bid_price(), bid_size(), symbol_id(), time_coinapi(), time_exchange()
- orderbook_current:
  - primary key: symbol_id
  - fields: asks(), bids(), symbol_id(), time_coinapi(), time_exchange()
- metrics_listing:
  - primary key: metric_id
  - fields: description(), metric_id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external CoinAPI REST API read of public market data
- approval: none; read-only market-data API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect coin-api
```

### Inspect as structured JSON

```bash
pm connectors inspect coin-api --json
```

## Agent Rules

- Run pm connectors inspect coin-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
