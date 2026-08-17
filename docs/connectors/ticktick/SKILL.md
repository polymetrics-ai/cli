---
name: pm-ticktick
description: TickTick connector knowledge and safe action guide.
---

# pm-ticktick

## Purpose

Reads projects and project tasks, and writes task create/complete/delete actions, through the TickTick Open API.

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
- project_id
- access_token (secret)
- bearer_token (secret)
- client_access_token (secret)

## ETL Streams

- projects:
  - primary key: id
  - fields: closed(), color(), groupId(), id(), kind(), name(), permission(), sortOrder(), viewMode()
- tasks:
  - primary key: id
  - fields: completedTime(), content(), desc(), dueDate(), id(), isAllDay(), modifiedTime(), priority(), projectId(), reminders(), repeatFlag(), sortOrder(), startDate(), status(), timeZone(), title()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_task:
  - endpoint: POST /task
  - risk: creates a new task in the caller's TickTick account (in the given projectId, or the default Inbox if omitted); low-risk external mutation, no approval required
- complete_task:
  - endpoint: POST /project/{{ record.projectId }}/task/{{ record.id }}/complete
  - required fields: projectId, id
  - risk: marks an existing task as completed; a completed task is removed from active task lists/reminders for every collaborator on the project
- delete_task:
  - endpoint: DELETE /project/{{ record.projectId }}/task/{{ record.id }}
  - required fields: projectId, id
  - risk: permanently removes a task from the given project; irreversible, no undo via the API

## Security

- read risk: external TickTick API read of project and task data
- write risk: external TickTick API mutation: creates a task, marks a task complete, or permanently deletes a task
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ticktick
```

### Inspect as structured JSON

```bash
pm connectors inspect ticktick --json
```

## Agent Rules

- Run pm connectors inspect ticktick before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
