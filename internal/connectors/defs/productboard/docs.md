# Overview

Reads Productboard features, notes, components, and products through the public API.

Readable streams: `features`, `notes`, `components`, `products`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.productboard.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Productboard API access token, sent as 'Authorization:
  Bearer <access_token>'. Never logged.
- `base_url` (optional, string); default `https://api.productboard.com`; format `uri`; Productboard
  API base URL override for tests or proxies.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound sent as 'updated_since'
  on the first request of every stream, when set.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.productboard.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/features` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: single request; no pagination.

- `features`: GET `/features` - records path `data`; query `limit`=`100`; `updated_since` from
  template `{{ config.start_date }}`, omitted when absent; follows a next-page URL from the response
  body; URL path `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `notes`: GET `/notes` - records path `data`; query `limit`=`100`; `updated_since` from template
  `{{ config.start_date }}`, omitted when absent; follows a next-page URL from the response body;
  URL path `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `components`: GET `/components` - records path `data`; query `limit`=`100`; `updated_since` from
  template `{{ config.start_date }}`, omitted when absent; follows a next-page URL from the response
  body; URL path `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `products`: GET `/products` - records path `data`; query `limit`=`100`; `updated_since` from
  template `{{ config.start_date }}`, omitted when absent; follows a next-page URL from the response
  body; URL path `links.next`; next URLs stay on the configured API host; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Productboard API read of feature, note,
component, and product data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
