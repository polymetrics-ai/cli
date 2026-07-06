# Overview

Reads Guru collections, groups, members, and teams through the Guru REST API using HTTP Basic
authentication (email + API token).

Readable streams: `collections`, `groups`, `members`, `teams`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.getguru.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.getguru.com/api/v1`; format `uri`; Guru API
  base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-250).
- `password` (required, secret, string); Guru API token, used as the HTTP Basic auth password. Never
  logged.
- `username` (required, string); Guru account email address, used as the HTTP Basic auth username.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://api.getguru.com/api/v1`, `max_pages=0`,
`page_size=50`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/collections`.

## Streams notes

Default pagination: follows RFC 5988 Link headers with rel=next.

- `collections`: GET `/collections` - records at response root; query `pageSize`=`{{
  config.page_size }}`; follows RFC 5988 Link headers with rel=next.
- `groups`: GET `/groups` - records at response root; query `pageSize`=`{{ config.page_size }}`;
  follows RFC 5988 Link headers with rel=next.
- `members`: GET `/members` - records at response root; query `pageSize`=`{{ config.page_size }}`;
  follows RFC 5988 Link headers with rel=next; computed output fields `email`, `firstName`, `id`,
  `lastName`.
- `teams`: GET `/teams` - records at response root; query `pageSize`=`{{ config.page_size }}`;
  follows RFC 5988 Link headers with rel=next.

## Write actions & risks

This connector is read-only. Read behavior: external Guru API read of collections, groups, members,
and teams.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
