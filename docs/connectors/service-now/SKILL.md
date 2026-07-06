---
name: pm-service-now
description: ServiceNow connector knowledge and safe action guide.
---

# pm-service-now

## Purpose

Reads and writes ServiceNow incident, user, and group table data through the ServiceNow Table API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- username
- password (secret)

## ETL Streams

- incidents:
  - primary key: sys_id
  - cursor: updated_on
  - fields: name(), number(), priority(), short_description(), state(), sys_created_on(), sys_id(), updated_on()
- users:
  - primary key: sys_id
  - cursor: updated_on
  - fields: active(), email(), name(), number(), sys_id(), updated_on(), user_name()
- groups:
  - primary key: sys_id
  - cursor: updated_on
  - fields: active(), description(), name(), number(), sys_id(), updated_on()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_incident:
  - endpoint: POST /api/now/table/incident
  - risk: creates a new incident record; low-risk external mutation (a new ticket), no approval required
- create_user:
  - endpoint: POST /api/now/table/sys_user
  - risk: creates a new ServiceNow user account record; a new user account granted whatever role/ACL defaults the instance applies is a higher-scrutiny mutation than an incident/group create
- create_group:
  - endpoint: POST /api/now/table/sys_user_group
  - risk: creates a new user group record; low-risk external mutation, no approval required
- update_incident:
  - endpoint: PATCH /api/now/table/incident/{{ record.sys_id }}
  - required fields: sys_id
  - risk: mutates an existing incident's recorded fields (only fields present in the submitted record are changed; ServiceNow's Table API PATCH/PUT both modify only the submitted fields, never the whole record) by sys_id
- update_user:
  - endpoint: PATCH /api/now/table/sys_user/{{ record.sys_id }}
  - required fields: sys_id
  - risk: mutates an existing user account's profile fields by sys_id, including active (deactivating a user's account revokes their instance access); higher-scrutiny than incident/group updates
- update_group:
  - endpoint: PATCH /api/now/table/sys_user_group/{{ record.sys_id }}
  - required fields: sys_id
  - risk: mutates an existing group's recorded fields by sys_id, including active/manager; can change who is considered the group's membership owner

## Security

- read risk: external ServiceNow API read of incident, user, and group table data
- write risk: creates incident/user/group records and updates their fields by sys_id (ServiceNow Table API PATCH, which modifies only submitted fields); creating/deactivating a user account is a higher-scrutiny mutation than incident/group create-update
- approval: none for incident/group create-update (low-risk ticketing/CRM-style data); review user create/update before enabling in a caller with untrusted input, since it can grant or revoke ServiceNow instance access
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect service-now
```

### Inspect as structured JSON

```bash
pm connectors inspect service-now --json
```

## Agent Rules

- Run pm connectors inspect service-now before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
