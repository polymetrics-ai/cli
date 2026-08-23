---
name: pm-teamwork
description: Teamwork connector knowledge and safe action guide.
---

# pm-teamwork

## Purpose

Reads Teamwork projects, people, companies, tags, time entries, tasklists, milestones, and tasks, and writes approved project/tasklist/task/milestone/company/time-entry mutations through the Teamwork API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- username (required)
- password (secret) (required)

## ETL Streams

- projects:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), name(string)
- people:
  - primary key: id
  - fields: administrator(boolean), company-id(string), email-address(string), first_name(string), id(string), last_name(string), user-name(string)
- companies:
  - primary key: id
  - fields: address_one(string), id(string), name(string), phone(string), website(string)
- tags:
  - primary key: id
  - fields: color(string), id(string), name(string)
- time_entries:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), date(string), description(string), hours(string), id(string), isbillable(string), minutes(string), person_id(string), project_id(string), todo_item_id(string)
- tasklists:
  - primary key: id
  - fields: complete(boolean), description(string), id(string), milestone-id(string), name(string), project_id(string)
- milestones:
  - primary key: id
  - cursor: created_at
  - fields: completed(boolean), created_at(string), deadline(string), description(string), id(string), project_id(string), title(string)
- tasks:
  - primary key: id
  - cursor: created_at
  - fields: content(string), created_at(string), description(string), id(string), priority(string), project-id(string), project-name(string), status(string), todo-list-id(string), todo-list-name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_project:
  - endpoint: POST /projects.json
  - required fields: project
  - risk: creates a new project; low-risk external mutation, no approval required. Body is wrapped under a top-level "project" key (Teamwork's V1 API convention) — the record itself must carry that wrapper, since the engine's write dialect sends record fields verbatim as the JSON body with no nested-wrapper construction primitive.
- update_project:
  - endpoint: PUT /projects/{{ record.id }}.json
  - required fields: id, project
  - risk: mutates an existing project's name or description; visible to every project member
- create_tasklist:
  - endpoint: POST /projects/{{ record.project_id }}/tasklists.json
  - required fields: project_id, todo-list
  - risk: creates a new task list under the target project; low-risk external mutation, no approval required
- create_task:
  - endpoint: POST /tasklists/{{ record.tasklist_id }}/tasks.json
  - required fields: tasklist_id, todo-item
  - risk: creates a new task in the target task list; low-risk external mutation, no approval required
- update_task:
  - endpoint: PUT /tasks/{{ record.id }}.json
  - required fields: id, todo-item
  - risk: mutates an existing task's content, description, or priority
- complete_task:
  - endpoint: PUT /tasks/{{ record.id }}/complete.json
  - required fields: id
  - risk: marks an existing task as complete; a visible, notifiable state change for every task follower
- create_milestone:
  - endpoint: POST /projects/{{ record.project_id }}/milestones.json
  - required fields: project_id, milestone
  - risk: creates a new milestone under the target project; low-risk external mutation, no approval required
- create_company:
  - endpoint: POST /companies.json
  - required fields: company
  - risk: creates a new company record; low-risk external mutation, no approval required
- create_time_entry:
  - endpoint: POST /projects/{{ record.project_id }}/time_entries.json
  - required fields: project_id, time-entry
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
