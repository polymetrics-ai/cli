# pm connectors inspect opinion-stage

```text
NAME
  pm connectors inspect opinion-stage - Opinion Stage connector manual

SYNOPSIS
  pm connectors inspect opinion-stage
  pm connectors inspect opinion-stage --json
  pm credentials add <name> --connector opinion-stage [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Opinion Stage items (polls, quizzes, and forms) through the Opinion Stage Public Result API. Read-only.

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
  items:
    primary key: id
    fields: created(string), embed(object), id(string), links(object), modified(string), relationships(object), status(string), title(string), type(string)
  responses:
    primary key: id
    fields: answers(array), created(string), duration(number), id(string), item_id(string), links(object), result(object), result_text(string), result_title(string), type(string), utm(object)
  questions:
    primary key: id
    fields: created(string), id(string), item_id(string), kind(string), lead(boolean), modified(string), title(string), type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Opinion Stage API read of item directory
  approval: none; read-only API-key access
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect opinion-stage

  # Inspect as structured JSON
  pm connectors inspect opinion-stage --json

AGENT WORKFLOW
  - Run pm connectors inspect opinion-stage before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
