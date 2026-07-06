# pm connectors inspect commercetools

```text
NAME
  pm connectors inspect commercetools - commercetools connector manual

SYNOPSIS
  pm connectors inspect commercetools
  pm connectors inspect commercetools --json
  pm credentials add <name> --connector commercetools [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads commercetools customers, orders, and products through the HTTP API.

ICON
  asset: icons/commercetools.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.commercetools.com/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  project_key
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: createdAt
    fields: addresses(), authenticationMode(), createdAt(), customerNumber(), email(), firstName(), id(), isEmailVerified(), lastModifiedAt(), lastName(), version()
  orders:
    primary key: id
    cursor: createdAt
    fields: createdAt(), customerId(), id(), lastModifiedAt(), lineItems(), orderNumber(), orderState(), totalPrice(), version()
  products:
    primary key: id
    cursor: createdAt
    fields: createdAt(), id(), lastModifiedAt(), masterData(), productType(), version()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external commercetools API read of customer, order, and product data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect commercetools

  # Inspect as structured JSON
  pm connectors inspect commercetools --json

AGENT WORKFLOW
  - Run pm connectors inspect commercetools before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
