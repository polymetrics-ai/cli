# Overview

Reads GoLogin browser profiles, folders, tags, and account information through the GoLogin REST API.

Readable streams: `profiles`, `folders`, `user`, `tags`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.gologin.com/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); GoLogin API key. Used only for Bearer auth (Authorization:
  Bearer <api_key>); never logged.
- `base_url` (optional, string); default `https://api.gologin.com`; format `uri`; GoLogin API base
  URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.gologin.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/browser/v2` with query `page`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `folders`, `user`, `tags`; page_number: `profiles`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `profiles`: GET `/browser/v2` - records path `profiles`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 30; incremental cursor `updatedAt`;
  formatted as `rfc3339`.
- `folders`: GET `/folders` - records at response root.
- `user`: GET `/user` - records at response root; incremental cursor `createdAt`; formatted as
  `rfc3339`.
- `tags`: GET `/tags/all` - records path `tags`.

## Write actions & risks

This connector is read-only. Read behavior: external GoLogin API read of browser profile and account
data.

## Known limits

- Batch defaults: read_page_size=30.
- API coverage includes 4 stream-backed endpoint group(s).
