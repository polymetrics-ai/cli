---
name: pm-hoorayhr
description: HoorayHR connector knowledge and safe action guide.
---

# pm-hoorayhr

## Purpose

Reads HoorayHR users, time-off, leave-types, and sick-leave records through the HoorayHR REST API using session-token authentication.

## Icon

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
- hoorayhrusername
- mode
- hoorayhrpassword (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: companyId(), companyStartDate(), createdAt(), email(), firstName(), id(), isAdmin(), jobTitle(), lastName(), status(), updatedAt()
- time_off:
  - primary key: id
  - fields: createdAt(), end(), id(), leaveTypeId(), leaveUnit(), notes(), start(), status(), timeOffType(), updatedAt(), userId()
- leave_types:
  - primary key: id
  - fields: budget(), color(), createdAt(), default(), icon(), id(), leaveInDays(), name(), unpaidLeave(), updatedAt()
- sick_leaves:
  - primary key: id
  - fields: actualReturn(), actualStart(), createdAt(), id(), notes(), percentage(), reportedReturn(), reportedStart(), status(), updatedAt(), userId()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
