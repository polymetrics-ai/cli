# Overview

Reads Close CRM leads, contacts, opportunities, activities, users, tasks, lead/opportunity statuses,
pipelines, roles, groups, and custom field definitions, and writes
leads/contacts/opportunities/tasks through the Close REST API.

Readable streams: `leads`, `contacts`, `opportunities`, `activities`, `users`, `tasks`,
`lead_statuses`, `opportunity_statuses`, `pipelines`, `roles`, `groups`, `custom_fields_lead`,
`custom_fields_contact`, `custom_fields_opportunity`.

Write actions: `create_lead`, `update_lead`, `delete_lead`, `create_contact`, `update_contact`,
`delete_contact`, `create_opportunity`, `update_opportunity`, `delete_opportunity`, `create_task`,
`update_task`, `delete_task`.

Service API documentation: https://developer.close.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Close API key, sent as the HTTP Basic auth username with an
  empty password. Used only for auth; never logged.
- `base_url` (optional, string); default `https://api.close.com/api/v1`; format `uri`; Close API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.close.com/api/v1`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/lead/`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `_skip`; limit parameter `_limit`;
page size 100.

Pagination by stream: none: `lead_statuses`, `opportunity_statuses`, `pipelines`, `roles`, `groups`;
offset_limit: `leads`, `contacts`, `opportunities`, `activities`, `users`, `tasks`,
`custom_fields_lead`, `custom_fields_contact`, `custom_fields_opportunity`.

- `leads`: GET `/lead/` - records path `data`; offset/limit pagination; offset parameter `_skip`;
  limit parameter `_limit`; page size 100.
- `contacts`: GET `/contact/` - records path `data`; offset/limit pagination; offset parameter
  `_skip`; limit parameter `_limit`; page size 100.
- `opportunities`: GET `/opportunity/` - records path `data`; offset/limit pagination; offset
  parameter `_skip`; limit parameter `_limit`; page size 100.
- `activities`: GET `/activity/` - records path `data`; offset/limit pagination; offset parameter
  `_skip`; limit parameter `_limit`; page size 100.
- `users`: GET `/user/` - records path `data`; offset/limit pagination; offset parameter `_skip`;
  limit parameter `_limit`; page size 100.
- `tasks`: GET `/task/` - records path `data`; offset/limit pagination; offset parameter `_skip`;
  limit parameter `_limit`; page size 100.
- `lead_statuses`: GET `/status/lead/` - records path `data`.
- `opportunity_statuses`: GET `/status/opportunity/` - records path `data`.
- `pipelines`: GET `/pipeline/` - records path `data`.
- `roles`: GET `/role/` - records path `data`.
- `groups`: GET `/group/` - records path `data`.
- `custom_fields_lead`: GET `/custom_field/lead/` - records path `data`; offset/limit pagination;
  offset parameter `_skip`; limit parameter `_limit`; page size 100.
- `custom_fields_contact`: GET `/custom_field/contact/` - records path `data`; offset/limit
  pagination; offset parameter `_skip`; limit parameter `_limit`; page size 100.
- `custom_fields_opportunity`: GET `/custom_field/opportunity/` - records path `data`; offset/limit
  pagination; offset parameter `_skip`; limit parameter `_limit`; page size 100.

## Write actions & risks

Overall write risk: external mutation; creates/updates/deletes live Close leads, contacts,
opportunities, and tasks.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_lead`: POST `/lead/` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `contacts`, `description`, `name`, `status_id`, `url`; risk: external mutation;
  creates a live Close lead; approval required.
- `update_lead`: PUT `/lead/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `description`, `id`, `name`, `status_id`, `url`;
  risk: external mutation; overwrites a live Close lead's fields; approval required.
- `delete_lead`: DELETE `/lead/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; risk: external mutation; irreversibly deletes a
  live Close lead and its contacts/opportunities; approval required.
- `create_contact`: POST `/contact/` - kind `create`; body type `json`; required record fields
  `lead_id`, `name`; accepted fields `emails`, `lead_id`, `name`, `phones`, `title`, `urls`; risk:
  external mutation; creates a live Close contact under a lead; approval required.
- `update_contact`: PUT `/contact/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `emails`, `id`, `name`, `phones`, `title`,
  `urls`; risk: external mutation; overwrites a live Close contact's fields; approval required.
- `delete_contact`: DELETE `/contact/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; irreversibly
  deletes a live Close contact; approval required.
- `create_opportunity`: POST `/opportunity/` - kind `create`; body type `json`; required record
  fields `lead_id`; accepted fields `confidence`, `contact_id`, `lead_id`, `note`, `status_id`,
  `user_id`, `value`, `value_period`; risk: external mutation; creates a live Close opportunity
  under a lead; approval required.
- `update_opportunity`: PUT `/opportunity/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `confidence`, `id`, `note`, `status_id`,
  `user_id`, `value`, `value_period`; risk: external mutation; overwrites a live Close opportunity's
  fields; approval required.
- `delete_opportunity`: DELETE `/opportunity/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live Close opportunity; approval required.
- `create_task`: POST `/task/` - kind `create`; body type `json`; required record fields `_type`,
  `lead_id`, `text`; accepted fields `_type`, `assigned_to`, `contact_id`, `due_date`,
  `is_complete`, `lead_id`, `text`; risk: external mutation; creates a live Close task on a lead;
  approval required.
- `update_task`: PUT `/task/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `assigned_to`, `due_date`, `id`, `is_complete`,
  `text`; risk: external mutation; overwrites a live Close task's fields; approval required.
- `delete_task`: DELETE `/task/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; risk: external mutation; irreversibly deletes a
  live Close task; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 14 stream-backed endpoint group(s), 12 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=17, duplicate_of=57, non_data_endpoint=16, out_of_scope=164,
  requires_elevated_scope=16.
