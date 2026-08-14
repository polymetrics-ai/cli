---
name: pm-whisky-hunter
description: Whisky Hunter connector knowledge and safe action guide.
---

# pm-whisky-hunter

## Purpose

Reads public Whisky Hunter auction and distillery data. Read-only, no credentials required.

## Icon

- id: whiskyhunter
- asset: icons/whiskyhunter.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://whiskyhunter.net/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url

## ETL Streams

- auctions:
  - primary key: id
  - fields: dt(string), id(integer), winning_bid(number)
- distilleries:
  - primary key: id
  - fields: country(string), id(integer), name(string)
- auctions_data:
  - primary key: auction_slug, dt
  - fields: all_auctions_lots_count(integer), auction_lots_count(integer), auction_name(string), auction_slug(string), auction_trading_volume(number), dt(string), winning_bid_mean(number)
- auctions_info:
  - primary key: slug
  - fields: base_currency(string), buyers_fee(number), listing_fee(number), name(string), reserve_fee(number), sellers_fee(number), slug(string), url(string)
- distilleries_info:
  - primary key: slug
  - fields: country(string), name(string), slug(string)
- auction_data:
  - primary key: auction_slug, dt
  - fields: all_auctions_lots_count(integer), auction_lots_count(integer), auction_name(string), auction_slug(string), auction_trading_volume(number), dt(string), winning_bid_mean(number)
- distillery_data:
  - primary key: slug, dt
  - fields: dt(string), lots_count(integer), name(string), slug(string), trading_volume(number), winning_bid_max(number), winning_bid_mean(number), winning_bid_min(number)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Whisky Hunter API read of public auction and distillery data
- approval: none; read-only public API, no credentials
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect whisky-hunter
```

### Inspect as structured JSON

```bash
pm connectors inspect whisky-hunter --json
```

## Agent Rules

- Run pm connectors inspect whisky-hunter before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
