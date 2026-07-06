---
name: pm-finnhub
description: Finnhub connector knowledge and safe action guide.
---

# pm-finnhub

## Purpose

Reads Finnhub stock symbols, market news, per-symbol company profiles, and per-symbol analyst recommendation trends through the Finnhub REST API.

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
- exchange
- market_news_category
- mode
- symbols
- api_key (secret)

## ETL Streams

- stock_symbols:
  - primary key: symbol
  - fields: currency(), description(), displaySymbol(), figi(), mic(), symbol(), type()
- market_news:
  - primary key: id
  - cursor: datetime
  - fields: category(), datetime(), headline(), id(), image(), related(), source(), summary(), symbol(), url()
- company_profile:
  - primary key: ticker
  - fields: country(), currency(), exchange(), finnhubIndustry(), ipo(), logo(), marketCapitalization(), name(), phone(), shareOutstanding(), ticker(), weburl()
- stock_recommendations:
  - primary key: symbol, period
  - fields: buy(), hold(), period(), sell(), strongBuy(), strongSell(), symbol()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Finnhub API read of market data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect finnhub
```

### Inspect as structured JSON

```bash
pm connectors inspect finnhub --json
```

## Agent Rules

- Run pm connectors inspect finnhub before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
