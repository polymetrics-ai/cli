# pm connectors inspect gong

```text
NAME
  pm connectors inspect gong - Gong connector manual

SYNOPSIS
  pm connectors inspect gong
  pm connectors inspect gong --json
  pm credentials add <name> --connector gong [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Gong users, calls, and scorecards through the Gong REST API (read-only).

ICON
  asset: icons/gong.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://us-66463.app.gong.io/settings/api/documentation

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  start_date
  access_key (secret)
  access_key_secret (secret)

ETL STREAMS
  users:
    primary key: id
    cursor: created
    fields: active(), created(), email_address(), first_name(), id(), last_name(), manager_id(), phone_number(), title()
  calls:
    primary key: id
    cursor: started
    fields: direction(), duration(), id(), is_private(), language(), media(), scheduled(), scope(), started(), system(), title(), url()
  scorecards:
    primary key: scorecardId
    cursor: updated
    fields: created(), enabled(), scorecardId(), scorecardName(), updated(), workspaceId()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Gong API read of call, user, and scorecard data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gong

  # Inspect as structured JSON
  pm connectors inspect gong --json

AGENT WORKFLOW
  - Run pm connectors inspect gong before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
