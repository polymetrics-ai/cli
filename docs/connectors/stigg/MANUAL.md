# pm connectors inspect stigg

```text
NAME
  pm connectors inspect stigg - Stigg connector manual

SYNOPSIS
  pm connectors inspect stigg
  pm connectors inspect stigg --json
  pm credentials add <name> --connector stigg [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Stigg products, plans, customers, and subscriptions through the Stigg GraphQL-over-HTTP API. Read-only.

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
  api_key (secret)

ETL STREAMS
  products:
    primary key: id
    fields: displayName(), id(), refId(), status()
  plans:
    primary key: id
    fields: displayName(), id(), refId(), status()
  customers:
    primary key: id
    fields: displayName(), id(), refId(), status()
  subscriptions:
    primary key: id
    fields: customerId(), id(), refId(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Stigg GraphQL API read of product/plan/customer/subscription entitlement metadata
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect stigg

  # Inspect as structured JSON
  pm connectors inspect stigg --json

AGENT WORKFLOW
  - Run pm connectors inspect stigg before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
