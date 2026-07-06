# Overview

Reads Partnerize conversions, campaigns, and publishers through the REST API.

Readable streams: `conversions`, `campaigns`, `publishers`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api-documentation.partnerize.com/.

## Auth setup

Connection fields:

- `application_key` (required, secret, string); Partnerize application key, sent as the Basic auth
  username. Never logged.
- `base_url` (optional, string); default `https://api.partnerize.com/v2`; format `uri`; Partnerize
  API base URL override for tests or proxies.
- `user_api_key` (required, secret, string); Partnerize user API key, sent as the Basic auth
  password. Never logged.

Secret fields are redacted in logs and write previews: `application_key`, `user_api_key`.

Default configuration values: `base_url=https://api.partnerize.com/v2`.

Authentication behavior:

- HTTP Basic authentication using `secrets.application_key`, `secrets.user_api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/conversions` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `conversions`: GET `/conversions` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; incremental cursor `created_at`; formatted as
  `rfc3339`.
- `campaigns`: GET `/campaigns` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; incremental cursor `created_at`; formatted as
  `rfc3339`.
- `publishers`: GET `/publishers` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; incremental cursor `created_at`; formatted as
  `rfc3339`.

## Write actions & risks

This connector is read-only. Read behavior: external Partnerize API read of conversion, campaign,
and publisher data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
