# Overview

Reads PagerDuty incidents, users, services, and teams through the REST API.

Readable streams: `incidents`, `users`, `services`, `teams`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.pagerduty.com/api-reference/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); PagerDuty REST API key, sent as the Authorization header
  (Authorization: Token token=<api_key>). Never logged.
- `base_url` (optional, string); default `https://api.pagerduty.com`; format `uri`; PagerDuty API
  base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.pagerduty.com`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token token=` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/incidents` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `incidents`: GET `/incidents` - records path `incidents`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100.
- `users`: GET `/users` - records path `users`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.
- `services`: GET `/services` - records path `services`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `teams`: GET `/teams` - records path `teams`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external PagerDuty API read of incident, user, service,
and team data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, non_data_endpoint=1, out_of_scope=4.
