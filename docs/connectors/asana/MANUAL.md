# pm connectors inspect asana

```text
NAME
  pm connectors inspect asana - Asana connector manual

SYNOPSIS
  pm connectors inspect asana
  pm connectors inspect asana --json
  pm credentials add <name> --connector asana [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Asana workspaces, projects, tasks, sections, tags, stories, users, teams, custom fields, project statuses, and team/workspace memberships through the Asana v1 REST API. Writes task/project/section/tag create-update-delete and task comments.

ICON
  asset: icons/asana.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.asana.com/reference/rest-api-reference

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  assignee
  base_url
  mode
  project_id
  team_id
  workspace_id
  access_token (secret)

ETL STREAMS
  workspaces:
    primary key: gid
    fields: gid(), name(), resource_type()
  projects:
    primary key: gid
    fields: created_at(), gid(), modified_at(), name(), resource_type()
  tasks:
    primary key: gid
    fields: completed(), created_at(), gid(), modified_at(), name(), resource_type()
  users:
    primary key: gid
    fields: email(), gid(), name(), resource_type()
  teams:
    primary key: gid
    fields: description(), gid(), name(), resource_type(), visibility()
  tags:
    primary key: gid
    fields: color(), created_at(), gid(), name(), notes(), resource_type()
  sections:
    primary key: gid
    fields: created_at(), gid(), name(), project_gid(), resource_type()
  stories:
    primary key: gid
    fields: created_at(), gid(), resource_subtype(), resource_type(), task_gid(), text(), type()
  custom_fields:
    primary key: gid
    fields: description(), enabled(), gid(), name(), resource_type(), type()
  project_statuses:
    primary key: gid
    fields: color(), created_at(), gid(), modified_at(), project_gid(), resource_type(), text(), title()
  team_memberships:
    primary key: gid
    fields: gid(), is_admin(), is_guest(), is_limited_access(), resource_type()
  workspace_memberships:
    primary key: gid
    fields: gid(), is_active(), is_admin(), is_guest(), resource_type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_task:
    endpoint: POST /tasks
    risk: external mutation; creates a new task visible to the whole team/project; low risk, no approval required
  update_task:
    endpoint: PUT /tasks/{{ record.gid }}
    required fields: gid
    risk: external mutation; overwrites task fields (e.g. can mark completed, reassign, reschedule); approval required
  delete_task:
    endpoint: DELETE /tasks/{{ record.gid }}
    required fields: gid
    risk: irreversible external deletion of a task; approval required
  create_project:
    endpoint: POST /projects
    risk: external mutation; creates a new project visible to the team/workspace; low risk, no approval required
  update_project:
    endpoint: PUT /projects/{{ record.gid }}
    required fields: gid
    risk: external mutation; overwrites project fields (can archive, reassign owner); approval required
  delete_project:
    endpoint: DELETE /projects/{{ record.gid }}
    required fields: gid
    risk: irreversible external deletion of a project and its association with its tasks; approval required
  create_section:
    endpoint: POST /projects/{{ record.project_gid }}/sections
    required fields: project_gid
    risk: external mutation; creates a new section in a project's board/list view; low risk, no approval required
  update_section:
    endpoint: PUT /sections/{{ record.gid }}
    required fields: gid
    risk: external mutation; renames a section; low risk, no approval required
  delete_section:
    endpoint: DELETE /sections/{{ record.gid }}
    required fields: gid
    risk: irreversible external deletion of a section (Asana requires the section be empty of tasks first); approval required
  create_tag:
    endpoint: POST /tags
    risk: external mutation; creates a new workspace-visible tag; low risk, no approval required
  update_tag:
    endpoint: PUT /tags/{{ record.gid }}
    required fields: gid
    risk: external mutation; renames/recolors a tag visible workspace-wide; low risk, no approval required
  delete_tag:
    endpoint: DELETE /tags/{{ record.gid }}
    required fields: gid
    risk: irreversible external deletion of a tag, removed from every task that carries it; approval required
  add_comment:
    endpoint: POST /tasks/{{ record.task_gid }}/stories
    required fields: task_gid
    risk: external mutation; posts a comment visible to everyone with access to the task; approval required

SECURITY
  read risk: external Asana API read of workspace, project, task, and PM metadata
  write risk: external mutations: creates/updates/deletes tasks, projects, sections, and tags, and posts task comments
  approval: required for every write action; deletes are flagged destructive/irreversible in writes.json's per-action risk field
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect asana

  # Inspect as structured JSON
  pm connectors inspect asana --json

AGENT WORKFLOW
  - Run pm connectors inspect asana before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
