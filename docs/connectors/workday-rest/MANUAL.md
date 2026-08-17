# pm connectors inspect workday-rest

```text
NAME
  pm connectors inspect workday-rest - Workday REST connector manual

SYNOPSIS
  pm connectors inspect workday-rest
  pm connectors inspect workday-rest --json
  pm credentials add <name> --connector workday-rest [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Workday REST API resources (workers, organizations, job profiles) with bearer-token authentication. Read-only.

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
  tenant
  access_token (secret)

ETL STREAMS
  workers:
    primary key: id
    cursor: updated
    fields: descriptor(), id(), updated()
  organizations:
    primary key: id
    cursor: updated
    fields: descriptor(), id(), type()
  jobs:
    primary key: id
    cursor: updated
    fields: descriptor(), id(), updated()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Workday REST API read of worker, organization, and job profile data (HR/PII-adjacent)
  approval: none; read-only, bearer-token auth
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect workday-rest

  # Inspect as structured JSON
  pm connectors inspect workday-rest --json

AGENT WORKFLOW
  - Run pm connectors inspect workday-rest before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
