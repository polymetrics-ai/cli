---
name: pm-coinmarketcap
description: CoinMarketCap connector knowledge and safe action guide.
---

# pm-coinmarketcap

## Purpose

Reads CoinMarketCap Pro API global market metrics, id/slug/symbol-keyed cryptocurrency detail and quote lookups, price conversion, fear-and-greed index, and altcoin season index. Read-only.

## Icon

- asset: icons/coinmarketcap.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://coinmarketcap.com/api/documentation/v1/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- altcoin_season_timeframe
- base_url
- convert
- cryptocurrency_ids
- historical_count
- historical_interval
- historical_time_end
- historical_time_start
- mode
- price_conversion_amount
- price_conversion_id
- price_conversion_symbol
- api_key (secret)

## ETL Streams

- global_metrics:
  - primary key: active_cryptocurrencies
  - fields: active_cryptocurrencies(), active_exchanges(), active_market_pairs(), btc_dominance(), eth_dominance(), last_updated(), quote(), total_cryptocurrencies(), total_exchanges()
- global_metrics_quotes_historical:
  - primary key: timestamp
  - fields: active_cryptocurrencies(), active_exchanges(), active_market_pairs(), btc_dominance(), eth_dominance(), quote(), timestamp()
- cryptocurrency_info:
  - primary key: cmc_id
  - fields: category(), cmc_id(), date_added(), date_launched(), description(), id(), logo(), name(), notice(), platform(), slug(), subreddit(), symbol(), tags(), urls()
- cryptocurrency_quotes_latest:
  - primary key: cmc_id
  - fields: circulating_supply(), cmc_id(), cmc_rank(), id(), last_updated(), max_supply(), name(), quote(), slug(), symbol(), total_supply()
- price_conversion:
  - primary key: id
  - fields: amount(), id(), last_updated(), name(), quote(), symbol()
- fear_and_greed_latest:
  - primary key: update_time
  - fields: update_time(), value(), value_classification()
- altcoin_season_index_latest:
  - primary key: snapshot_time
  - fields: altcoin_index(), altcoin_marketcap(), snapshot_time()
- altcoin_season_index_historical:
  - primary key: timestamp
  - fields: altcoin_index(), altcoin_marketcap(), timestamp()
- key_info:
  - primary key: plan
  - fields: plan(), usage()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external CoinMarketCap Pro API read of aggregate global market metrics
- approval: none; read-only market-data API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect coinmarketcap
```

### Inspect as structured JSON

```bash
pm connectors inspect coinmarketcap --json
```

## Agent Rules

- Run pm connectors inspect coinmarketcap before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
