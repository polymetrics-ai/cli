# Overview

Reads Avni subjects and encounters through a read-only HTTP API using HTTP Basic authentication.

Readable streams: `subjects`, `encounters`, `program_enrolments`, `program_encounters`,
`group_subjects`, `locations`, `approval_statuses`.

This connector is read-only; no write actions are declared.

Service API documentation: https://avniproject.org/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://app.avniproject.org`; format `uri`; Avni API base
  URL override for tests, proxies, or self-hosted instances.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page.
- `password` (required, secret, string); Avni account password, sent via HTTP Basic auth. Never
  logged.
- `start_date` (optional, string); format `date-time`; Optional RFC3339 lower bound sent as the
  start_date query parameter on every request.
- `username` (required, string); Avni account username, sent via HTTP Basic auth.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://app.avniproject.org`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/subjects` with query `page_size`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from `next_page`; maximum
100 page(s).

- `subjects`: GET `/api/subjects` - records path `items`; query `page_size`=`{{ config.page_size
  }}`; `start_date` from template `{{ config.start_date }}`, omitted when absent; cursor pagination;
  cursor parameter `page`; next token from `next_page`; maximum 100 page(s); emits passthrough
  records.
- `encounters`: GET `/api/encounters` - records path `items`; query `page_size`=`{{ config.page_size
  }}`; `start_date` from template `{{ config.start_date }}`, omitted when absent; cursor pagination;
  cursor parameter `page`; next token from `next_page`; maximum 100 page(s); emits passthrough
  records.
- `program_enrolments`: GET `/api/programEnrolments` - records path `items`; query `page_size`=`{{
  config.page_size }}`; `start_date` from template `{{ config.start_date }}`, omitted when absent;
  cursor pagination; cursor parameter `page`; next token from `next_page`; maximum 100 page(s).
- `program_encounters`: GET `/api/programEncounters` - records path `items`; query `page_size`=`{{
  config.page_size }}`; `start_date` from template `{{ config.start_date }}`, omitted when absent;
  cursor pagination; cursor parameter `page`; next token from `next_page`; maximum 100 page(s).
- `group_subjects`: GET `/api/groupSubjects` - records path `items`; query `page_size`=`{{
  config.page_size }}`; `start_date` from template `{{ config.start_date }}`, omitted when absent;
  cursor pagination; cursor parameter `page`; next token from `next_page`; maximum 100 page(s).
- `locations`: GET `/api/locations` - records path `items`; query `page_size`=`{{ config.page_size
  }}`; `start_date` from template `{{ config.start_date }}`, omitted when absent; cursor pagination;
  cursor parameter `page`; next token from `next_page`; maximum 100 page(s).
- `approval_statuses`: GET `/api/approvalStatuses` - records path `items`; query
  `lastModifiedDateTime` from template `{{ config.start_date }}`, omitted when absent;
  `page_size`=`{{ config.page_size }}`; cursor pagination; cursor parameter `page`; next token from
  `next_page`; maximum 100 page(s).

## Write actions & risks

This connector is read-only. Read behavior: external Avni API read of subjects and encounters.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=6, duplicate_of=5, non_data_endpoint=1, out_of_scope=11,
  requires_elevated_scope=1.
