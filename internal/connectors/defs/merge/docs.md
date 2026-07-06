# Overview

Reads Merge ATS common-model objects (candidates, applications, jobs, offers, departments, users)
through the Merge unified REST API.

Readable streams: `candidates`, `applications`, `jobs`, `offers`, `departments`, `users`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.merge.dev/api-reference/.

## Auth setup

Connection fields:

- `account_token` (required, secret, string); Merge per-linked-account token selecting which
  connected end-customer account to read, sent as the X-Account-Token header. Never logged.
- `api_token` (required, secret, string); Merge account-wide API token, sent as Bearer auth
  (Authorization: Bearer <api_token>). Never logged.
- `base_url` (optional, string); default `https://api.merge.dev/api/ats/v1`; format `uri`; Merge ATS
  Common Model v1 API base URL. Override for other Merge categories (hris, accounting, ...) or test
  servers.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only objects modified at
  or after this time are read on a fresh sync (no persisted cursor yet).

Secret fields are redacted in logs and write previews: `account_token`, `api_token`.

Default configuration values: `base_url=https://api.merge.dev/api/ats/v1`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/candidates`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `next`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `candidates`: GET `/candidates` - records path `results`; query `page_size`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `next`; incremental cursor `modified_at`;
  sent as `modified_after`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `applications`: GET `/applications` - records path `results`; query `page_size`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `next`; incremental cursor `modified_at`;
  sent as `modified_after`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `jobs`: GET `/jobs` - records path `results`; query `page_size`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `next`; incremental cursor `modified_at`; sent as
  `modified_after`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `offers`: GET `/offers` - records path `results`; query `page_size`=`100`; cursor pagination;
  cursor parameter `cursor`; next token from `next`; incremental cursor `modified_at`; sent as
  `modified_after`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `departments`: GET `/departments` - records path `results`; query `page_size`=`100`; cursor
  pagination; cursor parameter `cursor`; next token from `next`; incremental cursor `modified_at`;
  sent as `modified_after`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `users`: GET `/users` - records path `results`; query `page_size`=`100`; cursor pagination; cursor
  parameter `cursor`; next token from `next`; incremental cursor `modified_at`; sent as
  `modified_after`; formatted as `rfc3339`; initial lower bound from `start_date`.

## Write actions & risks

This connector is read-only. Read behavior: external Merge unified API read of ATS candidate and
hiring data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=5.
