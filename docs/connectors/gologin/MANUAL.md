# pm connectors inspect gologin

```text
NAME
  pm connectors inspect gologin - GoLogin connector manual

SYNOPSIS
  pm connectors inspect gologin
  pm connectors inspect gologin --json
  pm credentials add <name> --connector gologin [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads GoLogin browser profiles, folders, tags, and account information through the GoLogin REST API.

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
  api_key (secret)

ETL STREAMS
  profiles:
    primary key: id
    cursor: updatedAt
    fields: browserType(), createdAt(), folderName(), id(), name(), notes(), os(), role(), updatedAt()
  folders:
    primary key: id
    fields: id(), name(), profilesCount()
  user:
    primary key: _id
    cursor: createdAt
    fields: _id(), createdAt(), email(), firstName(), lastName(), plan(), profilesCount()
  tags:
    primary key: _id
    fields: _id(), color(), field(), title()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external GoLogin API read of browser profile and account data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gologin

  # Inspect as structured JSON
  pm connectors inspect gologin --json

AGENT WORKFLOW
  - Run pm connectors inspect gologin before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
