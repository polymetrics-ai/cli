---
name: pm-marketstack
description: Marketstack connector knowledge and safe action guide.
---

# pm-marketstack

## Purpose

Reads Marketstack exchanges, tickers, end-of-day prices, splits, and dividends through the Marketstack REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- start_date
- symbols
- api_key (secret)

## ETL Streams

- exchanges:
  - primary key: mic
  - fields: acronym(), city(), country(), country_code(), currency_code(), currency_name(), currency_symbol(), mic(), name(), timezone(), timezone_abbr(), website()
- tickers:
  - primary key: symbol
  - fields: has_eod(), has_intraday(), name(), stock_exchange_mic(), stock_exchange_name(), symbol()
- eod:
  - primary key: symbol, date
  - cursor: date
  - fields: adj_close(), adj_high(), adj_low(), adj_open(), adj_volume(), close(), date(), dividend(), exchange(), high(), low(), open(), split_factor(), symbol(), volume()
- splits:
  - primary key: symbol, date
  - cursor: date
  - fields: date(), split_factor(), symbol()
- dividends:
  - primary key: symbol, date
  - cursor: date
  - fields: date(), dividend(), symbol()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Marketstack API read of financial market data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect marketstack
```

### Inspect as structured JSON

```bash
pm connectors inspect marketstack --json
```

## Agent Rules

- Run pm connectors inspect marketstack before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
