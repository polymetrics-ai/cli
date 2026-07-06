# Overview

Reads Fulcrum forms, records, projects, choice lists, and classification sets through the Fulcrum
REST API v2.

Readable streams: `forms`, `records`, `projects`, `choice_lists`, `classification_sets`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.fulcrumapp.com/reference.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Fulcrum REST API v2 token. Sent as the X-ApiToken header;
  never logged.
- `base_url` (optional, string); default `https://api.fulcrumapp.com/api/v2`; format `uri`; Fulcrum
  API base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-20000).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.fulcrumapp.com/api/v2`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- API key authentication in `X-ApiToken` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/forms.json`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

- `forms`: GET `/forms.json` - records path `forms`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `records`: GET `/records.json` - records path `records`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `projects`: GET `/projects.json` - records path `projects`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `choice_lists`: GET `/choice_lists.json` - records path `choice_lists`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `classification_sets`: GET `/classification_sets.json` - records path `classification_sets`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100.

## Write actions & risks

This connector is read-only. Read behavior: external Fulcrum API read of form, record, and project
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
