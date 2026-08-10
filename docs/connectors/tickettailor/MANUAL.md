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
  api_key (secret)

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
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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

COMMAND SURFACE
  Run Ticket Tailor's declared streams and reverse-ETL actions.
  Usage: pm tickettailor <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete v1 event-series event-series-id bundles bundle-id - Documented DELETE /v1/event_series/{event_series_id}/bundles/{bundle_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.delete.v1-event-series-event-series-id-bundles-bundle-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 event-series event-series-id events event-occurrence-id - Documented DELETE /v1/event_series/{event_series_id}/events/{event_occurrence_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.delete.v1-event-series-event-series-id-events-event-occurrence-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 event-series event-series-id ticket-groups ticket-group-id - Documented DELETE /v1/event_series/{event_series_id}/ticket_groups/{ticket_group_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.delete.v1-event-series-event-series-id-ticket-groups-ticket-group-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 event-series event-series-id ticket-types ticket-type-id - Documented DELETE /v1/event_series/{event_series_id}/ticket_types/{ticket_type_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.delete.v1-event-series-event-series-id-ticket-types-ticket-type-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 event-series event-series-id waitlist-signup waitlist-signup-id - Documented DELETE /v1/event_series/{event_series_id}/waitlist_signup/{waitlist_signup_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.delete.v1-event-series-event-series-id-waitlist-signup-waitlist-signup-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get v1 check-ins - Documented GET /v1/check_ins (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-check-ins]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 checkout-forms checkout-form-id elements checkout-form-element-id - Documented GET /v1/checkout_forms/{checkout_form_id}/elements/{checkout_form_element_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-checkout-forms-checkout-form-id-elements-checkout-form-element-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 discounts discount-id - Documented GET /v1/discounts/{discount_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-discounts-discount-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 event-series event-series-id - Documented GET /v1/event_series/{event_series_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-event-series-event-series-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 event-series event-series-id events - Documented GET /v1/event_series/{event_series_id}/events (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-event-series-event-series-id-events]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 event-series event-series-id events event-occurrence-id - Documented GET /v1/event_series/{event_series_id}/events/{event_occurrence_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-event-series-event-series-id-events-event-occurrence-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 events event-id - Documented GET /v1/events/{event_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-events-event-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 holds hold-id - Documented GET /v1/holds/{hold_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-holds-hold-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 issued-memberships issued-membership-id - Documented GET /v1/issued_memberships/{issued_membership_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-issued-memberships-issued-membership-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 issued-tickets issued-ticket-id - Documented GET /v1/issued_tickets/{issued_ticket_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-issued-tickets-issued-ticket-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 membership-photo-share - Documented GET /v1/membership_photo_share (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-membership-photo-share]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 membership-types membership-type-id - Documented GET /v1/membership_types/{membership_type_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-membership-types-membership-type-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 orders order-id - Documented GET /v1/orders/{order_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-orders-order-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 ping - Documented GET /v1/ping (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-ping]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 products product-id - Documented GET /v1/products/{product_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-products-product-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 stores store-id - Documented GET /v1/stores/{store_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-stores-store-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 vouchers voucher-id - Documented GET /v1/vouchers/{voucher_id} (not implemented) [intent=direct_read availability=not_implemented operation=tickettailor.get.v1-vouchers-voucher-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post v1 checkout-forms checkout-form-id elements checkout-form-element-id - Documented POST /v1/checkout_forms/{checkout_form_id}/elements/{checkout_form_element_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-checkout-forms-checkout-form-id-elements-checkout-form-element-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id bundles - Documented POST /v1/event_series/{event_series_id}/bundles (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-bundles]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id bundles bundle-id - Documented POST /v1/event_series/{event_series_id}/bundles/{bundle_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-bundles-bundle-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id events - Documented POST /v1/event_series/{event_series_id}/events (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-events]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id events event-occurrence-id - Documented POST /v1/event_series/{event_series_id}/events/{event_occurrence_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-events-event-occurrence-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id overrides - Documented POST /v1/event_series/{event_series_id}/overrides (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-overrides]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id overrides override-id - Documented POST /v1/event_series/{event_series_id}/overrides/{override_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-overrides-override-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id ticket-groups - Documented POST /v1/event_series/{event_series_id}/ticket_groups (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-ticket-groups]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id ticket-groups ticket-group-id - Documented POST /v1/event_series/{event_series_id}/ticket_groups/{ticket_group_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-ticket-groups-ticket-group-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id ticket-types - Documented POST /v1/event_series/{event_series_id}/ticket_types (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-ticket-types]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 event-series event-series-id ticket-types ticket-type-id - Documented POST /v1/event_series/{event_series_id}/ticket_types/{ticket_type_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-event-series-event-series-id-ticket-types-ticket-type-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 holds - Documented POST /v1/holds (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-holds]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 holds hold-id - Documented POST /v1/holds/{hold_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-holds-hold-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 issued-membership-redemptions event-id - Documented POST /v1/issued_membership_redemptions/{event_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-issued-membership-redemptions-event-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 membership-types membership-type-id - Documented POST /v1/membership_types/{membership_type_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-membership-types-membership-type-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 stores store-id - Documented POST /v1/stores/{store_id} (not implemented) [intent=direct_write availability=not_implemented operation=tickettailor.post.v1-stores-store-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    change event series status apply - Plan and execute the change event series status reverse-ETL action [intent=reverse_etl availability=implemented write=change_event_series_status]; approval: requires plan, preview, approval, and execute; risk: changes an event series' publication status; setting to draft/sales_closed immediately stops further public ticket sales; flags: --id (required), --status (required)
    checkout form elements list - Run the checkout form elements ETL stream [intent=etl availability=implemented stream=checkout_form_elements]
    checkout forms list - Run the checkout forms ETL stream [intent=etl availability=implemented stream=checkout_forms]
    confirm order payment received apply - Plan and execute the confirm order payment received reverse-ETL action [intent=reverse_etl availability=implemented write=confirm_order_payment_received]; approval: requires plan, preview, approval, and execute; risk: marks an order (typically an offline/manual payment method) as paid, releasing its tickets from pending status; flags: --id (required)
    create check in apply - Plan and execute the create check in reverse-ETL action [intent=reverse_etl availability=implemented write=create_check_in]; approval: requires plan, preview, approval, and execute; risk: checks an attendee's issued ticket in (or out, when quantity is -1) at the door; low-risk operational mutation, no approval required; flags: --issued_ticket_id (required), --quantity (required)
    create discount apply - Plan and execute the create discount reverse-ETL action [intent=reverse_etl availability=implemented write=create_discount]; approval: requires plan, preview, approval, and execute; risk: creates a discount code redeemable at checkout; low-risk additive external mutation, no approval required; flags: --code (required), --name (required), --type (required)
    create event series apply - Plan and execute the create event series reverse-ETL action [intent=reverse_etl availability=implemented write=create_event_series]; approval: requires plan, preview, approval, and execute; risk: creates a new event series (a recurring/template event definition); low-risk additive external mutation, no approval required; flags: --name (required)
    create issued membership apply - Plan and execute the create issued membership reverse-ETL action [intent=reverse_etl availability=implemented write=create_issued_membership]; approval: requires plan, preview, approval, and execute; risk: issues a new membership directly to a member; low-risk additive external mutation, no approval required; flags: --email (required), --first_name (required), --last_name (required), --membership_type_id (required)
    create issued ticket apply - Plan and execute the create issued ticket reverse-ETL action [intent=reverse_etl availability=implemented write=create_issued_ticket]; approval: requires plan, preview, approval, and execute; risk: issues a new ticket directly (bypassing checkout), consuming inventory from either a ticket type or an existing hold; low-risk additive external mutation, no approval required; flags: --full_name (required)
    create membership type apply - Plan and execute the create membership type reverse-ETL action [intent=reverse_etl availability=implemented write=create_membership_type]; approval: requires plan, preview, approval, and execute; risk: creates a new membership type template; low-risk additive external mutation, no approval required; flags: --name (required), --valid_from_type (required), --valid_to_type (required)
    create product apply - Plan and execute the create product reverse-ETL action [intent=reverse_etl availability=implemented write=create_product]; approval: requires plan, preview, approval, and execute; risk: creates a new sellable add-on product; low-risk additive external mutation, no approval required; flags: --name (required), --price (required)
    create voucher apply - Plan and execute the create voucher reverse-ETL action [intent=reverse_etl availability=implemented write=create_voucher]; approval: requires plan, preview, approval, and execute; risk: creates a new voucher and its redeemable codes; low-risk additive external mutation, no approval required; flags: --name (required), --value (required)
    delete discount apply - Plan and execute the delete discount reverse-ETL action [intent=reverse_etl availability=implemented write=delete_discount]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a discount code; any customer relying on the code at checkout will see it rejected; flags: --id (required)
    delete event series apply - Plan and execute the delete event series reverse-ETL action [intent=reverse_etl availability=implemented write=delete_event_series]; approval: requires plan, preview, approval, and execute; risk: permanently deletes an event series and every event occurrence within it; destructive, approval required; flags: --id (required)
    delete hold apply - Plan and execute the delete hold reverse-ETL action [intent=reverse_etl availability=implemented write=delete_hold]; approval: requires plan, preview, approval, and execute; risk: releases a hold, returning its reserved tickets to public sale immediately; flags: --id (required)
    delete membership type apply - Plan and execute the delete membership type reverse-ETL action [intent=reverse_etl availability=implemented write=delete_membership_type]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a membership type; any issued membership referencing it is orphaned; flags: --id (required)
    delete product apply - Plan and execute the delete product reverse-ETL action [intent=reverse_etl availability=implemented write=delete_product]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a sellable product; it becomes unavailable at checkout immediately; flags: --id (required)
    delete voucher apply - Plan and execute the delete voucher reverse-ETL action [intent=reverse_etl availability=implemented write=delete_voucher]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a voucher and every un-redeemed code issued under it; flags: --id (required)
    discounts list - Run the discounts ETL stream [intent=etl availability=implemented stream=discounts]
    event series list - Run the event series ETL stream [intent=etl availability=implemented stream=event_series]
    event series overrides list - Run the event series overrides ETL stream [intent=etl availability=implemented stream=event_series_overrides]
    event series waitlist signups list - Run the event series waitlist signups ETL stream [intent=etl availability=implemented stream=event_series_waitlist_signups]
    events list - Run the events ETL stream [intent=etl availability=implemented stream=events]
    holds list - Run the holds ETL stream [intent=etl availability=implemented stream=holds]
    issued memberships list - Run the issued memberships ETL stream [intent=etl availability=implemented stream=issued_memberships]
    issued tickets list - Run the issued tickets ETL stream [intent=etl availability=implemented stream=issued_tickets]
    membership types list - Run the membership types ETL stream [intent=etl availability=implemented stream=membership_types]
    orders list - Run the orders ETL stream [intent=etl availability=implemented stream=orders]
    overview list - Run the overview ETL stream [intent=etl availability=implemented stream=overview]
    products list - Run the products ETL stream [intent=etl availability=implemented stream=products]
    stores list - Run the stores ETL stream [intent=etl availability=implemented stream=stores]
    update discount apply - Plan and execute the update discount reverse-ETL action [intent=reverse_etl availability=implemented write=update_discount]; approval: requires plan, preview, approval, and execute; risk: mutates an existing discount code's name, code, or usage limit; changing the code invalidates any already-shared link using the old code; flags: --id (required)
    update event series apply - Plan and execute the update event series reverse-ETL action [intent=reverse_etl availability=implemented write=update_event_series]; approval: requires plan, preview, approval, and execute; risk: mutates an existing event series' public-facing name/description/currency; flags: --id (required)
    update issued membership apply - Plan and execute the update issued membership reverse-ETL action [intent=reverse_etl availability=implemented write=update_issued_membership]; approval: requires plan, preview, approval, and execute; risk: mutates an existing issued membership's holder details or validity window; flags: --id (required)
    update order apply - Plan and execute the update order reverse-ETL action [intent=reverse_etl availability=implemented write=update_order]; approval: requires plan, preview, approval, and execute; risk: mutates an existing order's buyer contact/address details; flags: --id (required)
    update product apply - Plan and execute the update product reverse-ETL action [intent=reverse_etl availability=implemented write=update_product]; approval: requires plan, preview, approval, and execute; risk: mutates an existing product's name, price, or description, directly changing checkout pricing for it; flags: --id (required)
    update voucher apply - Plan and execute the update voucher reverse-ETL action [intent=reverse_etl availability=implemented write=update_voucher]; approval: requires plan, preview, approval, and execute; risk: mutates an existing voucher's value or expiry, directly changing what every un-redeemed code is worth; flags: --id (required)
    void issued membership apply - Plan and execute the void issued membership reverse-ETL action [intent=reverse_etl availability=implemented write=void_issued_membership]; approval: requires plan, preview, approval, and execute; risk: voids an issued membership, invalidating it immediately for entry/redemption; flags: --id (required)
    void issued ticket apply - Plan and execute the void issued ticket reverse-ETL action [intent=reverse_etl availability=implemented write=void_issued_ticket]; approval: requires plan, preview, approval, and execute; risk: voids an issued ticket, invalidating it for entry; optionally returns its inventory to a hold rather than public sale; flags: --id (required)
    void voucher code apply - Plan and execute the void voucher code reverse-ETL action [intent=reverse_etl availability=implemented write=void_voucher_code]; approval: requires plan, preview, approval, and execute; risk: voids a single voucher code, invalidating it for redemption immediately; flags: --id (required), --voucher_id (required)
    voucher codes list - Run the voucher codes ETL stream [intent=etl availability=implemented stream=voucher_codes]
    vouchers list - Run the vouchers ETL stream [intent=etl availability=implemented stream=vouchers]

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
