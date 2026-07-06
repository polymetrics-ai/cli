# Overview

Reads incident.io incidents, severities, incident roles, users, and follow-ups through the
incident.io REST API.

Readable streams: `incidents`, `severities`, `incident_roles`, `users`, `follow_ups`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api-docs.incident.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); incident.io API key. Used only for Bearer auth; never
  logged.
- `base_url` (optional, string); default `https://api.incident.io`; format `uri`; incident.io API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page for paginated streams (1-250).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.incident.io`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/severities`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `incidents`, `users`, `follow_ups`; none: `severities`,
`incident_roles`.

- `incidents`: GET `/v2/incidents` - records path `incidents`; query `page_size`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `after`; next token from
  `pagination_meta.after`; computed output fields `severity_id`, `severity_name`, `status_category`,
  `status_id`, `status_name`.
- `severities`: GET `/v1/severities` - records path `severities`.
- `incident_roles`: GET `/v2/incident_roles` - records path `incident_roles`.
- `users`: GET `/v2/users` - records path `users`; query `page_size`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `after`; next token from `pagination_meta.after`; computed
  output fields `base_role_id`, `base_role_name`.
- `follow_ups`: GET `/v2/follow_ups` - records path `follow_ups`; query `page_size`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `after`; next token from
  `pagination_meta.after`; computed output fields `assignee_id`, `assignee_name`.

## Write actions & risks

This connector is read-only. Read behavior: external incident.io API read of incidents, severities,
roles, users, and follow-ups.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=7.
