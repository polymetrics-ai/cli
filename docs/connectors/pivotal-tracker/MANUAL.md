# pm connectors inspect pivotal-tracker

```text
NAME
  pm connectors inspect pivotal-tracker - Pivotal Tracker connector manual

SYNOPSIS
  pm connectors inspect pivotal-tracker
  pm connectors inspect pivotal-tracker --json
  pm credentials add <name> --connector pivotal-tracker [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Pivotal Tracker projects, stories, iterations, and epics through API v5.

ICON
  asset: icons/pivotal-tracker.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  project_id
  api_token (secret)

ETL STREAMS
  projects:
    primary key: id
    fields: id(), name(), state(), updated_at()
  stories:
    primary key: id
    fields: id(), name(), state(), updated_at()
  iterations:
    primary key: id
    fields: id(), name(), state(), updated_at()
  epics:
    primary key: id
    fields: id(), name(), state(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Pivotal Tracker API read of project, story, iteration, and epic data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pivotal-tracker

  # Inspect as structured JSON
  pm connectors inspect pivotal-tracker --json

AGENT WORKFLOW
  - Run pm connectors inspect pivotal-tracker before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
