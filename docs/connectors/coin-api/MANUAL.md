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
  api_key (secret)

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

COMMAND SURFACE
  Run Coin API's declared streams and reverse-ETL actions.
  Usage: pm coin-api <command> [flags]
  Read streams
  Other Commands
    api get v1 assets asset-id - Documented GET /v1/assets/{asset_id} (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-assets-asset-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 assets icons size - Documented GET /v1/assets/icons/{size} (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-assets-icons-size]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 chains - Documented GET /v1/chains (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-chains]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 chains chain-id - Documented GET /v1/chains/{chain_id} (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-chains-chain-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 exchangerate asset-id-base asset-id-quote - Documented GET /v1/exchangerate/{asset_id_base}/{asset_id_quote} (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-exchangerate-asset-id-base-asset-id-quote]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 exchangerate asset-id-base asset-id-quote history - Documented GET /v1/exchangerate/{asset_id_base}/{asset_id_quote}/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-exchangerate-asset-id-base-asset-id-quote-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 exchangerate history periods - Documented GET /v1/exchangerate/history/periods (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-exchangerate-history-periods]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 exchanges exchange-id - Documented GET /v1/exchanges/{exchange_id} (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-exchanges-exchange-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 exchanges icons size - Documented GET /v1/exchanges/icons/{size} (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-exchanges-icons-size]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics asset current - Documented GET /v1/metrics/asset/current (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-metrics-asset-current]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics asset history - Documented GET /v1/metrics/asset/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-metrics-asset-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics asset listing - Documented GET /v1/metrics/asset/listing (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-metrics-asset-listing]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics exchange current - Documented GET /v1/metrics/exchange/current (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-metrics-exchange-current]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics exchange history - Documented GET /v1/metrics/exchange/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-metrics-exchange-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics exchange listing - Documented GET /v1/metrics/exchange/listing (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-metrics-exchange-listing]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics symbol current - Documented GET /v1/metrics/symbol/current (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-metrics-symbol-current]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics symbol history - Documented GET /v1/metrics/symbol/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-metrics-symbol-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 metrics symbol listing - Documented GET /v1/metrics/symbol/listing (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-metrics-symbol-listing]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 ohlcv exchanges exchange-id history - Documented GET /v1/ohlcv/exchanges/{exchange_id}/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-ohlcv-exchanges-exchange-id-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 ohlcv periods - Documented GET /v1/ohlcv/periods (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-ohlcv-periods]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 ohlcv symbol-id latest - Documented GET /v1/ohlcv/{symbol_id}/latest (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-ohlcv-symbol-id-latest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 options exchange-id current - Documented GET /v1/options/{exchange_id}/current (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-options-exchange-id-current]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 orderbooks symbol-id depth current - Documented GET /v1/orderbooks/{symbol_id}/depth/current (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-orderbooks-symbol-id-depth-current]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 orderbooks symbol-id history - Documented GET /v1/orderbooks/{symbol_id}/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-orderbooks-symbol-id-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 orderbooks3 current - Documented GET /v1/orderbooks3/current (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-orderbooks3-current]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 orderbooks3 symbol-id current - Documented GET /v1/orderbooks3/{symbol_id}/current (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-orderbooks3-symbol-id-current]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 quotes current - Documented GET /v1/quotes/current (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-quotes-current]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 quotes latest - Documented GET /v1/quotes/latest (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-quotes-latest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 quotes symbol-id history - Documented GET /v1/quotes/{symbol_id}/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-quotes-symbol-id-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 quotes symbol-id latest - Documented GET /v1/quotes/{symbol_id}/latest (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-quotes-symbol-id-latest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 symbols exchange-id active - Documented GET /v1/symbols/{exchange_id}/active (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-symbols-exchange-id-active]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 symbols exchange-id history - Documented GET /v1/symbols/{exchange_id}/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-symbols-exchange-id-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 symbols map exchange-id - Documented GET /v1/symbols/map/{exchange_id} (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-symbols-map-exchange-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 trades latest - Documented GET /v1/trades/latest (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-trades-latest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 trades symbol-id latest - Documented GET /v1/trades/{symbol_id}/latest (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v1-trades-symbol-id-latest]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 metrics asset history - Documented GET /v2/metrics/asset/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v2-metrics-asset-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 metrics asset listing - Documented GET /v2/metrics/asset/listing (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v2-metrics-asset-listing]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 metrics chain history - Documented GET /v2/metrics/chain/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v2-metrics-chain-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 metrics chain listing - Documented GET /v2/metrics/chain/listing (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v2-metrics-chain-listing]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 metrics exchange history - Documented GET /v2/metrics/exchange/history (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v2-metrics-exchange-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 metrics exchange listing - Documented GET /v2/metrics/exchange/listing (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v2-metrics-exchange-listing]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 metrics listing - Documented GET /v2/metrics/listing (not implemented) [intent=direct_read availability=not_implemented operation=coin-api.get.v2-metrics-listing]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    assets list - Run the assets ETL stream [intent=etl availability=implemented stream=assets]
    exchange rates list - Run the exchange rates ETL stream [intent=etl availability=implemented stream=exchange_rates]
    exchanges list - Run the exchanges ETL stream [intent=etl availability=implemented stream=exchanges]
    metrics listing list - Run the metrics listing ETL stream [intent=etl availability=implemented stream=metrics_listing]
    ohlcv historical data list - Run the ohlcv historical data ETL stream [intent=etl availability=implemented stream=ohlcv_historical_data]
    orderbook current list - Run the orderbook current ETL stream [intent=etl availability=implemented stream=orderbook_current]
    quotes current list - Run the quotes current ETL stream [intent=etl availability=implemented stream=quotes_current]
    symbols list - Run the symbols ETL stream [intent=etl availability=implemented stream=symbols]; notes: discrepancy=present-in-surface-absent-from-artifact
    trades historical data list - Run the trades historical data ETL stream [intent=etl availability=implemented stream=trades_historical_data]

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
