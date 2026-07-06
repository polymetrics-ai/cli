# Overview

Datadog reads 15 stream(s), and writes through 27 action(s).

Readable streams: `monitors`, `dashboards`, `users`, `slo`, `downtimes`, `dashboard_lists`,
`notebooks`, `organizations`, `hosts`, `slo_corrections`, `s_tests`, `s_locations`, `s_variables`,
`api_keys`, `application_keys`.

Write actions: `create_monitor`, `update_monitor`, `delete_monitor`, `create_dashboard`,
`update_dashboard`, `delete_dashboard`, `create_dashboard_list`, `update_dashboard_list`,
`delete_dashboard_list`, `create_downtime`, `update_downtime`, `cancel_downtime`, `create_notebook`,
`update_notebook`, `delete_notebook`, `create_slo`, `update_slo`, `delete_slo`, `create_user`,
`update_user`, `disable_user`, `create_event`, `create_s_api_test`, `update_s_api_test`,
`create_api_key`, `update_api_key`, `delete_api_key`.

Service API documentation: https://docs.datadoghq.com/api/latest/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Datadog API key, sent as the DD-API-KEY header. Never
  logged.
- `application_key` (required, secret, string); Datadog application key, sent as the
  DD-APPLICATION-KEY header. Never logged.
- `base_url` (optional, string); default `https://api.datadoghq.com`; format `uri`; Datadog API base
  URL override for tests, proxies, or a regional site (e.g. https://api.datadoghq.eu,
  https://api.us3.datadoghq.com).

Secret fields are redacted in logs and write previews: `api_key`, `application_key`.

Default configuration values: `base_url=https://api.datadoghq.com`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v1/dashboard`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `dashboards`, `downtimes`, `dashboard_lists`, `organizations`,
`s_locations`, `s_variables`, `api_keys`, `application_keys`; offset_limit: `notebooks`, `hosts`,
`slo_corrections`; page_number: `monitors`, `users`, `slo`, `s_tests`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `monitors`: GET `/api/v1/monitor` - records path `.`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 0; page size 100; incremental cursor `modified`;
  formatted as `rfc3339`.
- `dashboards`: GET `/api/v1/dashboard` - records path `dashboards`.
- `users`: GET `/api/v2/users` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 0; page size 100; computed output fields
  `created_at`, `disabled`, `email`, `handle`, `name`, `status`, `verified`.
- `slo`: GET `/api/v1/slo` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 0; page size 100.
- `downtimes`: GET `/api/v1/downtime` - records path `.`.
- `dashboard_lists`: GET `/api/v1/dashboard/lists/manual` - records path `dashboard_lists`.
- `notebooks`: GET `/api/v1/notebooks` - records path `data`; offset/limit pagination; offset
  parameter `start`; limit parameter `count`; page size 100; computed output fields `author_handle`,
  `created`, `modified`, `name`.
- `organizations`: GET `/api/v1/org` - records path `orgs`.
- `hosts`: GET `/api/v1/hosts` - records path `host_list`; offset/limit pagination; offset parameter
  `start`; limit parameter `count`; page size 100.
- `slo_corrections`: GET `/api/v1/slo/correction` - records path `data`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; computed output fields
  `category`, `created_at`, `description`, `duration`, `end`, `modified_at`, `slo_id`, `start`,
  `timezone`.
- `stream`: GET connector-managed request path - records path `tests`; page-number pagination; page
  parameter `page_number`; size parameter `page_size`; starts at 0; page size 100.
- `stream`: GET connector-managed request path - records path `locations`.
- `stream`: GET connector-managed request path - records path `variables`.
- `api_keys`: GET `/api/v1/api_key` - records path `api_keys`.
- `application_keys`: GET `/api/v1/application_key` - records path `application_keys`.

## Write actions & risks

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_monitor`: POST `/api/v1/monitor` - kind `create`; body type `json`; required record fields
  `name`, `type`, `query`, `message`; accepted fields `message`, `name`, `options`, `priority`,
  `query`, `tags`, `type`; risk: creates a new alerting monitor; low-risk external mutation, no
  approval required.
- `update_monitor`: PUT `/api/v1/monitor/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `id`, `message`, `name`, `options`,
  `priority`, `query`, `tags`; risk: mutates an existing monitor's alert condition/notification
  message; a changed query/threshold affects live alerting behavior, approval required.
- `delete_monitor`: DELETE `/api/v1/monitor/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversibly removes a monitor and its alerting history reference;
  approval required.
- `create_dashboard`: POST `/api/v1/dashboard` - kind `create`; body type `json`; required record
  fields `title`, `layout_type`, `widgets`; accepted fields `description`, `layout_type`,
  `notify_list`, `tags`, `title`, `widgets`; risk: creates a new dashboard; low-risk external
  mutation, no approval required.
- `update_dashboard`: PUT `/api/v1/dashboard/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `title`, `layout_type`, `widgets`; accepted fields
  `description`, `id`, `layout_type`, `notify_list`, `tags`, `title`, `widgets`; risk: replaces an
  existing dashboard's full widget layout; external mutation, approval required.
- `delete_dashboard`: DELETE `/api/v1/dashboard/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: irreversibly removes a dashboard; approval required.
- `create_dashboard_list`: POST `/api/v1/dashboard/lists/manual` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `name`; risk: creates a new dashboard list
  (folder); low-risk external mutation, no approval required.
- `update_dashboard_list`: PUT `/api/v1/dashboard/lists/manual/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`, `name`; accepted fields `id`,
  `name`; risk: renames an existing dashboard list; external mutation, approval required.
- `delete_dashboard_list`: DELETE `/api/v1/dashboard/lists/manual/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; risk: irreversibly removes a dashboard list (folder);
  the dashboards themselves are unaffected, approval required.
- `create_downtime`: POST `/api/v1/downtime` - kind `create`; body type `json`; required record
  fields `scope`; accepted fields `end`, `message`, `monitor_id`, `monitor_tags`, `recurrence`,
  `scope`, `start`, `timezone`; risk: schedules a downtime that silences monitor alerts for the
  given scope; suppresses real alerting during the window, approval required.
- `update_downtime`: PUT `/api/v1/downtime/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `end`, `id`, `message`, `scope`,
  `start`; risk: mutates an existing downtime's window/scope; changes which alerts are currently
  suppressed, approval required.
- `cancel_downtime`: DELETE `/api/v1/downtime/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: cancels a scheduled/active downtime; alerting resumes immediately
  for its scope, approval required.
- `create_notebook`: POST `/api/v1/notebooks` - kind `create`; body type `json`; required record
  fields `name`, `cells`, `time`; accepted fields `cells`, `name`, `status`, `time`; risk: creates a
  new notebook; low-risk external mutation, no approval required.
- `update_notebook`: PUT `/api/v1/notebooks/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `name`, `cells`, `time`; accepted fields `cells`, `id`,
  `name`, `status`, `time`; risk: replaces an existing notebook's content; external mutation,
  approval required.
- `delete_notebook`: DELETE `/api/v1/notebooks/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: irreversibly removes a notebook; approval required.
- `create_slo`: POST `/api/v1/slo` - kind `create`; body type `json`; required record fields `name`,
  `type`, `thresholds`; accepted fields `description`, `monitor_ids`, `name`, `query`, `tags`,
  `thresholds`, `type`; risk: creates a new SLO target; low-risk external mutation, no approval
  required.
- `update_slo`: PUT `/api/v1/slo/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `description`, `id`, `name`, `tags`,
  `thresholds`; risk: mutates an existing SLO's target thresholds; affects SLO burn-rate alerting,
  approval required.
- `delete_slo`: DELETE `/api/v1/slo/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: irreversibly removes an SLO and its historical error-budget tracking; approval
  required.
- `create_user`: POST `/api/v1/user` - kind `create`; body type `json`; required record fields
  `email`; accepted fields `access_role`, `email`, `name`; risk: invites a new user into the Datadog
  organization with the given role; approval required.
- `update_user`: PUT `/api/v1/user/{{ record.handle }}` - kind `update`; body type `json`; path
  fields `handle`; required record fields `handle`; accepted fields `access_role`, `disabled`,
  `email`, `handle`, `name`; risk: mutates an existing user's role/profile; a changed access_role
  directly changes that user's permissions, approval required.
- `disable_user`: DELETE `/api/v1/user/{{ record.handle }}` - kind `delete`; body type `none`; path
  fields `handle`; required record fields `handle`; accepted fields `handle`; missing records
  treated as success for status `404`; risk: disables a user's access to the Datadog organization;
  approval required.
- `create_event`: POST `/api/v1/events` - kind `create`; body type `json`; required record fields
  `title`, `text`; accepted fields `aggregation_key`, `alert_type`, `date_happened`, `host`,
  `priority`, `tags`, `text`, `title`; risk: posts a custom event into the Datadog event stream;
  low-risk external mutation, no approval required.
- `create_s_api_test`: POST connector-managed endpoint - kind `create`; body type `json`; required
  record fields `name`, `type`, `config`, `locations`; accepted fields `config`, `locations`,
  `message`, `name`, `options`, `subtype`, `tags`, `type`.
- `update_s_api_test`: PUT connector-managed endpoint - kind `update`; body type `json`; path fields
  `public_id`; required record fields `public_id`; accepted fields `config`, `locations`, `message`,
  `name`, `options`, `public_id`, `tags`.
- `create_api_key`: POST `/api/v1/api_key` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `name`; risk: creates a new organization API key with full
  agent-submission scope; a newly-minted long-lived credential, approval required.
- `update_api_key`: PUT `/api/v1/api_key/{{ record.key }}` - kind `update`; body type `json`; path
  fields `key`; required record fields `key`, `name`; accepted fields `key`, `name`; risk: renames
  an existing API key; low-risk external mutation, no approval required.
- `delete_api_key`: DELETE `/api/v1/api_key/{{ record.key }}` - kind `delete`; body type `none`;
  path fields `key`; required record fields `key`; accepted fields `key`; missing records treated as
  success for status `404`; risk: irreversibly revokes an organization API key; every
  agent/integration still using it immediately loses ingest access, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 12 stream-backed endpoint group(s), 25 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=8, duplicate_of=27, non_data_endpoint=8, out_of_scope=97,
  requires_elevated_scope=52.
