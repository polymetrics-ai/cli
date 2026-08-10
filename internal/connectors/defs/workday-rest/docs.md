# Overview

Reads and writes the documented Workday REST surface across the 52 independently versioned services
Workday's own directory publishes, with bearer-token authentication.

`api_surface.json` owns Workday's source-backed endpoint ledger, while `cli_surface.json` owns
per-command availability. The bundle includes stream-backed reads, bounded direct and binary reads,
typed reverse-ETL actions, and an explicitly blocked deprecated operation; use runtime help or
`cli_surface.json` to determine which command paths are currently executable. Certification is
fixture-only; no live provider calls were made.

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

The connector declares typed write actions across the HCM, Financials, Student, and Platform
services, including HR/PII-adjacent worker, absence, payroll, and financial records. `writes.json`
owns those declarations; `cli_surface.json` records whether each action has an executable command
path.

Writes are only available through reverse ETL plan -> preview -> explicit approval -> execute. Every
DELETE action is gated as destructive and additionally requires a typed confirmation. The bundle
does not expose arbitrary request bodies, raw query strings, generic method/path/body, file bytes,
shell commands, or passthrough HTTP tools.

Read behavior: external Workday REST API read across HCM, Financials, Student and Platform services
(HR/PII-adjacent).

## Known limits

- Batch defaults: read_page_size=100.
- `api_surface.json` is the authoritative per-endpoint coverage and blocked-operation ledger.
- Fixture-only evidence: no live Workday credentials, provider calls, provider writes, or
  certification run were used.
