# pm connectors inspect coin-api

```text
NAME
  pm connectors inspect coin-api - Coin API connector manual

SYNOPSIS
  pm connectors inspect coin-api
  pm connectors inspect coin-api --json
  pm credentials add <name> --connector coin-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads CoinAPI market data: symbols, exchanges, assets, exchange rates, current quotes, current order book, the metrics catalog, and historical OHLCV and trades for a configured symbol via the CoinAPI REST API.

ICON
  id: coinapi
  asset: icons/coinapi.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.coinapi.io/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  asset_id_base
  base_url
  end_date
  environment
  limit
  period
  start_date
  symbol_id
  api_key (secret) (required)

ETL STREAMS
  symbols:
    primary key: symbol_id
    fields: asset_id_base(string), asset_id_quote(string), data_end(string), data_start(string), exchange_id(string), symbol_id(string), symbol_type(string)
  exchanges:
    primary key: exchange_id
    fields: data_quote_end(string), data_quote_start(string), data_symbols_count(integer), exchange_id(string), name(string), website(string)
  assets:
    primary key: asset_id
    fields: asset_id(string), data_end(string), data_start(string), name(string), price_usd(number), type_is_crypto(integer)
  ohlcv_historical_data:
    primary key: symbol_id, time_period_start
    cursor: time_period_start
    fields: period_id(string), price_close(number), price_high(number), price_low(number), price_open(number), symbol_id(string), time_close(string), time_open(string), time_period_end(string), time_period_start(string), trades_count(integer), volume_traded(number)
  trades_historical_data:
    primary key: symbol_id, uuid
    cursor: time_exchange
    fields: price(number), size(number), symbol_id(string), taker_side(string), time_coinapi(string), time_exchange(string), uuid(string)
  exchange_rates:
    primary key: asset_id_base, asset_id_quote
    fields: asset_id_base(string), asset_id_quote(string), rate(number), time(string)
  quotes_current:
    primary key: symbol_id
    fields: ask_price(number), ask_size(number), bid_price(number), bid_size(number), symbol_id(string), time_coinapi(string), time_exchange(string)
  orderbook_current:
    primary key: symbol_id
    fields: asks(array), bids(array), symbol_id(string), time_coinapi(string), time_exchange(string)
  metrics_listing:
    primary key: metric_id
    fields: description(string), metric_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external CoinAPI REST API read of public market data
  approval: none; read-only market-data API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect coin-api

  # Inspect as structured JSON
  pm connectors inspect coin-api --json

AGENT WORKFLOW
  - Run pm connectors inspect coin-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
