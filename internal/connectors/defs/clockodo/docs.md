# Overview

Reads Clockodo customers, projects, services, users, time entries, absences, teams, surcharges,
lump-sum services, nonbusiness groups/days, holiday/overtime carryovers, target hours, and
current-user settings, and writes customers/projects/services/teams/lump-sum services through the
Clockodo REST API.

Readable streams: `customers`, `projects`, `services`, `users`, `current_user_settings`, `teams`,
`surcharges`, `lumpsum_services`, `nonbusiness_groups`, `nonbusiness_days`, `holidays_carry`,
`holidays_quota`, `overtime_carry`, `target_hours`, `absences`, `entries`.

Write actions: `create_customer`, `update_customer`, `delete_customer`, `create_project`,
`update_project`, `delete_project`, `create_service`, `update_service`, `delete_service`,
`create_team`, `update_team`, `delete_team`, `create_lumpsum_service`, `update_lumpsum_service`,
`delete_lumpsum_service`.

Service API documentation: https://www.clockodo.com/en/api/.

## Auth setup

Connection fields:

- `absences_year` (optional, string); 4-digit year (e.g. "2026") to query for the absences stream;
  required only for that stream, which Clockodo's own API requires a year query param for.
- `api_key` (required, secret, string); Clockodo API key, sent as the X-ClockodoApiKey header. Used
  only for auth; never logged.
- `base_url` (optional, string); default `https://my.clockodo.com/api`; format `uri`; Clockodo API
  base URL override for tests or proxies.
- `email_address` (required, string); Clockodo account email address, sent as the X-ClockodoApiUser
  header.
- `entries_time_since` (optional, string); ISO 8601 UTC start of the time-entries window (e.g.
  "2026-01-01T00:00:00Z"); required only for the entries stream, which Clockodo's own API requires a
  bounded time_since/time_until range for.
- `entries_time_until` (optional, string); ISO 8601 UTC end of the time-entries window (e.g.
  "2026-12-31T23:59:59Z"); required only for the entries stream, paired with entries_time_since.
- `external_application` (required, string); Application;contact identifier required by Clockodo,
  sent as the X-Clockodo-External-Application header (e.g. "polymetrics;ops@example.com").
- `language` (optional, string); Optional Accept-Language header override (e.g. "en").
- `nonbusinessdays_year` (optional, string); 4-digit year (e.g. "2026") to query for the
  nonbusiness_days stream; required only for that stream, which Clockodo's own API requires a year
  query param for.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://my.clockodo.com/api`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/users`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; no page-size parameter; starts at
1; page size 50.

Pagination by stream: none: `services`, `users`, `current_user_settings`, `teams`, `surcharges`,
`lumpsum_services`, `nonbusiness_groups`, `nonbusiness_days`, `holidays_carry`, `holidays_quota`,
`overtime_carry`, `target_hours`, `absences`; page_number: `customers`, `projects`, `entries`.

- `customers`: GET `/v2/customers` - records path `customers`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 50.
- `projects`: GET `/v2/projects` - records path `projects`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 50.
- `services`: GET `/v2/services` - records path `services`.
- `users`: GET `/v2/users` - records path `users`.
- `current_user_settings`: GET `/v2/aggregates/users/me` - records path `.`.
- `teams`: GET `/v2/teams` - records path `team`.
- `surcharges`: GET `/v2/surcharges` - records path `surcharges`.
- `lumpsum_services`: GET `/v2/lumpsumservices` - records path `lumpSumServices`.
- `nonbusiness_groups`: GET `/nonbusinessgroups` - records path `nonbusinessgroups`.
- `nonbusiness_days`: GET `/nonbusinessdays` - records path `nonbusinessdays`; query `year`=`{{
  config.nonbusinessdays_year }}`.
- `holidays_carry`: GET `/holidayscarry` - records path `holidayscarry`.
- `holidays_quota`: GET `/holidaysquota` - records path `holidaysquota`.
- `overtime_carry`: GET `/overtimecarry` - records path `overtimecarry`.
- `target_hours`: GET `/targethours` - records path `targethours`.
- `absences`: GET `/absences` - records path `absences`; query `year`=`{{ config.absences_year }}`.
- `entries`: GET `/v2/entries` - records path `entries`; query `time_since`=`{{
  config.entries_time_since }}`; `time_until`=`{{ config.entries_time_until }}`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 50.

## Write actions & risks

Overall write risk: external mutation; creates/updates/deletes live Clockodo customers, projects,
services, teams, and lump-sum services.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_customer`: POST `/v2/customers` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `active`, `billable_default`, `name`, `note`, `number`; risk: external
  mutation; creates a live Clockodo customer; approval required.
- `update_customer`: PUT `/v2/customers/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `active`, `billable_default`, `id`,
  `name`, `note`, `number`; risk: external mutation; overwrites a live Clockodo customer's fields;
  approval required.
- `delete_customer`: DELETE `/v2/customers/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live Clockodo customer; approval required.
- `create_project`: POST `/v2/projects` - kind `create`; body type `json`; required record fields
  `name`, `customers_id`; accepted fields `active`, `billable_default`, `budget_is_hours`,
  `budget_money`, `customers_id`, `name`, `note`, `number`; risk: external mutation; creates a live
  Clockodo project; approval required.
- `update_project`: PUT `/v2/projects/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `active`, `billable_default`,
  `budget_money`, `id`, `name`, `note`; risk: external mutation; overwrites a live Clockodo
  project's fields; approval required.
- `delete_project`: DELETE `/v2/projects/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly removes a live Clockodo project; approval required.
- `create_service`: POST `/v2/services` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `active`, `name`, `note`, `number`; risk: external mutation; creates a
  live Clockodo service; approval required.
- `update_service`: PUT `/v2/services/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `active`, `id`, `name`, `note`; risk:
  external mutation; overwrites a live Clockodo service's fields; approval required.
- `delete_service`: DELETE `/v2/services/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  irreversibly deletes a live Clockodo service; approval required.
- `create_team`: POST `/v2/teams` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `leader`, `name`; risk: external mutation; creates a live Clockodo team; approval
  required.
- `update_team`: PUT `/v2/teams/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `id`, `leader`, `name`; risk: external
  mutation; overwrites a live Clockodo team's fields; approval required.
- `delete_team`: DELETE `/v2/teams/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; irreversibly
  deletes a live Clockodo team; approval required.
- `create_lumpsum_service`: POST `/v2/lumpsumservices` - kind `create`; body type `json`; required
  record fields `name`, `price`; accepted fields `active`, `name`, `note`, `number`, `price`,
  `unit`; risk: external mutation; creates a live Clockodo lump-sum service; approval required.
- `update_lumpsum_service`: PUT `/v2/lumpsumservices/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `active`, `id`, `name`,
  `note`, `price`; risk: external mutation; overwrites a live Clockodo lump-sum service's fields;
  approval required.
- `delete_lumpsum_service`: DELETE `/v2/lumpsumservices/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  mutation; irreversibly deletes a live Clockodo lump-sum service; approval required.

## Known limits

- API coverage includes 16 stream-backed endpoint group(s), 15 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=4, duplicate_of=13, non_data_endpoint=2, out_of_scope=31,
  requires_elevated_scope=2.
