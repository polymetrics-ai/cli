---
name: pm-asana
description: Asana connector knowledge and safe action guide.
---

# pm-asana

## Purpose

Reads Asana workspaces, projects, tasks, sections, tags, stories, users, teams, custom fields, project statuses, and team/workspace memberships through the Asana v1 REST API. Writes task/project/section/tag create-update-delete and task comments.

## Icon

- asset: icons/asana.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.asana.com/reference/rest-api-reference

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- assignee
- base_url
- mode
- project_id
- team_id
- workspace_id
- access_token (secret)

## ETL Streams

- workspaces:
  - primary key: gid
  - fields: gid(), name(), resource_type()
- projects:
  - primary key: gid
  - fields: created_at(), gid(), modified_at(), name(), resource_type()
- tasks:
  - primary key: gid
  - fields: completed(), created_at(), gid(), modified_at(), name(), resource_type()
- users:
  - primary key: gid
  - fields: email(), gid(), name(), resource_type()
- teams:
  - primary key: gid
  - fields: description(), gid(), name(), resource_type(), visibility()
- tags:
  - primary key: gid
  - fields: color(), created_at(), gid(), name(), notes(), resource_type()
- sections:
  - primary key: gid
  - fields: created_at(), gid(), name(), project_gid(), resource_type()
- stories:
  - primary key: gid
  - fields: created_at(), gid(), resource_subtype(), resource_type(), task_gid(), text(), type()
- custom_fields:
  - primary key: gid
  - fields: description(), enabled(), gid(), name(), resource_type(), type()
- project_statuses:
  - primary key: gid
  - fields: color(), created_at(), gid(), modified_at(), project_gid(), resource_type(), text(), title()
- team_memberships:
  - primary key: gid
  - fields: gid(), is_admin(), is_guest(), is_limited_access(), resource_type()
- workspace_memberships:
  - primary key: gid
  - fields: gid(), is_active(), is_admin(), is_guest(), resource_type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_task:
  - endpoint: POST /tasks
  - risk: external mutation; creates a new task visible to the whole team/project; approval required
- update_task:
  - endpoint: PUT /tasks/{{ record.gid }}
  - required fields: gid
  - risk: external mutation; overwrites task fields (e.g. can mark completed, reassign, reschedule); approval required
- delete_task:
  - endpoint: DELETE /tasks/{{ record.gid }}
  - required fields: gid
  - risk: irreversible external deletion of a task; approval required
- create_project:
  - endpoint: POST /projects
  - risk: external mutation; creates a new project visible to the team/workspace; approval required
- update_project:
  - endpoint: PUT /projects/{{ record.gid }}
  - required fields: gid
  - risk: external mutation; overwrites project fields (can archive, reassign owner); approval required
- delete_project:
  - endpoint: DELETE /projects/{{ record.gid }}
  - required fields: gid
  - risk: irreversible external deletion of a project and its association with its tasks; approval required
- create_section:
  - endpoint: POST /projects/{{ record.project_gid }}/sections
  - required fields: project_gid
  - risk: external mutation; creates a new section in a project's board/list view; approval required
- update_section:
  - endpoint: PUT /sections/{{ record.gid }}
  - required fields: gid
  - risk: external mutation; renames a section; approval required
- delete_section:
  - endpoint: DELETE /sections/{{ record.gid }}
  - required fields: gid
  - risk: irreversible external deletion of a section (Asana requires the section be empty of tasks first); approval required
- create_tag:
  - endpoint: POST /tags
  - risk: external mutation; creates a new workspace-visible tag; approval required
- update_tag:
  - endpoint: PUT /tags/{{ record.gid }}
  - required fields: gid
  - risk: external mutation; renames/recolors a tag visible workspace-wide; approval required
- delete_tag:
  - endpoint: DELETE /tags/{{ record.gid }}
  - required fields: gid
  - risk: irreversible external deletion of a tag, removed from every task that carries it; approval required
- add_comment:
  - endpoint: POST /tasks/{{ record.task_gid }}/stories
  - required fields: task_gid
  - risk: external mutation; posts a comment visible to everyone with access to the task; approval required

## Security

- read risk: external Asana API read of workspace, project, task, and PM metadata
- write risk: external mutations: creates/updates/deletes tasks, projects, sections, and tags, and posts task comments
- approval: required for every write action; deletes are flagged destructive/irreversible in writes.json's per-action risk field
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Inspect and safely plan changes to Asana workspaces, projects, and tasks.
- Usage: pm asana <topic> <leaf> [flags]
- Global flags:
  - --json (boolean): Write deterministic machine-readable JSON output.
- Organization
  - workspaces list - List available Asana workspaces. [intent=etl availability=implemented stream=workspaces]
  - users list - List Asana users, optionally scoped to a workspace. [intent=etl availability=implemented stream=users]; flags: --workspace
  - teams list - List teams in an Asana workspace. [intent=etl availability=implemented stream=teams]; flags: --workspace
  - team-memberships list - List Asana team memberships. [intent=etl availability=implemented stream=team_memberships]; flags: --team, --workspace
  - workspace-memberships list - List memberships in an Asana workspace. [intent=etl availability=implemented stream=workspace_memberships]; flags: --workspace
- Work Management
  - projects list - List Asana projects, optionally scoped to a workspace. [intent=etl availability=implemented stream=projects]; flags: --workspace
  - projects create - Plan creation of an Asana project. [intent=reverse_etl availability=implemented write=create_project]; approval: Requires plan, preview, explicit approval, then execute.; risk: external mutation; creates a new project visible to the team/workspace; approval required; flags: --data, --name
  - projects update - Plan updates to an Asana project. [intent=reverse_etl availability=implemented write=update_project]; approval: Requires plan, preview, explicit approval, then execute.; risk: external mutation; overwrites project fields (can archive, reassign owner); approval required; flags: --gid, --data
  - projects delete - Plan deletion of an Asana project. [intent=reverse_etl availability=implemented write=delete_project]; approval: Requires plan, preview, explicit approval, then execute.; risk: irreversible external deletion of a project and its association with its tasks; approval required; flags: --gid
  - project-statuses list - List statuses for discovered Asana projects. [intent=etl availability=implemented stream=project_statuses]
  - sections list - List sections for discovered Asana projects. [intent=etl availability=implemented stream=sections]
  - sections create - Plan creation of an Asana project section. [intent=reverse_etl availability=implemented write=create_section]; approval: Requires plan, preview, explicit approval, then execute.; risk: external mutation; creates a new section in a project's board/list view; approval required; flags: --project-gid, --data, --name
  - sections update - Plan updates to an Asana section. [intent=reverse_etl availability=implemented write=update_section]; approval: Requires plan, preview, explicit approval, then execute.; risk: external mutation; renames a section; approval required; flags: --gid, --data, --name
  - sections delete - Plan deletion of an empty Asana section. [intent=reverse_etl availability=implemented write=delete_section]; approval: Requires plan, preview, explicit approval, then execute.; risk: irreversible external deletion of a section (Asana requires the section be empty of tasks first); approval required; flags: --gid
  - tasks list - List Asana tasks with optional workspace, project, and assignee scopes. [intent=etl availability=implemented stream=tasks]; flags: --assignee, --project, --workspace
  - tasks create - Plan creation of an Asana task. [intent=reverse_etl availability=implemented write=create_task]; approval: Requires plan, preview, explicit approval, then execute.; risk: external mutation; creates a new task visible to the whole team/project; approval required; flags: --data, --name, --workspace
  - tasks update - Plan updates to an Asana task. [intent=reverse_etl availability=implemented write=update_task]; approval: Requires plan, preview, explicit approval, then execute.; risk: external mutation; overwrites task fields (e.g. can mark completed, reassign, reschedule); approval required; flags: --gid, --data
  - tasks delete - Plan deletion of an Asana task. [intent=reverse_etl availability=implemented write=delete_task]; approval: Requires plan, preview, explicit approval, then execute.; risk: irreversible external deletion of a task; approval required; flags: --gid
  - tasks comment - Plan an Asana task comment. [intent=reverse_etl availability=implemented write=add_comment]; approval: Requires plan, preview, explicit approval, then execute.; risk: external mutation; posts a comment visible to everyone with access to the task; approval required; flags: --task-gid, --data, --text
  - stories list - List stories for discovered Asana tasks. [intent=etl availability=implemented stream=stories]
- Metadata
  - tags list - List Asana tags, optionally scoped to a workspace. [intent=etl availability=implemented stream=tags]; flags: --workspace
  - tags create - Plan creation of an Asana tag. [intent=reverse_etl availability=implemented write=create_tag]; approval: Requires plan, preview, explicit approval, then execute.; risk: external mutation; creates a new workspace-visible tag; approval required; flags: --data, --name
  - tags update - Plan updates to an Asana tag. [intent=reverse_etl availability=implemented write=update_tag]; approval: Requires plan, preview, explicit approval, then execute.; risk: external mutation; renames/recolors a tag visible workspace-wide; approval required; flags: --gid, --data
  - tags delete - Plan deletion of an Asana tag. [intent=reverse_etl availability=implemented write=delete_tag]; approval: Requires plan, preview, explicit approval, then execute.; risk: irreversible external deletion of a tag, removed from every task that carries it; approval required; flags: --gid
  - custom-fields list - List custom fields in an Asana workspace. [intent=etl availability=implemented stream=custom_fields]; flags: --workspace

## Commands

### Inspect as a manual

```bash
pm connectors inspect asana
```

### Inspect as structured JSON

```bash
pm connectors inspect asana --json
```

## Agent Rules

- Run pm connectors inspect asana before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
