---
name: pm-tiktok-marketing
description: TikTok Marketing connector knowledge and safe action guide.
---

# pm-tiktok-marketing

## Purpose

Reads TikTok Business advertisers, campaigns, ad groups, and ads through the TikTok Marketing (Business) API.

## Icon

- id: tiktok
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
- access_token (secret) (required)

## ETL Streams

- advertisers:
  - primary key: advertiser_id
  - fields: advertiser_id(string), advertiser_name(string), company(string), country(string), currency(string), role(string), status(string), timezone(string)
- campaigns:
  - primary key: campaign_id
  - cursor: modify_time
  - fields: advertiser_id(string), budget(number), budget_mode(string), campaign_id(string), campaign_name(string), create_time(string), modify_time(string), objective_type(string), operation_status(string)
- adgroups:
  - primary key: adgroup_id
  - cursor: modify_time
  - fields: adgroup_id(string), adgroup_name(string), advertiser_id(string), budget(number), budget_mode(string), campaign_id(string), create_time(string), modify_time(string), operation_status(string), placement_type(string)
- ads:
  - primary key: ad_id
  - cursor: modify_time
  - fields: ad_id(string), ad_name(string), adgroup_id(string), advertiser_id(string), call_to_action(string), campaign_id(string), create_time(string), modify_time(string), operation_status(string)

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
