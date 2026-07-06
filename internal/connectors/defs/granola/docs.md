# Overview

Reads Granola meeting notes metadata and full note detail (summary, owner, attendees, calendar
event) through the Granola public API (read-only).

Readable streams: `notes`, `detailed_notes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.granola.ai/introduction.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Granola public API key (grn_ prefix), sent as Authorization:
  Bearer <api_key>; never logged.
- `base_url` (optional, string); default `https://public-api.granola.ai/v1`; format `uri`; Granola
  public API base URL override, for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `30`; Records per page (1-30).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound applied as the notes
  stream's created_after filter on a fresh sync; superseded by the persisted incremental cursor on
  repeat syncs.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://public-api.granola.ai/v1`, `page_size=30`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/notes` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop
flag `hasMore`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `notes`: GET `/notes` - records path `notes`; query `limit` from template `{{ config.page_size
  }}`, default `30`; cursor pagination; cursor parameter `cursor`; next token from `cursor`; stop
  flag `hasMore`; incremental cursor `created_at`; sent as `created_after`; formatted as `rfc3339`;
  initial lower bound from `start_date`; computed output fields `owner_email`, `owner_name`.
- `detailed_notes`: GET `/notes/{{ fanout.id }}` - records at response root; query
  `include`=`transcript`; cursor pagination; cursor parameter `cursor`; next token from `cursor`;
  stop flag `hasMore`; computed output fields `owner_email`, `owner_name`; fan-out; ids from request
  `/notes`; id-list records path `notes`; id field `id`; id inserted into the request path.

## Write actions & risks

This connector is read-only. Read behavior: external Granola API read of meeting notes metadata.

## Known limits

- Batch defaults: read_page_size=30.
- API coverage includes 2 stream-backed endpoint group(s).
