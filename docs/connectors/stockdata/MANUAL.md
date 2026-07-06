# pm connectors inspect stockdata

```text
NAME
  pm connectors inspect stockdata - StockData connector manual

SYNOPSIS
  pm connectors inspect stockdata
  pm connectors inspect stockdata --json
  pm credentials add <name> --connector stockdata [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads StockData market quotes, prices, splits, dividends, news, entity stats, entities, and source metadata through the StockData API.

ICON
  asset: icons/stockdata.svg
  source: official
  review_status: official_verified
  review_url: https://www.stockdata.org/documentation

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  countries
  date_from
  date_to
  entity_search
  entity_types
  group_by
  group_domains
  interval
  language
  news_uuids
  published_after
  published_before
  published_on
  sentiment_gte
  sentiment_lte
  symbols
  api_token (secret)

ETL STREAMS
  tickers:
    primary key: symbol
    fields: exchange(), name(), symbol()
  eod_prices:
    primary key: ticker, date
    cursor: date
    fields: close(), date(), ticker()
  intraday_prices:
    primary key: ticker, date
    cursor: date
    fields: close(), date(), ticker()
  news:
    primary key: url
    fields: published_at(), title(), url()
  quotes:
    primary key: ticker
    fields: day_high(), day_low(), day_open(), exchange(), last_trade_time(), name(), previous_close_price(), price(), ticker()
  intraday_adjusted_prices:
    primary key: ticker, date
    cursor: date
    fields: close(), date(), ticker(), volume()
  splits:
    primary key: ticker, date
    cursor: date
    fields: date(), split_factor(), split_ratio(), ticker()
  dividends:
    primary key: ticker, date
    cursor: date
    fields: currency(), date(), dividend(), ticker()
  news_similar:
    primary key: uuid
    fields: entities(), language(), published_at(), source(), title(), url(), uuid()
  news_by_uuid:
    primary key: uuid
    fields: entities(), language(), published_at(), source(), title(), url(), uuid()
  news_stats_intraday:
    primary key: date
    fields: data(), date()
  news_stats_aggregation:
    primary key: key
    fields: key(), sentiment_avg(), total_documents()
  news_stats_trending:
    primary key: key
    fields: key(), score(), sentiment_avg(), total_documents()
  entity_search:
    primary key: symbol
    fields: country(), exchange(), industry(), name(), symbol(), type()
  news_sources:
    primary key: source_id
    fields: domain(), language(), source_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external StockData API read of ticker, quote, price, split, dividend, news, entity, stats, and source data
  approval: none; read-only public market data API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect stockdata

  # Inspect as structured JSON
  pm connectors inspect stockdata --json

AGENT WORKFLOW
  - Run pm connectors inspect stockdata before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
