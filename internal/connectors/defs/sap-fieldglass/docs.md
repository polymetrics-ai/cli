# Overview

Reads SAP Fieldglass workers, job postings, and time sheets through the SAP Fieldglass API.
Read-only.

Readable streams: `workers`, `job_postings`, `time_sheets`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.sap.com/package/SAPFieldglass/rest.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); SAP Fieldglass OAuth access token, sent as a Bearer
  token (Authorization: Bearer <access_token>). Never logged.
- `base_url` (optional, string); default `https://api.fieldglass.net`; format `uri`; SAP Fieldglass
  API base URL override for tests or proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.fieldglass.net`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v1/workers`.

## Streams notes

Default pagination: single request; no pagination.

- `workers`: GET `/api/v1/workers` - records path `data`; query `limit`=`2`; `page`=`1`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host;
  computed output fields `id`, `stream`; emits passthrough records.
- `job_postings`: GET `/api/v1/job_postings` - records path `data`; query `limit`=`2`; `page`=`1`;
  follows a next-page URL from the response body; URL path `next`; next URLs stay on the configured
  API host; computed output fields `id`, `stream`; emits passthrough records.
- `time_sheets`: GET `/api/v1/time_sheets` - records path `data`; query `limit`=`2`; `page`=`1`;
  follows a next-page URL from the response body; URL path `next`; next URLs stay on the configured
  API host; computed output fields `id`, `stream`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external SAP Fieldglass API read of worker, job posting,
and time sheet data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
