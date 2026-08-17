---
name: pm-squarespace
description: Squarespace connector knowledge and safe action guide.
---

# pm-squarespace

## Purpose

Reads Squarespace orders, products, inventory, profiles, transactions, store pages, webhook subscriptions, and contacts, and writes webhook subscription mutations through the Squarespace Commerce API.

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

- base_url
- api_key (secret)

## ETL Streams

- orders:
  - primary key: id
  - cursor: modifiedOn
  - fields: createdOn(), id(), modifiedOn(), orderNumber()
- products:
  - primary key: id
  - cursor: modifiedOn
  - fields: createdOn(), id(), modifiedOn(), name()
- inventory:
  - primary key: sku
  - fields: modifiedOn(), quantity(), sku()
- profiles:
  - primary key: id
  - fields: createdOn(), id(), modifiedOn(), name()
- transactions:
  - primary key: id
  - fields: createdOn(), customerEmail(), discounts(), id(), modifiedOn(), payments(), salesLineItems(), salesOrderId(), shippingLineItems(), total(), totalNetPayment(), totalNetSales(), totalNetShipping(), totalSales(), totalTaxes(), voided()
- store_pages:
  - primary key: id
  - fields: id(), isEnabled(), title(), urlSlug()
- webhook_subscriptions:
  - primary key: id
  - fields: clientId(), createdOn(), endpointUrl(), id(), topics(), updatedOn(), websiteId()
- contacts:
  - primary key: id
  - fields: createdOn(), defaultShippingAddress(), firstName(), id(), lastName(), locale(), primaryEmail()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_webhook_subscription:
  - endpoint: POST /webhook_subscriptions
  - risk: registers a new HTTPS endpoint to receive live order/contact/address event notifications; low-risk external mutation, no approval required
- delete_webhook_subscription:
  - endpoint: DELETE /webhook_subscriptions/{{ record.id }}
  - required fields: id
  - risk: permanently removes a webhook subscription, stopping future event notifications to that endpoint; external mutation, approval required

## Security

- read risk: external Squarespace API read of commerce orders, products, inventory, profiles, transactions, store pages, webhook subscriptions, and contacts
- write risk: external Squarespace API mutation (webhook subscription create/delete)
- approval: reverse ETL plan approval required before destructive writes (delete_webhook_subscription); create_webhook_subscription is low-risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect squarespace
```

### Inspect as structured JSON

```bash
pm connectors inspect squarespace --json
```

## Agent Rules

- Run pm connectors inspect squarespace before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
