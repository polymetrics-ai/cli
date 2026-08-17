# pm connectors inspect aha

```text
NAME
  pm connectors inspect aha - Aha! connector manual

SYNOPSIS
  pm connectors inspect aha
  pm connectors inspect aha --json
  pm credentials add <name> --connector aha [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Aha! features, products, ideas, releases, initiatives, goals, epics, and users through the Aha! REST API (read-only).

ICON
  asset: icons/aha.svg
  source: official
  review_status: official_verified
  review_url: https://www.aha.io/api

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  features:
    primary key: id
    cursor: updated_at
    fields: created_at(), due_date(), id(), name(), reference_num(), resource(), start_date(), updated_at(), url(), workflow_status()
  products:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), product_line(), reference_prefix(), resource(), updated_at(), url()
  ideas:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), reference_num(), resource(), score(), updated_at(), url(), votes(), workflow_status()
  releases:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), reference_num(), release_date(), released(), resource(), start_date(), updated_at(), url()
  initiatives:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), reference_num(), resource(), updated_at(), url(), workflow_status()
  goals:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), reference_num(), resource(), updated_at(), url(), workflow_status()
  epics:
    primary key: id
    cursor: updated_at
    fields: created_at(), due_date(), id(), name(), reference_num(), resource(), start_date(), updated_at(), url(), workflow_status()
  users:
    primary key: id
    fields: administrator(), email(), enabled(), id(), name(), resource()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Aha! API read of planning and roadmap data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect aha

  # Inspect as structured JSON
  pm connectors inspect aha --json

AGENT WORKFLOW
  - Run pm connectors inspect aha before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
