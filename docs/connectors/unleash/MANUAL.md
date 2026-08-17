# pm connectors inspect unleash

```text
NAME
  pm connectors inspect unleash - Unleash connector manual

SYNOPSIS
  pm connectors inspect unleash
  pm connectors inspect unleash --json
  pm credentials add <name> --connector unleash [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Unleash projects, feature toggles, environments, and segments through admin API list endpoints.

ICON
  asset: icons/unleash.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.getunleash.io/reference/api/unleash

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  project_id
  api_token (secret)

ETL STREAMS
  projects:
    primary key: id
    fields: id(), name()
  features:
    primary key: name
    fields: enabled(), name(), project(), type()
  environments:
    primary key: id
    fields: id(), name()
  segments:
    primary key: id
    fields: id(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Unleash admin API read of project, feature toggle, environment, and segment data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect unleash

  # Inspect as structured JSON
  pm connectors inspect unleash --json

AGENT WORKFLOW
  - Run pm connectors inspect unleash before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
