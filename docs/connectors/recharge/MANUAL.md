# pm connectors inspect recharge

```text
NAME
  pm connectors inspect recharge - Recharge connector manual

SYNOPSIS
  pm connectors inspect recharge
  pm connectors inspect recharge --json
  pm credentials add <name> --connector recharge [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Recharge customers, subscriptions, and orders through the Recharge REST API.

ICON
  asset: icons/recharge.svg
  source: official
  review_status: official_verified
  review_url: https://docs.getrecharge.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  api_version
  base_url
  mode
  access_token (secret)

ETL STREAMS
  customers:
    primary key: id
    fields: created_at(), email(), id(), updated_at()
  subscriptions:
    primary key: id
    fields: created_at(), customer_id(), id(), status(), updated_at()
  orders:
    primary key: id
    fields: created_at(), customer_id(), id(), status(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Recharge API read of customer, subscription, and order data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect recharge

  # Inspect as structured JSON
  pm connectors inspect recharge --json

AGENT WORKFLOW
  - Run pm connectors inspect recharge before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
