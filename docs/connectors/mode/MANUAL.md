# pm connectors inspect mode

```text
NAME
  pm connectors inspect mode - Mode connector manual

SYNOPSIS
  pm connectors inspect mode
  pm connectors inspect mode --json
  pm credentials add <name> --connector mode [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Mode collections (spaces), reports, data sources, groups, and memberships through the Mode REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
  mode
  workspace
  api_secret (secret)
  api_token (secret)

ETL STREAMS
  spaces:
    primary key: token
    cursor: updated_at
    fields: created_at(), description(), id(), name(), restricted(), space_type(), state(), token(), updated_at()
  reports:
    primary key: token
    cursor: updated_at
    fields: account_username(), archived(), created_at(), description(), id(), last_run_at(), name(), public(), space_token(), token(), updated_at()
  data_sources:
    primary key: token
    cursor: updated_at
    fields: adapter(), asleep(), created_at(), database(), description(), host(), id(), name(), public(), queryable(), token(), updated_at()
  groups:
    primary key: token
    cursor: updated_at
    fields: created_at(), description(), id(), name(), state(), token(), updated_at()
  memberships:
    primary key: token
    cursor: created_at
    fields: admin(), created_at(), email(), id(), token(), username()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Mode API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect mode

  # Inspect as structured JSON
  pm connectors inspect mode --json

AGENT WORKFLOW
  - Run pm connectors inspect mode before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
