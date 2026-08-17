# pm connectors inspect partnerize

```text
NAME
  pm connectors inspect partnerize - Partnerize connector manual

SYNOPSIS
  pm connectors inspect partnerize
  pm connectors inspect partnerize --json
  pm credentials add <name> --connector partnerize [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Partnerize conversions, campaigns, and publishers through the REST API.

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
  application_key (secret)
  user_api_key (secret)

ETL STREAMS
  conversions:
    primary key: id
    cursor: created_at
    fields: created_at(), currency(), id(), status(), value()
  campaigns:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name(), status()
  publishers:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Partnerize API read of conversion, campaign, and publisher data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect partnerize

  # Inspect as structured JSON
  pm connectors inspect partnerize --json

AGENT WORKFLOW
  - Run pm connectors inspect partnerize before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
