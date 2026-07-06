# Overview

Reads Ruddr clients, projects, and time entries through the Ruddr API. Read-only.

Readable streams: `clients`, `projects`, `time_entries`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.ruddr.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Ruddr API token, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.ruddr.io`; format `uri`; Ruddr API base URL
  override for tests or proxies.
- `mode` (optional, string).
- `workspace_id` (required, string); Ruddr workspace id, substituted into every stream's
  workspace-scoped path.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.ruddr.io`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/workspaces/{{ config.workspace_id }}/projects`.

## Streams notes

Default pagination: single request; no pagination.

- `clients`: GET `/api/workspaces/{{ config.workspace_id }}/clients` - records path `results`; query
  `page`=`1`; `page_size`=`2`; follows a next-page URL from the response body; URL path `next`; next
  URLs stay on the configured API host; computed output fields `stream`; emits passthrough records.
- `projects`: GET `/api/workspaces/{{ config.workspace_id }}/projects` - records path `results`;
  query `page`=`1`; `page_size`=`2`; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host; computed output fields `stream`; emits
  passthrough records.
- `time_entries`: GET `/api/workspaces/{{ config.workspace_id }}/time_entries` - records path
  `results`; query `page`=`1`; `page_size`=`2`; follows a next-page URL from the response body; URL
  path `next`; next URLs stay on the configured API host; computed output fields `stream`; emits
  passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Ruddr API read of client, project, and
time-entry data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
