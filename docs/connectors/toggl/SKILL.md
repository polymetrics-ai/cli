---
name: pm-toggl
description: Toggl connector knowledge and safe action guide.
---

# pm-toggl

## Purpose

Reads and writes time entries, projects, clients, tags, tasks, and users through the Toggl Track API.

## Icon

- id: toggl
- asset: icons/toggl.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.track.toggl.com/docs/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- end_date
- organization_id
- start_date
- workspace_id
- api_token (secret) (required)

## ETL Streams

- time_entries:
  - primary key: id
  - fields: at(string), billable(boolean), created_with(string), description(string), duration(integer), duronly(boolean), id(integer), pid(integer), project_id(integer), server_deleted_at(string), start(string), stop(string), tag_ids(array), tags(array), task_id(integer), tid(integer), uid(integer), user_id(integer), wid(integer), workspace_id(integer)
- projects:
  - primary key: id
  - fields: active(boolean), actual_hours(integer), at(string), auto_estimates(boolean), billable(boolean), cid(integer), client_id(integer), color(string), created_at(string), currency(string), estimated_hours(integer), fixed_fee(number), id(integer), is_private(boolean), name(string), rate(number), rate_last_updated(string), recurring(boolean), recurring_parameters(object), server_deleted_at(string), template(boolean), wid(integer), workspace_id(integer)
- clients:
  - primary key: id
  - fields: archived(boolean), at(string), id(integer), name(string), server_deleted_at(string), wid(integer), workspace_id(integer)
- workspace_users:
  - primary key: id
  - fields: active(boolean), admin(boolean), at(string), avatar_file_name(string), email(string), group_ids(array), id(integer), inactive(boolean), invitation_code(string), invite_url(string), is_direct(boolean), labour_cost(number), name(string), organization_admin(boolean), rate(number), rate_last_updated(string), timezone(string), user_id(integer), workspace_admin(boolean), workspace_id(integer)
- organization_users:
  - primary key: id
  - fields: admin(boolean), avatar_url(string), can_edit_email(boolean), email(string), groups(array), id(integer), inactive(boolean), invitation_code(string), joined(boolean), name(string), owner(boolean), user_id(integer), workspaces(array)
- tags:
  - primary key: id
  - fields: at(string), creator_id(integer), deleted_at(string), id(integer), name(string), workspace_id(integer)
- tasks:
  - primary key: id
  - fields: active(boolean), at(string), client_id(integer), client_name(string), estimated_seconds(integer), external_reference(string), id(integer), name(string), project_billable(boolean), project_color(string), project_id(integer), project_is_private(boolean), project_name(string), rate(number), rate_last_updated(string), recurring(boolean), tracked_seconds(integer), user_id(integer), user_name(string), workspace_id(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_time_entry:
  - endpoint: POST /workspaces/{{ config.workspace_id }}/time_entries
  - required fields: start, duration, created_with
  - risk: creates a new time entry on the caller's account; external mutation, no approval required
- update_time_entry:
  - endpoint: PUT /workspaces/{{ config.workspace_id }}/time_entries/{{ record.id }}
  - required fields: id
  - optional fields: description, start, stop, duration, project_id, task_id, tag_ids, tags, tag_action, billable, created_with
  - risk: mutates an existing time entry's timing, project/task association, tags, or billable flag
- stop_time_entry:
  - endpoint: PATCH /workspaces/{{ config.workspace_id }}/time_entries/{{ record.id }}/stop
  - required fields: id
  - risk: stops a currently-running time entry by setting its stop time to now; no effect on an already-stopped entry beyond the API's own idempotency
- delete_time_entry:
  - endpoint: DELETE /workspaces/{{ config.workspace_id }}/time_entries/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a time entry; irreversible
- create_project:
  - endpoint: POST /workspaces/{{ config.workspace_id }}/projects
  - required fields: name
  - risk: creates a new project in the target workspace; external mutation, no approval required
- update_project:
  - endpoint: PUT /workspaces/{{ config.workspace_id }}/projects/{{ record.id }}
  - required fields: id
  - optional fields: name, client_id, is_private, active, color, billable, rate, currency, start_date, end_date
  - risk: mutates an existing project's name, client association, active/private state, or billing settings
- delete_project:
  - endpoint: DELETE /workspaces/{{ config.workspace_id }}/projects/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a project; also removes its association from any time entries that referenced it
- create_client:
  - endpoint: POST /workspaces/{{ config.workspace_id }}/clients
  - required fields: name
  - risk: creates a new client in the target workspace; external mutation, no approval required
- update_client:
  - endpoint: PUT /workspaces/{{ config.workspace_id }}/clients/{{ record.id }}
  - required fields: id
  - optional fields: name, notes, external_reference
  - risk: mutates an existing client's name or notes
- delete_client:
  - endpoint: DELETE /workspaces/{{ config.workspace_id }}/clients/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a client; projects previously associated with it lose that association
- create_tag:
  - endpoint: POST /workspaces/{{ config.workspace_id }}/tags
  - required fields: name
  - risk: creates a new tag in the target workspace; external mutation, no approval required
- update_tag:
  - endpoint: PUT /workspaces/{{ config.workspace_id }}/tags/{{ record.id }}
  - required fields: id, name
  - risk: renames an existing tag; the new name applies retroactively everywhere the tag is shown
- delete_tag:
  - endpoint: DELETE /workspaces/{{ config.workspace_id }}/tags/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a tag; it is removed from every time entry that referenced it
- create_task:
  - endpoint: POST /workspaces/{{ config.workspace_id }}/projects/{{ record.project_id }}/tasks
  - required fields: project_id, name
  - optional fields: active, estimated_seconds, user_id, external_reference
  - risk: creates a new task under the given project; external mutation, no approval required
- update_task:
  - endpoint: PUT /workspaces/{{ config.workspace_id }}/projects/{{ record.project_id }}/tasks/{{ record.id }}
  - required fields: project_id, id
  - optional fields: name, active, estimated_seconds, user_id, external_reference
  - risk: mutates an existing task's name, active/done state, estimate, or assignee; setting active:false marks the task done
- delete_task:
  - endpoint: DELETE /workspaces/{{ config.workspace_id }}/projects/{{ record.project_id }}/tasks/{{ record.id }}
  - required fields: project_id, id
  - risk: permanently deletes a task; time entries previously linked to it lose that association

## Security

- read risk: external Toggl Track API read of time-tracking and workspace data
- write risk: external mutation of Toggl time entries, projects, clients, tags, and tasks; no destructive-admin or elevated-scope actions modeled
- approval: create/update actions execute without approval (low-risk, reversible via a follow-up write); delete_* actions are irreversible and should be gated by the caller's own reverse-ETL approval policy
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect toggl
```

### Inspect as structured JSON

```bash
pm connectors inspect toggl --json
```

## Agent Rules

- Run pm connectors inspect toggl before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
