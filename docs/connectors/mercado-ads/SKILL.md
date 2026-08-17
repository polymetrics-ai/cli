---
name: pm-mercado-ads
description: Mercado Ads connector knowledge and safe action guide.
---

# pm-mercado-ads

## Purpose

Reads Mercado Ads brand, display, and product advertisers and daily campaign metrics from the Mercado Libre Advertising API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
- end_date
- lookback_days
- mode
- start_date
- client_id (secret)
- client_refresh_token (secret)
- client_secret (secret)

## ETL Streams

- brand_advertisers:
  - primary key: advertiser_id
  - fields: account_name(), advertiser_id(), advertiser_name(), site_id()
- display_advertisers:
  - primary key: advertiser_id
  - fields: account_name(), advertiser_id(), advertiser_name(), site_id()
- product_advertisers:
  - primary key: advertiser_id
  - fields: account_name(), advertiser_id(), advertiser_name(), site_id()
- brand_campaigns_metrics:
  - primary key: date, advertiser_id, campaign_id
  - cursor: date
  - fields: acos(), advertiser_id(), campaign_id(), clicks(), cost(), cpc(), ctr(), date(), direct_amount(), indirect_amount(), prints(), total_amount(), units_quantity()
- display_campaigns_metrics:
  - primary key: date, advertiser_id, campaign_id
  - cursor: date
  - fields: acos(), advertiser_id(), campaign_id(), clicks(), cost(), cpc(), ctr(), date(), direct_amount(), indirect_amount(), prints(), total_amount(), units_quantity()
- product_campaigns_metrics:
  - primary key: date, advertiser_id, campaign_id
  - cursor: date
  - fields: acos(), advertiser_id(), campaign_id(), clicks(), cost(), cpc(), ctr(), date(), direct_amount(), indirect_amount(), prints(), total_amount(), units_quantity()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Mercado Ads API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect mercado-ads
```

### Inspect as structured JSON

```bash
pm connectors inspect mercado-ads --json
```

## Agent Rules

- Run pm connectors inspect mercado-ads before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
