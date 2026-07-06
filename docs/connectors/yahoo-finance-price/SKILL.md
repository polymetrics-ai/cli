---
name: pm-yahoo-finance-price
description: Yahoo Finance Price connector knowledge and safe action guide.
---

# pm-yahoo-finance-price

## Purpose

Reads public Yahoo Finance chart prices and flattens them into OHLCV records. Read-only. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/yahoo-finance-price.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.yahoofinanceapi.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- interval
- mode
- range
- symbol

## ETL Streams

- prices:
  - primary key: symbol, timestamp
  - cursor: timestamp
  - fields: adjclose(), close(), currency(), high(), low(), open(), symbol(), timestamp(), volume()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Yahoo Finance Price API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect yahoo-finance-price
```

### Inspect as structured JSON

```bash
pm connectors inspect yahoo-finance-price --json
```

## Agent Rules

- Run pm connectors inspect yahoo-finance-price before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
