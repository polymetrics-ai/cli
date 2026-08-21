---
name: pm-churnkey
description: Churnkey connector knowledge and safe action guide.
---

# pm-churnkey

## Purpose

Reads Churnkey cancel-flow sessions and aggregated session counts through the Churnkey Data API, and sends usage/billing events and customer attribute updates through the Churnkey Event Tracking API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- app_id (required)
- base_url
- api_key (secret) (required)

## ETL Streams

- sessions:
  - primary key: _id
  - cursor: created_at
  - fields: _id(string), aborted(boolean), abtest(string), accepted_offer(object), blueprint_id(string), canceled(boolean), created_at(string), customer(object), customer_billing_interval(string), customer_email(string), customer_id(string), customer_plan_id(string), discount_cooldown_applied(boolean), feedback(string), mode(string), offer_type(string), org(string), provider(string), segment_id(string), survey_choice_id(string), survey_choice_value(string), survey_id(string), updated_at(string)
- session_aggregation:
  - fields: aborted(boolean), billing_interval(string), canceled(boolean), count(integer), month(string), offer_type(string), plan_id(string), save_type(string), trial(boolean)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_event:
  - endpoint: POST /v1/api/events/new
  - required fields: event, customerId
  - risk: external mutation; records a usage/billing event against a Churnkey customer, influencing cancel-flow offer targeting; approval required
- update_customer:
  - endpoint: POST /v1/api/events/customer-update
  - risk: external mutation; overwrites a Churnkey customer's tracked attributes used to drive cancel-flow segmentation and offer eligibility; approval required
- set_billing_users:
  - endpoint: POST /v1/api/events/customer-update/set-users
  - required fields: customerId, users
  - risk: external mutation; overwrites which users on a Churnkey customer account receive Payment Recovery billing-contact emails; approval required

## Security

- read risk: external Churnkey API read of cancel-flow session and customer data
- write risk: external mutation of Churnkey customer event/attribute data used to drive cancel-flow targeting; approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect churnkey
```

### Inspect as structured JSON

```bash
pm connectors inspect churnkey --json
```

## Agent Rules

- Run pm connectors inspect churnkey before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
