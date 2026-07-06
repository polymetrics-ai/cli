# Overview

Reads Pingdom checks, probes, actions, maintenance windows/occurrences, alerting contacts/teams,
credits, transaction checks, and reference data, and writes check/contact/team/maintenance mutations
through API 3.1.

Readable streams: `checks`, `probes`, `actions`, `maintenance`, `reference`, `alerting_contacts`,
`alerting_teams`, `maintenance_occurrences`, `credits`, `tms_checks`.

Write actions: `create_check`, `update_check`, `delete_check`, `create_contact`, `update_contact`,
`delete_contact`, `create_team`, `update_team`, `delete_team`, `create_maintenance`,
`delete_maintenance`.

Service API documentation: https://docs.pingdom.com/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Pingdom API 3.1 Bearer token. Sent as Authorization: Bearer
  <api_key>. Never logged.
- `base_url` (optional, string); default `https://api.pingdom.com/api/3.1`; format `uri`; Pingdom
  API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.pingdom.com/api/3.1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/checks`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100.

Pagination by stream: none: `alerting_contacts`, `alerting_teams`, `maintenance_occurrences`,
`credits`; offset_limit: `checks`, `probes`, `actions`, `maintenance`, `reference`, `tms_checks`.

- `checks`: GET `/checks` - records path `checks`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; emits passthrough records.
- `probes`: GET `/probes` - records path `probes`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; emits passthrough records.
- `actions`: GET `/actions` - records path `actions`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; emits passthrough records.
- `maintenance`: GET `/maintenance` - records path `maintenance`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; emits passthrough records.
- `reference`: GET `/reference` - single-object response; records path `.`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; emits passthrough records.
- `alerting_contacts`: GET `/alerting/contacts` - records path `contacts`; emits passthrough
  records.
- `alerting_teams`: GET `/alerting/teams` - records path `teams`; emits passthrough records.
- `maintenance_occurrences`: GET `/maintenance.occurrences` - records path `occurrences`; emits
  passthrough records.
- `credits`: GET `/credits` - single-object response; records path `credits`; emits passthrough
  records.
- `tms_checks`: GET `/tms/check` - records path `checks`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; emits passthrough records.

## Write actions & risks

Overall write risk: creates/updates/deletes uptime checks, alerting contacts and teams, and
maintenance windows.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_check`: POST `/checks` - kind `create`; body type `json`; required record fields `name`,
  `host`, `type`; accepted fields `host`, `ipv6`, `name`, `notifyagainevery`, `notifywhenbackup`,
  `paused`, `resolution`, `sendnotificationwhendown`, `tags`, `type`; risk: creates a new Pingdom
  uptime check (this action models the common HTTP-type check shape; Pingdom's other 8 check types
  share the same name/host/type/paused/resolution/notification fields plus type-specific attributes
  not modeled here, see docs.md Known limits); low-risk external mutation, no approval required.
- `update_check`: PUT `/checks/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `host`, `id`, `name`, `paused`, `resolution`, `tags`;
  risk: updates an existing check's settings (name/host/paused/resolution/tags); external mutation,
  approval required.
- `delete_check`: DELETE `/checks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  permanently deletes an uptime check and its historical results; destructive external mutation,
  approval required.
- `create_contact`: POST `/alerting/contacts` - kind `create`; body type `json`; required record
  fields `name`, `notification_targets`; accepted fields `name`, `notification_targets`, `paused`;
  risk: creates a new alerting contact with email/SMS notification targets; low-risk external
  mutation, no approval required.
- `update_contact`: PUT `/alerting/contacts/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `name`, `paused`, `notification_targets`; accepted
  fields `id`, `name`, `notification_targets`, `paused`; risk: updates an existing alerting
  contact's name/paused state/notification targets (Pingdom's PUT is a full replacement, requiring
  name/paused/notification_targets together); external mutation, approval required.
- `delete_contact`: DELETE `/alerting/contacts/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`;
  risk: permanently deletes an alerting contact and its notification targets; destructive external
  mutation, approval required.
- `create_team`: POST `/alerting/teams` - kind `create`; body type `json`; required record fields
  `name`, `member_ids`; accepted fields `member_ids`, `name`; risk: creates a new alerting team from
  a list of contact ids; low-risk external mutation, no approval required.
- `update_team`: PUT `/alerting/teams/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `name`, `member_ids`; accepted fields `id`,
  `member_ids`, `name`; risk: updates an existing alerting team's name/member list; external
  mutation, approval required.
- `delete_team`: DELETE `/alerting/teams/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`; risk:
  permanently deletes an alerting team; destructive external mutation, approval required.
- `create_maintenance`: POST `/maintenance` - kind `create`; body type `json`; required record
  fields `description`, `from`, `to`; accepted fields `description`, `effectiveto`, `from`,
  `recurrencetype`, `repeatevery`, `tmsids`, `to`, `uptimeids`; risk: creates a new maintenance
  window that suppresses alerting for the assigned checks during the scheduled period; low-risk
  external mutation, no approval required.
- `delete_maintenance`: DELETE `/maintenance/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; confirmation `destructive`;
  risk: permanently deletes a maintenance window, immediately resuming alerting for its assigned
  checks; destructive external mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 11 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=3, duplicate_of=6, non_data_endpoint=13, out_of_scope=7.
