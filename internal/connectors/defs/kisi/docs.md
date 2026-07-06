# Overview

Reads Kisi physical access-control data: members, locks, groups, users, and logins via the Kisi REST
API.

Readable streams: `members`, `locks`, `groups`, `users`, `logins`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.kisi.io/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Kisi API key, sent as "Authorization: KISI-LOGIN <api_key>".
  Never logged.
- `base_url` (optional, string); default `https://api.kisi.io`; format `uri`; Kisi API base URL
  override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.kisi.io`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `KISI-LOGIN` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/members` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `members`: GET `/members` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `locks`: GET `/locks` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `groups`: GET `/groups` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `users`: GET `/users` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.
- `logins`: GET `/logins` - records at response root; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Kisi API read of physical access-control data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=2.
