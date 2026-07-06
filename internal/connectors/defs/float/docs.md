# Overview

Reads Float people, projects, clients, tasks, and departments through the Float v3 REST API.

Readable streams: `people`, `projects`, `clients`, `tasks`, `departments`.

This connector is read-only; no write actions are declared.

Service API documentation: https://dev.float.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Float personal access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `base_url` (optional, string); default `https://api.float.com/v3`; format `uri`; Float API base
  URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `200`; Records per page (1-200), sent as the 'per-page'
  query param.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.float.com/v3`, `page_size=200`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/departments` with query `per-page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per-page`; starts
at 1; page size 200.

- `people`: GET `/people` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `per-page`; starts at 1; page size 200.
- `projects`: GET `/projects` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `per-page`; starts at 1; page size 200.
- `clients`: GET `/clients` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `per-page`; starts at 1; page size 200.
- `tasks`: GET `/tasks` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `per-page`; starts at 1; page size 200.
- `departments`: GET `/departments` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `per-page`; starts at 1; page size 200.

## Write actions & risks

This connector is read-only. Read behavior: external Float API read of resource-planning and
staffing data.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
