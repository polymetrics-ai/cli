# Overview

Reads Statuspage pages, components, incidents, subscribers, component groups, metrics, metrics
providers, page access groups/users, and incident templates through the Statuspage API.

Readable streams: `pages`, `components`, `incidents`, `subscribers`, `component_groups`, `metrics`,
`metrics_providers`, `page_access_groups`, `page_access_users`, `incident_templates`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.statuspage.io/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Statuspage API key, sent as the Authorization header with an
  'OAuth ' prefix (Authorization: OAuth <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.statuspage.io/v1`; format `uri`; Statuspage
  API base URL override for tests or proxies.
- `page_id` (optional, string); Statuspage page ID that the 'components', 'incidents', and
  'subscribers' streams are scoped to (required for those streams; substituted into the page-scoped
  path). Not required for the 'pages' stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.statuspage.io/v1`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `OAuth` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/pages` with query `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `pages`: GET `/pages` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100.
- `components`: GET `/pages/{{ config.page_id }}/components` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `incidents`: GET `/pages/{{ config.page_id }}/incidents` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100;
  incremental cursor `created_at`; formatted as `rfc3339`.
- `subscribers`: GET `/pages/{{ config.page_id }}/subscribers` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `component_groups`: GET `/pages/{{ config.page_id }}/component-groups` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100.
- `metrics`: GET `/pages/{{ config.page_id }}/metrics` - records path `.`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `metrics_providers`: GET `/pages/{{ config.page_id }}/metrics_providers` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100.
- `page_access_groups`: GET `/pages/{{ config.page_id }}/page_access_groups` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100.
- `page_access_users`: GET `/pages/{{ config.page_id }}/page_access_users` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100.
- `incident_templates`: GET `/pages/{{ config.page_id }}/incident_templates` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100.

## Write actions & risks

This connector is read-only. Read behavior: external Statuspage API read of page, component,
incident, subscriber, component group, metric, metrics provider, page access group/user, and
incident template data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3, destructive_admin=7, duplicate_of=19, non_data_endpoint=1, out_of_scope=67,
  requires_elevated_scope=5.
