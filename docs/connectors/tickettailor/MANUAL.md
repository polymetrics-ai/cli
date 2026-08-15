# pm connectors inspect tickettailor

```text
NAME
  pm connectors inspect tickettailor - Ticket Tailor connector manual

SYNOPSIS
  pm connectors inspect tickettailor
  pm connectors inspect tickettailor --json
  pm credentials add <name> --connector tickettailor [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes events, orders, issued tickets, event series, holds, discounts, memberships, products, stores, and vouchers through the Ticket Tailor API.

ICON
  id: simple-icons-tickettailor
  asset: icons/simple-icons/tickettailor.svg
  title: Ticket Tailor
  simple_icon_slug: tickettailor
  simple_icon_hex: 222432
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Ticket%20Tailor
  match: exact-name-or-slug
  matched_by: tickettailor

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  events:
    primary key: id
    fields: end_date(string), id(string), name(string), start_date(string), status(string)
  orders:
    primary key: id
    fields: created_at(string), email(string), event_id(string), id(string), total(string)
  issued_tickets:
    primary key: id
    fields: event_id(string), id(string), order_id(string), status(string), ticket_type_id(string)
  event_series:
    primary key: id
    fields: created_at(integer), currency(string), description(string), id(string), name(string)
  holds:
    primary key: id
    fields: created_at(integer), event_id(string), id(string), note(string), total_on_hold(integer), updated_at(integer)
  discounts:
    primary key: id
    fields: code(string), id(string), max_redemptions(integer), name(string), times_redeemed(integer), type(string)
  membership_types:
    primary key: id
    fields: id(string), max_redemptions(integer), name(string), valid_from_type(string), valid_to_type(string)
  issued_memberships:
    primary key: id
    fields: code(string), email(string), first_name(string), full_name(string), id(string), is_valid(boolean), last_name(string), membership_type_id(string), membership_type_name(string)
  products:
    primary key: id
    fields: created_at(integer), currency(string), description(string), id(string), name(string), price(integer)
  stores:
    primary key: id
    fields: currency(string), id(string), name(string)
  vouchers:
    primary key: id
    fields: available_codes(integer), expiry(integer), id(string), name(string), total_codes(integer), type(string), value(integer)
  checkout_forms:
    primary key: id
    fields: created_at(integer), event_series_id(string), id(string)
  voucher_codes:
    primary key: id
    fields: code(string), expiry(integer), id(string), used(boolean), value(integer), voucher_id(string)
  checkout_form_elements:
    primary key: id, checkout_form_id
    fields: checkout_form_id(string), id(string), per_ticket(boolean), question(string), required(boolean), type(string)
  event_series_overrides:
    primary key: id, event_series_id
    fields: created_at(integer), event_series_id(string), id(string), max_sellable_tickets(integer), name(string)
  event_series_waitlist_signups:
    primary key: id, event_series_id
    fields: created_at(integer), email(string), event_id(string), event_series_id(string), id(string), notified_date(integer)
  overview:
    primary key: id
    fields: box_office_name(string), credits(number), currency(string), id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_event_series:
    endpoint: POST /event_series
    required fields: name
    risk: creates a new event series (a recurring/template event definition); low-risk additive external mutation, no approval required
  update_event_series:
    endpoint: POST /event_series/{{ record.id }}
    required fields: id
    risk: mutates an existing event series' public-facing name/description/currency
  delete_event_series:
    endpoint: DELETE /event_series/{{ record.id }}
    required fields: id
    risk: permanently deletes an event series and every event occurrence within it; destructive, approval required
  change_event_series_status:
    endpoint: POST /event_series/{{ record.id }}/status
    required fields: id, status
    risk: changes an event series' publication status; setting to draft/sales_closed immediately stops further public ticket sales
  create_discount:
    endpoint: POST /discounts
    required fields: name, code, type
    risk: creates a discount code redeemable at checkout; low-risk additive external mutation, no approval required
  update_discount:
    endpoint: POST /discounts/{{ record.id }}
    required fields: id
    risk: mutates an existing discount code's name, code, or usage limit; changing the code invalidates any already-shared link using the old code
  delete_discount:
    endpoint: DELETE /discounts/{{ record.id }}
    required fields: id
    risk: permanently deletes a discount code; any customer relying on the code at checkout will see it rejected
  delete_hold:
    endpoint: DELETE /holds/{{ record.id }}
    required fields: id
    risk: releases a hold, returning its reserved tickets to public sale immediately
  create_check_in:
    endpoint: POST /check_ins
    required fields: issued_ticket_id, quantity
    risk: checks an attendee's issued ticket in (or out, when quantity is -1) at the door; low-risk operational mutation, no approval required
  create_issued_ticket:
    endpoint: POST /issued_tickets
    required fields: full_name
    risk: issues a new ticket directly (bypassing checkout), consuming inventory from either a ticket type or an existing hold; low-risk additive external mutation, no approval required
  void_issued_ticket:
    endpoint: POST /issued_tickets/{{ record.id }}/void
    required fields: id
    risk: voids an issued ticket, invalidating it for entry; optionally returns its inventory to a hold rather than public sale
  update_order:
    endpoint: POST /orders/{{ record.id }}
    required fields: id
    risk: mutates an existing order's buyer contact/address details
  confirm_order_payment_received:
    endpoint: POST /orders/{{ record.id }}/confirm-payment-received
    required fields: id
    risk: marks an order (typically an offline/manual payment method) as paid, releasing its tickets from pending status
  create_membership_type:
    endpoint: POST /membership_types
    required fields: name, valid_from_type, valid_to_type
    risk: creates a new membership type template; low-risk additive external mutation, no approval required
  delete_membership_type:
    endpoint: DELETE /membership_types/{{ record.id }}
    required fields: id
    risk: permanently deletes a membership type; any issued membership referencing it is orphaned
  create_issued_membership:
    endpoint: POST /issued_memberships
    required fields: membership_type_id, first_name, last_name, email
    risk: issues a new membership directly to a member; low-risk additive external mutation, no approval required
  update_issued_membership:
    endpoint: POST /issued_memberships/{{ record.id }}
    required fields: id
    risk: mutates an existing issued membership's holder details or validity window
  void_issued_membership:
    endpoint: POST /issued_memberships/{{ record.id }}/void
    required fields: id
    risk: voids an issued membership, invalidating it immediately for entry/redemption
  create_voucher:
    endpoint: POST /vouchers
    required fields: name, value
    risk: creates a new voucher and its redeemable codes; low-risk additive external mutation, no approval required
  update_voucher:
    endpoint: POST /vouchers/{{ record.id }}
    required fields: id
    risk: mutates an existing voucher's value or expiry, directly changing what every un-redeemed code is worth
  delete_voucher:
    endpoint: DELETE /vouchers/{{ record.id }}
    required fields: id
    risk: permanently deletes a voucher and every un-redeemed code issued under it
  void_voucher_code:
    endpoint: POST /vouchers/{{ record.voucher_id }}/codes/{{ record.id }}/void
    required fields: voucher_id, id
    risk: voids a single voucher code, invalidating it for redemption immediately
  create_product:
    endpoint: POST /products
    required fields: name, price
    risk: creates a new sellable add-on product; low-risk additive external mutation, no approval required
  update_product:
    endpoint: POST /products/{{ record.id }}
    required fields: id
    risk: mutates an existing product's name, price, or description, directly changing checkout pricing for it
  delete_product:
    endpoint: DELETE /products/{{ record.id }}
    required fields: id
    risk: permanently deletes a sellable product; it becomes unavailable at checkout immediately

SECURITY
  read risk: external Ticket Tailor API read of event, order, issued ticket, event series, hold, discount, membership, product, store, and voucher data
  write risk: external Ticket Tailor API mutations covering event series/hold/discount/membership/voucher/product lifecycle, ticket issuance/voiding/check-in, and order payment confirmation; delete_event_series is destructive/confirm-gated
  approval: required for delete_event_series (confirm: destructive); other writes are low-risk additive/idempotent mutations, no approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect tickettailor

  # Inspect as structured JSON
  pm connectors inspect tickettailor --json

AGENT WORKFLOW
  - Run pm connectors inspect tickettailor before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
