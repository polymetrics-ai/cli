# pm connectors inspect qualaroo

```text
NAME
  pm connectors inspect qualaroo - Qualaroo connector manual

SYNOPSIS
  pm connectors inspect qualaroo
  pm connectors inspect qualaroo --json
  pm credentials add <name> --connector qualaroo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Qualaroo nudges and reporting response records through the Qualaroo API. Read-only.

ICON
  asset: icons/qualaroo.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://help.qualaroo.com/hc/en-us/articles/201969438-The-Qualaroo-API

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  survey_id
  api_key (secret)
  api_secret (secret)

ETL STREAMS
  nudges:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), status(), updated_at()
  responses:
    primary key: id
    cursor: created_at
    fields: created_at(), email(), id(), nudge_id(), updated_at()
  survey_responses:
    primary key: id
    cursor: time
    fields: answered_questions(), id(), identity(), ip_address(), page(), properties(), referrer(), time(), token(), user_agent()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Qualaroo API read of survey nudge and reporting response data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect qualaroo

  # Inspect as structured JSON
  pm connectors inspect qualaroo --json

AGENT WORKFLOW
  - Run pm connectors inspect qualaroo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
