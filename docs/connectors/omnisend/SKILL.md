---
name: pm-omnisend
description: Omnisend connector knowledge and safe action guide.
---

# pm-omnisend

## Purpose

Reads Omnisend contacts, campaigns, carts, orders, and products through the Omnisend REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- api_key (secret)

## ETL Streams

- contacts:
  - primary key: contactID
  - cursor: createdAt
  - fields: city(), contactID(), country(), countryCode(), createdAt(), email(), firstName(), lastName(), segments(), state(), status(), tags()
- campaigns:
  - primary key: campaignID
  - cursor: createdAt
  - fields: bounced(), campaignID(), clicked(), createdAt(), endDate(), fromName(), name(), opened(), sent(), startDate(), status(), subject(), type(), unsubscribed()
- carts:
  - primary key: cartID
  - cursor: createdAt
  - fields: cartID(), cartRecoveryUrl(), cartSum(), contactID(), createdAt(), currency(), email(), phone(), products(), updatedAt()
- orders:
  - primary key: orderID
  - cursor: createdAt
  - fields: cartID(), contactID(), createdAt(), currency(), discountSum(), email(), fulfillmentStatus(), orderID(), orderNumber(), orderSum(), paymentStatus(), products(), shippingSum(), subTotalSum(), taxSum(), updatedAt()
- products:
  - primary key: productID
  - cursor: createdAt
  - fields: categoryIDs(), createdAt(), currency(), description(), images(), productID(), productUrl(), status(), title(), type(), updatedAt(), variants(), vendor()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Omnisend API read of contact, campaign, and ecommerce order data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect omnisend
```

### Inspect as structured JSON

```bash
pm connectors inspect omnisend --json
```

## Agent Rules

- Run pm connectors inspect omnisend before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
