# pm connectors inspect when-i-work

```text
NAME
  pm connectors inspect when-i-work - When I Work connector manual

SYNOPSIS
  pm connectors inspect when-i-work
  pm connectors inspect when-i-work --json
  pm credentials add <name> --connector when-i-work [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes When I Work workforce-scheduling data: users, locations, positions, shifts, sites, shift templates, annotations, availability events, request types, time entries, timezones, payrolls, open-shift approval requests, and shift swaps.

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
  email (secret)
  password (secret)

ETL STREAMS
  users:
    primary key: id
    fields: email(), first_name(), id(), last_name()
  locations:
    primary key: id
    fields: address(), id(), name()
  positions:
    primary key: id
    fields: color(), id(), name()
  shifts:
    primary key: id
    fields: end_time(), id(), start_time(), user_id()
  sites:
    primary key: id
    fields: address(), id(), is_deleted(), location_id(), name()
  blocks:
    primary key: id
    fields: end_time(), id(), location_id(), position_id(), start_time()
  annotations:
    primary key: id
    fields: end_date(), id(), message(), start_date(), title()
  availabilityevents:
    primary key: id
    fields: end_time(), id(), start_time(), type(), user_id()
  requesttypes:
    primary key: id
    fields: built_in(), enabled(), id(), is_deleted(), name()
  times:
    primary key: id
    fields: end_time(), id(), is_approved(), shift_id(), start_time(), user_id()
  timezones:
    primary key: timezone_id
    fields: offset(), olson_id(), timezone_id(), timezone_name()
  payrolls:
    primary key: id
    fields: end_date(), id(), is_closed(), is_finalized(), start_date()
  openshiftapprovalrequests:
    primary key: id
    fields: created_at(), id(), shift_id(), status()
  swaps:
    primary key: id
    fields: id(), shift_id(), status(), type(), user_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_user:
    endpoint: POST /2/users
    risk: external mutation; creates a workforce-scheduling user account; approval required
  update_user:
    endpoint: PUT /2/users/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_user:
    endpoint: DELETE /2/users/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a workforce-scheduling user account; approval required
  create_location:
    endpoint: POST /2/locations
    risk: external mutation; approval required
  update_location:
    endpoint: PUT /2/locations/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_location:
    endpoint: DELETE /2/locations/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a schedule location; approval required
  create_position:
    endpoint: POST /2/positions
    risk: external mutation; approval required
  update_position:
    endpoint: PUT /2/positions/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_position:
    endpoint: DELETE /2/positions/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a position; approval required
  create_site:
    endpoint: POST /2/sites
    risk: external mutation; approval required
  update_site:
    endpoint: PUT /2/sites/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_site:
    endpoint: DELETE /2/sites/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a site; approval required
  create_block:
    endpoint: POST /2/blocks
    risk: external mutation; approval required
  update_block:
    endpoint: PUT /2/blocks/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_block:
    endpoint: DELETE /2/blocks/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a shift template; approval required
  create_annotation:
    endpoint: POST /2/annotations
    risk: external mutation; approval required
  update_annotation:
    endpoint: PUT /2/annotations/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_annotation:
    endpoint: DELETE /2/annotations/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a schedule annotation; approval required
  create_availability_event:
    endpoint: POST /2/availabilityevents
    risk: external mutation; writes a user's availability/unavailability preference; approval required
  update_availability_event:
    endpoint: PUT /2/availabilityevents/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_availability_event:
    endpoint: DELETE /2/availabilityevents/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a user availability event; approval required
  create_time:
    endpoint: POST /2/times
    risk: external mutation; creates a worked-time entry feeding payroll; approval required
  update_time:
    endpoint: PUT /2/times/{{ record.id }}
    required fields: id
    risk: external mutation; edits a worked-time entry feeding payroll; approval required
  delete_time:
    endpoint: DELETE /2/times/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a worked-time entry feeding payroll; approval required
  create_shift:
    endpoint: POST /2/shifts
    risk: external mutation; creates a scheduled shift; approval required
  delete_shift:
    endpoint: DELETE /2/shifts/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a scheduled shift; approval required

SECURITY
  read risk: external When I Work API read of the caller's own workforce-scheduling data
  write risk: external mutation of workforce-scheduling records (users, locations, positions, sites, shift templates, annotations, availability events, time entries feeding payroll, and shifts); create/update/delete all require approval, deletes are irreversible
  approval: read: none; write: required for every create/update/delete action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect when-i-work

  # Inspect as structured JSON
  pm connectors inspect when-i-work --json

AGENT WORKFLOW
  - Run pm connectors inspect when-i-work before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
