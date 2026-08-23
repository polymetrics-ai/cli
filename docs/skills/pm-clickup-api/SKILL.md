---
name: pm-clickup-api
description: ClickUp connector knowledge and safe action guide.
---

# pm-clickup-api

## Purpose

Reads ClickUp workspaces (teams), spaces, folders, lists, tasks, goals, space tags, and webhooks, and writes task/folder/list/space/webhook lifecycle mutations, task comments, tags, custom field values, and goal creation, through the ClickUp v2 REST API using a personal API token.

## Icon

- id: clickup
- asset: icons/clickup.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://clickup.com/api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- folder_id
- include_archived
- include_closed_tasks
- list_id
- mode
- space_id
- team_id
- api_token (secret) (required)

## ETL Streams

- tasks:
  - primary key: id
  - cursor: date_updated
  - fields: creator_id(integer), date_closed(string), date_created(string), date_updated(string), folder_id(string), id(string), list_id(string), name(string), space_id(string), status(string), url(string)
- teams:
  - primary key: id
  - fields: avatar(string), color(string), id(string), name(string)
- spaces:
  - primary key: id
  - fields: archived(boolean), id(string), multiple_assignees(boolean), name(string), private(boolean)
- folders:
  - primary key: id
  - fields: archived(boolean), hidden(boolean), id(string), name(string), orderindex(integer), space_id(string), task_count(string)
- lists:
  - primary key: id
  - fields: archived(boolean), id(string), name(string), orderindex(integer), space_id(string), task_count(integer)
- goals:
  - primary key: id
  - fields: archived(boolean), color(string), creator(integer), date_created(string), description(string), due_date(string), id(string), multiple_owners(boolean), name(string), percent_completed(integer), private(boolean), start_date(string), team_id(string)
- space_tags:
  - primary key: name
  - fields: name(string), space_id(string), tag_bg(string), tag_fg(string)
- webhooks:
  - primary key: id
  - fields: client_id(string), endpoint(string), events(array), folder_id(integer), health(object), id(string), list_id(integer), space_id(integer), task_id(string), team_id(integer), userid(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_task:
  - endpoint: POST /list/{{ config.list_id }}/task
  - required fields: name
  - risk: creates a new ClickUp task in the configured list; low-risk (additive)
- update_task:
  - endpoint: PUT /task/{{ record.id }}
  - required fields: id
  - risk: updates fields on an existing ClickUp task (name, description, status, dates, priority, archived); approval required
- delete_task:
  - endpoint: DELETE /task/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a ClickUp task; irreversible; approval required
- create_task_comment:
  - endpoint: POST /task/{{ record.task_id }}/comment
  - required fields: task_id, comment_text, notify_all
  - optional fields: assignee, group_assignee
  - risk: adds a new comment to a ClickUp task, visible to all task watchers when notify_all is true; low-risk
- add_tag_to_task:
  - endpoint: POST /task/{{ record.task_id }}/tag/{{ record.tag_name }}
  - required fields: task_id, tag_name
  - risk: attaches an existing Space Tag to a task; low-risk
- remove_tag_from_task:
  - endpoint: DELETE /task/{{ record.task_id }}/tag/{{ record.tag_name }}
  - required fields: task_id, tag_name
  - risk: removes a tag from a task (does not delete the tag from the Space); low-risk
- set_custom_field_value:
  - endpoint: POST /task/{{ record.task_id }}/field/{{ record.field_id }}
  - required fields: task_id, field_id, value
  - optional fields: value_options
  - risk: sets a Custom Field value on a task; the accepted value shape varies by the field's type (text/number/date/dropdown/label/people/task-relationship/manual-progress/location/button); approval required since an incorrectly-typed value can silently fail or corrupt a differently-typed field
- create_goal:
  - endpoint: POST /team/{{ config.team_id }}/goal
  - required fields: name
  - risk: creates a new ClickUp Goal in the configured team/workspace; low-risk (additive)
- create_folder:
  - endpoint: POST /space/{{ config.space_id }}/folder
  - required fields: name
  - risk: creates a new Folder in the configured space; low-risk (additive)
- update_folder:
  - endpoint: PUT /folder/{{ record.id }}
  - required fields: id, name
  - risk: renames an existing ClickUp Folder; approval required
- delete_folder:
  - endpoint: DELETE /folder/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a ClickUp Folder and every List/task inside it; irreversible; approval required
- create_list:
  - endpoint: POST /folder/{{ config.folder_id }}/list
  - required fields: name
  - risk: creates a new List in the configured Folder; low-risk (additive)
- update_list:
  - endpoint: PUT /list/{{ record.id }}
  - required fields: id, name
  - risk: updates an existing ClickUp List's name/description/due date/priority/assignee/color; approval required
- delete_list:
  - endpoint: DELETE /list/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a ClickUp List and every task inside it; irreversible; approval required
- create_space:
  - endpoint: POST /team/{{ config.team_id }}/space
  - required fields: name
  - risk: creates a new Space in the configured Workspace; low-risk (additive)
- update_space:
  - endpoint: PUT /space/{{ record.id }}
  - required fields: id, name
  - risk: updates an existing ClickUp Space's name/color/privacy/ClickApp feature toggles; ClickUp's own docs mark every body field required (a partial update still needs the full current feature set re-sent to avoid resetting unspecified features); approval required
- delete_space:
  - endpoint: DELETE /space/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a ClickUp Space and every Folder/List/task inside it; irreversible; approval required
- create_webhook:
  - endpoint: POST /team/{{ config.team_id }}/webhook
  - required fields: endpoint, events
  - risk: registers or repoints an outbound event-delivery URL of the caller's choosing; approval required
- update_webhook:
  - endpoint: PUT /webhook/{{ record.id }}
  - required fields: id
  - risk: changes which events are delivered to (or repoints) an existing outbound webhook; approval required
- delete_webhook:
  - endpoint: DELETE /webhook/{{ record.id }}
  - required fields: id
  - risk: stops event delivery to a registered webhook endpoint; approval required (irreversible without re-registering)

## Security

- read risk: external ClickUp API read of workspace, space, folder, list, task, goal, tag, and webhook data
- write risk: external mutation of ClickUp tasks, folders, lists, spaces, webhooks, tags, custom field values, and goals; delete_task/delete_folder/delete_list/delete_space are irreversible cascading deletes, and create_webhook/update_webhook register or repoint an outbound event-delivery URL of the caller's choosing — every write ships with an explicit per-action risk string
- approval: required for every delete_* action, every update_* action, create_webhook/update_webhook, and set_custom_field_value; create_task/create_task_comment/add_tag_to_task/remove_tag_from_task/create_goal/create_folder/create_list/create_space are low-risk (additive or non-destructive)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect clickup-api
```

### Inspect as structured JSON

```bash
pm connectors inspect clickup-api --json
```

## Agent Rules

- Run pm connectors inspect clickup-api before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
