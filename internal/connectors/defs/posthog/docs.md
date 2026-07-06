# Overview

Reads PostHog events and persons for a project via the PostHog REST API. Read-only.

Readable streams: `events`, `persons`.

This connector is read-only; no write actions are declared.

Service API documentation: https://posthog.com/docs/api.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); PostHog personal API key. Sent as a Bearer token; never
  logged.
- `base_url` (optional, string); default `https://app.posthog.com`; format `uri`; Base PostHog URL.
  Defaults to PostHog Cloud; override for self-hosted instances or tests.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records requested per page (sent as the limit query
  param).
- `project_id` (required, string); PostHog project ID; scopes every request to
  /api/projects/{project_id}/.
- `start_date` (optional, string); RFC3339 lower bound sent as the events stream's after filter. Any
  data before this date will not be replicated.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.posthog.com`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/projects/{{ config.project_id }}/events/` with query `limit`=`1`.

## Streams notes

Default pagination: single request; no pagination.

- `events`: GET `/api/projects/{{ config.project_id }}/events/` - records path `results`; query
  `after` from template `{{ config.start_date }}`, omitted when absent; `limit` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host; emits passthrough records.
- `persons`: GET `/api/projects/{{ config.project_id }}/persons/` - records path `results`; query
  `limit` from template `{{ config.page_size }}`, default `100`; follows a next-page URL from the
  response body; URL path `next`; next URLs stay on the configured API host; emits passthrough
  records.

## Write actions & risks

This connector is read-only. Read behavior: external PostHog API read of project event and person
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 2 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
