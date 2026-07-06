---
name: pm-spotify-ads
description: Spotify Ads connector knowledge and safe action guide.
---

# pm-spotify-ads

## Purpose

Reads Spotify Ads ad accounts, campaigns, ad sets, ads, businesses, business-scoped ad accounts, and assets, and writes campaign mutations through the Spotify Ads API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- ad_account_id
- base_url
- mode
- access_token (secret)

## ETL Streams

- ad_accounts:
  - primary key: id
  - fields: country(), id(), name()
- campaigns:
  - primary key: id
  - fields: id(), name(), objective(), status()
- ad_sets:
  - primary key: id
  - fields: id(), name(), objective(), status()
- ads:
  - primary key: id
  - fields: id(), name(), objective(), status()
- businesses:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), updated_at()
- business_ad_accounts:
  - primary key: id
  - fields: business_id(), country_code(), created_at(), currency_code(), id(), name(), status(), updated_at()
- assets:
  - primary key: id
  - cursor: updated_at
  - fields: asset_type(), created_at(), id(), name(), status(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- update_campaign:
  - endpoint: PATCH /ad_accounts/{{ record.ad_account_id }}/campaigns/{{ record.id }}
  - required fields: ad_account_id, id
  - risk: mutates a live campaign's name, purchase-order reference, or status; setting status to PAUSED/ARCHIVED stops that campaign's ad delivery and spend, approval required

## Security

- read risk: external Spotify Ads API read of ad account, campaign, ad set, ad, business, and asset data
- write risk: external Spotify Ads API mutation of a campaign's name, purchase order reference, or status (active/paused/archived)
- approval: reverse ETL plan approval required before writes; update_campaign can pause or archive live ad spend
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect spotify-ads
```

### Inspect as structured JSON

```bash
pm connectors inspect spotify-ads --json
```

## Agent Rules

- Run pm connectors inspect spotify-ads before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
