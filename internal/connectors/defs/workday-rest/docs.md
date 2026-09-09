# Overview

Reads and writes the documented Workday REST surface across the 52 independently versioned services
Workday's own directory publishes, with bearer-token authentication.

Current official operation ledger: 911 documented HTTP operations (651 GET, 153 POST, 56 PATCH, 32
DELETE, 19 PUT). Implemented rows: 911 commands = 654 bounded direct reads + 252 typed writes + 5
binary downloads. Those commands, together with the 3 stream-backed reads, classify 910 documented
endpoints; 1 endpoint is blocked as deprecated. Validated rows: 0 (fixture-only; no live provider
calls were made).

Readable streams: `workers`, `organizations`, `jobs`.

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

The connector declares 252 typed write actions (145 POST creates, 56 PATCH updates, 32 DELETE
removals, 19 PUT upserts) across the HCM, Financials, Student and Platform services, including
HR/PII-adjacent worker, absence, payroll, and financial records.

Writes are only available through reverse ETL plan -> preview -> explicit approval -> execute. Every
DELETE action is gated as destructive and additionally requires a typed confirmation. The bundle
does not expose arbitrary request bodies, raw query strings, generic method/path/body, file bytes,
shell commands, or passthrough HTTP tools.

Read behavior: external Workday REST API read across HCM, Financials, Student and Platform services
(HR/PII-adjacent).

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s); the remaining documented reads are
  exposed as bounded direct reads rather than ETL streams.
- Other documented endpoints are not exposed by this connector where they are blocked in the
  operation ledger as deprecated=1.
- Fixture-only evidence: no live Workday credentials, provider calls, provider writes, or
  validation run were used.
