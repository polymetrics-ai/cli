# Overview

Reads Rocketlane projects, tasks, customers, users, and time entries through the REST API.

Readable streams: `projects`, `tasks`, `customers`, `users`, `time_entries`.

This connector is read-only; no write actions are declared.

Service API documentation: https://apidocs.rocketlane.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Rocketlane API key, sent as the api-key header. Never
  logged.
- `base_url` (optional, string); default `https://api.rocketlane.com/api/1.0`; format `uri`;
  Rocketlane API base URL override for tests or proxies.
- `created_after` (optional, string); Optional RFC3339 lower-bound filter passed through as the
  'created_after' parameter.
- `mode` (optional, string).
- `project_id` (optional, string); Optional project id filter passed through as the 'projectId'
  parameter.
- `status` (optional, string); Optional status filter passed through as the 'status' parameter.
- `updated_after` (optional, string); Optional RFC3339 lower-bound filter passed through as the
  'updated_after' parameter.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.rocketlane.com/api/1.0`.

Authentication behavior:

- API key authentication in `api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1; page size 100.

- `projects`: GET `/projects` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `status` from template `{{ config.status }}`,
  omitted when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1;
  page size 100; computed output fields `stream`; emits passthrough records.
- `tasks`: GET `/tasks` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `projectId` from template `{{ config.project_id
  }}`, omitted when absent; `status` from template `{{ config.status }}`, omitted when absent;
  `updated_after` from template `{{ config.updated_after }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `pageSize`; starts at 1; page size 100; computed
  output fields `stream`; emits passthrough records.
- `customers`: GET `/customers` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `updated_after` from template `{{
  config.updated_after }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1; page size 100; computed output fields `stream`; emits
  passthrough records.
- `users`: GET `/users` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `status` from template `{{ config.status }}`,
  omitted when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1;
  page size 100; computed output fields `stream`; emits passthrough records.
- `time_entries`: GET `/time-entries` - records path `data`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `projectId` from template `{{ config.project_id
  }}`, omitted when absent; `updated_after` from template `{{ config.updated_after }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1;
  page size 100; computed output fields `stream`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Rocketlane API read of project, task, customer,
and time-entry data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
