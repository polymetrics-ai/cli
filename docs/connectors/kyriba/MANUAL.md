# pm connectors inspect kyriba

```text
NAME
  pm connectors inspect kyriba - Kyriba connector manual

SYNOPSIS
  pm connectors inspect kyriba
  pm connectors inspect kyriba --json
  pm credentials add <name> --connector kyriba [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Kyriba bank accounts, transactions, statements, and payments through tenant REST API collection endpoints. Read-only.

ICON
  asset: icons/kyriba.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.kyriba.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  scope
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  bank_accounts:
    primary key: id
    fields: account_number(), currency(), id(), status()
  transactions:
    primary key: id
    fields: account_number(), amount(), currency(), id(), status()
  statements:
    primary key: id
    fields: account_number(), currency(), id(), status()
  payments:
    primary key: id
    fields: amount(), currency(), id(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Kyriba tenant REST API read of bank accounts/transactions/statements/payments
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect kyriba

  # Inspect as structured JSON
  pm connectors inspect kyriba --json

AGENT WORKFLOW
  - Run pm connectors inspect kyriba before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
