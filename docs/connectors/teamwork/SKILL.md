---
name: pm-teamwork
description: Teamwork connector knowledge and safe action guide.
---

# pm-teamwork

## Purpose

Reads Teamwork projects, people, companies, tags, time entries, tasklists, milestones, and tasks, and writes approved project/tasklist/task/milestone/company/time-entry mutations through the Teamwork API.

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
- username
- password (secret)

## ETL Streams

- projects:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), id(), name()
- people:
  - primary key: id
  - fields: administrator(), company-id(), email-address(), first_name(), id(), last_name(), user-name()
- companies:
  - primary key: id
  - fields: address_one(), id(), name(), phone(), website()
- tags:
  - primary key: id
  - fields: color(), id(), name()
- time_entries:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), date(), description(), hours(), id(), isbillable(), minutes(), person_id(), project_id(), todo_item_id()
- tasklists:
  - primary key: id
  - fields: complete(), description(), id(), milestone-id(), name(), project_id()
- milestones:
  - primary key: id
  - cursor: created_at
  - fields: completed(), created_at(), deadline(), description(), id(), project_id(), title()
- tasks:
  - primary key: id
  - cursor: created_at
  - fields: content(), created_at(), description(), id(), priority(), project-id(), project-name(), status(), todo-list-id(), todo-list-name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_project:
  - endpoint: POST /projects.json
  - risk: creates a new project; low-risk external mutation, no approval required. Body is wrapped under a top-level "project" key (Teamwork's V1 API convention) — the record itself must carry that wrapper, since the engine's write dialect sends record fields verbatim as the JSON body with no nested-wrapper construction primitive.
- update_project:
  - endpoint: PUT /projects/{{ record.id }}.json
  - required fields: id
  - risk: mutates an existing project's name or description; visible to every project member
- create_tasklist:
  - endpoint: POST /projects/{{ record.project_id }}/tasklists.json
  - required fields: project_id
  - risk: creates a new task list under the target project; low-risk external mutation, no approval required
- create_task:
  - endpoint: POST /tasklists/{{ record.tasklist_id }}/tasks.json
  - required fields: tasklist_id
  - risk: creates a new task in the target task list; low-risk external mutation, no approval required
- update_task:
  - endpoint: PUT /tasks/{{ record.id }}.json
  - required fields: id
  - risk: mutates an existing task's content, description, or priority
- complete_task:
  - endpoint: PUT /tasks/{{ record.id }}/complete.json
  - required fields: id
  - risk: marks an existing task as complete; a visible, notifiable state change for every task follower
- create_milestone:
  - endpoint: POST /projects/{{ record.project_id }}/milestones.json
  - required fields: project_id
  - risk: creates a new milestone under the target project; low-risk external mutation, no approval required
- create_company:
  - endpoint: POST /companies.json
  - risk: creates a new company record; low-risk external mutation, no approval required
- create_time_entry:
  - endpoint: POST /projects/{{ record.project_id }}/time_entries.json
  - required fields: project_id
  - risk: logs a new time entry against the target project; contributes to billable-hours totals and any linked invoice

## Security

- read risk: external Teamwork API read of project, people, company, tag, time-entry, tasklist, milestone, and task data
- write risk: external Teamwork API mutation (create/update projects, tasklists, tasks, milestones, companies, time entries; complete tasks)
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect teamwork
```

### Inspect as structured JSON

```bash
pm connectors inspect teamwork --json
```

## Agent Rules

- Run pm connectors inspect teamwork before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
