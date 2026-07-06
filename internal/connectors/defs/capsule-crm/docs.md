# Overview

Reads Capsule CRM parties, opportunities, cases, tasks, users, tags, custom field definitions,
teams, pipelines, milestones, lost reasons, task categories, boards, and stages, and writes
party/opportunity/case/task create, update, and delete actions, through the Capsule v2 REST API.

Readable streams: `parties`, `opportunities`, `kases`, `tasks`, `users`, `tags`, `custom_fields`,
`teams`, `pipelines`, `milestones`, `lost_reasons`, `categories`, `boards`, `stages`.

Write actions: `create_party`, `update_party`, `delete_party`, `create_opportunity`,
`update_opportunity`, `delete_opportunity`, `create_kase`, `update_kase`, `delete_kase`,
`create_task`, `update_task`, `delete_task`.

Service API documentation: https://developer.capsulecrm.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.capsulecrm.com/api/v2`; format `uri`; Capsule
  CRM API base URL override for tests or proxies.
- `bearer_token` (required, secret, string); Capsule CRM personal access token. Used only for Bearer
  auth; never logged.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `bearer_token`.

Default configuration values: `base_url=https://api.capsulecrm.com/api/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.bearer_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/parties` with query `page`=`1`; `perPage`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `perPage`; starts
at 1; page size 50.

Pagination by stream: none: `tags`, `custom_fields`, `teams`, `pipelines`, `milestones`,
`lost_reasons`, `categories`, `boards`, `stages`; page_number: `parties`, `opportunities`, `kases`,
`tasks`, `users`.

- `parties`: GET `/parties` - records path `parties`; page-number pagination; page parameter `page`;
  size parameter `perPage`; starts at 1; page size 50; computed output fields `created_at`,
  `first_name`, `job_title`, `last_contacted_at`, `last_name`, `organisation_name`, `updated_at`.
- `opportunities`: GET `/opportunities` - records path `opportunities`; page-number pagination; page
  parameter `page`; size parameter `perPage`; starts at 1; page size 50; computed output fields
  `closed_on`, `created_at`, `expected_close_on`, `lost_reason`, `milestone_id`, `milestone_name`,
  `party_id`, `updated_at`, `value_amount`, `value_currency`.
- `kases`: GET `/kases` - records path `kases`; page-number pagination; page parameter `page`; size
  parameter `perPage`; starts at 1; page size 50; computed output fields `closed_on`, `created_at`,
  `party_id`, `updated_at`.
- `tasks`: GET `/tasks` - records path `tasks`; page-number pagination; page parameter `page`; size
  parameter `perPage`; starts at 1; page size 50; computed output fields `category_id`,
  `created_at`, `due_on`, `kase_id`, `opportunity_id`, `party_id`, `updated_at`.
- `users`: GET `/users` - records path `users`; page-number pagination; page parameter `page`; size
  parameter `perPage`; starts at 1; page size 50; computed output fields `created_at`, `updated_at`.
- `tags`: GET `/tags` - records path `tags`.
- `custom_fields`: GET `/customfields` - records path `definitions`; computed output fields
  `entity_type`, `restricted_to_type`.
- `teams`: GET `/teams` - records path `teams`.
- `pipelines`: GET `/pipelines` - records path `pipelines`; computed output fields `created_at`,
  `display_order`, `updated_at`.
- `milestones`: GET `/milestones` - records path `milestones`; computed output fields `pipeline_id`.
- `lost_reasons`: GET `/lostreasons` - records path `lostReasons`.
- `categories`: GET `/categories` - records path `categories`.
- `boards`: GET `/boards` - records path `boards`; computed output fields `created_at`,
  `entity_type`, `updated_at`.
- `stages`: GET `/stages` - records path `stages`; computed output fields `board_id`,
  `display_order`.

## Write actions & risks

Overall write risk: external mutation of live Capsule CRM parties, opportunities, cases, and tasks
including irreversible deletes; approval required for every write action.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_party`: POST `/parties` - kind `create`; body type `json`; required record fields `party`;
  accepted fields `party`; risk: external mutation; creates a live Capsule CRM contact; approval
  required. Body wraps the record under a top-level "party" key (Capsule's resource-envelope
  convention) - the record itself must carry that wrapper, since the engine's write dialect sends
  record fields verbatim as the JSON body with no nested-wrapper construction primitive.
- `update_party`: PUT `/parties/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `party`; required record fields `id`, `party`; accepted fields `id`, `party`;
  risk: external mutation; updates a live Capsule CRM contact; approval required. Body wraps the
  record under a top-level "party" key; "id" is path-only (path_fields) and excluded from the body
  via body_fields.
- `delete_party`: DELETE `/parties/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; irreversibly
  deletes a live Capsule CRM contact and its associated history; approval required.
- `create_opportunity`: POST `/opportunities` - kind `create`; body type `json`; required record
  fields `opportunity`; accepted fields `opportunity`; risk: external mutation; creates a live
  Capsule CRM sales opportunity; approval required. Body wraps the record under a top-level
  "opportunity" key.
- `update_opportunity`: PUT `/opportunities/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `opportunity`; required record fields `id`, `opportunity`; accepted
  fields `id`, `opportunity`; risk: external mutation; updates a live Capsule CRM sales opportunity
  (including moving pipeline stage or closing/losing it); approval required.
- `delete_opportunity`: DELETE `/opportunities/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live Capsule CRM sales opportunity; approval required.
- `create_kase`: POST `/kases` - kind `create`; body type `json`; required record fields `kase`;
  accepted fields `kase`; risk: external mutation; creates a live Capsule CRM case/project; approval
  required. Body wraps the record under a top-level "kase" key (Capsule kept the "kase" spelling in
  the API after renaming Cases to Projects in the product UI, to avoid a breaking change; see
  docs.md).
- `update_kase`: PUT `/kases/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  body fields `kase`; required record fields `id`, `kase`; accepted fields `id`, `kase`; risk:
  external mutation; updates a live Capsule CRM case/project, including closing it; approval
  required.
- `delete_kase`: DELETE `/kases/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; irreversibly
  deletes a live Capsule CRM case/project; approval required.
- `create_task`: POST `/tasks` - kind `create`; body type `json`; required record fields `task`;
  accepted fields `task`; risk: external mutation; creates a live Capsule CRM task/reminder;
  approval required. Body wraps the record under a top-level "task" key.
- `update_task`: PUT `/tasks/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  body fields `task`; required record fields `id`, `task`; accepted fields `id`, `task`; risk:
  external mutation; updates a live Capsule CRM task, including marking it complete; approval
  required.
- `delete_task`: DELETE `/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; irreversibly
  deletes a live Capsule CRM task; approval required.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 14 stream-backed endpoint group(s), 12 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=14, duplicate_of=25, non_data_endpoint=3, out_of_scope=35,
  requires_elevated_scope=27.
