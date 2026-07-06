---
name: pm-twelve-data
description: Twelve Data connector knowledge and safe action guide.
---

# pm-twelve-data

## Purpose

Reads Twelve Data time series, quotes, stocks, and forex pair reference data.

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
- interval
- output_size
- symbol
- api_key (secret)

## ETL Streams

- time_series:
  - primary key: symbol, datetime
  - cursor: datetime
  - fields: close(), datetime(), high(), low(), open(), symbol(), volume()
- quote:
  - primary key: symbol
  - fields: close(), currency(), name(), symbol()
- stocks:
  - primary key: symbol
  - fields: currency(), name(), symbol()
- forex_pairs:
  - primary key: symbol
  - fields: currency(), name(), symbol()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Twelve Data API read of market time series, quote, and reference data
- approval: none; read-only, no reverse-ETL writes implemented by legacy
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect twelve-data
```

### Inspect as structured JSON

```bash
pm connectors inspect twelve-data --json
```

## Agent Rules

- Run pm connectors inspect twelve-data before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
