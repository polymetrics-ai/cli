# Overview

Reads Pivotal Tracker projects, stories, iterations, and epics through API v5.

Readable streams: `projects`, `stories`, `iterations`, `epics`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.pivotaltracker.com/help/api/rest/v5.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Pivotal Tracker API token. Sent as the X-TrackerToken
  header on every request; never logged.
- `base_url` (optional, string); default `https://www.pivotaltracker.com/services/v5`; format `uri`;
  Pivotal Tracker API base URL override for tests or proxies.
- `project_id` (optional, string); Pivotal Tracker project id. Required by the stories, iterations,
  and epics streams, which read from projects/{project_id}/{resource}; not used by the projects
  stream itself.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://www.pivotaltracker.com/services/v5`.

Authentication behavior:

- API key authentication in `X-TrackerToken` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `projects`: GET `/projects` - records path `.`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; computed output fields `id`, `name`, `state`,
  `updated_at`.
- `stories`: GET `/projects/{{ config.project_id }}/stories` - records path `.`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; computed output
  fields `id`, `name`, `state`, `updated_at`.
- `iterations`: GET `/projects/{{ config.project_id }}/iterations` - records path `.`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; computed output
  fields `id`, `name`, `state`, `updated_at`.
- `epics`: GET `/projects/{{ config.project_id }}/epics` - records path `.`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; computed output
  fields `id`, `name`, `state`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Pivotal Tracker API read of project, story,
iteration, and epic data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=3.
