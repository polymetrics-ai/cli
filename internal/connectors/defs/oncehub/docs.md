# Overview

Reads OnceHub bookings, contacts, booking pages, users, and event types through the OnceHub REST
API.

Readable streams: `bookings`, `contacts`, `booking_pages`, `users`, `event_types`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.oncehub.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); OnceHub API key. Sent as the API-Key header; never logged.
- `base_url` (optional, string); default `https://api.oncehub.com`; format `uri`; OnceHub API base
  URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only bookings with a
  last_updated_time at or after this time are read.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.oncehub.com`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- API key authentication in `API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/users` with query `limit`=`1`.

## Streams notes

Default pagination: follows RFC 5988 Link headers with rel=next.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `bookings`: GET `/v2/bookings` - records path `data`; query `limit`=`{{ config.page_size }}`;
  follows RFC 5988 Link headers with rel=next; incremental cursor `last_updated_time`; sent as
  `last_updated_time.gt`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `contacts`: GET `/v2/contacts` - records path `data`; query `limit`=`{{ config.page_size }}`;
  follows RFC 5988 Link headers with rel=next.
- `booking_pages`: GET `/v2/booking-pages` - records path `data`; query `limit`=`{{ config.page_size
  }}`; follows RFC 5988 Link headers with rel=next.
- `users`: GET `/v2/users` - records path `data`; query `limit`=`{{ config.page_size }}`; follows
  RFC 5988 Link headers with rel=next.
- `event_types`: GET `/v2/event-types` - records path `data`; query `limit`=`{{ config.page_size
  }}`; follows RFC 5988 Link headers with rel=next.

## Write actions & risks

This connector is read-only. Read behavior: external OnceHub API read of scheduling, contact, and
user data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
