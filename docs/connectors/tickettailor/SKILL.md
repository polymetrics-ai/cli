---
name: pm-tickettailor
description: Ticket Tailor connector knowledge and safe action guide.
---

# pm-tickettailor

## Purpose

Reads and writes events, orders, issued tickets, event series, holds, discounts, memberships, products, stores, and vouchers through the Ticket Tailor API.

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

- events:
  - primary key: id
  - fields: end_date(), id(), name(), start_date(), status()
- orders:
  - primary key: id
  - fields: created_at(), email(), event_id(), id(), total()
- issued_tickets:
  - primary key: id
  - fields: event_id(), id(), order_id(), status(), ticket_type_id()
- event_series:
  - primary key: id
  - fields: created_at(), currency(), description(), id(), name()
- holds:
  - primary key: id
  - fields: created_at(), event_id(), id(), note(), total_on_hold(), updated_at()
- discounts:
  - primary key: id
  - fields: code(), id(), max_redemptions(), name(), times_redeemed(), type()
- membership_types:
  - primary key: id
  - fields: id(), max_redemptions(), name(), valid_from_type(), valid_to_type()
- issued_memberships:
  - primary key: id
  - fields: code(), email(), first_name(), full_name(), id(), is_valid(), last_name(), membership_type_id(), membership_type_name()
- products:
  - primary key: id
  - fields: created_at(), currency(), description(), id(), name(), price()
- stores:
  - primary key: id
  - fields: currency(), id(), name()
- vouchers:
  - primary key: id
  - fields: available_codes(), expiry(), id(), name(), total_codes(), type(), value()
- checkout_forms:
  - primary key: id
  - fields: created_at(), event_series_id(), id()
- voucher_codes:
  - primary key: id
  - fields: code(), expiry(), id(), used(), value(), voucher_id()
- checkout_form_elements:
  - primary key: id, checkout_form_id
  - fields: checkout_form_id(), id(), per_ticket(), question(), required(), type()
- event_series_overrides:
  - primary key: id, event_series_id
  - fields: created_at(), event_series_id(), id(), max_sellable_tickets(), name()
- event_series_waitlist_signups:
  - primary key: id, event_series_id
  - fields: created_at(), email(), event_id(), event_series_id(), id(), notified_date()
- overview:
  - primary key: id
  - fields: box_office_name(), credits(), currency(), id()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_event_series:
  - endpoint: POST /event_series
  - risk: creates a new event series (a recurring/template event definition); low-risk additive external mutation, no approval required
- update_event_series:
  - endpoint: POST /event_series/{{ record.id }}
  - required fields: id
  - risk: mutates an existing event series' public-facing name/description/currency
- delete_event_series:
  - endpoint: DELETE /event_series/{{ record.id }}
  - required fields: id
  - risk: permanently deletes an event series and every event occurrence within it; destructive, approval required
- change_event_series_status:
  - endpoint: POST /event_series/{{ record.id }}/status
  - required fields: id
  - optional fields: status
  - risk: changes an event series' publication status; setting to draft/sales_closed immediately stops further public ticket sales
- create_discount:
  - endpoint: POST /discounts
  - risk: creates a discount code redeemable at checkout; low-risk additive external mutation, no approval required
- update_discount:
  - endpoint: POST /discounts/{{ record.id }}
  - required fields: id
  - risk: mutates an existing discount code's name, code, or usage limit; changing the code invalidates any already-shared link using the old code
- delete_discount:
  - endpoint: DELETE /discounts/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a discount code; any customer relying on the code at checkout will see it rejected
- delete_hold:
  - endpoint: DELETE /holds/{{ record.id }}
  - required fields: id
  - risk: releases a hold, returning its reserved tickets to public sale immediately
- create_check_in:
  - endpoint: POST /check_ins
  - risk: checks an attendee's issued ticket in (or out, when quantity is -1) at the door; low-risk operational mutation, no approval required
- create_issued_ticket:
  - endpoint: POST /issued_tickets
  - risk: issues a new ticket directly (bypassing checkout), consuming inventory from either a ticket type or an existing hold; low-risk additive external mutation, no approval required
- void_issued_ticket:
  - endpoint: POST /issued_tickets/{{ record.id }}/void
  - required fields: id
  - risk: voids an issued ticket, invalidating it for entry; optionally returns its inventory to a hold rather than public sale
- update_order:
  - endpoint: POST /orders/{{ record.id }}
  - required fields: id
  - risk: mutates an existing order's buyer contact/address details
- confirm_order_payment_received:
  - endpoint: POST /orders/{{ record.id }}/confirm-payment-received
  - required fields: id
  - risk: marks an order (typically an offline/manual payment method) as paid, releasing its tickets from pending status
- create_membership_type:
  - endpoint: POST /membership_types
  - risk: creates a new membership type template; low-risk additive external mutation, no approval required
- delete_membership_type:
  - endpoint: DELETE /membership_types/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a membership type; any issued membership referencing it is orphaned
- create_issued_membership:
  - endpoint: POST /issued_memberships
  - risk: issues a new membership directly to a member; low-risk additive external mutation, no approval required
- update_issued_membership:
  - endpoint: POST /issued_memberships/{{ record.id }}
  - required fields: id
  - risk: mutates an existing issued membership's holder details or validity window
- void_issued_membership:
  - endpoint: POST /issued_memberships/{{ record.id }}/void
  - required fields: id
  - risk: voids an issued membership, invalidating it immediately for entry/redemption
- create_voucher:
  - endpoint: POST /vouchers
  - risk: creates a new voucher and its redeemable codes; low-risk additive external mutation, no approval required
- update_voucher:
  - endpoint: POST /vouchers/{{ record.id }}
  - required fields: id
  - risk: mutates an existing voucher's value or expiry, directly changing what every un-redeemed code is worth
- delete_voucher:
  - endpoint: DELETE /vouchers/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a voucher and every un-redeemed code issued under it
- void_voucher_code:
  - endpoint: POST /vouchers/{{ record.voucher_id }}/codes/{{ record.id }}/void
  - required fields: voucher_id, id
  - risk: voids a single voucher code, invalidating it for redemption immediately
- create_product:
  - endpoint: POST /products
  - risk: creates a new sellable add-on product; low-risk additive external mutation, no approval required
- update_product:
  - endpoint: POST /products/{{ record.id }}
  - required fields: id
  - risk: mutates an existing product's name, price, or description, directly changing checkout pricing for it
- delete_product:
  - endpoint: DELETE /products/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a sellable product; it becomes unavailable at checkout immediately

## Security

- read risk: external Ticket Tailor API read of event, order, issued ticket, event series, hold, discount, membership, product, store, and voucher data
- write risk: external Ticket Tailor API mutations covering event series/hold/discount/membership/voucher/product lifecycle, ticket issuance/voiding/check-in, and order payment confirmation; delete_event_series is destructive/confirm-gated
- approval: required for delete_event_series (confirm: destructive); other writes are low-risk additive/idempotent mutations, no approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect tickettailor
```

### Inspect as structured JSON

```bash
pm connectors inspect tickettailor --json
```

## Agent Rules

- Run pm connectors inspect tickettailor before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
