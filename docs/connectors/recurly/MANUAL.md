# pm connectors inspect recurly

```text
NAME
  pm connectors inspect recurly - Recurly connector manual

SYNOPSIS
  pm connectors inspect recurly
  pm connectors inspect recurly --json
  pm credentials add <name> --connector recurly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Recurly accounts, subscriptions, invoices, transactions, and plans through the Recurly v3 REST API.

ICON
  asset: icons/recurly.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.recurly.com/api/v2021-02-25/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  accounts:
    primary key: id
    cursor: updated_at
    fields: code(), created_at(), email(), id(), state(), updated_at()
  subscriptions:
    primary key: id
    cursor: updated_at
    fields: account_id(), created_at(), id(), plan_id(), state(), updated_at()
  invoices:
    primary key: id
    cursor: created_at
    fields: account_id(), created_at(), id(), state(), total()
  transactions:
    primary key: id
    cursor: created_at
    fields: account_id(), amount(), created_at(), id(), status()
  plans:
    primary key: id
    cursor: updated_at
    fields: code(), id(), name(), state(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Recurly API read of subscription billing data
  approval: none; read-only billing API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect recurly

  # Inspect as structured JSON
  pm connectors inspect recurly --json

AGENT WORKFLOW
  - Run pm connectors inspect recurly before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
