# pm connectors inspect pipeliner

```text
NAME
  pm connectors inspect pipeliner - Pipeliner connector manual

SYNOPSIS
  pm connectors inspect pipeliner
  pm connectors inspect pipeliner --json
  pm credentials add <name> --connector pipeliner [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Pipeliner CRM accounts, contacts, opportunities, and leads through the REST API.

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
  mode
  space_id
  password (secret)
  username (secret)

ETL STREAMS
  accounts:
    primary key: id
    fields: id(), name(), status(), updated_at()
  contacts:
    primary key: id
    fields: id(), name(), status(), updated_at()
  opportunities:
    primary key: id
    fields: id(), name(), status(), updated_at()
  leads:
    primary key: id
    fields: id(), name(), status(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Pipeliner CRM API read of account, contact, opportunity, and lead data
  approval: none; read-only CRM sync
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pipeliner

  # Inspect as structured JSON
  pm connectors inspect pipeliner --json

AGENT WORKFLOW
  - Run pm connectors inspect pipeliner before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
