# Overview

Reads Salesflare accounts, contacts, opportunities, users, tags, tasks, workflows, groups, stages,
pipelines, persons, currencies, custom-field types, and email data sources, and writes CRM lifecycle
mutations, through the Salesflare REST API.

Readable streams: `accounts`, `contacts`, `opportunities`, `users`, `tags`, `tasks`, `workflows`,
`groups`, `stages`, `pipelines`, `persons`, `currencies`, `custom_field_types`,
`email_data_sources`.

Write actions: `create_account`, `update_account`, `delete_account`, `create_contact`,
`update_contact`, `delete_contact`, `create_opportunity`, `update_opportunity`,
`delete_opportunity`, `create_tag`, `update_tag`, `delete_tag`, `create_task`, `update_task`,
`delete_task`, `create_meeting`, `update_meeting`, `delete_meeting`, `create_call`,
`create_internal_note`, `update_internal_note`, `delete_internal_note`.

Service API documentation: https://api.salesflare.com/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Salesflare API key. Sent as Authorization: Bearer <api_key>;
  never logged.
- `base_url` (optional, string); default `https://api.salesflare.com`; format `uri`; Salesflare API
  base URL. Also usable as a base URL override for tests/proxies.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.salesflare.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `page`; next token from
`pagination.next_page`; page size 100; maximum 100 page(s).

Pagination by stream: cursor: `accounts`, `contacts`, `opportunities`; none: `groups`, `stages`,
`pipelines`, `persons`, `currencies`, `custom_field_types`, `email_data_sources`; offset_limit:
`users`, `tags`, `tasks`, `workflows`.

- `accounts`: GET `/accounts` - records path `data`; query `limit`=`100`; `page`=`1`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next_page`; page size 100;
  maximum 100 page(s); emits passthrough records.
- `contacts`: GET `/contacts` - records path `data`; query `limit`=`100`; `page`=`1`; cursor
  pagination; cursor parameter `page`; next token from `pagination.next_page`; page size 100;
  maximum 100 page(s); emits passthrough records.
- `opportunities`: GET `/opportunities` - records path `data`; query `limit`=`100`; `page`=`1`;
  cursor pagination; cursor parameter `page`; next token from `pagination.next_page`; page size 100;
  maximum 100 page(s); emits passthrough records.
- `users`: GET `/users` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 2; emits passthrough records.
- `tags`: GET `/tags` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 2; emits passthrough records.
- `tasks`: GET `/tasks` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 2; emits passthrough records.
- `workflows`: GET `/workflows` - records path `data`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 2; emits passthrough records.
- `groups`: GET `/groups` - records path `data`; emits passthrough records.
- `stages`: GET `/stages` - records path `data`; emits passthrough records.
- `pipelines`: GET `/pipelines` - records path `data`; emits passthrough records.
- `persons`: GET `/persons` - records path `data`; emits passthrough records.
- `currencies`: GET `/currencies` - records path `data`; emits passthrough records.
- `custom_field_types`: GET `/customfields/types` - records path `data`; emits passthrough records.
- `email_data_sources`: GET `/datasources/email` - records path `data`; emits passthrough records.

## Write actions & risks

Overall write risk: external Salesflare mutations:
account/contact/opportunity/tag/task/meeting/call/internal-note create-update-delete.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_account`: POST `/accounts` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `city`, `country`, `domain`, `email`, `name`, `phone_number`; risk:
  creates a new CRM account; low-risk external mutation, no approval required.
- `update_account`: PUT `/accounts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `city`, `country`, `domain`, `email`, `id`,
  `name`, `phone_number`; risk: external mutation updating a CRM account; approval required.
- `delete_account`: DELETE `/accounts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: destructive/irreversible: permanently deletes a CRM account; approval
  required.
- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `account_id`, `email`, `first_name`, `last_name`, `name`, `phone_number`;
  risk: creates a new CRM contact; low-risk external mutation, no approval required.
- `update_contact`: PUT `/contacts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `account_id`, `email`, `first_name`, `id`,
  `last_name`, `name`, `phone_number`; risk: external mutation updating a CRM contact; approval
  required.
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: destructive/irreversible: permanently deletes a CRM contact; approval
  required.
- `create_opportunity`: POST `/opportunities` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `account_id`, `currency`, `name`, `pipeline_id`, `stage_id`,
  `value`; risk: creates a new CRM opportunity/deal; low-risk external mutation, no approval
  required.
- `update_opportunity`: PUT `/opportunities/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `closed`, `currency`, `id`, `name`,
  `stage_id`, `value`; risk: external mutation updating a CRM opportunity/deal (may change
  stage/close state); approval required.
- `delete_opportunity`: DELETE `/opportunities/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: destructive/irreversible: permanently deletes a CRM
  opportunity/deal; approval required.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`; risk: creates a new CRM tag; low-risk external mutation, no approval
  required.
- `update_tag`: PUT `/tags/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `id`, `name`; risk: external mutation renaming a CRM
  tag; approval required.
- `delete_tag`: DELETE `/tags/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; risk: destructive/irreversible: permanently deletes a CRM tag from every record it's
  applied to; approval required.
- `create_task`: POST `/tasks` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `account_id`, `assignee_id`, `description`, `due_date`, `name`; risk: creates a
  new CRM task; low-risk external mutation, no approval required.
- `update_task`: PUT `/tasks/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `completed`, `description`, `due_date`, `id`, `name`;
  risk: external mutation updating a CRM task (may mark complete); approval required.
- `delete_task`: DELETE `/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: destructive/irreversible: permanently deletes a CRM task; approval required.
- `create_meeting`: POST `/meetings` - kind `create`; body type `json`; required record fields
  `title`, `start_date`, `end_date`; accepted fields `account_id`, `description`, `end_date`,
  `start_date`, `title`; risk: creates a new CRM meeting/calendar entry; low-risk external mutation,
  no approval required.
- `update_meeting`: PUT `/meetings/{{ record.meeting_id }}` - kind `update`; body type `json`; path
  fields `meeting_id`; required record fields `meeting_id`; accepted fields `end_date`,
  `meeting_id`, `start_date`, `title`; risk: external mutation updating a CRM meeting/calendar
  entry; approval required.
- `delete_meeting`: DELETE `/meetings/{{ record.meeting_id }}` - kind `delete`; body type `none`;
  path fields `meeting_id`; required record fields `meeting_id`; accepted fields `meeting_id`;
  missing records treated as success for status `404`; risk: destructive/irreversible: permanently
  deletes a CRM meeting/calendar entry; approval required.
- `create_call`: POST `/calls` - kind `create`; body type `json`; required record fields
  `account_id`; accepted fields `account_id`, `date`, `notes`; risk: logs a new call activity
  against a CRM account; low-risk external mutation, no approval required.
- `create_internal_note`: POST `/messages` - kind `create`; body type `json`; required record fields
  `content`; accepted fields `account_id`, `contact_id`, `content`, `opportunity_id`; risk: creates
  a new internal note on a CRM record; low-risk external mutation, no approval required.
- `update_internal_note`: PUT `/messages/{{ record.message_id }}` - kind `update`; body type `json`;
  path fields `message_id`; required record fields `message_id`; accepted fields `content`,
  `message_id`; risk: external mutation editing a CRM internal note; approval required.
- `delete_internal_note`: DELETE `/messages/{{ record.message_id }}` - kind `delete`; body type
  `none`; path fields `message_id`; required record fields `message_id`; accepted fields
  `message_id`; missing records treated as success for status `404`; risk: destructive/irreversible:
  permanently deletes a CRM internal note; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 14 stream-backed endpoint group(s), 22 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=12, non_data_endpoint=5, out_of_scope=12,
  requires_elevated_scope=6.
