# pm connectors inspect clickup-api

```text
NAME
  pm connectors inspect clickup-api - ClickUp connector manual

SYNOPSIS
  pm connectors inspect clickup-api
  pm connectors inspect clickup-api --json
  pm credentials add <name> --connector clickup-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads ClickUp workspaces (teams), spaces, folders, lists, tasks, goals, space tags, and webhooks, and writes task/folder/list/space/webhook lifecycle mutations, task comments, tags, custom field values, and goal creation, through the ClickUp v2 REST API using a personal API token.

ICON
  id: clickup
  asset: icons/clickup.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://clickup.com/api/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  folder_id
  include_archived
  include_closed_tasks
  list_id
  mode
  space_id
  team_id
  api_token (secret) (required)

ETL STREAMS
  tasks:
    primary key: id
    cursor: date_updated
    fields: creator_id(integer), date_closed(string), date_created(string), date_updated(string), folder_id(string), id(string), list_id(string), name(string), space_id(string), status(string), url(string)
  teams:
    primary key: id
    fields: avatar(string), color(string), id(string), name(string)
  spaces:
    primary key: id
    fields: archived(boolean), id(string), multiple_assignees(boolean), name(string), private(boolean)
  folders:
    primary key: id
    fields: archived(boolean), hidden(boolean), id(string), name(string), orderindex(integer), space_id(string), task_count(string)
  lists:
    primary key: id
    fields: archived(boolean), id(string), name(string), orderindex(integer), space_id(string), task_count(integer)
  goals:
    primary key: id
    fields: archived(boolean), color(string), creator(integer), date_created(string), description(string), due_date(string), id(string), multiple_owners(boolean), name(string), percent_completed(integer), private(boolean), start_date(string), team_id(string)
  space_tags:
    primary key: name
    fields: name(string), space_id(string), tag_bg(string), tag_fg(string)
  webhooks:
    primary key: id
    fields: client_id(string), endpoint(string), events(array), folder_id(integer), health(object), id(string), list_id(integer), space_id(integer), task_id(string), team_id(integer), userid(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_task:
    endpoint: POST /list/{{ config.list_id }}/task
    required fields: name
    risk: creates a new ClickUp task in the configured list; low-risk (additive)
  update_task:
    endpoint: PUT /task/{{ record.id }}
    required fields: id
    risk: updates fields on an existing ClickUp task (name, description, status, dates, priority, archived); approval required
  delete_task:
    endpoint: DELETE /task/{{ record.id }}
    required fields: id
    risk: permanently deletes a ClickUp task; irreversible; approval required
  create_task_comment:
    endpoint: POST /task/{{ record.task_id }}/comment
    required fields: task_id, comment_text, notify_all
    optional fields: assignee, group_assignee
    risk: adds a new comment to a ClickUp task, visible to all task watchers when notify_all is true; low-risk
  add_tag_to_task:
    endpoint: POST /task/{{ record.task_id }}/tag/{{ record.tag_name }}
    required fields: task_id, tag_name
    risk: attaches an existing Space Tag to a task; low-risk
  remove_tag_from_task:
    endpoint: DELETE /task/{{ record.task_id }}/tag/{{ record.tag_name }}
    required fields: task_id, tag_name
    risk: removes a tag from a task (does not delete the tag from the Space); low-risk
  set_custom_field_value:
    endpoint: POST /task/{{ record.task_id }}/field/{{ record.field_id }}
    required fields: task_id, field_id, value
    optional fields: value_options
    risk: sets a Custom Field value on a task; the accepted value shape varies by the field's type (text/number/date/dropdown/label/people/task-relationship/manual-progress/location/button); approval required since an incorrectly-typed value can silently fail or corrupt a differently-typed field
  create_goal:
    endpoint: POST /team/{{ config.team_id }}/goal
    required fields: name
    risk: creates a new ClickUp Goal in the configured team/workspace; low-risk (additive)
  create_folder:
    endpoint: POST /space/{{ config.space_id }}/folder
    required fields: name
    risk: creates a new Folder in the configured space; low-risk (additive)
  update_folder:
    endpoint: PUT /folder/{{ record.id }}
    required fields: id, name
    risk: renames an existing ClickUp Folder; approval required
  delete_folder:
    endpoint: DELETE /folder/{{ record.id }}
    required fields: id
    risk: permanently deletes a ClickUp Folder and every List/task inside it; irreversible; approval required
  create_list:
    endpoint: POST /folder/{{ config.folder_id }}/list
    required fields: name
    risk: creates a new List in the configured Folder; low-risk (additive)
  update_list:
    endpoint: PUT /list/{{ record.id }}
    required fields: id, name
    risk: updates an existing ClickUp List's name/description/due date/priority/assignee/color; approval required
  delete_list:
    endpoint: DELETE /list/{{ record.id }}
    required fields: id
    risk: permanently deletes a ClickUp List and every task inside it; irreversible; approval required
  create_space:
    endpoint: POST /team/{{ config.team_id }}/space
    required fields: name
    risk: creates a new Space in the configured Workspace; low-risk (additive)
  update_space:
    endpoint: PUT /space/{{ record.id }}
    required fields: id, name
    risk: updates an existing ClickUp Space's name/color/privacy/ClickApp feature toggles; ClickUp's own docs mark every body field required (a partial update still needs the full current feature set re-sent to avoid resetting unspecified features); approval required
  delete_space:
    endpoint: DELETE /space/{{ record.id }}
    required fields: id
    risk: permanently deletes a ClickUp Space and every Folder/List/task inside it; irreversible; approval required
  create_webhook:
    endpoint: POST /team/{{ config.team_id }}/webhook
    required fields: endpoint, events
    risk: registers or repoints an outbound event-delivery URL of the caller's choosing; approval required
  update_webhook:
    endpoint: PUT /webhook/{{ record.id }}
    required fields: id
    risk: changes which events are delivered to (or repoints) an existing outbound webhook; approval required
  delete_webhook:
    endpoint: DELETE /webhook/{{ record.id }}
    required fields: id
    risk: stops event delivery to a registered webhook endpoint; approval required (irreversible without re-registering)

SECURITY
  read risk: external ClickUp API read of workspace, space, folder, list, task, goal, tag, and webhook data
  write risk: external mutation of ClickUp tasks, folders, lists, spaces, webhooks, tags, custom field values, and goals; delete_task/delete_folder/delete_list/delete_space are irreversible cascading deletes, and create_webhook/update_webhook register or repoint an outbound event-delivery URL of the caller's choosing — every write ships with an explicit per-action risk string
  approval: required for every delete_* action, every update_* action, create_webhook/update_webhook, and set_custom_field_value; create_task/create_task_comment/add_tag_to_task/remove_tag_from_task/create_goal/create_folder/create_list/create_space are low-risk (additive or non-destructive)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run ClickUp's declared typed write actions.
  Usage: pm clickup-api <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Reverse ETL writes
  Other Commands
    add tag to task apply - Typed action add_tag_to_task [intent=reverse_etl availability=partial write=add_tag_to_task]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /task/{task_id}/tag/{tag_name} disagrees with covered api_surface path /v2/task/{task_id}/tag/{tag_name}.; risk: attaches an existing Space Tag to a task; low-risk; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /task/{task_id}/tag/{tag_name} disagrees with covered api_surface path /v2/task/{task_id}/tag/{tag_name}.; flags: --tag-name (required), --task-id (required)
    create folder apply - Typed action create_folder [intent=reverse_etl availability=partial write=create_folder]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.space_id }}.; risk: creates a new Folder in the configured space; low-risk (additive); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.space_id }}.; flags: --name (required)
    create goal apply - Typed action create_goal [intent=reverse_etl availability=partial write=create_goal]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.team_id }}.; risk: creates a new ClickUp Goal in the configured team/workspace; low-risk (additive); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.team_id }}.; flags: --name (required)
    create list apply - Typed action create_list [intent=reverse_etl availability=partial write=create_list]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.folder_id }}.; risk: creates a new List in the configured Folder; low-risk (additive); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.folder_id }}.; flags: --name (required)
    create space apply - Typed action create_space [intent=reverse_etl availability=partial write=create_space]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.team_id }}.; risk: creates a new Space in the configured Workspace; low-risk (additive); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.team_id }}.; flags: --name (required)
    create task apply - Typed action create_task [intent=reverse_etl availability=partial write=create_task]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.list_id }}.; risk: creates a new ClickUp task in the configured list; low-risk (additive); notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.list_id }}.; flags: --name (required)
    create task comment apply - Typed action create_task_comment [intent=reverse_etl availability=partial write=create_task_comment]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /task/{task_id}/comment disagrees with covered api_surface path /v2/task/{task_id}/comment.; risk: adds a new comment to a ClickUp task, visible to all task watchers when notify_all is true; low-risk; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /task/{task_id}/comment disagrees with covered api_surface path /v2/task/{task_id}/comment.; flags: --comment-text (required), --notify-all (required), --task-id (required)
    create webhook apply - Typed action create_webhook [intent=reverse_etl availability=partial write=create_webhook]; approval: Blocked pending a faithful CLI record binding: declaration-pending: typed action path uses unrecognized placeholder {{ config.team_id }}.; risk: registers or repoints an outbound event-delivery URL of the caller's choosing; approval required; notes: Generated from the connector-owned typed action; declaration-pending: typed action path uses unrecognized placeholder {{ config.team_id }}.; flags: --endpoint (required), --events (required)
    delete folder apply - Typed action delete_folder [intent=reverse_etl availability=partial write=delete_folder]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /folder/{id} disagrees with covered api_surface path /v2/folder/{folder_id}.; risk: permanently deletes a ClickUp Folder and every List/task inside it; irreversible; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /folder/{id} disagrees with covered api_surface path /v2/folder/{folder_id}.; flags: --id (required)
    delete list apply - Typed action delete_list [intent=reverse_etl availability=partial write=delete_list]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /list/{id} disagrees with covered api_surface path /v2/list/{list_id}.; risk: permanently deletes a ClickUp List and every task inside it; irreversible; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /list/{id} disagrees with covered api_surface path /v2/list/{list_id}.; flags: --id (required)
    delete space apply - Typed action delete_space [intent=reverse_etl availability=partial write=delete_space]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /space/{id} disagrees with covered api_surface path /v2/space/{space_id}.; risk: permanently deletes a ClickUp Space and every Folder/List/task inside it; irreversible; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /space/{id} disagrees with covered api_surface path /v2/space/{space_id}.; flags: --id (required)
    delete task apply - Typed action delete_task [intent=reverse_etl availability=partial write=delete_task]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /task/{id} disagrees with covered api_surface path /v2/task/{task_id}.; risk: permanently deletes a ClickUp task; irreversible; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /task/{id} disagrees with covered api_surface path /v2/task/{task_id}.; flags: --id (required)
    delete webhook apply - Typed action delete_webhook [intent=reverse_etl availability=partial write=delete_webhook]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /webhook/{id} disagrees with covered api_surface path /v2/webhook/{webhook_id}.; risk: stops event delivery to a registered webhook endpoint; approval required (irreversible without re-registering); notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /webhook/{id} disagrees with covered api_surface path /v2/webhook/{webhook_id}.; flags: --id (required)
    remove tag from task apply - Typed action remove_tag_from_task [intent=reverse_etl availability=partial write=remove_tag_from_task]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /task/{task_id}/tag/{tag_name} disagrees with covered api_surface path /v2/task/{task_id}/tag/{tag_name}.; risk: removes a tag from a task (does not delete the tag from the Space); low-risk; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /task/{task_id}/tag/{tag_name} disagrees with covered api_surface path /v2/task/{task_id}/tag/{tag_name}.; flags: --tag-name (required), --task-id (required)
    set custom field value apply - Typed action set_custom_field_value [intent=reverse_etl availability=partial write=set_custom_field_value]; approval: Blocked pending a faithful CLI record binding: declaration-pending: the closed CLI flag set cannot faithfully represent required record field value.; risk: sets a Custom Field value on a task; the accepted value shape varies by the field's type (text/number/date/dropdown/label/people/task-relationship/manual-progress/location/button); approval required since an incorrectly-typed value can silently fail or corrupt a differently-typed field; notes: Generated from the connector-owned typed action; declaration-pending: the closed CLI flag set cannot faithfully represent required record field value.
    update folder apply - Typed action update_folder [intent=reverse_etl availability=partial write=update_folder]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /folder/{id} disagrees with covered api_surface path /v2/folder/{folder_id}.; risk: renames an existing ClickUp Folder; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /folder/{id} disagrees with covered api_surface path /v2/folder/{folder_id}.; flags: --id (required), --name (required)
    update list apply - Typed action update_list [intent=reverse_etl availability=partial write=update_list]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /list/{id} disagrees with covered api_surface path /v2/list/{list_id}.; risk: updates an existing ClickUp List's name/description/due date/priority/assignee/color; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /list/{id} disagrees with covered api_surface path /v2/list/{list_id}.; flags: --id (required), --name (required)
    update space apply - Typed action update_space [intent=reverse_etl availability=partial write=update_space]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /space/{id} disagrees with covered api_surface path /v2/space/{space_id}.; risk: updates an existing ClickUp Space's name/color/privacy/ClickApp feature toggles; ClickUp's own docs mark every body field required (a partial update still needs the full current feature set re-sent to avoid resetting unspecified features); approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /space/{id} disagrees with covered api_surface path /v2/space/{space_id}.; flags: --id (required), --name (required)
    update task apply - Typed action update_task [intent=reverse_etl availability=partial write=update_task]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /task/{id} disagrees with covered api_surface path /v2/task/{task_id}.; risk: updates fields on an existing ClickUp task (name, description, status, dates, priority, archived); approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /task/{id} disagrees with covered api_surface path /v2/task/{task_id}.; flags: --id (required)
    update webhook apply - Typed action update_webhook [intent=reverse_etl availability=partial write=update_webhook]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /webhook/{id} disagrees with covered api_surface path /v2/webhook/{webhook_id}.; risk: changes which events are delivered to (or repoints) an existing outbound webhook; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /webhook/{id} disagrees with covered api_surface path /v2/webhook/{webhook_id}.; flags: --id (required)

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

EXAMPLES
  # Inspect as a manual
  pm connectors inspect clickup-api

  # Inspect as structured JSON
  pm connectors inspect clickup-api --json

AGENT WORKFLOW
  - Run pm connectors inspect clickup-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
