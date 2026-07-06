# Overview

Reads GlassFrog circles, roles, people, projects, and assignments through the GlassFrog API v3
(read-only full-refresh source).

Readable streams: `assignments`, `circles`, `people`, `projects`, `roles`.

This connector is read-only; no write actions are declared.

Service API documentation: https://documenter.getpostman.com/view/1014385/glassfrog-api-v3/2SJViY.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); GlassFrog API token, sent as the X-Auth-Token header. Never
  logged.
- `base_url` (optional, string); default `https://api.glassfrog.com/api/v3`; format `uri`; GlassFrog
  API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.glassfrog.com/api/v3`.

Authentication behavior:

- API key authentication in `X-Auth-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/circles`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

- `assignments`: GET `/assignments` - records path `assignments`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `person_id`, `role_id`.
- `circles`: GET `/circles` - records path `circles`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; computed output fields `supported_role_id`.
- `people`: GET `/people` - records path `people`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `projects`: GET `/projects` - records path `projects`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `roles`: GET `/roles` - records path `roles`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external GlassFrog API read of circle, role, person,
project, and assignment data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=3.
