# Overview

Reads and writes documented GoCardless REST API resources through the connector engine.

Readable streams: `payments`, `mandates`, `payouts`, `refunds`, `bank_authorisation`,
`billing_requests`, `billing_request`, `billing_request_templates`, `billing_request_template`,
`institutions`, `billing_request_institutions`, `balances`, `bank_account_detail`,
`bank_account_holder_verification`, `block`, `blocks`, `creditors`, `creditor`,
`creditor_bank_accounts`, `creditor_bank_account`, `currency_exchange_rates`, `customers`,
`customer`, `customer_bank_accounts`, `customer_bank_account`, `events`, `event`, `export`,
`exports`, `funds_availability`, `instalment_schedules`, `instalment_schedule`, `mandate`,
`mandate_import`, `mandate_import_entries`, `negative_balance_limits`, `outbound_payment`,
`outbound_payments`, `outbound_payment_stats`, `outbound_payment_import`,
`outbound_payment_imports`, `outbound_payment_import_entries`, `payer_authorisation`, `payment`,
`payment_account`, `payment_accounts`, `payment_account_transaction`,
`payment_account_transactions_by_payment_account`, `payout`, `payout_items`, `redirect_flow`,
`refund`, `scheme_identifiers`, `scheme_identifier`, `subscriptions`, `subscription`, `tax_rates`,
`tax_rate`, `transferred_mandate`, `verification_details`, `webhooks`, `webhook`,
`customer_bank_account_token`.

Write actions: `create_a_bank_authorisation`, `create_a_billing_request`,
`collect_customer_details`, `collect_bank_account_details`, `confirm_the_payer_details`,
`fulfil_a_billing_request`, `cancel_a_billing_request`, `notify_the_customer`, `trigger_fallback`,
`change_currency`, `select_institution_for_a_billing_request`, `create_a_billing_request_flow`,
`initialise_a_billing_request_flow`, `create_a_billing_request_template`,
`update_a_billing_request_template`, `create_a_billing_request_with_actions`,
`create_a_bank_account_holder_verification`, `create_a_block`, `disable_a_block`, `enable_a_block`,
`create_blocks_by_reference`, `create_a_creditor`, `update_a_creditor`,
`create_a_creditor_bank_account`, `disable_a_creditor_bank_account`, `create_a_customer`,
`update_a_customer`, `remove_a_customer`, `create_a_customer_bank_account`,
`update_a_customer_bank_account`, `disable_a_customer_bank_account`, `handle_a_notification`,
`create_with_dates`, `update_an_instalment_schedule`, `cancel_an_instalment_schedule`,
`create_a_logo_associated_with_a_creditor`, `create_a_mandate`, `update_a_mandate`,
`cancel_a_mandate`, `reinstate_a_mandate`, `create_a_new_mandate_import`, `submit_a_mandate_import`,
`cancel_a_mandate_import`, `add_a_mandate_import_entry`, `create_an_outbound_payment`,
`create_a_withdrawal_outbound_payment`, `cancel_an_outbound_payment`, `approve_an_outbound_payment`,
`update_an_outbound_payment`, `create_an_outbound_payment_import`, `create_a_payer_authorisation`,
`update_a_payer_authorisation`, `submit_a_payer_authorisation`, `confirm_a_payer_authorisation`,
`create_a_payer_theme_associated_with_a_creditor`, `create_a_payment`, `update_a_payment`,
`cancel_a_payment`, `retry_a_payment`, `update_a_payout`, `create_a_redirect_flow`,
`complete_a_redirect_flow`, `create_a_refund`, `update_a_refund`, `simulate_a_scenario`,
`create_a_scheme_identifier`, `create_a_subscription`, `update_a_subscription`,
`pause_a_subscription`, `resume_a_subscription`, `cancel_a_subscription`,
`create_a_verification_detail`, `retry_a_webhook`, `perform_a_bank_details_lookup`,
`create_a_mandate_pdf`, `create_a_customer_bank_account_token`.

Service API documentation: https://developer.gocardless.com/api-reference/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); GoCardless access token, sent as Authorization: Bearer
  <access_token>. Never logged.
- `balances_creditor_id` (optional, string); Balances Creditor Id query value for the matching
  GoCardless stream.
- `bank_account_detail_id` (optional, string); Bank Account Detail Id used by GoCardless stream
  paths.
- `bank_account_holder_verification_id` (optional, string); Bank Account Holder Verification Id used
  by GoCardless stream paths.
- `bank_authorisation_id` (optional, string); Bank Authorisation Id used by GoCardless stream paths.
- `base_url` (required, string); format `uri`; GoCardless API base URL. Use
  https://api.gocardless.com for live or https://api-sandbox.gocardless.com for sandbox.
- `billing_request_id` (optional, string); Billing Request Id used by GoCardless stream paths.
- `billing_request_institutions_country_code` (optional, string); Billing Request Institutions
  Country Code query value for the matching GoCardless stream.
- `billing_request_template_id` (optional, string); Billing Request Template Id used by GoCardless
  stream paths.
- `block_id` (optional, string); Block Id used by GoCardless stream paths.
- `creditor_bank_account_id` (optional, string); Creditor Bank Account Id used by GoCardless stream
  paths.
- `creditor_id` (optional, string); Creditor Id used by GoCardless stream paths.
- `customer_bank_account_id` (optional, string); Customer Bank Account Id used by GoCardless stream
  paths.
- `customer_bank_account_token_id` (optional, string); Customer Bank Account Token Id used by
  GoCardless stream paths.
- `customer_id` (optional, string); Customer Id used by GoCardless stream paths.
- `event_id` (optional, string); Event Id used by GoCardless stream paths.
- `export_id` (optional, string); Export Id used by GoCardless stream paths.
- `funds_availability_amount` (optional, string); Funds Availability Amount query value for the
  matching GoCardless stream.
- `gc_key_id` (optional, string); Optional Gc-Key-Id public key identifier sent for encrypted
  bank-account-detail reads; omitted when unset.
- `gocardless_version` (optional, string); default `2015-07-06`; GoCardless-Version header value
  sent on every request.
- `instalment_schedule_id` (optional, string); Instalment Schedule Id used by GoCardless stream
  paths.
- `mandate_id` (optional, string); Mandate Id used by GoCardless stream paths.
- `mandate_import_id` (optional, string); Mandate Import Id used by GoCardless stream paths.
- `outbound_payment_id` (optional, string); Outbound Payment Id used by GoCardless stream paths.
- `outbound_payment_import_id` (optional, string); Outbound Payment Import Id used by GoCardless
  stream paths.
- `payer_authorisation_id` (optional, string); Payer Authorisation Id used by GoCardless stream
  paths.
- `payment_account_id` (optional, string); Payment Account Id used by GoCardless stream paths.
- `payment_account_transaction_id` (optional, string); Payment Account Transaction Id used by
  GoCardless stream paths.
- `payment_account_transactions_value_date_from` (optional, string); Payment Account Transactions
  Value Date From query value for the matching GoCardless stream.
- `payment_account_transactions_value_date_to` (optional, string); Payment Account Transactions
  Value Date To query value for the matching GoCardless stream.
- `payment_id` (optional, string); Payment Id used by GoCardless stream paths.
- `payout_id` (optional, string); Payout Id used by GoCardless stream paths.
- `redirect_flow_id` (optional, string); Redirect Flow Id used by GoCardless stream paths.
- `refund_id` (optional, string); Refund Id used by GoCardless stream paths.
- `scheme_identifier_id` (optional, string); Scheme Identifier Id used by GoCardless stream paths.
- `start_date` (optional, string); format `date-time`; Sent as created_at[gt].
- `subscription_id` (optional, string); Subscription Id used by GoCardless stream paths.
- `tax_rate_id` (optional, string); Tax Rate Id used by GoCardless stream paths.
- `transferred_mandate_id` (optional, string); Transferred Mandate Id used by GoCardless stream
  paths.
- `verification_details_creditor_id` (optional, string); Verification Details Creditor Id query
  value for the matching GoCardless stream.
- `webhook_id` (optional, string); Webhook Id used by GoCardless stream paths.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `gocardless_version=2015-07-06`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/payments`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `after`; next token from
`meta.cursors.after`.

Pagination by stream: cursor: `payments`, `mandates`, `payouts`, `refunds`, `billing_requests`,
`billing_request_templates`, `balances`, `blocks`, `creditors`, `creditor_bank_accounts`,
`currency_exchange_rates`, `customers`, `customer_bank_accounts`, `events`, `exports`,
`instalment_schedules`, `mandate_import_entries`, `negative_balance_limits`, `outbound_payments`,
`outbound_payment_imports`, `outbound_payment_import_entries`, `payment_accounts`,
`payment_account_transactions_by_payment_account`, `payout_items`, `scheme_identifiers`,
`subscriptions`, `tax_rates`, `verification_details`, `webhooks`; none: `bank_authorisation`,
`billing_request`, `billing_request_template`, `institutions`, `billing_request_institutions`,
`bank_account_detail`, `bank_account_holder_verification`, `block`, `creditor`,
`creditor_bank_account`, `customer`, `customer_bank_account`, `event`, `export`,
`funds_availability`, `instalment_schedule`, `mandate`, `mandate_import`, `outbound_payment`,
`outbound_payment_stats`, `outbound_payment_import`, `payer_authorisation`, `payment`,
`payment_account`, `payment_account_transaction`, `payout`, `redirect_flow`, `refund`,
`scheme_identifier`, `subscription`, `tax_rate`, `transferred_mandate`, `webhook`,
`customer_bank_account_token`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `payments`: GET `/payments` - records path `payments`; query `limit`=`50`; cursor pagination;
  cursor parameter `after`; next token from `meta.cursors.after`; incremental cursor `created_at`;
  sent as `created_at[gt]`; formatted as `rfc3339`; initial lower bound from `start_date`; computed
  output fields `mandate`, `payout`.
- `mandates`: GET `/mandates` - records path `mandates`; query `limit`=`50`; cursor pagination;
  cursor parameter `after`; next token from `meta.cursors.after`; incremental cursor `created_at`;
  sent as `created_at[gt]`; formatted as `rfc3339`; initial lower bound from `start_date`; computed
  output fields `creditor`, `customer_bank_account`.
- `payouts`: GET `/payouts` - records path `payouts`; query `limit`=`50`; cursor pagination; cursor
  parameter `after`; next token from `meta.cursors.after`; incremental cursor `created_at`; sent as
  `created_at[gt]`; formatted as `rfc3339`; initial lower bound from `start_date`; computed output
  fields `creditor`, `creditor_bank_account`.
- `refunds`: GET `/refunds` - records path `refunds`; query `limit`=`50`; cursor pagination; cursor
  parameter `after`; next token from `meta.cursors.after`; incremental cursor `created_at`; sent as
  `created_at[gt]`; formatted as `rfc3339`; initial lower bound from `start_date`; computed output
  fields `mandate`, `payment`.
- `bank_authorisation`: GET `/bank_authorisations/{{ config.bank_authorisation_id }}` - records path
  `bank_authorisations`; emits passthrough records.
- `billing_requests`: GET `/billing_requests` - records path `billing_requests`; query `limit`=`50`;
  cursor pagination; cursor parameter `after`; next token from `meta.cursors.after`; emits
  passthrough records.
- `billing_request`: GET `/billing_requests/{{ config.billing_request_id }}` - records path
  `billing_requests`; emits passthrough records.
- `billing_request_templates`: GET `/billing_request_templates` - records path
  `billing_request_templates`; query `limit`=`50`; cursor pagination; cursor parameter `after`; next
  token from `meta.cursors.after`; emits passthrough records.
- `billing_request_template`: GET `/billing_request_templates/{{ config.billing_request_template_id
  }}` - records path `billing_request_templates`; emits passthrough records.
- `institutions`: GET `/institutions` - records path `institutions`; emits passthrough records.
- `billing_request_institutions`: GET `/billing_requests/{{ config.billing_request_id
  }}/institutions` - records path `institutions`; query `country_code`=`{{
  config.billing_request_institutions_country_code }}`; emits passthrough records.
- `balances`: GET `/balances` - records path `balances`; query `creditor`=`{{
  config.balances_creditor_id }}`; `limit`=`50`; cursor pagination; cursor parameter `after`; next
  token from `meta.cursors.after`; emits passthrough records.
- `bank_account_detail`: GET `/bank_account_details/{{ config.bank_account_detail_id }}` - records
  path `bank_account_details`; emits passthrough records.
- `bank_account_holder_verification`: GET `/bank_account_holder_verifications/{{
  config.bank_account_holder_verification_id }}` - records path `bank_account_holder_verifications`;
  emits passthrough records.
- `block`: GET `/blocks/{{ config.block_id }}` - records path `blocks`; emits passthrough records.
- `blocks`: GET `/blocks` - records path `blocks`; query `limit`=`50`; cursor pagination; cursor
  parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `creditors`: GET `/creditors` - records path `creditors`; query `limit`=`50`; cursor pagination;
  cursor parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `creditor`: GET `/creditors/{{ config.creditor_id }}` - records path `creditors`; emits
  passthrough records.
- `creditor_bank_accounts`: GET `/creditor_bank_accounts` - records path `creditor_bank_accounts`;
  query `limit`=`50`; cursor pagination; cursor parameter `after`; next token from
  `meta.cursors.after`; emits passthrough records.
- `creditor_bank_account`: GET `/creditor_bank_accounts/{{ config.creditor_bank_account_id }}` -
  records path `creditor_bank_accounts`; emits passthrough records.
- `currency_exchange_rates`: GET `/currency_exchange_rates` - records path
  `currency_exchange_rates`; query `limit`=`50`; cursor pagination; cursor parameter `after`; next
  token from `meta.cursors.after`; emits passthrough records.
- `customers`: GET `/customers` - records path `customers`; query `limit`=`50`; cursor pagination;
  cursor parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `customer`: GET `/customers/{{ config.customer_id }}` - records path `customers`; emits
  passthrough records.
- `customer_bank_accounts`: GET `/customer_bank_accounts` - records path `customer_bank_accounts`;
  query `limit`=`50`; cursor pagination; cursor parameter `after`; next token from
  `meta.cursors.after`; emits passthrough records.
- `customer_bank_account`: GET `/customer_bank_accounts/{{ config.customer_bank_account_id }}` -
  records path `customer_bank_accounts`; emits passthrough records.
- `events`: GET `/events` - records path `events`; query `limit`=`50`; cursor pagination; cursor
  parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `event`: GET `/events/{{ config.event_id }}` - records path `events`; emits passthrough records.
- `export`: GET `/exports/{{ config.export_id }}` - records path `exports`; emits passthrough
  records.
- `exports`: GET `/exports` - records path `exports`; query `limit`=`50`; cursor pagination; cursor
  parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `funds_availability`: GET `/funds_availability/{{ config.mandate_id }}` - records path `.`; query
  `amount`=`{{ config.funds_availability_amount }}`; emits passthrough records.
- `instalment_schedules`: GET `/instalment_schedules` - records path `instalment_schedules`; query
  `limit`=`50`; cursor pagination; cursor parameter `after`; next token from `meta.cursors.after`;
  emits passthrough records.
- `instalment_schedule`: GET `/instalment_schedules/{{ config.instalment_schedule_id }}` - records
  path `instalment_schedules`; emits passthrough records.
- `mandate`: GET `/mandates/{{ config.mandate_id }}` - records path `mandates`; emits passthrough
  records.
- `mandate_import`: GET `/mandate_imports/{{ config.mandate_import_id }}` - records path
  `mandate_imports`; emits passthrough records.
- `mandate_import_entries`: GET `/mandate_import_entries` - records path `mandate_import_entries`;
  query `limit`=`50`; `mandate_import`=`{{ config.mandate_import_id }}`; cursor pagination; cursor
  parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `negative_balance_limits`: GET `/negative_balance_limits` - records path
  `negative_balance_limits`; query `limit`=`50`; cursor pagination; cursor parameter `after`; next
  token from `meta.cursors.after`; emits passthrough records.
- `outbound_payment`: GET `/outbound_payments/{{ config.outbound_payment_id }}` - records path
  `outbound_payments`; emits passthrough records.
- `outbound_payments`: GET `/outbound_payments` - records path `outbound_payments`; query
  `limit`=`50`; cursor pagination; cursor parameter `after`; next token from `meta.cursors.after`;
  emits passthrough records.
- `outbound_payment_stats`: GET `/outbound_payments/stats` - records path `outbound_payments`; emits
  passthrough records.
- `outbound_payment_import`: GET `/outbound_payment_imports/{{ config.outbound_payment_import_id }}`
  - records path `outbound_payment_imports`; emits passthrough records.
- `outbound_payment_imports`: GET `/outbound_payment_imports` - records path
  `outbound_payment_imports`; query `limit`=`50`; cursor pagination; cursor parameter `after`; next
  token from `meta.cursors.after`; emits passthrough records.
- `outbound_payment_import_entries`: GET `/outbound_payment_import_entries` - records path
  `outbound_payment_import_entries`; query `limit`=`50`; `outbound_payment_import`=`{{
  config.outbound_payment_import_id }}`; cursor pagination; cursor parameter `after`; next token
  from `meta.cursors.after`; emits passthrough records.
- `payer_authorisation`: GET `/payer_authorisations/{{ config.payer_authorisation_id }}` - records
  path `payer_authorisations`; emits passthrough records.
- `payment`: GET `/payments/{{ config.payment_id }}` - records path `payments`; emits passthrough
  records.
- `payment_account`: GET `/payment_accounts/{{ config.payment_account_id }}` - records path
  `payment_accounts`; emits passthrough records.
- `payment_accounts`: GET `/payment_accounts` - records path `payment_accounts`; query `limit`=`50`;
  cursor pagination; cursor parameter `after`; next token from `meta.cursors.after`; emits
  passthrough records.
- `payment_account_transaction`: GET `/payment_account_transactions/{{
  config.payment_account_transaction_id }}` - records path `payment_account_transactions`; emits
  passthrough records.
- `payment_account_transactions_by_payment_account`: GET `/payment_accounts/{{
  config.payment_account_id }}/transactions` - records path `payment_account_transactions`; query
  `limit`=`50`; `value_date_from`=`{{ config.payment_account_transactions_value_date_from }}`;
  `value_date_to`=`{{ config.payment_account_transactions_value_date_to }}`; cursor pagination;
  cursor parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `payout`: GET `/payouts/{{ config.payout_id }}` - records path `payouts`; emits passthrough
  records.
- `payout_items`: GET `/payout_items` - records path `payout_items`; query `limit`=`50`;
  `payout`=`{{ config.payout_id }}`; cursor pagination; cursor parameter `after`; next token from
  `meta.cursors.after`; emits passthrough records.
- `redirect_flow`: GET `/redirect_flows/{{ config.redirect_flow_id }}` - records path
  `redirect_flows`; emits passthrough records.
- `refund`: GET `/refunds/{{ config.refund_id }}` - records path `refunds`; emits passthrough
  records.
- `scheme_identifiers`: GET `/scheme_identifiers` - records path `scheme_identifiers`; query
  `limit`=`50`; cursor pagination; cursor parameter `after`; next token from `meta.cursors.after`;
  emits passthrough records.
- `scheme_identifier`: GET `/scheme_identifiers/{{ config.scheme_identifier_id }}` - records path
  `scheme_identifiers`; emits passthrough records.
- `subscriptions`: GET `/subscriptions` - records path `subscriptions`; query `limit`=`50`; cursor
  pagination; cursor parameter `after`; next token from `meta.cursors.after`; emits passthrough
  records.
- `subscription`: GET `/subscriptions/{{ config.subscription_id }}` - records path `subscriptions`;
  emits passthrough records.
- `tax_rates`: GET `/tax_rates` - records path `tax_rates`; query `limit`=`50`; cursor pagination;
  cursor parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `tax_rate`: GET `/tax_rates/{{ config.tax_rate_id }}` - records path `tax_rates`; emits
  passthrough records.
- `transferred_mandate`: GET `/transferred_mandates/{{ config.transferred_mandate_id }}` - records
  path `transferred_mandates`; emits passthrough records.
- `verification_details`: GET `/verification_details` - records path `verification_details`; query
  `creditor`=`{{ config.verification_details_creditor_id }}`; `limit`=`50`; cursor pagination;
  cursor parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `webhooks`: GET `/webhooks` - records path `webhooks`; query `limit`=`50`; cursor pagination;
  cursor parameter `after`; next token from `meta.cursors.after`; emits passthrough records.
- `webhook`: GET `/webhooks/{{ config.webhook_id }}` - records path `webhooks`; emits passthrough
  records.
- `customer_bank_account_token`: GET `/customer_bank_account_tokens/{{
  config.customer_bank_account_token_id }}` - records path `customer_bank_account_tokens`; emits
  passthrough records.

## Write actions & risks

Overall write risk: external GoCardless API mutations including creates, updates, cancellations,
retries, simulations, and customer/account changes.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_a_bank_authorisation`: POST `/bank_authorisations` - kind `create`; body type `json`;
  risk: GoCardless mutation: Create a Bank Authorisation.
- `create_a_billing_request`: POST `/billing_requests` - kind `create`; body type `json`; risk:
  GoCardless mutation: Create a Billing Request.
- `collect_customer_details`: POST `/billing_requests/{{ record.billing_request_id
  }}/actions/collect_customer_details` - kind `create`; body type `json`; path fields
  `billing_request_id`; required record fields `billing_request_id`; accepted fields
  `billing_request_id`; risk: GoCardless mutation: Collect customer details.
- `collect_bank_account_details`: POST `/billing_requests/{{ record.billing_request_id
  }}/actions/collect_bank_account` - kind `create`; body type `json`; path fields
  `billing_request_id`; required record fields `billing_request_id`; accepted fields
  `billing_request_id`; risk: GoCardless mutation: Collect bank account details.
- `confirm_the_payer_details`: POST `/billing_requests/{{ record.billing_request_id
  }}/actions/confirm_payer_details` - kind `create`; body type `json`; path fields
  `billing_request_id`; required record fields `billing_request_id`; accepted fields
  `billing_request_id`; risk: GoCardless mutation: Confirm the payer details.
- `fulfil_a_billing_request`: POST `/billing_requests/{{ record.billing_request_id
  }}/actions/fulfil` - kind `create`; body type `json`; path fields `billing_request_id`; required
  record fields `billing_request_id`; accepted fields `billing_request_id`; risk: GoCardless
  mutation: Fulfil a Billing Request.
- `cancel_a_billing_request`: POST `/billing_requests/{{ record.billing_request_id
  }}/actions/cancel` - kind `delete`; body type `none`; path fields `billing_request_id`; required
  record fields `billing_request_id`; accepted fields `billing_request_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: Destructive GoCardless mutation:
  Cancel a Billing Request.
- `notify_the_customer`: POST `/billing_requests/{{ record.billing_request_id }}/actions/notify` -
  kind `create`; body type `json`; path fields `billing_request_id`; required record fields
  `billing_request_id`; accepted fields `billing_request_id`; risk: GoCardless mutation: Notify the
  customer.
- `trigger_fallback`: POST `/billing_requests/{{ record.billing_request_id }}/actions/fallback` -
  kind `create`; body type `json`; path fields `billing_request_id`; required record fields
  `billing_request_id`; accepted fields `billing_request_id`; risk: GoCardless mutation: Trigger
  fallback.
- `change_currency`: POST `/billing_requests/{{ record.billing_request_id
  }}/actions/choose_currency` - kind `create`; body type `json`; path fields `billing_request_id`;
  required record fields `billing_request_id`; accepted fields `billing_request_id`; risk:
  GoCardless mutation: Change currency.
- `select_institution_for_a_billing_request`: POST `/billing_requests/{{ record.billing_request_id
  }}/actions/select_institution` - kind `create`; body type `json`; path fields
  `billing_request_id`; required record fields `billing_request_id`; accepted fields
  `billing_request_id`; risk: GoCardless mutation: Select institution for a Billing Request.
- `create_a_billing_request_flow`: POST `/billing_request_flows` - kind `create`; body type `json`;
  risk: GoCardless mutation: Create a Billing Request Flow.
- `initialise_a_billing_request_flow`: POST `/billing_request_flows/{{
  record.billing_request_flow_id }}/actions/initialise` - kind `create`; body type `json`; path
  fields `billing_request_flow_id`; required record fields `billing_request_flow_id`; accepted
  fields `billing_request_flow_id`; risk: GoCardless mutation: Initialise a Billing Request Flow.
- `create_a_billing_request_template`: POST `/billing_request_templates` - kind `create`; body type
  `json`; risk: GoCardless mutation: Create a Billing Request Template.
- `update_a_billing_request_template`: PUT `/billing_request_templates/{{
  record.billing_request_template_id }}` - kind `update`; body type `json`; path fields
  `billing_request_template_id`; required record fields `billing_request_template_id`; accepted
  fields `billing_request_template_id`; risk: GoCardless mutation: Update a Billing Request
  Template.
- `create_a_billing_request_with_actions`: POST `/billing_requests/create_with_actions` - kind
  `create`; body type `json`; risk: GoCardless mutation: Create a Billing Request with Actions.
- `create_a_bank_account_holder_verification`: POST `/bank_account_holder_verifications` - kind
  `create`; body type `json`; risk: GoCardless mutation: Create a bank account holder verification.
- `create_a_block`: POST `/blocks` - kind `create`; body type `json`; risk: GoCardless mutation:
  Create a block.
- `disable_a_block`: POST `/blocks/{{ record.block_id }}/actions/disable` - kind `delete`; body type
  `none`; path fields `block_id`; required record fields `block_id`; accepted fields `block_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Destructive
  GoCardless mutation: Disable a block.
- `enable_a_block`: POST `/blocks/{{ record.block_id }}/actions/enable` - kind `create`; body type
  `json`; path fields `block_id`; required record fields `block_id`; accepted fields `block_id`;
  risk: GoCardless mutation: Enable a block.
- `create_blocks_by_reference`: POST `/blocks/block_by_ref` - kind `create`; body type `json`; risk:
  GoCardless mutation: Create blocks by reference.
- `create_a_creditor`: POST `/creditors` - kind `create`; body type `json`; risk: GoCardless
  mutation: Create a creditor.
- `update_a_creditor`: PUT `/creditors/{{ record.creditor_id }}` - kind `update`; body type `json`;
  path fields `creditor_id`; required record fields `creditor_id`; accepted fields `creditor_id`;
  risk: GoCardless mutation: Update a creditor.
- `create_a_creditor_bank_account`: POST `/creditor_bank_accounts` - kind `create`; body type
  `json`; risk: GoCardless mutation: Create a creditor bank account.
- `disable_a_creditor_bank_account`: POST `/creditor_bank_accounts/{{
  record.creditor_bank_account_id }}/actions/disable` - kind `delete`; body type `none`; path fields
  `creditor_bank_account_id`; required record fields `creditor_bank_account_id`; accepted fields
  `creditor_bank_account_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Destructive GoCardless mutation: Disable a creditor bank account.
- `create_a_customer`: POST `/customers` - kind `create`; body type `json`; risk: GoCardless
  mutation: Create a customer.
- `update_a_customer`: PUT `/customers/{{ record.customer_id }}` - kind `update`; body type `json`;
  path fields `customer_id`; required record fields `customer_id`; accepted fields `customer_id`;
  risk: GoCardless mutation: Update a customer.
- `remove_a_customer`: DELETE `/customers/{{ record.customer_id }}` - kind `delete`; body type
  `none`; path fields `customer_id`; required record fields `customer_id`; accepted fields
  `customer_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Destructive GoCardless mutation: Remove a customer.
- `create_a_customer_bank_account`: POST `/customer_bank_accounts` - kind `create`; body type
  `json`; risk: GoCardless mutation: Create a customer bank account.
- `update_a_customer_bank_account`: PUT `/customer_bank_accounts/{{ record.customer_bank_account_id
  }}` - kind `update`; body type `json`; path fields `customer_bank_account_id`; required record
  fields `customer_bank_account_id`; accepted fields `customer_bank_account_id`; risk: GoCardless
  mutation: Update a customer bank account.
- `disable_a_customer_bank_account`: POST `/customer_bank_accounts/{{
  record.customer_bank_account_id }}/actions/disable` - kind `delete`; body type `none`; path fields
  `customer_bank_account_id`; required record fields `customer_bank_account_id`; accepted fields
  `customer_bank_account_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Destructive GoCardless mutation: Disable a customer bank account.
- `handle_a_notification`: POST `/customer_notifications/{{ record.customer_notification_id
  }}/actions/handle` - kind `create`; body type `json`; path fields `customer_notification_id`;
  required record fields `customer_notification_id`; accepted fields `customer_notification_id`;
  risk: GoCardless mutation: Handle a notification.
- `create_with_dates`: POST `/instalment_schedules` - kind `create`; body type `json`; risk:
  GoCardless mutation: Create (with dates).
- `update_an_instalment_schedule`: PUT `/instalment_schedules/{{ record.instalment_schedule_id }}` -
  kind `update`; body type `json`; path fields `instalment_schedule_id`; required record fields
  `instalment_schedule_id`; accepted fields `instalment_schedule_id`; risk: GoCardless mutation:
  Update an instalment schedule.
- `cancel_an_instalment_schedule`: POST `/instalment_schedules/{{ record.instalment_schedule_id
  }}/actions/cancel` - kind `delete`; body type `none`; path fields `instalment_schedule_id`;
  required record fields `instalment_schedule_id`; accepted fields `instalment_schedule_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Destructive
  GoCardless mutation: Cancel an instalment schedule.
- `create_a_logo_associated_with_a_creditor`: POST `/branding/logos` - kind `create`; body type
  `json`; risk: GoCardless mutation: Create a logo associated with a creditor.
- `create_a_mandate`: POST `/mandates` - kind `create`; body type `json`; risk: GoCardless mutation:
  Create a mandate.
- `update_a_mandate`: PUT `/mandates/{{ record.mandate_id }}` - kind `update`; body type `json`;
  path fields `mandate_id`; required record fields `mandate_id`; accepted fields `mandate_id`; risk:
  GoCardless mutation: Update a mandate.
- `cancel_a_mandate`: POST `/mandates/{{ record.mandate_id }}/actions/cancel` - kind `delete`; body
  type `none`; path fields `mandate_id`; required record fields `mandate_id`; accepted fields
  `mandate_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Destructive GoCardless mutation: Cancel a mandate.
- `reinstate_a_mandate`: POST `/mandates/{{ record.mandate_id }}/actions/reinstate` - kind `create`;
  body type `json`; path fields `mandate_id`; required record fields `mandate_id`; accepted fields
  `mandate_id`; risk: GoCardless mutation: Reinstate a mandate.
- `create_a_new_mandate_import`: POST `/mandate_imports` - kind `create`; body type `json`; risk:
  GoCardless mutation: Create a new mandate import.
- `submit_a_mandate_import`: POST `/mandate_imports/{{ record.mandate_import_id }}/actions/submit` -
  kind `create`; body type `json`; path fields `mandate_import_id`; required record fields
  `mandate_import_id`; accepted fields `mandate_import_id`; risk: GoCardless mutation: Submit a
  mandate import.
- `cancel_a_mandate_import`: POST `/mandate_imports/{{ record.mandate_import_id }}/actions/cancel` -
  kind `delete`; body type `none`; path fields `mandate_import_id`; required record fields
  `mandate_import_id`; accepted fields `mandate_import_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Destructive GoCardless mutation: Cancel a mandate
  import.
- `add_a_mandate_import_entry`: POST `/mandate_import_entries` - kind `create`; body type `json`;
  risk: GoCardless mutation: Add a mandate import entry.
- `create_an_outbound_payment`: POST `/outbound_payments` - kind `create`; body type `json`; risk:
  GoCardless mutation: Create an outbound payment.
- `create_a_withdrawal_outbound_payment`: POST `/outbound_payments/withdrawal` - kind `create`; body
  type `json`; risk: GoCardless mutation: Create a withdrawal outbound payment.
- `cancel_an_outbound_payment`: POST `/outbound_payments/{{ record.outbound_payment_id
  }}/actions/cancel` - kind `delete`; body type `none`; path fields `outbound_payment_id`; required
  record fields `outbound_payment_id`; accepted fields `outbound_payment_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Destructive GoCardless
  mutation: Cancel an outbound payment.
- `approve_an_outbound_payment`: POST `/outbound_payments/{{ record.outbound_payment_id
  }}/actions/approve` - kind `create`; body type `json`; path fields `outbound_payment_id`; required
  record fields `outbound_payment_id`; accepted fields `outbound_payment_id`; risk: GoCardless
  mutation: Approve an outbound payment.
- `update_an_outbound_payment`: PUT `/outbound_payments/{{ record.outbound_payment_id }}` - kind
  `update`; body type `json`; path fields `outbound_payment_id`; required record fields
  `outbound_payment_id`; accepted fields `outbound_payment_id`; risk: GoCardless mutation: Update an
  outbound payment.
- `create_an_outbound_payment_import`: POST `/outbound_payment_imports` - kind `create`; body type
  `json`; risk: GoCardless mutation: Create an outbound payment import.
- `create_a_payer_authorisation`: POST `/payer_authorisations` - kind `create`; body type `json`;
  risk: GoCardless mutation: Create a Payer Authorisation.
- `update_a_payer_authorisation`: PUT `/payer_authorisations/{{ record.payer_authorisation_id }}` -
  kind `update`; body type `json`; path fields `payer_authorisation_id`; required record fields
  `payer_authorisation_id`; accepted fields `payer_authorisation_id`; risk: GoCardless mutation:
  Update a Payer Authorisation.
- `submit_a_payer_authorisation`: POST `/payer_authorisations/{{ record.payer_authorisation_id
  }}/actions/submit` - kind `create`; body type `json`; path fields `payer_authorisation_id`;
  required record fields `payer_authorisation_id`; accepted fields `payer_authorisation_id`; risk:
  GoCardless mutation: Submit a Payer Authorisation.
- `confirm_a_payer_authorisation`: POST `/payer_authorisations/{{ record.payer_authorisation_id
  }}/actions/confirm` - kind `create`; body type `json`; path fields `payer_authorisation_id`;
  required record fields `payer_authorisation_id`; accepted fields `payer_authorisation_id`; risk:
  GoCardless mutation: Confirm a Payer Authorisation.
- `create_a_payer_theme_associated_with_a_creditor`: POST `/branding/payer_themes` - kind `create`;
  body type `json`; risk: GoCardless mutation: Create a payer theme associated with a creditor.
- `create_a_payment`: POST `/payments` - kind `create`; body type `json`; risk: GoCardless mutation:
  Create a payment.
- `update_a_payment`: PUT `/payments/{{ record.payment_id }}` - kind `update`; body type `json`;
  path fields `payment_id`; required record fields `payment_id`; accepted fields `payment_id`; risk:
  GoCardless mutation: Update a payment.
- `cancel_a_payment`: POST `/payments/{{ record.payment_id }}/actions/cancel` - kind `delete`; body
  type `none`; path fields `payment_id`; required record fields `payment_id`; accepted fields
  `payment_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Destructive GoCardless mutation: Cancel a payment.
- `retry_a_payment`: POST `/payments/{{ record.payment_id }}/actions/retry` - kind `create`; body
  type `json`; path fields `payment_id`; required record fields `payment_id`; accepted fields
  `payment_id`; risk: GoCardless mutation: Retry a payment.
- `update_a_payout`: PUT `/payouts/{{ record.payout_id }}` - kind `update`; body type `json`; path
  fields `payout_id`; required record fields `payout_id`; accepted fields `payout_id`; risk:
  GoCardless mutation: Update a payout.
- `create_a_redirect_flow`: POST `/redirect_flows` - kind `create`; body type `json`; risk:
  GoCardless mutation: Create a redirect flow.
- `complete_a_redirect_flow`: POST `/redirect_flows/{{ record.redirect_flow_id }}/actions/complete`
  - kind `create`; body type `json`; path fields `redirect_flow_id`; required record fields
  `redirect_flow_id`; accepted fields `redirect_flow_id`; risk: GoCardless mutation: Complete a
  redirect flow.
- `create_a_refund`: POST `/refunds` - kind `create`; body type `json`; risk: GoCardless mutation:
  Create a refund.
- `update_a_refund`: PUT `/refunds/{{ record.refund_id }}` - kind `update`; body type `json`; path
  fields `refund_id`; required record fields `refund_id`; accepted fields `refund_id`; risk:
  GoCardless mutation: Update a refund.
- `simulate_a_scenario`: POST `/scenario_simulators/{{ record.scenario }}/actions/run` - kind
  `create`; body type `json`; path fields `scenario`; required record fields `scenario`; accepted
  fields `scenario`; risk: GoCardless mutation: Simulate a scenario.
- `create_a_scheme_identifier`: POST `/scheme_identifiers` - kind `create`; body type `json`; risk:
  GoCardless mutation: Create a scheme identifier.
- `create_a_subscription`: POST `/subscriptions` - kind `create`; body type `json`; risk: GoCardless
  mutation: Create a subscription.
- `update_a_subscription`: PUT `/subscriptions/{{ record.subscription_id }}` - kind `update`; body
  type `json`; path fields `subscription_id`; required record fields `subscription_id`; accepted
  fields `subscription_id`; risk: GoCardless mutation: Update a subscription.
- `pause_a_subscription`: POST `/subscriptions/{{ record.subscription_id }}/actions/pause` - kind
  `create`; body type `json`; path fields `subscription_id`; required record fields
  `subscription_id`; accepted fields `subscription_id`; risk: GoCardless mutation: Pause a
  subscription.
- `resume_a_subscription`: POST `/subscriptions/{{ record.subscription_id }}/actions/resume` - kind
  `create`; body type `json`; path fields `subscription_id`; required record fields
  `subscription_id`; accepted fields `subscription_id`; risk: GoCardless mutation: Resume a
  subscription.
- `cancel_a_subscription`: POST `/subscriptions/{{ record.subscription_id }}/actions/cancel` - kind
  `delete`; body type `none`; path fields `subscription_id`; required record fields
  `subscription_id`; accepted fields `subscription_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Destructive GoCardless mutation: Cancel a
  subscription.
- `create_a_verification_detail`: POST `/verification_details` - kind `create`; body type `json`;
  risk: GoCardless mutation: Create a verification detail.
- `retry_a_webhook`: POST `/webhooks/{{ record.webhook_id }}/actions/retry` - kind `create`; body
  type `json`; path fields `webhook_id`; required record fields `webhook_id`; accepted fields
  `webhook_id`; risk: GoCardless mutation: Retry a webhook.
- `perform_a_bank_details_lookup`: POST `/bank_details_lookups` - kind `create`; body type `json`;
  risk: GoCardless mutation: Perform a bank details lookup.
- `create_a_mandate_pdf`: POST `/mandate_pdfs` - kind `create`; body type `json`; risk: GoCardless
  mutation: Create a mandate PDF.
- `create_a_customer_bank_account_token`: POST `/customer_bank_account_tokens` - kind `create`; body
  type `json`; risk: GoCardless mutation: Create a customer bank account token.

## Known limits

- Batch defaults: read_page_size=50, write_batch_size=1.
- API coverage includes 63 stream-backed endpoint group(s), 76 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=3.
