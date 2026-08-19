---
name: pm-finnworlds
description: Finnworlds connector knowledge and safe action guide.
---

# pm-finnworlds

## Purpose

Reads global financial data (dividends, stock splits, historical candlesticks, and commodity prices) from the Finnworlds REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- commodities
- tickers
- key (secret) (required)

## ETL Streams

- dividends:
  - primary key: ticker, date
  - cursor: date
  - fields: date(string), dividend_rate(string), ticker(string)
- stock_splits:
  - primary key: ticker, date
  - cursor: date
  - fields: date(string), stock_split(string), ticker(string)
- historical_candlestick:
  - primary key: ticker, date
  - cursor: date
  - fields: adjusted_close(string), close(string), closetime(integer), date(string), high(string), low(string), open(string), opentime(integer), ticker(string), trade_volume(string)
- commodities:
  - primary key: commodity_name, datetime
  - cursor: datetime
  - fields: commodity_name(string), commodity_price(string), commodity_unit(string), datetime(string), percentage_day(string), percentage_month(string), percentage_week(string), percentage_year(string), price_change_day(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Finnworlds API read of global financial/market data for the configured tickers/commodities
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect finnworlds
```

### Inspect as structured JSON

```bash
pm connectors inspect finnworlds --json
```

## Agent Rules

- Run pm connectors inspect finnworlds before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
