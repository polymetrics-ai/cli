# Overview

Reads Workday tenant data (workers, organizations, positions) through conservative Workday API
endpoints. Read-only.

Readable streams: `workers`, `organizations`, `positions`.

This connector is read-only; no write actions are declared.

Service API documentation: https://community.workday.com/sites/default/files/file-hosting/restapi/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://wd2-impl-services1.workday.com`; format `uri`;
  Workday API base URL override for a tenant's actual Workday instance, tests, or proxies.
- `password` (required, secret, string); Workday tenant API password, sent via HTTP Basic auth.
  Never logged.
- `tenant` (required, string); Workday tenant name, substituted as a path segment into every
  stream's resource URL (e.g. ccx/api/v1/<tenant>/workers).
- `username` (required, secret, string); Workday tenant API username, sent via HTTP Basic auth.
  Never logged.

Secret fields are redacted in logs and write previews: `password`, `username`.

Default configuration values: `base_url=https://wd2-impl-services1.workday.com`.

Authentication behavior:

- HTTP Basic authentication using `secrets.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/ccx/api/v1/{{ config.tenant }}/workers` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100; maximum 1 page(s).

- `workers`: GET `/ccx/api/v1/{{ config.tenant }}/workers` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1
  page(s); emits passthrough records.
- `organizations`: GET `/ccx/api/v1/{{ config.tenant }}/organizations` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  maximum 1 page(s); emits passthrough records.
- `positions`: GET `/ccx/api/v1/{{ config.tenant }}/positions` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1
  page(s); emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Workday tenant API read of worker,
organization, and position data (HR/PII-adjacent).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, requires_elevated_scope=1.
