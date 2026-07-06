# Overview

Reads and writes events, orders, issued tickets, event series, holds, discounts, memberships,
products, stores, and vouchers through the Ticket Tailor API.

Readable streams: `events`, `orders`, `issued_tickets`, `event_series`, `holds`, `discounts`,
`membership_types`, `issued_memberships`, `products`, `stores`, `vouchers`, `checkout_forms`,
`voucher_codes`, `checkout_form_elements`, `event_series_overrides`,
`event_series_waitlist_signups`, `overview`.

Write actions: `create_event_series`, `update_event_series`, `delete_event_series`,
`change_event_series_status`, `create_discount`, `update_discount`, `delete_discount`,
`delete_hold`, `create_check_in`, `create_issued_ticket`, `void_issued_ticket`, `update_order`,
`confirm_order_payment_received`, `create_membership_type`, `delete_membership_type`,
`create_issued_membership`, `update_issued_membership`, `void_issued_membership`, `create_voucher`,
`update_voucher`, `delete_voucher`, `void_voucher_code`, `create_product`, `update_product`,
`delete_product`.

Service API documentation: https://developers.tickettailor.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Ticket Tailor API key, sent as the username of HTTP Basic
  auth with an empty password (Authorization: Basic base64(<api_key>:)). Never logged.
- `base_url` (optional, string); default `https://api.tickettailor.com/v1`; format `uri`; Ticket
  Tailor API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.tickettailor.com/v1`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/events` with query `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `links.next`;
cross-host next URLs are allowed.

Pagination by stream: next_url: `events`, `orders`, `issued_tickets`, `event_series`, `holds`,
`discounts`, `membership_types`, `issued_memberships`, `products`, `stores`, `vouchers`,
`checkout_forms`, `voucher_codes`, `checkout_form_elements`, `event_series_overrides`,
`event_series_waitlist_signups`; none: `overview`.

- `events`: GET `/events` - records path `data`; query `limit`=`100`; follows a next-page URL from
  the response body; URL path `links.next`; cross-host next URLs are allowed; emits passthrough
  records.
- `orders`: GET `/orders` - records path `data`; query `limit`=`100`; follows a next-page URL from
  the response body; URL path `links.next`; cross-host next URLs are allowed; emits passthrough
  records.
- `issued_tickets`: GET `/issued_tickets` - records path `data`; query `limit`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; cross-host next URLs are allowed;
  emits passthrough records.
- `event_series`: GET `/event_series` - records path `data`; query `limit`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; cross-host next URLs are allowed;
  emits passthrough records.
- `holds`: GET `/holds` - records path `data`; query `limit`=`100`; follows a next-page URL from the
  response body; URL path `links.next`; cross-host next URLs are allowed; emits passthrough records.
- `discounts`: GET `/discounts` - records path `data`; query `limit`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; cross-host next URLs are allowed; emits passthrough
  records.
- `membership_types`: GET `/membership_types` - records path `data`; query `limit`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; cross-host next URLs are allowed;
  emits passthrough records.
- `issued_memberships`: GET `/issued_memberships` - records path `data`; query `limit`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; cross-host next URLs are
  allowed; emits passthrough records.
- `products`: GET `/products` - records path `data`; query `limit`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; cross-host next URLs are allowed; emits passthrough
  records.
- `stores`: GET `/stores` - records path `data`; query `limit`=`100`; follows a next-page URL from
  the response body; URL path `links.next`; cross-host next URLs are allowed; emits passthrough
  records.
- `vouchers`: GET `/vouchers` - records path `data`; query `limit`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; cross-host next URLs are allowed; emits passthrough
  records.
- `checkout_forms`: GET `/checkout_forms` - records path `data`; query `limit`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; cross-host next URLs are allowed;
  emits passthrough records.
- `voucher_codes`: GET `/vouchers/{{ fanout.id }}/codes` - records path `data`; query `limit`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; cross-host next URLs are
  allowed; fan-out; ids from request `/vouchers`; id-list records path `data`; id field `id`; id
  inserted into the request path; stamps `voucher_id`; emits passthrough records.
- `checkout_form_elements`: GET `/checkout_forms/{{ fanout.id }}/elements` - records path `data`;
  query `limit`=`100`; follows a next-page URL from the response body; URL path `links.next`;
  cross-host next URLs are allowed; fan-out; ids from request `/checkout_forms`; id-list records
  path `data`; id field `id`; id inserted into the request path; stamps `checkout_form_id`; emits
  passthrough records.
- `event_series_overrides`: GET `/event_series/{{ fanout.id }}/overrides` - records path `data`;
  query `limit`=`100`; follows a next-page URL from the response body; URL path `links.next`;
  cross-host next URLs are allowed; fan-out; ids from request `/event_series`; id-list records path
  `data`; id field `id`; id inserted into the request path; stamps `event_series_id`; emits
  passthrough records.
- `event_series_waitlist_signups`: GET `/event_series/{{ fanout.id }}/waitlist_signups` - records
  path `data`; query `limit`=`100`; follows a next-page URL from the response body; URL path
  `links.next`; cross-host next URLs are allowed; fan-out; ids from request `/event_series`; id-list
  records path `data`; id field `id`; id inserted into the request path; stamps `event_series_id`;
  emits passthrough records.
- `overview`: GET `/overview` - single-object response; records path `.`; computed output fields
  `id`; emits passthrough records.

## Write actions & risks

Overall write risk: external Ticket Tailor API mutations covering event
series/hold/discount/membership/voucher/product lifecycle, ticket issuance/voiding/check-in, and
order payment confirmation; delete_event_series is destructive/confirm-gated.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_event_series`: POST `/event_series` - kind `create`; body type `form`; required record
  fields `name`; accepted fields `access_code`, `country`, `currency`, `description`, `name`,
  `postal_code`, `venue`; risk: creates a new event series (a recurring/template event definition);
  low-risk additive external mutation, no approval required.
- `update_event_series`: POST `/event_series/{{ record.id }}` - kind `update`; body type `form`;
  path fields `id`; required record fields `id`; accepted fields `call_to_action`, `currency`,
  `description`, `id`, `name`; risk: mutates an existing event series' public-facing
  name/description/currency.
- `delete_event_series`: DELETE `/event_series/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: permanently deletes an event series
  and every event occurrence within it; destructive, approval required.
- `change_event_series_status`: POST `/event_series/{{ record.id }}/status` - kind `update`; body
  type `form`; path fields `id`; body fields `status`; required record fields `id`, `status`;
  accepted fields `id`, `status`; risk: changes an event series' publication status; setting to
  draft/sales_closed immediately stops further public ticket sales.
- `create_discount`: POST `/discounts` - kind `create`; body type `form`; required record fields
  `name`, `code`, `type`; accepted fields `code`, `expires`, `max_redemptions`, `name`, `price`,
  `price_percent`, `type`; risk: creates a discount code redeemable at checkout; low-risk additive
  external mutation, no approval required.
- `update_discount`: POST `/discounts/{{ record.id }}` - kind `update`; body type `form`; path
  fields `id`; required record fields `id`; accepted fields `code`, `id`, `max_redemptions`, `name`;
  risk: mutates an existing discount code's name, code, or usage limit; changing the code
  invalidates any already-shared link using the old code.
- `delete_discount`: DELETE `/discounts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently deletes a discount code; any customer relying on the code at
  checkout will see it rejected.
- `delete_hold`: DELETE `/holds/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: releases a hold, returning its reserved tickets to public sale immediately.
- `create_check_in`: POST `/check_ins` - kind `create`; body type `form`; required record fields
  `issued_ticket_id`, `quantity`; accepted fields `check_in_at`, `issued_ticket_id`,
  `local_unique_id`, `quantity`; risk: checks an attendee's issued ticket in (or out, when quantity
  is -1) at the door; low-risk operational mutation, no approval required.
- `create_issued_ticket`: POST `/issued_tickets` - kind `create`; body type `form`; required record
  fields `full_name`; accepted fields `email`, `event_id`, `full_name`, `hold_id`, `send_email`,
  `ticket_type_id`; risk: issues a new ticket directly (bypassing checkout), consuming inventory
  from either a ticket type or an existing hold; low-risk additive external mutation, no approval
  required.
- `void_issued_ticket`: POST `/issued_tickets/{{ record.id }}/void` - kind `update`; body type
  `form`; path fields `id`; required record fields `id`; accepted fields `id`, `void_to_hold`; risk:
  voids an issued ticket, invalidating it for entry; optionally returns its inventory to a hold
  rather than public sale.
- `update_order`: POST `/orders/{{ record.id }}` - kind `update`; body type `form`; path fields
  `id`; required record fields `id`; accepted fields `address_1`, `email`, `first_name`, `id`,
  `last_name`, `phone`, `postal_code`; risk: mutates an existing order's buyer contact/address
  details.
- `confirm_order_payment_received`: POST `/orders/{{ record.id }}/confirm-payment-received` - kind
  `update`; body type `form`; path fields `id`; required record fields `id`; accepted fields `id`,
  `transaction_id`; risk: marks an order (typically an offline/manual payment method) as paid,
  releasing its tickets from pending status.
- `create_membership_type`: POST `/membership_types` - kind `create`; body type `form`; required
  record fields `name`, `valid_from_type`, `valid_to_type`; accepted fields `max_redemptions`,
  `name`, `photo_required`, `valid_from_type`, `valid_to_type`; risk: creates a new membership type
  template; low-risk additive external mutation, no approval required.
- `delete_membership_type`: DELETE `/membership_types/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently deletes a membership type; any issued
  membership referencing it is orphaned.
- `create_issued_membership`: POST `/issued_memberships` - kind `create`; body type `form`; required
  record fields `membership_type_id`, `first_name`, `last_name`, `email`; accepted fields `email`,
  `first_name`, `last_name`, `membership_type_id`, `valid_from_date`, `valid_to_date`; risk: issues
  a new membership directly to a member; low-risk additive external mutation, no approval required.
- `update_issued_membership`: POST `/issued_memberships/{{ record.id }}` - kind `update`; body type
  `form`; path fields `id`; required record fields `id`; accepted fields `email`, `first_name`,
  `id`, `last_name`, `valid_to_date`; risk: mutates an existing issued membership's holder details
  or validity window.
- `void_issued_membership`: POST `/issued_memberships/{{ record.id }}/void` - kind `update`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: voids an
  issued membership, invalidating it immediately for entry/redemption.
- `create_voucher`: POST `/vouchers` - kind `create`; body type `form`; required record fields
  `name`, `value`; accepted fields `codes`, `expiry`, `name`, `usable_on_any_event`, `value`,
  `voucher_type`; risk: creates a new voucher and its redeemable codes; low-risk additive external
  mutation, no approval required.
- `update_voucher`: POST `/vouchers/{{ record.id }}` - kind `update`; body type `form`; path fields
  `id`; required record fields `id`; accepted fields `expiry`, `id`, `name`, `value`; risk: mutates
  an existing voucher's value or expiry, directly changing what every un-redeemed code is worth.
- `delete_voucher`: DELETE `/vouchers/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently deletes a voucher and every un-redeemed code issued under it.
- `void_voucher_code`: POST `/vouchers/{{ record.voucher_id }}/codes/{{ record.id }}/void` - kind
  `update`; body type `none`; path fields `voucher_id`, `id`; required record fields `voucher_id`,
  `id`; accepted fields `id`, `voucher_id`; risk: voids a single voucher code, invalidating it for
  redemption immediately.
- `create_product`: POST `/products` - kind `create`; body type `form`; required record fields
  `name`, `price`; accepted fields `booking_fee`, `currency`, `description`, `name`, `price`; risk:
  creates a new sellable add-on product; low-risk additive external mutation, no approval required.
- `update_product`: POST `/products/{{ record.id }}` - kind `update`; body type `form`; path fields
  `id`; required record fields `id`; accepted fields `description`, `id`, `name`, `price`; risk:
  mutates an existing product's name, price, or description, directly changing checkout pricing for
  it.
- `delete_product`: DELETE `/products/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently deletes a sellable product; it becomes unavailable at checkout
  immediately.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 17 stream-backed endpoint group(s), 25 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=13, non_data_endpoint=2, out_of_scope=23.
