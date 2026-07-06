# Overview

Reads and writes When I Work workforce-scheduling data: users, locations, positions, shifts, sites,
shift templates, annotations, availability events, request types, time entries, timezones, payrolls,
open-shift approval requests, and shift swaps.

Readable streams: `users`, `locations`, `positions`, `shifts`, `sites`, `blocks`, `annotations`,
`availabilityevents`, `requesttypes`, `times`, `timezones`, `payrolls`, `openshiftapprovalrequests`,
`swaps`.

Write actions: `create_user`, `update_user`, `delete_user`, `create_location`, `update_location`,
`delete_location`, `create_position`, `update_position`, `delete_position`, `create_site`,
`update_site`, `delete_site`, `create_block`, `update_block`, `delete_block`, `create_annotation`,
`update_annotation`, `delete_annotation`, `create_availability_event`, `update_availability_event`,
`delete_availability_event`, `create_time`, `update_time`, `delete_time`, `create_shift`,
`delete_shift`.

Service API documentation: https://apidocs.wheniwork.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.wheniwork.com`; format `uri`; When I Work API
  base URL override for tests or proxies.
- `email` (required, secret, string); When I Work account email, sent as the Basic auth username on
  every request; never logged.
- `mode` (optional, string).
- `password` (required, secret, string); When I Work account password, sent as the Basic auth
  password on every request; never logged.

Secret fields are redacted in logs and write previews: `email`, `password`.

Default configuration values: `base_url=https://api.wheniwork.com`.

Authentication behavior:

- HTTP Basic authentication using `secrets.email`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/2/users`.

## Streams notes

Default pagination: single request; no pagination.

- `users`: GET `/2/users` - records path `users`; emits passthrough records.
- `locations`: GET `/2/locations` - records path `locations`; emits passthrough records.
- `positions`: GET `/2/positions` - records path `positions`; emits passthrough records.
- `shifts`: GET `/2/shifts` - records path `shifts`; emits passthrough records.
- `sites`: GET `/2/sites` - records path `sites`; emits passthrough records.
- `blocks`: GET `/2/blocks` - records path `blocks`; emits passthrough records.
- `annotations`: GET `/2/annotations` - records path `annotations`; emits passthrough records.
- `availabilityevents`: GET `/2/availabilityevents` - records path `availabilityevents`; emits
  passthrough records.
- `requesttypes`: GET `/2/requesttypes` - records path `request-types`; emits passthrough records.
- `times`: GET `/2/times` - records path `times`; emits passthrough records.
- `timezones`: GET `/2/timezones` - records path `timezone`; emits passthrough records.
- `payrolls`: GET `/2/payrolls` - records path `payrolls`; emits passthrough records.
- `openshiftapprovalrequests`: GET `/2/openshiftapprovalrequests` - records path
  `openshiftapprovalrequests`; emits passthrough records.
- `swaps`: GET `/2/swaps` - records path `swaps`; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation of workforce-scheduling records (users, locations, positions,
sites, shift templates, annotations, availability events, time entries feeding payroll, and shifts);
create/update/delete all require approval, deletes are irreversible.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user`: POST `/2/users` - kind `create`; body type `json`; accepted fields `email`,
  `employee_code`, `first_name`, `hourly_rate`, `last_name`, `phone_number`, `role`; risk: external
  mutation; creates a workforce-scheduling user account; approval required.
- `update_user`: PUT `/2/users/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `email`, `first_name`, `hourly_rate`, `id`,
  `last_name`, `phone_number`; risk: external mutation; approval required.
- `delete_user`: DELETE `/2/users/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: irreversible external deletion of a
  workforce-scheduling user account; approval required.
- `create_location`: POST `/2/locations` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `address`, `is_default`, `max_hours`, `name`, `radius`; risk: external
  mutation; approval required.
- `update_location`: PUT `/2/locations/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `address`, `id`, `max_hours`, `name`;
  risk: external mutation; approval required.
- `delete_location`: DELETE `/2/locations/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: irreversible external deletion of a schedule
  location; approval required.
- `create_position`: POST `/2/positions` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `color`, `name`, `sort`, `tips_tracking`; risk: external mutation;
  approval required.
- `update_position`: PUT `/2/positions/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `color`, `id`, `name`, `sort`; risk:
  external mutation; approval required.
- `delete_position`: DELETE `/2/positions/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: irreversible external deletion of a position;
  approval required.
- `create_site`: POST `/2/sites` - kind `create`; body type `json`; required record fields
  `location_id`, `name`; accepted fields `address`, `color`, `description`, `location_id`, `name`;
  risk: external mutation; approval required.
- `update_site`: PUT `/2/sites/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `address`, `description`, `id`, `name`; risk:
  external mutation; approval required.
- `delete_site`: DELETE `/2/sites/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: irreversible external deletion of a site; approval
  required.
- `create_block`: POST `/2/blocks` - kind `create`; body type `json`; required record fields
  `start_time`, `end_time`, `location_id`; accepted fields `break_time`, `color`, `end_time`,
  `location_id`, `notes`, `position_id`, `start_time`; risk: external mutation; approval required.
- `update_block`: PUT `/2/blocks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `end_time`, `id`, `notes`, `start_time`; risk:
  external mutation; approval required.
- `delete_block`: DELETE `/2/blocks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: irreversible external deletion of a shift
  template; approval required.
- `create_annotation`: POST `/2/annotations` - kind `create`; body type `json`; required record
  fields `start_date`, `end_date`, `title`; accepted fields `all_locations`, `announcement`,
  `business_closed`, `color`, `end_date`, `message`, `start_date`, `title`; risk: external mutation;
  approval required.
- `update_annotation`: PUT `/2/annotations/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `end_date`, `id`, `message`,
  `start_date`, `title`; risk: external mutation; approval required.
- `delete_annotation`: DELETE `/2/annotations/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: irreversible external deletion of a
  schedule annotation; approval required.
- `create_availability_event`: POST `/2/availabilityevents` - kind `create`; body type `json`;
  required record fields `start_time`, `type`; accepted fields `all_day`, `end_time`, `notes`,
  `start_time`, `type`, `user_id`; risk: external mutation; writes a user's
  availability/unavailability preference; approval required.
- `update_availability_event`: PUT `/2/availabilityevents/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `end_time`, `id`,
  `notes`, `start_time`; risk: external mutation; approval required.
- `delete_availability_event`: DELETE `/2/availabilityevents/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: irreversible external
  deletion of a user availability event; approval required.
- `create_time`: POST `/2/times` - kind `create`; body type `json`; required record fields
  `user_id`, `start_time`, `end_time`; accepted fields `end_time`, `location_id`, `notes`,
  `position_id`, `shift_id`, `site_id`, `start_time`, `user_id`; risk: external mutation; creates a
  worked-time entry feeding payroll; approval required.
- `update_time`: PUT `/2/times/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `end_time`, `id`, `is_approved`, `notes`,
  `start_time`; risk: external mutation; edits a worked-time entry feeding payroll; approval
  required.
- `delete_time`: DELETE `/2/times/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: irreversible external deletion of a worked-time
  entry feeding payroll; approval required.
- `create_shift`: POST `/2/shifts` - kind `create`; body type `json`; required record fields
  `start_time`, `end_time`, `location_id`; accepted fields `end_time`, `location_id`, `notes`,
  `position_id`, `published`, `site_id`, `start_time`, `user_id`; risk: external mutation; creates a
  scheduled shift; approval required.
- `delete_shift`: DELETE `/2/shifts/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: irreversible external deletion of a scheduled
  shift; approval required.

## Known limits

- API coverage includes 14 stream-backed endpoint group(s), 26 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=3, duplicate_of=26, non_data_endpoint=14, out_of_scope=36.
