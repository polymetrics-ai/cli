# pm connectors inspect height

```text
NAME
  pm connectors inspect height - Height connector manual

SYNOPSIS
  pm connectors inspect height
  pm connectors inspect height --json
  pm credentials add <name> --connector height [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Height tasks, lists, field templates, users, and workspace through the Height REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  tasks:
    primary key: id
    cursor: createdAt
    fields: assigneesIds(array), completed(boolean), completedAt(string), createdAt(string), createdUserId(string), deleted(boolean), description(string), id(string), index(integer), lastActivityAt(string), listIds(array), model(string), name(string), parentTaskId(string), status(string), url(string)
  lists:
    primary key: id
    cursor: createdAt
    fields: createdAt(string), defaultList(boolean), description(string), id(string), key(string), model(string), name(string), type(string), updatedAt(string), url(string), userId(string), visualization(string)
  field_templates:
    primary key: id
    fields: archived(boolean), hidden(boolean), id(string), labels(array), model(string), name(string), required(boolean), standardType(string), type(string)
  users:
    primary key: id
    cursor: createdAt
    fields: admin(boolean), createdAt(string), deleted(boolean), email(string), firstname(string), id(string), key(string), lastname(string), model(string), signedUpAt(string), state(string), username(string)
  workspace:
    primary key: id
    cursor: createdAt
    fields: createdAt(string), createdUserId(string), frozen(boolean), id(string), key(string), model(string), name(string), url(string), urlType(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Height API read of task, list, field-template, user, and workspace data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect height

  # Inspect as structured JSON
  pm connectors inspect height --json

AGENT WORKFLOW
  - Run pm connectors inspect height before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
