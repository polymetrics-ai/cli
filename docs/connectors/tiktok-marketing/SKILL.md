---
name: pm-tiktok-marketing
description: TikTok Marketing connector knowledge and safe action guide.
---

# pm-tiktok-marketing

## Purpose

Reads TikTok Business advertisers, campaigns, ad groups, and ads through the TikTok Marketing (Business) API.

## Icon

- asset: icons/tiktok.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://business-api.tiktok.com/portal/docs?id=1740029169927169

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- advertiser_id
- base_url
- access_token (secret)

## ETL Streams

- advertisers:
  - primary key: advertiser_id
  - fields: advertiser_id(), advertiser_name(), company(), country(), currency(), role(), status(), timezone()
- campaigns:
  - primary key: campaign_id
  - cursor: modify_time
  - fields: advertiser_id(), budget(), budget_mode(), campaign_id(), campaign_name(), create_time(), modify_time(), objective_type(), operation_status()
- adgroups:
  - primary key: adgroup_id
  - cursor: modify_time
  - fields: adgroup_id(), adgroup_name(), advertiser_id(), budget(), budget_mode(), campaign_id(), create_time(), modify_time(), operation_status(), placement_type()
- ads:
  - primary key: ad_id
  - cursor: modify_time
  - fields: ad_id(), ad_name(), adgroup_id(), advertiser_id(), call_to_action(), campaign_id(), create_time(), modify_time(), operation_status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external TikTok Business API read of advertiser and campaign management data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect tiktok-marketing
```

### Inspect as structured JSON

```bash
pm connectors inspect tiktok-marketing --json
```

## Agent Rules

- Run pm connectors inspect tiktok-marketing before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
