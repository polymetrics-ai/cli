# pm connectors inspect braintree

```text
NAME
  pm connectors inspect braintree - Braintree connector manual

SYNOPSIS
  pm connectors inspect braintree
  pm connectors inspect braintree --json
  pm credentials add <name> --connector braintree [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Braintree transactions, customers, subscriptions, reference data, payment methods, disputes, merchant accounts, and Apple Pay domains through the gateway HTTP API.

ICON
  asset: icons/braintree.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.paypal.com/braintree/docs/reference/general/server-sdk-deprecation-policy

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  merchant_id
  mode
  page_size
  private_key (secret)
  public_key (secret)

ETL STREAMS
  transactions:
    primary key: id
    fields: amount(), id(), status()
  customers:
    primary key: id
    fields: amount(), id(), status()
  subscriptions:
    primary key: id
    fields: amount(), id(), status()
  add_ons:
    primary key: id
    fields: amount(), id(), kind(), name()
  discounts:
    primary key: id
    fields: amount(), id(), kind(), name()
  plans:
    primary key: id
    fields: billing_frequency(), currency_iso_code(), id(), name(), price()
  merchant_accounts:
    primary key: id
    fields: currency_iso_code(), default(), id(), status()
  payment_methods:
    primary key: token
    fields: customer_id(), default(), payment_instrument_type(), token()
  disputes:
    primary key: id
    fields: amount(), id(), reason(), status()
  apple_pay_domains:
    primary key: domain
    fields: domain(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Braintree API read of transaction, customer, subscription, reference, dispute, payment method, and merchant account data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect braintree

  # Inspect as structured JSON
  pm connectors inspect braintree --json

AGENT WORKFLOW
  - Run pm connectors inspect braintree before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
