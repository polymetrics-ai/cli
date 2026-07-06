# Overview

Reads Gridly views, per-view records (with flattened column cells), and per-view branches through
the Gridly REST API.

Readable streams: `views`, `records`, `branches`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.gridly.com/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Gridly API key, sent as the Authorization header ("ApiKey
  <key>"). Never logged.
- `base_url` (optional, string); default `https://api.gridly.com/v1`; format `uri`; Gridly API base
  URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages per view; use 0, all, or unlimited to
  exhaust each stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000).
- `view_ids` (optional, string); Comma-separated list of Gridly view ids to fan out over for the
  records and branches streams (one request per id). Required for records/branches; not used by the
  views stream, which lists all views in one request.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.gridly.com/v1`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `ApiKey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `views` with query `page`=`1`; `pageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1; page size 100.

- `views`: GET `views` - records at response root; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1; page size 100.
- `records`: GET `views/{{ fanout.id }}/records` - records at response root; page-number pagination;
  page parameter `page`; size parameter `pageSize`; starts at 1; page size 100; fan-out; ids from
  config field `view_ids`; id inserted into the request path; stamps `view_id`.
- `branches`: GET `views/{{ fanout.id }}/branches` - records at response root; page-number
  pagination; page parameter `page`; size parameter `pageSize`; starts at 1; page size 100; fan-out;
  ids from config field `view_ids`; id inserted into the request path; stamps `view_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Gridly API read of view/grid content.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
