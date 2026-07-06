# Overview

Reads Delighted survey responses, people, bounces, unsubscribes, and aggregate metrics through the
Delighted REST API; can create/update and delete people.

Readable streams: `survey_responses`, `people`, `bounces`, `unsubscribes`, `metrics`.

Write actions: `create_person`, `delete_person`.

Service API documentation: https://delighted.com/docs/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Delighted API key, sent as the HTTP Basic auth username with
  a blank password. Never logged.
- `base_url` (optional, string); default `https://api.delighted.com/v1`; format `uri`; Delighted API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; RFC3339 timestamp or Unix-seconds lower bound
  for survey_responses/metrics; only records updated at or after this time are read.
  stripe/chargebee/aircall) and accepts RFC3339 or Unix-seconds only (documented deviation/scope
  narrowing, see docs.md). The wire query parameter Delighted itself receives is still named
  'since', unchanged.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.delighted.com/v1`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/metrics.json`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `metrics`; page_number: `survey_responses`, `people`, `bounces`,
`unsubscribes`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `survey_responses`: GET `/survey_responses.json` - records at response root; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100;
  incremental cursor `updated_at`; sent as `since`; formatted as Unix-seconds timestamp; initial
  lower bound from `start_date`.
- `people`: GET `/people.json` - records at response root; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `bounces`: GET `/bounces.json` - records at response root; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `unsubscribes`: GET `/unsubscribes.json` - records at response root; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `metrics`: GET `/metrics.json` - records at response root; incremental sent as `since`; formatted
  as Unix-seconds timestamp; initial lower bound from `start_date`.

## Write actions & risks

Overall write risk: creates/updates Delighted people and deletes existing people.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_person`: POST `/people.json` - kind `create`; body type `form`; required record fields
  `email`; accepted fields `delay`, `email`, `name`, `phone_number`, `properties`, `send`; risk:
  creates or updates a Delighted person and may trigger survey workflow depending on account
  settings.
- `delete_person`: DELETE `/people/{{ record.person_id }}.json` - kind `delete`; body type `none`;
  path fields `person_id`; required record fields `person_id`; accepted fields `person_id`; missing
  records treated as success for status `404`; risk: deletes a Delighted person record.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 2 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1.
