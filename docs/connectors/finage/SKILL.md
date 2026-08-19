---
name: pm-finage
description: Finage connector knowledge and safe action guide.
---

# pm-finage

## Purpose

Reads Finage US market data: most active stocks, top gainers and losers, sector performance, delisted companies, and per-symbol market news via the Finage REST API.

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
- calendar_from
- calendar_to
- mode
- symbols
- api_key (secret) (required)

## ETL Streams

- most_active_us_stocks:
  - primary key: symbol
  - fields: change(number), change_percentage(string), company_name(string), price(string), symbol(string)
- most_gainers:
  - primary key: symbol
  - fields: change(number), change_percentage(string), company_name(string), price(string), symbol(string)
- most_losers:
  - primary key: symbol
  - fields: change(number), change_percentage(string), company_name(string), price(string), symbol(string)
- sector_performance:
  - primary key: sector
  - fields: change_percentage(string), sector(string)
- delisted_companies:
  - primary key: symbol
  - fields: company_name(string), delisted_date(string), exchange(string), ipo_date(string), symbol(string)
- market_news:
  - primary key: url
  - fields: date(string), description(string), source(string), symbol(string), title(string), url(string)
- earnings_calendar:
  - primary key: symbol, date
  - fields: date(string), eps(number), estimated_eps(number), estimated_revenue(number), revenue(number), symbol(string), time(string)
- ipo_calendar:
  - primary key: symbol, date
  - fields: company(string), date(string), exchange(string), market_cap(number), price_range(string), shares(integer), status(string), symbol(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Finage API read of market data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect finage
```

### Inspect as structured JSON

```bash
pm connectors inspect finage --json
```

## Agent Rules

- Run pm connectors inspect finage before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
