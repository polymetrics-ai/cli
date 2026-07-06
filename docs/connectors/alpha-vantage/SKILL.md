---
name: pm-alpha-vantage
description: Alpha Vantage connector knowledge and safe action guide.
---

# pm-alpha-vantage

## Purpose

Reads Alpha Vantage daily, weekly, monthly, and intraday OHLCV time series plus the latest global quote for a configured stock symbol. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

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
- base_url
- interval
- mode
- outputsize
- symbol
- api_key (secret)

## ETL Streams

- time_series_daily:
  - primary key: symbol, date
  - cursor: date
  - fields: close(), date(), high(), low(), open(), symbol(), volume()
- time_series_weekly:
  - primary key: symbol, date
  - cursor: date
  - fields: close(), date(), high(), low(), open(), symbol(), volume()
- time_series_monthly:
  - primary key: symbol, date
  - cursor: date
  - fields: close(), date(), high(), low(), open(), symbol(), volume()
- time_series_intraday:
  - primary key: symbol, date
  - cursor: date
  - fields: close(), date(), high(), low(), open(), symbol(), volume()
- global_quote:
  - primary key: symbol, latest_trading_day
  - cursor: latest_trading_day
  - fields: change(), change_percent(), high(), latest_trading_day(), low(), open(), previous_close(), price(), symbol(), volume()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Alpha Vantage API reads performed by the legacy connector via a Tier-2 hook
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
