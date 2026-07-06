# Overview

Reads Deputy locations, employees, departments, timesheets, tasks, leave, rosters, webhooks, and
teams, and writes department/leave/roster/webhook/team mutations, through the Deputy REST API (full
refresh).

Readable streams: `locations`, `employees`, `departments`, `timesheets`, `tasks`, `leave`,
`rosters`, `webhooks`, `teams`.

Write actions: `create_department`, `update_department`, `delete_department`, `create_leave`,
`update_leave`, `delete_leave`, `create_roster`, `update_roster`, `delete_roster`, `create_webhook`,
`update_webhook`, `delete_webhook`, `create_team`, `update_team`, `delete_team`.

Service API documentation: https://developer.deputy.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Deputy bearer access token, sent as the Authorization
  header. Never logged.
- `base_url` (required, string); format `uri`; Deputy install-specific API base URL, e.g.
  https://{installname}.{geo}.deputy.com.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v1/resource/Company` with query `max`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `start`; limit parameter `max`; page
size 500.

Pagination by stream: none: `employees`, `timesheets`, `tasks`; offset_limit: `locations`,
`departments`, `leave`, `rosters`, `webhooks`, `teams`.

- `locations`: GET `/api/v1/resource/Company` - records at response root; offset/limit pagination;
  offset parameter `start`; limit parameter `max`; page size 500; computed output fields `active`,
  `address`, `code`, `company_name`, `country`, `created`, `creator`, `id`, `modified`.
- `employees`: GET `/api/v1/supervise/employee` - records at response root; computed output fields
  `active`, `company`, `created`, `display_name`, `first_name`, `id`, `last_name`, `modified`,
  `role`.
- `departments`: GET `/api/v1/resource/OperationalUnit` - records at response root; offset/limit
  pagination; offset parameter `start`; limit parameter `max`; page size 500; computed output fields
  `active`, `company`, `created`, `creator`, `id`, `modified`, `operational_unit_name`.
- `timesheets`: GET `/api/v1/my/timesheets` - records at response root; computed output fields
  `created`, `date`, `employee`, `end_time`, `id`, `is_in_progress`, `modified`, `operational_unit`,
  `start_time`, `total_time`.
- `tasks`: GET `/api/v1/my/tasks` - records at response root; computed output fields `completed`,
  `created`, `creator`, `due_time`, `id`, `modified`, `priority`, `title`.
- `leave`: GET `/api/v1/resource/Leave` - records at response root; offset/limit pagination; offset
  parameter `start`; limit parameter `max`; page size 500; computed output fields `all_day`,
  `comment`, `created`, `creator`, `date_end`, `date_start`, `days`, `employee`, `id`, `leave_rule`,
  `modified`, `status`.
- `rosters`: GET `/api/v1/resource/Roster` - records at response root; offset/limit pagination;
  offset parameter `start`; limit parameter `max`; page size 500; computed output fields `cost`,
  `created`, `creator`, `date`, `employee`, `end_time`, `id`, `modified`, `open`,
  `operational_unit`, `published`, `start_time`, `total_time`.
- `webhooks`: GET `/api/v1/resource/Webhook` - records at response root; offset/limit pagination;
  offset parameter `start`; limit parameter `max`; page size 500; computed output fields `address`,
  `created`, `creator`, `enabled`, `id`, `modified`, `topic`, `type`.
- `teams`: GET `/api/v1/resource/Team` - records at response root; offset/limit pagination; offset
  parameter `start`; limit parameter `max`; page size 500; computed output fields `created`,
  `creator`, `id`, `leader_employee`, `modified`, `name`.

## Write actions & risks

Overall write risk: external mutation of departments, leave requests (approval status),
rosters/shifts (may notify employees), webhook subscriptions, and teams; approval required.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_department`: POST `/api/v1/resource/OperationalUnit` - kind `create`; body type `json`;
  required record fields `Company`, `OperationalUnitName`; accepted fields `Active`, `Colour`,
  `Company`, `OperationalUnitName`, `ParentOperationalUnit`, `ShowOnRoster`; risk: external
  mutation; creates a real Deputy department/operational unit; approval required.
- `update_department`: POST `/api/v1/resource/OperationalUnit/{{ record.Id }}` - kind `update`; body
  type `json`; path fields `Id`; required record fields `Id`; accepted fields `Active`, `Colour`,
  `Id`, `OperationalUnitName`, `ShowOnRoster`; risk: external mutation; updates a real Deputy
  department/operational unit; approval required.
- `delete_department`: DELETE `/api/v1/resource/OperationalUnit/{{ record.Id }}` - kind `delete`;
  body type `none`; path fields `Id`; required record fields `Id`; accepted fields `Id`; missing
  records treated as success for status `404`; risk: irreversible deletion of a real Deputy
  department/operational unit; approval required.
- `create_leave`: POST `/api/v1/resource/Leave` - kind `create`; body type `json`; required record
  fields `Employee`, `DateStart`, `DateEnd`; accepted fields `AllDay`, `Comment`, `DateEnd`,
  `DateStart`, `Employee`, `LeaveRule`; risk: external mutation; creates a real leave request for a
  Deputy employee; approval required.
- `update_leave`: POST `/api/v1/resource/Leave/{{ record.Id }}` - kind `update`; body type `json`;
  path fields `Id`; required record fields `Id`; accepted fields `Comment`, `DateEnd`, `DateStart`,
  `Id`, `Status`; risk: external mutation; updates a real Deputy leave request, including its
  approval status; approval required.
- `delete_leave`: DELETE `/api/v1/resource/Leave/{{ record.Id }}` - kind `delete`; body type `none`;
  path fields `Id`; required record fields `Id`; accepted fields `Id`; missing records treated as
  success for status `404`; risk: irreversible deletion of a real Deputy leave request; approval
  required.
- `create_roster`: POST `/api/v1/resource/Roster` - kind `create`; body type `json`; required record
  fields `StartTime`, `EndTime`, `OperationalUnit`; accepted fields `Comment`, `Employee`,
  `EndTime`, `Open`, `OperationalUnit`, `StartTime`; risk: external mutation; creates a real Deputy
  roster/shift, potentially notifying the assigned employee; approval required.
- `update_roster`: POST `/api/v1/resource/Roster/{{ record.Id }}` - kind `update`; body type `json`;
  path fields `Id`; required record fields `Id`; accepted fields `Comment`, `Employee`, `EndTime`,
  `Id`, `Published`, `StartTime`; risk: external mutation; updates a real Deputy roster/shift,
  potentially notifying the assigned employee; approval required.
- `delete_roster`: DELETE `/api/v1/resource/Roster/{{ record.Id }}` - kind `delete`; body type
  `none`; path fields `Id`; required record fields `Id`; accepted fields `Id`; missing records
  treated as success for status `404`; risk: irreversible deletion of a real Deputy roster/shift;
  approval required.
- `create_webhook`: POST `/api/v1/resource/Webhook` - kind `create`; body type `json`; required
  record fields `Topic`, `Address`, `Type`; accepted fields `Address`, `Enabled`, `Topic`, `Type`;
  risk: external mutation; registers a real Deputy webhook subscription that will deliver events to
  the given address; approval required.
- `update_webhook`: POST `/api/v1/resource/Webhook/{{ record.Id }}` - kind `update`; body type
  `json`; path fields `Id`; required record fields `Id`; accepted fields `Address`, `Enabled`, `Id`;
  risk: external mutation; updates a real Deputy webhook subscription; approval required.
- `delete_webhook`: DELETE `/api/v1/resource/Webhook/{{ record.Id }}` - kind `delete`; body type
  `none`; path fields `Id`; required record fields `Id`; accepted fields `Id`; missing records
  treated as success for status `404`; risk: irreversible deletion of a real Deputy webhook
  subscription; approval required.
- `create_team`: POST `/api/v1/resource/Team` - kind `create`; body type `json`; required record
  fields `Name`; accepted fields `LeaderEmployee`, `Name`; risk: external mutation; creates a real
  Deputy team; approval required.
- `update_team`: POST `/api/v1/resource/Team/{{ record.Id }}` - kind `update`; body type `json`;
  path fields `Id`; required record fields `Id`; accepted fields `Id`, `LeaderEmployee`, `Name`;
  risk: external mutation; updates a real Deputy team; approval required.
- `delete_team`: DELETE `/api/v1/resource/Team/{{ record.Id }}` - kind `delete`; body type `none`;
  path fields `Id`; required record fields `Id`; accepted fields `Id`; missing records treated as
  success for status `404`; risk: irreversible deletion of a real Deputy team; approval required.

## Known limits

- Batch defaults: read_page_size=500.
- API coverage includes 9 stream-backed endpoint group(s), 15 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, destructive_admin=3, duplicate_of=11, non_data_endpoint=5, out_of_scope=14,
  requires_elevated_scope=6.
