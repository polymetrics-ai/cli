---
name: pm-alpha-vantage
description: Alpha Vantage connector knowledge and safe action guide.
---

# pm-alpha-vantage

## Purpose

Reads daily, weekly, monthly, and intraday OHLCV time series plus latest global quotes through fixed Alpha Vantage query operations.

## Icon

- id: alpha-vantage
- asset: icons/alpha-vantage.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- adjusted
- interval
- outputsize
- symbol (required)
- api_key (secret) (required)

## ETL Streams

- time_series_daily:
  - primary key: symbol, date
  - cursor: date
  - fields: close(number), date(string), high(number), low(number), open(number), symbol(string), volume(integer)
- time_series_weekly:
  - primary key: symbol, date
  - cursor: date
  - fields: close(number), date(string), high(number), low(number), open(number), symbol(string), volume(integer)
- time_series_monthly:
  - primary key: symbol, date
  - cursor: date
  - fields: close(number), date(string), high(number), low(number), open(number), symbol(string), volume(integer)
- time_series_intraday:
  - primary key: symbol, date
  - cursor: date
  - fields: close(number), date(string), high(number), low(number), open(number), symbol(string), volume(integer)
- global_quote:
  - primary key: symbol, latest_trading_day
  - cursor: latest_trading_day
  - fields: change(number), change_percent(string), high(number), latest_trading_day(string), low(number), open(number), previous_close(number), price(number), symbol(string), volume(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded Alpha Vantage query reads use the fixed provider origin and declared API-key query authentication.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect alpha-vantage
```

### Inspect as structured JSON

```bash
pm connectors inspect alpha-vantage --json
```

## Agent Rules

- Run pm connectors inspect alpha-vantage before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
