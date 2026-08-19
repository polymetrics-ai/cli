---
name: pm-finnhub
description: Finnhub connector knowledge and safe action guide.
---

# pm-finnhub

## Purpose

Reads Finnhub stock symbols, market news, per-symbol company profiles, and per-symbol analyst recommendation trends through the Finnhub REST API.

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
- exchange
- market_news_category
- mode
- symbols
- api_key (secret) (required)

## ETL Streams

- stock_symbols:
  - primary key: symbol
  - fields: currency(string), description(string), displaySymbol(string), figi(string), mic(string), symbol(string), type(string)
- market_news:
  - primary key: id
  - cursor: datetime
  - fields: category(string), datetime(integer), headline(string), id(integer), image(string), related(string), source(string), summary(string), symbol(string), url(string)
- company_profile:
  - primary key: ticker
  - fields: country(string), currency(string), exchange(string), finnhubIndustry(string), ipo(string), logo(string), marketCapitalization(number), name(string), phone(string), shareOutstanding(number), ticker(string), weburl(string)
- stock_recommendations:
  - primary key: symbol, period
  - fields: buy(integer), hold(integer), period(string), sell(integer), strongBuy(integer), strongSell(integer), symbol(string)

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
