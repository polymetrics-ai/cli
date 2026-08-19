---
name: pm-squarespace
description: Squarespace connector knowledge and safe action guide.
---

# pm-squarespace

## Purpose

Reads Squarespace orders, products, inventory, profiles, transactions, store pages, webhook subscriptions, and contacts, and writes webhook subscription mutations through the Squarespace Commerce API.

## Icon

- id: simple-icons-squarespace
- asset: icons/simple-icons/squarespace.svg
- title: Squarespace
- simple_icon_slug: squarespace
- simple_icon_hex: 000000
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Squarespace
- match: exact-name-or-slug
- matched_by: squarespace

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret) (required)

## ETL Streams

- orders:
  - primary key: id
  - cursor: modifiedOn
  - fields: createdOn(string), id(string), modifiedOn(string), orderNumber(string)
- products:
  - primary key: id
  - cursor: modifiedOn
  - fields: createdOn(string), id(string), modifiedOn(string), name(string)
- inventory:
  - primary key: sku
  - fields: modifiedOn(string), quantity(integer), sku(string)
- profiles:
  - primary key: id
  - fields: createdOn(string), id(string), modifiedOn(string), name(string)
- transactions:
  - primary key: id
  - fields: createdOn(string), customerEmail(string), discounts(array), id(string), modifiedOn(string), payments(array), salesLineItems(array), salesOrderId(string), shippingLineItems(array), total(object), totalNetPayment(object), totalNetSales(object), totalNetShipping(object), totalSales(object), totalTaxes(object), voided(boolean)
- store_pages:
  - primary key: id
  - fields: id(string), isEnabled(boolean), title(string), urlSlug(string)
- webhook_subscriptions:
  - primary key: id
  - fields: clientId(string), createdOn(string), endpointUrl(string), id(string), topics(array), updatedOn(string), websiteId(string)
- contacts:
  - primary key: id
  - fields: createdOn(string), defaultShippingAddress(object), firstName(string), id(string), lastName(string), locale(string), primaryEmail(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_webhook_subscription:
  - endpoint: POST /webhook_subscriptions
  - required fields: endpointUrl
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
