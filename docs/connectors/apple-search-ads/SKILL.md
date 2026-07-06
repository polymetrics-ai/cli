---
name: pm-apple-search-ads
description: Apple Ads connector knowledge and safe action guide.
---

# pm-apple-search-ads

## Purpose

Reads Apple Search Ads campaigns, ad groups, targeting keywords, and ads via the Apple Search Ads Campaign Management API using an OAuth2 client-credentials grant scoped to an organization. Read-only.

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
- max_pages
- org_id
- page_size
- token_refresh_endpoint
- client_id (secret)
- client_secret (secret)

## ETL Streams

- campaigns:
  - primary key: id
  - cursor: modification_time
  - fields: ad_channel_type(), billing_event(), budget_amount(), countries_or_regions(), creation_time(), daily_budget_amount(), deleted(), display_status(), id(), modification_time(), name(), org_id(), serving_status(), status(), supply_sources()
- adgroups:
  - primary key: id
  - cursor: modification_time
  - fields: campaign_id(), cpa_goal(), creation_time(), default_bid_amount(), deleted(), display_status(), end_time(), id(), modification_time(), name(), pricing_model(), serving_status(), start_time(), status()
- keywords:
  - primary key: id
  - cursor: modification_time
  - fields: ad_group_id(), bid_amount(), campaign_id(), deleted(), id(), match_type(), modification_time(), status(), text()
- ads:
  - primary key: id
  - cursor: modification_time
  - fields: ad_group_id(), campaign_id(), creation_time(), creative_id(), creative_type(), deleted(), id(), modification_time(), name(), serving_status(), status()

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
