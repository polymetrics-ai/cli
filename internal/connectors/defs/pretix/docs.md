# Overview

Reads pretix organizers, events, items, and orders through the pretix REST API.

Readable streams: `organizers`, `events`, `items`, `orders`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.pretix.eu/en/latest/api/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Pretix API token, sent as 'Authorization: Token
  <api_token>'. Never logged.
- `base_url` (optional, string); default `https://pretix.eu/api/v1`; format `uri`; Pretix API base
  URL override for self-hosted instances, tests, or proxies.
- `event` (optional, string); Pretix event slug, required for the items and orders streams
  (substituted into the event-scoped path).
- `organizer` (optional, string); Pretix organizer slug, required for the events, items, and orders
  streams (substituted into the organizer-scoped path).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://pretix.eu/api/v1`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/organizers/`.

## Streams notes

Default pagination: single request; no pagination.

- `organizers`: GET `/organizers/` - records path `results`; query `page_size`=`100`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host;
  computed output fields `id`.
- `events`: GET `/organizers/{{ config.organizer }}/events/` - records path `results`; query
  `page_size`=`100`; follows a next-page URL from the response body; URL path `next`; next URLs stay
  on the configured API host; computed output fields `id`, `updated_at`.
- `items`: GET `/organizers/{{ config.organizer }}/events/{{ config.event }}/items/` - records path
  `results`; query `page_size`=`100`; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host.
- `orders`: GET `/organizers/{{ config.organizer }}/events/{{ config.event }}/orders/` - records
  path `results`; query `page_size`=`100`; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host; computed output fields `id`.

## Write actions & risks

This connector is read-only. Read behavior: external pretix API read of organizer, event, item, and
order data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
