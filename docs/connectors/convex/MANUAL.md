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
  asset: icons/convex.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.convex.dev/http-api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  deployment_url
  mode
  table
  access_key (secret)

ETL STREAMS
  tables:
    primary key: name
    fields: name()
  documents:
    primary key: id
    fields: _id(), id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
