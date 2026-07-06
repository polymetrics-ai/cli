# Overview

Reads Workday REST API resources (workers, organizations, job profiles) with bearer-token
authentication. Read-only.

Readable streams: `workers`, `organizations`, `jobs`.

This connector is read-only; no write actions are declared.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Workday REST API bearer access token (Authorization:
  Bearer <access_token>). Never logged.
- `base_url` (optional, string); default `https://wd2-impl-services1.workday.com`; format `uri`;
  Workday API base URL override for a tenant's actual Workday instance, tests, or proxies.
- `tenant` (required, string); Workday tenant name, substituted as a path segment into every
  stream's resource URL (e.g. ccx/api/hcm/v1/<tenant>/workers).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://wd2-impl-services1.workday.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/ccx/api/hcm/v1/{{ config.tenant }}/workers` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100; maximum 1 page(s).

- `workers`: GET `/ccx/api/hcm/v1/{{ config.tenant }}/workers` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1
  page(s); emits passthrough records.
- `organizations`: GET `/ccx/api/hcm/v1/{{ config.tenant }}/organizations` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  maximum 1 page(s); emits passthrough records.
- `jobs`: GET `/ccx/api/hcm/v1/{{ config.tenant }}/jobs` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; maximum 1
  page(s); emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Workday REST API read of worker, organization,
and job profile data (HR/PII-adjacent).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=1.
