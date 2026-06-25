# pm connectors inspect source-stripe

```text
NAME
  pm connectors inspect source-stripe - Stripe connector manual

SYNOPSIS
  pm connectors inspect source-stripe
  pm connectors inspect source-stripe --json
  pm credentials add <name> --connector source-stripe [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Stripe native Go source connector. Runtime family: declarative_http_go. Documentation: https://docs.airbyte.com/integrations/sources/stripe.

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

IMPLEMENTATION STATUS
  implementation_status: enabled
  runtime_kind: declarative_http_go
  pm connector: stripe
  notes: Implemented as the built-in Stripe connector on the connsdk declarative-HTTP template.
  upstream image reference: airbyte/source-stripe:6.0.7 (metadata only; not executed)

RUNTIME CAPABILITIES
  metadata=true
  check=true
  catalog=true
  read=true
  write=true
  query=false
  etl=true
  reverse_etl=true

NATIVE PORT PLAN
  family: native_saas
  priority_wave: 0
  etl_operations: catalog, check, read, write
  reverse_etl_operations: preview, validate_write, write
  conformance: catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, secret_redaction, spec, state_checkpoint, write_validation

OFFICIAL APPLICATION DOCUMENTATION
  Stripe API reference: https://stripe.com/docs/api
  Stripe authentication: https://stripe.com/docs/api/authentication
  API changelog: https://docs.stripe.com/changelog
  API upgrades: https://docs.stripe.com/upgrades
  Stripe API changelog: https://stripe.com/docs/upgrades
  Stripe rate limits: https://stripe.com/docs/rate-limits
  Stripe API OpenAPI specification: https://github.com/stripe/openapi
  Stripe Status: https://status.stripe.com/
  Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/stripe

AUTHENTICATION
  fixture: Fixture-backed conformance mode; no credentials required.
    supports: read=true write=true
  api_key: Live Stripe secret API key via Bearer auth.
    secrets: client_secret
    supports: read=true write=true

CONFIGURATION
  base_url default=https://api.stripe.com/v1: Stripe API base URL override for tests or proxies.
  account_id: Optional Stripe account ID; sent as the Stripe-Account header for Connect.
  start_date: RFC3339 lower bound; only objects created at or after this time are read.
  page_size default=100: Records per page (1-100).
  max_pages default=0: Maximum pages; use 0, all, or unlimited to exhaust the stream.
  mode: Runtime mode: live (default) or fixture for credential-free conformance.
  client_secret (secret) (required): Stripe secret API key (sk_...). Used only for Bearer auth; never logged.

ETL STREAMS
  customers: Stripe customers.
    primary key: id
    cursor: created
    fields: id(string), object(string), created(integer), email(string), name(string), description(string), phone(string), currency(string), balance(integer), delinquent(boolean), livemode(boolean)
  charges: Stripe charges.
    primary key: id
    cursor: created
    fields: id(string), object(string), created(integer), amount(integer), amount_captured(integer), amount_refunded(integer), currency(string), customer(string), status(string), paid(boolean), refunded(boolean), livemode(boolean)
  invoices: Stripe invoices.
    primary key: id
    cursor: created
    fields: id(string), object(string), created(integer), customer(string), subscription(string), status(string), currency(string), amount_due(integer), amount_paid(integer), amount_remaining(integer), total(integer), paid(boolean), livemode(boolean)
  subscriptions: Stripe subscriptions.
    primary key: id
    cursor: created
    fields: id(string), object(string), created(integer), customer(string), status(string), currency(string), current_period_start(integer), current_period_end(integer), cancel_at_period_end(boolean), canceled_at(integer), livemode(boolean)
  products: Stripe products.
    primary key: id
    cursor: created
    fields: id(string), object(string), created(integer), updated(integer), name(string), description(string), active(boolean), type(string), livemode(boolean)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
  Source modes: full_refresh, incremental

REVERSE ETL ACTIONS
  create_customer: Create a Stripe customer.
    endpoint: POST /customers
    required fields: email
    optional fields: name, description, phone
    risk: external mutation; approval required
  update_customer: Update an existing Stripe customer addressed by id.
    endpoint: POST /customers/{id}
    required fields: id
    optional fields: email, name, description, phone
    risk: external mutation; approval required

PAGINATION
  type: id_cursor
  page size field: page_size
  page limit field: max_pages
  default limit: 0

SECURITY
  read risk: external Stripe API read of customer and billing data
  write risk: external Stripe API mutation
  mutation risk: creates or updates Stripe customers through allow-listed reverse ETL actions
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect source-stripe

  # Inspect as structured JSON
  pm connectors inspect source-stripe --json

AGENT WORKFLOW
  - This catalog slug delegates to the enabled pm connector stripe.
  - Run pm connectors inspect source-stripe before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

SEE ALSO
  Stripe connector documentation: https://docs.airbyte.com/integrations/sources/stripe
  Stripe API reference: https://stripe.com/docs/api
  Stripe authentication: https://stripe.com/docs/api/authentication
  API changelog: https://docs.stripe.com/changelog
  API upgrades: https://docs.stripe.com/upgrades
  Stripe API changelog: https://stripe.com/docs/upgrades
  Stripe rate limits: https://stripe.com/docs/rate-limits
  Stripe API OpenAPI specification: https://github.com/stripe/openapi
  Stripe Status: https://status.stripe.com/

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
