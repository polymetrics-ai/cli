---
name: pm-mercado-ads
description: Mercado Ads connector knowledge and safe action guide.
---

# pm-mercado-ads

## Purpose

Reads Mercado Ads brand, display, and product advertisers and daily campaign metrics from the Mercado Libre Advertising API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
- end_date
- lookback_days (required)
- mode
- start_date
- client_id (secret) (required)
- client_refresh_token (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- brand_advertisers:
  - primary key: advertiser_id
  - fields: account_name(string), advertiser_id(integer), advertiser_name(string), site_id(string)
- display_advertisers:
  - primary key: advertiser_id
  - fields: account_name(string), advertiser_id(integer), advertiser_name(string), site_id(string)
- product_advertisers:
  - primary key: advertiser_id
  - fields: account_name(string), advertiser_id(integer), advertiser_name(string), site_id(string)
- brand_campaigns_metrics:
  - primary key: date, advertiser_id, campaign_id
  - cursor: date
  - fields: acos(number), advertiser_id(string), campaign_id(string), clicks(number), cost(number), cpc(number), ctr(number), date(string), direct_amount(number), indirect_amount(number), prints(number), total_amount(number), units_quantity(number)
- display_campaigns_metrics:
  - primary key: date, advertiser_id, campaign_id
  - cursor: date
  - fields: acos(number), advertiser_id(string), campaign_id(string), clicks(number), cost(number), cpc(number), ctr(number), date(string), direct_amount(number), indirect_amount(number), prints(number), total_amount(number), units_quantity(number)
- product_campaigns_metrics:
  - primary key: date, advertiser_id, campaign_id
  - cursor: date
  - fields: acos(number), advertiser_id(string), campaign_id(string), clicks(number), cost(number), cpc(number), ctr(number), date(string), direct_amount(number), indirect_amount(number), prints(number), total_amount(number), units_quantity(number)

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
