---
name: pm-deputy
description: Deputy connector knowledge and safe action guide.
---

# pm-deputy

## Purpose

Reads Deputy locations, employees, departments, timesheets, tasks, leave, rosters, webhooks, and teams, and writes department/leave/roster/webhook/team mutations, through the Deputy REST API (full refresh).

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- mode
- api_key (secret) (required)

## ETL Streams

- locations:
  - primary key: id
  - fields: active(boolean), address(integer), code(string), company_name(string), country(integer), created(string), creator(integer), id(integer), modified(string)
- employees:
  - primary key: id
  - fields: active(boolean), company(integer), created(string), display_name(string), first_name(string), id(integer), last_name(string), modified(string), role(integer)
- departments:
  - primary key: id
  - fields: active(boolean), company(integer), created(string), creator(integer), id(integer), modified(string), operational_unit_name(string)
- timesheets:
  - primary key: id
  - fields: created(string), date(string), employee(integer), end_time(integer), id(integer), is_in_progress(boolean), modified(string), operational_unit(integer), start_time(integer), total_time(number)
- tasks:
  - primary key: id
  - fields: completed(boolean), created(string), creator(integer), due_time(string), id(integer), modified(string), priority(integer), title(string)
- leave:
  - primary key: id
  - fields: all_day(boolean), comment(string), created(string), creator(integer), date_end(string), date_start(string), days(number), employee(integer), id(integer), leave_rule(integer), modified(string), status(integer)
- rosters:
  - primary key: id
  - fields: cost(number), created(string), creator(integer), date(string), employee(integer), end_time(integer), id(integer), modified(string), open(boolean), operational_unit(integer), published(boolean), start_time(integer), total_time(number)
- webhooks:
  - primary key: id
  - fields: address(string), created(string), creator(integer), enabled(boolean), id(integer), modified(string), topic(string), type(string)
- teams:
  - primary key: id
  - fields: created(string), creator(integer), id(integer), leader_employee(integer), modified(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_department:
  - endpoint: POST /api/v1/resource/OperationalUnit
  - required fields: Company, OperationalUnitName
  - risk: external mutation; creates a real Deputy department/operational unit; approval required
- update_department:
  - endpoint: POST /api/v1/resource/OperationalUnit/{{ record.Id }}
  - required fields: Id
  - risk: external mutation; updates a real Deputy department/operational unit; approval required
- delete_department:
  - endpoint: DELETE /api/v1/resource/OperationalUnit/{{ record.Id }}
  - required fields: Id
  - risk: irreversible deletion of a real Deputy department/operational unit; approval required
- create_leave:
  - endpoint: POST /api/v1/resource/Leave
  - required fields: Employee, DateStart, DateEnd
  - risk: external mutation; creates a real leave request for a Deputy employee; approval required
- update_leave:
  - endpoint: POST /api/v1/resource/Leave/{{ record.Id }}
  - required fields: Id
  - risk: external mutation; updates a real Deputy leave request, including its approval status; approval required
- delete_leave:
  - endpoint: DELETE /api/v1/resource/Leave/{{ record.Id }}
  - required fields: Id
  - risk: irreversible deletion of a real Deputy leave request; approval required
- create_roster:
  - endpoint: POST /api/v1/resource/Roster
  - required fields: StartTime, EndTime, OperationalUnit
  - risk: external mutation; creates a real Deputy roster/shift, potentially notifying the assigned employee; approval required
- update_roster:
  - endpoint: POST /api/v1/resource/Roster/{{ record.Id }}
  - required fields: Id
  - risk: external mutation; updates a real Deputy roster/shift, potentially notifying the assigned employee; approval required
- delete_roster:
  - endpoint: DELETE /api/v1/resource/Roster/{{ record.Id }}
  - required fields: Id
  - risk: irreversible deletion of a real Deputy roster/shift; approval required
- create_webhook:
  - endpoint: POST /api/v1/resource/Webhook
  - required fields: Topic, Address, Type
  - risk: external mutation; registers a real Deputy webhook subscription that will deliver events to the given address; approval required
- update_webhook:
  - endpoint: POST /api/v1/resource/Webhook/{{ record.Id }}
  - required fields: Id
  - risk: external mutation; updates a real Deputy webhook subscription; approval required
- delete_webhook:
  - endpoint: DELETE /api/v1/resource/Webhook/{{ record.Id }}
  - required fields: Id
  - risk: irreversible deletion of a real Deputy webhook subscription; approval required
- create_team:
  - endpoint: POST /api/v1/resource/Team
  - required fields: Name
  - risk: external mutation; creates a real Deputy team; approval required
- update_team:
  - endpoint: POST /api/v1/resource/Team/{{ record.Id }}
  - required fields: Id
  - risk: external mutation; updates a real Deputy team; approval required
- delete_team:
  - endpoint: DELETE /api/v1/resource/Team/{{ record.Id }}
  - required fields: Id
  - risk: irreversible deletion of a real Deputy team; approval required

## Security

- read risk: external Deputy API read of workforce scheduling, employee, timesheet, leave, and roster data
- write risk: external mutation of departments, leave requests (approval status), rosters/shifts (may notify employees), webhook subscriptions, and teams; approval required
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect deputy
```

### Inspect as structured JSON

```bash
pm connectors inspect deputy --json
```

## Agent Rules

- Run pm connectors inspect deputy before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
