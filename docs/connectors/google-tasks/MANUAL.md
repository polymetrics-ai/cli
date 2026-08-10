# pm connectors inspect google-tasks

```text
NAME
  pm connectors inspect google-tasks - Google Tasks connector manual

SYNOPSIS
  pm connectors inspect google-tasks
  pm connectors inspect google-tasks --json
  pm credentials add <name> --connector google-tasks [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google task lists and tasks through the Google Tasks REST API.

ICON
  id: simple-icons-googletasks
  asset: icons/simple-icons/googletasks.svg
  title: Google Tasks
  simple_icon_slug: googletasks
  simple_icon_hex: 2684FC
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Google%20Tasks
  match: exact-name-or-slug
  matched_by: google-tasks

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  records_limit
  api_key (secret)

ETL STREAMS
  tasklists:
    primary key: id
    cursor: updated
    fields: etag(string), id(string), kind(string), self_link(string), title(string), updated(string)
  tasks:
    primary key: id
    cursor: updated
    fields: completed(string), deleted(boolean), due(string), etag(string), hidden(boolean), id(string), kind(string), notes(string), parent(string), position(string), self_link(string), status(string), tasklist_id(string), title(string), updated(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Google Tasks API read of the authenticated user's task lists and tasks
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Google Tasks's declared streams and reverse-ETL actions.
  Usage: pm google-tasks <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete tasks v1 lists tasklist tasks task - Documented DELETE /tasks/v1/lists/{tasklist}/tasks/{task} (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.delete.tasks-v1-lists-tasklist-tasks-task]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete tasks v1 users me lists tasklist - Documented DELETE /tasks/v1/users/@me/lists/{tasklist} (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.delete.tasks-v1-users-me-lists-tasklist]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get tasks v1 lists tasklist tasks - Documented GET /tasks/v1/lists/{tasklist}/tasks (not implemented) [intent=direct_read availability=not_implemented operation=google-tasks.get.tasks-v1-lists-tasklist-tasks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get tasks v1 lists tasklist tasks task - Documented GET /tasks/v1/lists/{tasklist}/tasks/{task} (not implemented) [intent=direct_read availability=not_implemented operation=google-tasks.get.tasks-v1-lists-tasklist-tasks-task]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get tasks v1 users me lists - Documented GET /tasks/v1/users/@me/lists (not implemented) [intent=direct_read availability=not_implemented operation=google-tasks.get.tasks-v1-users-me-lists]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get tasks v1 users me lists tasklist - Documented GET /tasks/v1/users/@me/lists/{tasklist} (not implemented) [intent=direct_read availability=not_implemented operation=google-tasks.get.tasks-v1-users-me-lists-tasklist]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch tasks v1 lists tasklist tasks task - Documented PATCH /tasks/v1/lists/{tasklist}/tasks/{task} (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.patch.tasks-v1-lists-tasklist-tasks-task]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch tasks v1 users me lists tasklist - Documented PATCH /tasks/v1/users/@me/lists/{tasklist} (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.patch.tasks-v1-users-me-lists-tasklist]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post tasks v1 lists tasklist clear - Documented POST /tasks/v1/lists/{tasklist}/clear (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.post.tasks-v1-lists-tasklist-clear]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post tasks v1 lists tasklist tasks - Documented POST /tasks/v1/lists/{tasklist}/tasks (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.post.tasks-v1-lists-tasklist-tasks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post tasks v1 lists tasklist tasks task move - Documented POST /tasks/v1/lists/{tasklist}/tasks/{task}/move (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.post.tasks-v1-lists-tasklist-tasks-task-move]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post tasks v1 users me lists - Documented POST /tasks/v1/users/@me/lists (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.post.tasks-v1-users-me-lists]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put tasks v1 lists tasklist tasks task - Documented PUT /tasks/v1/lists/{tasklist}/tasks/{task} (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.put.tasks-v1-lists-tasklist-tasks-task]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put tasks v1 users me lists tasklist - Documented PUT /tasks/v1/users/@me/lists/{tasklist} (not implemented) [intent=direct_write availability=not_implemented operation=google-tasks.put.tasks-v1-users-me-lists-tasklist]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    tasklists list - Run the tasklists ETL stream [intent=etl availability=implemented stream=tasklists]; notes: discrepancy=present-in-surface-absent-from-artifact
    tasks list - Run the tasks ETL stream [intent=etl availability=implemented stream=tasks]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-tasks

  # Inspect as structured JSON
  pm connectors inspect google-tasks --json

AGENT WORKFLOW
  - Run pm connectors inspect google-tasks before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
