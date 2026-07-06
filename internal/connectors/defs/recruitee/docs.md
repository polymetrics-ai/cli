# Overview

Reads Recruitee offers, candidates, departments, sources, and tags through the Recruitee REST API.

Readable streams: `offers`, `candidates`, `departments`, `sources`, `tags`.

This connector is read-only; no write actions are declared.

Service API documentation: https://apidocs.recruitee.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Recruitee personal API token, sent as a Bearer token
  (Authorization: Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.recruitee.com`; format `uri`; Recruitee API
  base URL override for tests or proxies.
- `company_id` (required, string); Recruitee company (tenant) ID; every request path is scoped under
  /c/{company_id}/ so paths stay allow-listed rather than accepting arbitrary request paths.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.recruitee.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/c/{{ config.company_id }}/offers` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

- `offers`: GET `/c/{{ config.company_id }}/offers` - records path `offers`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `created_at`, `id`, `status`, `title`, `updated_at`.
- `candidates`: GET `/c/{{ config.company_id }}/candidates` - records path `candidates`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed
  output fields `created_at`, `email`, `id`, `name`, `updated_at`.
- `departments`: GET `/c/{{ config.company_id }}/departments` - records path `departments`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  computed output fields `id`, `name`.
- `sources`: GET `/c/{{ config.company_id }}/sources` - records path `sources`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed
  output fields `id`, `name`.
- `tags`: GET `/c/{{ config.company_id }}/tags` - records path `tags`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields `id`,
  `name`.

## Write actions & risks

This connector is read-only. Read behavior: external Recruitee API read of ATS offer and candidate
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=48, deprecated=1, destructive_admin=109, non_data_endpoint=53, out_of_scope=648,
  requires_elevated_scope=84.
