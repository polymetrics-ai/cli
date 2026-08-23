---
name: pm-hoorayhr
description: HoorayHR connector knowledge and safe action guide.
---

# pm-hoorayhr

## Purpose

Reads HoorayHR users, time-off, leave-types, and sick-leave records through the HoorayHR REST API using session-token authentication.

## Icon

- id: hoorayhr
- asset: icons/hoorayhr.svg
- source: official
- review_status: official_verified
- review_url: https://api.hoorayhr.io/swagger.json

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- hoorayhrusername (required)
- mode
- hoorayhrpassword (secret) (required)

## ETL Streams

- users:
  - primary key: id
  - fields: companyId(integer), companyStartDate(string), createdAt(string), email(string), firstName(string), id(integer), isAdmin(boolean), jobTitle(string), lastName(string), status(string), updatedAt(string)
- time_off:
  - primary key: id
  - fields: createdAt(string), end(string), id(integer), leaveTypeId(integer), leaveUnit(string), notes(string), start(string), status(string), timeOffType(string), updatedAt(string), userId(integer)
- leave_types:
  - primary key: id
  - fields: budget(number), color(string), createdAt(string), default(boolean), icon(string), id(integer), leaveInDays(boolean), name(string), unpaidLeave(boolean), updatedAt(string)
- sick_leaves:
  - primary key: id
  - fields: actualReturn(string), actualStart(string), createdAt(string), id(integer), notes(string), percentage(number), reportedReturn(string), reportedStart(string), status(string), updatedAt(string), userId(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external HoorayHR API read of employee, time-off, leave-type, and sick-leave data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect hoorayhr
```

### Inspect as structured JSON

```bash
pm connectors inspect hoorayhr --json
```

## Agent Rules

- Run pm connectors inspect hoorayhr before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
