# Overview

Reads Pipeliner CRM accounts, contacts, opportunities, and leads through the REST API.

Readable streams: `accounts`, `contacts`, `opportunities`, `leads`.

This connector is read-only; no write actions are declared.

Service API documentation: https://workspace.pipelinersales.com/community/api/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.pipelinersales.com/api/v100/rest`; format
  `uri`; Pipeliner API base URL override for tests or proxies.
- `mode` (optional, string).
- `password` (required, secret, string); Pipeliner API password, sent as the HTTP Basic auth
  password. Never logged.
- `space_id` (required, string); Pipeliner space (account) id; sent as the
  /spaces/{space_id}/entities/... path segment.
- `username` (required, secret, string); Pipeliner API username, sent as the HTTP Basic auth
  username. Never logged.

Secret fields are redacted in logs and write previews: `password`, `username`.

Default configuration values: `base_url=https://api.pipelinersales.com/api/v100/rest`.

Authentication behavior:

- HTTP Basic authentication using `secrets.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/spaces/{{ config.space_id }}/entities/Accounts` with query `limit`=`1`;
`offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

- `accounts`: GET `/spaces/{{ config.space_id }}/entities/Accounts` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  computed output fields `updated_at`.
- `contacts`: GET `/spaces/{{ config.space_id }}/entities/Contacts` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  computed output fields `updated_at`.
- `opportunities`: GET `/spaces/{{ config.space_id }}/entities/Opportunities` - records path `data`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  computed output fields `updated_at`.
- `leads`: GET `/spaces/{{ config.space_id }}/entities/Leads` - records path `data`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; computed output
  fields `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Pipeliner CRM API read of account, contact,
opportunity, and lead data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=5.
