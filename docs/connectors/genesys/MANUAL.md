# pm connectors inspect genesys

```text
NAME
  pm connectors inspect genesys - Genesys connector manual

SYNOPSIS
  pm connectors inspect genesys
  pm connectors inspect genesys --json
  pm credentials add <name> --connector genesys [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Genesys Cloud users, queues, groups, and divisions through the Genesys Cloud Platform API.

ICON
  id: genesys
  asset: icons/genesys.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.genesys.cloud/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  max_pages
  mode
  page_size
  scope
  token_url (required)
  client_id (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  users:
    primary key: id
    fields: display_name(string), email(string), id(string), name(string), state(string)
  queues:
    primary key: id
    fields: description(string), id(string), name(string)
  groups:
    primary key: id
    fields: description(string), id(string), name(string)
  divisions:
    primary key: id
    fields: description(string), id(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Genesys Cloud Platform API read of user, queue, group, and division data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect genesys

  # Inspect as structured JSON
  pm connectors inspect genesys --json

AGENT WORKFLOW
  - Run pm connectors inspect genesys before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
