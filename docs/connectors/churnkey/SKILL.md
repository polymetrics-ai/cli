---
name: pm-churnkey
description: Churnkey connector knowledge and safe action guide.
---

# pm-churnkey

## Purpose

Reads Churnkey cancel-flow sessions and aggregated session counts through the Churnkey Data API, and sends usage/billing events and customer attribute updates through the Churnkey Event Tracking API.

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

- app_id
- base_url
- api_key (secret)

## ETL Streams

- sessions:
  - primary key: _id
  - cursor: created_at
  - fields: _id(), aborted(), abtest(), accepted_offer(), blueprint_id(), canceled(), created_at(), customer(), customer_billing_interval(), customer_email(), customer_id(), customer_plan_id(), discount_cooldown_applied(), feedback(), mode(), offer_type(), org(), provider(), segment_id(), survey_choice_id(), survey_choice_value(), survey_id(), updated_at()
- session_aggregation:
  - fields: aborted(), billing_interval(), canceled(), count(), month(), offer_type(), plan_id(), save_type(), trial()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_event:
  - endpoint: POST /v1/api/events/new
  - risk: external mutation; records a usage/billing event against a Churnkey customer, influencing cancel-flow offer targeting; approval required
- update_customer:
  - endpoint: POST /v1/api/events/customer-update
  - risk: external mutation; overwrites a Churnkey customer's tracked attributes used to drive cancel-flow segmentation and offer eligibility; approval required
- set_billing_users:
  - endpoint: POST /v1/api/events/customer-update/set-users
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
