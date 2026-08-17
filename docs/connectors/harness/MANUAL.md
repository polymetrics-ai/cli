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
  asset: icons/harness.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  base_url
  mode
  page_size
  api_key (secret)

ETL STREAMS
  organizations:
    primary key: identifier
    fields: account_identifier(), description(), identifier(), name()
  projects:
    primary key: identifier
    fields: account_identifier(), color(), description(), identifier(), modules(), name(), org_identifier()
  services:
    primary key: identifier
    fields: account_identifier(), deleted(), description(), identifier(), name(), org_identifier(), project_identifier()
  connectors:
    primary key: identifier
    fields: description(), identifier(), name(), org_identifier(), project_identifier(), type()
  pipelines:
    primary key: identifier
    fields: description(), identifier(), name(), org_identifier(), project_identifier(), stage_count()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
