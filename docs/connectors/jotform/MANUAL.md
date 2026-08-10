# pm connectors inspect jotform

```text
NAME
  pm connectors inspect jotform - Jotform connector manual

SYNOPSIS
  pm connectors inspect jotform
  pm connectors inspect jotform --json
  pm credentials add <name> --connector jotform [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Jotform forms, submissions, reports, folders, and the account profile through the Jotform REST API.

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
  forms:
    primary key: id
    cursor: created_at
    fields: count(string), created_at(string), id(string), last_submission(string), new(string), status(string), title(string), type(string), updated_at(string), url(string), username(string)
  submissions:
    primary key: id
    cursor: created_at
    fields: answers(object), created_at(string), flag(string), form_id(string), id(string), ip(string), new(string), notes(string), status(string), updated_at(string)
  reports:
    primary key: id
    cursor: created_at
    fields: created_at(string), fields(string), form_id(string), id(string), status(string), title(string), type(string), updated_at(string), url(string)
  folders:
    primary key: id
    fields: color(string), forms(object), id(string), name(string), owner(string), parent(string), subfolders(object)
  user:
    primary key: username
    fields: account_type(string), created_at(string), email(string), name(string), status(string), time_zone(string), updated_at(string), usage(string), username(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Jotform API read of form, submission, report, and folder data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect jotform

  # Inspect as structured JSON
  pm connectors inspect jotform --json

AGENT WORKFLOW
  - Run pm connectors inspect jotform before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
