# pm connectors inspect gocardless

```text
NAME
  pm connectors inspect gocardless - GoCardless connector manual

SYNOPSIS
  pm connectors inspect gocardless
  pm connectors inspect gocardless --json
  pm credentials add <name> --connector gocardless [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes documented GoCardless REST API resources through the declarative connector engine.

ICON
  asset: icons/gocardless.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.gocardless.com/api-reference/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  balances_creditor_id
  bank_account_detail_id
  bank_account_holder_verification_id
  bank_authorisation_id
  base_url
  billing_request_id
  billing_request_institutions_country_code
  billing_request_template_id
  block_id
  creditor_bank_account_id
  creditor_id
  customer_bank_account_id
  customer_bank_account_token_id
  customer_id
  event_id
  export_id
  funds_availability_amount
  gc_key_id
  gocardless_version
  instalment_schedule_id
  mandate_id
  mandate_import_id
  outbound_payment_id
  outbound_payment_import_id
  payer_authorisation_id
  payment_account_id
  payment_account_transaction_id
  payment_account_transactions_value_date_from
  payment_account_transactions_value_date_to
  payment_id
  payout_id
  redirect_flow_id
  refund_id
  scheme_identifier_id
  start_date
  subscription_id
  tax_rate_id
  transferred_mandate_id
  verification_details_creditor_id
  webhook_id
  access_token (secret)

ETL STREAMS
  payments:
    primary key: id
    cursor: created_at
    fields: amount(), amount_refunded(), charge_date(), created_at(), currency(), description(), id(), mandate(), payout(), reference(), status()
  mandates:
    primary key: id
    cursor: created_at
    fields: created_at(), creditor(), customer_bank_account(), id(), next_possible_charge_date(), payments_require_approval(), reference(), scheme(), status()
  payouts:
    primary key: id
    cursor: created_at
    fields: amount(), arrival_date(), created_at(), creditor(), creditor_bank_account(), currency(), deducted_fees(), id(), payout_type(), reference(), status()
  refunds:
    primary key: id
    cursor: created_at
    fields: amount(), created_at(), currency(), id(), mandate(), payment(), reference()
  bank_authorisation:
    primary key: id
    fields: id()
  billing_requests:
    primary key: id
    fields: id()
  billing_request:
    primary key: id
    fields: id()
  billing_request_templates:
    primary key: id
    fields: id()
  billing_request_template:
    primary key: id
    fields: id()
  institutions:
    primary key: id
    fields: id()
  billing_request_institutions:
    primary key: id
    fields: id()
  balances:
  bank_account_detail:
    primary key: id
    fields: id()
  bank_account_holder_verification:
    primary key: id
    fields: id()
  block:
    primary key: id
    fields: id()
  blocks:
    primary key: id
    fields: id()
  creditors:
    primary key: id
    fields: id()
  creditor:
    primary key: id
    fields: id()
  creditor_bank_accounts:
    primary key: id
    fields: id()
  creditor_bank_account:
    primary key: id
    fields: id()
  currency_exchange_rates:
  customers:
    primary key: id
    fields: id()
  customer:
    primary key: id
    fields: id()
  customer_bank_accounts:
    primary key: id
    fields: id()
  customer_bank_account:
    primary key: id
    fields: id()
  events:
    primary key: id
    fields: id()
  event:
    primary key: id
    fields: id()
  export:
    primary key: id
    fields: id()
  exports:
    primary key: id
    fields: id()
  funds_availability:
  instalment_schedules:
    primary key: id
    fields: id()
  instalment_schedule:
    primary key: id
    fields: id()
  mandate:
    primary key: id
    fields: id()
  mandate_import:
    primary key: id
    fields: id()
  mandate_import_entries:
    primary key: id
    fields: id()
  negative_balance_limits:
    primary key: id
    fields: id()
  outbound_payment:
    primary key: id
    fields: id()
  outbound_payments:
    primary key: id
    fields: id()
  outbound_payment_stats:
  outbound_payment_import:
    primary key: id
    fields: id()
  outbound_payment_imports:
    primary key: id
    fields: id()
  outbound_payment_import_entries:
    primary key: id
    fields: id()
  payer_authorisation:
    primary key: id
    fields: id()
  payment:
    primary key: id
    fields: id()
  payment_account:
    primary key: id
    fields: id()
  payment_accounts:
    primary key: id
    fields: id()
  payment_account_transaction:
    primary key: id
    fields: id()
  payment_account_transactions_by_payment_account:
    primary key: id
    fields: id()
  payout:
    primary key: id
    fields: id()
  payout_items:
  redirect_flow:
    primary key: id
    fields: id()
  refund:
    primary key: id
    fields: id()
  scheme_identifiers:
    primary key: id
    fields: id()
  scheme_identifier:
    primary key: id
    fields: id()
  subscriptions:
    primary key: id
    fields: id()
  subscription:
    primary key: id
    fields: id()
  tax_rates:
    primary key: id
    fields: id()
  tax_rate:
    primary key: id
    fields: id()
  transferred_mandate:
    primary key: id
    fields: id()
  verification_details:
    primary key: id
    fields: id()
  webhooks:
    primary key: id
    fields: id()
  webhook:
    primary key: id
    fields: id()
  customer_bank_account_token:
    primary key: id
    fields: id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_a_bank_authorisation:
    endpoint: POST /bank_authorisations
    risk: GoCardless mutation: Create a Bank Authorisation.
  create_a_billing_request:
    endpoint: POST /billing_requests
    risk: GoCardless mutation: Create a Billing Request.
  collect_customer_details:
    endpoint: POST /billing_requests/{{ record.billing_request_id }}/actions/collect_customer_details
    required fields: billing_request_id
    risk: GoCardless mutation: Collect customer details.
  collect_bank_account_details:
    endpoint: POST /billing_requests/{{ record.billing_request_id }}/actions/collect_bank_account
    required fields: billing_request_id
    risk: GoCardless mutation: Collect bank account details.
  confirm_the_payer_details:
    endpoint: POST /billing_requests/{{ record.billing_request_id }}/actions/confirm_payer_details
    required fields: billing_request_id
    risk: GoCardless mutation: Confirm the payer details.
  fulfil_a_billing_request:
    endpoint: POST /billing_requests/{{ record.billing_request_id }}/actions/fulfil
    required fields: billing_request_id
    risk: GoCardless mutation: Fulfil a Billing Request.
  cancel_a_billing_request:
    endpoint: POST /billing_requests/{{ record.billing_request_id }}/actions/cancel
    required fields: billing_request_id
    risk: Destructive GoCardless mutation: Cancel a Billing Request.
  notify_the_customer:
    endpoint: POST /billing_requests/{{ record.billing_request_id }}/actions/notify
    required fields: billing_request_id
    risk: GoCardless mutation: Notify the customer.
  trigger_fallback:
    endpoint: POST /billing_requests/{{ record.billing_request_id }}/actions/fallback
    required fields: billing_request_id
    risk: GoCardless mutation: Trigger fallback.
  change_currency:
    endpoint: POST /billing_requests/{{ record.billing_request_id }}/actions/choose_currency
    required fields: billing_request_id
    risk: GoCardless mutation: Change currency.
  select_institution_for_a_billing_request:
    endpoint: POST /billing_requests/{{ record.billing_request_id }}/actions/select_institution
    required fields: billing_request_id
    risk: GoCardless mutation: Select institution for a Billing Request.
  create_a_billing_request_flow:
    endpoint: POST /billing_request_flows
    risk: GoCardless mutation: Create a Billing Request Flow.
  initialise_a_billing_request_flow:
    endpoint: POST /billing_request_flows/{{ record.billing_request_flow_id }}/actions/initialise
    required fields: billing_request_flow_id
    risk: GoCardless mutation: Initialise a Billing Request Flow.
  create_a_billing_request_template:
    endpoint: POST /billing_request_templates
    risk: GoCardless mutation: Create a Billing Request Template.
  update_a_billing_request_template:
    endpoint: PUT /billing_request_templates/{{ record.billing_request_template_id }}
    required fields: billing_request_template_id
    risk: GoCardless mutation: Update a Billing Request Template.
  create_a_billing_request_with_actions:
    endpoint: POST /billing_requests/create_with_actions
    risk: GoCardless mutation: Create a Billing Request with Actions.
  create_a_bank_account_holder_verification:
    endpoint: POST /bank_account_holder_verifications
    risk: GoCardless mutation: Create a bank account holder verification..
  create_a_block:
    endpoint: POST /blocks
    risk: GoCardless mutation: Create a block.
  disable_a_block:
    endpoint: POST /blocks/{{ record.block_id }}/actions/disable
    required fields: block_id
    risk: Destructive GoCardless mutation: Disable a block.
  enable_a_block:
    endpoint: POST /blocks/{{ record.block_id }}/actions/enable
    required fields: block_id
    risk: GoCardless mutation: Enable a block.
  create_blocks_by_reference:
    endpoint: POST /blocks/block_by_ref
    risk: GoCardless mutation: Create blocks by reference.
  create_a_creditor:
    endpoint: POST /creditors
    risk: GoCardless mutation: Create a creditor.
  update_a_creditor:
    endpoint: PUT /creditors/{{ record.creditor_id }}
    required fields: creditor_id
    risk: GoCardless mutation: Update a creditor.
  create_a_creditor_bank_account:
    endpoint: POST /creditor_bank_accounts
    risk: GoCardless mutation: Create a creditor bank account.
  disable_a_creditor_bank_account:
    endpoint: POST /creditor_bank_accounts/{{ record.creditor_bank_account_id }}/actions/disable
    required fields: creditor_bank_account_id
    risk: Destructive GoCardless mutation: Disable a creditor bank account.
  create_a_customer:
    endpoint: POST /customers
    risk: GoCardless mutation: Create a customer.
  update_a_customer:
    endpoint: PUT /customers/{{ record.customer_id }}
    required fields: customer_id
    risk: GoCardless mutation: Update a customer.
  remove_a_customer:
    endpoint: DELETE /customers/{{ record.customer_id }}
    required fields: customer_id
    risk: Destructive GoCardless mutation: Remove a customer.
  create_a_customer_bank_account:
    endpoint: POST /customer_bank_accounts
    risk: GoCardless mutation: Create a customer bank account.
  update_a_customer_bank_account:
    endpoint: PUT /customer_bank_accounts/{{ record.customer_bank_account_id }}
    required fields: customer_bank_account_id
    risk: GoCardless mutation: Update a customer bank account.
  disable_a_customer_bank_account:
    endpoint: POST /customer_bank_accounts/{{ record.customer_bank_account_id }}/actions/disable
    required fields: customer_bank_account_id
    risk: Destructive GoCardless mutation: Disable a customer bank account.
  handle_a_notification:
    endpoint: POST /customer_notifications/{{ record.customer_notification_id }}/actions/handle
    required fields: customer_notification_id
    risk: GoCardless mutation: Handle a notification.
  create_with_dates:
    endpoint: POST /instalment_schedules
    risk: GoCardless mutation: Create (with dates).
  update_an_instalment_schedule:
    endpoint: PUT /instalment_schedules/{{ record.instalment_schedule_id }}
    required fields: instalment_schedule_id
    risk: GoCardless mutation: Update an instalment schedule.
  cancel_an_instalment_schedule:
    endpoint: POST /instalment_schedules/{{ record.instalment_schedule_id }}/actions/cancel
    required fields: instalment_schedule_id
    risk: Destructive GoCardless mutation: Cancel an instalment schedule.
  create_a_logo_associated_with_a_creditor:
    endpoint: POST /branding/logos
    risk: GoCardless mutation: Create a logo associated with a creditor.
  create_a_mandate:
    endpoint: POST /mandates
    risk: GoCardless mutation: Create a mandate.
  update_a_mandate:
    endpoint: PUT /mandates/{{ record.mandate_id }}
    required fields: mandate_id
    risk: GoCardless mutation: Update a mandate.
  cancel_a_mandate:
    endpoint: POST /mandates/{{ record.mandate_id }}/actions/cancel
    required fields: mandate_id
    risk: Destructive GoCardless mutation: Cancel a mandate.
  reinstate_a_mandate:
    endpoint: POST /mandates/{{ record.mandate_id }}/actions/reinstate
    required fields: mandate_id
    risk: GoCardless mutation: Reinstate a mandate.
  create_a_new_mandate_import:
    endpoint: POST /mandate_imports
    risk: GoCardless mutation: Create a new mandate import.
  submit_a_mandate_import:
    endpoint: POST /mandate_imports/{{ record.mandate_import_id }}/actions/submit
    required fields: mandate_import_id
    risk: GoCardless mutation: Submit a mandate import.
  cancel_a_mandate_import:
    endpoint: POST /mandate_imports/{{ record.mandate_import_id }}/actions/cancel
    required fields: mandate_import_id
    risk: Destructive GoCardless mutation: Cancel a mandate import.
  add_a_mandate_import_entry:
    endpoint: POST /mandate_import_entries
    risk: GoCardless mutation: Add a mandate import entry.
  create_an_outbound_payment:
    endpoint: POST /outbound_payments
    risk: GoCardless mutation: Create an outbound payment.
  create_a_withdrawal_outbound_payment:
    endpoint: POST /outbound_payments/withdrawal
    risk: GoCardless mutation: Create a withdrawal outbound payment.
  cancel_an_outbound_payment:
    endpoint: POST /outbound_payments/{{ record.outbound_payment_id }}/actions/cancel
    required fields: outbound_payment_id
    risk: Destructive GoCardless mutation: Cancel an outbound payment.
  approve_an_outbound_payment:
    endpoint: POST /outbound_payments/{{ record.outbound_payment_id }}/actions/approve
    required fields: outbound_payment_id
    risk: GoCardless mutation: Approve an outbound payment.
  update_an_outbound_payment:
    endpoint: PUT /outbound_payments/{{ record.outbound_payment_id }}
    required fields: outbound_payment_id
    risk: GoCardless mutation: Update an outbound payment.
  create_an_outbound_payment_import:
    endpoint: POST /outbound_payment_imports
    risk: GoCardless mutation: Create an outbound payment import.
  create_a_payer_authorisation:
    endpoint: POST /payer_authorisations
    risk: GoCardless mutation: Create a Payer Authorisation.
  update_a_payer_authorisation:
    endpoint: PUT /payer_authorisations/{{ record.payer_authorisation_id }}
    required fields: payer_authorisation_id
    risk: GoCardless mutation: Update a Payer Authorisation.
  submit_a_payer_authorisation:
    endpoint: POST /payer_authorisations/{{ record.payer_authorisation_id }}/actions/submit
    required fields: payer_authorisation_id
    risk: GoCardless mutation: Submit a Payer Authorisation.
  confirm_a_payer_authorisation:
    endpoint: POST /payer_authorisations/{{ record.payer_authorisation_id }}/actions/confirm
    required fields: payer_authorisation_id
    risk: GoCardless mutation: Confirm a Payer Authorisation.
  create_a_payer_theme_associated_with_a_creditor:
    endpoint: POST /branding/payer_themes
    risk: GoCardless mutation: Create a payer theme associated with a creditor.
  create_a_payment:
    endpoint: POST /payments
    risk: GoCardless mutation: Create a payment.
  update_a_payment:
    endpoint: PUT /payments/{{ record.payment_id }}
    required fields: payment_id
    risk: GoCardless mutation: Update a payment.
  cancel_a_payment:
    endpoint: POST /payments/{{ record.payment_id }}/actions/cancel
    required fields: payment_id
    risk: Destructive GoCardless mutation: Cancel a payment.
  retry_a_payment:
    endpoint: POST /payments/{{ record.payment_id }}/actions/retry
    required fields: payment_id
    risk: GoCardless mutation: Retry a payment.
  update_a_payout:
    endpoint: PUT /payouts/{{ record.payout_id }}
    required fields: payout_id
    risk: GoCardless mutation: Update a payout.
  create_a_redirect_flow:
    endpoint: POST /redirect_flows
    risk: GoCardless mutation: Create a redirect flow.
  complete_a_redirect_flow:
    endpoint: POST /redirect_flows/{{ record.redirect_flow_id }}/actions/complete
    required fields: redirect_flow_id
    risk: GoCardless mutation: Complete a redirect flow.
  create_a_refund:
    endpoint: POST /refunds
    risk: GoCardless mutation: Create a refund.
  update_a_refund:
    endpoint: PUT /refunds/{{ record.refund_id }}
    required fields: refund_id
    risk: GoCardless mutation: Update a refund.
  simulate_a_scenario:
    endpoint: POST /scenario_simulators/{{ record.scenario }}/actions/run
    required fields: scenario
    risk: GoCardless mutation: Simulate a scenario.
  create_a_scheme_identifier:
    endpoint: POST /scheme_identifiers
    risk: GoCardless mutation: Create a scheme identifier.
  create_a_subscription:
    endpoint: POST /subscriptions
    risk: GoCardless mutation: Create a subscription.
  update_a_subscription:
    endpoint: PUT /subscriptions/{{ record.subscription_id }}
    required fields: subscription_id
    risk: GoCardless mutation: Update a subscription.
  pause_a_subscription:
    endpoint: POST /subscriptions/{{ record.subscription_id }}/actions/pause
    required fields: subscription_id
    risk: GoCardless mutation: Pause a subscription.
  resume_a_subscription:
    endpoint: POST /subscriptions/{{ record.subscription_id }}/actions/resume
    required fields: subscription_id
    risk: GoCardless mutation: Resume a subscription.
  cancel_a_subscription:
    endpoint: POST /subscriptions/{{ record.subscription_id }}/actions/cancel
    required fields: subscription_id
    risk: Destructive GoCardless mutation: Cancel a subscription.
  create_a_verification_detail:
    endpoint: POST /verification_details
    risk: GoCardless mutation: Create a verification detail.
  retry_a_webhook:
    endpoint: POST /webhooks/{{ record.webhook_id }}/actions/retry
    required fields: webhook_id
    risk: GoCardless mutation: Retry a webhook.
  perform_a_bank_details_lookup:
    endpoint: POST /bank_details_lookups
    risk: GoCardless mutation: Perform a bank details lookup.
  create_a_mandate_pdf:
    endpoint: POST /mandate_pdfs
    risk: GoCardless mutation: Create a mandate PDF.
  create_a_customer_bank_account_token:
    endpoint: POST /customer_bank_account_tokens
    risk: GoCardless mutation: Create a customer bank account token.

SECURITY
  read risk: external GoCardless API read of payments, mandates, payouts, refunds, and other documented REST resources
  write risk: external GoCardless API mutations including creates, updates, cancellations, retries, simulations, and customer/account changes
  approval: required for every write action; destructive actions are marked confirm: destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gocardless

  # Inspect as structured JSON
  pm connectors inspect gocardless --json

AGENT WORKFLOW
  - Run pm connectors inspect gocardless before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
