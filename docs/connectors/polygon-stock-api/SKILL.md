---
name: pm-polygon-stock-api
description: Polygon Stock API connector knowledge and safe action guide.
---

# pm-polygon-stock-api

## Purpose

Reads Polygon.io stock tickers, dividends, and splits through the Polygon.io reference REST API.

## Icon

- asset: icons/polygon.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://polygon.io/docs/stocks/getting-started

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- active
- base_url
- ex_dividend_date
- execution_date
- locale
- market
- mode
- order
- page_size
- sort
- ticker
- type
- api_key (secret)

## ETL Streams

- tickers:
  - primary key: ticker
  - fields: active(), currency_name(), locale(), market(), name(), primary_exchange(), ticker()
- dividends:
  - primary key: id
  - cursor: ex_dividend_date
  - fields: cash_amount(), currency(), ex_dividend_date(), id(), ticker()
- splits:
  - primary key: id
  - cursor: execution_date
  - fields: execution_date(), id(), split_from(), split_to(), ticker()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Polygon.io API read of stock reference data (tickers, dividends, splits)
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect polygon-stock-api
```

### Inspect as structured JSON

```bash
pm connectors inspect polygon-stock-api --json
```

## Agent Rules

- Run pm connectors inspect polygon-stock-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
