---
name: pm-ebay-fulfillment
description: eBay Fulfillment connector knowledge and safe action guide.
---

# pm-ebay-fulfillment

## Purpose

Reads eBay seller orders, exploded line items, shipping fulfillments, and payment disputes through the eBay Sell Fulfillment REST API.

## Icon

- id: simple-icons-ebay-fulfillment
- asset: icons/simple-icons/ebay-fulfillment.svg
- title: eBay
- simple_icon_slug: ebay
- simple_icon_hex: E53238
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=eBay
- match: curated-alias
- matched_by: ebay

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- refresh_token_endpoint
- scope
- start_date
- password (secret)
- refresh_token (secret) (required)
- username (secret)

## ETL Streams

- orders:
  - primary key: order_id
  - cursor: creation_date
  - fields: buyer_username(string), creation_date(string), last_modified_date(string), legacy_order_id(string), line_item_count(integer), order_fulfillment_status(string), order_id(string), order_payment_status(string), sales_record_reference(string), seller_id(string), total_currency(string), total_value(string)
- order_line_items:
  - primary key: line_item_id
  - cursor: creation_date
  - fields: creation_date(string), legacy_item_id(string), line_item_fulfillment_status(string), line_item_id(string), order_id(string), quantity(integer), sku(string), title(string), total_currency(string), total_value(string)
- shipping_fulfillments:
  - primary key: order_id
  - cursor: creation_date
  - fields: creation_date(string), legacy_order_id(string), order_fulfillment_status(string), order_id(string), ship_to_city(string), ship_to_country_code(string), ship_to_name(string), ship_to_postal_code(string), ship_to_state_or_province(string), shipping_step(string)
- payment_disputes:
  - primary key: payment_dispute_id
  - cursor: open_date
  - fields: amount_currency(string), amount_value(string), buyer_username(string), dispute_state(string), dispute_status(string), open_date(string), order_id(string), payment_dispute_id(string), reason(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external eBay Sell Fulfillment API read of a seller's order, shipment, and dispute data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Declared ebay-fulfillment API commands.
- Usage: pm ebay-fulfillment <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Other Commands
  - operations get-sell-fulfillment-v1-order - Declared etl: GET /sell/fulfillment/v1/order. [intent=etl availability=partial]; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.get-sell-fulfillment-v1-order-1 has no declaration-owned executable etl route.
  - operations get-sell-fulfillment-v1-order-exploded-line-items - Declared etl: GET /sell/fulfillment/v1/order (exploded lineItems). [intent=etl availability=partial]; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.get-sell-fulfillment-v1-order-exploded-lineitems-2 has no declaration-owned executable etl route.
  - operations get-sell-fulfillment-v1-order-shipping-projection - Declared etl: GET /sell/fulfillment/v1/order (shipping projection). [intent=etl availability=partial]; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.get-sell-fulfillment-v1-order-shipping-projection-3 has no declaration-owned executable etl route.
  - operations get-sell-fulfillment-v1-payment-dispute - Declared etl: GET /sell/fulfillment/v1/payment_dispute. [intent=etl availability=partial]; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.get-sell-fulfillment-v1-payment-dispute-4 has no declaration-owned executable etl route.
  - operations get-sell-fulfillment-v1-order-order-id - Declared direct read: GET /sell/fulfillment/v1/order/{orderId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.get-sell-fulfillment-v1-order-orderid-5 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-sell-fulfillment-v1-order-order-id-shipping-fulfillment - Declared direct read: GET /sell/fulfillment/v1/order/{orderId}/shipping_fulfillment. [intent=direct_read availability=partial]; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.get-sell-fulfillment-v1-order-orderid-shipping-fulfillment-6 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-sell-fulfillment-v1-order-order-id-shipping-fulfillment-fulfillment-id - Declared direct read: GET /sell/fulfillment/v1/order/{orderId}/shipping_fulfillment/{fulfillmentId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.get-sell-fulfillment-v1-order-orderid-shipping-fulfillment-fulfillmentid-7 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations post-sell-fulfillment-v1-order-order-id-shipping-fulfillment - Declared direct write: POST /sell/fulfillment/v1/order/{orderId}/shipping_fulfillment. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /sell/fulfillment/v1/order/{orderId}/shipping_fulfillment.; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.post-sell-fulfillment-v1-order-orderid-shipping-fulfillment-8 has no declaration-owned executable direct_write route.
  - operations get-sell-fulfillment-v1-payment-dispute-payment-dispute-id - Declared direct read: GET /sell/fulfillment/v1/payment_dispute/{payment_dispute_id}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.get-sell-fulfillment-v1-payment-dispute-payment-dispute-id-9 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations post-sell-fulfillment-v1-payment-dispute-payment-dispute-id-accept - Declared direct write: POST /sell/fulfillment/v1/payment_dispute/{payment_dispute_id}/accept. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /sell/fulfillment/v1/payment_dispute/{payment_dispute_id}/accept.; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.post-sell-fulfillment-v1-payment-dispute-payment-dispute-id-accept-10 has no declaration-owned executable direct_write route.
  - operations post-sell-fulfillment-v1-payment-dispute-payment-dispute-id-contest - Declared direct write: POST /sell/fulfillment/v1/payment_dispute/{payment_dispute_id}/contest. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /sell/fulfillment/v1/payment_dispute/{payment_dispute_id}/contest.; notes: Blocked: locked source operation ebay-fulfillment.local-api-surface.post-sell-fulfillment-v1-payment-dispute-payment-dispute-id-contest-11 has no declaration-owned executable direct_write route.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ebay-fulfillment
```

### Inspect as structured JSON

```bash
pm connectors inspect ebay-fulfillment --json
```

## Agent Rules

- Run pm connectors inspect ebay-fulfillment before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
