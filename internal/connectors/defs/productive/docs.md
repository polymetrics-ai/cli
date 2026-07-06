# Overview

Reads Productive projects, people, companies, and tasks through the Productive JSON:API-style REST
API (read-only).

Readable streams: `projects`, `people`, `companies`, `tasks`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.productive.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Productive API token, sent as the X-Auth-Token header. Never
  logged.
- `base_url` (optional, string); default `https://api.productive.io/api/v2`; format `uri`;
  Productive API base URL. Defaults to https://api.productive.io/api/v2.
- `organization_id` (required, string); Productive organization id, sent as the X-Organization-Id
  header on every request. Required by Productive's API for every call.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.productive.io/api/v2`.

Authentication behavior:

- API key authentication in `X-Auth-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects` with query `page`=`1`; `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `projects`: GET `/projects` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`; formatted
  as `rfc3339`; computed output fields `created_at`, `name`, `type`, `updated_at`.
- `people`: GET `/people` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`; formatted as
  `rfc3339`; computed output fields `created_at`, `name`, `type`, `updated_at`.
- `companies`: GET `/companies` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`;
  formatted as `rfc3339`; computed output fields `created_at`, `name`, `type`, `updated_at`.
- `tasks`: GET `/tasks` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`; formatted as
  `rfc3339`; computed output fields `created_at`, `name`, `type`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Productive API read of projects, people,
companies, and tasks.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=5.
