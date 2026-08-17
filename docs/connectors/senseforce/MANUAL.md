# pm connectors inspect senseforce

```text
NAME
  pm connectors inspect senseforce - Senseforce connector manual

SYNOPSIS
  pm connectors inspect senseforce
  pm connectors inspect senseforce --json
  pm credentials add <name> --connector senseforce [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads records from a configured Senseforce dataset through the Senseforce API.

ICON
  asset: icons/senseforce.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  backend_url
  dataset_id
  access_token (secret)

ETL STREAMS
  records:
    primary key: id
    fields: Timestamp(), id(), value()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Senseforce API read of a configured dataset's rows
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect senseforce

  # Inspect as structured JSON
  pm connectors inspect senseforce --json

AGENT WORKFLOW
  - Run pm connectors inspect senseforce before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
