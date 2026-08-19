---
name: pm-criteo-marketing
description: Criteo Marketing connector knowledge and safe action guide.
---

# pm-criteo-marketing

## Purpose

Reads Criteo Marketing Solutions ad sets, advertisers, campaigns, audiences, ad spend statistics, and Marketplace Performance Outcomes (MPO) advertisers/sellers/budgets/seller-campaigns through the Criteo REST API using OAuth2 client-credentials auth.

## Icon

- id: criteo
- asset: icons/criteo.svg
- source: official
- review_status: official_verified
- review_url: https://developers.criteo.com/marketing-solutions/reference/getting-started

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- currency
- end_date
- start_date
- token_url
- client_id (secret) (required)
- client_secret (secret) (required)

## ETL Streams

- ad_sets:
  - primary key: id
  - fields: advertiserId(string), campaignId(string), datasetId(string), destinationEnvironment(string), id(string), mediaType(string), name(string), objective(string), status(string), type(string)
- advertisers:
  - primary key: id
  - fields: country(string), currency(string), id(string), name(string), timezone(string), type(string)
- campaigns:
  - primary key: id
  - fields: advertiserId(string), goal(string), id(string), name(string), objective(string), spendLimit(object), type(string)
- audiences:
  - primary key: id
  - fields: advertiserId(string), description(string), id(string), name(string), nbActiveUsers(integer), type(string)
- statistics:
  - primary key: AdvertiserId, CampaignId, Day
  - cursor: Day
  - fields: AdvertiserId(string), CampaignId(string), Clicks(integer), Currency(string), Day(string), Displays(integer), Spend(number)
- mpo_advertisers:
  - primary key: id
  - fields: advertiserName(string), currencyName(string), id(integer), timeZoneId(string)
- mpo_sellers:
  - primary key: id
  - fields: id(string), sellerName(string)
- mpo_budgets:
  - primary key: id
  - fields: amount(number), budgetType(string), campaignIds(array), endDate(string), id(string), isSuspended(boolean), sellerId(string), spend(number), startDate(string), status(string)
- mpo_seller_campaigns:
  - primary key: id
  - fields: bid(number), campaignId(integer), id(string), sellerId(string), suspendedSince(string), suspensionReasons(array)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Criteo Marketing Solutions API read of advertiser, campaign, and ad spend data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect criteo-marketing
```

### Inspect as structured JSON

```bash
pm connectors inspect criteo-marketing --json
```

## Agent Rules

- Run pm connectors inspect criteo-marketing before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
