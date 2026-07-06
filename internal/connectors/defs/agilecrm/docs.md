# Overview

Reads AgileCRM contacts, deals, tasks, milestone pipelines, campaigns, and support tickets, and
writes contact/deal/task create, update, and delete actions, through the AgileCRM REST API.

Readable streams: `contacts`, `deals`, `tasks`, `milestone`, `campaigns`, `tickets_filters`,
`tickets`.

Write actions: `create_contact`, `update_contact`, `delete_contact`, `create_deal`, `update_deal`,
`delete_deal`, `create_task`, `update_task`, `delete_task`.

Service API documentation: https://github.com/agilecrm/rest-api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); AgileCRM API key, used as the HTTP Basic auth password.
  Never logged.
- `domain` (required, string); AgileCRM account subdomain (the <domain> in
  https://<domain>.agilecrm.com). Templated directly into the base URL; restrict to the safe label
  charset.
- `email` (required, string); AgileCRM account email, used as the HTTP Basic auth username.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- HTTP Basic authentication using `config.email`, `secrets.api_key`.

Requests use base URL `https://{{ config.domain }}.agilecrm.com/dev/api` after applying
configuration defaults.

Connection checks call GET `/contacts`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `contacts`, `deals`, `campaigns`, `tickets`; none: `tasks`,
`milestone`, `tickets_filters`.

- `contacts`: GET `/contacts` - records path `.`; query `page_size`=`50`; cursor pagination; cursor
  parameter `cursor`; next cursor from last record field `cursor`; computed output fields
  `owner_id`.
- `deals`: GET `/opportunity` - records path `.`; query `page_size`=`50`; cursor pagination; cursor
  parameter `cursor`; next cursor from last record field `cursor`; computed output fields
  `owner_id`.
- `tasks`: GET `/tasks` - records path `.`; computed output fields `owner_id`.
- `milestone`: GET `/milestone/pipelines` - records path `.`.
- `campaigns`: GET `/workflows` - records path `.`; query `page_size`=`50`; cursor pagination;
  cursor parameter `cursor`; next cursor from last record field `cursor`.
- `tickets_filters`: GET `/tickets/filters` - records path `.`.
- `tickets`: GET `/tickets/filter` - records path `.`; query `page_size`=`25`; cursor pagination;
  cursor parameter `cursor`; next cursor from last record field `cursor`; fan-out; ids from request
  `/tickets/filters`; id-list records path `.`; id field `id`; id sent as query parameter
  `filter_id`; stamps `filter_id`.

## Write actions & risks

Overall write risk: external mutation of live AgileCRM contacts, deals, and tasks including
irreversible deletes; approval required for every write action.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `properties`; accepted fields `lead_score`, `properties`, `star_value`, `tags`; risk: external
  mutation; creates a live AgileCRM contact/company; approval required.
- `update_contact`: PUT `/contacts/edit-properties` - kind `update`; body type `json`; required
  record fields `id`, `properties`; accepted fields `id`, `properties`; risk: external mutation;
  overwrites live AgileCRM contact property fields; approval required.
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live AgileCRM contact; approval required.
- `create_deal`: POST `/opportunity` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `close_date`, `expected_value`, `milestone`, `name`, `pipeline_id`,
  `probability`; risk: external mutation; creates a live AgileCRM deal; approval required.
- `update_deal`: PUT `/opportunity/partial-update` - kind `update`; body type `json`; required
  record fields `id`; accepted fields `expected_value`, `id`, `milestone`, `name`; risk: external
  mutation; overwrites live AgileCRM deal fields; approval required.
- `delete_deal`: DELETE `/opportunity/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live AgileCRM deal; approval required.
- `create_task`: POST `/tasks` - kind `create`; body type `json`; required record fields `subject`,
  `type`; accepted fields `contacts`, `due`, `priority_type`, `status`, `subject`, `type`; risk:
  external mutation; creates a live AgileCRM task; approval required.
- `update_task`: PUT `/tasks/partial-update` - kind `update`; body type `json`; required record
  fields `id`; accepted fields `due`, `id`, `status`, `subject`; risk: external mutation; overwrites
  live AgileCRM task fields; approval required.
- `delete_task`: DELETE `/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; irreversibly
  deletes a live AgileCRM task; approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 7 stream-backed endpoint group(s), 9 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=15, out_of_scope=39.
