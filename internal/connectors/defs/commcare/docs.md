# Overview

Reads CommCare HQ project, application, form, case, user, group, report, location, lookup table,
export, and messaging API data; writes supported JSON mutations for cases, users, groups, locations,
and lookup tables.

Readable streams: `forms`, `cases`, `applications`, `application`, `multimedia_upload_status`,
`forms_v1`, `form`, `cases_v1`, `case_v1`, `cases_v2`, `case_v2`, `case_v2_by_external_id`,
`case_v2_bulk_by_ids`, `case_v2_index_children`, `mobile_workers`, `mobile_worker`, `bulk_users`,
`web_users`, `web_user`, `user_domains`, `user_identity`, `groups`, `group`, `reports`,
`report_data`, `locations_v1`, `location_v1`, `locations_v2`, `location_v2`, `location_types`,
`location_type`, `fixtures`, `fixture_table_items`, `fixture_item`, `lookup_tables`,
`lookup_table_rows`, `det_exports`, `messaging_events`.

Write actions: `create_case_v2`, `update_case_v2`, `upsert_case_v2_by_external_id`,
`upsert_case_v2`, `create_mobile_worker`, `update_mobile_worker`, `delete_mobile_worker`,
`send_mobile_worker_password_reset`, `create_web_user_invitation`, `update_web_user`,
`enable_web_user`, `disable_web_user`, `create_group`, `create_groups_bulk`, `update_group`,
`delete_group`, `create_location_v2`, `update_location_v2`, `bulk_upsert_locations_v2`,
`create_lookup_table`, `update_lookup_table`, `delete_lookup_table`, `create_lookup_table_row`,
`update_lookup_table_row`, `delete_lookup_table_row`.

Service API documentation: https://commcare-hq.readthedocs.io/api/index.html.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); CommCare HQ API key. Sent as an ApiKey-scheme Authorization
  header (Authorization: ApiKey <api_key>); never logged.
- `app_id` (optional, string); Optional CommCare application id; used by application detail,
  multimedia status, and as a form/case filter where supported.
- `base_url` (optional, string); default `https://www.commcarehq.org`; format `uri`; CommCare HQ
  base URL override for self-hosted instances, tests, or proxies.
- `case_id` (optional, string); Case id for case detail, child-index, and case update endpoints.
- `case_ids` (optional, string); Comma-separated case ids for the Case Data API v2 GET bulk-by-id
  endpoint.
- `external_id` (optional, string); External id for Case Data API v2 external-id lookups and
  upserts.
- `fixture_item_id` (optional, string).
- `fixture_type` (optional, string).
- `form_id` (optional, string); Form submission id for form detail endpoints.
- `group_id` (optional, string); Group id for group detail and mutation endpoints.
- `location_id` (optional, string); Location id for Location API v1/v2 detail and update endpoints.
- `location_type_id` (optional, string); Location type id for location type detail endpoints.
- `lookup_table_id` (optional, string); Lookup table id for update and delete actions.
- `lookup_table_item_id` (optional, string); Lookup table row id for update and delete actions.
- `mobile_worker_id` (optional, string); Mobile worker id for user detail and mutation endpoints.
- `processing_id` (optional, string); Application multimedia upload processing id for polling upload
  status.
- `project_space` (required, string); CommCare HQ project space slug; substituted into
  project-scoped API paths.
- `report_id` (optional, string); Report id for the configurable report data endpoint.
- `web_user_id` (optional, string); Web user id for web-user detail and mutation endpoints.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://www.commcarehq.org`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `ApiKey` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/a/{{ config.project_space }}/api/v0.5/form/` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100; maximum 100 page(s).

Pagination by stream: next_url: `cases_v2`, `case_v2_index_children`, `messaging_events`; none:
`application`, `multimedia_upload_status`, `form`, `case_v1`, `case_v2`, `case_v2_by_external_id`,
`case_v2_bulk_by_ids`, `mobile_worker`, `web_user`, `user_identity`, `group`, `reports`,
`location_v1`, `location_v2`, `location_type`, `fixtures`, `fixture_table_items`, `fixture_item`;
offset_limit: `forms`, `cases`, `applications`, `forms_v1`, `cases_v1`, `mobile_workers`,
`bulk_users`, `web_users`, `user_domains`, `groups`, `report_data`, `locations_v1`, `locations_v2`,
`location_types`, `lookup_tables`, `lookup_table_rows`, `det_exports`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `forms`: GET `/a/{{ config.project_space }}/api/v0.5/form/` - records path `objects`; query
  `app_id` from template `{{ config.app_id }}`, omitted when absent; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 100 page(s); incremental
  cursor `received_on`; formatted as `rfc3339`; emits passthrough records.
- `cases`: GET `/a/{{ config.project_space }}/api/v0.5/case/` - records path `objects`; query
  `app_id` from template `{{ config.app_id }}`, omitted when absent; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 100 page(s); incremental
  cursor `server_modified_on`; formatted as `rfc3339`; emits passthrough records.
- `applications`: GET `/a/{{ config.project_space }}/api/application/v1/` - records path `objects`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  maximum 100 page(s); emits passthrough records.
- `application`: GET `/a/{{ config.project_space }}/api/application/v1/{{ config.app_id }}` -
  records path `objects`; emits passthrough records.
- `multimedia_upload_status`: GET `/a/{{ config.project_space }}/apps/api/{{ config.app_id
  }}/multimedia/status/{{ config.processing_id }}/` - single-object response; records path `.`;
  emits passthrough records.
- `forms_v1`: GET `/a/{{ config.project_space }}/api/form/v1/` - records path `objects`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  maximum 100 page(s); emits passthrough records.
- `form`: GET `/a/{{ config.project_space }}/api/form/v1/{{ config.form_id }}/` - single-object
  response; records path `.`; emits passthrough records.
- `cases_v1`: GET `/a/{{ config.project_space }}/api/case/v1/` - records path `.`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; maximum 100
  page(s); emits passthrough records.
- `case_v1`: GET `/a/{{ config.project_space }}/api/case/v1/{{ config.case_id }}/` - single-object
  response; records path `.`; emits passthrough records.
- `cases_v2`: GET `/a/{{ config.project_space }}/api/case/v2/` - records path `cases`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host;
  emits passthrough records.
- `case_v2`: GET `/a/{{ config.project_space }}/api/case/v2/{{ config.case_id }}` - single-object
  response; records path `.`; emits passthrough records.
- `case_v2_by_external_id`: GET `/a/{{ config.project_space }}/api/case/v2/ext/{{ config.external_id
  }}/` - single-object response; records path `.`; emits passthrough records.
- `case_v2_bulk_by_ids`: GET `/a/{{ config.project_space }}/api/case/v2/{{ config.case_ids }}` -
  records path `cases`; emits passthrough records.
- `case_v2_index_children`: GET `/a/{{ config.project_space }}/api/case/v2/` - records path `cases`;
  query `indices.parent`=`{{ config.case_id }}`; follows a next-page URL from the response body; URL
  path `next`; next URLs stay on the configured API host; emits passthrough records.
- `mobile_workers`: GET `/a/{{ config.project_space }}/api/user/v1/` - records path `objects`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  maximum 100 page(s); emits passthrough records.
- `mobile_worker`: GET `/a/{{ config.project_space }}/api/user/v1/{{ config.mobile_worker_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `bulk_users`: GET `/a/{{ config.project_space }}/api/bulk-user/v1/` - records path `objects`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  maximum 100 page(s); emits passthrough records.
- `web_users`: GET `/a/{{ config.project_space }}/api/web-user/v1/` - records path `objects`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  maximum 100 page(s); emits passthrough records.
- `web_user`: GET `/a/{{ config.project_space }}/api/web-user/v1/{{ config.web_user_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `user_domains`: GET `/api/user_domains/v1/` - records path `objects`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 100 page(s); emits
  passthrough records.
- `user_identity`: GET `/api/identity/v1/` - single-object response; records path `.`; emits
  passthrough records.
- `groups`: GET `/a/{{ config.project_space }}/api/group/v1/` - records path `objects`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; maximum 100
  page(s); emits passthrough records.
- `group`: GET `/a/{{ config.project_space }}/api/group/v1/{{ config.group_id }}/` - single-object
  response; records path `.`; emits passthrough records.
- `reports`: GET `/a/{{ config.project_space }}/api/simplereportconfiguration/v1/` - records path
  `.`; query `format`=`json`; emits passthrough records.
- `report_data`: GET `/a/{{ config.project_space }}/api/configurablereportdata/v1/{{
  config.report_id }}/` - records path `data`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 50; maximum 100 page(s); emits passthrough records.
- `locations_v1`: GET `/a/{{ config.project_space }}/api/location/v1/` - records path `objects`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  maximum 100 page(s); emits passthrough records.
- `location_v1`: GET `/a/{{ config.project_space }}/api/location/v1/{{ config.location_id }}` -
  single-object response; records path `.`; emits passthrough records.
- `locations_v2`: GET `/a/{{ config.project_space }}/api/location/v2/` - records path `objects`;
  query `format`=`json`; offset/limit pagination; offset parameter `offset`; limit parameter
  `limit`; page size 100; maximum 100 page(s); emits passthrough records.
- `location_v2`: GET `/a/{{ config.project_space }}/api/location/v2/{{ config.location_id }}` -
  single-object response; records path `.`; query `format`=`json`; emits passthrough records.
- `location_types`: GET `/a/{{ config.project_space }}/api/location_type/v1/` - records path
  `objects`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size
  100; maximum 100 page(s); emits passthrough records.
- `location_type`: GET `/a/{{ config.project_space }}/api/location_type/v1/{{
  config.location_type_id }}` - single-object response; records path `.`; emits passthrough records.
- `fixtures`: GET `/a/{{ config.project_space }}/api/fixture/v1/` - records path `.`; emits
  passthrough records.
- `fixture_table_items`: GET `/a/{{ config.project_space }}/api/fixture/v1/` - records path `.`;
  query `fixture_type`=`{{ config.fixture_type }}`; emits passthrough records.
- `fixture_item`: GET `/a/{{ config.project_space }}/api/fixture/v1/{{ config.fixture_item_id }}/` -
  single-object response; records path `.`; emits passthrough records.
- `lookup_tables`: GET `/a/{{ config.project_space }}/api/lookup_table/v1/` - records path
  `objects`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size
  100; maximum 100 page(s); emits passthrough records.
- `lookup_table_rows`: GET `/a/{{ config.project_space }}/api/lookup_table_item/v1/` - records path
  `objects`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size
  100; maximum 100 page(s); emits passthrough records.
- `det_exports`: GET `/a/{{ config.project_space }}/api/det_export_instance/v1/` - records path
  `objects`; offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size
  100; maximum 100 page(s); emits passthrough records.
- `messaging_events`: GET `/a/{{ config.project_space }}/api/messaging-event/v1/` - records path
  `objects`; follows a next-page URL from the response body; URL path `meta.next`; next URLs stay on
  the configured API host; emits passthrough records.

## Write actions & risks

Overall write risk: external CommCare HQ mutations for cases, mobile workers, web-user invitations
and access, groups, locations, lookup tables, and lookup table rows.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_case_v2`: POST `/a/{{ config.project_space }}/api/case/v2/` - kind `create`; body type
  `json`; required record fields `case_type`, `case_name`, `owner_id`; accepted fields `case_name`,
  `case_type`, `close`, `external_id`, `indices`, `owner_id`, `properties`; risk: creates a CommCare
  case by submitting a server-generated XForm.
- `update_case_v2`: PUT `/a/{{ config.project_space }}/api/case/v2/{{ record.case_id }}` - kind
  `update`; body type `json`; path fields `case_id`; required record fields `case_id`; accepted
  fields `case_id`, `case_name`, `case_type`, `close`, `external_id`, `indices`, `owner_id`,
  `properties`; risk: updates an existing CommCare case by id.
- `upsert_case_v2_by_external_id`: PUT `/a/{{ config.project_space }}/api/case/v2/ext/{{
  record.external_id }}/` - kind `upsert`; body type `json`; path fields `external_id`; required
  record fields `external_id`; accepted fields `case_name`, `case_type`, `close`, `external_id`,
  `indices`, `owner_id`, `properties`; risk: updates or creates a CommCare case matched by external
  id.
- `upsert_case_v2`: PUT `/a/{{ config.project_space }}/api/case/v2/` - kind `upsert`; body type
  `json`; required record fields `external_id`; accepted fields `case_name`, `case_type`, `close`,
  `external_id`, `indices`, `owner_id`, `properties`; risk: updates or creates a CommCare case
  matched by the request body's external_id.
- `create_mobile_worker`: POST `/a/{{ config.project_space }}/api/user/v1/` - kind `create`; body
  type `json`; required record fields `username`, `email`; accepted fields `default_phone_number`,
  `email`, `first_name`, `groups`, `language`, `last_name`, `locations`, `password`,
  `phone_numbers`, `primary_location`, `require_account_confirmation`,
  `send_confirmation_email_now`, `user_data`, `username`; risk: creates a mobile worker account;
  password-bearing creation is intentionally not represented in fixtures or docs.
- `update_mobile_worker`: PUT `/a/{{ config.project_space }}/api/user/v1/{{ record.mobile_worker_id
  }}/` - kind `update`; body type `json`; path fields `mobile_worker_id`; required record fields
  `mobile_worker_id`; accepted fields `default_phone_number`, `email`, `first_name`, `groups`,
  `language`, `last_name`, `locations`, `mobile_worker_id`, `password`, `phone_numbers`,
  `primary_location`, `send_confirmation_email_now`, `user_data`; risk: updates a mobile worker
  profile and assignments.
- `delete_mobile_worker`: DELETE `/a/{{ config.project_space }}/api/user/v1/{{
  record.mobile_worker_id }}/` - kind `delete`; body type `none`; path fields `mobile_worker_id`;
  required record fields `mobile_worker_id`; accepted fields `mobile_worker_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a mobile worker.
- `send_mobile_worker_password_reset`: POST `/a/{{ config.project_space }}/api/user/v1/{{
  record.mobile_worker_id }}/email_password_reset/` - kind `custom`; body type `none`; path fields
  `mobile_worker_id`; required record fields `mobile_worker_id`; accepted fields `mobile_worker_id`;
  risk: sends a password reset email to a mobile worker.
- `create_web_user_invitation`: POST `/a/{{ config.project_space }}/api/invitation/v1/` - kind
  `create`; body type `json`; required record fields `email`, `role`; accepted fields
  `assigned_location_ids`, `email`, `primary_location_id`, `profile`, `role`, `tableau_groups`,
  `tableau_role`, `user_data`; risk: invites a web user to the project.
- `update_web_user`: PATCH `/a/{{ config.project_space }}/api/web-user/v1/{{ record.web_user_id }}/`
  - kind `update`; body type `json`; path fields `web_user_id`; required record fields
  `web_user_id`; accepted fields `assigned_location_ids`, `primary_location_id`, `profile`, `role`,
  `tableau_groups`, `tableau_role`, `user_data`, `web_user_id`; risk: updates a web user's role,
  locations, profile, and custom data.
- `enable_web_user`: POST `/a/{{ config.project_space }}/api/web-user/v1/{{ record.web_user_id
  }}/enable` - kind `custom`; body type `none`; path fields `web_user_id`; required record fields
  `web_user_id`; accepted fields `web_user_id`; risk: enables a web user account.
- `disable_web_user`: POST `/a/{{ config.project_space }}/api/web-user/v1/{{ record.web_user_id
  }}/disable` - kind `custom`; body type `none`; path fields `web_user_id`; required record fields
  `web_user_id`; accepted fields `web_user_id`; risk: disables a web user account.
- `create_group`: POST `/a/{{ config.project_space }}/api/group/v1/` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `case_sharing`, `metadata`, `name`,
  `reporting`, `users`; risk: creates a user group.
- `create_groups_bulk`: PATCH `/a/{{ config.project_space }}/api/group/v1/` - kind `create`; body
  type `json`; required record fields `objects`; accepted fields `objects`; risk: creates multiple
  user groups from one request body.
- `update_group`: PUT `/a/{{ config.project_space }}/api/group/v1/{{ record.group_id }}/` - kind
  `update`; body type `json`; path fields `group_id`; required record fields `group_id`; accepted
  fields `case_sharing`, `group_id`, `metadata`, `name`, `reporting`, `users`; risk: updates a user
  group and replaces provided assignments/custom metadata.
- `delete_group`: DELETE `/a/{{ config.project_space }}/api/group/v1/{{ record.group_id }}/` - kind
  `delete`; body type `none`; path fields `group_id`; required record fields `group_id`; accepted
  fields `group_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a user group.
- `create_location_v2`: POST `/a/{{ config.project_space }}/api/location/v2/` - kind `create`; body
  type `json`; required record fields `name`, `location_type_code`; accepted fields `latitude`,
  `location_data`, `location_type_code`, `longitude`, `name`, `parent_location_id`, `site_code`;
  risk: creates a location in the project hierarchy.
- `update_location_v2`: PUT `/a/{{ config.project_space }}/api/location/v2/{{ record.location_id }}`
  - kind `update`; body type `json`; path fields `location_id`; required record fields
  `location_id`; accepted fields `latitude`, `location_data`, `location_id`, `location_type_code`,
  `longitude`, `name`, `parent_location_id`, `site_code`; risk: updates a location in the project
  hierarchy.
- `bulk_upsert_locations_v2`: PATCH `/a/{{ config.project_space }}/api/location/v2/` - kind
  `upsert`; body type `json`; required record fields `objects`; accepted fields `objects`; risk:
  atomically creates and updates multiple locations.
- `create_lookup_table`: POST `/a/{{ config.project_space }}/api/lookup_table/v1/` - kind `create`;
  body type `json`; required record fields `tag`, `fields`; accepted fields `fields`, `is_global`,
  `tag`; risk: creates a lookup table definition.
- `update_lookup_table`: PUT `/a/{{ config.project_space }}/api/lookup_table/v1/{{
  record.lookup_table_id }}` - kind `update`; body type `json`; path fields `lookup_table_id`;
  required record fields `lookup_table_id`, `tag`, `fields`; accepted fields `fields`, `is_global`,
  `lookup_table_id`, `tag`; risk: updates a lookup table definition.
- `delete_lookup_table`: DELETE `/a/{{ config.project_space }}/api/lookup_table/v1/{{
  record.lookup_table_id }}` - kind `delete`; body type `none`; path fields `lookup_table_id`;
  required record fields `lookup_table_id`; accepted fields `lookup_table_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a lookup table
  definition.
- `create_lookup_table_row`: POST `/a/{{ config.project_space }}/api/lookup_table_item/v1/` - kind
  `create`; body type `json`; required record fields `data_type_id`, `fields`; accepted fields
  `data_type_id`, `fields`, `sort_key`; risk: creates a lookup table row.
- `update_lookup_table_row`: PUT `/a/{{ config.project_space }}/api/lookup_table_item/v1/{{
  record.lookup_table_item_id }}` - kind `update`; body type `json`; path fields
  `lookup_table_item_id`; required record fields `lookup_table_item_id`, `data_type_id`, `fields`;
  accepted fields `data_type_id`, `fields`, `lookup_table_item_id`, `sort_key`; risk: updates a
  lookup table row.
- `delete_lookup_table_row`: DELETE `/a/{{ config.project_space }}/api/lookup_table_item/v1/{{
  record.lookup_table_item_id }}` - kind `delete`; body type `none`; path fields
  `lookup_table_item_id`; required record fields `lookup_table_item_id`; accepted fields
  `lookup_table_item_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a lookup table row.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 38 stream-backed endpoint group(s), 25 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=5, out_of_scope=7.
