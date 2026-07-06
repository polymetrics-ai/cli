---
name: pm-finage
description: Finage connector knowledge and safe action guide.
---

# pm-finage

## Purpose

Reads Finage US market data: most active stocks, top gainers and losers, sector performance, delisted companies, and per-symbol market news via the Finage REST API.

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
- calendar_from
- calendar_to
- mode
- symbols
- api_key (secret)

## ETL Streams

- most_active_us_stocks:
  - primary key: symbol
  - fields: change(), change_percentage(), company_name(), price(), symbol()
- most_gainers:
  - primary key: symbol
  - fields: change(), change_percentage(), company_name(), price(), symbol()
- most_losers:
  - primary key: symbol
  - fields: change(), change_percentage(), company_name(), price(), symbol()
- sector_performance:
  - primary key: sector
  - fields: change_percentage(), sector()
- delisted_companies:
  - primary key: symbol
  - fields: company_name(), delisted_date(), exchange(), ipo_date(), symbol()
- market_news:
  - primary key: url
  - fields: date(), description(), source(), symbol(), title(), url()
- earnings_calendar:
  - primary key: symbol, date
  - fields: date(), eps(), estimated_eps(), estimated_revenue(), revenue(), symbol(), time()
- ipo_calendar:
  - primary key: symbol, date
  - fields: company(), date(), exchange(), market_cap(), price_range(), shares(), status(), symbol()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
