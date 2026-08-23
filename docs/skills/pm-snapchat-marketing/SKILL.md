---
name: pm-snapchat-marketing
description: Snapchat Marketing connector knowledge and safe action guide.
---

# pm-snapchat-marketing

## Purpose

Reads Snapchat Marketing (Ads API) organizations, ad accounts, campaigns, ad squads, and ads via the OAuth2 refresh-token grant.

## Icon

- id: snapchat
- asset: icons/snapchat.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.snap.com/api/marketing-api/Ads-API/announcements

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- ad_account_ids
- base_url
- organization_ids
- token_url
- client_id (secret) (required)
- client_secret (secret) (required)
- refresh_token (secret) (required)

## ETL Streams

- organizations:
  - primary key: id
  - cursor: updated_at
  - fields: address_line_1(string), administrative_district_level_1(string), country(string), created_at(string), id(string), locality(string), name(string), postal_code(string), type(string), updated_at(string)
- adaccounts:
  - primary key: id
  - cursor: updated_at
  - fields: advertiser(string), created_at(string), currency(string), id(string), name(string), organization_id(string), status(string), timezone(string), type(string), updated_at(string)
- campaigns:
  - primary key: id
  - cursor: updated_at
  - fields: ad_account_id(string), created_at(string), daily_budget_micro(integer), end_time(string), id(string), lifetime_spend_cap_micro(integer), name(string), objective(string), start_time(string), status(string), updated_at(string)
- adsquads:
  - primary key: id
  - cursor: updated_at
  - fields: bid_micro(integer), billing_event(string), campaign_id(string), created_at(string), daily_budget_micro(integer), id(string), name(string), optimization_goal(string), status(string), type(string), updated_at(string)
- ads:
  - primary key: id
  - cursor: updated_at
  - fields: ad_squad_id(string), created_at(string), creative_id(string), id(string), name(string), review_status(string), status(string), type(string), updated_at(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Snapchat Ads API read of organizations, ad accounts, campaigns, ad squads, and ads under the configured organization/ad-account ids
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect snapchat-marketing
```

### Inspect as structured JSON

```bash
pm connectors inspect snapchat-marketing --json
```

## Agent Rules

- Run pm connectors inspect snapchat-marketing before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
