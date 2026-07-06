# Overview

Reads Gainsight PX accounts, users, features, and segments through the aptrinsic REST API
(read-only).

Readable streams: `accounts`, `users`, `feature`, `segments`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://support.gainsight.com/PX/API_for_Developers/02Usage_of_Different_APIs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Gainsight PX (aptrinsic) REST API key. Sent as the
  X-APTRINSIC-API-KEY header; never logged.
- `base_url` (optional, string); default `https://api.aptrinsic.com/v1`; format `uri`; Gainsight PX
  API base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-500).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.aptrinsic.com/v1`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- API key authentication in `X-APTRINSIC-API-KEY` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `scrollId`; next token from `scrollId`.

- `accounts`: GET `/accounts` - records path `accounts`; query `pageSize`=`100`; cursor pagination;
  cursor parameter `scrollId`; next token from `scrollId`.
- `users`: GET `/users` - records path `users`; query `pageSize`=`100`; cursor pagination; cursor
  parameter `scrollId`; next token from `scrollId`.
- `feature`: GET `/feature` - records path `features`; query `pageSize`=`100`; cursor pagination;
  cursor parameter `scrollId`; next token from `scrollId`.
- `segments`: GET `/segment` - records path `segments`; query `pageSize`=`100`; cursor pagination;
  cursor parameter `scrollId`; next token from `scrollId`.

## Write actions & risks

This connector is read-only. Read behavior: external Gainsight PX (aptrinsic) API read of account,
user, feature, and segment data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
