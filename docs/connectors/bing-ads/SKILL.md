---
name: pm-bing-ads
description: Bing Ads connector knowledge and safe action guide.
---

# pm-bing-ads

## Purpose

Reads Microsoft Advertising (Bing Ads) accounts, users, campaigns, ad groups, and ads through the v13 Customer Management and Campaign Management REST APIs. Read-only.

## Icon

- id: bingads
- asset: icons/bingads.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://learn.microsoft.com/en-us/advertising/guides/release-notes

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- account_ids
- ad_group_id
- base_url
- campaign_base_url
- campaign_id
- customer_account_id
- customer_id
- token_url
- client_id (secret) (required)
- client_secret (secret)
- developer_token (secret) (required)
- refresh_token (secret) (required)
- tenant_id (secret)

## ETL Streams

- accounts:
  - primary key: Id
  - fields: AccountLifeCycleStatus(string), Id(string), Name(string), Number(string), PauseReason(string)
- users:
  - primary key: Id
  - fields: CustomerId(string), Id(string), JobTitle(string), LastModifiedTime(string), UserLifeCycleStatus(string), UserName(string)
- campaigns:
  - primary key: Id
  - fields: BudgetType(string), CampaignType(string), DailyBudget(number), Id(string), Name(string), Status(string), TimeZone(string)
- ad_groups:
  - primary key: Id
  - fields: AdRotation(string), EndDate(string), Id(string), Name(string), Network(string), StartDate(string), Status(string)
- ads:
  - primary key: Id
  - fields: DevicePreference(string), EditorialStatus(string), Id(string), Status(string), Type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Microsoft Advertising REST API read of account/user/campaign/ad-group/ad metadata
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect bing-ads
```

### Inspect as structured JSON

```bash
pm connectors inspect bing-ads --json
```

## Agent Rules

- Run pm connectors inspect bing-ads before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
