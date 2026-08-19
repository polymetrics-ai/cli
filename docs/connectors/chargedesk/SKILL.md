---
name: pm-chargedesk
description: ChargeDesk connector knowledge and safe action guide.
---

# pm-chargedesk

## Purpose

Reads ChargeDesk charges, customers, subscriptions, and products through the ChargeDesk REST API.

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

- base_url
- mode
- username
- password (secret) (required)

## ETL Streams

- charges:
  - primary key: charge_id
  - cursor: occurred
  - fields: amount(string), amount_refunded(string), charge_id(string), currency(string), customer_email(string), customer_id(string), customer_name(string), description(string), object(string), occurred(integer), payment_method(string), product_id(string), status(string), subscription_id(string), transaction_id(string)
- customers:
  - primary key: customer_id
  - cursor: occurred
  - fields: country(string), currency(string), customer_id(string), delinquent(boolean), email(string), name(string), object(string), occurred(integer), phone(string), tax_number(string)
- subscriptions:
  - primary key: subscription_id
  - cursor: occurred
  - fields: amount(string), currency(string), current_period_end(integer), current_period_start(integer), customer_id(string), interval(string), object(string), occurred(integer), product_id(string), status(string), subscription_id(string)
- products:
  - primary key: product_id
  - cursor: occurred
  - fields: amount(string), currency(string), interval(string), name(string), object(string), occurred(integer), product_id(string), status(string)
- log_activity:
  - cursor: occurred
  - fields: action_params(boolean), action_reason(string), action_type(string), company(string), context(string), description(string), event(string), ip(string), object_id(string), object_type(string), occurred(integer), params(string), source(string), sub_description(string)
- log_cancellations:
  - cursor: occurred
  - fields: action(string), customer_id(string), email(string), ip(string), method(string), occurred(integer), reason(string), subscription_id(string)
- webhook_notifications:
  - primary key: notification
  - fields: description(string), name(string), notification(string), object(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /customers
  - risk: external mutation creating a new ChargeDesk customer record; approval required
- update_customer:
  - endpoint: POST /customers/{{ record.customer_id }}
  - required fields: customer_id
  - risk: external mutation updating an existing ChargeDesk customer record; approval required
- delete_customer:
  - endpoint: DELETE /customers/{{ record.customer_id }}
  - required fields: customer_id
  - risk: irreversible deletion of a customer record (and, by ChargeDesk's own default, all associated charges/tickets); approval required
- update_charge:
  - endpoint: POST /charges/{{ record.charge_id }}
  - required fields: charge_id
  - risk: external mutation updating an existing charge record's stored data; approval required
- delete_charge:
  - endpoint: DELETE /charges/{{ record.charge_id }}
  - required fields: charge_id
  - risk: irreversible deletion of a charge record; approval required
- refund_charge:
  - endpoint: POST /gateway/charges/{{ record.charge_id }}/refund
  - required fields: charge_id
  - risk: gateway method; irreversibly refunds a charge (full or partial) on the originating payment gateway as well as ChargeDesk; approval required
- capture_charge:
  - endpoint: POST /gateway/charges/{{ record.charge_id }}/capture
  - required fields: charge_id
  - risk: gateway method; captures (settles) a previously authorized charge on the originating payment gateway; approval required
- void_charge:
  - endpoint: POST /gateway/charges/{{ record.charge_id }}/void
  - required fields: charge_id
  - risk: gateway method; voids an authorized charge or cancels an outstanding payment request on the originating payment gateway; approval required
- cancel_subscription:
  - endpoint: POST /gateway/subscriptions/{{ record.subscription_id }}/cancel
  - required fields: subscription_id
  - risk: gateway method; irreversibly cancels future recurring charges for a subscription on the originating payment gateway as well as ChargeDesk; approval required
- create_webhook:
  - endpoint: POST /webhooks
  - required fields: url
  - risk: external mutation creating a new outbound webhook subscription that will POST ChargeDesk event data to a third-party URL; approval required
- delete_webhook:
  - endpoint: DELETE /webhooks/{{ record.webhook_id }}
  - required fields: webhook_id
  - risk: irreversible removal of an outbound webhook subscription; approval required
- create_agent:
  - endpoint: POST /agents
  - required fields: name, email, role
  - risk: external mutation inviting a new support agent (or updating an existing agent's role) with account access to ChargeDesk; approval required
- delete_agent:
  - endpoint: DELETE /agents/{{ record.email }}
  - required fields: email
  - risk: irreversible removal of a support agent's ChargeDesk account access; approval required

## Security

- read risk: external ChargeDesk API read of billing/charge, customer, subscription, product, activity-log, and cancellation-log data
- write risk: external mutations creating/updating/deleting customers, charges, webhooks, and agents, plus live gateway methods (refund/capture/void a charge, cancel a subscription) that mutate the connected payment gateway; every write action requires approval
- approval: read: none; write: required for every action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect chargedesk
```

### Inspect as structured JSON

```bash
pm connectors inspect chargedesk --json
```

## Agent Rules

- Run pm connectors inspect chargedesk before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
