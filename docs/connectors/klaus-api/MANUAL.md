# pm connectors inspect klaus-api

```text
NAME
  pm connectors inspect klaus-api - Klaus API connector manual

SYNOPSIS
  pm connectors inspect klaus-api
  pm connectors inspect klaus-api --json
  pm credentials add <name> --connector klaus-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Klaus (Zendesk QA) users and rating categories through the Klaus public REST API. The reviews stream is not yet migrated (ENGINE_GAP, see docs.md).

ICON
  asset: icons/klaus-api.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://help.klausapp.com/en/articles/2911907-klaus-api

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account
  base_url
  mode
  workspace
  api_key (secret)

ETL STREAMS
  users:
    primary key: id
    fields: email(), id(), name()
  categories:
    primary key: id
    fields: archived(), critical(), description(), groupId(), groupName(), groupPosition(), id(), maxRating(), name(), position(), rootCauses(), scorecards(), weight()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Klaus API read of user and quality-review configuration data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect klaus-api

  # Inspect as structured JSON
  pm connectors inspect klaus-api --json

AGENT WORKFLOW
  - Run pm connectors inspect klaus-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
