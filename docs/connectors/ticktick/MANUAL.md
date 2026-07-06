# pm connectors inspect ticktick

```text
NAME
  pm connectors inspect ticktick - TickTick connector manual

SYNOPSIS
  pm connectors inspect ticktick
  pm connectors inspect ticktick --json
  pm credentials add <name> --connector ticktick [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads projects and project tasks, and writes task create/complete/delete actions, through the TickTick Open API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  project_id
  access_token (secret)
  bearer_token (secret)
  client_access_token (secret)

ETL STREAMS
  projects:
    primary key: id
    fields: closed(), color(), groupId(), id(), kind(), name(), permission(), sortOrder(), viewMode()
  tasks:
    primary key: id
    fields: completedTime(), content(), desc(), dueDate(), id(), isAllDay(), modifiedTime(), priority(), projectId(), reminders(), repeatFlag(), sortOrder(), startDate(), status(), timeZone(), title()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_task:
    endpoint: POST /task
    risk: creates a new task in the caller's TickTick account (in the given projectId, or the default Inbox if omitted); low-risk external mutation, no approval required
  complete_task:
    endpoint: POST /project/{{ record.projectId }}/task/{{ record.id }}/complete
    required fields: projectId, id
    risk: marks an existing task as completed; a completed task is removed from active task lists/reminders for every collaborator on the project
  delete_task:
    endpoint: DELETE /project/{{ record.projectId }}/task/{{ record.id }}
    required fields: projectId, id
    risk: permanently removes a task from the given project; irreversible, no undo via the API

SECURITY
  read risk: external TickTick API read of project and task data
  write risk: external TickTick API mutation: creates a task, marks a task complete, or permanently deletes a task
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ticktick

  # Inspect as structured JSON
  pm connectors inspect ticktick --json

AGENT WORKFLOW
  - Run pm connectors inspect ticktick before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
