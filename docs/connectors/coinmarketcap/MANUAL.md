# pm connectors inspect coinmarketcap

```text
NAME
  pm connectors inspect coinmarketcap - CoinMarketCap connector manual

SYNOPSIS
  pm connectors inspect coinmarketcap
  pm connectors inspect coinmarketcap --json
  pm credentials add <name> --connector coinmarketcap [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads CoinMarketCap Pro API global market metrics, id/slug/symbol-keyed cryptocurrency detail and quote lookups, price conversion, fear-and-greed index, and altcoin season index. Read-only.

ICON
  id: coinmarketcap
  asset: icons/coinmarketcap.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://coinmarketcap.com/api/documentation/v1/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  altcoin_season_timeframe
  base_url
  convert
  cryptocurrency_ids
  historical_count
  historical_interval
  historical_time_end
  historical_time_start
  mode
  price_conversion_amount
  price_conversion_id
  price_conversion_symbol
  api_key (secret) (required)

ETL STREAMS
  global_metrics:
    primary key: active_cryptocurrencies
    fields: active_cryptocurrencies(integer), active_exchanges(integer), active_market_pairs(integer), btc_dominance(number), eth_dominance(number), last_updated(string), quote(object), total_cryptocurrencies(integer), total_exchanges(integer)
  global_metrics_quotes_historical:
    primary key: timestamp
    fields: active_cryptocurrencies(integer), active_exchanges(integer), active_market_pairs(integer), btc_dominance(number), eth_dominance(number), quote(object), timestamp(string)
  cryptocurrency_info:
    primary key: cmc_id
    fields: category(string), cmc_id(string), date_added(string), date_launched(string), description(string), id(integer), logo(string), name(string), notice(string), platform(object), slug(string), subreddit(string), symbol(string), tags(array), urls(object)
  cryptocurrency_quotes_latest:
    primary key: cmc_id
    fields: circulating_supply(number), cmc_id(string), cmc_rank(integer), id(integer), last_updated(string), max_supply(number), name(string), quote(object), slug(string), symbol(string), total_supply(number)
  price_conversion:
    primary key: id
    fields: amount(number), id(integer), last_updated(string), name(string), quote(object), symbol(string)
  fear_and_greed_latest:
    primary key: update_time
    fields: update_time(string), value(integer), value_classification(string)
  altcoin_season_index_latest:
    primary key: snapshot_time
    fields: altcoin_index(integer), altcoin_marketcap(number), snapshot_time(string)
  altcoin_season_index_historical:
    primary key: timestamp
    fields: altcoin_index(integer), altcoin_marketcap(number), timestamp(string)
  key_info:
    primary key: plan
    fields: plan(object), usage(object)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external CoinMarketCap Pro API read of aggregate global market metrics
  approval: none; read-only market-data API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect coinmarketcap

  # Inspect as structured JSON
  pm connectors inspect coinmarketcap --json

AGENT WORKFLOW
  - Run pm connectors inspect coinmarketcap before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
