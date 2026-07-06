# Overview

Reads k6 Cloud organizations, projects, and load tests through the k6 Cloud REST API.

Readable streams: `organizations`, `k6_tests`, `projects`.

This connector is read-only; no write actions are declared.

Service API documentation: https://k6.io/docs/cloud/cloud-reference/cloud-rest-api/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); k6 Cloud API token. Used only for Bearer auth
  (Authorization: Bearer <api_token>); never logged.
- `base_url` (optional, string); default `https://api.k6.io`; format `uri`; k6 Cloud API base URL
  override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, string); default `32`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.k6.io`, `page_size=32`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v3/organizations`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `organizations`; page_number: `k6_tests`, `projects`.

- `organizations`: GET `/v3/organizations` - records path `organizations`.
- `k6_tests`: GET `loadtests/v2/tests` - records path `k6-tests`; query `page_size` from template
  `{{ config.page_size }}`, default `32`; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 32.
- `projects`: GET `/v3/organizations/{{ fanout.id }}/projects` - records path `projects`; query
  `page_size` from template `{{ config.page_size }}`, default `32`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 32; fan-out; ids from request
  `/v3/organizations`; id-list records path `organizations`; id field `id`; id inserted into the
  request path.

## Write actions & risks

This connector is read-only. Read behavior: external k6 Cloud API read of organizations, projects,
and load tests.

## Known limits

- Batch defaults: read_page_size=32.
- API coverage includes 3 stream-backed endpoint group(s).
