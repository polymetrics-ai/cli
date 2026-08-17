# pm connectors inspect deputy

```text
NAME
  pm connectors inspect deputy - Deputy connector manual

SYNOPSIS
  pm connectors inspect deputy
  pm connectors inspect deputy --json
  pm credentials add <name> --connector deputy [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Deputy locations, employees, departments, timesheets, tasks, leave, rosters, webhooks, and teams, and writes department/leave/roster/webhook/team mutations, through the Deputy REST API (full refresh).

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  locations:
    primary key: id
    fields: active(), address(), code(), company_name(), country(), created(), creator(), id(), modified()
  employees:
    primary key: id
    fields: active(), company(), created(), display_name(), first_name(), id(), last_name(), modified(), role()
  departments:
    primary key: id
    fields: active(), company(), created(), creator(), id(), modified(), operational_unit_name()
  timesheets:
    primary key: id
    fields: created(), date(), employee(), end_time(), id(), is_in_progress(), modified(), operational_unit(), start_time(), total_time()
  tasks:
    primary key: id
    fields: completed(), created(), creator(), due_time(), id(), modified(), priority(), title()
  leave:
    primary key: id
    fields: all_day(), comment(), created(), creator(), date_end(), date_start(), days(), employee(), id(), leave_rule(), modified(), status()
  rosters:
    primary key: id
    fields: cost(), created(), creator(), date(), employee(), end_time(), id(), modified(), open(), operational_unit(), published(), start_time(), total_time()
  webhooks:
    primary key: id
    fields: address(), created(), creator(), enabled(), id(), modified(), topic(), type()
  teams:
    primary key: id
    fields: created(), creator(), id(), leader_employee(), modified(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_department:
    endpoint: POST /api/v1/resource/OperationalUnit
    risk: external mutation; creates a real Deputy department/operational unit; approval required
  update_department:
    endpoint: POST /api/v1/resource/OperationalUnit/{{ record.Id }}
    required fields: Id
    risk: external mutation; updates a real Deputy department/operational unit; approval required
  delete_department:
    endpoint: DELETE /api/v1/resource/OperationalUnit/{{ record.Id }}
    required fields: Id
    risk: irreversible deletion of a real Deputy department/operational unit; approval required
  create_leave:
    endpoint: POST /api/v1/resource/Leave
    risk: external mutation; creates a real leave request for a Deputy employee; approval required
  update_leave:
    endpoint: POST /api/v1/resource/Leave/{{ record.Id }}
    required fields: Id
    risk: external mutation; updates a real Deputy leave request, including its approval status; approval required
  delete_leave:
    endpoint: DELETE /api/v1/resource/Leave/{{ record.Id }}
    required fields: Id
    risk: irreversible deletion of a real Deputy leave request; approval required
  create_roster:
    endpoint: POST /api/v1/resource/Roster
    risk: external mutation; creates a real Deputy roster/shift, potentially notifying the assigned employee; approval required
  update_roster:
    endpoint: POST /api/v1/resource/Roster/{{ record.Id }}
    required fields: Id
    risk: external mutation; updates a real Deputy roster/shift, potentially notifying the assigned employee; approval required
  delete_roster:
    endpoint: DELETE /api/v1/resource/Roster/{{ record.Id }}
    required fields: Id
    risk: irreversible deletion of a real Deputy roster/shift; approval required
  create_webhook:
    endpoint: POST /api/v1/resource/Webhook
    risk: external mutation; registers a real Deputy webhook subscription that will deliver events to the given address; approval required
  update_webhook:
    endpoint: POST /api/v1/resource/Webhook/{{ record.Id }}
    required fields: Id
    risk: external mutation; updates a real Deputy webhook subscription; approval required
  delete_webhook:
    endpoint: DELETE /api/v1/resource/Webhook/{{ record.Id }}
    required fields: Id
    risk: irreversible deletion of a real Deputy webhook subscription; approval required
  create_team:
    endpoint: POST /api/v1/resource/Team
    risk: external mutation; creates a real Deputy team; approval required
  update_team:
    endpoint: POST /api/v1/resource/Team/{{ record.Id }}
    required fields: Id
    risk: external mutation; updates a real Deputy team; approval required
  delete_team:
    endpoint: DELETE /api/v1/resource/Team/{{ record.Id }}
    required fields: Id
    risk: irreversible deletion of a real Deputy team; approval required

SECURITY
  read risk: external Deputy API read of workforce scheduling, employee, timesheet, leave, and roster data
  write risk: external mutation of departments, leave requests (approval status), rosters/shifts (may notify employees), webhook subscriptions, and teams; approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect deputy

  # Inspect as structured JSON
  pm connectors inspect deputy --json

AGENT WORKFLOW
  - Run pm connectors inspect deputy before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
