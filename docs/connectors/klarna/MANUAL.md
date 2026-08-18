# pm connectors inspect klarna

```text
NAME
  pm connectors inspect klarna - Klarna connector manual

SYNOPSIS
  pm connectors inspect klarna
  pm connectors inspect klarna --json
  pm credentials add <name> --connector klarna [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Klarna settlement payouts and transactions through the Klarna Settlements API.

ICON
  id: klarna
  asset: icons/klarna.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.klarna.com/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url (required)
  mode
  payment_references
  summary_currency_code
  summary_end_date
  summary_start_date
  password (secret) (required)
  username (secret) (required)

ETL STREAMS
  payouts:
    primary key: payout_reference
    fields: currency_code(string), merchant_settlement_type(string), payment_reference(string), payout_reference(string), settlement_amount(integer), totals(object)
  transactions:
    primary key: transaction_id
    fields: amount(integer), capture_date(string), capture_id(string), currency_code(string), merchant_reference1(string), merchant_reference2(string), order_id(string), payout_reference(string), sale_date(string), short_order_id(string), transaction_id(string), type(string)
  payout_details:
    primary key: payment_reference
    fields: currency_code(string), currency_code_of_registration_country(string), merchant_id(string), merchant_settlement_type(string), payment_reference(string), payout_date(string), totals(object), transactions(string)
  payout_summaries:
    primary key: summary_settlement_currency, summary_payout_date_start, summary_payout_date_end
    fields: summary_payout_date_end(string), summary_payout_date_start(string), summary_settlement_currency(string), summary_total_commission_amount(integer), summary_total_commission_reversal_amount(integer), summary_total_fee_amount(integer), summary_total_fee_correction_amount(integer), summary_total_holdback_amount(integer), summary_total_release_amount(integer), summary_total_repay_amount(integer), summary_total_return_amount(integer), summary_total_reversal_amount(integer), summary_total_sale_amount(integer), summary_total_settlement_amount(integer), summary_total_tax_amount(integer)
  payout_summary:
    primary key: payout_reference
    fields: currency_code(string), fee_amount(integer), payout_reference(string), return_amount(integer), sale_amount(integer), settlement_amount(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: external Klarna Settlements API read of payout and transaction data
  approval: none; read-only source
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect klarna

  # Inspect as structured JSON
  pm connectors inspect klarna --json

AGENT WORKFLOW
  - Run pm connectors inspect klarna before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
