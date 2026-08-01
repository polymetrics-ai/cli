# Overview

Workday REST is generated from the official Workday REST Directory 2026.30 `productionConfidenceLevel` service manifest and its 52 OpenAPI v2 service files. Archived services in the official directory are not counted as current production surface.

Post-change operation dispositions:

| disposition | count |
| --- | ---: |
| fixture-backed ETL streams | 463 |
| bounded direct reads | 174 |
| typed reverse ETL writes | 251 |
| blocked binary/file or contract-gap operations | 32 |
| total official production operations | 920 |

The connector does not claim Workday certification and this implementation makes no live provider calls.

## Auth setup

- `access_token` is required and marked `x-secret`; provide it through credentials, environment, or stdin-backed secret flows, never prompt text or fixtures.
- `base_url` defaults to `https://wd2-impl-services1.workday.com` and should be set to the tenant-specific Workday REST hostname/base URL for live use.
- `tenant` is retained as an optional legacy compatibility field but is not interpolated into the official 2026.30 service paths.
- Optional path-parameter config fields such as `id`, `subresource_id`, and other generated Workday ID fields are needed only when reading streams whose official paths contain those parameters.

## Streams notes

The bundle declares 463 fixture-backed streams. Stream paths use official service `basePath + path` values from the 2026.30 OpenAPI files. Collection responses read `data`; single-resource responses emit the response object. Pagination is bounded to one page (`limit=100`, `offset=0`) unless a stream is a single-object read, in which case pagination is disabled. This keeps fixture replay and accidental live reads bounded.

## Write actions & risks

The bundle declares 251 typed non-binary write actions. Each action has a closed root `record_schema`, explicit path fields, risk text, redaction for path identifiers, and uses the existing reverse ETL plan -> preview -> explicit approval -> execute flow. DELETE actions are marked destructive and include documented 404 idempotent-missing handling.

Multipart Workday business-process writes that carry a `jsonData` root part are represented as bounded multipart field writes. Attachment/file upload and binary/download operations are not implemented as writes.

## Known limits

- 32 official operations remain blocked/planned because they are binary/file/attachment operations, destructive contract gaps, or the generic POST `/wql/v1/data` WQL query passthrough that cannot be represented as a named typed action in the current connector contract.
- Direct reads are limited to Workday values/search endpoints and cap JSON responses at 1 MiB with clinical JSON redaction. No raw method/path/body/query passthrough is exposed.
- Fixture-backed parity is not live certification; live-safe certification requires separate approved credentials, redacted artifacts, and provider-safe execution.
- Official source evidence: `services2026.30.json` plus each service `specFilePath` from `https://community.workday.com/sites/default/files/file-hosting/restapi/`.
