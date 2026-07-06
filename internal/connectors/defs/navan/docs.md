# Overview

Reads Navan flight, hotel, car, and rail travel bookings through the Navan REST API using OAuth2
client-credentials authentication.

Readable streams: `bookings`, `hotel_bookings`, `car_bookings`, `rail_bookings`.

This connector is read-only; no write actions are declared.

Service API documentation: https://navan.com/api-docs.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.navan.com`; format `uri`; Defaults to
  production.
- `client_id` (required, secret, string); Navan OAuth2 client-credentials client ID. Used only for
  the token exchange; never logged.
- `client_secret` (required, secret, string); Navan OAuth2 client-credentials client secret. Used
  only for the token exchange; never logged.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound sent as the createdFrom
  filter; only bookings created at or after this time are read. Optional; omitted entirely on a full
  sync with no prior cursor.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.navan.com`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.base_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/bookings` with query `bookingType`=`FLIGHT`; `page`=`0`; `size`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `size`; starts at
0; page size 50.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `bookings`: GET `/v1/bookings` - records path `data`; query `bookingType`=`FLIGHT`; page-number
  pagination; page parameter `page`; size parameter `size`; starts at 0; page size 50; incremental
  cursor `last_modified`; sent as `createdFrom`; formatted as `rfc3339`; initial lower bound from
  `start_date`; computed output fields `approval_status`, `base_price`, `booking_fee`, `booking_id`,
  `booking_method`, `booking_status`, `booking_type`, `cancelled_at`, `confirmation_number`,
  `end_date`, `grand_total`, `last_modified`, `start_date`.
- `hotel_bookings`: GET `/v1/bookings` - records path `data`; query `bookingType`=`HOTEL`;
  page-number pagination; page parameter `page`; size parameter `size`; starts at 0; page size 50;
  incremental cursor `last_modified`; sent as `createdFrom`; formatted as `rfc3339`; initial lower
  bound from `start_date`; computed output fields `approval_status`, `base_price`, `booking_fee`,
  `booking_id`, `booking_method`, `booking_status`, `booking_type`, `cancelled_at`,
  `confirmation_number`, `end_date`, `grand_total`, `last_modified`, `start_date`.
- `car_bookings`: GET `/v1/bookings` - records path `data`; query `bookingType`=`CAR`; page-number
  pagination; page parameter `page`; size parameter `size`; starts at 0; page size 50; incremental
  cursor `last_modified`; sent as `createdFrom`; formatted as `rfc3339`; initial lower bound from
  `start_date`; computed output fields `approval_status`, `base_price`, `booking_fee`, `booking_id`,
  `booking_method`, `booking_status`, `booking_type`, `cancelled_at`, `confirmation_number`,
  `end_date`, `grand_total`, `last_modified`, `start_date`.
- `rail_bookings`: GET `/v1/bookings` - records path `data`; query `bookingType`=`RAIL`; page-number
  pagination; page parameter `page`; size parameter `size`; starts at 0; page size 50; incremental
  cursor `last_modified`; sent as `createdFrom`; formatted as `rfc3339`; initial lower bound from
  `start_date`; computed output fields `approval_status`, `base_price`, `booking_fee`, `booking_id`,
  `booking_method`, `booking_status`, `booking_type`, `cancelled_at`, `confirmation_number`,
  `end_date`, `grand_total`, `last_modified`, `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external Navan API read of travel booking data (flight,
hotel, car, rail).

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
