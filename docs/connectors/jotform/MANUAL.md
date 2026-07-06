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
  forms:
    primary key: id
    cursor: created_at
    fields: count(), created_at(), id(), last_submission(), new(), status(), title(), type(), updated_at(), url(), username()
  submissions:
    primary key: id
    cursor: created_at
    fields: answers(), created_at(), flag(), form_id(), id(), ip(), new(), notes(), status(), updated_at()
  reports:
    primary key: id
    cursor: created_at
    fields: created_at(), fields(), form_id(), id(), status(), title(), type(), updated_at(), url()
  folders:
    primary key: id
    fields: color(), forms(), id(), name(), owner(), parent(), subfolders()
  user:
    primary key: username
    fields: account_type(), created_at(), email(), name(), status(), time_zone(), updated_at(), usage(), username()

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
