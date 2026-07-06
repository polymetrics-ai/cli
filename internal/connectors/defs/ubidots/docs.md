# Overview

Reads Ubidots devices, variables, variable values, device groups, device types, dashboards, and
events, and writes device/variable lifecycle mutations and new variable data points through API
v2.0.

Readable streams: `devices`, `variables`, `dashboards`, `events`, `device_groups`, `device_types`,
`variable_values`.

Write actions: `create_device`, `update_device`, `delete_device`, `create_variable`,
`update_variable`, `delete_variable`, `create_variable_value`.

Service API documentation: https://docs.ubidots.com/reference/welcome.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://industrial.api.ubidots.com`; format `uri`; Ubidots
  API base URL override for tests or proxies.
- `token` (required, secret, string); Ubidots API token, sent as the X-Auth-Token header on every
  request.

Secret fields are redacted in logs and write previews: `token`.

Default configuration values: `base_url=https://industrial.api.ubidots.com`.

Authentication behavior:

- API key authentication in `X-Auth-Token` using `secrets.token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `api/v2.0/devices/`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 100; maximum 1 page(s).

- `devices`: GET `api/v2.0/devices/` - records path `results`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s);
  computed output fields `created_at`.
- `variables`: GET `api/v2.0/variables/` - records path `results`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s);
  computed output fields `created_at`.
- `dashboards`: GET `api/v2.0/dashboards/` - records path `results`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s);
  computed output fields `created_at`.
- `events`: GET `api/v2.0/events/` - records path `results`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s); computed output
  fields `created_at`.
- `device_groups`: GET `api/v2.0/device_groups/` - records path `results`; page-number pagination;
  page parameter `page`; size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s).
- `device_types`: GET `api/v2.0/device_types/` - records path `results`; page-number pagination;
  page parameter `page`; size parameter `page_size`; starts at 1; page size 100; maximum 1 page(s).
- `variable_values`: GET `api/v1.6/variables/{{ fanout.id }}/values/` - records path `results`;
  page-number pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size
  100; maximum 1 page(s); fan-out; ids from request `api/v2.0/variables/`; id-list records path
  `results`; id field `id`; id inserted into the request path; stamps `variable_id`.

## Write actions & risks

Overall write risk: external mutation of Ubidots devices and variables (create/update/delete) and
injection of new variable data points; device/variable delete is destructive and irreversible.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_device`: POST `api/v2.0/devices/` - kind `create`; body type `json`; required record
  fields `label`; accepted fields `description`, `label`, `name`, `organization`, `properties`,
  `tags`; risk: creates a new Ubidots device; low-risk external mutation, no approval required.
- `update_device`: PATCH `api/v2.0/devices/{{ record.id }}/` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `description`, `id`, `label`, `name`,
  `properties`, `tags`; risk: updates the fields of an existing Ubidots device; external mutation,
  no approval required.
- `delete_device`: DELETE `api/v2.0/devices/{{ record.id }}/` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: permanently deletes a device and all of its variables/values;
  destructive and irreversible; approval required.
- `create_variable`: POST `api/v2.0/variables/` - kind `create`; body type `json`; required record
  fields `label`, `device`; accepted fields `description`, `device`, `label`, `name`, `tags`,
  `unit`; risk: creates a new variable under an existing device; low-risk external mutation, no
  approval required.
- `update_variable`: PATCH `api/v2.0/variables/{{ record.id }}/` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `description`, `id`, `label`,
  `name`, `tags`, `unit`; risk: updates the fields of an existing variable; external mutation, no
  approval required.
- `delete_variable`: DELETE `api/v2.0/variables/{{ record.id }}/` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: permanently deletes a variable and all of its stored values;
  destructive and irreversible; approval required.
- `create_variable_value`: POST `api/v1.6/variables/{{ record.variable_id }}/values/` - kind
  `create`; body type `json`; path fields `variable_id`; body fields `value`, `timestamp`,
  `context`; required record fields `variable_id`, `value`; accepted fields `context`, `timestamp`,
  `value`, `variable_id`; risk: injects a new data point (dot) into an existing variable; low-risk
  external mutation, no approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 7 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=7, duplicate_of=12, non_data_endpoint=3, out_of_scope=11,
  requires_elevated_scope=2.
