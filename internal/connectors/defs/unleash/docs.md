# Overview

Reads Unleash projects, feature toggles, environments, and segments through admin API list
endpoints.

Readable streams: `projects`, `features`, `environments`, `segments`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.getunleash.io/reference/api/unleash.

## Auth setup

Connection fields:

- `api_token` (optional, secret, string); Unleash admin API token. Used only for Bearer auth; never
  logged.
- `base_url` (optional, string); default `https://app.unleash-hosted.com`; format `uri`; Unleash
  instance base URL override for self-hosted instances, tests, or proxies.
- `mode` (optional, string).
- `project_id` (optional, string); default `default`; Unleash project id the (project-scoped)
  features stream reads from.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://app.unleash-hosted.com`, `project_id=default`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/admin/projects/{{ config.project_id }}/features` with query
`limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100; maximum 1 page(s).

- `projects`: GET `/api/admin/projects` - records path `projects`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 1 page(s).
- `features`: GET `/api/admin/projects/{{ config.project_id }}/features` - records path `features`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  maximum 1 page(s).
- `environments`: GET `/api/admin/environments` - records path `environments`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; maximum 1 page(s);
  computed output fields `id`.
- `segments`: GET `/api/admin/segments` - records path `segments`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 1 page(s).

## Write actions & risks

This connector is read-only. Read behavior: external Unleash admin API read of project, feature
toggle, environment, and segment data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, out_of_scope=3, requires_elevated_scope=1.
