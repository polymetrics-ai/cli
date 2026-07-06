# pm connectors inspect todoist

```text
NAME
  pm connectors inspect todoist - Todoist connector manual

SYNOPSIS
  pm connectors inspect todoist
  pm connectors inspect todoist --json
  pm credentials add <name> --connector todoist [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads projects, sections, tasks, comments, labels, and project collaborators, and writes project/section/task/comment/label create, update, and delete actions (plus task close/reopen), through the Todoist REST API.

ICON
  asset: icons/todoist.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.todoist.com/rest/v2/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  project_id
  task_id
  bearer_token (secret)
  token (secret)

ETL STREAMS
  projects:
    primary key: id
    fields: color(), id(), is_favorite(), name()
  sections:
    primary key: id
    fields: id(), name(), order(), project_id()
  tasks:
    primary key: id
    fields: content(), created_at(), description(), due(), id(), is_completed(), project_id(), section_id()
  comments:
    primary key: id
    fields: content(), id(), posted_at(), project_id(), task_id()
  labels:
    primary key: id
    fields: color(), id(), is_favorite(), name(), order()
  collaborators:
    primary key: id, project_id
    fields: email(), id(), name(), project_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_project:
    endpoint: POST /projects
    risk: creates a new project in the caller's Todoist account; low-risk external mutation, no approval required
  update_project:
    endpoint: POST /projects/{{ record.id }}
    required fields: id
    risk: mutates an existing project's name, color, favorite flag, or display style
  delete_project:
    endpoint: DELETE /projects/{{ record.id }}
    required fields: id
    risk: permanently removes a project and everything in it (its sections, tasks, and comments); irreversible
  create_section:
    endpoint: POST /sections
    risk: creates a new section within an existing project; low-risk external mutation, no approval required
  update_section:
    endpoint: POST /sections/{{ record.id }}
    required fields: id
    risk: renames an existing section
  delete_section:
    endpoint: DELETE /sections/{{ record.id }}
    required fields: id
    risk: permanently removes a section and every task in it; irreversible
  create_task:
    endpoint: POST /tasks
    risk: creates a new task in the caller's Todoist account (in the given project, or Inbox if omitted); low-risk external mutation, no approval required
  update_task:
    endpoint: POST /tasks/{{ record.id }}
    required fields: id
    risk: mutates an existing task's content, description, labels, priority, due date, assignee, or duration
  close_task:
    endpoint: POST /tasks/{{ record.id }}/close
    required fields: id
    risk: marks an existing task as completed (mirrors clicking the checkbox in the Todoist UI); recurring tasks advance to their next occurrence instead of disappearing
  reopen_task:
    endpoint: POST /tasks/{{ record.id }}/reopen
    required fields: id
    risk: reopens a previously completed task, returning it to the active task list
  delete_task:
    endpoint: DELETE /tasks/{{ record.id }}
    required fields: id
    risk: permanently removes a task; irreversible
  create_comment:
    endpoint: POST /comments
    risk: posts a new comment on a task or project; low-risk external mutation, no approval required
  update_comment:
    endpoint: POST /comments/{{ record.id }}
    required fields: id
    risk: edits the content of an existing comment
  delete_comment:
    endpoint: DELETE /comments/{{ record.id }}
    required fields: id
    risk: permanently removes a comment; irreversible
  create_label:
    endpoint: POST /labels
    risk: creates a new personal label; low-risk external mutation, no approval required
  update_label:
    endpoint: POST /labels/{{ record.id }}
    required fields: id
    risk: renames an existing label or changes its order/color/favorite flag; renaming changes how it appears on every task already using it
  delete_label:
    endpoint: DELETE /labels/{{ record.id }}
    required fields: id
    risk: permanently deletes a personal label; it is removed from every task that used it

SECURITY
  read risk: external Todoist API read of project, section, task, comment, label, and collaborator data
  write risk: external Todoist API mutation: creates, updates, or deletes projects/sections/tasks/comments/labels, and closes/reopens tasks
  approval: reverse ETL plan approval required before writes; delete actions on projects/sections/tasks/comments/labels are irreversible
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect todoist

  # Inspect as structured JSON
  pm connectors inspect todoist --json

AGENT WORKFLOW
  - Run pm connectors inspect todoist before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
