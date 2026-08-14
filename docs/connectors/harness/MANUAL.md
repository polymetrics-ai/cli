# pm connectors inspect harness

```text
NAME
  pm connectors inspect harness - Harness connector manual

SYNOPSIS
  pm connectors inspect harness
  pm connectors inspect harness --json
  pm credentials add <name> --connector harness [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Harness NextGen organizations, projects, services, connectors, and pipelines through the Harness platform REST API.

ICON
  id: harness
  asset: icons/harness.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id (required)
  base_url
  mode
  page_size
  api_key (secret) (required)

ETL STREAMS
  organizations:
    primary key: identifier
    fields: account_identifier(string), description(string), identifier(string), name(string)
  projects:
    primary key: identifier
    fields: account_identifier(string), color(string), description(string), identifier(string), modules(array), name(string), org_identifier(string)
  services:
    primary key: identifier
    fields: account_identifier(string), deleted(boolean), description(string), identifier(string), name(string), org_identifier(string), project_identifier(string)
  connectors:
    primary key: identifier
    fields: description(string), identifier(string), name(string), org_identifier(string), project_identifier(string), type(string)
  pipelines:
    primary key: identifier
    fields: description(string), identifier(string), name(string), org_identifier(string), project_identifier(string), stage_count(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Harness NextGen platform API read of organization/project/service/connector/pipeline metadata
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect harness

  # Inspect as structured JSON
  pm connectors inspect harness --json

AGENT WORKFLOW
  - Run pm connectors inspect harness before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
