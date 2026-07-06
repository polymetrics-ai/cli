---
name: pm-whisky-hunter
description: Whisky Hunter connector knowledge and safe action guide.
---

# pm-whisky-hunter

## Purpose

Reads public Whisky Hunter auction and distillery data. Read-only, no credentials required.

## Icon

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
  - fields: dt(), id(), winning_bid()
- distilleries:
  - primary key: id
  - fields: country(), id(), name()
- auctions_data:
  - primary key: auction_slug, dt
  - fields: all_auctions_lots_count(), auction_lots_count(), auction_name(), auction_slug(), auction_trading_volume(), dt(), winning_bid_mean()
- auctions_info:
  - primary key: slug
  - fields: base_currency(), buyers_fee(), listing_fee(), name(), reserve_fee(), sellers_fee(), slug(), url()
- distilleries_info:
  - primary key: slug
  - fields: country(), name(), slug()
- auction_data:
  - primary key: auction_slug, dt
  - fields: all_auctions_lots_count(), auction_lots_count(), auction_name(), auction_slug(), auction_trading_volume(), dt(), winning_bid_mean()
- distillery_data:
  - primary key: slug, dt
  - fields: dt(), lots_count(), name(), slug(), trading_volume(), winning_bid_max(), winning_bid_mean(), winning_bid_min()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
