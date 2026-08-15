---
name: pm-when-i-work
description: When I Work connector knowledge and safe action guide.
---

# pm-when-i-work

## Purpose

Reads and writes When I Work workforce-scheduling data: users, locations, positions, shifts, sites, shift templates, annotations, availability events, request types, time entries, timezones, payrolls, open-shift approval requests, and shift swaps.

## Icon

- id: simple-icons-wheniwork
- asset: icons/simple-icons/wheniwork.svg
- title: When I Work
- simple_icon_slug: wheniwork
- simple_icon_hex: 51A33D
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=When%20I%20Work
- match: exact-name-or-slug
- matched_by: when-i-work

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- email (secret) (required)
- password (secret) (required)

## ETL Streams

- users:
  - primary key: id
  - fields: email(string), first_name(string), id(integer), last_name(string)
- locations:
  - primary key: id
  - fields: address(string), id(integer), name(string)
- positions:
  - primary key: id
  - fields: color(string), id(integer), name(string)
- shifts:
  - primary key: id
  - fields: end_time(string), id(integer), start_time(string), user_id(integer)
- sites:
  - primary key: id
  - fields: address(string), id(integer), is_deleted(boolean), location_id(integer), name(string)
- blocks:
  - primary key: id
  - fields: end_time(string), id(integer), location_id(integer), position_id(integer), start_time(string)
- annotations:
  - primary key: id
  - fields: end_date(string), id(integer), message(string), start_date(string), title(string)
- availabilityevents:
  - primary key: id
  - fields: end_time(string), id(integer), start_time(string), type(integer), user_id(integer)
- requesttypes:
  - primary key: id
  - fields: built_in(boolean), enabled(boolean), id(integer), is_deleted(boolean), name(string)
- times:
  - primary key: id
  - fields: end_time(string), id(integer), is_approved(boolean), shift_id(integer), start_time(string), user_id(integer)
- timezones:
  - primary key: timezone_id
  - fields: offset(number), olson_id(string), timezone_id(integer), timezone_name(string)
- payrolls:
  - primary key: id
  - fields: end_date(string), id(integer), is_closed(boolean), is_finalized(boolean), start_date(string)
- openshiftapprovalrequests:
  - primary key: id
  - fields: created_at(string), id(integer), shift_id(integer), status(integer)
- swaps:
  - primary key: id
  - fields: id(integer), shift_id(integer), status(integer), type(integer), user_id(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_user:
  - endpoint: POST /2/users
  - risk: external mutation; creates a workforce-scheduling user account; approval required
- update_user:
  - endpoint: PUT /2/users/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_user:
  - endpoint: DELETE /2/users/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a workforce-scheduling user account; approval required
- create_location:
  - endpoint: POST /2/locations
  - required fields: name
  - risk: external mutation; approval required
- update_location:
  - endpoint: PUT /2/locations/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_location:
  - endpoint: DELETE /2/locations/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a schedule location; approval required
- create_position:
  - endpoint: POST /2/positions
  - required fields: name
  - risk: external mutation; approval required
- update_position:
  - endpoint: PUT /2/positions/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_position:
  - endpoint: DELETE /2/positions/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a position; approval required
- create_site:
  - endpoint: POST /2/sites
  - required fields: location_id, name
  - risk: external mutation; approval required
- update_site:
  - endpoint: PUT /2/sites/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_site:
  - endpoint: DELETE /2/sites/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a site; approval required
- create_block:
  - endpoint: POST /2/blocks
  - required fields: start_time, end_time, location_id
  - risk: external mutation; approval required
- update_block:
  - endpoint: PUT /2/blocks/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_block:
  - endpoint: DELETE /2/blocks/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a shift template; approval required
- create_annotation:
  - endpoint: POST /2/annotations
  - required fields: start_date, end_date, title
  - risk: external mutation; approval required
- update_annotation:
  - endpoint: PUT /2/annotations/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_annotation:
  - endpoint: DELETE /2/annotations/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a schedule annotation; approval required
- create_availability_event:
  - endpoint: POST /2/availabilityevents
  - required fields: start_time, type
  - risk: external mutation; writes a user's availability/unavailability preference; approval required
- update_availability_event:
  - endpoint: PUT /2/availabilityevents/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_availability_event:
  - endpoint: DELETE /2/availabilityevents/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a user availability event; approval required
- create_time:
  - endpoint: POST /2/times
  - required fields: user_id, start_time, end_time
  - risk: external mutation; creates a worked-time entry feeding payroll; approval required
- update_time:
  - endpoint: PUT /2/times/{{ record.id }}
  - required fields: id
  - risk: external mutation; edits a worked-time entry feeding payroll; approval required
- delete_time:
  - endpoint: DELETE /2/times/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a worked-time entry feeding payroll; approval required
- create_shift:
  - endpoint: POST /2/shifts
  - required fields: start_time, end_time, location_id
  - risk: external mutation; creates a scheduled shift; approval required
- delete_shift:
  - endpoint: DELETE /2/shifts/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a scheduled shift; approval required

## Security

- read risk: external When I Work API read of the caller's own workforce-scheduling data
- write risk: external mutation of workforce-scheduling records (users, locations, positions, sites, shift templates, annotations, availability events, time entries feeding payroll, and shifts); create/update/delete all require approval, deletes are irreversible
- approval: read: none; write: required for every create/update/delete action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect when-i-work
```

### Inspect as structured JSON

```bash
pm connectors inspect when-i-work --json
```

## Agent Rules

- Run pm connectors inspect when-i-work before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
