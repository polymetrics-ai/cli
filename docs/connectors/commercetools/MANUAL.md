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
  id: commercetools
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
  base_url (required)
  mode
  project_key (required)
  token_url (required)
  client_id (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  customers:
    primary key: id
    cursor: createdAt
    fields: addresses(array), authenticationMode(string), createdAt(string), customerNumber(string), email(string), firstName(string), id(string), isEmailVerified(boolean), lastModifiedAt(string), lastName(string), version(integer)
  orders:
    primary key: id
    cursor: createdAt
    fields: createdAt(string), customerId(string), id(string), lastModifiedAt(string), lineItems(array), orderNumber(string), orderState(string), totalPrice(object), version(integer)
  products:
    primary key: id
    cursor: createdAt
    fields: createdAt(string), id(string), lastModifiedAt(string), masterData(object), productType(object), version(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external commercetools API read of customer, order, and product data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run commercetools's declared typed write actions.
  Usage: pm commercetools <command> [flags]

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

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
