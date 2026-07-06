# pm connectors inspect hellobaton

```text
NAME
  pm connectors inspect hellobaton - Hellobaton connector manual

SYNOPSIS
  pm connectors inspect hellobaton
  pm connectors inspect hellobaton --json
  pm credentials add <name> --connector hellobaton [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Hellobaton projects, milestones, tasks, phases, companies, and users through the Hellobaton REST API.

ICON
  asset: icons/hellobaton.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  projects:
    primary key: id
    cursor: modified
    fields: _self(), annual_contract_value(), archived(), completed_datetime(), cost(), created(), creator(), id(), modified(), name()
  milestones:
    primary key: id
    cursor: modified
    fields: _self(), created(), deadline_datetime(), deadline_fixed(), description(), duration(), finish_datetime(), id(), modified(), project()
  tasks:
    primary key: id
    cursor: modified
    fields: _self(), created(), description(), id(), modified(), name(), project()
  phases:
    primary key: id
    cursor: modified
    fields: _self(), created(), id(), modified(), name()
  companies:
    primary key: id
    cursor: modified
    fields: _self(), created(), id(), modified(), name()
  users:
    primary key: id
    cursor: modified
    fields: _self(), created(), first_name(), id(), last_name(), modified()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Hellobaton API read of project, milestone, task, phase, company, and user data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect hellobaton

  # Inspect as structured JSON
  pm connectors inspect hellobaton --json

AGENT WORKFLOW
  - Run pm connectors inspect hellobaton before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
