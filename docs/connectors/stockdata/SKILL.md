---
name: pm-stockdata
description: StockData connector knowledge and safe action guide.
---

# pm-stockdata

## Purpose

Reads StockData market quotes, prices, splits, dividends, news, entity stats, entities, and source metadata through the StockData API.

## Icon

- id: stockdata
- asset: icons/stockdata.svg
- source: official
- review_status: official_verified
- review_url: https://www.stockdata.org/documentation

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- countries
- date_from
- date_to
- entity_search
- entity_types
- group_by
- group_domains
- interval
- language
- news_uuids
- published_after
- published_before
- published_on
- sentiment_gte
- sentiment_lte
- symbols
- api_token (secret) (required)

## ETL Streams

- tickers:
  - primary key: symbol
  - fields: exchange(string), name(string), symbol(string)
- eod_prices:
  - primary key: ticker, date
  - cursor: date
  - fields: close(number), date(string), ticker(string)
- intraday_prices:
  - primary key: ticker, date
  - cursor: date
  - fields: close(number), date(string), ticker(string)
- news:
  - primary key: url
  - fields: published_at(string), title(string), url(string)
- quotes:
  - primary key: ticker
  - fields: day_high(number), day_low(number), day_open(number), exchange(string), last_trade_time(string), name(string), previous_close_price(number), price(number), ticker(string)
- intraday_adjusted_prices:
  - primary key: ticker, date
  - cursor: date
  - fields: close(number), date(string), ticker(string), volume(number)
- splits:
  - primary key: ticker, date
  - cursor: date
  - fields: date(string), split_factor(number), split_ratio(string), ticker(string)
- dividends:
  - primary key: ticker, date
  - cursor: date
  - fields: currency(string), date(string), dividend(number), ticker(string)
- news_similar:
  - primary key: uuid
  - fields: entities(array), language(string), published_at(string), source(string), title(string), url(string), uuid(string)
- news_by_uuid:
  - primary key: uuid
  - fields: entities(array), language(string), published_at(string), source(string), title(string), url(string), uuid(string)
- news_stats_intraday:
  - primary key: date
  - fields: data(array), date(string)
- news_stats_aggregation:
  - primary key: key
  - fields: key(string), sentiment_avg(number), total_documents(number)
- news_stats_trending:
  - primary key: key
  - fields: key(string), score(number), sentiment_avg(number), total_documents(number)
- entity_search:
  - primary key: symbol
  - fields: country(string), exchange(string), industry(string), name(string), symbol(string), type(string)
- news_sources:
  - primary key: source_id
  - fields: domain(string), language(string), source_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external StockData API read of ticker, quote, price, split, dividend, news, entity, stats, and source data
- approval: none; read-only public market data API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect stockdata
```

### Inspect as structured JSON

```bash
pm connectors inspect stockdata --json
```

## Agent Rules

- Run pm connectors inspect stockdata before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
