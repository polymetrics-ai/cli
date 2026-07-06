# Overview

Reads NinjaOne RMM organizations, devices, locations, activities, and policies through the NinjaOne
v2 REST API.

Readable streams: `organizations`, `devices`, `locations`, `activities`, `policies`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.ninjarmm.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); NinjaOne RMM API bearer token. Used only for Bearer auth;
  never logged.
- `base_url` (optional, string); default `https://app.ninjarmm.com`; format `uri`; NinjaOne RMM API
  base URL override for tests, proxies, or a regional instance.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-1000) for paginated streams.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.ninjarmm.com`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/organizations` with query `pageSize`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `after`; next cursor from last record field
`id`.

Pagination by stream: cursor: `organizations`, `devices`, `locations`, `activities`; none:
`policies`.

- `organizations`: GET `/v2/organizations` - records path `.`; query `pageSize` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `after`; next cursor from
  last record field `id`; computed output fields `description`, `id`, `name`, `node_approval_mode`.
- `devices`: GET `/v2/devices` - records path `.`; query `pageSize` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `after`; next cursor from
  last record field `id`; computed output fields `approval_status`, `dns_name`, `id`, `location_id`,
  `node_class`, `offline`, `organization_id`, `system_name`.
- `locations`: GET `/v2/locations` - records path `.`; query `pageSize` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `after`; next cursor from
  last record field `id`; computed output fields `address`, `description`, `id`, `name`,
  `organization_id`.
- `activities`: GET `/v2/activities` - records path `.`; query `pageSize` from template `{{
  config.page_size }}`, default `100`; cursor pagination; cursor parameter `after`; next cursor from
  last record field `id`; computed output fields `activityTime`, `activity_type`, `device_id`, `id`,
  `message`, `status`.
- `policies`: GET `/v2/policies` - records path `.`; computed output fields `description`, `id`,
  `name`, `node_class`.

## Write actions & risks

This connector is read-only. Read behavior: external NinjaOne RMM API read of managed device and
organization data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
