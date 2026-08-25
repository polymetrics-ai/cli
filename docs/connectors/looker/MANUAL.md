# pm connectors inspect looker

```text
NAME
  pm connectors inspect looker - Looker connector manual

SYNOPSIS
  pm connectors inspect looker
  pm connectors inspect looker --json
  pm credentials add <name> --connector looker [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Looker users, groups, folders, looks, and dashboards through the Looker API 4.0.

ICON
  id: looker
  asset: icons/looker.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://cloud.google.com/looker/docs/reference/looker-api/latest

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  mode
  token_url
  access_token (secret)
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  users:
    primary key: id
    fields: display_name(string), email(string), id(string)
  groups:
    primary key: id
    fields: id(string), name(string)
  folders:
    primary key: id
    fields: id(string), name(string)
  looks:
    primary key: id
    fields: folder_id(string), id(string), title(string)
  dashboards:
    primary key: id
    fields: folder_id(string), id(string), title(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Looker API read of users, groups, folders, looks, and dashboards
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Looker's declared typed write actions.
  Usage: pm looker <command> [flags]

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

EXAMPLES
  # Inspect as a manual
  pm connectors inspect looker

  # Inspect as structured JSON
  pm connectors inspect looker --json

AGENT WORKFLOW
  - Run pm connectors inspect looker before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
