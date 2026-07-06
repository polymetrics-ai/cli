# pm connectors inspect missive

```text
NAME
  pm connectors inspect missive - Missive connector manual

SYNOPSIS
  pm connectors inspect missive
  pm connectors inspect missive --json
  pm credentials add <name> --connector missive [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Missive contacts, contact groups, users, teams, and shared labels through the Missive REST API.

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
  kind
  api_key (secret)

ETL STREAMS
  contacts:
    primary key: id
    fields: first_name(), id(), last_name(), modified_at()
  contact_groups:
    primary key: id
    fields: id(), kind(), name()
  users:
    primary key: id
    fields: email(), id(), name()
  teams:
    primary key: id
    fields: id(), name(), organization()
  shared_labels:
    primary key: id
    fields: color(), id(), name(), name_with_parent_names()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Missive API read of contact, user, team, and label data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect missive

  # Inspect as structured JSON
  pm connectors inspect missive --json

AGENT WORKFLOW
  - Run pm connectors inspect missive before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
