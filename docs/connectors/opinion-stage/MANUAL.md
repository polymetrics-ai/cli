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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  items:
    primary key: id
    fields: created(), embed(), id(), links(), modified(), relationships(), status(), title(), type()
  responses:
    primary key: id
    fields: answers(), created(), duration(), id(), item_id(), links(), result(), result_text(), result_title(), type(), utm()
  questions:
    primary key: id
    fields: created(), id(), item_id(), kind(), lead(), modified(), title(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
