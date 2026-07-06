---
name: pm-clockify
description: Clockify connector knowledge and safe action guide.
---

# pm-clockify

## Purpose

Reads Clockify workspaces, clients, projects, tags, users, tasks, time entries, custom fields, user groups, holidays, expense categories, and time-off policies, and writes clients/projects/tags/tasks through the Clockify REST API v1.

## Icon

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
- api_key (secret)

## ETL Streams

- workspaces:
  - primary key: id
  - fields: featureSubscriptionType(), hourlyRate(), id(), imageUrl(), memberships(), name(), workspaceSettings()
- clients:
  - primary key: id
  - fields: address(), archived(), email(), id(), name(), note(), workspaceId()
- projects:
  - primary key: id
  - fields: archived(), billable(), clientId(), clientName(), color(), duration(), id(), name(), note(), public(), workspaceId()
- tags:
  - primary key: id
  - fields: archived(), id(), name(), workspaceId()
- users:
  - primary key: id
  - fields: activeWorkspace(), defaultWorkspace(), email(), id(), name(), profilePicture(), status()
- current_user:
  - primary key: id
  - fields: activeWorkspace(), customFields(), defaultWorkspace(), email(), id(), memberships(), name(), profilePicture(), settings(), status()
- custom_fields:
  - primary key: id
  - fields: allowedValues(), description(), entityType(), id(), name(), onlyAdminCanEdit(), placeholder(), projectDefaultValues(), required(), status(), type(), workspaceDefaultValue(), workspaceId()
- user_groups:
  - primary key: id
  - fields: id(), name(), teamManagers(), userIds(), workspaceId()
- holidays:
  - primary key: id
  - fields: automaticTimeEntryCreation(), datePeriod(), everyoneIncludingNew(), id(), name(), occursAnnually(), projectId(), taskId(), userGroupIds(), userIds(), workspaceId()
- expense_categories:
  - primary key: id
  - fields: archived(), hasUnitPrice(), id(), name(), priceInCents(), unit(), workspaceId()
- time_off_policies:
  - primary key: id
  - fields: allowHalfDay(), allowNegativeBalance(), approve(), archived(), automaticAccrual(), automaticTimeEntryCreation(), everyoneIncludingNew(), id(), name(), negativeBalance(), projectId(), timeUnit(), userGroupIds(), userIds(), workspaceId()
- tasks:
  - primary key: id
  - fields: assigneeId(), assigneeIds(), billable(), budgetEstimate(), costRate(), duration(), estimate(), hourlyRate(), id(), name(), projectId(), status(), userGroupIds()
- time_entries:
  - primary key: id
  - fields: billable(), costRate(), customFieldValues(), description(), hourlyRate(), id(), isLocked(), kioskId(), projectId(), tagIds(), taskId(), timeInterval(), type(), userId(), workspaceId()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_client:
  - endpoint: POST /v1/workspaces/{{ config.workspace_id }}/clients
  - risk: external mutation; creates a live Clockify client; approval required
- update_client:
  - endpoint: PUT /v1/workspaces/{{ config.workspace_id }}/clients/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Clockify client's fields; approval required
- delete_client:
  - endpoint: DELETE /v1/workspaces/{{ config.workspace_id }}/clients/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockify client; approval required
- create_project:
  - endpoint: POST /v1/workspaces/{{ config.workspace_id }}/projects
  - risk: external mutation; creates a live Clockify project; approval required
- update_project:
  - endpoint: PUT /v1/workspaces/{{ config.workspace_id }}/projects/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Clockify project's fields; approval required
- delete_project:
  - endpoint: DELETE /v1/workspaces/{{ config.workspace_id }}/projects/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockify project; approval required
- create_tag:
  - endpoint: POST /v1/workspaces/{{ config.workspace_id }}/tags
  - risk: external mutation; creates a live Clockify tag; approval required
- update_tag:
  - endpoint: PUT /v1/workspaces/{{ config.workspace_id }}/tags/{{ record.id }}
  - required fields: id
  - risk: external mutation; overwrites a live Clockify tag's fields; approval required
- delete_tag:
  - endpoint: DELETE /v1/workspaces/{{ config.workspace_id }}/tags/{{ record.id }}
  - required fields: id
  - risk: external mutation; irreversibly deletes a live Clockify tag; approval required
- create_task:
  - endpoint: POST /v1/workspaces/{{ config.workspace_id }}/projects/{{ record.projectId }}/tasks
  - required fields: projectId
  - risk: external mutation; creates a live Clockify task on a project; approval required
- update_task:
  - endpoint: PUT /v1/workspaces/{{ config.workspace_id }}/projects/{{ record.projectId }}/tasks/{{ record.id }}
  - required fields: projectId, id
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
