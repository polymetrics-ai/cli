---
name: pm-apple-search-ads
description: Apple Ads connector knowledge and safe action guide.
---

# pm-apple-search-ads

## Purpose

Reads Apple Search Ads campaigns, ad groups, targeting keywords, and ads via the Apple Search Ads Campaign Management API using an OAuth2 client-credentials grant scoped to an organization. Read-only.

## Icon

- id: simple-icons-apple
- asset: icons/simple-icons/apple.svg
- title: Apple
- simple_icon_slug: apple
- simple_icon_hex: 000000
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Apple
- match: curated-alias
- matched_by: apple

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- org_id (required)
- page_size
- token_refresh_endpoint
- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- campaigns:
  - primary key: id
  - cursor: modification_time
  - fields: ad_channel_type(string), billing_event(string), budget_amount(object), countries_or_regions(array), creation_time(string), daily_budget_amount(object), deleted(boolean), display_status(string), id(integer), modification_time(string), name(string), org_id(integer), serving_status(string), status(string), supply_sources(array)
- adgroups:
  - primary key: id
  - cursor: modification_time
  - fields: campaign_id(integer), cpa_goal(object), creation_time(string), default_bid_amount(object), deleted(boolean), display_status(string), end_time(string), id(integer), modification_time(string), name(string), pricing_model(string), serving_status(string), start_time(string), status(string)
- keywords:
  - primary key: id
  - cursor: modification_time
  - fields: ad_group_id(integer), bid_amount(object), campaign_id(integer), deleted(boolean), id(integer), match_type(string), modification_time(string), status(string), text(string)
- ads:
  - primary key: id
  - cursor: modification_time
  - fields: ad_group_id(integer), campaign_id(integer), creation_time(string), creative_id(integer), creative_type(string), deleted(boolean), id(integer), modification_time(string), name(string), serving_status(string), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Apple Search Ads API read of campaign, ad group, keyword, and ad data
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect apple-search-ads
```

### Inspect as structured JSON

```bash
pm connectors inspect apple-search-ads --json
```

## Agent Rules

- Run pm connectors inspect apple-search-ads before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
