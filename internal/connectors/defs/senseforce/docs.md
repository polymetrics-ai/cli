# Overview

Reads records from a configured Senseforce dataset through the Senseforce API.

Readable streams: `records`.

This connector is read-only; no write actions are declared.

Service API documentation: https://manual.senseforce.io/manual/sf-platform/public-api/endpoints.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Senseforce API access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `backend_url` (required, string); format `uri`; Your Senseforce backend base URL (e.g.
  https://yourtenant.senseforce.io).
- `dataset_id` (required, string); The Senseforce dataset ID to read records from; substituted into
  /api/v1/datasets/<dataset_id>/records.

Secret fields are redacted in logs and write previews: `access_token`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `backend_url` value after applying defaults.

Connection checks call GET `/api/v1/datasets/{{ config.dataset_id }}/records`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100; maximum 1 page(s).

- `records`: GET `/api/v1/datasets/{{ config.dataset_id }}/records` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100; maximum 1 page(s); emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Senseforce API read of a configured dataset's
rows.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 1 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=1.
