---
name: pm-facebook-marketing
description: Facebook Marketing connector knowledge and safe action guide.
---

# pm-facebook-marketing

## Purpose

Reads Facebook Marketing ad accounts, campaigns, ads, ad sets, ad creatives, custom audiences, and performance insights, and creates/updates campaigns and ad sets, through the Graph API.

## Icon

- id: facebook
- asset: icons/facebook.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.facebook.com/docs/marketing-api/marketing-api-changelog

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- ad_account_id
- base_url
- access_token (secret) (required)

## ETL Streams

- ad_accounts:
  - primary key: id
  - fields: account_id(string), account_status(string), currency(string), id(string), name(string), timezone_name(string)
- campaigns:
  - primary key: id
  - fields: created_time(string), effective_status(string), id(string), name(string), objective(string), status(string), updated_time(string)
- ads:
  - primary key: id
  - fields: created_time(string), effective_status(string), id(string), name(string), status(string), updated_time(string)
- ad_sets:
  - primary key: id
  - fields: bid_amount(integer), billing_event(string), campaign_id(string), created_time(string), daily_budget(string), effective_status(string), end_time(string), id(string), lifetime_budget(string), name(string), optimization_goal(string), start_time(string), status(string), updated_time(string)
- ad_creatives:
  - primary key: id
  - fields: id(string), name(string), object_story_id(string), object_type(string), status(string), thumbnail_url(string)
- custom_audiences:
  - primary key: id
  - fields: approximate_count_lower_bound(integer), approximate_count_upper_bound(integer), description(string), id(string), name(string), operation_status(object), subtype(string), time_created(string), time_updated(string)
- insights:
  - primary key: id
  - fields: ad_id(string), ad_name(string), adset_id(string), adset_name(string), campaign_id(string), campaign_name(string), clicks(string), cpc(string), cpm(string), ctr(string), date_start(string), date_stop(string), frequency(string), id(string), impressions(string), reach(string), spend(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_campaign:
  - endpoint: POST /{{ config.ad_account_id }}/campaigns
  - required fields: name, objective, status, special_ad_categories
  - risk: external mutation on a live Facebook ad account; creates a campaign that can incur ad spend once ads are attached; approval required
- update_campaign:
  - endpoint: POST /{{ record.id }}
  - required fields: id
  - risk: external mutation on a live Facebook ad account (e.g. pausing/resuming spend); approval required
- create_ad_set:
  - endpoint: POST /{{ config.ad_account_id }}/adsets
  - required fields: name, campaign_id, billing_event, optimization_goal, targeting, status
  - risk: external mutation on a live Facebook ad account; creates an ad set that can incur ad spend once ads are attached; approval required

## Security

- read risk: external Facebook Graph API read of ad account, campaign, ad, ad set, ad creative, custom audience, and insights (performance metrics) data
- write risk: external mutation of a live Facebook ad account; creating/updating campaigns and ad sets can incur real ad spend once ads are attached and the campaign/ad set is active
- approval: writes require approval; reads are unrestricted
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Facebook Marketing's declared streams and reverse-ETL actions.
- Usage: pm facebook-marketing <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Reverse ETL writes
- Other Commands
  - create ad set apply - Plan and execute the create ad set reverse-ETL action. [intent=reverse_etl availability=implemented write=create_ad_set]; approval: requires plan, preview, approval, and execute; risk: external mutation on a live Facebook ad account; creates an ad set that can incur ad spend once ads are attached; approval required; flags: --billing_event (required), --campaign_id (required), --name (required), --optimization_goal (required), --status (required), --targeting (required)
  - create campaign apply - Plan and execute the create campaign reverse-ETL action. [intent=reverse_etl availability=implemented write=create_campaign]; approval: requires plan, preview, approval, and execute; risk: external mutation on a live Facebook ad account; creates a campaign that can incur ad spend once ads are attached; approval required; flags: --name (required), --objective (required), --special_ad_categories (required), --status (required)
  - update campaign apply - Plan and execute the update campaign reverse-ETL action. [intent=reverse_etl availability=implemented write=update_campaign]; approval: requires plan, preview, approval, and execute; risk: external mutation on a live Facebook ad account (e.g. pausing/resuming spend); approval required; flags: --id (required)

## Commands

### Inspect as a manual

```bash
pm connectors inspect facebook-marketing
```

### Inspect as structured JSON

```bash
pm connectors inspect facebook-marketing --json
```

## Agent Rules

- Run pm connectors inspect facebook-marketing before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
