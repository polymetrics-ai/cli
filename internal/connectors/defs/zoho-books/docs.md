# Overview

Reads and writes Zoho Books API v3 accounting resources using the connector engine.

Readable streams: `contacts`, `invoices`, `items`, `list_bank_accounts`, `get_bank_account`,
`get_last_imported_bank_statement`, `list_bank_account_rules`, `get_bank_account_rule`,
`list_bank_transactions`, `get_matching_bank_transactions`, `get_bank_transaction`,
`list_base_currency_adjustments`, `list_base_currency_adjustment_accounts`,
`list_base_currency_adjustment_contacts`, `get_base_currency_adjustment`, `list_bills`,
`convert_purchase_order_to_bill`, `get_bill`, `get_bill_comments`, `list_bill_payments`,
`list_chart_of_accounts`, `list_chart_of_account_transactions`, `get_chart_of_account`,
`list_contact_persons`, `get_contact_person`, `get_contact`, `get_contact_address`,
`list_contact_autobill_recurring_invoices`, `list_contact_comments`, `get_unused_retainer_payments`,
`list_contact_credit_note_refunds`, `get_contact_statement_mail`, `list_credit_notes`,
`list_credit_note_refunds_of_all_credit_notes`, `get_credit_note_refund_by_id`,
`list_credit_note_templates`, `get_credit_note`, `list_credit_note_comments`,
`get_credit_note_custom_fields`, `get_credit_note_email`, `get_credit_note_email_history`,
`list_invoices_of_credit_note`, `list_credit_note_refunds_of_a_credit_note`,
`get_credit_note_refund`, `list_currencies`, `get_currency`, `list_exchange_rates`,
`get_exchange_rate`, `list_custom_modules`, `get_custom_module`, `list_custom_module_records`,
`get_custom_module_record`, `get_customer_debit_note`, `list_customer_payments`,
`list_customer_payment_refunds`, `get_customer_payment_refund`, `get_customer_payment`,
`list_delivery_challans`, `list_delivery_challan_templates`, `get_delivery_challan`,
`list_estimates`, `list_estimate_templates`, `get_estimate`, `list_estimate_comments`,
`get_estimate_email`, `list_employees`, `get_employee`, `list_expenses`, `get_expense`,
`list_expense_comments`, `list_fixed_assets`, `get_fixed_asset`, `get_fixed_asset_forecast`,
`get_fixed_asset_history`, `get_fixed_asset_type_list`, `list_invoice_templates`, `get_invoice`,
`list_invoice_comments`, `list_invoice_credits_applied`, `get_invoice_document_details`,
`get_invoice_email`, `get_payment_reminder_mail_content_for_invoice`, `list_invoice_payments`,
`generate_invoice_payment_link`, `list_item_details`, `get_item`, `list_journals`, `get_journal`,
`list_journal_credits`, `list_recurring_journals`, `get_recurring_journal`, `list_child_journals`,
`get_transaction_journal_view`, `list_locations`, `list_opening_balance_transactions`,
`list_opening_balance_details`, `get_opening_balance`, `list_organizations`,
`list_organizations_for_user`, `get_organization`, `list_pricebooks`, `list_projects`,
`get_project`, `list_project_comments`, `list_project_invoices`, `list_project_tasks`,
`get_project_task`, `list_project_users`, `get_project_user`, `list_purchase_orders`,
`list_purchase_order_templates`, `get_purchase_order`, `list_purchase_order_comments`,
`get_recurring_bill`, `list_recurring_bills`, `list_recurring_expenses`, `get_recurring_expense`,
`list_recurring_expense_history`, `list_child_expenses_of_recurring_expense`,
`list_recurring_invoices`, `get_recurring_invoice`, `list_recurring_invoice_history`,
`list_recurring_invoice_child_invoices`, `get_register_budget_vs_actuals`,
`list_register_bulk_action_history`, `get_register_bulk_action_history`,
`get_register_bulk_update_editpage`, `list_register_transactions`, `get_tags`, `all_tag_options`,
`get_all_tag_options`, `list_retainer_invoices`, `list_retainer_invoice_templates`,
`get_retainer_invoice`, `list_retainer_invoice`, `get_retainer_invoice_email`, `list_sales_orders`,
`list_sales_order_templates`, `get_sales_order`, `get_sales_order_email`, `list_sales_receipts`,
`get_sales_receipt`, `list_tasks`, `get_task`, `list_task_comments`, `get_task_document`,
`list_tax_authorities`, `get_tax_authority`, `list_taxes`, `get_tax`, `list_tax_exemptions`,
`get_tax_exemption`, `get_tax_group`, `list_time_entries`, `get_running_timer`, `get_time_entry`,
`get_accounting_period_transaction_lock`, `get_transaction_lock`, `list_transaction_locks`,
`list_users`, `get_current_user`, `get_user`, `list_vendor_credits`,
`list_vendor_credit_refunds_of_all_vendor_credits`, `get_vendor_credit`, `list_bills_credited`,
`list_vendor_credit_comments`, `list_vendor_credit_refunds_of_a_vendor_credit`,
`get_vendor_credit_refund`, `list_vendor_payments`, `get_vendor_payment`,
`get_vendor_payment_email_content`, `list_vendor_payment_refunds`, `get_vendor_payment_refund`.

Write actions: `create_bank_account`, `update_bank_account`, `delete_bank_account`,
`mark_bank_account_active`, `mark_bank_account_inactive`, `update_bank_account_preferences`,
`create_bank_reconciliation`, `update_bank_reconciliation`, `delete_bank_reconciliation`,
`add_bank_reconciliation_attachment`, `delete_bank_reconciliation_document`,
`save_bank_reconciliation_draft`, `delete_last_imported_bank_statement`, `import_bank_statements`,
`create_bank_account_match_filter`, `update_bank_account_match_filter`,
`delete_bank_account_match_filter`, `create_bank_account_rule`, `bulk_update_bank_account_rules`,
`bulk_delete_bank_account_rules`, `reorder_bank_account_rules`, `skip_suggested_bank_account_rule`,
`update_bank_account_rule`, `delete_bank_account_rule`, `create_bank_transaction`,
`categorize_bank_transaction_as_payment_refund`, `categorize_as_vendor_payment_refund`,
`categorize_bank_transaction`, `categorize_as_credit_note_refunds`,
`categorize_bank_transaction_as_customer_payment`, `categorize_bank_transaction_as_expense`,
`categorize_as_vendor_credit_refunds`, `categorize_bank_transaction_as_vendor_payment`,
`exclude_bank_transaction`, `match_bank_transaction`, `restore_bank_transaction`,
`update_bank_transaction`, `delete_bank_transaction`, `uncategorize_bank_transaction`,
`unmatch_bank_transaction`, `create_base_currency_adjustment`,
`bulk_delete_base_currency_adjustments`, `delete_base_currency_adjustment`,
`reevaluate_base_currency_adjustment`, `update_custom_fields_in_bill`, `create_bill`, `update_bill`,
`delete_bill`, `update_bill_billing_address`, `approve_bill`, `add_bill_attachment`,
`delete_bill_attachment`, `add_bill_comment`, `delete_bill_comment`, `apply_credits_to_bill`,
`delete_bill_payment`, `mark_bill_open`, `mark_bill_void`, `submit_bill`, `create_chart_of_account`,
`bulk_mark_chart_of_accounts_active`, `bulk_delete_chart_of_accounts`,
`bulk_mark_chart_of_accounts_inactive`, `delete_chart_of_account_transaction`,
`update_chart_of_account`, `delete_chart_of_account`, `mark_chart_of_account_active`,
`mark_chart_of_account_inactive`, `create_contact_person`, `update_contact_person`,
`delete_contact_person`, `mark_contact_person_primary`, `create_contact`, `delete_contacts`,
`create_contact_person_2`, `update_contact_person_2`, `delete_contact_person_2`,
`invite_contact_person_to_portal`, `resend_contact_person_portal_invite`,
`mark_contact_person_primary_2`, and 489 more.

Service API documentation: https://www.zoho.com/books/api/v3/.

## Auth setup

Connection fields:

- `accept` (optional, string); Zoho Books path or required-query parameter 'accept' used by one or
  more documented endpoints.
- `access_token` (required, secret, string); Zoho OAuth access token. Sent as the Authorization
  header with a 'Zoho-oauthtoken ' prefix; never logged.
- `account_id` (optional, string); Zoho Books path or required-query parameter 'account_id' used by
  one or more documented endpoints.
- `adjustment_date` (optional, string); Zoho Books path or required-query parameter
  'adjustment_date' used by one or more documented endpoints.
- `bank_account_id` (optional, string); Zoho Books path or required-query parameter
  'bank_account_id' used by one or more documented endpoints.
- `bank_transaction_id` (optional, string); Zoho Books path or required-query parameter
  'bank_transaction_id' used by one or more documented endpoints.
- `base_currency_adjustment_id` (optional, string); Zoho Books path or required-query parameter
  'base_currency_adjustment_id' used by one or more documented endpoints.
- `base_url` (optional, string); default `https://www.zohoapis.com/books/v3`; format `uri`; Zoho
  Books API base URL override for tests or region-specific data centers.
- `bill_id` (optional, string); Zoho Books path or required-query parameter 'bill_id' used by one or
  more documented endpoints.
- `card_id` (optional, string); Zoho Books path or required-query parameter 'card_id' used by one or
  more documented endpoints.
- `comment_id` (optional, string); Zoho Books path or required-query parameter 'comment_id' used by
  one or more documented endpoints.
- `contact_id` (optional, string); Zoho Books path or required-query parameter 'contact_id' used by
  one or more documented endpoints.
- `contact_person_id` (optional, string); Zoho Books path or required-query parameter
  'contact_person_id' used by one or more documented endpoints.
- `contactperson_id` (optional, string); Zoho Books path or required-query parameter
  'contactperson_id' used by one or more documented endpoints.
- `creditnote_id` (optional, string); Zoho Books path or required-query parameter 'creditnote_id'
  used by one or more documented endpoints.
- `creditnote_ids` (optional, string); Zoho Books path or required-query parameter 'creditnote_ids'
  used by one or more documented endpoints.
- `creditnote_refund_id` (optional, string); Zoho Books path or required-query parameter
  'creditnote_refund_id' used by one or more documented endpoints.
- `currency_id` (optional, string); Zoho Books path or required-query parameter 'currency_id' used
  by one or more documented endpoints.
- `customer_payment_id` (optional, string); Zoho Books path or required-query parameter
  'customer_payment_id' used by one or more documented endpoints.
- `debit_note_id` (optional, string); Zoho Books path or required-query parameter 'debit_note_id'
  used by one or more documented endpoints.
- `deliverychallan_id` (optional, string); Zoho Books path or required-query parameter
  'deliverychallan_id' used by one or more documented endpoints.
- `document_id` (optional, string); Zoho Books path or required-query parameter 'document_id' used
  by one or more documented endpoints.
- `email_template_id` (optional, string); Zoho Books path or required-query parameter
  'email_template_id' used by one or more documented endpoints.
- `employee_id` (optional, string); Zoho Books path or required-query parameter 'employee_id' used
  by one or more documented endpoints.
- `entity_type` (optional, string); Zoho Books path or required-query parameter 'entity_type' used
  by one or more documented endpoints.
- `estimate_id` (optional, string); Zoho Books path or required-query parameter 'estimate_id' used
  by one or more documented endpoints.
- `estimate_ids` (optional, string); Zoho Books path or required-query parameter 'estimate_ids' used
  by one or more documented endpoints.
- `exchange_rate` (optional, string); Zoho Books path or required-query parameter 'exchange_rate'
  used by one or more documented endpoints.
- `exchange_rate_id` (optional, string); Zoho Books path or required-query parameter
  'exchange_rate_id' used by one or more documented endpoints.
- `expense_id` (optional, string); Zoho Books path or required-query parameter 'expense_id' used by
  one or more documented endpoints.
- `expiry_time` (optional, string); Zoho Books path or required-query parameter 'expiry_time' used
  by one or more documented endpoints.
- `fixed_asset_id` (optional, string); Zoho Books path or required-query parameter 'fixed_asset_id'
  used by one or more documented endpoints.
- `invoice_id` (optional, string); Zoho Books path or required-query parameter 'invoice_id' used by
  one or more documented endpoints.
- `invoice_ids` (optional, string); Zoho Books path or required-query parameter 'invoice_ids' used
  by one or more documented endpoints.
- `item_id` (optional, string); Zoho Books path or required-query parameter 'item_id' used by one or
  more documented endpoints.
- `item_ids` (optional, string); Zoho Books path or required-query parameter 'item_ids' used by one
  or more documented endpoints.
- `journal_id` (optional, string); Zoho Books path or required-query parameter 'journal_id' used by
  one or more documented endpoints.
- `line1` (optional, string); Zoho Books path or required-query parameter 'line1' used by one or
  more documented endpoints.
- `link_type` (optional, string); Zoho Books path or required-query parameter 'link_type' used by
  one or more documented endpoints.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `metadata_name` (optional, string); Zoho Books path or required-query parameter 'metadata_name'
  used by one or more documented endpoints.
- `mode` (optional, string).
- `module_api_name` (optional, string); Zoho Books path or required-query parameter
  'module_api_name' used by one or more documented endpoints.
- `module_id` (optional, string); Zoho Books path or required-query parameter 'module_id' used by
  one or more documented endpoints.
- `module_name` (optional, string); Zoho Books path or required-query parameter 'module_name' used
  by one or more documented endpoints.
- `notes` (optional, string); Zoho Books path or required-query parameter 'notes' used by one or
  more documented endpoints.
- `organization_id` (optional, string); Zoho Books organization ID; sent as organization_id on read
  requests when set and required for write actions that target organization-scoped resources.
- `page_size` (optional, string); default `200`.
- `payment_id` (optional, string); Zoho Books path or required-query parameter 'payment_id' used by
  one or more documented endpoints.
- `project_id` (optional, string); Zoho Books path or required-query parameter 'project_id' used by
  one or more documented endpoints.
- `purchaseorder_id` (optional, string); Zoho Books path or required-query parameter
  'purchaseorder_id' used by one or more documented endpoints.
- `purchaseorder_ids` (optional, string); Zoho Books path or required-query parameter
  'purchaseorder_ids' used by one or more documented endpoints.
- `reconciliation_id` (optional, string); Zoho Books path or required-query parameter
  'reconciliation_id' used by one or more documented endpoints.
- `recurring_bill_id` (optional, string); Zoho Books path or required-query parameter
  'recurring_bill_id' used by one or more documented endpoints.
- `recurring_expense_id` (optional, string); Zoho Books path or required-query parameter
  'recurring_expense_id' used by one or more documented endpoints.
- `recurring_invoice_id` (optional, string); Zoho Books path or required-query parameter
  'recurring_invoice_id' used by one or more documented endpoints.
- `recurring_journal_id` (optional, string); Zoho Books path or required-query parameter
  'recurring_journal_id' used by one or more documented endpoints.
- `reference_id` (optional, string); Zoho Books path or required-query parameter 'reference_id' used
  by one or more documented endpoints.
- `refund_id` (optional, string); Zoho Books path or required-query parameter 'refund_id' used by
  one or more documented endpoints.
- `retainerinvoice_id` (optional, string); Zoho Books path or required-query parameter
  'retainerinvoice_id' used by one or more documented endpoints.
- `rule_id` (optional, string); Zoho Books path or required-query parameter 'rule_id' used by one or
  more documented endpoints.
- `sales_receipt_id` (optional, string); Zoho Books path or required-query parameter
  'sales_receipt_id' used by one or more documented endpoints.
- `salesorder_id` (optional, string); Zoho Books path or required-query parameter 'salesorder_id'
  used by one or more documented endpoints.
- `tag_id` (optional, string); Zoho Books path or required-query parameter 'tag_id' used by one or
  more documented endpoints.
- `task_id` (optional, string); Zoho Books path or required-query parameter 'task_id' used by one or
  more documented endpoints.
- `tax_authority_id` (optional, string); Zoho Books path or required-query parameter
  'tax_authority_id' used by one or more documented endpoints.
- `tax_exemption_id` (optional, string); Zoho Books path or required-query parameter
  'tax_exemption_id' used by one or more documented endpoints.
- `tax_group_id` (optional, string); Zoho Books path or required-query parameter 'tax_group_id' used
  by one or more documented endpoints.
- `tax_id` (optional, string); Zoho Books path or required-query parameter 'tax_id' used by one or
  more documented endpoints.
- `time_entry_id` (optional, string); Zoho Books path or required-query parameter 'time_entry_id'
  used by one or more documented endpoints.
- `transaction_id` (optional, string); Zoho Books path or required-query parameter 'transaction_id'
  used by one or more documented endpoints.
- `transaction_type` (optional, string); Zoho Books path or required-query parameter
  'transaction_type' used by one or more documented endpoints.
- `user_id` (optional, string); Zoho Books path or required-query parameter 'user_id' used by one or
  more documented endpoints.
- `vendor_credit_id` (optional, string); Zoho Books path or required-query parameter
  'vendor_credit_id' used by one or more documented endpoints.
- `vendor_credit_refund_id` (optional, string); Zoho Books path or required-query parameter
  'vendor_credit_refund_id' used by one or more documented endpoints.
- `vendorpayment_refund_id` (optional, string); Zoho Books path or required-query parameter
  'vendorpayment_refund_id' used by one or more documented endpoints.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://www.zohoapis.com/books/v3`, `max_pages=0`,
`page_size=200`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Zoho-oauthtoken` using
  `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contacts`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 200.

- `contacts`: GET `/contacts` - records path `contacts`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `invoices`: GET `/invoices` - records path `invoices`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `items`: GET `/items` - records path `items`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `list_bank_accounts`: GET `/bankaccounts` - records path `bankaccounts`; query `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `get_bank_account`: GET `/bankaccounts/{{ config.account_id }}` - records path `bankaccount`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_last_imported_bank_statement`: GET `/bankaccounts/{{ config.account_id
  }}/statement/lastimported` - records path `statement`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `list_bank_account_rules`: GET `/bankaccounts/rules` - records path `rules`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_bank_account_rule`: GET `/bankaccounts/rules/{{ config.rule_id }}` - records path
  `target_accounts`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_bank_transactions`: GET `/banktransactions` - records path `banktransactions`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_matching_bank_transactions`: GET `/banktransactions/uncategorized/{{ config.transaction_id
  }}/match` - records path `matching_transactions`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_bank_transaction`: GET `/banktransactions/{{ config.bank_transaction_id }}` - records path
  `banktransaction`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_base_currency_adjustments`: GET `/basecurrencyadjustment` - records path
  `base_currency_adjustments`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `list_base_currency_adjustment_accounts`: GET `/basecurrencyadjustment/accounts` - records path
  `data`; query `adjustment_date` from template `{{ config.adjustment_date }}`; `currency_id` from
  template `{{ config.currency_id }}`; `exchange_rate` from template `{{ config.exchange_rate }}`;
  `notes` from template `{{ config.notes }}`; `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `list_base_currency_adjustment_contacts`: GET `/basecurrencyadjustment/contacts` - records path
  `contacts`; query `account_id` from template `{{ config.account_id }}`; `adjustment_date` from
  template `{{ config.adjustment_date }}`; `currency_id` from template `{{ config.currency_id }}`;
  `exchange_rate` from template `{{ config.exchange_rate }}`; `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_base_currency_adjustment`: GET `/basecurrencyadjustment/{{ config.base_currency_adjustment_id
  }}` - records path `data`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `list_bills`: GET `/bills` - records path `bills`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `convert_purchase_order_to_bill`: GET `/bills/editpage/frompurchaseorders` - records path
  `purchaseorder_ids`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; `purchaseorder_ids` from template `{{ config.purchaseorder_ids }}`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_bill`: GET `/bills/{{ config.bill_id }}` - records path `bill`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `get_bill_comments`: GET `/bills/{{ config.bill_id }}/comments` - records path `comments`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_bill_payments`: GET `/bills/{{ config.bill_id }}/payments` - records path `payments`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_chart_of_accounts`: GET `/chartofaccounts` - records path `chartofaccounts`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_chart_of_account_transactions`: GET `/chartofaccounts/accounttransactions` - records path
  `transactions`; query `account_id` from template `{{ config.account_id }}`; `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `get_chart_of_account`: GET `/chartofaccounts/{{ config.account_id }}` - records path
  `custom_fields`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_contact_persons`: GET `/contacts/{{ config.contact_id }}/contactpersons` - records path
  `contact_persons`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_contact_person`: GET `/contacts/{{ config.contact_id }}/contactpersons/{{
  config.contact_person_id }}` - records path `contact_person`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `get_contact`: GET `/contacts/{{ config.contact_id }}` - records path `contact`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_contact_address`: GET `/contacts/{{ config.contact_id }}/address` - records path `addresses`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_contact_autobill_recurring_invoices`: GET `/contacts/{{ config.contact_id }}/card/{{
  config.card_id }}/recurringinvoices` - records path `recurring_invoices`; query `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_contact_comments`: GET `/contacts/{{ config.contact_id }}/comments` - records path
  `contact_comments`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_unused_retainer_payments`: GET `/contacts/{{ config.contact_id
  }}/receivables/unusedretainerpayments` - records path `unused_credits_receivable`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_contact_credit_note_refunds`: GET `/contacts/{{ config.contact_id }}/refunds` - records path
  `creditnote_refunds`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `get_contact_statement_mail`: GET `/contacts/{{ config.contact_id }}/statements/email` - records
  path `data`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_credit_notes`: GET `/creditnotes` - records path `creditnotes`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_credit_note_refunds_of_all_credit_notes`: GET `/creditnotes/refunds` - records path
  `creditnote_refunds`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `get_credit_note_refund_by_id`: GET `/creditnotes/refunds/{{ config.refund_id }}` - records path
  `creditnote_refund`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_credit_note_templates`: GET `/creditnotes/templates` - records path `templates`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_credit_note`: GET `/creditnotes/{{ config.creditnote_id }}` - records path `creditnote`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_credit_note_comments`: GET `/creditnotes/{{ config.creditnote_id }}/comments` - records path
  `comments`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_credit_note_custom_fields`: GET `/creditnotes/{{ config.creditnote_id }}/customfields` -
  records path `custom_fields`; query `organization_id` from template `{{ config.organization_id
  }}`, omitted when absent; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits
  passthrough records.
- `get_credit_note_email`: GET `/creditnotes/{{ config.creditnote_id }}/email` - records path
  `data`; query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_credit_note_email_history`: GET `/creditnotes/{{ config.creditnote_id }}/emailhistory` -
  records path `email_history`; query `organization_id` from template `{{ config.organization_id
  }}`, omitted when absent; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits
  passthrough records.
- `list_invoices_of_credit_note`: GET `/creditnotes/{{ config.creditnote_id }}/invoices` - records
  path `invoices_credited`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `list_credit_note_refunds_of_a_credit_note`: GET `/creditnotes/{{ config.creditnote_id }}/refunds`
  - records path `creditnote_refunds`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_credit_note_refund`: GET `/creditnotes/{{ config.creditnote_id }}/refunds/{{
  config.creditnote_refund_id }}` - records path `creditnote_refund`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_currencies`: GET `/settings/currencies` - records path `currencies`; query `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `get_currency`: GET `/settings/currencies/{{ config.currency_id }}` - records path `currency`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_exchange_rates`: GET `/settings/currencies/{{ config.currency_id }}/exchangerates` - records
  path `exchange_rates`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `get_exchange_rate`: GET `/settings/currencies/{{ config.currency_id }}/exchangerates/{{
  config.exchange_rate_id }}` - records path `exchange_rate`; query `organization_id` from template
  `{{ config.organization_id }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.
- `list_custom_modules`: GET `/settings/modules` - records path `modules`; query `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `get_custom_module`: GET `/settings/modules/{{ config.module_api_name }}` - records path `module`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_custom_module_records`: GET `/{{ config.module_name }}` - records path `module_record`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_custom_module_record`: GET `/{{ config.module_name }}/{{ config.module_id }}` - records path
  `users`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_customer_debit_note`: GET `/invoices/{{ config.debit_note_id }}` - records path `invoice`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_customer_payments`: GET `/customerpayments` - records path `customer_payments`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_customer_payment_refunds`: GET `/customerpayments/{{ config.customer_payment_id }}/refunds`
  - records path `payment_refunds`; query `organization_id` from template `{{ config.organization_id
  }}`, omitted when absent; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits
  passthrough records.
- `get_customer_payment_refund`: GET `/customerpayments/{{ config.customer_payment_id }}/refunds/{{
  config.refund_id }}` - records path `payment_refunds`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_customer_payment`: GET `/customerpayments/{{ config.payment_id }}` - records path `payment`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_delivery_challans`: GET `/deliverychallans` - records path `deliverychallans`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_delivery_challan_templates`: GET `/deliverychallans/templates` - records path `templates`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_delivery_challan`: GET `/deliverychallans/{{ config.deliverychallan_id }}` - records path
  `deliverychallan`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_estimates`: GET `/estimates` - records path `estimates`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_estimate_templates`: GET `/estimates/templates` - records path `templates`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_estimate`: GET `/estimates/{{ config.estimate_id }}` - records path `estimate`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_estimate_comments`: GET `/estimates/{{ config.estimate_id }}/comments` - records path
  `comments`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_estimate_email`: GET `/estimates/{{ config.estimate_id }}/email` - records path `error_list`;
  query `email_template_id` from template `{{ config.email_template_id }}`; `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_employees`: GET `/employees` - records path `employees`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `get_employee`: GET `/employees/{{ config.employee_id }}` - records path `employee`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_expenses`: GET `/expenses` - records path `expenses`; query `organization_id` from template
  `{{ config.organization_id }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.
- `get_expense`: GET `/expenses/{{ config.expense_id }}` - records path `expense`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_expense_comments`: GET `/expenses/{{ config.expense_id }}/comments` - records path
  `comments`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_fixed_assets`: GET `/fixedassets` - records path `fixedassets`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `get_fixed_asset`: GET `/fixedassets/{{ config.fixed_asset_id }}` - records path `fissed-asset`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_fixed_asset_forecast`: GET `/fixedassets/{{ config.fixed_asset_id }}/forecast` - records path
  `data`; query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_fixed_asset_history`: GET `/fixedassets/{{ config.fixed_asset_id }}/history` - records path
  `asset_history`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_fixed_asset_type_list`: GET `/fixedassettypes` - records path `fixed_asset_types`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_invoice_templates`: GET `/invoices/templates` - records path `templates`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_invoice`: GET `/invoices/{{ config.invoice_id }}` - records path `invoice`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_invoice_comments`: GET `/invoices/{{ config.invoice_id }}/comments` - records path
  `comments`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_invoice_credits_applied`: GET `/invoices/{{ config.invoice_id }}/creditsapplied` - records
  path `credits`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_invoice_document_details`: GET `/invoices/{{ config.invoice_id }}/documents/{{
  config.document_id }}` - records path `document`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_invoice_email`: GET `/invoices/{{ config.invoice_id }}/email` - records path `data`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_payment_reminder_mail_content_for_invoice`: GET `/invoices/{{ config.invoice_id
  }}/paymentreminder` - records path `data`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `list_invoice_payments`: GET `/invoices/{{ config.invoice_id }}/payments` - records path
  `payments`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `generate_invoice_payment_link`: GET `/share/paymentlink` - records path `data`; query
  `expiry_time` from template `{{ config.expiry_time }}`; `link_type` from template `{{
  config.link_type }}`; `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; `transaction_id` from template `{{ config.transaction_id }}`; `transaction_type` from
  template `{{ config.transaction_type }}`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `list_item_details`: GET `/itemdetails` - records path `items`; query `item_ids` from template `{{
  config.item_ids }}`; `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_item`: GET `/items/{{ config.item_id }}` - records path `item`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_journals`: GET `/journals` - records path `journals`; query `organization_id` from template
  `{{ config.organization_id }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.
- `get_journal`: GET `/journals/{{ config.journal_id }}` - records path `journal`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_journal_credits`: GET `/journals/{{ config.journal_id }}/credits` - records path `credits`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_recurring_journals`: GET `/recurringjournals` - records path `recurring_journals`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_recurring_journal`: GET `/recurringjournals/{{ config.recurring_journal_id }}` - records path
  `journal`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_child_journals`: GET `/recurringjournals/{{ config.recurring_journal_id }}/journals` -
  records path `journals`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `get_transaction_journal_view`: GET `/transactions/{{ config.transaction_id }}/journals` - records
  path `journal`; query `entity_type` from template `{{ config.entity_type }}`; `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_locations`: GET `/locations` - records path `locations`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_opening_balance_transactions`: GET `/openingbalances/transactions` - records path
  `transactions`; query `account_id` from template `{{ config.account_id }}`; `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_opening_balance_details`: GET `/settings/openingbalancedetails` - records path
  `opening_balance`; query `account_id` from template `{{ config.account_id }}`; `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `get_opening_balance`: GET `/settings/openingbalances` - records path `opening_balance`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_organizations`: GET `/organizations` - records path `organizations`; query `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_organizations_for_user`: GET `/organizations/user` - records path `organizations`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_organization`: GET `/organizations/{{ config.organization_id }}` - records path
  `organization`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_pricebooks`: GET `/pricebooks` - records path `pricebooks`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_projects`: GET `/projects` - records path `projects`; query `organization_id` from template
  `{{ config.organization_id }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.
- `get_project`: GET `/projects/{{ config.project_id }}` - records path `project`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_project_comments`: GET `/projects/{{ config.project_id }}/comments` - records path
  `comments`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_project_invoices`: GET `/projects/{{ config.project_id }}/invoices` - records path
  `invoices`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_project_tasks`: GET `/projects/{{ config.project_id }}/tasks` - records path `tasks`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_project_task`: GET `/projects/{{ config.project_id }}/tasks/{{ config.task_id }}` - records
  path `task`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_project_users`: GET `/projects/{{ config.project_id }}/users` - records path `users`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_project_user`: GET `/projects/{{ config.project_id }}/users/{{ config.user_id }}` - records
  path `user`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_purchase_orders`: GET `/purchaseorders` - records path `purchaseorders`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_purchase_order_templates`: GET `/purchaseorders/templates` - records path `templates`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_purchase_order`: GET `/purchaseorders/{{ config.purchaseorder_id }}` - records path
  `purchaseorder`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_purchase_order_comments`: GET `/purchaseorders/{{ config.purchaseorder_id }}/comments` -
  records path `comments`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `get_recurring_bill`: GET `/recurring_bills/{{ config.recurring_bill_id }}` - records path
  `recurring_bill`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_recurring_bills`: GET `/recurringbills` - records path `recurringbills`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_recurring_expenses`: GET `/recurringexpenses` - records path `recurring_expenses`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_recurring_expense`: GET `/recurringexpenses/{{ config.recurring_expense_id }}` - records path
  `recurring_expense`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_recurring_expense_history`: GET `/recurringexpenses/{{ config.recurring_expense_id
  }}/comments` - records path `comments`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `list_child_expenses_of_recurring_expense`: GET `/recurringexpenses/{{ config.recurring_expense_id
  }}/expenses` - records path `expensehistory`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `list_recurring_invoices`: GET `/recurringinvoices` - records path `recurring_invoices`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_recurring_invoice`: GET `/recurringinvoices/{{ config.recurring_invoice_id }}` - records path
  `recurring_invoice`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_recurring_invoice_history`: GET `/recurringinvoices/{{ config.recurring_invoice_id
  }}/comments` - records path `comments`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `list_recurring_invoice_child_invoices`: GET `/recurringinvoices/{{ config.recurring_invoice_id
  }}/invoices` - records path `invoice_history`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_register_budget_vs_actuals`: GET `/registers/{{ config.account_id }}/budgetvsactuals` -
  records path `register_budget_vs_actual`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `list_register_bulk_action_history`: GET `/registers/{{ config.account_id }}/bulkhistory` -
  records path `bulk_actions_history`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_register_bulk_action_history`: GET `/registers/{{ config.account_id }}/bulkhistory/{{
  config.comment_id }}` - records path `bulk_action_details`; query `organization_id` from template
  `{{ config.organization_id }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.
- `get_register_bulk_update_editpage`: GET `/registers/{{ config.account_id }}/bulkupdate/editpage`
  - records path `fields`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `list_register_transactions`: GET `/registers/{{ config.account_id }}/transactions` - records path
  `register_transactions`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `get_tags`: GET `/reportingtags` - records path `reporting_tags`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `all_tag_options`: GET `/reportingtags/(\d+)/options/all` - records path `results`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; `tag_id` from
  template `{{ config.tag_id }}`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits
  passthrough records.
- `get_all_tag_options`: GET `/reportingtags/options` - records path `dependent_tags`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; `tag_id` from
  template `{{ config.tag_id }}`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits
  passthrough records.
- `list_retainer_invoices`: GET `/retainerinvoices` - records path `retainerinvoices`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_retainer_invoice_templates`: GET `/retainerinvoices/templates` - records path `templates`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_retainer_invoice`: GET `/retainerinvoices/{{ config.retainerinvoice_id }}` - records path
  `retainerinvoice`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_retainer_invoice`: GET `/retainerinvoices/{{ config.retainerinvoice_id }}/comments` -
  records path `comments`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `get_retainer_invoice_email`: GET `/retainerinvoices/{{ config.retainerinvoice_id }}/email` -
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_sales_orders`: GET `/salesorders` - records path `salesorders`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_sales_order_templates`: GET `/salesorders/templates` - records path `templates`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_sales_order`: GET `/salesorders/{{ config.salesorder_id }}` - records path `salesorder`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_sales_order_email`: GET `/salesorders/{{ config.salesorder_id }}/email` - records path
  `data`; query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_sales_receipts`: GET `/salesreceipts` - records path `salesreceipts`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_sales_receipt`: GET `/salesreceipts/{{ config.sales_receipt_id }}` - records path
  `salesreceipt`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_tasks`: GET `/tasks` - records path `tasks`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_task`: GET `/tasks/{{ config.task_id }}` - records path `task`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_task_comments`: GET `/tasks/{{ config.task_id }}/comments` - records path `comments`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_task_document`: GET `/tasks/{{ config.task_id }}/documents/{{ config.document_id }}` -
  records path `documents`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `list_tax_authorities`: GET `/settings/taxauthorities` - records path `tax_authorities`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_tax_authority`: GET `/settings/taxauthorities/{{ config.tax_authority_id }}` - records path
  `tax_authority`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_taxes`: GET `/settings/taxes` - records path `taxes`; query `organization_id` from template
  `{{ config.organization_id }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.
- `get_tax`: GET `/settings/taxes/{{ config.tax_id }}` - records path `tax`; query `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_tax_exemptions`: GET `/settings/taxexemptions` - records path `tax_exemptions`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_tax_exemption`: GET `/settings/taxexemptions/{{ config.tax_exemption_id }}` - records path
  `tax_exemption`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_tax_group`: GET `/settings/taxgroups/{{ config.tax_group_id }}` - records path `tax_group`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_time_entries`: GET `/projects/timeentries` - records path `time_entries`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_running_timer`: GET `/projects/timeentries/runningtimer/me` - records path `time_entry`;
  query `organization_id` from template `{{ config.organization_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_time_entry`: GET `/projects/timeentries/{{ config.time_entry_id }}` - records path
  `time_entry`; query `organization_id` from template `{{ config.organization_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_accounting_period_transaction_lock`: GET `/accountingperiods/transactionlock` - records path
  `transaction_lock`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_transaction_lock`: GET `/transactionlock` - records path `transaction_lock`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_transaction_locks`: GET `/transactionlocks` - records path `transaction_locks`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_users`: GET `/users` - records path `users`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_current_user`: GET `/users/me` - records path `user`; query `organization_id` from template
  `{{ config.organization_id }}`, omitted when absent; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.
- `get_user`: GET `/users/{{ config.user_id }}` - records path `user`; query `organization_id` from
  template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_vendor_credits`: GET `/vendorcredits` - records path `vendorcredits`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_vendor_credit_refunds_of_all_vendor_credits`: GET `/vendorcredits/refunds` - records path
  `vendor_credit_refunds`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `get_vendor_credit`: GET `/vendorcredits/{{ config.vendor_credit_id }}` - records path
  `vendor_credit`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_bills_credited`: GET `/vendorcredits/{{ config.vendor_credit_id }}/bills` - records path
  `bills_credited`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_vendor_credit_comments`: GET `/vendorcredits/{{ config.vendor_credit_id }}/comments` -
  records path `comments`; query `organization_id` from template `{{ config.organization_id }}`,
  omitted when absent; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough
  records.
- `list_vendor_credit_refunds_of_a_vendor_credit`: GET `/vendorcredits/{{ config.vendor_credit_id
  }}/refunds` - records path `vendor_credit_refunds`; query `organization_id` from template `{{
  config.organization_id }}`, omitted when absent; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `get_vendor_credit_refund`: GET `/vendorcredits/{{ config.vendor_credit_id }}/refunds/{{
  config.vendor_credit_refund_id }}` - records path `vendor_credit_refund`; query `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `list_vendor_payments`: GET `/vendorpayments` - records path `vendorpayments`; query
  `organization_id` from template `{{ config.organization_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed
  output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_vendor_payment`: GET `/vendorpayments/{{ config.payment_id }}` - records path
  `vendorpayment`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `get_vendor_payment_email_content`: GET `/vendorpayments/{{ config.payment_id }}/email` - records
  path `to_contacts`; query `organization_id` from template `{{ config.organization_id }}`, omitted
  when absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 200; computed output fields `id`, `name`, `updated_at`; emits passthrough records.
- `list_vendor_payment_refunds`: GET `/vendorpayments/{{ config.payment_id }}/refunds` - records
  path `vendorpayment_refunds`; query `organization_id` from template `{{ config.organization_id
  }}`, omitted when absent; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 200; computed output fields `id`, `name`, `updated_at`; emits
  passthrough records.
- `get_vendor_payment_refund`: GET `/vendorpayments/{{ config.payment_id }}/refunds/{{
  config.vendorpayment_refund_id }}` - records path `vendorpayment_refund`; query `organization_id`
  from template `{{ config.organization_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.

## Write actions & risks

Overall write risk: external Zoho Books API mutation of accounting, contact, inventory, project,
banking, tax, and organization-adjacent records.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_bank_account`: POST `/bankaccounts?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_bank_account`: PUT `/bankaccounts/{{ record.account_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `account_id`; required
  record fields `account_id`; accepted fields `account_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_bank_account`: DELETE `/bankaccounts/{{ record.account_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `account_id`; required
  record fields `account_id`; accepted fields `account_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `mark_bank_account_active`: POST `/bankaccounts/{{ record.account_id }}/active?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `account_id`; required
  record fields `account_id`; accepted fields `account_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_bank_account_inactive`: POST `/bankaccounts/{{ record.account_id
  }}/inactive?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `account_id`; required record fields `account_id`; accepted fields `account_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `update_bank_account_preferences`: PUT `/bankaccounts/{{ record.account_id
  }}/preferences?organization_id={{ config.organization_id }}` - kind `update`; body type `none`;
  path fields `account_id`; required record fields `account_id`; accepted fields `account_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `create_bank_reconciliation`: POST `/bankaccounts/{{ record.account_id
  }}/reconciliations?organization_id={{ config.organization_id }}` - kind `create`; body type
  `json`; path fields `account_id`; required record fields `account_id`; accepted fields
  `account_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_bank_reconciliation`: PUT `/bankaccounts/{{ record.account_id }}/reconciliations/{{
  record.reconciliation_id }}?organization_id={{ config.organization_id }}` - kind `update`; body
  type `json`; path fields `account_id`, `reconciliation_id`; required record fields `account_id`,
  `reconciliation_id`; accepted fields `account_id`, `reconciliation_id`; risk: external mutation in
  Zoho Books accounting data; approval required.
- `delete_bank_reconciliation`: DELETE `/bankaccounts/{{ record.account_id }}/reconciliations/{{
  record.reconciliation_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body
  type `none`; path fields `account_id`, `reconciliation_id`; required record fields `account_id`,
  `reconciliation_id`; accepted fields `account_id`, `reconciliation_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho
  Books; approval required.
- `add_bank_reconciliation_attachment`: POST `/bankaccounts/{{ record.account_id
  }}/reconciliations/{{ record.reconciliation_id }}/attachment?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `account_id`,
  `reconciliation_id`; required record fields `account_id`, `reconciliation_id`; accepted fields
  `account_id`, `reconciliation_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_bank_reconciliation_document`: DELETE `/bankaccounts/{{ record.account_id
  }}/reconciliations/{{ record.reconciliation_id }}/documents/{{ record.document_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `account_id`, `reconciliation_id`, `document_id`; required record fields `account_id`,
  `reconciliation_id`, `document_id`; accepted fields `account_id`, `document_id`,
  `reconciliation_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `save_bank_reconciliation_draft`: PUT `/bankaccounts/{{ record.account_id }}/reconciliations/{{
  record.reconciliation_id }}/draft?organization_id={{ config.organization_id }}` - kind `update`;
  body type `json`; path fields `account_id`, `reconciliation_id`; required record fields
  `account_id`, `reconciliation_id`; accepted fields `account_id`, `reconciliation_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `delete_last_imported_bank_statement`: DELETE `/bankaccounts/{{ record.account_id }}/statement/{{
  record.statement_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `account_id`, `statement_id`; required record fields `account_id`,
  `statement_id`; accepted fields `account_id`, `statement_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `import_bank_statements`: POST `/bankstatements?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `create_bank_account_match_filter`: POST `/bankaccounts/matchfilters?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `update_bank_account_match_filter`: PUT `/bankaccounts/matchfilters/{{ record.match_filter_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `match_filter_id`; required record fields `match_filter_id`; accepted fields `match_filter_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `delete_bank_account_match_filter`: DELETE `/bankaccounts/matchfilters/{{ record.match_filter_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `match_filter_id`; required record fields `match_filter_id`; accepted fields `match_filter_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `create_bank_account_rule`: POST `/bankaccounts/rules?organization_id={{ config.organization_id
  }}` - kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `bulk_update_bank_account_rules`: PUT `/bankaccounts/rules?organization_id={{
  config.organization_id }}` - kind `update`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `bulk_delete_bank_account_rules`: DELETE `/bankaccounts/rules?organization_id={{
  config.organization_id }}&rule_ids={{ record.rule_ids }}` - kind `delete`; body type `none`; path
  fields `rule_ids`; required record fields `rule_ids`; accepted fields `rule_ids`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `reorder_bank_account_rules`: POST `/bankaccounts/rules/order?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `skip_suggested_bank_account_rule`: POST `/bankaccounts/rules/skipsuggest?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `update_bank_account_rule`: PUT `/bankaccounts/rules/{{ record.rule_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `rule_id`; required
  record fields `rule_id`; accepted fields `rule_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_bank_account_rule`: DELETE `/bankaccounts/rules/{{ record.rule_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `rule_id`; required
  record fields `rule_id`; accepted fields `rule_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books; approval
  required.
- `create_bank_transaction`: POST `/banktransactions?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `categorize_bank_transaction_as_payment_refund`: POST `/banktransactions/uncategorized/{{
  record.statement_line_id }}/categorize/paymentrefunds?organization_id={{ config.organization_id
  }}` - kind `create`; body type `json`; path fields `statement_line_id`; required record fields
  `statement_line_id`; accepted fields `statement_line_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `categorize_as_vendor_payment_refund`: POST `/banktransactions/uncategorized/{{
  record.statement_line_id }}/categorize/vendorpaymentrefunds?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `statement_line_id`;
  required record fields `statement_line_id`; accepted fields `statement_line_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `categorize_bank_transaction`: POST `/banktransactions/uncategorized/{{ record.transaction_id
  }}/categorize?organization_id={{ config.organization_id }}` - kind `create`; body type `json`;
  path fields `transaction_id`; required record fields `transaction_id`; accepted fields
  `transaction_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `categorize_as_credit_note_refunds`: POST `/banktransactions/uncategorized/{{
  record.transaction_id }}/categorize/creditnoterefunds?organization_id={{ config.organization_id
  }}` - kind `create`; body type `json`; path fields `transaction_id`; required record fields
  `transaction_id`; accepted fields `transaction_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `categorize_bank_transaction_as_customer_payment`: POST `/banktransactions/uncategorized/{{
  record.transaction_id }}/categorize/customerpayments?organization_id={{ config.organization_id }}`
  - kind `create`; body type `json`; path fields `transaction_id`; required record fields
  `transaction_id`; accepted fields `transaction_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `categorize_bank_transaction_as_expense`: POST `/banktransactions/uncategorized/{{
  record.transaction_id }}/categorize/expenses?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; path fields `transaction_id`; required record fields `transaction_id`;
  accepted fields `transaction_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `categorize_as_vendor_credit_refunds`: POST `/banktransactions/uncategorized/{{
  record.transaction_id }}/categorize/vendorcreditrefunds?organization_id={{ config.organization_id
  }}` - kind `create`; body type `json`; path fields `transaction_id`; required record fields
  `transaction_id`; accepted fields `transaction_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `categorize_bank_transaction_as_vendor_payment`: POST `/banktransactions/uncategorized/{{
  record.transaction_id }}/categorize/vendorpayments?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; path fields `transaction_id`; required record fields
  `transaction_id`; accepted fields `transaction_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `exclude_bank_transaction`: POST `/banktransactions/uncategorized/{{ record.transaction_id
  }}/exclude?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `transaction_id`; required record fields `transaction_id`; accepted fields
  `transaction_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `match_bank_transaction`: POST `/banktransactions/uncategorized/{{ record.transaction_id
  }}/match?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `transaction_id`; required record fields `transaction_id`; accepted fields
  `transaction_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `restore_bank_transaction`: POST `/banktransactions/uncategorized/{{ record.transaction_id
  }}/restore?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `transaction_id`; required record fields `transaction_id`; accepted fields
  `transaction_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_bank_transaction`: PUT `/banktransactions/{{ record.bank_transaction_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `bank_transaction_id`; required record fields `bank_transaction_id`; accepted fields
  `bank_transaction_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_bank_transaction`: DELETE `/banktransactions/{{ record.bank_transaction_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `bank_transaction_id`; required record fields `bank_transaction_id`; accepted fields
  `bank_transaction_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `uncategorize_bank_transaction`: POST `/banktransactions/{{ record.transaction_id
  }}/uncategorize?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `transaction_id`; required record fields `transaction_id`; accepted fields
  `transaction_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `unmatch_bank_transaction`: POST `/banktransactions/{{ record.transaction_id
  }}/unmatch?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `transaction_id`; required record fields `transaction_id`; accepted fields
  `transaction_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_base_currency_adjustment`: POST `/basecurrencyadjustment?organization_id={{
  config.organization_id }}&account_ids={{ record.account_ids }}` - kind `create`; body type `json`;
  path fields `account_ids`; required record fields `account_ids`; accepted fields `account_ids`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `bulk_delete_base_currency_adjustments`: DELETE
  `/basecurrencyadjustment/bulkdelete?organization_id={{ config.organization_id
  }}&base_currency_adjustment_ids={{ record.base_currency_adjustment_ids }}` - kind `delete`; body
  type `none`; path fields `base_currency_adjustment_ids`; required record fields
  `base_currency_adjustment_ids`; accepted fields `base_currency_adjustment_ids`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `delete_base_currency_adjustment`: DELETE `/basecurrencyadjustment/{{
  record.base_currency_adjustment_id }}?organization_id={{ config.organization_id }}` - kind
  `delete`; body type `none`; path fields `base_currency_adjustment_id`; required record fields
  `base_currency_adjustment_id`; accepted fields `base_currency_adjustment_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `reevaluate_base_currency_adjustment`: POST `/basecurrencyadjustment/{{
  record.base_currency_adjustment_id }}/reevaluate?organization_id={{ config.organization_id }}` -
  kind `create`; body type `none`; path fields `base_currency_adjustment_id`; required record fields
  `base_currency_adjustment_id`; accepted fields `base_currency_adjustment_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `update_custom_fields_in_bill`: PUT `/bill/{{ record.bill_id }}/customfields?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `bill_id`; required
  record fields `bill_id`; accepted fields `bill_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `create_bill`: POST `/bills?organization_id={{ config.organization_id }}` - kind `create`; body
  type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_bill`: PUT `/bills/{{ record.bill_id }}?organization_id={{ config.organization_id }}` -
  kind `update`; body type `json`; path fields `bill_id`; required record fields `bill_id`; accepted
  fields `bill_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_bill`: DELETE `/bills/{{ record.bill_id }}?organization_id={{ config.organization_id }}` -
  kind `delete`; body type `none`; path fields `bill_id`; required record fields `bill_id`; accepted
  fields `bill_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: external destructive mutation in Zoho Books; approval required.
- `update_bill_billing_address`: PUT `/bills/{{ record.bill_id }}/address/billing?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `bill_id`; required
  record fields `bill_id`; accepted fields `bill_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `approve_bill`: POST `/bills/{{ record.bill_id }}/approve?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `bill_id`; required
  record fields `bill_id`; accepted fields `bill_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_bill_attachment`: POST `/bills/{{ record.bill_id }}/attachment?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `bill_id`; required
  record fields `bill_id`; accepted fields `bill_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_bill_attachment`: DELETE `/bills/{{ record.bill_id }}/attachment?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `bill_id`; required
  record fields `bill_id`; accepted fields `bill_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books; approval
  required.
- `add_bill_comment`: POST `/bills/{{ record.bill_id }}/comments?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `bill_id`; required
  record fields `bill_id`; accepted fields `bill_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_bill_comment`: DELETE `/bills/{{ record.bill_id }}/comments/{{ record.comment_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `bill_id`, `comment_id`; required record fields `bill_id`, `comment_id`; accepted fields
  `bill_id`, `comment_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `apply_credits_to_bill`: POST `/bills/{{ record.bill_id }}/credits?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `bill_id`; required
  record fields `bill_id`; accepted fields `bill_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_bill_payment`: DELETE `/bills/{{ record.bill_id }}/payments/{{ record.bill_payment_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `bill_id`, `bill_payment_id`; required record fields `bill_id`, `bill_payment_id`; accepted fields
  `bill_id`, `bill_payment_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `mark_bill_open`: POST `/bills/{{ record.bill_id }}/status/open?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `bill_id`; required
  record fields `bill_id`; accepted fields `bill_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_bill_void`: POST `/bills/{{ record.bill_id }}/status/void?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `bill_id`; required
  record fields `bill_id`; accepted fields `bill_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `submit_bill`: POST `/bills/{{ record.bill_id }}/submit?organization_id={{ config.organization_id
  }}` - kind `create`; body type `none`; path fields `bill_id`; required record fields `bill_id`;
  accepted fields `bill_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `create_chart_of_account`: POST `/chartofaccounts?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `bulk_mark_chart_of_accounts_active`: POST `/chartofaccounts/active?organization_id={{
  config.organization_id }}&account_ids={{ record.account_ids }}` - kind `create`; body type `none`;
  path fields `account_ids`; required record fields `account_ids`; accepted fields `account_ids`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `bulk_delete_chart_of_accounts`: DELETE `/chartofaccounts/bulkdelete?organization_id={{
  config.organization_id }}&account_ids={{ record.account_ids }}` - kind `delete`; body type `none`;
  path fields `account_ids`; required record fields `account_ids`; accepted fields `account_ids`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `bulk_mark_chart_of_accounts_inactive`: POST `/chartofaccounts/inactive?organization_id={{
  config.organization_id }}&account_ids={{ record.account_ids }}` - kind `create`; body type `none`;
  path fields `account_ids`; required record fields `account_ids`; accepted fields `account_ids`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `delete_chart_of_account_transaction`: DELETE `/chartofaccounts/transactions/{{
  record.transaction_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `transaction_id`; required record fields `transaction_id`; accepted fields
  `transaction_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: external destructive mutation in Zoho Books; approval required.
- `update_chart_of_account`: PUT `/chartofaccounts/{{ record.account_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `account_id`; required
  record fields `account_id`; accepted fields `account_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_chart_of_account`: DELETE `/chartofaccounts/{{ record.account_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `account_id`; required
  record fields `account_id`; accepted fields `account_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `mark_chart_of_account_active`: POST `/chartofaccounts/{{ record.account_id
  }}/active?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `account_id`; required record fields `account_id`; accepted fields `account_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `mark_chart_of_account_inactive`: POST `/chartofaccounts/{{ record.account_id
  }}/inactive?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `account_id`; required record fields `account_id`; accepted fields `account_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `create_contact_person`: POST `/contacts/contactpersons?organization_id={{ config.organization_id
  }}` - kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `update_contact_person`: PUT `/contacts/contactpersons/{{ record.contact_person_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `contact_person_id`; required record fields `contact_person_id`; accepted fields
  `contact_person_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_contact_person`: DELETE `/contacts/contactpersons/{{ record.contact_person_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `contact_person_id`; required record fields `contact_person_id`; accepted fields
  `contact_person_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `mark_contact_person_primary`: POST `/contacts/contactpersons/{{ record.contact_person_id
  }}/primary?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `contact_person_id`; required record fields `contact_person_id`; accepted fields
  `contact_person_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_contact`: POST `/contacts?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_contacts`: DELETE `/contacts?organization_id={{ config.organization_id }}&contact_ids={{
  record.contact_ids }}` - kind `delete`; body type `none`; path fields `contact_ids`; required
  record fields `contact_ids`; accepted fields `contact_ids`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `create_contact_person_2`: POST `/contacts/contactpersons?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `update_contact_person_2`: PUT `/contacts/contactpersons/{{ record.contactperson_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `contactperson_id`; required record fields `contactperson_id`; accepted fields `contactperson_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `delete_contact_person_2`: DELETE `/contacts/contactpersons/{{ record.contactperson_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `contactperson_id`; required record fields `contactperson_id`; accepted fields `contactperson_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `invite_contact_person_to_portal`: POST `/contacts/contactpersons/{{ record.contactperson_id
  }}/portal/invite?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `contactperson_id`; required record fields `contactperson_id`; accepted fields
  `contactperson_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `resend_contact_person_portal_invite`: POST `/contacts/contactpersons/{{ record.contactperson_id
  }}/portal/invite/resend?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `contactperson_id`; required record fields `contactperson_id`; accepted fields
  `contactperson_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_contact_person_primary_2`: POST `/contacts/contactpersons/{{ record.contactperson_id
  }}/primary?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `contactperson_id`; required record fields `contactperson_id`; accepted fields
  `contactperson_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `disable_contact_person_sms`: POST `/contacts/contactpersons/{{ record.contactperson_id
  }}/sms/disable?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `contactperson_id`; required record fields `contactperson_id`; accepted fields
  `contactperson_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `enable_contact_person_sms`: POST `/contacts/contactpersons/{{ record.contactperson_id
  }}/sms/enable?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `contactperson_id`; required record fields `contactperson_id`; accepted fields
  `contactperson_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `restore_contact_documents`: POST `/contacts/documents/restore?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `assign_owner_to_contacts`: POST `/contacts/owner?organization_id={{ config.organization_id
  }}&contact_ids={{ record.contact_ids }}` - kind `create`; body type `none`; path fields
  `contact_ids`; required record fields `contact_ids`; accepted fields `contact_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `send_contacts_sms`: POST `/contacts/sms?organization_id={{ config.organization_id }}` - kind
  `create`; body type `none`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `mark_contacts_for_1099_tracking`: POST `/contacts/track1099?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `update_contact`: PUT `/contacts/{{ record.contact_id }}?organization_id={{ config.organization_id
  }}` - kind `update`; body type `json`; path fields `contact_id`; required record fields
  `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_contact`: DELETE `/contacts/{{ record.contact_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `mark_contact_active`: POST `/contacts/{{ record.contact_id }}/active?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_contact_address`: POST `/contacts/{{ record.contact_id }}/address?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_contact_address`: PUT `/contacts/{{ record.contact_id }}/address/{{ record.address_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `contact_id`, `address_id`; required record fields `contact_id`, `address_id`; accepted fields
  `address_id`, `contact_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_contact_address`: DELETE `/contacts/{{ record.contact_id }}/address/{{ record.address_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `contact_id`, `address_id`; required record fields `contact_id`, `address_id`; accepted fields
  `address_id`, `contact_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `mark_contact_address_as_billing`: POST `/contacts/{{ record.contact_id }}/address/{{
  record.address_id }}/markasbilling?organization_id={{ config.organization_id }}` - kind `create`;
  body type `none`; path fields `contact_id`, `address_id`; required record fields `contact_id`,
  `address_id`; accepted fields `address_id`, `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_contact_address_as_shipping`: POST `/contacts/{{ record.contact_id }}/address/{{
  record.address_id }}/markasshipping?organization_id={{ config.organization_id }}` - kind `create`;
  body type `none`; path fields `contact_id`, `address_id`; required record fields `contact_id`,
  `address_id`; accepted fields `address_id`, `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `verify_contact_address_by_id`: POST `/contacts/{{ record.contact_id }}/address/{{
  record.address_id }}/verify?organization_id={{ config.organization_id }}` - kind `create`; body
  type `none`; path fields `contact_id`, `address_id`; required record fields `contact_id`,
  `address_id`; accepted fields `address_id`, `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_contact_attachment`: POST `/contacts/{{ record.contact_id }}/attachment?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_contact_bank_account`: POST `/contacts/{{ record.contact_id }}/bankaccount?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_contact_bank_account`: PUT `/contacts/{{ record.contact_id }}/bankaccount/{{
  record.bank_account_id }}?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `contact_id`, `bank_account_id`; required record fields `contact_id`,
  `bank_account_id`; accepted fields `bank_account_id`, `contact_id`; risk: external mutation in
  Zoho Books accounting data; approval required.
- `delete_contact_bank_account`: DELETE `/contacts/{{ record.contact_id }}/bankaccount/{{
  record.bank_account_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `contact_id`, `bank_account_id`; required record fields `contact_id`,
  `bank_account_id`; accepted fields `bank_account_id`, `contact_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho
  Books; approval required.
- `approve_contact_bank_account`: POST `/contacts/{{ record.contact_id }}/bankaccount/{{
  record.bank_account_id }}/approve?organization_id={{ config.organization_id }}` - kind `create`;
  body type `none`; path fields `contact_id`, `bank_account_id`; required record fields
  `contact_id`, `bank_account_id`; accepted fields `bank_account_id`, `contact_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `decline_contact_bank_account`: POST `/contacts/{{ record.contact_id }}/bankaccount/{{
  record.bank_account_id }}/decline?organization_id={{ config.organization_id }}` - kind `create`;
  body type `none`; path fields `contact_id`, `bank_account_id`; required record fields
  `contact_id`, `bank_account_id`; accepted fields `bank_account_id`, `contact_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `verify_contact_bank_account`: POST `/contacts/{{ record.contact_id }}/bankaccount/{{
  record.bank_account_id }}/verify?organization_id={{ config.organization_id }}` - kind `create`;
  body type `none`; path fields `contact_id`, `bank_account_id`; required record fields
  `contact_id`, `bank_account_id`; accepted fields `bank_account_id`, `contact_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `add_contact_card`: POST `/contacts/{{ record.contact_id }}/card?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_contact_card`: PUT `/contacts/{{ record.contact_id }}/card/{{ record.card_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `contact_id`, `card_id`; required record fields `contact_id`, `card_id`; accepted fields
  `card_id`, `contact_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_contact_card`: DELETE `/contacts/{{ record.contact_id }}/card/{{ record.card_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `contact_id`, `card_id`; required record fields `contact_id`, `card_id`; accepted fields
  `card_id`, `contact_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `send_contact_client_review_email`: POST `/contacts/{{ record.contact_id
  }}/clientreviews/email?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `contact_id`; required record fields `contact_id`; accepted fields
  `contact_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `add_contact_comment`: POST `/contacts/{{ record.contact_id }}/comments?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_contact_comment`: DELETE `/contacts/{{ record.contact_id }}/comments/{{ record.comment_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `contact_id`, `comment_id`; required record fields `contact_id`, `comment_id`; accepted fields
  `comment_id`, `contact_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `update_contact_document`: PUT `/contacts/{{ record.contact_id }}/documents/{{ record.document_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `none`; path fields
  `contact_id`, `document_id`; required record fields `contact_id`, `document_id`; accepted fields
  `contact_id`, `document_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_contact_document`: DELETE `/contacts/{{ record.contact_id }}/documents/{{
  record.document_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `contact_id`, `document_id`; required record fields `contact_id`,
  `document_id`; accepted fields `contact_id`, `document_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `verify_contact_einvoice`: POST `/contacts/{{ record.contact_id
  }}/einvoice/verify?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `contact_id`; required record fields `contact_id`; accepted fields
  `contact_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `email_contact`: POST `/contacts/{{ record.contact_id }}/email?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_contact_inactive`: POST `/contacts/{{ record.contact_id }}/inactive?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `merge_contact`: POST `/contacts/{{ record.contact_id }}/merge?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `assign_contact_owner`: POST `/contacts/{{ record.contact_id }}/owner?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `send_contact_payment_method_email`: POST `/contacts/{{ record.contact_id
  }}/paymentmethod/email?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `contact_id`; required record fields `contact_id`; accepted fields
  `contact_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `disable_contact_payment_reminder`: POST `/contacts/{{ record.contact_id
  }}/paymentreminder/disable?organization_id={{ config.organization_id }}` - kind `create`; body
  type `none`; path fields `contact_id`; required record fields `contact_id`; accepted fields
  `contact_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `enable_contact_payment_reminder`: POST `/contacts/{{ record.contact_id
  }}/paymentreminder/enable?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `contact_id`; required record fields `contact_id`; accepted fields
  `contact_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `disable_contact_portal`: POST `/contacts/{{ record.contact_id
  }}/portal/disable?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `contact_id`; required record fields `contact_id`; accepted fields `contact_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `enable_contact_portal`: POST `/contacts/{{ record.contact_id }}/portal/enable?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `send_contact_sms`: POST `/contacts/{{ record.contact_id }}/sms?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `email_contact_statement`: POST `/contacts/{{ record.contact_id
  }}/statements/email?organization_id={{ config.organization_id }}` - kind `create`; body type
  `json`; path fields `contact_id`; required record fields `contact_id`; accepted fields
  `contact_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_contact_tags`: PUT `/contacts/{{ record.contact_id }}/tags?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_contact_tag`: DELETE `/contacts/{{ record.contact_id }}/tags/{{ record.tag_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `contact_id`, `tag_id`; required record fields `contact_id`, `tag_id`; accepted fields
  `contact_id`, `tag_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `add_contact_tax_info`: POST `/contacts/{{ record.contact_id }}/taxinfo?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_contact_tax_info`: PUT `/contacts/{{ record.contact_id }}/taxinfo/{{ record.tax_info_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `contact_id`, `tax_info_id`; required record fields `contact_id`, `tax_info_id`; accepted fields
  `contact_id`, `tax_info_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_contact_tax_info`: DELETE `/contacts/{{ record.contact_id }}/taxinfo/{{ record.tax_info_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `contact_id`, `tax_info_id`; required record fields `contact_id`, `tax_info_id`; accepted fields
  `contact_id`, `tax_info_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `track_contact_1099`: POST `/contacts/{{ record.contact_id }}/track1099?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_contact_trn_status`: POST `/contacts/{{ record.contact_id }}/trnstatus?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `untrack_contact_1099`: POST `/contacts/{{ record.contact_id }}/untrack1099?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `contact_id`; required
  record fields `contact_id`; accepted fields `contact_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `send_contact_vendor_statement_email`: POST `/contacts/{{ record.contact_id
  }}/vendorstatements/email?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `contact_id`; required record fields `contact_id`; accepted fields
  `contact_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_credit_note`: POST `/creditnotes?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `approve_credit_notes`: POST `/creditnotes/approve?organization_id={{ config.organization_id
  }}&creditnote_ids={{ record.creditnote_ids }}` - kind `create`; body type `none`; path fields
  `creditnote_ids`; required record fields `creditnote_ids`; accepted fields `creditnote_ids`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `cancel_credit_notes_einvoice`: POST `/creditnotes/einvoice/cancel?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `push_credit_notes_einvoice`: POST `/creditnotes/einvoice/push?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `submit_credit_notes`: POST `/creditnotes/submit?organization_id={{ config.organization_id
  }}&creditnote_ids={{ record.creditnote_ids }}` - kind `create`; body type `none`; path fields
  `creditnote_ids`; required record fields `creditnote_ids`; accepted fields `creditnote_ids`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `update_credit_note`: PUT `/creditnotes/{{ record.creditnote_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `creditnote_id`;
  required record fields `creditnote_id`; accepted fields `creditnote_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `delete_credit_note`: DELETE `/creditnotes/{{ record.creditnote_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `creditnote_id`;
  required record fields `creditnote_id`; accepted fields `creditnote_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: external destructive mutation in
  Zoho Books; approval required.
- `update_credit_note_billing_address`: PUT `/creditnotes/{{ record.creditnote_id
  }}/address/billing?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_credit_note_shipping_address`: PUT `/creditnotes/{{ record.creditnote_id
  }}/address/shipping?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `approve_credit_note`: POST `/creditnotes/{{ record.creditnote_id }}/approve?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `creditnote_id`;
  required record fields `creditnote_id`; accepted fields `creditnote_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `finalize_credit_note_approval`: POST `/creditnotes/{{ record.creditnote_id
  }}/approve/final?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `add_credit_note_attachment`: POST `/creditnotes/{{ record.creditnote_id
  }}/attachment?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_credit_note_cfdi_status`: POST `/creditnotes/{{ record.creditnote_id
  }}/cfdi/status?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `add_credit_note_comment`: POST `/creditnotes/{{ record.creditnote_id
  }}/comments?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `creditnote_id`; required record fields `creditnote_id`; accepted fields `creditnote_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `delete_credit_note_comment`: DELETE `/creditnotes/{{ record.creditnote_id }}/comments/{{
  record.comment_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `creditnote_id`, `comment_id`; required record fields `creditnote_id`,
  `comment_id`; accepted fields `comment_id`, `creditnote_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `update_credit_note_custom_fields`: POST `/creditnotes/{{ record.creditnote_id
  }}/customfields?organization_id={{ config.organization_id }}` - kind `create`; body type `json`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_credit_note_document`: PUT `/creditnotes/{{ record.creditnote_id }}/documents/{{
  record.document_id }}?organization_id={{ config.organization_id }}` - kind `update`; body type
  `none`; path fields `creditnote_id`, `document_id`; required record fields `creditnote_id`,
  `document_id`; accepted fields `creditnote_id`, `document_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `delete_credit_note_document`: DELETE `/creditnotes/{{ record.creditnote_id }}/documents/{{
  record.document_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `creditnote_id`, `document_id`; required record fields `creditnote_id`,
  `document_id`; accepted fields `creditnote_id`, `document_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `add_credit_note_digital_signature`: POST `/creditnotes/{{ record.creditnote_id
  }}/dsign?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `creditnote_id`; required record fields `creditnote_id`; accepted fields `creditnote_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `upload_credit_note_digital_signature`: POST `/creditnotes/{{ record.creditnote_id
  }}/dsign/upload?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `cancel_credit_note_einvoice`: POST `/creditnotes/{{ record.creditnote_id
  }}/einvoice/cancel?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `fetch_credit_note_einvoice`: POST `/creditnotes/{{ record.creditnote_id
  }}/einvoice/fetch?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `push_credit_note_einvoice`: POST `/creditnotes/{{ record.creditnote_id
  }}/einvoice/push?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_credit_note_einvoice_status`: DELETE `/creditnotes/{{ record.creditnote_id
  }}/einvoice/status?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: external destructive mutation in Zoho Books; approval required.
- `mark_credit_note_einvoice_cancelled`: POST `/creditnotes/{{ record.creditnote_id
  }}/einvoice/status/cancel?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_credit_note_einvoice_pushed`: POST `/creditnotes/{{ record.creditnote_id
  }}/einvoice/status/push?organization_id={{ config.organization_id }}` - kind `create`; body type
  `json`; path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `recall_credit_note_einvoice_status`: POST `/creditnotes/{{ record.creditnote_id
  }}/einvoice/status/recall?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `email_credit_note`: POST `/creditnotes/{{ record.creditnote_id }}/email?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `creditnote_id`;
  required record fields `creditnote_id`; accepted fields `creditnote_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `apply_credit_note_to_invoice`: POST `/creditnotes/{{ record.creditnote_id
  }}/invoices?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `creditnote_id`; required record fields `creditnote_id`; accepted fields `creditnote_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `delete_invoice_of_credit_note`: DELETE `/creditnotes/{{ record.creditnote_id }}/invoices/{{
  record.creditnote_invoice_id }}?organization_id={{ config.organization_id }}` - kind `delete`;
  body type `none`; path fields `creditnote_id`, `creditnote_invoice_id`; required record fields
  `creditnote_id`, `creditnote_invoice_id`; accepted fields `creditnote_id`,
  `creditnote_invoice_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `create_credit_note_refund`: POST `/creditnotes/{{ record.creditnote_id
  }}/refunds?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `creditnote_id`; required record fields `creditnote_id`; accepted fields `creditnote_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `push_credit_note_refund_einvoice`: POST `/creditnotes/{{ record.creditnote_id
  }}/refunds/einvoice/push?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_credit_note_refund`: PUT `/creditnotes/{{ record.creditnote_id }}/refunds/{{
  record.creditnote_refund_id }}?organization_id={{ config.organization_id }}` - kind `update`; body
  type `json`; path fields `creditnote_id`, `creditnote_refund_id`; required record fields
  `creditnote_id`, `creditnote_refund_id`; accepted fields `creditnote_id`, `creditnote_refund_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `delete_credit_note_refund`: DELETE `/creditnotes/{{ record.creditnote_id }}/refunds/{{
  record.creditnote_refund_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body
  type `none`; path fields `creditnote_id`, `creditnote_refund_id`; required record fields
  `creditnote_id`, `creditnote_refund_id`; accepted fields `creditnote_id`, `creditnote_refund_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `reject_credit_note`: POST `/creditnotes/{{ record.creditnote_id }}/reject?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `creditnote_id`;
  required record fields `creditnote_id`; accepted fields `creditnote_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `mark_credit_note_draft`: POST `/creditnotes/{{ record.creditnote_id
  }}/status/draft?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_credit_note_open`: POST `/creditnotes/{{ record.creditnote_id
  }}/status/open?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_credit_note_ready_to_push`: POST `/creditnotes/{{ record.creditnote_id
  }}/status/readytopush?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_credit_note_void`: POST `/creditnotes/{{ record.creditnote_id
  }}/status/void?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `creditnote_id`; required record fields `creditnote_id`; accepted fields
  `creditnote_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `submit_credit_note`: POST `/creditnotes/{{ record.creditnote_id }}/submit?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `creditnote_id`;
  required record fields `creditnote_id`; accepted fields `creditnote_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `apply_credit_note_substatus`: POST `/creditnotes/{{ record.creditnote_id }}/substatus/{{
  record.substatus_id }}?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `creditnote_id`, `substatus_id`; required record fields `creditnote_id`,
  `substatus_id`; accepted fields `creditnote_id`, `substatus_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `delete_credit_note_substatus`: DELETE `/creditnotes/{{ record.creditnote_id }}/substatus/{{
  record.substatus_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `creditnote_id`, `substatus_id`; required record fields `creditnote_id`,
  `substatus_id`; accepted fields `creditnote_id`, `substatus_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho
  Books; approval required.
- `update_credit_note_template`: PUT `/creditnotes/{{ record.creditnote_id }}/templates/{{
  record.template_id }}?organization_id={{ config.organization_id }}` - kind `update`; body type
  `none`; path fields `creditnote_id`, `template_id`; required record fields `creditnote_id`,
  `template_id`; accepted fields `creditnote_id`, `template_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `cancel_einvoice_credit_note`: POST `/einvoices/creditnotes/{{ record.creditnote_id
  }}/cancel?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `creditnote_id`; required record fields `creditnote_id`; accepted fields `creditnote_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `create_currency`: POST `/settings/currencies?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_currency`: PUT `/settings/currencies/{{ record.currency_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `currency_id`; required
  record fields `currency_id`; accepted fields `currency_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_currency`: DELETE `/settings/currencies/{{ record.currency_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `currency_id`; required
  record fields `currency_id`; accepted fields `currency_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `create_exchange_rate`: POST `/settings/currencies/{{ record.currency_id
  }}/exchangerates?organization_id={{ config.organization_id }}` - kind `create`; body type `json`;
  path fields `currency_id`; required record fields `currency_id`; accepted fields `currency_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `update_exchange_rate`: PUT `/settings/currencies/{{ record.currency_id }}/exchangerates/{{
  record.exchange_rate_id }}?organization_id={{ config.organization_id }}` - kind `update`; body
  type `json`; path fields `currency_id`, `exchange_rate_id`; required record fields `currency_id`,
  `exchange_rate_id`; accepted fields `currency_id`, `exchange_rate_id`; risk: external mutation in
  Zoho Books accounting data; approval required.
- `delete_exchange_rate`: DELETE `/settings/currencies/{{ record.currency_id }}/exchangerates/{{
  record.exchange_rate_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body
  type `none`; path fields `currency_id`, `exchange_rate_id`; required record fields `currency_id`,
  `exchange_rate_id`; accepted fields `currency_id`, `exchange_rate_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho
  Books; approval required.
- `create_custom_module`: POST `/settings/modules?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_custom_module`: PUT `/settings/modules/{{ record.module_api_name }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `module_api_name`;
  required record fields `module_api_name`; accepted fields `module_api_name`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `delete_custom_module`: DELETE `/settings/modules/{{ record.module_api_name }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `module_api_name`;
  required record fields `module_api_name`; accepted fields `module_api_name`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `create_custom_module_record`: POST `/{{ record.module_name }}?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `module_name`; required
  record fields `module_name`; accepted fields `module_name`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `bulk_update_custom_module_records`: PUT `/{{ record.module_name }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `module_name`; required
  record fields `module_name`; accepted fields `module_name`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_custom_module_records`: DELETE `/{{ record.module_name }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `module_name`; required
  record fields `module_name`; accepted fields `module_name`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `update_custom_module_record`: PUT `/{{ record.module_name }}/{{ record.module_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `module_name`, `module_id`; required record fields `module_name`, `module_id`; accepted fields
  `module_id`, `module_name`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_custom_module_record`: DELETE `/{{ record.module_name }}/{{ record.module_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `module_name`, `module_id`; required record fields `module_name`, `module_id`; accepted fields
  `module_id`, `module_name`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `create_customer_debit_note`: POST `/invoices?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_customer_debit_note`: PUT `/invoices/{{ record.debit_note_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `debit_note_id`;
  required record fields `debit_note_id`; accepted fields `debit_note_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `delete_customer_debit_note`: DELETE `/invoices/{{ record.debit_note_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `debit_note_id`;
  required record fields `debit_note_id`; accepted fields `debit_note_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: external destructive mutation in
  Zoho Books; approval required.
- `update_custom_fields_in_customer_payment`: PUT `/customerpayment/{{ record.customer_payment_id
  }}/customfields?organization_id={{ config.organization_id }}` - kind `update`; body type `json`;
  path fields `customer_payment_id`; required record fields `customer_payment_id`; accepted fields
  `customer_payment_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_customer_payment`: POST `/customerpayments?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `bulk_delete_customer_payments`: DELETE `/customerpayments?organization_id={{
  config.organization_id }}&payment_ids={{ record.payment_ids }}&bulk_delete={{ record.bulk_delete
  }}` - kind `delete`; body type `none`; path fields `payment_ids`, `bulk_delete`; required record
  fields `payment_ids`, `bulk_delete`; accepted fields `bulk_delete`, `payment_ids`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `create_customer_payment_refund`: POST `/customerpayments/{{ record.customer_payment_id
  }}/refunds?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `customer_payment_id`; required record fields `customer_payment_id`; accepted fields
  `customer_payment_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_customer_payment_refund`: PUT `/customerpayments/{{ record.customer_payment_id
  }}/refunds/{{ record.refund_id }}?organization_id={{ config.organization_id }}` - kind `update`;
  body type `json`; path fields `customer_payment_id`, `refund_id`; required record fields
  `customer_payment_id`, `refund_id`; accepted fields `customer_payment_id`, `refund_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `delete_customer_payment_refund`: DELETE `/customerpayments/{{ record.customer_payment_id
  }}/refunds/{{ record.refund_id }}?organization_id={{ config.organization_id }}` - kind `delete`;
  body type `none`; path fields `customer_payment_id`, `refund_id`; required record fields
  `customer_payment_id`, `refund_id`; accepted fields `customer_payment_id`, `refund_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `update_customer_payment`: PUT `/customerpayments/{{ record.payment_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `payment_id`; required
  record fields `payment_id`; accepted fields `payment_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_customer_payment`: DELETE `/customerpayments/{{ record.payment_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `payment_id`; required
  record fields `payment_id`; accepted fields `payment_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `create_delivery_challan`: POST `/deliverychallans?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `return_delivery_challans`: PUT `/deliverychallans/return?organization_id={{
  config.organization_id }}&deliverychallan_ids={{ record.deliverychallan_ids }}` - kind `update`;
  body type `json`; path fields `deliverychallan_ids`; required record fields `deliverychallan_ids`;
  accepted fields `deliverychallan_ids`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `undo_return_delivery_challans`: PUT `/deliverychallans/undo/return?organization_id={{
  config.organization_id }}&deliverychallan_ids={{ record.deliverychallan_ids }}` - kind `update`;
  body type `none`; path fields `deliverychallan_ids`; required record fields `deliverychallan_ids`;
  accepted fields `deliverychallan_ids`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `update_delivery_challan`: PUT `/deliverychallans/{{ record.deliverychallan_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `deliverychallan_id`; required record fields `deliverychallan_id`; accepted fields
  `deliverychallan_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_delivery_challan`: DELETE `/deliverychallans/{{ record.deliverychallan_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `deliverychallan_id`; required record fields `deliverychallan_id`; accepted fields
  `deliverychallan_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `update_delivery_challan_shipping_address`: PUT `/deliverychallans/{{ record.deliverychallan_id
  }}/address/shipping?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `deliverychallan_id`; required record fields `deliverychallan_id`; accepted
  fields `deliverychallan_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_delivery_challan_attachment`: DELETE `/deliverychallans/{{ record.deliverychallan_id
  }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}` - kind
  `delete`; body type `none`; path fields `deliverychallan_id`, `document_id`; required record
  fields `deliverychallan_id`, `document_id`; accepted fields `deliverychallan_id`, `document_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `mark_delivery_challan_as_delivered`: POST `/deliverychallans/{{ record.deliverychallan_id
  }}/status/delivered?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `deliverychallan_id`; required record fields `deliverychallan_id`; accepted
  fields `deliverychallan_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `mark_delivery_challan_as_open`: POST `/deliverychallans/{{ record.deliverychallan_id
  }}/status/open?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `deliverychallan_id`; required record fields `deliverychallan_id`; accepted fields
  `deliverychallan_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_delivery_challan_as_returned`: POST `/deliverychallans/{{ record.deliverychallan_id
  }}/status/returned?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `deliverychallan_id`; required record fields `deliverychallan_id`; accepted
  fields `deliverychallan_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `mark_delivery_challan_as_undelivered`: POST `/deliverychallans/{{ record.deliverychallan_id
  }}/status/undelivered?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `deliverychallan_id`; required record fields `deliverychallan_id`; accepted
  fields `deliverychallan_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_delivery_challan_template`: PUT `/deliverychallans/{{ record.deliverychallan_id
  }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}` - kind
  `update`; body type `none`; path fields `deliverychallan_id`, `template_id`; required record
  fields `deliverychallan_id`, `template_id`; accepted fields `deliverychallan_id`, `template_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `update_custom_fields_in_estimate`: PUT `/estimate/{{ record.estimate_id
  }}/customfields?organization_id={{ config.organization_id }}` - kind `update`; body type `json`;
  path fields `estimate_id`; required record fields `estimate_id`; accepted fields `estimate_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `create_estimate`: POST `/estimates?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `email_multiple_estimates`: POST `/estimates/email?organization_id={{ config.organization_id
  }}&estimate_ids={{ record.estimate_ids }}` - kind `create`; body type `none`; path fields
  `estimate_ids`; required record fields `estimate_ids`; accepted fields `estimate_ids`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `update_estimate`: PUT `/estimates/{{ record.estimate_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `estimate_id`; required
  record fields `estimate_id`; accepted fields `estimate_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_estimate`: DELETE `/estimates/{{ record.estimate_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `estimate_id`; required
  record fields `estimate_id`; accepted fields `estimate_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `update_estimate_billing_address`: PUT `/estimates/{{ record.estimate_id
  }}/address/billing?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `estimate_id`; required record fields `estimate_id`; accepted fields
  `estimate_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_estimate_shipping_address`: PUT `/estimates/{{ record.estimate_id
  }}/address/shipping?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `estimate_id`; required record fields `estimate_id`; accepted fields
  `estimate_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `approve_estimate`: POST `/estimates/{{ record.estimate_id }}/approve?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `estimate_id`; required
  record fields `estimate_id`; accepted fields `estimate_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `create_estimate_comment`: POST `/estimates/{{ record.estimate_id }}/comments?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `estimate_id`; required
  record fields `estimate_id`; accepted fields `estimate_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_estimate_comment`: PUT `/estimates/{{ record.estimate_id }}/comments/{{ record.comment_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `estimate_id`, `comment_id`; required record fields `estimate_id`, `comment_id`; accepted fields
  `comment_id`, `estimate_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_estimate_comment`: DELETE `/estimates/{{ record.estimate_id }}/comments/{{
  record.comment_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `estimate_id`, `comment_id`; required record fields `estimate_id`,
  `comment_id`; accepted fields `comment_id`, `estimate_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `email_estimate`: POST `/estimates/{{ record.estimate_id }}/email?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `estimate_id`; required
  record fields `estimate_id`; accepted fields `estimate_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_estimate_accepted`: POST `/estimates/{{ record.estimate_id
  }}/status/accepted?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `estimate_id`; required record fields `estimate_id`; accepted fields
  `estimate_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_estimate_declined`: POST `/estimates/{{ record.estimate_id
  }}/status/declined?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `estimate_id`; required record fields `estimate_id`; accepted fields
  `estimate_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_estimate_sent`: POST `/estimates/{{ record.estimate_id }}/status/sent?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `estimate_id`; required
  record fields `estimate_id`; accepted fields `estimate_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `submit_estimate`: POST `/estimates/{{ record.estimate_id }}/submit?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `estimate_id`; required
  record fields `estimate_id`; accepted fields `estimate_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_estimate_template`: PUT `/estimates/{{ record.estimate_id }}/templates/{{
  record.template_id }}?organization_id={{ config.organization_id }}` - kind `update`; body type
  `none`; path fields `estimate_id`, `template_id`; required record fields `estimate_id`,
  `template_id`; accepted fields `estimate_id`, `template_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_employee`: DELETE `/employee/{{ record.employee_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `employee_id`; required
  record fields `employee_id`; accepted fields `employee_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `create_employee`: POST `/employees?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_expense`: POST `/expenses?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_expense`: PUT `/expenses/{{ record.expense_id }}?organization_id={{ config.organization_id
  }}` - kind `update`; body type `json`; path fields `expense_id`; required record fields
  `expense_id`; accepted fields `expense_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_expense`: DELETE `/expenses/{{ record.expense_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `expense_id`; required
  record fields `expense_id`; accepted fields `expense_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `create_expense_receipt`: POST `/expenses/{{ record.expense_id }}/receipt?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `expense_id`; required
  record fields `expense_id`; accepted fields `expense_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_expense_receipt`: DELETE `/expenses/{{ record.expense_id }}/receipt?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `expense_id`; required
  record fields `expense_id`; accepted fields `expense_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `create_fixed_asset`: POST `/fixedassets?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_fixed_asset`: PUT `/fixedassets/{{ record.fixed_asset_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `fixed_asset_id`;
  required record fields `fixed_asset_id`; accepted fields `fixed_asset_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `delete_fixed_asset`: DELETE `/fixedassets/{{ record.fixed_asset_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `fixed_asset_id`;
  required record fields `fixed_asset_id`; accepted fields `fixed_asset_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: external destructive mutation in
  Zoho Books; approval required.
- `create_fixed_asset_comment`: POST `/fixedassets/{{ record.fixed_asset_id
  }}/comments?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `fixed_asset_id`; required record fields `fixed_asset_id`; accepted fields
  `fixed_asset_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_fixed_asset_comment`: DELETE `/fixedassets/{{ record.fixed_asset_id }}/comments/{{
  record.comment_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `fixed_asset_id`, `comment_id`; required record fields `fixed_asset_id`,
  `comment_id`; accepted fields `comment_id`, `fixed_asset_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `sell_fixed_asset`: POST `/fixedassets/{{ record.fixed_asset_id }}/sell?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `fixed_asset_id`;
  required record fields `fixed_asset_id`; accepted fields `fixed_asset_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `mark_fixed_asset_active`: POST `/fixedassets/{{ record.fixed_asset_id
  }}/status/active?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `fixed_asset_id`; required record fields `fixed_asset_id`; accepted fields
  `fixed_asset_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_fixed_asset_cancel`: POST `/fixedassets/{{ record.fixed_asset_id
  }}/status/cancel?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `fixed_asset_id`; required record fields `fixed_asset_id`; accepted fields
  `fixed_asset_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_fixed_asset_draft`: POST `/fixedassets/{{ record.fixed_asset_id
  }}/status/draft?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `fixed_asset_id`; required record fields `fixed_asset_id`; accepted fields
  `fixed_asset_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `write_off_fixed_asset`: POST `/fixedassets/{{ record.fixed_asset_id
  }}/writeoff?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `fixed_asset_id`; required record fields `fixed_asset_id`; accepted fields
  `fixed_asset_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_fixed_asset_type`: POST `/fixedassettypes?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_fixed_asset_type`: PUT `/fixedassettypes/{{ record.fixed_asset_type_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `fixed_asset_type_id`; required record fields `fixed_asset_type_id`; accepted fields
  `fixed_asset_type_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_fixed_asset_type`: DELETE `/fixedassettypes/{{ record.fixed_asset_type_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `fixed_asset_type_id`; required record fields `fixed_asset_type_id`; accepted fields
  `fixed_asset_type_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `import_customer_using_crm_account_id`: POST `/crm/account/{{ record.crm_account_id
  }}/import?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `crm_account_id`; required record fields `crm_account_id`; accepted fields
  `crm_account_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `import_customer_using_crm_contact_id`: POST `/crm/contact/{{ record.crm_contact_id
  }}/import?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `crm_contact_id`; required record fields `crm_contact_id`; accepted fields
  `crm_contact_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `import_item_using_crm_product_id`: POST `/crm/item/{{ record.crm_product_id
  }}/import?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `crm_product_id`; required record fields `crm_product_id`; accepted fields
  `crm_product_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `import_vendor_using_crm_vendor_id`: POST `/crm/vendor/{{ record.crm_vendor_id
  }}/import?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `crm_vendor_id`; required record fields `crm_vendor_id`; accepted fields `crm_vendor_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `cancel_einvoice_invoice`: POST `/einvoices/invoices/{{ record.invoice_id
  }}/cancel?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `update_custom_fields_in_invoice`: PUT `/invoice/{{ record.invoice_id
  }}/customfields?organization_id={{ config.organization_id }}` - kind `update`; body type `json`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `create_invoice`: POST `/invoices?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_invoices`: DELETE `/invoices?organization_id={{ config.organization_id }}&invoice_ids={{
  record.invoice_ids }}` - kind `delete`; body type `none`; path fields `invoice_ids`; required
  record fields `invoice_ids`; accepted fields `invoice_ids`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `approve_invoices`: POST `/invoices/approve?organization_id={{ config.organization_id
  }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `none`; path fields
  `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `preview_invoice_coupons`: POST `/invoices/coupons/preview?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `cancel_invoices_einvoice`: POST `/invoices/einvoice/cancel?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `push_invoices_einvoice`: POST `/invoices/einvoice/push?organization_id={{ config.organization_id
  }}` - kind `create`; body type `none`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `email_invoices`: POST `/invoices/email?organization_id={{ config.organization_id
  }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `json`; path fields
  `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `delete_invoice_expense_receipt`: DELETE `/invoices/expenses/{{ record.expense_id
  }}/receipt?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path
  fields `expense_id`; required record fields `expense_id`; accepted fields `expense_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `create_invoices_from_estimates`: POST `/invoices/fromestimates?organization_id={{
  config.organization_id }}&estimate_ids={{ record.estimate_ids }}` - kind `create`; body type
  `none`; path fields `estimate_ids`; required record fields `estimate_ids`; accepted fields
  `estimate_ids`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_invoices_from_projects`: POST `/invoices/fromprojects?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `create_invoice_from_salesorder`: POST `/invoices/fromsalesorder?organization_id={{
  config.organization_id }}&salesorder_id={{ record.salesorder_id }}` - kind `create`; body type
  `none`; path fields `salesorder_id`; required record fields `salesorder_id`; accepted fields
  `salesorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `map_invoice_with_salesorder`: POST `/invoices/mapwithorder?organization_id={{
  config.organization_id }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `none`;
  path fields `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `mark_invoices_shipped`: POST `/invoices/markasshipped?organization_id={{ config.organization_id
  }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `none`; path fields
  `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `bulk_invoice_reminder`: POST `/invoices/paymentreminder?organization_id={{ config.organization_id
  }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `none`; path fields
  `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `mark_invoices_sent`: POST `/invoices/status/sent?organization_id={{ config.organization_id
  }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `none`; path fields
  `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `void_invoices`: POST `/invoices/status/void?organization_id={{ config.organization_id
  }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `none`; path fields
  `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `submit_invoices`: POST `/invoices/submit?organization_id={{ config.organization_id
  }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `none`; path fields
  `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `unmap_invoices_from_salesorders`: PUT `/invoices/unmap/salesorders?organization_id={{
  config.organization_id }}` - kind `update`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `unship_invoices`: POST `/invoices/unship?organization_id={{ config.organization_id
  }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `none`; path fields
  `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `write_off_invoices`: POST `/invoices/writeoff?organization_id={{ config.organization_id
  }}&invoice_ids={{ record.invoice_ids }}` - kind `create`; body type `none`; path fields
  `invoice_ids`; required record fields `invoice_ids`; accepted fields `invoice_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `update_invoice`: PUT `/invoices/{{ record.invoice_id }}?organization_id={{ config.organization_id
  }}` - kind `update`; body type `json`; path fields `invoice_id`; required record fields
  `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_invoice`: DELETE `/invoices/{{ record.invoice_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `update_invoice_billing_address`: PUT `/invoices/{{ record.invoice_id
  }}/address/billing?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_invoice_shipping_address`: PUT `/invoices/{{ record.invoice_id
  }}/address/shipping?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_invoice_advanced_tracking_details`: PUT `/invoices/{{ record.invoice_id
  }}/advancedtrackingdetails?organization_id={{ config.organization_id }}` - kind `update`; body
  type `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `approve_invoice`: POST `/invoices/{{ record.invoice_id }}/approve?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `finalize_invoice_approval`: POST `/invoices/{{ record.invoice_id
  }}/approve/final?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `add_invoice_attachment`: POST `/invoices/{{ record.invoice_id }}/attachment?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_invoice_attachment_preference`: PUT `/invoices/{{ record.invoice_id
  }}/attachment?organization_id={{ config.organization_id }}&can_send_in_mail={{
  record.can_send_in_mail }}` - kind `update`; body type `none`; path fields `invoice_id`,
  `can_send_in_mail`; required record fields `invoice_id`, `can_send_in_mail`; accepted fields
  `can_send_in_mail`, `invoice_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_invoice_attachment`: DELETE `/invoices/{{ record.invoice_id
  }}/attachment?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `update_invoice_cfdi_status`: POST `/invoices/{{ record.invoice_id
  }}/cfdi/status?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `add_invoice_comment`: POST `/invoices/{{ record.invoice_id }}/comments?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_invoice_comment`: PUT `/invoices/{{ record.invoice_id }}/comments/{{ record.comment_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `invoice_id`, `comment_id`; required record fields `invoice_id`, `comment_id`; accepted fields
  `comment_id`, `invoice_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_invoice_comment`: DELETE `/invoices/{{ record.invoice_id }}/comments/{{ record.comment_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `invoice_id`, `comment_id`; required record fields `invoice_id`, `comment_id`; accepted fields
  `comment_id`, `invoice_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `apply_credits_to_invoice`: POST `/invoices/{{ record.invoice_id }}/credits?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_invoice_applied_credit`: DELETE `/invoices/{{ record.invoice_id }}/creditsapplied/{{
  record.creditnotes_invoice_id }}?organization_id={{ config.organization_id }}` - kind `delete`;
  body type `none`; path fields `invoice_id`, `creditnotes_invoice_id`; required record fields
  `invoice_id`, `creditnotes_invoice_id`; accepted fields `creditnotes_invoice_id`, `invoice_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `add_invoice_document`: POST `/invoices/{{ record.invoice_id }}/documents/{{ record.document_id
  }}?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path fields
  `invoice_id`, `document_id`; required record fields `invoice_id`, `document_id`; accepted fields
  `document_id`, `invoice_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_invoice_document`: DELETE `/invoices/{{ record.invoice_id }}/documents/{{
  record.document_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `invoice_id`, `document_id`; required record fields `invoice_id`,
  `document_id`; accepted fields `document_id`, `invoice_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `upload_invoice_document`: POST `/invoices/{{ record.invoice_id }}/documents/{{ record.document_id
  }}/upload?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `invoice_id`, `document_id`; required record fields `invoice_id`, `document_id`; accepted
  fields `document_id`, `invoice_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `add_invoice_digital_signature`: POST `/invoices/{{ record.invoice_id }}/dsign?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `upload_invoice_digital_signature`: POST `/invoices/{{ record.invoice_id
  }}/dsign/upload?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `cancel_invoice_einvoice`: POST `/invoices/{{ record.invoice_id
  }}/einvoice/cancel?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `fetch_invoice_einvoice`: POST `/invoices/{{ record.invoice_id
  }}/einvoice/fetch?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `update_invoice_einvoice_payment_status`: PUT `/invoices/{{ record.invoice_id
  }}/einvoice/paymentstatus?organization_id={{ config.organization_id }}` - kind `update`; body type
  `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `push_invoice_einvoice`: POST `/invoices/{{ record.invoice_id }}/einvoice/push?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_invoice_einvoice_status`: DELETE `/invoices/{{ record.invoice_id
  }}/einvoice/status?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: external destructive mutation in Zoho Books; approval required.
- `mark_invoice_einvoice_cancelled`: POST `/invoices/{{ record.invoice_id
  }}/einvoice/status/cancel?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_invoice_einvoice_pushed`: POST `/invoices/{{ record.invoice_id
  }}/einvoice/status/push?organization_id={{ config.organization_id }}` - kind `create`; body type
  `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `recall_invoice_einvoice_status`: POST `/invoices/{{ record.invoice_id
  }}/einvoice/status/recall?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `email_invoice`: POST `/invoices/{{ record.invoice_id }}/email?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `schedule_invoice_email`: POST `/invoices/{{ record.invoice_id
  }}/email/schedule?organization_id={{ config.organization_id }}` - kind `create`; body type `json`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `cancel_scheduled_invoice_email`: DELETE `/invoices/{{ record.invoice_id
  }}/email/schedule?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `force_pay_invoice`: POST `/invoices/{{ record.invoice_id }}/forcepay?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_invoice_line_item`: DELETE `/invoices/{{ record.invoice_id }}/lineitems/{{
  record.line_item_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `invoice_id`, `line_item_id`; required record fields `invoice_id`,
  `line_item_id`; accepted fields `invoice_id`, `line_item_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `mail_invoice_pdf`: POST `/invoices/{{ record.invoice_id }}/mailpdf?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_invoice_metadata`: PUT `/invoices/{{ record.invoice_id }}/metadata/{{ record.metadata_name
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `none`; path fields
  `invoice_id`, `metadata_name`; required record fields `invoice_id`, `metadata_name`; accepted
  fields `invoice_id`, `metadata_name`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `create_invoice_asynchronous_online_payment`: POST `/invoices/{{ record.invoice_id
  }}/onlinepayments/asynchronous?organization_id={{ config.organization_id }}` - kind `create`; body
  type `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `add_invoice_online_payment_bank_account`: POST `/invoices/{{ record.invoice_id
  }}/onlinepayments/bankaccount?organization_id={{ config.organization_id }}` - kind `create`; body
  type `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_invoice_synchronous_online_payment`: POST `/invoices/{{ record.invoice_id
  }}/onlinepayments/synchronous?organization_id={{ config.organization_id }}` - kind `create`; body
  type `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `remind_customer_for_invoice_payment`: POST `/invoices/{{ record.invoice_id
  }}/paymentreminder?organization_id={{ config.organization_id }}` - kind `create`; body type
  `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `disable_invoice_payment_reminder`: POST `/invoices/{{ record.invoice_id
  }}/paymentreminder/disable?organization_id={{ config.organization_id }}` - kind `create`; body
  type `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `enable_invoice_payment_reminder`: POST `/invoices/{{ record.invoice_id
  }}/paymentreminder/enable?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_invoice_payment`: DELETE `/invoices/{{ record.invoice_id }}/payments/{{
  record.invoice_payment_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body
  type `none`; path fields `invoice_id`, `invoice_payment_id`; required record fields `invoice_id`,
  `invoice_payment_id`; accepted fields `invoice_id`, `invoice_payment_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: external destructive mutation in
  Zoho Books; approval required.
- `apply_pricebook_to_invoice`: PUT `/invoices/{{ record.invoice_id }}/pricebooks/{{
  record.pricebook_id }}?organization_id={{ config.organization_id }}` - kind `update`; body type
  `none`; path fields `invoice_id`, `pricebook_id`; required record fields `invoice_id`,
  `pricebook_id`; accepted fields `invoice_id`, `pricebook_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `reject_invoice`: POST `/invoices/{{ record.invoice_id }}/reject?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `send_invoice_dunning_notifications`: POST `/invoices/{{ record.invoice_id
  }}/senddunningnotifications?organization_id={{ config.organization_id }}` - kind `create`; body
  type `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `send_invoice_retry_sms`: POST `/invoices/{{ record.invoice_id }}/sendretrysms?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `send_invoice_sms`: POST `/invoices/{{ record.invoice_id }}/sms?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `send_invoice_via_snail_mail`: POST `/invoices/{{ record.invoice_id
  }}/snailmail?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `cancel_invoice`: POST `/invoices/{{ record.invoice_id }}/status/cancel?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_invoice_draft`: POST `/invoices/{{ record.invoice_id }}/status/draft?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_invoice_ready_to_push`: POST `/invoices/{{ record.invoice_id
  }}/status/readytopush?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_invoice_sent`: POST `/invoices/{{ record.invoice_id }}/status/sent?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_invoice_void`: POST `/invoices/{{ record.invoice_id }}/status/void?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `submit_invoice`: POST `/invoices/{{ record.invoice_id }}/submit?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `apply_invoice_substatus`: POST `/invoices/{{ record.invoice_id }}/substatus/{{
  record.substatus_id }}?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `invoice_id`, `substatus_id`; required record fields `invoice_id`,
  `substatus_id`; accepted fields `invoice_id`, `substatus_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `delete_invoice_substatus`: DELETE `/invoices/{{ record.invoice_id }}/substatus/{{
  record.substatus_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `invoice_id`, `substatus_id`; required record fields `invoice_id`,
  `substatus_id`; accepted fields `invoice_id`, `substatus_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `update_invoice_template`: PUT `/invoices/{{ record.invoice_id }}/templates/{{ record.template_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `none`; path fields
  `invoice_id`, `template_id`; required record fields `invoice_id`, `template_id`; accepted fields
  `invoice_id`, `template_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `write_off_invoice`: POST `/invoices/{{ record.invoice_id }}/writeoff?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `invoice_id`; required
  record fields `invoice_id`; accepted fields `invoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `cancel_write_off_invoice`: POST `/invoices/{{ record.invoice_id
  }}/writeoff/cancel?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_custom_fields_in_item`: PUT `/item/{{ record.item_id }}/customfields?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `item_id`; required
  record fields `item_id`; accepted fields `item_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `create_item`: POST `/items?organization_id={{ config.organization_id }}` - kind `create`; body
  type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `add_items_to_portal`: POST `/items/addtoportal?organization_id={{ config.organization_id
  }}&item_ids={{ record.item_ids }}` - kind `create`; body type `none`; path fields `item_ids`;
  required record fields `item_ids`; accepted fields `item_ids`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `update_item`: PUT `/items/{{ record.item_id }}?organization_id={{ config.organization_id }}` -
  kind `update`; body type `json`; path fields `item_id`; required record fields `item_id`; accepted
  fields `item_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_item`: DELETE `/items/{{ record.item_id }}?organization_id={{ config.organization_id }}` -
  kind `delete`; body type `none`; path fields `item_id`; required record fields `item_id`; accepted
  fields `item_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: external destructive mutation in Zoho Books; approval required.
- `mark_item_active`: POST `/items/{{ record.item_id }}/active?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `item_id`; required
  record fields `item_id`; accepted fields `item_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_item_to_portal`: POST `/items/{{ record.item_id }}/addtoportal?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `item_id`; required
  record fields `item_id`; accepted fields `item_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_item_inactive`: POST `/items/{{ record.item_id }}/inactive?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `item_id`; required
  record fields `item_id`; accepted fields `item_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `remove_item_from_portal`: POST `/items/{{ record.item_id }}/removefromportal?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `item_id`; required
  record fields `item_id`; accepted fields `item_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `create_journal`: POST `/journals?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `bulk_approve_journals`: POST `/journals/approve?organization_id={{ config.organization_id
  }}&journal_ids={{ record.journal_ids }}` - kind `create`; body type `none`; path fields
  `journal_ids`; required record fields `journal_ids`; accepted fields `journal_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `bulk_delete_journals`: DELETE `/journals/bulkdelete?organization_id={{ config.organization_id
  }}&journal_ids={{ record.journal_ids }}` - kind `delete`; body type `none`; path fields
  `journal_ids`; required record fields `journal_ids`; accepted fields `journal_ids`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `bulk_publish_journals`: POST `/journals/status/publish?organization_id={{ config.organization_id
  }}&journal_ids={{ record.journal_ids }}` - kind `create`; body type `none`; path fields
  `journal_ids`; required record fields `journal_ids`; accepted fields `journal_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `bulk_submit_journals`: POST `/journals/submit?organization_id={{ config.organization_id
  }}&journal_ids={{ record.journal_ids }}` - kind `create`; body type `none`; path fields
  `journal_ids`; required record fields `journal_ids`; accepted fields `journal_ids`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `update_journal`: PUT `/journals/{{ record.journal_id }}?organization_id={{ config.organization_id
  }}` - kind `update`; body type `json`; path fields `journal_id`; required record fields
  `journal_id`; accepted fields `journal_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_journal`: DELETE `/journals/{{ record.journal_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `journal_id`; required
  record fields `journal_id`; accepted fields `journal_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `approve_journal`: POST `/journals/{{ record.journal_id }}/approve?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `journal_id`; required
  record fields `journal_id`; accepted fields `journal_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_journal_attachment`: POST `/journals/{{ record.journal_id }}/attachment?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `journal_id`; required
  record fields `journal_id`; accepted fields `journal_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_journal_comment`: POST `/journals/{{ record.journal_id }}/comments?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `journal_id`; required
  record fields `journal_id`; accepted fields `journal_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_journal_comment`: DELETE `/journals/{{ record.journal_id }}/comments/{{ record.comment_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `journal_id`, `comment_id`; required record fields `journal_id`, `comment_id`; accepted fields
  `comment_id`, `journal_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `apply_journal_credits_to_bills`: POST `/journals/{{ record.journal_id }}/credits/{{
  record.journal_line_id }}/bills?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; path fields `journal_id`, `journal_line_id`; required record fields
  `journal_id`, `journal_line_id`; accepted fields `journal_id`, `journal_line_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `apply_journal_credits_to_invoices`: POST `/journals/{{ record.journal_id }}/credits/{{
  record.journal_line_id }}/invoices?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; path fields `journal_id`, `journal_line_id`; required record fields
  `journal_id`, `journal_line_id`; accepted fields `journal_id`, `journal_line_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `delete_journal_credits_payables`: DELETE `/journals/{{ record.journal_id }}/credits/{{
  record.journal_line_id }}/payables?organization_id={{ config.organization_id }}` - kind `delete`;
  body type `none`; path fields `journal_id`, `journal_line_id`; required record fields
  `journal_id`, `journal_line_id`; accepted fields `journal_id`, `journal_line_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `delete_journal_credits_receivables`: DELETE `/journals/{{ record.journal_id }}/credits/{{
  record.journal_line_id }}/receivables?organization_id={{ config.organization_id }}` - kind
  `delete`; body type `none`; path fields `journal_id`, `journal_line_id`; required record fields
  `journal_id`, `journal_line_id`; accepted fields `journal_id`, `journal_line_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `reject_journal`: POST `/journals/{{ record.journal_id }}/reject?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `journal_id`; required
  record fields `journal_id`; accepted fields `journal_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `reverse_journal`: POST `/journals/{{ record.journal_id }}/reverse?organization_id={{
  config.organization_id }}&reversal_date={{ record.reversal_date }}&is_reversal_scheduled={{
  record.is_reversal_scheduled }}` - kind `create`; body type `none`; path fields `journal_id`,
  `reversal_date`, `is_reversal_scheduled`; required record fields `journal_id`, `reversal_date`,
  `is_reversal_scheduled`; accepted fields `is_reversal_scheduled`, `journal_id`, `reversal_date`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `mark_journal_published`: POST `/journals/{{ record.journal_id
  }}/status/publish?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `journal_id`; required record fields `journal_id`; accepted fields `journal_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `submit_journal_for_approval`: POST `/journals/{{ record.journal_id }}/submit?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `journal_id`; required
  record fields `journal_id`; accepted fields `journal_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `create_recurring_journal`: POST `/recurringjournals?organization_id={{ config.organization_id }}`
  - kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_recurring_journal`: PUT `/recurringjournals/{{ record.recurring_journal_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `recurring_journal_id`; required record fields `recurring_journal_id`; accepted fields
  `recurring_journal_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_recurring_journal`: DELETE `/recurringjournals/{{ record.recurring_journal_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `recurring_journal_id`; required record fields `recurring_journal_id`; accepted fields
  `recurring_journal_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `resume_recurring_journal`: POST `/recurringjournals/{{ record.recurring_journal_id
  }}/status/resume?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `recurring_journal_id`; required record fields `recurring_journal_id`; accepted fields
  `recurring_journal_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `stop_recurring_journal`: POST `/recurringjournals/{{ record.recurring_journal_id
  }}/status/stop?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `recurring_journal_id`; required record fields `recurring_journal_id`; accepted fields
  `recurring_journal_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_location`: POST `/locations?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_location`: PUT `/locations/{{ record.location_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `location_id`; required
  record fields `location_id`; accepted fields `location_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_location`: DELETE `/locations/{{ record.location_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `location_id`; required
  record fields `location_id`; accepted fields `location_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `mark_location_active`: POST `/locations/{{ record.location_id }}/active?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `location_id`; required
  record fields `location_id`; accepted fields `location_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_location_inactive`: POST `/locations/{{ record.location_id }}/inactive?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `location_id`; required
  record fields `location_id`; accepted fields `location_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_location_primary`: POST `/locations/{{ record.location_id
  }}/markasprimary?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `location_id`; required record fields `location_id`; accepted fields `location_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `enable_locations`: POST `/settings/locations/enable?organization_id={{ config.organization_id }}`
  - kind `create`; body type `none`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `writeoff_opening_balance`: POST `/openingbalances/{{ record.opening_balance_id
  }}/writeoff?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `opening_balance_id`; required record fields `opening_balance_id`; accepted fields
  `opening_balance_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `cancel_writeoff_opening_balance`: POST `/openingbalances/{{ record.opening_balance_id
  }}/writeoff/cancel?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `opening_balance_id`; required record fields `opening_balance_id`; accepted
  fields `opening_balance_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `create_opening_balance`: POST `/settings/openingbalances?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `update_opening_balance`: PUT `/settings/openingbalances?organization_id={{ config.organization_id
  }}` - kind `update`; body type `json`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_opening_balance`: DELETE `/settings/openingbalances?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `create_organization`: POST `/organizations` - kind `create`; body type `json`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `create_organization_address`: POST `/organizations/address` - kind `create`; body type `json`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `update_organization_address`: PUT `/organizations/address/{{ record.address_id }}` - kind
  `update`; body type `json`; path fields `address_id`; required record fields `address_id`;
  accepted fields `address_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_organization_address`: DELETE `/organizations/address/{{ record.address_id }}` - kind
  `delete`; body type `none`; path fields `address_id`; required record fields `address_id`;
  accepted fields `address_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `update_organization`: PUT `/organizations/{{ record.organization_id }}` - kind `update`; body
  type `json`; path fields `organization_id`; required record fields `organization_id`; accepted
  fields `organization_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `copy_organization_settings`: POST `/organizations/{{ record.organization_id
  }}/copysettings?settings_to_copy={{ record.settings_to_copy }}` - kind `create`; body type `none`;
  path fields `organization_id`, `settings_to_copy`; required record fields `organization_id`,
  `settings_to_copy`; accepted fields `organization_id`, `settings_to_copy`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `downgrade_organization_to_invoice`: POST `/organizations/{{ record.organization_id
  }}/downgradetoinvoice` - kind `create`; body type `none`; path fields `organization_id`; required
  record fields `organization_id`; accepted fields `organization_id`; risk: external mutation in
  Zoho Books accounting data; approval required.
- `mark_organization_inactive`: POST `/organizations/{{ record.organization_id }}/inactive` - kind
  `create`; body type `none`; path fields `organization_id`; required record fields
  `organization_id`; accepted fields `organization_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `upgrade_organization_to_books`: POST `/organizations/{{ record.organization_id }}/upgradetobooks`
  - kind `create`; body type `none`; path fields `organization_id`; required record fields
  `organization_id`; accepted fields `organization_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `create_pricebook`: POST `/pricebooks?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_pricebook`: PUT `/pricebooks/{{ record.pricebook_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `pricebook_id`; required
  record fields `pricebook_id`; accepted fields `pricebook_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `delete_pricebook`: DELETE `/pricebooks/{{ record.pricebook_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `pricebook_id`; required
  record fields `pricebook_id`; accepted fields `pricebook_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `mark_pricebook_active`: POST `/pricebooks/{{ record.pricebook_id }}/active?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `pricebook_id`; required
  record fields `pricebook_id`; accepted fields `pricebook_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `mark_pricebook_inactive`: POST `/pricebooks/{{ record.pricebook_id }}/inactive?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `pricebook_id`; required
  record fields `pricebook_id`; accepted fields `pricebook_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `create_project`: POST `/projects?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_project`: PUT `/projects/{{ record.project_id }}?organization_id={{ config.organization_id
  }}` - kind `update`; body type `json`; path fields `project_id`; required record fields
  `project_id`; accepted fields `project_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_project`: DELETE `/projects/{{ record.project_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `project_id`; required
  record fields `project_id`; accepted fields `project_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `mark_project_active`: POST `/projects/{{ record.project_id }}/active?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `project_id`; required
  record fields `project_id`; accepted fields `project_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `clone_project`: POST `/projects/{{ record.project_id }}/clone?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `project_id`; required
  record fields `project_id`; accepted fields `project_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_project_comment`: POST `/projects/{{ record.project_id }}/comments?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `project_id`; required
  record fields `project_id`; accepted fields `project_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_project_comment`: DELETE `/projects/{{ record.project_id }}/comments/{{ record.comment_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `project_id`, `comment_id`; required record fields `project_id`, `comment_id`; accepted fields
  `comment_id`, `project_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `mark_project_inactive`: POST `/projects/{{ record.project_id }}/inactive?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `project_id`; required
  record fields `project_id`; accepted fields `project_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_project_task`: POST `/projects/{{ record.project_id }}/tasks?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `project_id`; required
  record fields `project_id`; accepted fields `project_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_project_task`: PUT `/projects/{{ record.project_id }}/tasks/{{ record.task_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `project_id`, `task_id`; required record fields `project_id`, `task_id`; accepted fields
  `project_id`, `task_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_project_task`: DELETE `/projects/{{ record.project_id }}/tasks/{{ record.task_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `project_id`, `task_id`; required record fields `project_id`, `task_id`; accepted fields
  `project_id`, `task_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `add_project_user`: POST `/projects/{{ record.project_id }}/users?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `project_id`; required
  record fields `project_id`; accepted fields `project_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `invite_project_user`: POST `/projects/{{ record.project_id }}/users/invite?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `project_id`; required
  record fields `project_id`; accepted fields `project_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_project_user`: PUT `/projects/{{ record.project_id }}/users/{{ record.user_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `project_id`, `user_id`; required record fields `project_id`, `user_id`; accepted fields
  `project_id`, `user_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_project_user`: DELETE `/projects/{{ record.project_id }}/users/{{ record.user_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `project_id`, `user_id`; required record fields `project_id`, `user_id`; accepted fields
  `project_id`, `user_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `update_custom_fields_in_purchase_order`: PUT `/purchaseorder/{{ record.purchaseorder_id
  }}/customfields?organization_id={{ config.organization_id }}` - kind `update`; body type `json`;
  path fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_purchase_order`: POST `/purchaseorders?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_purchase_order`: PUT `/purchaseorders/{{ record.purchaseorder_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `purchaseorder_id`;
  required record fields `purchaseorder_id`; accepted fields `purchaseorder_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `delete_purchase_order`: DELETE `/purchaseorders/{{ record.purchaseorder_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `purchaseorder_id`;
  required record fields `purchaseorder_id`; accepted fields `purchaseorder_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `update_purchase_order_billing_address`: PUT `/purchaseorders/{{ record.purchaseorder_id
  }}/address/billing?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `approve_purchase_order`: POST `/purchaseorders/{{ record.purchaseorder_id
  }}/approve?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `add_purchase_order_attachment`: POST `/purchaseorders/{{ record.purchaseorder_id
  }}/attachment?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_purchase_order_attachment`: PUT `/purchaseorders/{{ record.purchaseorder_id
  }}/attachment?organization_id={{ config.organization_id }}&can_send_in_mail={{
  record.can_send_in_mail }}` - kind `update`; body type `none`; path fields `purchaseorder_id`,
  `can_send_in_mail`; required record fields `purchaseorder_id`, `can_send_in_mail`; accepted fields
  `can_send_in_mail`, `purchaseorder_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_purchase_order_attachment`: DELETE `/purchaseorders/{{ record.purchaseorder_id
  }}/attachment?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`;
  path fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `add_purchase_order_comment`: POST `/purchaseorders/{{ record.purchaseorder_id
  }}/comments?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_purchase_order_comment`: PUT `/purchaseorders/{{ record.purchaseorder_id }}/comments/{{
  record.comment_id }}?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `purchaseorder_id`, `comment_id`; required record fields `purchaseorder_id`,
  `comment_id`; accepted fields `comment_id`, `purchaseorder_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `delete_purchase_order_comment`: DELETE `/purchaseorders/{{ record.purchaseorder_id }}/comments/{{
  record.comment_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `purchaseorder_id`, `comment_id`; required record fields `purchaseorder_id`,
  `comment_id`; accepted fields `comment_id`, `purchaseorder_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `email_purchase_order`: POST `/purchaseorders/{{ record.purchaseorder_id
  }}/email?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `reject_purchase_orders`: POST `/purchaseorders/{{ record.purchaseorder_id
  }}/reject?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_purchase_order_billed`: POST `/purchaseorders/{{ record.purchaseorder_id
  }}/status/billed?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_purchase_order_cancelled`: POST `/purchaseorders/{{ record.purchaseorder_id
  }}/status/cancelled?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_purchase_order_open`: POST `/purchaseorders/{{ record.purchaseorder_id
  }}/status/open?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `submit_purchase_order`: POST `/purchaseorders/{{ record.purchaseorder_id
  }}/submit?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `purchaseorder_id`; required record fields `purchaseorder_id`; accepted fields
  `purchaseorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_purchase_order_template`: PUT `/purchaseorders/{{ record.purchaseorder_id }}/templates/{{
  record.template_id }}?organization_id={{ config.organization_id }}` - kind `update`; body type
  `none`; path fields `purchaseorder_id`, `template_id`; required record fields `purchaseorder_id`,
  `template_id`; accepted fields `purchaseorder_id`, `template_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `delete_recurring_bill`: DELETE `/recurring_bills/{{ record.recurring_bill_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `recurring_bill_id`; required record fields `recurring_bill_id`; accepted fields
  `recurring_bill_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `create_recurring_bill`: POST `/recurringbills?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_recurring_bill`: PUT `/recurringbills/{{ record.recurring_bill_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `recurring_bill_id`;
  required record fields `recurring_bill_id`; accepted fields `recurring_bill_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `resume_recurring_bill`: POST `/recurringbills/{{ record.recurring_bill_id
  }}/status/resume?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `recurring_bill_id`; required record fields `recurring_bill_id`; accepted fields
  `recurring_bill_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `stop_recurring_bill`: POST `/recurringbills/{{ record.recurring_bill_id
  }}/status/stop?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `recurring_bill_id`; required record fields `recurring_bill_id`; accepted fields
  `recurring_bill_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_recurring_expense`: POST `/recurringexpenses?organization_id={{ config.organization_id }}`
  - kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_recurring_expense`: PUT `/recurringexpenses/{{ record.recurring_expense_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `recurring_expense_id`; required record fields `recurring_expense_id`; accepted fields
  `recurring_expense_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_recurring_expense`: DELETE `/recurringexpenses/{{ record.recurring_expense_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `recurring_expense_id`; required record fields `recurring_expense_id`; accepted fields
  `recurring_expense_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `resume_recurring_expense`: POST `/recurringexpenses/{{ record.recurring_expense_id
  }}/status/resume?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `recurring_expense_id`; required record fields `recurring_expense_id`; accepted fields
  `recurring_expense_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `stop_recurring_expense`: POST `/recurringexpenses/{{ record.recurring_expense_id
  }}/status/stop?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `recurring_expense_id`; required record fields `recurring_expense_id`; accepted fields
  `recurring_expense_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_recurring_invoice`: POST `/recurringinvoices?organization_id={{ config.organization_id }}`
  - kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `resume_recurring_invoices`: POST `/recurringinvoices/status/resume?organization_id={{
  config.organization_id }}&recurring_invoice_ids={{ record.recurring_invoice_ids }}` - kind
  `create`; body type `none`; path fields `recurring_invoice_ids`; required record fields
  `recurring_invoice_ids`; accepted fields `recurring_invoice_ids`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `stop_recurring_invoices`: POST `/recurringinvoices/status/stop?organization_id={{
  config.organization_id }}&recurring_invoice_ids={{ record.recurring_invoice_ids }}` - kind
  `create`; body type `none`; path fields `recurring_invoice_ids`; required record fields
  `recurring_invoice_ids`; accepted fields `recurring_invoice_ids`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `update_recurring_invoice`: PUT `/recurringinvoices/{{ record.recurring_invoice_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `recurring_invoice_id`; required record fields `recurring_invoice_id`; accepted fields
  `recurring_invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_recurring_invoice`: DELETE `/recurringinvoices/{{ record.recurring_invoice_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `recurring_invoice_id`; required record fields `recurring_invoice_id`; accepted fields
  `recurring_invoice_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `disable_recurring_invoice_autobill`: POST `/recurringinvoices/{{ record.recurring_invoice_id
  }}/autobill/disable?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `recurring_invoice_id`; required record fields `recurring_invoice_id`;
  accepted fields `recurring_invoice_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `enable_recurring_invoice_autobill`: POST `/recurringinvoices/{{ record.recurring_invoice_id
  }}/autobill/enable?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `recurring_invoice_id`; required record fields `recurring_invoice_id`;
  accepted fields `recurring_invoice_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_recurring_invoice_bank_account`: DELETE `/recurringinvoices/{{ record.recurring_invoice_id
  }}/bankaccount?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`;
  path fields `recurring_invoice_id`; required record fields `recurring_invoice_id`; accepted fields
  `recurring_invoice_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `associate_recurring_invoice_bank_account`: POST `/recurringinvoices/{{
  record.recurring_invoice_id }}/bankaccount/{{ record.account_id }}?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `recurring_invoice_id`,
  `account_id`; required record fields `recurring_invoice_id`, `account_id`; accepted fields
  `account_id`, `recurring_invoice_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_recurring_invoice_card`: DELETE `/recurringinvoices/{{ record.recurring_invoice_id
  }}/card?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path
  fields `recurring_invoice_id`; required record fields `recurring_invoice_id`; accepted fields
  `recurring_invoice_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `associate_recurring_invoice_card`: POST `/recurringinvoices/{{ record.recurring_invoice_id
  }}/card/{{ record.card_id }}?organization_id={{ config.organization_id }}` - kind `create`; body
  type `none`; path fields `recurring_invoice_id`, `card_id`; required record fields
  `recurring_invoice_id`, `card_id`; accepted fields `card_id`, `recurring_invoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `resume_recurring_invoice`: POST `/recurringinvoices/{{ record.recurring_invoice_id
  }}/status/resume?organization_id={{ config.organization_id }}` - kind `create`; body type `json`;
  path fields `recurring_invoice_id`; required record fields `recurring_invoice_id`; accepted fields
  `recurring_invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `stop_recurring_invoice`: POST `/recurringinvoices/{{ record.recurring_invoice_id
  }}/status/stop?organization_id={{ config.organization_id }}` - kind `create`; body type `json`;
  path fields `recurring_invoice_id`; required record fields `recurring_invoice_id`; accepted fields
  `recurring_invoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_recurring_invoice_template`: PUT `/recurringinvoices/{{ record.recurring_invoice_id
  }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}` - kind
  `update`; body type `json`; path fields `recurring_invoice_id`, `template_id`; required record
  fields `recurring_invoice_id`, `template_id`; accepted fields `recurring_invoice_id`,
  `template_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `bulk_delete_register_transactions`: PUT `/registers/{{ record.account_id
  }}/transactions/bulkdelete?organization_id={{ config.organization_id }}` - kind `update`; body
  type `json`; path fields `account_id`; required record fields `account_id`; accepted fields
  `account_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `bulk_update_register_transactions`: PUT `/registers/{{ record.account_id
  }}/transactions/bulkupdate?organization_id={{ config.organization_id }}` - kind `update`; body
  type `json`; path fields `account_id`; required record fields `account_id`; accepted fields
  `account_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_tag`: POST `/reportingtags?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `reorder_tags`: PUT `/reportingtags/reorder?organization_id={{ config.organization_id }}` - kind
  `update`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `mark_default_option`: POST `/reportingtags/{{ record.tag_id }}?organization_id={{
  config.organization_id }}&default_option_id={{ record.default_option_id }}` - kind `create`; body
  type `none`; path fields `tag_id`, `default_option_id`; required record fields `tag_id`,
  `default_option_id`; accepted fields `default_option_id`, `tag_id`; risk: external mutation in
  Zoho Books accounting data; approval required.
- `update_tag`: PUT `/reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id
  }}` - kind `update`; body type `json`; path fields `tag_id`; required record fields `tag_id`;
  accepted fields `tag_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_tag`: DELETE `/reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id
  }}` - kind `delete`; body type `none`; path fields `tag_id`; required record fields `tag_id`;
  accepted fields `tag_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `active_tag`: POST `/reportingtags/{{ record.tag_id }}/active?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `tag_id`; required
  record fields `tag_id`; accepted fields `tag_id`; risk: external mutation in Zoho Books accounting
  data; approval required.
- `update_tag_criteria`: PUT `/reportingtags/{{ record.tag_id }}/criteria?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `tag_id`; required
  record fields `tag_id`; accepted fields `tag_id`; risk: external mutation in Zoho Books accounting
  data; approval required.
- `inactive_tag`: POST `/reportingtags/{{ record.tag_id }}/inactive?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `tag_id`; required
  record fields `tag_id`; accepted fields `tag_id`; risk: external mutation in Zoho Books accounting
  data; approval required.
- `active_tag_option`: POST `/reportingtags/{{ record.tag_id
  }}/option/(\d+)/active?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `tag_id`; required record fields `tag_id`; accepted fields `tag_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `inactive_tag_option`: POST `/reportingtags/{{ record.tag_id
  }}/option/(\d+)/inactive?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `tag_id`; required record fields `tag_id`; accepted fields `tag_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `update_tag_options`: PUT `/reportingtags/{{ record.tag_id }}/options?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `tag_id`; required
  record fields `tag_id`; accepted fields `tag_id`; risk: external mutation in Zoho Books accounting
  data; approval required.
- `create_retainer_invoice`: POST `/retainerinvoices?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_retainer_invoice`: PUT `/retainerinvoices/{{ record.retainerinvoice_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_retainer_invoice`: DELETE `/retainerinvoices/{{ record.retainerinvoice_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `update_retainer_invoice_billing_address`: PUT `/retainerinvoices/{{ record.retainerinvoice_id
  }}/address/billing?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted
  fields `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `approve_retainer_invoice`: POST `/retainerinvoices/{{ record.retainerinvoice_id
  }}/approve?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `add_retainer_invoice_attachment`: POST `/retainerinvoices/{{ record.retainerinvoice_id
  }}/attachment?organization_id={{ config.organization_id }}` - kind `create`; body type `json`;
  path fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `add_retainer_invoice_comment`: POST `/retainerinvoices/{{ record.retainerinvoice_id
  }}/comments?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_retainer_invoice_comment`: PUT `/retainerinvoices/{{ record.retainerinvoice_id
  }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}` - kind `update`;
  body type `json`; path fields `retainerinvoice_id`, `comment_id`; required record fields
  `retainerinvoice_id`, `comment_id`; accepted fields `comment_id`, `retainerinvoice_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `delete_retainer_invoice_comment`: DELETE `/retainerinvoices/{{ record.retainerinvoice_id
  }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}` - kind `delete`;
  body type `none`; path fields `retainerinvoice_id`, `comment_id`; required record fields
  `retainerinvoice_id`, `comment_id`; accepted fields `comment_id`, `retainerinvoice_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `delete_retainer_invoice_attachment`: DELETE `/retainerinvoices/{{ record.retainerinvoice_id
  }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}` - kind
  `delete`; body type `none`; path fields `retainerinvoice_id`, `document_id`; required record
  fields `retainerinvoice_id`, `document_id`; accepted fields `document_id`, `retainerinvoice_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `email_retainer_invoice`: POST `/retainerinvoices/{{ record.retainerinvoice_id
  }}/email?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `apply_retainer_payments_to_invoices`: POST `/retainerinvoices/{{ record.retainerinvoice_id
  }}/invoices?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_applied_retainer_payment`: DELETE `/retainerinvoices/{{ record.retainerinvoice_id
  }}/invoices/{{ record.invoice_id }}?organization_id={{ config.organization_id }}` - kind `delete`;
  body type `none`; path fields `retainerinvoice_id`, `invoice_id`; required record fields
  `retainerinvoice_id`, `invoice_id`; accepted fields `invoice_id`, `retainerinvoice_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `create_retainer_invoice_async_online_payment`: POST `/retainerinvoices/{{
  record.retainerinvoice_id }}/onlinepayments/asynchronous?organization_id={{ config.organization_id
  }}` - kind `create`; body type `none`; path fields `retainerinvoice_id`; required record fields
  `retainerinvoice_id`; accepted fields `retainerinvoice_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_retainer_invoice_draft`: POST `/retainerinvoices/{{ record.retainerinvoice_id
  }}/status/draft?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_retainer_invoice_sent`: POST `/retainerinvoices/{{ record.retainerinvoice_id
  }}/status/sent?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_retainer_invoice_void`: POST `/retainerinvoices/{{ record.retainerinvoice_id
  }}/status/void?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `submit_retainer_invoice`: POST `/retainerinvoices/{{ record.retainerinvoice_id
  }}/submit?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `retainerinvoice_id`; required record fields `retainerinvoice_id`; accepted fields
  `retainerinvoice_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_retainer_invoice_template`: PUT `/retainerinvoices/{{ record.retainerinvoice_id
  }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}` - kind
  `update`; body type `none`; path fields `retainerinvoice_id`, `template_id`; required record
  fields `retainerinvoice_id`, `template_id`; accepted fields `retainerinvoice_id`, `template_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `update_salesorder_customfields`: PUT `/salesorder/{{ record.salesorder_id
  }}/customfields?organization_id={{ config.organization_id }}` - kind `update`; body type `json`;
  path fields `salesorder_id`; required record fields `salesorder_id`; accepted fields
  `salesorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_sales_order`: POST `/salesorders?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_sales_order`: PUT `/salesorders/{{ record.salesorder_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `salesorder_id`;
  required record fields `salesorder_id`; accepted fields `salesorder_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `delete_sales_order`: DELETE `/salesorders/{{ record.salesorder_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `salesorder_id`;
  required record fields `salesorder_id`; accepted fields `salesorder_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: external destructive mutation in
  Zoho Books; approval required.
- `update_sales_order_billing_address`: PUT `/salesorders/{{ record.salesorder_id
  }}/address/billing?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `salesorder_id`; required record fields `salesorder_id`; accepted fields
  `salesorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_sales_order_shipping_address`: PUT `/salesorders/{{ record.salesorder_id
  }}/address/shipping?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `salesorder_id`; required record fields `salesorder_id`; accepted fields
  `salesorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `approve_sales_order`: POST `/salesorders/{{ record.salesorder_id }}/approve?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `salesorder_id`;
  required record fields `salesorder_id`; accepted fields `salesorder_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `add_sales_order_attachment`: POST `/salesorders/{{ record.salesorder_id
  }}/attachment?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `salesorder_id`; required record fields `salesorder_id`; accepted fields
  `salesorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_sales_order_attachment_preference`: PUT `/salesorders/{{ record.salesorder_id
  }}/attachment?organization_id={{ config.organization_id }}&can_send_in_mail={{
  record.can_send_in_mail }}` - kind `update`; body type `none`; path fields `salesorder_id`,
  `can_send_in_mail`; required record fields `salesorder_id`, `can_send_in_mail`; accepted fields
  `can_send_in_mail`, `salesorder_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `delete_sales_order_attachment`: DELETE `/salesorders/{{ record.salesorder_id
  }}/attachment?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`;
  path fields `salesorder_id`; required record fields `salesorder_id`; accepted fields
  `salesorder_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: external destructive mutation in Zoho Books; approval required.
- `add_sales_order_comment`: POST `/salesorders/{{ record.salesorder_id
  }}/comments?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `salesorder_id`; required record fields `salesorder_id`; accepted fields `salesorder_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `update_sales_order_comment`: PUT `/salesorders/{{ record.salesorder_id }}/comments/{{
  record.comment_id }}?organization_id={{ config.organization_id }}` - kind `update`; body type
  `json`; path fields `salesorder_id`, `comment_id`; required record fields `salesorder_id`,
  `comment_id`; accepted fields `comment_id`, `salesorder_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_sales_order_comment`: DELETE `/salesorders/{{ record.salesorder_id }}/comments/{{
  record.comment_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `salesorder_id`, `comment_id`; required record fields `salesorder_id`,
  `comment_id`; accepted fields `comment_id`, `salesorder_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `email_sales_order`: POST `/salesorders/{{ record.salesorder_id }}/email?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `salesorder_id`;
  required record fields `salesorder_id`; accepted fields `salesorder_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `mark_sales_order_as_open`: POST `/salesorders/{{ record.salesorder_id
  }}/status/open?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `salesorder_id`; required record fields `salesorder_id`; accepted fields
  `salesorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_sales_order_as_void`: POST `/salesorders/{{ record.salesorder_id
  }}/status/void?organization_id={{ config.organization_id }}` - kind `create`; body type `json`;
  path fields `salesorder_id`; required record fields `salesorder_id`; accepted fields
  `salesorder_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `submit_sales_order`: POST `/salesorders/{{ record.salesorder_id }}/submit?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `salesorder_id`;
  required record fields `salesorder_id`; accepted fields `salesorder_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `update_sales_order_sub_status`: POST `/salesorders/{{ record.salesorder_id }}/substatus/{{
  record.status_code }}?organization_id={{ config.organization_id }}` - kind `create`; body type
  `none`; path fields `salesorder_id`, `status_code`; required record fields `salesorder_id`,
  `status_code`; accepted fields `salesorder_id`, `status_code`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `update_sales_order_template`: PUT `/salesorders/{{ record.salesorder_id }}/templates/{{
  record.template_id }}?organization_id={{ config.organization_id }}` - kind `update`; body type
  `none`; path fields `salesorder_id`, `template_id`; required record fields `salesorder_id`,
  `template_id`; accepted fields `salesorder_id`, `template_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `create_sales_receipt`: POST `/salesreceipts?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_sales_receipt`: PUT `/salesreceipts/{{ record.sales_receipt_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `sales_receipt_id`;
  required record fields `sales_receipt_id`; accepted fields `sales_receipt_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `delete_sales_receipt`: DELETE `/salesreceipts/{{ record.sales_receipt_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `sales_receipt_id`;
  required record fields `sales_receipt_id`; accepted fields `sales_receipt_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `email_sales_receipt`: POST `/salesreceipts/{{ record.sales_receipt_id }}/email` - kind `create`;
  body type `json`; path fields `sales_receipt_id`; required record fields `sales_receipt_id`;
  accepted fields `sales_receipt_id`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `add_task`: POST `/tasks?organization_id={{ config.organization_id }}` - kind `create`; body type
  `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_tasks`: PUT `/tasks?organization_id={{ config.organization_id }}&bulk_update={{
  record.bulk_update }}` - kind `update`; body type `json`; path fields `bulk_update`; required
  record fields `bulk_update`; accepted fields `bulk_update`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_tasks`: DELETE `/tasks?organization_id={{ config.organization_id }}&task_ids={{
  record.task_ids }}` - kind `delete`; body type `none`; path fields `task_ids`; required record
  fields `task_ids`; accepted fields `task_ids`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books; approval
  required.
- `update_a_task`: PUT `/tasks/{{ record.task_id }}?organization_id={{ config.organization_id }}` -
  kind `update`; body type `json`; path fields `task_id`; required record fields `task_id`; accepted
  fields `task_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_task`: DELETE `/tasks/{{ record.task_id }}?organization_id={{ config.organization_id }}` -
  kind `delete`; body type `none`; path fields `task_id`; required record fields `task_id`; accepted
  fields `task_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: external destructive mutation in Zoho Books; approval required.
- `add_task_attachment`: POST `/tasks/{{ record.task_id }}/attachment?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `task_id`; required
  record fields `task_id`; accepted fields `task_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `add_task_comment`: POST `/tasks/{{ record.task_id }}/comments?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `task_id`; required
  record fields `task_id`; accepted fields `task_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_task_comment`: DELETE `/tasks/{{ record.task_id }}/comments/{{ record.comment_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `task_id`, `comment_id`; required record fields `task_id`, `comment_id`; accepted fields
  `comment_id`, `task_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `delete_task_document`: DELETE `/tasks/{{ record.task_id }}/documents/{{ record.document_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `task_id`, `document_id`; required record fields `task_id`, `document_id`; accepted fields
  `document_id`, `task_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `mark_task_as_completed`: POST `/tasks/{{ record.task_id }}/markascompleted?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `task_id`; required
  record fields `task_id`; accepted fields `task_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_task_as_ongoing`: POST `/tasks/{{ record.task_id }}/markasongoing?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `task_id`; required
  record fields `task_id`; accepted fields `task_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_task_as_open`: POST `/tasks/{{ record.task_id }}/markasopen?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `task_id`; required
  record fields `task_id`; accepted fields `task_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `update_percentage_task`: POST `/tasks/{{ record.task_id }}/percentage?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `task_id`; required
  record fields `task_id`; accepted fields `task_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `create_tax_authority`: POST `/settings/taxauthorities?organization_id={{ config.organization_id
  }}` - kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `update_tax_authority`: PUT `/settings/taxauthorities/{{ record.tax_authority_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `tax_authority_id`; required record fields `tax_authority_id`; accepted fields `tax_authority_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `delete_tax_authority`: DELETE `/settings/taxauthorities/{{ record.tax_authority_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `tax_authority_id`; required record fields `tax_authority_id`; accepted fields `tax_authority_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `create_tax`: POST `/settings/taxes?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_tax`: PUT `/settings/taxes/{{ record.tax_id }}?organization_id={{ config.organization_id
  }}` - kind `update`; body type `json`; path fields `tax_id`; required record fields `tax_id`;
  accepted fields `tax_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_tax`: DELETE `/settings/taxes/{{ record.tax_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `tax_id`; required
  record fields `tax_id`; accepted fields `tax_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books; approval
  required.
- `create_tax_exemption`: POST `/settings/taxexemptions?organization_id={{ config.organization_id
  }}` - kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data;
  approval required.
- `update_tax_exemption`: PUT `/settings/taxexemptions/{{ record.tax_exemption_id
  }}?organization_id={{ config.organization_id }}` - kind `update`; body type `json`; path fields
  `tax_exemption_id`; required record fields `tax_exemption_id`; accepted fields `tax_exemption_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `delete_tax_exemption`: DELETE `/settings/taxexemptions/{{ record.tax_exemption_id
  }}?organization_id={{ config.organization_id }}` - kind `delete`; body type `none`; path fields
  `tax_exemption_id`; required record fields `tax_exemption_id`; accepted fields `tax_exemption_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `create_tax_group`: POST `/settings/taxgroups?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_tax_group`: PUT `/settings/taxgroups/{{ record.tax_group_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `tax_group_id`; required
  record fields `tax_group_id`; accepted fields `tax_group_id`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `delete_tax_group`: DELETE `/settings/taxgroups/{{ record.tax_group_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `tax_group_id`; required
  record fields `tax_group_id`; accepted fields `tax_group_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `create_time_entries`: POST `/projects/timeentries?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_time_entries`: DELETE `/projects/timeentries?organization_id={{ config.organization_id }}`
  - kind `delete`; body type `none`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `stop_entry_timer`: POST `/projects/timeentries/timer/stop?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `update_time_entry`: PUT `/projects/timeentries/{{ record.time_entry_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `time_entry_id`;
  required record fields `time_entry_id`; accepted fields `time_entry_id`; risk: external mutation
  in Zoho Books accounting data; approval required.
- `delete_time_entry`: DELETE `/projects/timeentries/{{ record.time_entry_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `time_entry_id`;
  required record fields `time_entry_id`; accepted fields `time_entry_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: external destructive mutation in
  Zoho Books; approval required.
- `start_entry_timer`: POST `/projects/timeentries/{{ record.time_entry_id
  }}/timer/start?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `time_entry_id`; required record fields `time_entry_id`; accepted fields
  `time_entry_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_transaction_lock`: PUT `/transactionlock?organization_id={{ config.organization_id }}` -
  kind `update`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_transaction_lock`: DELETE `/transactionlock?organization_id={{ config.organization_id
  }}&transaction_lock_id={{ record.transaction_lock_id }}` - kind `delete`; body type `json`; path
  fields `transaction_lock_id`; required record fields `transaction_lock_id`; accepted fields
  `transaction_lock_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `update_partial_unlock`: PUT `/transactionlock/partialunlock?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; risk: external mutation in Zoho
  Books accounting data; approval required.
- `create_user`: POST `/users?organization_id={{ config.organization_id }}` - kind `create`; body
  type `json`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_user`: PUT `/users/{{ record.user_id }}?organization_id={{ config.organization_id }}` -
  kind `update`; body type `json`; path fields `user_id`; required record fields `user_id`; accepted
  fields `user_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_user`: DELETE `/users/{{ record.user_id }}?organization_id={{ config.organization_id }}` -
  kind `delete`; body type `none`; path fields `user_id`; required record fields `user_id`; accepted
  fields `user_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: external destructive mutation in Zoho Books; approval required.
- `mark_user_active`: POST `/users/{{ record.user_id }}/active?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `user_id`; required
  record fields `user_id`; accepted fields `user_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `mark_user_inactive`: POST `/users/{{ record.user_id }}/inactive?organization_id={{
  config.organization_id }}` - kind `create`; body type `none`; path fields `user_id`; required
  record fields `user_id`; accepted fields `user_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `invite_user`: POST `/users/{{ record.user_id }}/invite?organization_id={{ config.organization_id
  }}` - kind `create`; body type `none`; path fields `user_id`; required record fields `user_id`;
  accepted fields `user_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `create_vendor_credit`: POST `/vendorcredits?organization_id={{ config.organization_id }}` - kind
  `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `update_vendor_credit`: PUT `/vendorcredits/{{ record.vendor_credit_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `vendor_credit_id`;
  required record fields `vendor_credit_id`; accepted fields `vendor_credit_id`; risk: external
  mutation in Zoho Books accounting data; approval required.
- `delete_vendor_credit`: DELETE `/vendorcredits/{{ record.vendor_credit_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `vendor_credit_id`;
  required record fields `vendor_credit_id`; accepted fields `vendor_credit_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external destructive
  mutation in Zoho Books; approval required.
- `approve_vendor_credit`: POST `/vendorcredits/{{ record.vendor_credit_id
  }}/approve?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `vendor_credit_id`; required record fields `vendor_credit_id`; accepted fields
  `vendor_credit_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `apply_credits_to_a_bill`: POST `/vendorcredits/{{ record.vendor_credit_id
  }}/bills?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `vendor_credit_id`; required record fields `vendor_credit_id`; accepted fields
  `vendor_credit_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_vendor_credit_bill`: DELETE `/vendorcredits/{{ record.vendor_credit_id }}/bills/{{
  record.vendor_credit_bill_id }}?organization_id={{ config.organization_id }}` - kind `delete`;
  body type `none`; path fields `vendor_credit_id`, `vendor_credit_bill_id`; required record fields
  `vendor_credit_id`, `vendor_credit_bill_id`; accepted fields `vendor_credit_bill_id`,
  `vendor_credit_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `add_vendor_credit_comment`: POST `/vendorcredits/{{ record.vendor_credit_id
  }}/comments?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `vendor_credit_id`; required record fields `vendor_credit_id`; accepted fields
  `vendor_credit_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `delete_vendor_credit_comment`: DELETE `/vendorcredits/{{ record.vendor_credit_id }}/comments/{{
  record.comment_id }}?organization_id={{ config.organization_id }}` - kind `delete`; body type
  `none`; path fields `vendor_credit_id`, `comment_id`; required record fields `vendor_credit_id`,
  `comment_id`; accepted fields `comment_id`, `vendor_credit_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `refund_vendor_credit`: POST `/vendorcredits/{{ record.vendor_credit_id
  }}/refunds?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `vendor_credit_id`; required record fields `vendor_credit_id`; accepted fields
  `vendor_credit_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `update_vendor_credit_refund`: PUT `/vendorcredits/{{ record.vendor_credit_id }}/refunds/{{
  record.vendor_credit_refund_id }}?organization_id={{ config.organization_id }}` - kind `update`;
  body type `json`; path fields `vendor_credit_id`, `vendor_credit_refund_id`; required record
  fields `vendor_credit_id`, `vendor_credit_refund_id`; accepted fields `vendor_credit_id`,
  `vendor_credit_refund_id`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `delete_vendor_credit_refund`: DELETE `/vendorcredits/{{ record.vendor_credit_id }}/refunds/{{
  record.vendor_credit_refund_id }}?organization_id={{ config.organization_id }}` - kind `delete`;
  body type `none`; path fields `vendor_credit_id`, `vendor_credit_refund_id`; required record
  fields `vendor_credit_id`, `vendor_credit_refund_id`; accepted fields `vendor_credit_id`,
  `vendor_credit_refund_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: external destructive mutation in Zoho Books; approval required.
- `mark_vendor_credit_open`: POST `/vendorcredits/{{ record.vendor_credit_id
  }}/status/open?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `vendor_credit_id`; required record fields `vendor_credit_id`; accepted fields
  `vendor_credit_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `mark_vendor_credit_void`: POST `/vendorcredits/{{ record.vendor_credit_id
  }}/status/void?organization_id={{ config.organization_id }}` - kind `create`; body type `none`;
  path fields `vendor_credit_id`; required record fields `vendor_credit_id`; accepted fields
  `vendor_credit_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `submit_vendor_credit`: POST `/vendorcredits/{{ record.vendor_credit_id
  }}/submit?organization_id={{ config.organization_id }}` - kind `create`; body type `none`; path
  fields `vendor_credit_id`; required record fields `vendor_credit_id`; accepted fields
  `vendor_credit_id`; risk: external mutation in Zoho Books accounting data; approval required.
- `create_vendor_payment`: POST `/vendorpayments?organization_id={{ config.organization_id }}` -
  kind `create`; body type `json`; risk: external mutation in Zoho Books accounting data; approval
  required.
- `bulk_delete_vendor_payments`: DELETE `/vendorpayments?organization_id={{ config.organization_id
  }}&vendorpayment_id={{ record.vendorpayment_id }}&bulk_delete={{ record.bulk_delete }}` - kind
  `delete`; body type `none`; path fields `vendorpayment_id`, `bulk_delete`; required record fields
  `vendorpayment_id`, `bulk_delete`; accepted fields `bulk_delete`, `vendorpayment_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.
- `update_vendor_payment`: PUT `/vendorpayments/{{ record.payment_id }}?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `payment_id`; required
  record fields `payment_id`; accepted fields `payment_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `delete_vendor_payment`: DELETE `/vendorpayments/{{ record.payment_id }}?organization_id={{
  config.organization_id }}` - kind `delete`; body type `none`; path fields `payment_id`; required
  record fields `payment_id`; accepted fields `payment_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: external destructive mutation in Zoho Books;
  approval required.
- `email_vendor_payment`: POST `/vendorpayments/{{ record.payment_id }}/email?organization_id={{
  config.organization_id }}` - kind `create`; body type `json`; path fields `payment_id`; required
  record fields `payment_id`; accepted fields `payment_id`; risk: external mutation in Zoho Books
  accounting data; approval required.
- `refund_excess_vendor_payment`: POST `/vendorpayments/{{ record.payment_id
  }}/refunds?organization_id={{ config.organization_id }}` - kind `create`; body type `json`; path
  fields `payment_id`; required record fields `payment_id`; accepted fields `payment_id`; risk:
  external mutation in Zoho Books accounting data; approval required.
- `update_vendor_payment_refund`: PUT `/vendorpayments/{{ record.payment_id }}/refunds/{{
  record.vendorpayment_refund_id }}?organization_id={{ config.organization_id }}` - kind `update`;
  body type `json`; path fields `payment_id`, `vendorpayment_refund_id`; required record fields
  `payment_id`, `vendorpayment_refund_id`; accepted fields `payment_id`, `vendorpayment_refund_id`;
  risk: external mutation in Zoho Books accounting data; approval required.
- `delete_vendor_payment_refund`: DELETE `/vendorpayments/{{ record.payment_id }}/refunds/{{
  record.vendorpayment_refund_id }}?organization_id={{ config.organization_id }}` - kind `delete`;
  body type `none`; path fields `payment_id`, `vendorpayment_refund_id`; required record fields
  `payment_id`, `vendorpayment_refund_id`; accepted fields `payment_id`, `vendorpayment_refund_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: external
  destructive mutation in Zoho Books; approval required.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 174 stream-backed endpoint group(s), 569 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=6, non_data_endpoint=84, out_of_scope=15.
