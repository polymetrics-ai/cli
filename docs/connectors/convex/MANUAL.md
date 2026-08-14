# pm connectors inspect convex

```text
NAME
  pm connectors inspect convex - Convex connector manual

SYNOPSIS
  pm connectors inspect convex
  pm connectors inspect convex --json
  pm credentials add <name> --connector convex [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Convex tables and documents through the deployment HTTP API.

ICON
  id: convex
  asset: icons/convex.svg
  source: official
  review_status: official_verified
  review_url: https://docs.convex.dev/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  deployment_url (required)
  mode
  table
  access_key (secret) (required)

ETL STREAMS
  tables:
    primary key: name
    fields: name(string)
  documents:
    primary key: id
    fields: _id(string), id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Convex deployment API read of table metadata and documents
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect convex

  # Inspect as structured JSON
  pm connectors inspect convex --json

AGENT WORKFLOW
  - Run pm connectors inspect convex before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
