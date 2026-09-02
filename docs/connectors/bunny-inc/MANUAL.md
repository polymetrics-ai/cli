# pm connectors inspect bunny-inc

```text
NAME
  pm connectors inspect bunny-inc - Bunny, Inc. connector manual

SYNOPSIS
  pm connectors inspect bunny-inc
  pm connectors inspect bunny-inc --json
  pm credentials add <name> --connector bunny-inc [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Bunny subscription-billing data through declared per-tenant GraphQL connection routes.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  subdomain (required)
  apikey (secret) (required)

ETL STREAMS
  accounts:
    primary key: id
    cursor: updatedAt
    fields: id(string), updatedAt(string)
  contacts:
    primary key: id
    cursor: updatedAt
    fields: id(string), updatedAt(string)
  invoices:
    primary key: id
    cursor: updatedAt
    fields: id(string), updatedAt(string)
  payments:
    primary key: id
    cursor: updatedAt
    fields: id(string), updatedAt(string)
  subscriptions:
    primary key: id
    cursor: updatedAt
    fields: id(string), updatedAt(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: Bounded declared Bunny GraphQL reads use a source-validated tenant subdomain and bearer API key.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bunny-inc

  # Inspect as structured JSON
  pm connectors inspect bunny-inc --json

AGENT WORKFLOW
  - Run pm connectors inspect bunny-inc before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
