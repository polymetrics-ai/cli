# pm connectors inspect signnow

```text
NAME
  pm connectors inspect signnow - signNow connector manual

SYNOPSIS
  pm connectors inspect signnow
  pm connectors inspect signnow --json
  pm credentials add <name> --connector signnow [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads signNow documents, templates, and users through the signNow REST API.

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
  page_size
  access_token (secret)

ETL STREAMS
  documents:
    primary key: id
    cursor: updated_at
    fields: id(), name(), updated_at()
  templates:
    primary key: id
    cursor: updated_at
    fields: id(), name(), updated_at()
  users:
    primary key: id
    cursor: updated_at
    fields: email(), id(), name(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external signNow API read of document, template, and user data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect signnow

  # Inspect as structured JSON
  pm connectors inspect signnow --json

AGENT WORKFLOW
  - Run pm connectors inspect signnow before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
