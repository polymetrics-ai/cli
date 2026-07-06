# Overview

Reads Acuity Scheduling appointments, clients, appointment types, calendars, forms, products,
orders, and labels, and writes appointment/block/certificate mutations, through the Acuity REST API.

Readable streams: `appointments`, `clients`, `appointment_types`, `calendars`, `forms`, `products`,
`orders`, `labels`.

Write actions: `create_appointment`, `update_appointment`, `cancel_appointment`, `create_block`,
`create_certificate`.

Service API documentation: https://developers.acuityscheduling.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://acuityscheduling.com/api/v1`; format `uri`; Acuity
  API base URL override for tests or proxies.
- `mode` (optional, string).
- `password` (required, secret, string); Acuity API key, sent as the HTTP Basic auth password. Never
  logged.
- `username` (required, string); Acuity account User ID, sent as the HTTP Basic auth username.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://acuityscheduling.com/api/v1`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/calendars`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `clients`, `appointment_types`, `calendars`, `forms`, `products`,
`orders`, `labels`; page_number: `appointments`.

- `appointments`: GET `/appointments` - records at response root; page-number pagination; page
  parameter `page`; size parameter `max`; starts at 1; page size 100; computed output fields
  `amount_paid`, `appointment_type_id`, `calendar_id`, `datetime_created`, `end_time`, `first_name`,
  `last_name`.
- `clients`: GET `/clients` - records at response root; computed output fields `first_name`,
  `last_name`.
- `appointment_types`: GET `/appointment-types` - records at response root.
- `calendars`: GET `/calendars` - records at response root.
- `forms`: GET `/forms` - records at response root.
- `products`: GET `/products` - records at response root.
- `orders`: GET `/orders` - records at response root; query `max`=`100`; computed output fields
  `first_name`, `last_name`.
- `labels`: GET `/labels` - records at response root.

## Write actions & risks

Overall write risk: external Acuity Scheduling mutation: creates/updates/cancels live appointments,
blocks off calendar time, and issues package/coupon certificates; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_appointment`: POST `/appointments` - kind `create`; body type `json`; required record
  fields `datetime`, `appointmentTypeID`, `firstName`, `lastName`, `email`; accepted fields
  `appointmentTypeID`, `calendarID`, `certificate`, `datetime`, `email`, `firstName`, `lastName`,
  `phone`, `timezone`; risk: creates a live appointment booking on the calendar and, depending on
  account settings, sends the client a confirmation email/SMS; external mutation, approval required.
- `update_appointment`: PUT `/appointments/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `email`, `firstName`, `id`, `lastName`,
  `notes`, `phone`; risk: updates a live appointment's client-facing details from Acuity's
  white-list of updatable attributes; external mutation, approval required.
- `cancel_appointment`: PUT `/appointments/{{ record.id }}/cancel` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `cancelNote`, `id`,
  `noShow`; risk: permanently cancels a live scheduled appointment; irreversible (Acuity's own docs:
  it is not possible to un-cancel), and by default sends the client a cancellation notification.
  External mutation, approval required.
- `create_block`: POST `/blocks` - kind `create`; body type `json`; required record fields `start`,
  `end`, `calendarID`; accepted fields `calendarID`, `end`, `notes`, `recurring`, `start`; risk:
  blocks off a time range on a live calendar, preventing clients from booking appointments in it;
  external mutation, approval required.
- `create_certificate`: POST `/certificates` - kind `create`; body type `json`; accepted fields
  `certificate`, `couponID`, `email`, `productID`; risk: issues a live, redeemable package or coupon
  certificate code; external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 5 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=1, non_data_endpoint=1, out_of_scope=6,
  requires_elevated_scope=1.
