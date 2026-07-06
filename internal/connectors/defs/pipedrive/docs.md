# Overview

Reads Pipedrive deals, persons, organizations, activities, products, users, notes, leads, saved
filters, activity types, roles, webhooks, and field/reference metadata, and writes
lead/note/filter/activity-type/lead-label/webhook mutations through REST API v1.

Readable streams: `deals`, `persons`, `organizations`, `activities`, `products`, `users`, `notes`,
`leads`, `deal_fields`, `person_fields`, `organization_fields`, `product_fields`, `lead_fields`,
`roles`, `filters`, `activity_types`, `teams`, `webhooks`, `lead_labels`, `lead_sources`,
`currencies`.

Write actions: `create_lead`, `update_lead`, `delete_lead`, `create_note`, `update_note`,
`delete_note`, `create_filter`, `update_filter`, `delete_filter`, `create_activity_type`,
`update_activity_type`, `delete_activity_type`, `create_lead_label`, `update_lead_label`,
`delete_lead_label`, `create_webhook`, `delete_webhook`.

Service API documentation: https://developers.pipedrive.com/docs/api/v1.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Pipedrive API token. Sent as the api_token query parameter
  on every request; never logged.
- `base_url` (optional, string); default `https://api.pipedrive.com/v1`; format `uri`; Pipedrive API
  base URL override for tests or proxies.
- `replication_start_date` (optional, string); RFC3339 or Pipedrive-accepted timestamp lower bound;
  passed as the stream's since parameter (since_timestamp for deals/persons/organizations, since for
  activities) when set. Not used by products or users.

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `base_url=https://api.pipedrive.com/v1`.

Authentication behavior:

- API key authentication in query parameter `api_token` using `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `limit`=`1`; `start`=`0`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `start`; next token from
`additional_data.pagination.next_start`.

Pagination by stream: cursor: `deals`, `persons`, `organizations`, `activities`, `products`,
`users`, `notes`; none: `filters`, `activity_types`, `teams`, `webhooks`, `lead_labels`,
`lead_sources`, `currencies`; offset_limit: `leads`, `deal_fields`, `person_fields`,
`organization_fields`, `product_fields`, `lead_fields`, `roles`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `deals`: GET `/deals` - records path `data`; query `limit`=`100`; `start`=`0`; cursor pagination;
  cursor parameter `start`; next token from `additional_data.pagination.next_start`; incremental
  cursor `update_time`; sent as `since_timestamp`; formatted as `rfc3339`; initial lower bound from
  `replication_start_date`; emits passthrough records.
- `persons`: GET `/persons` - records path `data`; query `limit`=`100`; `start`=`0`; cursor
  pagination; cursor parameter `start`; next token from `additional_data.pagination.next_start`;
  incremental cursor `update_time`; sent as `since_timestamp`; formatted as `rfc3339`; initial lower
  bound from `replication_start_date`; emits passthrough records.
- `organizations`: GET `/organizations` - records path `data`; query `limit`=`100`; `start`=`0`;
  cursor pagination; cursor parameter `start`; next token from
  `additional_data.pagination.next_start`; incremental cursor `update_time`; sent as
  `since_timestamp`; formatted as `rfc3339`; initial lower bound from `replication_start_date`;
  emits passthrough records.
- `activities`: GET `/activities` - records path `data`; query `limit`=`100`; `start`=`0`; cursor
  pagination; cursor parameter `start`; next token from `additional_data.pagination.next_start`;
  incremental cursor `update_time`; sent as `since`; formatted as `rfc3339`; initial lower bound
  from `replication_start_date`; emits passthrough records.
- `products`: GET `/products` - records path `data`; query `limit`=`100`; `start`=`0`; cursor
  pagination; cursor parameter `start`; next token from `additional_data.pagination.next_start`;
  emits passthrough records.
- `users`: GET `/users` - records path `data`; query `limit`=`100`; `start`=`0`; cursor pagination;
  cursor parameter `start`; next token from `additional_data.pagination.next_start`; emits
  passthrough records.
- `notes`: GET `/notes` - records path `data`; query `limit`=`100`; `start`=`0`; `updated_since`
  from template `{{ config.replication_start_date }}`, omitted when absent; cursor pagination;
  cursor parameter `start`; next token from `additional_data.pagination.next_start`; emits
  passthrough records.
- `leads`: GET `/leads` - records path `data`; query `limit`=`100`; `start`=`0`; offset/limit
  pagination; offset parameter `start`; limit parameter `limit`; page size 100; emits passthrough
  records.
- `deal_fields`: GET `/dealFields` - records path `data`; query `limit`=`100`; `start`=`0`;
  offset/limit pagination; offset parameter `start`; limit parameter `limit`; page size 100; emits
  passthrough records.
- `person_fields`: GET `/personFields` - records path `data`; query `limit`=`100`; `start`=`0`;
  offset/limit pagination; offset parameter `start`; limit parameter `limit`; page size 100; emits
  passthrough records.
- `organization_fields`: GET `/organizationFields` - records path `data`; query `limit`=`100`;
  `start`=`0`; offset/limit pagination; offset parameter `start`; limit parameter `limit`; page size
  100; emits passthrough records.
- `product_fields`: GET `/productFields` - records path `data`; query `limit`=`100`; `start`=`0`;
  offset/limit pagination; offset parameter `start`; limit parameter `limit`; page size 100; emits
  passthrough records.
- `lead_fields`: GET `/leadFields` - records path `data`; query `limit`=`100`; `start`=`0`;
  offset/limit pagination; offset parameter `start`; limit parameter `limit`; page size 100; emits
  passthrough records.
- `roles`: GET `/roles` - records path `data`; query `limit`=`100`; `start`=`0`; offset/limit
  pagination; offset parameter `start`; limit parameter `limit`; page size 100; emits passthrough
  records.
- `filters`: GET `/filters` - records path `data`; emits passthrough records.
- `activity_types`: GET `/activityTypes` - records path `data`; emits passthrough records.
- `stream`: GET connector-managed request path - records path `data`; emits passthrough records.
- `webhooks`: GET `/webhooks` - records path `data`; emits passthrough records.
- `lead_labels`: GET `/leadLabels` - records path `data`; emits passthrough records.
- `lead_sources`: GET `/leadSources` - records path `data`; emits passthrough records.
- `currencies`: GET `/currencies` - records path `data`; emits passthrough records.

## Write actions & risks

Overall write risk: creates/updates/deletes leads, notes, saved filters, custom activity types, lead
labels, and webhook subscriptions.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_lead`: POST `/leads` - kind `create`; body type `json`; required record fields `title`;
  accepted fields `channel`, `channel_id`, `expected_close_date`, `label_ids`, `organization_id`,
  `origin_id`, `owner_id`, `person_id`, `title`, `value`, `visible_to`, `was_seen`; risk: creates a
  new lead; low-risk external mutation, no approval required.
- `update_lead`: PATCH `/leads/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `channel`, `channel_id`, `expected_close_date`, `id`,
  `is_archived`, `label_ids`, `organization_id`, `owner_id`, `person_id`, `title`, `value`,
  `visible_to`, `was_seen`; risk: updates an existing lead's fields (partial patch); external
  mutation, approval required.
- `delete_lead`: DELETE `/leads/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  permanently deletes a lead; destructive external mutation, approval required.
- `create_note`: POST `/notes` - kind `create`; body type `json`; required record fields `content`;
  accepted fields `content`, `deal_id`, `lead_id`, `org_id`, `person_id`, `pinned_to_deal_flag`,
  `pinned_to_organization_flag`, `pinned_to_person_flag`, `project_id`, `task_id`, `user_id`; risk:
  creates a new note attached to a deal/person/organization/lead; low-risk external mutation, no
  approval required.
- `update_note`: PUT `/notes/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `content`; accepted fields `content`, `id`; risk: updates an existing
  note's content; external mutation, approval required.
- `delete_note`: DELETE `/notes/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  permanently deletes a note; destructive external mutation, approval required.
- `create_filter`: POST `/filters` - kind `create`; body type `json`; required record fields `name`,
  `conditions`, `type`; accepted fields `conditions`, `name`, `type`; risk: creates a new saved
  filter; low-risk external mutation, no approval required.
- `update_filter`: PUT `/filters/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `name`, `conditions`; accepted fields `conditions`, `id`,
  `name`; risk: updates an existing saved filter's name/conditions; external mutation, approval
  required.
- `delete_filter`: DELETE `/filters/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  permanently deletes a saved filter; destructive external mutation, approval required.
- `create_activity_type`: POST `/activityTypes` - kind `create`; body type `json`; required record
  fields `name`, `icon_key`; accepted fields `color`, `icon_key`, `name`; risk: creates a new custom
  activity type; low-risk external mutation, no approval required.
- `update_activity_type`: PUT `/activityTypes/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `color`, `id`, `name`, `order_nr`;
  risk: updates an existing activity type's name/color/order; external mutation, approval required.
- `delete_activity_type`: DELETE `/activityTypes/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`;
  risk: permanently deletes a custom activity type; destructive external mutation, approval
  required.
- `create_lead_label`: POST `/leadLabels` - kind `create`; body type `json`; required record fields
  `name`, `color`; accepted fields `color`, `name`; risk: creates a new lead label; low-risk
  external mutation, no approval required.
- `update_lead_label`: PATCH `/leadLabels/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `color`, `id`, `name`; risk: updates an
  existing lead label's name/color; external mutation, approval required.
- `delete_lead_label`: DELETE `/leadLabels/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  permanently deletes a lead label; destructive external mutation, approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `subscription_url`, `event_action`, `event_object`, `name`; accepted fields `event_action`,
  `event_object`, `name`, `subscription_url`, `user_id`; risk: registers a new webhook subscription
  that will receive event notifications; low-risk external mutation, no approval required.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  permanently deletes a webhook subscription; destructive external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 20 stream-backed endpoint group(s), 17 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=11, destructive_admin=11, duplicate_of=21, non_data_endpoint=21, out_of_scope=116.
