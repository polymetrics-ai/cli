# Overview

Reads Qualaroo nudges and reporting response records through the Qualaroo API. Read-only.

Readable streams: `nudges`, `responses`, `survey_responses`.

This connector is read-only; no write actions are declared.

Service API documentation: https://help.qualaroo.com/the-rest-reporting-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Qualaroo API key. Never logged.
- `api_secret` (optional, secret, string); Optional Qualaroo Reporting API secret. When provided,
  the bundle uses HTTP Basic auth for reporting response reads.
- `base_url` (optional, string); default `https://api.qualaroo.com/api/v1`; format `uri`; Qualaroo
  API base URL override for tests or proxies.
- `survey_id` (optional, string); Qualaroo survey/nudge id used by the documented Reporting API
  response endpoint.

Secret fields are redacted in logs and write previews: `api_key`, `api_secret`.

Default configuration values: `base_url=https://api.qualaroo.com/api/v1`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`, `secrets.api_secret` when `{{
  secrets.api_secret }}`.
- API key authentication in `Authorization` with prefix `Token token="` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/nudges` with query `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: offset_limit: `survey_responses`; page_number: `nudges`, `responses`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `nudges`: GET `/nudges` - records path `nudges`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`; formatted
  as `rfc3339`.
- `responses`: GET `/responses` - records path `responses`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor `created_at`;
  formatted as `rfc3339`.
- `survey_responses`: GET `/nudges/{{ config.survey_id }}/responses.json` - records at response
  root; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 500;
  incremental cursor `time`; formatted as `rfc3339`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Qualaroo API read of survey nudge and reporting
response data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
