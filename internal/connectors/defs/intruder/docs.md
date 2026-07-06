# Overview

Reads Intruder issues, issue occurrences, scans, and targets through the Intruder REST API
(read-only, full refresh).

Readable streams: `issues`, `scans`, `targets`, `occurrences`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.intruder.io/docs/welcome.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Intruder API access token. Used only for Bearer auth;
  never logged.
- `base_url` (optional, string); default `https://api.intruder.io/v1`; format `uri`; Intruder API
  base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.intruder.io/v1`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/targets` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `issues`: GET `/issues` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `scans`: GET `/scans` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `targets`: GET `/targets` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `occurrences`: GET `/issues/{{ fanout.id }}/occurrences` - records path `results`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; fan-out; ids from
  request `/issues`; id-list records path `results`; id field `id`; id inserted into the request
  path; stamps `issue_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Intruder API read of vulnerability issues,
issue occurrences, scans, and target data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
