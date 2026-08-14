---
name: pm-clockify
description: Clockify connector knowledge and safe action guide.
---

# pm-clockify

## Purpose

Reads Clockify workspaces, clients, projects, tags, users, tasks, time entries, custom fields, user groups, holidays, expense categories, and time-off policies, and writes clients/projects/tags/tasks through the Clockify REST API v1.

## Icon

- id: clockify
- asset: icons/clockify.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.clockify.me/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- workspace_id
- api_key (secret) (required)

## ETL Streams

- workspaces:
  - primary key: id
  - fields: featureSubscriptionType(string), hourlyRate(object), id(string), imageUrl(string), memberships(array), name(string), workspaceSettings(object)
- clients:
  - primary key: id
  - fields: address(string), archived(boolean), email(string), id(string), name(string), note(string), workspaceId(string)
- projects:
  - primary key: id
  - fields: archived(boolean), billable(boolean), clientId(string), clientName(string), color(string), duration(string), id(string), name(string), note(string), public(boolean), workspaceId(string)
- tags:
  - primary key: id
  - fields: archived(boolean), id(string), name(string), workspaceId(string)
- users:
  - primary key: id
  - fields: activeWorkspace(string), defaultWorkspace(string), email(string), id(string), name(string), profilePicture(string), status(string)
- current_user:
  - primary key: id
  - fields: activeWorkspace(string), customFields(array), defaultWorkspace(string), email(string), id(string), memberships(array), name(string), profilePicture(string), settings(object), status(string)
- custom_fields:
  - primary key: id
  - fields: allowedValues(array), description(string), entityType(string), id(string), name(string), onlyAdminCanEdit(boolean), placeholder(string), projectDefaultValues(array), required(boolean), status(string), type(string), workspaceDefaultValue(string), workspaceId(string)
- user_groups:
  - primary key: id
  - fields: id(string), name(string), teamManagers(array), userIds(array), workspaceId(string)
- holidays:
  - primary key: id
  - fields: automaticTimeEntryCreation(boolean), datePeriod(object), everyoneIncludingNew(boolean), id(string), name(string), occursAnnually(boolean), projectId(string), taskId(string), userGroupIds(array), userIds(array), workspaceId(string)
- expense_categories:
  - primary key: id
  - fields: archived(boolean), hasUnitPrice(boolean), id(string), name(string), priceInCents(integer), unit(string), workspaceId(string)
- time_off_policies:
  - primary key: id
  - fields: allowHalfDay(boolean), allowNegativeBalance(boolean), approve(boolean), archived(boolean), automaticAccrual(object), automaticTimeEntryCreation(boolean), everyoneIncludingNew(boolean), id(string), name(string), negativeBalance(object), projectId(string), timeUnit(string), userGroupIds(array), userIds(array), workspaceId(string)
- tasks:
  - primary key: id
  - fields: assigneeId(string), assigneeIds(array), billable(boolean), budgetEstimate(object), costRate(object), duration(string), estimate(string), hourlyRate(object), id(string), name(string), projectId(string), status(string), userGroupIds(array)
- time_entries:
  - primary key: id
  - fields: billable(boolean), costRate(object), customFieldValues(array), description(string), hourlyRate(object), id(string), isLocked(boolean), kioskId(string), projectId(string), tagIds(array), taskId(string), timeInterval(object), type(string), userId(string), workspaceId(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_client:
  - endpoint: POST /v1/workspaces/{{ config.workspace_id }}/clients
  - required fields: name
  - risk: external mutation; creates a live Clockify client; approval required
- update_client:
  - endpoint: PUT /v1/workspaces/{{ config.workspace_id }}/clients/{{ record.id }}
  - required fields: id, name
  - risk: external mutation; overwrites a live Clockify client's fields; approval required
- delete_client:
  - endpoint: DELETE /v1/workspaces/{{ config.workspace_id }}/clients/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockify client; approval required
- create_project:
  - endpoint: POST /v1/workspaces/{{ config.workspace_id }}/projects
  - required fields: name
  - risk: external mutation; creates a live Clockify project; approval required
- update_project:
  - endpoint: PUT /v1/workspaces/{{ config.workspace_id }}/projects/{{ record.id }}
  - required fields: id, name
  - risk: external mutation; overwrites a live Clockify project's fields; approval required
- delete_project:
  - endpoint: DELETE /v1/workspaces/{{ config.workspace_id }}/projects/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockify project; approval required
- create_tag:
  - endpoint: POST /v1/workspaces/{{ config.workspace_id }}/tags
  - required fields: name
  - risk: external mutation; creates a live Clockify tag; approval required
- update_tag:
  - endpoint: PUT /v1/workspaces/{{ config.workspace_id }}/tags/{{ record.id }}
  - required fields: id, name
  - risk: external mutation; overwrites a live Clockify tag's fields; approval required
- delete_tag:
  - endpoint: DELETE /v1/workspaces/{{ config.workspace_id }}/tags/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockify tag; approval required
- create_task:
  - endpoint: POST /v1/workspaces/{{ config.workspace_id }}/projects/{{ record.projectId }}/tasks
  - required fields: projectId, name
  - risk: external mutation; creates a live Clockify task on a project; approval required
- update_task:
  - endpoint: PUT /v1/workspaces/{{ config.workspace_id }}/projects/{{ record.projectId }}/tasks/{{ record.id }}
  - required fields: projectId, id, name
  - risk: external mutation; overwrites a live Clockify task's fields; approval required
- delete_task:
  - endpoint: DELETE /v1/workspaces/{{ config.workspace_id }}/projects/{{ record.projectId }}/tasks/{{ record.id }}
  - required fields: projectId, id
  - risk: external mutation; irreversibly deletes a live Clockify task; approval required

## Security

- read risk: external Clockify API read of workspace, client, project, tag, user, task, time entry, and workspace-configuration data
- write risk: external mutation; creates/updates/deletes live Clockify clients, projects, tags, and tasks
- approval: required for all write actions; reads remain none
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect clockify
```

### Inspect as structured JSON

```bash
pm connectors inspect clockify --json
```

## Agent Rules

- Run pm connectors inspect clockify before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
