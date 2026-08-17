# pm connectors inspect workday

```text
NAME
  pm connectors inspect workday - Workday connector manual

SYNOPSIS
  pm connectors inspect workday
  pm connectors inspect workday --json
  pm credentials add <name> --connector workday [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Workday tenant data (workers, organizations, positions) through conservative Workday API endpoints. Read-only.

ICON
  asset: icons/workday.svg
  source: official
  review_status: official_verified
  review_url: https://community.workday.com/sites/default/files/file-hosting/productionapi/index.html

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  tenant
  password (secret)
  username (secret)

ETL STREAMS
  workers:
    primary key: id
    cursor: updated_at
    fields: id(), name(), updated_at()
  organizations:
    primary key: id
    cursor: updated_at
    fields: id(), name(), type(), updated_at()
  positions:
    primary key: id
    cursor: updated_at
    fields: id(), title(), updated_at(), worker_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Workday tenant API read of worker, organization, and position data (HR/PII-adjacent)
  approval: none; read-only, HTTP Basic auth
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect workday

  # Inspect as structured JSON
  pm connectors inspect workday --json

AGENT WORKFLOW
  - Run pm connectors inspect workday before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
