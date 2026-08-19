# pm connectors inspect stripe

```text
NAME
  pm connectors inspect stripe - Stripe connector manual

SYNOPSIS
  pm connectors inspect stripe
  pm connectors inspect stripe --json
  pm credentials add <name> --connector stripe [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Stripe customers, charges, invoices, subscriptions, and products, and writes approved reverse ETL customer create, update, and typed destructive delete actions through the Stripe REST API.

ICON
  id: stripe
  asset: icons/stripe.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://stripe.com/docs/api

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  base_url
  max_pages
  mode
  page_size
  start_date
  client_secret (secret) (required)

ETL STREAMS
  customers:
    primary key: id
    cursor: created
    fields: balance(integer), created(integer), currency(string), delinquent(boolean), description(string), email(string), id(string), livemode(boolean), name(string), object(string), phone(string)
  charges:
    primary key: id
    cursor: created
    fields: amount(integer), amount_captured(integer), amount_refunded(integer), created(integer), currency(string), customer(string), id(string), livemode(boolean), object(string), paid(boolean), refunded(boolean), status(string)
  invoices:
    primary key: id
    cursor: created
    fields: amount_due(integer), amount_paid(integer), amount_remaining(integer), created(integer), currency(string), customer(string), id(string), livemode(boolean), object(string), paid(boolean), status(string), subscription(string), total(integer)
  subscriptions:
    primary key: id
    cursor: created
    fields: cancel_at_period_end(boolean), canceled_at(integer), created(integer), currency(string), current_period_end(integer), current_period_start(integer), customer(string), id(string), livemode(boolean), object(string), status(string)
  products:
    primary key: id
    cursor: created
    fields: active(boolean), created(integer), description(string), id(string), livemode(boolean), name(string), object(string), type(string), updated(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers
    risk: external mutation; approval required
  update_customer:
    endpoint: POST /customers/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_customer:
    endpoint: DELETE /customers/{{ record.id }}
    required fields: id
    risk: destructive external mutation; deletes a Stripe customer; reverse ETL approval and typed destructive confirmation required

SECURITY
  read risk: external Stripe API read of customer and billing data
  write risk: external Stripe API mutation, including typed destructive customer deletion when explicitly confirmed
  approval: reverse ETL plan approval required before writes; destructive actions require typed confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read Stripe billing streams and safely plan customer write actions.
  Usage: pm stripe <command> [flags]
  Source CLI: Stripe API (OpenAPI spec3 2026-07-29.dahlia)
  Global flags:
    --credential (string): Credential name to use for the Stripe request.
    --connection (string): Credential name alias used only when --credential is omitted; does not resolve pm connections.
    --config (string_array): Connector config override as key=value; never pass secret values here.
    --json (boolean): Emit machine-readable JSON output.
    --limit (integer): Maximum records to emit from stream commands.
    --plan (string): Execute an approved reverse-ETL plan by id.
    --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
    --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
  Customers
    customers list - Read Stripe customers through the declared ETL stream. [intent=etl availability=implemented stream=customers]
    customers create - Plan creation of a Stripe customer through reverse ETL. [intent=reverse_etl availability=implemented write=create_customer]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Creates customer data in Stripe; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --email, --name, --description, --phone
    customers update - Plan an update to a Stripe customer through reverse ETL. [intent=reverse_etl availability=implemented write=update_customer]; approval: Plan first, inspect preview output, then run only with the generated approval token.; risk: Mutates customer data in Stripe; requires reverse ETL plan, preview, explicit approval, then execute.; flags: --id, --email, --name, --description, --phone
    customers delete - Plan deletion of a Stripe customer with typed destructive confirmation and customer ID redaction. [intent=reverse_etl availability=implemented write=delete_customer]; approval: Plan first, inspect preview output, then run only with the generated approval token and typed --confirm destructive challenge.; risk: Destructive Stripe customer deletion; redacts the customer ID from previews and write errors; requires reverse ETL plan, preview, explicit approval, and --confirm destructive before execute.; flags: --id
  Billing streams
    charges list - Read Stripe charges through the declared ETL stream. [intent=etl availability=implemented stream=charges]
    invoices list - Read Stripe invoices through the declared ETL stream. [intent=etl availability=implemented stream=invoices]
    subscriptions list - Read Stripe subscriptions through the declared ETL stream. [intent=etl availability=implemented stream=subscriptions]
    products list - Read Stripe products through the declared ETL stream. [intent=etl availability=implemented stream=products]
  Help topics:
    destructive-confirmation - Stripe DELETE/destructive operations are in scope only through typed destructive confirmation and the reverse ETL plan → preview → approval → execute path.

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

EXAMPLES
  # Inspect as a manual
  pm connectors inspect stripe

  # Inspect as structured JSON
  pm connectors inspect stripe --json

AGENT WORKFLOW
  - Run pm connectors inspect stripe before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
