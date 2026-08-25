# pm connectors inspect zoho-books

```text
NAME
  pm connectors inspect zoho-books - Zoho Books connector manual

SYNOPSIS
  pm connectors inspect zoho-books
  pm connectors inspect zoho-books --json
  pm credentials add <name> --connector zoho-books [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Zoho Books API v3 accounting resources using the declarative connector engine.

ICON
  id: simple-icons-zoho-books
  asset: icons/simple-icons/zoho-books.svg
  title: Zoho
  simple_icon_slug: zoho
  simple_icon_hex: E42527
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Zoho
  match: curated-alias
  matched_by: zoho

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  accept
  account_id
  adjustment_date
  bank_account_id
  bank_transaction_id
  base_currency_adjustment_id
  base_url
  bill_id
  card_id
  comment_id
  contact_id
  contact_person_id
  contactperson_id
  creditnote_id
  creditnote_ids
  creditnote_refund_id
  currency_id
  customer_payment_id
  debit_note_id
  deliverychallan_id
  document_id
  email_template_id
  employee_id
  entity_type
  estimate_id
  estimate_ids
  exchange_rate
  exchange_rate_id
  expense_id
  expiry_time
  fixed_asset_id
  invoice_id
  invoice_ids
  item_id
  item_ids
  journal_id
  line1
  link_type
  max_pages
  metadata_name
  mode
  module_api_name
  module_id
  module_name
  notes
  organization_id
  page_size
  payment_id
  project_id
  purchaseorder_id
  purchaseorder_ids
  reconciliation_id
  recurring_bill_id
  recurring_expense_id
  recurring_invoice_id
  recurring_journal_id
  reference_id
  refund_id
  retainerinvoice_id
  rule_id
  sales_receipt_id
  salesorder_id
  tag_id
  task_id
  tax_authority_id
  tax_exemption_id
  tax_group_id
  tax_id
  time_entry_id
  transaction_id
  transaction_type
  user_id
  vendor_credit_id
  vendor_credit_refund_id
  vendorpayment_refund_id
  access_token (secret) (required)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: contact_id(string), contact_name(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  invoices:
    primary key: id
    cursor: updated_at
    fields: id(string), invoice_id(string), invoice_number(string), last_modified_time(string), name(string), status(string), updated_at(string)
  items:
    primary key: id
    cursor: updated_at
    fields: id(string), item_id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_bank_accounts:
    primary key: id
    cursor: updated_at
    fields: account_id(string), account_name(string), id(string), name(string), status(string), updated_at(string)
  get_bank_account:
    primary key: id
    cursor: updated_at
    fields: account_id(string), account_name(string), id(string), name(string), status(string), updated_at(string)
  get_last_imported_bank_statement:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), statement_id(string), status(string), updated_at(string)
  list_bank_account_rules:
    primary key: id
    cursor: updated_at
    fields: account_name(string), id(string), name(string), rule_id(string), status(string), updated_at(string)
  get_bank_account_rule:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), target_account_id(string), target_account_name(string), updated_at(string)
  list_bank_transactions:
    primary key: id
    cursor: updated_at
    fields: account_name(string), date(string), id(string), name(string), status(string), transaction_id(string), updated_at(string)
  get_matching_bank_transactions:
    primary key: id
    cursor: updated_at
    fields: contact_name(string), date(string), id(string), name(string), status(string), transaction_id(string), updated_at(string)
  get_bank_transaction:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), date(string), id(string), name(string), status(string), transaction_id(string), updated_at(string)
  list_base_currency_adjustments:
    primary key: id
    cursor: updated_at
    fields: base_currency_adjustment_id(string), id(string), name(string), status(string), updated_at(string)
  list_base_currency_adjustment_accounts:
    primary key: id
    cursor: updated_at
    fields: currency_id(string), id(string), name(string), status(string), updated_at(string)
  list_base_currency_adjustment_contacts:
    primary key: id
    cursor: updated_at
    fields: contact_id(string), contact_name(string), id(string), name(string), status(string), updated_at(string)
  get_base_currency_adjustment:
    primary key: id
    cursor: updated_at
    fields: base_currency_adjustment_id(string), id(string), name(string), status(string), updated_at(string)
  list_bills:
    primary key: id
    cursor: updated_at
    fields: bill_id(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string), vendor_name(string)
  convert_purchase_order_to_bill:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  get_bill:
    primary key: id
    cursor: updated_at
    fields: bill_id(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string), vendor_name(string)
  get_bill_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  list_bill_payments:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), name(string), payment_id(string), status(string), updated_at(string), vendor_name(string)
  list_chart_of_accounts:
    primary key: id
    cursor: updated_at
    fields: account_id(string), account_name(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_chart_of_account_transactions:
    primary key: id
    cursor: updated_at
    fields: categorized_transaction_id(string), id(string), name(string), offset_account_name(string), status(string), updated_at(string)
  get_chart_of_account:
    primary key: id
    cursor: updated_at
    fields: customfield_id(string), id(string), name(string), status(string), updated_at(string)
  list_contact_persons:
    primary key: id
    cursor: updated_at
    fields: contact_person_id(string), first_name(string), id(string), name(string), status(string), updated_at(string)
  get_contact_person:
    primary key: id
    cursor: updated_at
    fields: contact_id(string), first_name(string), id(string), name(string), status(string), updated_at(string)
  get_contact:
    primary key: id
    cursor: updated_at
    fields: contact_id(string), contact_name(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  get_contact_address:
    primary key: id
    cursor: updated_at
    fields: address_id(string), id(string), name(string), status(string), updated_at(string)
  list_contact_autobill_recurring_invoices:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), recurrence_name(string), recurring_invoice_id(string), status(string), updated_at(string)
  list_contact_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), contact_name(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_unused_retainer_payments:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), name(string), payment_id(string), payment_number(string), status(string), updated_at(string)
  list_contact_credit_note_refunds:
    primary key: id
    cursor: updated_at
    fields: creditnote_refund_id(string), customer_name(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_contact_statement_mail:
    primary key: id
    cursor: updated_at
    fields: contact_id(string), file_name(string), id(string), name(string), status(string), updated_at(string)
  list_credit_notes:
    primary key: id
    cursor: updated_at
    fields: creditnote_id(string), customer_name(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_credit_note_refunds_of_all_credit_notes:
    primary key: id
    cursor: updated_at
    fields: creditnote_refund_id(string), customer_name(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_credit_note_refund_by_id:
    primary key: id
    cursor: updated_at
    fields: creditnote_refund_id(string), customer_name(string), date(string), id(string), name(string), status(string), updated_at(string)
  list_credit_note_templates:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), template_id(string), template_name(string), updated_at(string)
  get_credit_note:
    primary key: id
    cursor: updated_at
    fields: creditnote_id(string), customer_name(string), id(string), name(string), status(string), updated_at(string), updated_time(string)
  list_credit_note_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_credit_note_custom_fields:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  get_credit_note_email:
    primary key: id
    cursor: updated_at
    fields: customer_id(string), file_name(string), id(string), name(string), status(string), updated_at(string)
  get_credit_note_email_history:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), mailhistory_id(string), name(string), status(string), updated_at(string)
  list_invoices_of_credit_note:
    primary key: id
    cursor: updated_at
    fields: creditnote_id(string), date(string), id(string), invoice_number(string), name(string), status(string), updated_at(string)
  list_credit_note_refunds_of_a_credit_note:
    primary key: id
    cursor: updated_at
    fields: creditnote_refund_id(string), customer_name(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_credit_note_refund:
    primary key: id
    cursor: updated_at
    fields: creditnote_refund_id(string), customer_name(string), date(string), id(string), name(string), status(string), updated_at(string)
  list_currencies:
    primary key: id
    cursor: updated_at
    fields: currency_id(string), currency_name(string), id(string), name(string), status(string), updated_at(string)
  get_currency:
    primary key: id
    cursor: updated_at
    fields: currency_id(string), currency_name(string), id(string), name(string), status(string), updated_at(string)
  list_exchange_rates:
    primary key: id
    cursor: updated_at
    fields: exchange_rate_id(string), id(string), name(string), status(string), updated_at(string)
  get_exchange_rate:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  list_custom_modules:
    primary key: id
    cursor: updated_at
    fields: id(string), module_id(string), module_name(string), name(string), status(string), updated_at(string)
  get_custom_module:
    primary key: id
    cursor: updated_at
    fields: id(string), module_id(string), module_name(string), name(string), status(string), updated_at(string)
  list_custom_module_records:
    primary key: id
    cursor: updated_at
    fields: id(string), last_modified_time(string), module_api_name(string), module_record_id(string), name(string), status(string), updated_at(string)
  get_custom_module_record:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  get_customer_debit_note:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), invoice_id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_customer_payments:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), invoice_number(string), name(string), payment_id(string), status(string), updated_at(string)
  list_customer_payment_refunds:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), date(string), id(string), name(string), payment_refund_id(string), status(string), updated_at(string)
  get_customer_payment_refund:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), name(string), payment_refund_id(string), reference_number(string), status(string), updated_at(string)
  get_customer_payment:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), date(string), id(string), name(string), payment_id(string), status(string), updated_at(string)
  list_delivery_challans:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), deliverychallan_id(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_delivery_challan_templates:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), template_id(string), template_name(string), updated_at(string)
  get_delivery_challan:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), deliverychallan_id(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_estimates:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), estimate_id(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_estimate_templates:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), template_id(string), template_name(string), updated_at(string)
  get_estimate:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), estimate_id(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_estimate_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_estimate_email:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  list_employees:
    primary key: id
    cursor: updated_at
    fields: employee_id(string), id(string), name(string), status(string), updated_at(string)
  get_employee:
    primary key: id
    cursor: updated_at
    fields: employee_id(string), id(string), name(string), status(string), updated_at(string)
  list_expenses:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), expense_id(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  get_expense:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), expense_id(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_expense_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), id(string), name(string), status(string), updated_at(string)
  list_fixed_assets:
    primary key: id
    cursor: updated_at
    fields: asset_name(string), fixed_asset_id(string), id(string), name(string), status(string), updated_at(string)
  get_fixed_asset:
    primary key: id
    cursor: updated_at
    fields: asset_name(string), fixed_asset_id(string), id(string), name(string), status(string), updated_at(string)
  get_fixed_asset_forecast:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  get_fixed_asset_history:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string), valuation_id(string)
  get_fixed_asset_type_list:
    primary key: id
    cursor: updated_at
    fields: fixed_asset_type_id(string), fixed_asset_type_name(string), id(string), name(string), status(string), updated_at(string)
  list_invoice_templates:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), template_id(string), template_name(string), updated_at(string)
  get_invoice:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), invoice_id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  list_invoice_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  list_invoice_credits_applied:
    primary key: id
    cursor: updated_at
    fields: creditnote_id(string), creditnotes_number(string), id(string), name(string), status(string), updated_at(string)
  get_invoice_document_details:
    primary key: id
    cursor: updated_at
    fields: document_id(string), document_name(string), id(string), last_modified_time(string), name(string), status(string), updated_at(string)
  get_invoice_email:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), entity_id(string), id(string), name(string), status(string), updated_at(string)
  get_payment_reminder_mail_content_for_invoice:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), entity_id(string), id(string), name(string), status(string), updated_at(string)
  list_invoice_payments:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), name(string), payment_id(string), payment_number(string), status(string), updated_at(string)
  generate_invoice_payment_link:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  list_item_details:
    primary key: id
    cursor: updated_at
    fields: id(string), item_id(string), name(string), status(string), updated_at(string)
  get_item:
    primary key: id
    cursor: updated_at
    fields: id(string), item_id(string), name(string), status(string), updated_at(string)
  list_journals:
    primary key: id
    cursor: updated_at
    fields: entry_number(string), id(string), journal_id(string), name(string), status(string), updated_at(string)
  get_journal:
    primary key: id
    cursor: updated_at
    fields: id(string), journal_id(string), last_modified_time(string), location_name(string), name(string), status(string), updated_at(string)
  list_journal_credits:
    primary key: id
    cursor: updated_at
    fields: credit_id(string), id(string), name(string), status(string), updated_at(string)
  list_recurring_journals:
    primary key: id
    cursor: updated_at
    fields: id(string), journal_id(string), last_modified_time(string), location_name(string), name(string), status(string), updated_at(string)
  get_recurring_journal:
    primary key: id
    cursor: updated_at
    fields: id(string), journal_id(string), last_modified_time(string), location_name(string), name(string), status(string), updated_at(string)
  list_child_journals:
    primary key: id
    cursor: updated_at
    fields: entry_number(string), id(string), journal_id(string), name(string), status(string), updated_at(string)
  get_transaction_journal_view:
    primary key: id
    cursor: updated_at
    fields: id(string), journal_id(string), last_modified_time(string), location_name(string), name(string), status(string), updated_at(string)
  list_locations:
    primary key: id
    cursor: updated_at
    fields: id(string), location_id(string), location_name(string), name(string), status(string), updated_at(string)
  list_opening_balance_transactions:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), transaction_id(string), updated_at(string)
  list_opening_balance_details:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), name(string), opening_balance_id(string), status(string), updated_at(string)
  get_opening_balance:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), name(string), opening_balance_id(string), status(string), updated_at(string)
  list_organizations:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), organization_id(string), status(string), updated_at(string)
  list_organizations_for_user:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), organization_id(string), status(string), updated_at(string)
  get_organization:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), organization_id(string), status(string), updated_at(string)
  list_pricebooks:
    primary key: id
    cursor: updated_at
    fields: currency_id(string), id(string), name(string), status(string), updated_at(string)
  list_projects:
    primary key: id
    cursor: updated_at
    fields: created_time(string), customer_name(string), id(string), name(string), project_id(string), status(string), updated_at(string)
  get_project:
    primary key: id
    cursor: updated_at
    fields: created_time(string), customer_name(string), id(string), name(string), project_id(string), status(string), updated_at(string)
  list_project_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  list_project_invoices:
    primary key: id
    cursor: updated_at
    fields: created_time(string), customer_name(string), id(string), invoice_id(string), name(string), status(string), updated_at(string)
  list_project_tasks:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), name(string), project_id(string), status(string), updated_at(string)
  get_project_task:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), project_id(string), project_name(string), status(string), updated_at(string)
  list_project_users:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string), user_id(string), user_name(string)
  get_project_user:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string), user_id(string), user_name(string)
  list_purchase_orders:
    primary key: id
    cursor: updated_at
    fields: id(string), last_modified_time(string), name(string), purchaseorder_id(string), status(string), updated_at(string), vendor_name(string)
  list_purchase_order_templates:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), template_id(string), template_name(string), updated_at(string)
  get_purchase_order:
    primary key: id
    cursor: updated_at
    fields: id(string), last_modified_time(string), name(string), purchaseorder_id(string), status(string), updated_at(string), vendor_name(string)
  list_purchase_order_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_recurring_bill:
    primary key: id
    cursor: updated_at
    fields: id(string), last_modified_time(string), name(string), recurring_bill_id(string), status(string), updated_at(string), vendor_name(string)
  list_recurring_bills:
    primary key: id
    cursor: updated_at
    fields: id(string), last_modified_time(string), name(string), recurring_bill_id(string), status(string), updated_at(string), vendor_name(string)
  list_recurring_expenses:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), last_modified_time(string), name(string), recurring_expense_id(string), status(string), updated_at(string)
  get_recurring_expense:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), last_modified_time(string), name(string), recurring_expense_id(string), status(string), updated_at(string)
  list_recurring_expense_history:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  list_child_expenses_of_recurring_expense:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), date(string), expense_id(string), id(string), name(string), status(string), updated_at(string)
  list_recurring_invoices:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), name(string), recurring_invoice_id(string), status(string), updated_at(string)
  get_recurring_invoice:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), name(string), recurring_invoice_id(string), status(string), updated_at(string)
  list_recurring_invoice_history:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  list_recurring_invoice_child_invoices:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), date(string), id(string), invoice_id(string), name(string), status(string), updated_at(string)
  get_register_budget_vs_actuals:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  list_register_bulk_action_history:
    primary key: id
    cursor: updated_at
    fields: bulk_action_id(string), date(string), id(string), name(string), status(string), updated_at(string), user_name(string)
  get_register_bulk_action_history:
    primary key: id
    cursor: updated_at
    fields: account_name(string), bulk_action_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_register_bulk_update_editpage:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  list_register_transactions:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  get_tags:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), tag_id(string), tag_name(string), updated_at(string)
  all_tag_options:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), option_id(string), option_name(string), status(string), updated_at(string)
  get_all_tag_options:
    primary key: id
    cursor: updated_at
    fields: dependent_id(string), dependent_name(string), id(string), name(string), status(string), updated_at(string)
  list_retainer_invoices:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), last_modified_time(string), name(string), retainerinvoice_id(string), status(string), updated_at(string)
  list_retainer_invoice_templates:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), template_id(string), template_name(string), updated_at(string)
  get_retainer_invoice:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), last_modified_time(string), name(string), retainerinvoice_id(string), status(string), updated_at(string)
  list_retainer_invoice:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_retainer_invoice_email:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string)
  list_sales_orders:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), last_modified_time(string), name(string), salesorder_id(string), status(string), updated_at(string)
  list_sales_order_templates:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), template_id(string), template_name(string), updated_at(string)
  get_sales_order:
    primary key: id
    cursor: updated_at
    fields: customer_id(string), date(string), id(string), name(string), salesperson_name(string), status(string), updated_at(string)
  get_sales_order_email:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), entity_id(string), id(string), name(string), status(string), updated_at(string)
  list_sales_receipts:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), last_modified_time(string), name(string), sales_receipt_id(string), status(string), updated_at(string)
  get_sales_receipt:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), id(string), last_modified_time(string), name(string), sales_receipt_id(string), status(string), updated_at(string)
  list_tasks:
    primary key: id
    cursor: updated_at
    fields: contact_name(string), created_time(string), id(string), name(string), status(string), task_id(string), updated_at(string)
  get_task:
    primary key: id
    cursor: updated_at
    fields: contact_name(string), id(string), last_modified_by_id(string), name(string), status(string), task_id(string), updated_at(string)
  list_task_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  get_task_document:
    primary key: id
    cursor: updated_at
    fields: document_id(string), file_name(string), id(string), name(string), status(string), updated_at(string)
  list_tax_authorities:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), tax_authority_id(string), tax_authority_name(string), updated_at(string)
  get_tax_authority:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), tax_authority_id(string), tax_authority_name(string), updated_at(string)
  list_taxes:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), tax_id(string), tax_name(string), updated_at(string)
  get_tax:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), tax_id(string), tax_name(string), updated_at(string)
  list_tax_exemptions:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), tax_exemption_id(string), updated_at(string)
  get_tax_exemption:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), tax_exemption_id(string), updated_at(string)
  get_tax_group:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), tax_group_id(string), tax_group_name(string), updated_at(string)
  list_time_entries:
    primary key: id
    cursor: updated_at
    fields: created_time(string), customer_name(string), id(string), name(string), status(string), time_entry_id(string), updated_at(string)
  get_running_timer:
    primary key: id
    cursor: updated_at
    fields: created_time(string), customer_name(string), id(string), name(string), status(string), time_entry_id(string), updated_at(string)
  get_time_entry:
    primary key: id
    cursor: updated_at
    fields: created_time(string), customer_name(string), id(string), name(string), status(string), time_entry_id(string), updated_at(string)
  get_accounting_period_transaction_lock:
    primary key: id
    cursor: updated_at
    fields: accounting_period_name(string), id(string), name(string), status(string), transaction_lock_id(string), updated_at(string)
  get_transaction_lock:
    primary key: id
    cursor: updated_at
    fields: accounting_period_name(string), id(string), name(string), status(string), transaction_lock_id(string), updated_at(string)
  list_transaction_locks:
    primary key: id
    cursor: updated_at
    fields: accounting_period_name(string), id(string), name(string), status(string), transaction_lock_id(string), updated_at(string)
  list_users:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), status(string), updated_at(string), user_id(string)
  get_current_user:
    primary key: id
    cursor: updated_at
    fields: created_time(string), id(string), name(string), status(string), updated_at(string), user_id(string)
  get_user:
    primary key: id
    cursor: updated_at
    fields: created_time(string), id(string), name(string), status(string), updated_at(string), user_id(string)
  list_vendor_credits:
    primary key: id
    cursor: updated_at
    fields: id(string), last_modified_time(string), name(string), status(string), updated_at(string), vendor_credit_id(string), vendor_name(string)
  list_vendor_credit_refunds_of_all_vendor_credits:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), date(string), id(string), name(string), status(string), updated_at(string), vendor_credit_refund_id(string)
  get_vendor_credit:
    primary key: id
    cursor: updated_at
    fields: id(string), last_modified_time(string), name(string), status(string), updated_at(string), vendor_credit_id(string), vendor_name(string)
  list_bills_credited:
    primary key: id
    cursor: updated_at
    fields: bill_number(string), date(string), id(string), name(string), status(string), updated_at(string), vendor_credit_id(string)
  list_vendor_credit_comments:
    primary key: id
    cursor: updated_at
    fields: comment_id(string), date(string), id(string), name(string), status(string), updated_at(string)
  list_vendor_credit_refunds_of_a_vendor_credit:
    primary key: id
    cursor: updated_at
    fields: customer_name(string), date(string), id(string), name(string), status(string), updated_at(string), vendor_credit_refund_id(string)
  get_vendor_credit_refund:
    primary key: id
    cursor: updated_at
    fields: account_name(string), date(string), id(string), name(string), status(string), updated_at(string), vendor_credit_refund_id(string)
  list_vendor_payments:
    primary key: id
    cursor: updated_at
    fields: id(string), last_modified_time(string), name(string), payment_id(string), status(string), updated_at(string), vendor_name(string)
  get_vendor_payment:
    primary key: id
    cursor: updated_at
    fields: id(string), last_modified_time(string), name(string), payment_id(string), status(string), updated_at(string), vendor_name(string)
  get_vendor_payment_email_content:
    primary key: id
    cursor: updated_at
    fields: contact_person_id(string), first_name(string), id(string), name(string), status(string), updated_at(string)
  list_vendor_payment_refunds:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), name(string), reference_number(string), status(string), updated_at(string), vendorpayment_refund_id(string)
  get_vendor_payment_refund:
    primary key: id
    cursor: updated_at
    fields: date(string), id(string), name(string), status(string), to_account_name(string), updated_at(string), vendorpayment_refund_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_bank_account:
    endpoint: POST /bankaccounts?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_bank_account:
    endpoint: PUT /bankaccounts/{{ record.account_id }}?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bank_account:
    endpoint: DELETE /bankaccounts/{{ record.account_id }}?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_bank_account_active:
    endpoint: POST /bankaccounts/{{ record.account_id }}/active?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_bank_account_inactive:
    endpoint: POST /bankaccounts/{{ record.account_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_bank_account_preferences:
    endpoint: PUT /bankaccounts/{{ record.account_id }}/preferences?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_bank_reconciliation:
    endpoint: POST /bankaccounts/{{ record.account_id }}/reconciliations?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_bank_reconciliation:
    endpoint: PUT /bankaccounts/{{ record.account_id }}/reconciliations/{{ record.reconciliation_id }}?organization_id={{ config.organization_id }}
    required fields: account_id, reconciliation_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bank_reconciliation:
    endpoint: DELETE /bankaccounts/{{ record.account_id }}/reconciliations/{{ record.reconciliation_id }}?organization_id={{ config.organization_id }}
    required fields: account_id, reconciliation_id
    risk: external destructive mutation in Zoho Books; approval required
  add_bank_reconciliation_attachment:
    endpoint: POST /bankaccounts/{{ record.account_id }}/reconciliations/{{ record.reconciliation_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: account_id, reconciliation_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bank_reconciliation_document:
    endpoint: DELETE /bankaccounts/{{ record.account_id }}/reconciliations/{{ record.reconciliation_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: account_id, reconciliation_id, document_id
    risk: external destructive mutation in Zoho Books; approval required
  save_bank_reconciliation_draft:
    endpoint: PUT /bankaccounts/{{ record.account_id }}/reconciliations/{{ record.reconciliation_id }}/draft?organization_id={{ config.organization_id }}
    required fields: account_id, reconciliation_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_last_imported_bank_statement:
    endpoint: DELETE /bankaccounts/{{ record.account_id }}/statement/{{ record.statement_id }}?organization_id={{ config.organization_id }}
    required fields: account_id, statement_id
    risk: external destructive mutation in Zoho Books; approval required
  import_bank_statements:
    endpoint: POST /bankstatements?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  create_bank_account_match_filter:
    endpoint: POST /bankaccounts/matchfilters?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_bank_account_match_filter:
    endpoint: PUT /bankaccounts/matchfilters/{{ record.match_filter_id }}?organization_id={{ config.organization_id }}
    required fields: match_filter_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bank_account_match_filter:
    endpoint: DELETE /bankaccounts/matchfilters/{{ record.match_filter_id }}?organization_id={{ config.organization_id }}
    required fields: match_filter_id
    risk: external destructive mutation in Zoho Books; approval required
  create_bank_account_rule:
    endpoint: POST /bankaccounts/rules?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_update_bank_account_rules:
    endpoint: PUT /bankaccounts/rules?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_delete_bank_account_rules:
    endpoint: DELETE /bankaccounts/rules?organization_id={{ config.organization_id }}&rule_ids={{ record.rule_ids }}
    required fields: rule_ids
    risk: external destructive mutation in Zoho Books; approval required
  reorder_bank_account_rules:
    endpoint: POST /bankaccounts/rules/order?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  skip_suggested_bank_account_rule:
    endpoint: POST /bankaccounts/rules/skipsuggest?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_bank_account_rule:
    endpoint: PUT /bankaccounts/rules/{{ record.rule_id }}?organization_id={{ config.organization_id }}
    required fields: rule_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bank_account_rule:
    endpoint: DELETE /bankaccounts/rules/{{ record.rule_id }}?organization_id={{ config.organization_id }}
    required fields: rule_id
    risk: external destructive mutation in Zoho Books; approval required
  create_bank_transaction:
    endpoint: POST /banktransactions?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  categorize_bank_transaction_as_payment_refund:
    endpoint: POST /banktransactions/uncategorized/{{ record.statement_line_id }}/categorize/paymentrefunds?organization_id={{ config.organization_id }}
    required fields: statement_line_id
    risk: external mutation in Zoho Books accounting data; approval required
  categorize_as_vendor_payment_refund:
    endpoint: POST /banktransactions/uncategorized/{{ record.statement_line_id }}/categorize/vendorpaymentrefunds?organization_id={{ config.organization_id }}
    required fields: statement_line_id
    risk: external mutation in Zoho Books accounting data; approval required
  categorize_bank_transaction:
    endpoint: POST /banktransactions/uncategorized/{{ record.transaction_id }}/categorize?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  categorize_as_credit_note_refunds:
    endpoint: POST /banktransactions/uncategorized/{{ record.transaction_id }}/categorize/creditnoterefunds?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  categorize_bank_transaction_as_customer_payment:
    endpoint: POST /banktransactions/uncategorized/{{ record.transaction_id }}/categorize/customerpayments?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  categorize_bank_transaction_as_expense:
    endpoint: POST /banktransactions/uncategorized/{{ record.transaction_id }}/categorize/expenses?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  categorize_as_vendor_credit_refunds:
    endpoint: POST /banktransactions/uncategorized/{{ record.transaction_id }}/categorize/vendorcreditrefunds?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  categorize_bank_transaction_as_vendor_payment:
    endpoint: POST /banktransactions/uncategorized/{{ record.transaction_id }}/categorize/vendorpayments?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  exclude_bank_transaction:
    endpoint: POST /banktransactions/uncategorized/{{ record.transaction_id }}/exclude?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  match_bank_transaction:
    endpoint: POST /banktransactions/uncategorized/{{ record.transaction_id }}/match?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  restore_bank_transaction:
    endpoint: POST /banktransactions/uncategorized/{{ record.transaction_id }}/restore?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_bank_transaction:
    endpoint: PUT /banktransactions/{{ record.bank_transaction_id }}?organization_id={{ config.organization_id }}
    required fields: bank_transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bank_transaction:
    endpoint: DELETE /banktransactions/{{ record.bank_transaction_id }}?organization_id={{ config.organization_id }}
    required fields: bank_transaction_id
    risk: external destructive mutation in Zoho Books; approval required
  uncategorize_bank_transaction:
    endpoint: POST /banktransactions/{{ record.transaction_id }}/uncategorize?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  unmatch_bank_transaction:
    endpoint: POST /banktransactions/{{ record.transaction_id }}/unmatch?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_base_currency_adjustment:
    endpoint: POST /basecurrencyadjustment?organization_id={{ config.organization_id }}&account_ids={{ record.account_ids }}
    required fields: account_ids
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_delete_base_currency_adjustments:
    endpoint: DELETE /basecurrencyadjustment/bulkdelete?organization_id={{ config.organization_id }}&base_currency_adjustment_ids={{ record.base_currency_adjustment_ids }}
    required fields: base_currency_adjustment_ids
    risk: external destructive mutation in Zoho Books; approval required
  delete_base_currency_adjustment:
    endpoint: DELETE /basecurrencyadjustment/{{ record.base_currency_adjustment_id }}?organization_id={{ config.organization_id }}
    required fields: base_currency_adjustment_id
    risk: external destructive mutation in Zoho Books; approval required
  reevaluate_base_currency_adjustment:
    endpoint: POST /basecurrencyadjustment/{{ record.base_currency_adjustment_id }}/reevaluate?organization_id={{ config.organization_id }}
    required fields: base_currency_adjustment_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_custom_fields_in_bill:
    endpoint: PUT /bill/{{ record.bill_id }}/customfields?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_bill:
    endpoint: POST /bills?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_bill:
    endpoint: PUT /bills/{{ record.bill_id }}?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bill:
    endpoint: DELETE /bills/{{ record.bill_id }}?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external destructive mutation in Zoho Books; approval required
  update_bill_billing_address:
    endpoint: PUT /bills/{{ record.bill_id }}/address/billing?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  approve_bill:
    endpoint: POST /bills/{{ record.bill_id }}/approve?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_bill_attachment:
    endpoint: POST /bills/{{ record.bill_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bill_attachment:
    endpoint: DELETE /bills/{{ record.bill_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external destructive mutation in Zoho Books; approval required
  add_bill_comment:
    endpoint: POST /bills/{{ record.bill_id }}/comments?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bill_comment:
    endpoint: DELETE /bills/{{ record.bill_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: bill_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  apply_credits_to_bill:
    endpoint: POST /bills/{{ record.bill_id }}/credits?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_bill_payment:
    endpoint: DELETE /bills/{{ record.bill_id }}/payments/{{ record.bill_payment_id }}?organization_id={{ config.organization_id }}
    required fields: bill_id, bill_payment_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_bill_open:
    endpoint: POST /bills/{{ record.bill_id }}/status/open?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_bill_void:
    endpoint: POST /bills/{{ record.bill_id }}/status/void?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  submit_bill:
    endpoint: POST /bills/{{ record.bill_id }}/submit?organization_id={{ config.organization_id }}
    required fields: bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_chart_of_account:
    endpoint: POST /chartofaccounts?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_mark_chart_of_accounts_active:
    endpoint: POST /chartofaccounts/active?organization_id={{ config.organization_id }}&account_ids={{ record.account_ids }}
    required fields: account_ids
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_delete_chart_of_accounts:
    endpoint: DELETE /chartofaccounts/bulkdelete?organization_id={{ config.organization_id }}&account_ids={{ record.account_ids }}
    required fields: account_ids
    risk: external destructive mutation in Zoho Books; approval required
  bulk_mark_chart_of_accounts_inactive:
    endpoint: POST /chartofaccounts/inactive?organization_id={{ config.organization_id }}&account_ids={{ record.account_ids }}
    required fields: account_ids
    risk: external mutation in Zoho Books accounting data; approval required
  delete_chart_of_account_transaction:
    endpoint: DELETE /chartofaccounts/transactions/{{ record.transaction_id }}?organization_id={{ config.organization_id }}
    required fields: transaction_id
    risk: external destructive mutation in Zoho Books; approval required
  update_chart_of_account:
    endpoint: PUT /chartofaccounts/{{ record.account_id }}?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_chart_of_account:
    endpoint: DELETE /chartofaccounts/{{ record.account_id }}?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_chart_of_account_active:
    endpoint: POST /chartofaccounts/{{ record.account_id }}/active?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_chart_of_account_inactive:
    endpoint: POST /chartofaccounts/{{ record.account_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_contact_person:
    endpoint: POST /contacts/contactpersons?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_contact_person:
    endpoint: PUT /contacts/contactpersons/{{ record.contact_person_id }}?organization_id={{ config.organization_id }}
    required fields: contact_person_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact_person:
    endpoint: DELETE /contacts/contactpersons/{{ record.contact_person_id }}?organization_id={{ config.organization_id }}
    required fields: contact_person_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_contact_person_primary:
    endpoint: POST /contacts/contactpersons/{{ record.contact_person_id }}/primary?organization_id={{ config.organization_id }}
    required fields: contact_person_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_contact:
    endpoint: POST /contacts?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contacts:
    endpoint: DELETE /contacts?organization_id={{ config.organization_id }}&contact_ids={{ record.contact_ids }}
    required fields: contact_ids
    risk: external destructive mutation in Zoho Books; approval required
  create_contact_person_2:
    endpoint: POST /contacts/contactpersons?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_contact_person_2:
    endpoint: PUT /contacts/contactpersons/{{ record.contactperson_id }}?organization_id={{ config.organization_id }}
    required fields: contactperson_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact_person_2:
    endpoint: DELETE /contacts/contactpersons/{{ record.contactperson_id }}?organization_id={{ config.organization_id }}
    required fields: contactperson_id
    risk: external destructive mutation in Zoho Books; approval required
  invite_contact_person_to_portal:
    endpoint: POST /contacts/contactpersons/{{ record.contactperson_id }}/portal/invite?organization_id={{ config.organization_id }}
    required fields: contactperson_id
    risk: external mutation in Zoho Books accounting data; approval required
  resend_contact_person_portal_invite:
    endpoint: POST /contacts/contactpersons/{{ record.contactperson_id }}/portal/invite/resend?organization_id={{ config.organization_id }}
    required fields: contactperson_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_contact_person_primary_2:
    endpoint: POST /contacts/contactpersons/{{ record.contactperson_id }}/primary?organization_id={{ config.organization_id }}
    required fields: contactperson_id
    risk: external mutation in Zoho Books accounting data; approval required
  disable_contact_person_sms:
    endpoint: POST /contacts/contactpersons/{{ record.contactperson_id }}/sms/disable?organization_id={{ config.organization_id }}
    required fields: contactperson_id
    risk: external mutation in Zoho Books accounting data; approval required
  enable_contact_person_sms:
    endpoint: POST /contacts/contactpersons/{{ record.contactperson_id }}/sms/enable?organization_id={{ config.organization_id }}
    required fields: contactperson_id
    risk: external mutation in Zoho Books accounting data; approval required
  restore_contact_documents:
    endpoint: POST /contacts/documents/restore?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  assign_owner_to_contacts:
    endpoint: POST /contacts/owner?organization_id={{ config.organization_id }}&contact_ids={{ record.contact_ids }}
    required fields: contact_ids
    risk: external mutation in Zoho Books accounting data; approval required
  send_contacts_sms:
    endpoint: POST /contacts/sms?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  mark_contacts_for_1099_tracking:
    endpoint: POST /contacts/track1099?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact:
    endpoint: DELETE /contacts/{{ record.contact_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_contact_active:
    endpoint: POST /contacts/{{ record.contact_id }}/active?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_contact_address:
    endpoint: POST /contacts/{{ record.contact_id }}/address?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_contact_address:
    endpoint: PUT /contacts/{{ record.contact_id }}/address/{{ record.address_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, address_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact_address:
    endpoint: DELETE /contacts/{{ record.contact_id }}/address/{{ record.address_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, address_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_contact_address_as_billing:
    endpoint: POST /contacts/{{ record.contact_id }}/address/{{ record.address_id }}/markasbilling?organization_id={{ config.organization_id }}
    required fields: contact_id, address_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_contact_address_as_shipping:
    endpoint: POST /contacts/{{ record.contact_id }}/address/{{ record.address_id }}/markasshipping?organization_id={{ config.organization_id }}
    required fields: contact_id, address_id
    risk: external mutation in Zoho Books accounting data; approval required
  verify_contact_address_by_id:
    endpoint: POST /contacts/{{ record.contact_id }}/address/{{ record.address_id }}/verify?organization_id={{ config.organization_id }}
    required fields: contact_id, address_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_contact_attachment:
    endpoint: POST /contacts/{{ record.contact_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_contact_bank_account:
    endpoint: POST /contacts/{{ record.contact_id }}/bankaccount?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_contact_bank_account:
    endpoint: PUT /contacts/{{ record.contact_id }}/bankaccount/{{ record.bank_account_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, bank_account_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact_bank_account:
    endpoint: DELETE /contacts/{{ record.contact_id }}/bankaccount/{{ record.bank_account_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, bank_account_id
    risk: external destructive mutation in Zoho Books; approval required
  approve_contact_bank_account:
    endpoint: POST /contacts/{{ record.contact_id }}/bankaccount/{{ record.bank_account_id }}/approve?organization_id={{ config.organization_id }}
    required fields: contact_id, bank_account_id
    risk: external mutation in Zoho Books accounting data; approval required
  decline_contact_bank_account:
    endpoint: POST /contacts/{{ record.contact_id }}/bankaccount/{{ record.bank_account_id }}/decline?organization_id={{ config.organization_id }}
    required fields: contact_id, bank_account_id
    risk: external mutation in Zoho Books accounting data; approval required
  verify_contact_bank_account:
    endpoint: POST /contacts/{{ record.contact_id }}/bankaccount/{{ record.bank_account_id }}/verify?organization_id={{ config.organization_id }}
    required fields: contact_id, bank_account_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_contact_card:
    endpoint: POST /contacts/{{ record.contact_id }}/card?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_contact_card:
    endpoint: PUT /contacts/{{ record.contact_id }}/card/{{ record.card_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, card_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact_card:
    endpoint: DELETE /contacts/{{ record.contact_id }}/card/{{ record.card_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, card_id
    risk: external destructive mutation in Zoho Books; approval required
  send_contact_client_review_email:
    endpoint: POST /contacts/{{ record.contact_id }}/clientreviews/email?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_contact_comment:
    endpoint: POST /contacts/{{ record.contact_id }}/comments?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact_comment:
    endpoint: DELETE /contacts/{{ record.contact_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  update_contact_document:
    endpoint: PUT /contacts/{{ record.contact_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, document_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact_document:
    endpoint: DELETE /contacts/{{ record.contact_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, document_id
    risk: external destructive mutation in Zoho Books; approval required
  verify_contact_einvoice:
    endpoint: POST /contacts/{{ record.contact_id }}/einvoice/verify?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  email_contact:
    endpoint: POST /contacts/{{ record.contact_id }}/email?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_contact_inactive:
    endpoint: POST /contacts/{{ record.contact_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  merge_contact:
    endpoint: POST /contacts/{{ record.contact_id }}/merge?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  assign_contact_owner:
    endpoint: POST /contacts/{{ record.contact_id }}/owner?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  send_contact_payment_method_email:
    endpoint: POST /contacts/{{ record.contact_id }}/paymentmethod/email?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  disable_contact_payment_reminder:
    endpoint: POST /contacts/{{ record.contact_id }}/paymentreminder/disable?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  enable_contact_payment_reminder:
    endpoint: POST /contacts/{{ record.contact_id }}/paymentreminder/enable?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  disable_contact_portal:
    endpoint: POST /contacts/{{ record.contact_id }}/portal/disable?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  enable_contact_portal:
    endpoint: POST /contacts/{{ record.contact_id }}/portal/enable?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  send_contact_sms:
    endpoint: POST /contacts/{{ record.contact_id }}/sms?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  email_contact_statement:
    endpoint: POST /contacts/{{ record.contact_id }}/statements/email?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_contact_tags:
    endpoint: PUT /contacts/{{ record.contact_id }}/tags?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact_tag:
    endpoint: DELETE /contacts/{{ record.contact_id }}/tags/{{ record.tag_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, tag_id
    risk: external destructive mutation in Zoho Books; approval required
  add_contact_tax_info:
    endpoint: POST /contacts/{{ record.contact_id }}/taxinfo?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_contact_tax_info:
    endpoint: PUT /contacts/{{ record.contact_id }}/taxinfo/{{ record.tax_info_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, tax_info_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_contact_tax_info:
    endpoint: DELETE /contacts/{{ record.contact_id }}/taxinfo/{{ record.tax_info_id }}?organization_id={{ config.organization_id }}
    required fields: contact_id, tax_info_id
    risk: external destructive mutation in Zoho Books; approval required
  track_contact_1099:
    endpoint: POST /contacts/{{ record.contact_id }}/track1099?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_contact_trn_status:
    endpoint: POST /contacts/{{ record.contact_id }}/trnstatus?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  untrack_contact_1099:
    endpoint: POST /contacts/{{ record.contact_id }}/untrack1099?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  send_contact_vendor_statement_email:
    endpoint: POST /contacts/{{ record.contact_id }}/vendorstatements/email?organization_id={{ config.organization_id }}
    required fields: contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_credit_note:
    endpoint: POST /creditnotes?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  approve_credit_notes:
    endpoint: POST /creditnotes/approve?organization_id={{ config.organization_id }}&creditnote_ids={{ record.creditnote_ids }}
    required fields: creditnote_ids
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_credit_notes_einvoice:
    endpoint: POST /creditnotes/einvoice/cancel?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  push_credit_notes_einvoice:
    endpoint: POST /creditnotes/einvoice/push?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  submit_credit_notes:
    endpoint: POST /creditnotes/submit?organization_id={{ config.organization_id }}&creditnote_ids={{ record.creditnote_ids }}
    required fields: creditnote_ids
    risk: external mutation in Zoho Books accounting data; approval required
  update_credit_note:
    endpoint: PUT /creditnotes/{{ record.creditnote_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_credit_note:
    endpoint: DELETE /creditnotes/{{ record.creditnote_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external destructive mutation in Zoho Books; approval required
  update_credit_note_billing_address:
    endpoint: PUT /creditnotes/{{ record.creditnote_id }}/address/billing?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_credit_note_shipping_address:
    endpoint: PUT /creditnotes/{{ record.creditnote_id }}/address/shipping?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  approve_credit_note:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/approve?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  finalize_credit_note_approval:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/approve/final?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_credit_note_attachment:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_credit_note_cfdi_status:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/cfdi/status?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_credit_note_comment:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/comments?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_credit_note_comment:
    endpoint: DELETE /creditnotes/{{ record.creditnote_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  update_credit_note_custom_fields:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/customfields?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_credit_note_document:
    endpoint: PUT /creditnotes/{{ record.creditnote_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id, document_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_credit_note_document:
    endpoint: DELETE /creditnotes/{{ record.creditnote_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id, document_id
    risk: external destructive mutation in Zoho Books; approval required
  add_credit_note_digital_signature:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/dsign?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  upload_credit_note_digital_signature:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/dsign/upload?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_credit_note_einvoice:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/einvoice/cancel?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  fetch_credit_note_einvoice:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/einvoice/fetch?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  push_credit_note_einvoice:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/einvoice/push?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_credit_note_einvoice_status:
    endpoint: DELETE /creditnotes/{{ record.creditnote_id }}/einvoice/status?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_credit_note_einvoice_cancelled:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/einvoice/status/cancel?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_credit_note_einvoice_pushed:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/einvoice/status/push?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  recall_credit_note_einvoice_status:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/einvoice/status/recall?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  email_credit_note:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/email?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  apply_credit_note_to_invoice:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/invoices?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_of_credit_note:
    endpoint: DELETE /creditnotes/{{ record.creditnote_id }}/invoices/{{ record.creditnote_invoice_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id, creditnote_invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  create_credit_note_refund:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/refunds?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  push_credit_note_refund_einvoice:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/refunds/einvoice/push?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_credit_note_refund:
    endpoint: PUT /creditnotes/{{ record.creditnote_id }}/refunds/{{ record.creditnote_refund_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id, creditnote_refund_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_credit_note_refund:
    endpoint: DELETE /creditnotes/{{ record.creditnote_id }}/refunds/{{ record.creditnote_refund_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id, creditnote_refund_id
    risk: external destructive mutation in Zoho Books; approval required
  reject_credit_note:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/reject?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_credit_note_draft:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/status/draft?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_credit_note_open:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/status/open?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_credit_note_ready_to_push:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/status/readytopush?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_credit_note_void:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/status/void?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  submit_credit_note:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/submit?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  apply_credit_note_substatus:
    endpoint: POST /creditnotes/{{ record.creditnote_id }}/substatus/{{ record.substatus_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id, substatus_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_credit_note_substatus:
    endpoint: DELETE /creditnotes/{{ record.creditnote_id }}/substatus/{{ record.substatus_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id, substatus_id
    risk: external destructive mutation in Zoho Books; approval required
  update_credit_note_template:
    endpoint: PUT /creditnotes/{{ record.creditnote_id }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}
    required fields: creditnote_id, template_id
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_einvoice_credit_note:
    endpoint: POST /einvoices/creditnotes/{{ record.creditnote_id }}/cancel?organization_id={{ config.organization_id }}
    required fields: creditnote_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_currency:
    endpoint: POST /settings/currencies?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_currency:
    endpoint: PUT /settings/currencies/{{ record.currency_id }}?organization_id={{ config.organization_id }}
    required fields: currency_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_currency:
    endpoint: DELETE /settings/currencies/{{ record.currency_id }}?organization_id={{ config.organization_id }}
    required fields: currency_id
    risk: external destructive mutation in Zoho Books; approval required
  create_exchange_rate:
    endpoint: POST /settings/currencies/{{ record.currency_id }}/exchangerates?organization_id={{ config.organization_id }}
    required fields: currency_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_exchange_rate:
    endpoint: PUT /settings/currencies/{{ record.currency_id }}/exchangerates/{{ record.exchange_rate_id }}?organization_id={{ config.organization_id }}
    required fields: currency_id, exchange_rate_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_exchange_rate:
    endpoint: DELETE /settings/currencies/{{ record.currency_id }}/exchangerates/{{ record.exchange_rate_id }}?organization_id={{ config.organization_id }}
    required fields: currency_id, exchange_rate_id
    risk: external destructive mutation in Zoho Books; approval required
  create_custom_module:
    endpoint: POST /settings/modules?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_custom_module:
    endpoint: PUT /settings/modules/{{ record.module_api_name }}?organization_id={{ config.organization_id }}
    required fields: module_api_name
    risk: external mutation in Zoho Books accounting data; approval required
  delete_custom_module:
    endpoint: DELETE /settings/modules/{{ record.module_api_name }}?organization_id={{ config.organization_id }}
    required fields: module_api_name
    risk: external destructive mutation in Zoho Books; approval required
  create_custom_module_record:
    endpoint: POST /{{ record.module_name }}?organization_id={{ config.organization_id }}
    required fields: module_name
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_update_custom_module_records:
    endpoint: PUT /{{ record.module_name }}?organization_id={{ config.organization_id }}
    required fields: module_name
    risk: external mutation in Zoho Books accounting data; approval required
  delete_custom_module_records:
    endpoint: DELETE /{{ record.module_name }}?organization_id={{ config.organization_id }}
    required fields: module_name
    risk: external destructive mutation in Zoho Books; approval required
  update_custom_module_record:
    endpoint: PUT /{{ record.module_name }}/{{ record.module_id }}?organization_id={{ config.organization_id }}
    required fields: module_name, module_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_custom_module_record:
    endpoint: DELETE /{{ record.module_name }}/{{ record.module_id }}?organization_id={{ config.organization_id }}
    required fields: module_name, module_id
    risk: external destructive mutation in Zoho Books; approval required
  create_customer_debit_note:
    endpoint: POST /invoices?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_customer_debit_note:
    endpoint: PUT /invoices/{{ record.debit_note_id }}?organization_id={{ config.organization_id }}
    required fields: debit_note_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_customer_debit_note:
    endpoint: DELETE /invoices/{{ record.debit_note_id }}?organization_id={{ config.organization_id }}
    required fields: debit_note_id
    risk: external destructive mutation in Zoho Books; approval required
  update_custom_fields_in_customer_payment:
    endpoint: PUT /customerpayment/{{ record.customer_payment_id }}/customfields?organization_id={{ config.organization_id }}
    required fields: customer_payment_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_customer_payment:
    endpoint: POST /customerpayments?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_delete_customer_payments:
    endpoint: DELETE /customerpayments?organization_id={{ config.organization_id }}&payment_ids={{ record.payment_ids }}&bulk_delete={{ record.bulk_delete }}
    required fields: payment_ids, bulk_delete
    risk: external destructive mutation in Zoho Books; approval required
  create_customer_payment_refund:
    endpoint: POST /customerpayments/{{ record.customer_payment_id }}/refunds?organization_id={{ config.organization_id }}
    required fields: customer_payment_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_customer_payment_refund:
    endpoint: PUT /customerpayments/{{ record.customer_payment_id }}/refunds/{{ record.refund_id }}?organization_id={{ config.organization_id }}
    required fields: customer_payment_id, refund_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_customer_payment_refund:
    endpoint: DELETE /customerpayments/{{ record.customer_payment_id }}/refunds/{{ record.refund_id }}?organization_id={{ config.organization_id }}
    required fields: customer_payment_id, refund_id
    risk: external destructive mutation in Zoho Books; approval required
  update_customer_payment:
    endpoint: PUT /customerpayments/{{ record.payment_id }}?organization_id={{ config.organization_id }}
    required fields: payment_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_customer_payment:
    endpoint: DELETE /customerpayments/{{ record.payment_id }}?organization_id={{ config.organization_id }}
    required fields: payment_id
    risk: external destructive mutation in Zoho Books; approval required
  create_delivery_challan:
    endpoint: POST /deliverychallans?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  return_delivery_challans:
    endpoint: PUT /deliverychallans/return?organization_id={{ config.organization_id }}&deliverychallan_ids={{ record.deliverychallan_ids }}
    required fields: deliverychallan_ids
    risk: external mutation in Zoho Books accounting data; approval required
  undo_return_delivery_challans:
    endpoint: PUT /deliverychallans/undo/return?organization_id={{ config.organization_id }}&deliverychallan_ids={{ record.deliverychallan_ids }}
    required fields: deliverychallan_ids
    risk: external mutation in Zoho Books accounting data; approval required
  update_delivery_challan:
    endpoint: PUT /deliverychallans/{{ record.deliverychallan_id }}?organization_id={{ config.organization_id }}
    required fields: deliverychallan_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_delivery_challan:
    endpoint: DELETE /deliverychallans/{{ record.deliverychallan_id }}?organization_id={{ config.organization_id }}
    required fields: deliverychallan_id
    risk: external destructive mutation in Zoho Books; approval required
  update_delivery_challan_shipping_address:
    endpoint: PUT /deliverychallans/{{ record.deliverychallan_id }}/address/shipping?organization_id={{ config.organization_id }}
    required fields: deliverychallan_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_delivery_challan_attachment:
    endpoint: DELETE /deliverychallans/{{ record.deliverychallan_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: deliverychallan_id, document_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_delivery_challan_as_delivered:
    endpoint: POST /deliverychallans/{{ record.deliverychallan_id }}/status/delivered?organization_id={{ config.organization_id }}
    required fields: deliverychallan_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_delivery_challan_as_open:
    endpoint: POST /deliverychallans/{{ record.deliverychallan_id }}/status/open?organization_id={{ config.organization_id }}
    required fields: deliverychallan_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_delivery_challan_as_returned:
    endpoint: POST /deliverychallans/{{ record.deliverychallan_id }}/status/returned?organization_id={{ config.organization_id }}
    required fields: deliverychallan_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_delivery_challan_as_undelivered:
    endpoint: POST /deliverychallans/{{ record.deliverychallan_id }}/status/undelivered?organization_id={{ config.organization_id }}
    required fields: deliverychallan_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_delivery_challan_template:
    endpoint: PUT /deliverychallans/{{ record.deliverychallan_id }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}
    required fields: deliverychallan_id, template_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_custom_fields_in_estimate:
    endpoint: PUT /estimate/{{ record.estimate_id }}/customfields?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_estimate:
    endpoint: POST /estimates?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  email_multiple_estimates:
    endpoint: POST /estimates/email?organization_id={{ config.organization_id }}&estimate_ids={{ record.estimate_ids }}
    required fields: estimate_ids
    risk: external mutation in Zoho Books accounting data; approval required
  update_estimate:
    endpoint: PUT /estimates/{{ record.estimate_id }}?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_estimate:
    endpoint: DELETE /estimates/{{ record.estimate_id }}?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external destructive mutation in Zoho Books; approval required
  update_estimate_billing_address:
    endpoint: PUT /estimates/{{ record.estimate_id }}/address/billing?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_estimate_shipping_address:
    endpoint: PUT /estimates/{{ record.estimate_id }}/address/shipping?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  approve_estimate:
    endpoint: POST /estimates/{{ record.estimate_id }}/approve?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_estimate_comment:
    endpoint: POST /estimates/{{ record.estimate_id }}/comments?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_estimate_comment:
    endpoint: PUT /estimates/{{ record.estimate_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: estimate_id, comment_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_estimate_comment:
    endpoint: DELETE /estimates/{{ record.estimate_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: estimate_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  email_estimate:
    endpoint: POST /estimates/{{ record.estimate_id }}/email?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_estimate_accepted:
    endpoint: POST /estimates/{{ record.estimate_id }}/status/accepted?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_estimate_declined:
    endpoint: POST /estimates/{{ record.estimate_id }}/status/declined?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_estimate_sent:
    endpoint: POST /estimates/{{ record.estimate_id }}/status/sent?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  submit_estimate:
    endpoint: POST /estimates/{{ record.estimate_id }}/submit?organization_id={{ config.organization_id }}
    required fields: estimate_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_estimate_template:
    endpoint: PUT /estimates/{{ record.estimate_id }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}
    required fields: estimate_id, template_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_employee:
    endpoint: DELETE /employee/{{ record.employee_id }}?organization_id={{ config.organization_id }}
    required fields: employee_id
    risk: external destructive mutation in Zoho Books; approval required
  create_employee:
    endpoint: POST /employees?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  create_expense:
    endpoint: POST /expenses?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_expense:
    endpoint: PUT /expenses/{{ record.expense_id }}?organization_id={{ config.organization_id }}
    required fields: expense_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_expense:
    endpoint: DELETE /expenses/{{ record.expense_id }}?organization_id={{ config.organization_id }}
    required fields: expense_id
    risk: external destructive mutation in Zoho Books; approval required
  create_expense_receipt:
    endpoint: POST /expenses/{{ record.expense_id }}/receipt?organization_id={{ config.organization_id }}
    required fields: expense_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_expense_receipt:
    endpoint: DELETE /expenses/{{ record.expense_id }}/receipt?organization_id={{ config.organization_id }}
    required fields: expense_id
    risk: external destructive mutation in Zoho Books; approval required
  create_fixed_asset:
    endpoint: POST /fixedassets?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_fixed_asset:
    endpoint: PUT /fixedassets/{{ record.fixed_asset_id }}?organization_id={{ config.organization_id }}
    required fields: fixed_asset_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_fixed_asset:
    endpoint: DELETE /fixedassets/{{ record.fixed_asset_id }}?organization_id={{ config.organization_id }}
    required fields: fixed_asset_id
    risk: external destructive mutation in Zoho Books; approval required
  create_fixed_asset_comment:
    endpoint: POST /fixedassets/{{ record.fixed_asset_id }}/comments?organization_id={{ config.organization_id }}
    required fields: fixed_asset_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_fixed_asset_comment:
    endpoint: DELETE /fixedassets/{{ record.fixed_asset_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: fixed_asset_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  sell_fixed_asset:
    endpoint: POST /fixedassets/{{ record.fixed_asset_id }}/sell?organization_id={{ config.organization_id }}
    required fields: fixed_asset_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_fixed_asset_active:
    endpoint: POST /fixedassets/{{ record.fixed_asset_id }}/status/active?organization_id={{ config.organization_id }}
    required fields: fixed_asset_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_fixed_asset_cancel:
    endpoint: POST /fixedassets/{{ record.fixed_asset_id }}/status/cancel?organization_id={{ config.organization_id }}
    required fields: fixed_asset_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_fixed_asset_draft:
    endpoint: POST /fixedassets/{{ record.fixed_asset_id }}/status/draft?organization_id={{ config.organization_id }}
    required fields: fixed_asset_id
    risk: external mutation in Zoho Books accounting data; approval required
  write_off_fixed_asset:
    endpoint: POST /fixedassets/{{ record.fixed_asset_id }}/writeoff?organization_id={{ config.organization_id }}
    required fields: fixed_asset_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_fixed_asset_type:
    endpoint: POST /fixedassettypes?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_fixed_asset_type:
    endpoint: PUT /fixedassettypes/{{ record.fixed_asset_type_id }}?organization_id={{ config.organization_id }}
    required fields: fixed_asset_type_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_fixed_asset_type:
    endpoint: DELETE /fixedassettypes/{{ record.fixed_asset_type_id }}?organization_id={{ config.organization_id }}
    required fields: fixed_asset_type_id
    risk: external destructive mutation in Zoho Books; approval required
  import_customer_using_crm_account_id:
    endpoint: POST /crm/account/{{ record.crm_account_id }}/import?organization_id={{ config.organization_id }}
    required fields: crm_account_id
    risk: external mutation in Zoho Books accounting data; approval required
  import_customer_using_crm_contact_id:
    endpoint: POST /crm/contact/{{ record.crm_contact_id }}/import?organization_id={{ config.organization_id }}
    required fields: crm_contact_id
    risk: external mutation in Zoho Books accounting data; approval required
  import_item_using_crm_product_id:
    endpoint: POST /crm/item/{{ record.crm_product_id }}/import?organization_id={{ config.organization_id }}
    required fields: crm_product_id
    risk: external mutation in Zoho Books accounting data; approval required
  import_vendor_using_crm_vendor_id:
    endpoint: POST /crm/vendor/{{ record.crm_vendor_id }}/import?organization_id={{ config.organization_id }}
    required fields: crm_vendor_id
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_einvoice_invoice:
    endpoint: POST /einvoices/invoices/{{ record.invoice_id }}/cancel?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_custom_fields_in_invoice:
    endpoint: PUT /invoice/{{ record.invoice_id }}/customfields?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_invoice:
    endpoint: POST /invoices?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoices:
    endpoint: DELETE /invoices?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external destructive mutation in Zoho Books; approval required
  approve_invoices:
    endpoint: POST /invoices/approve?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  preview_invoice_coupons:
    endpoint: POST /invoices/coupons/preview?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_invoices_einvoice:
    endpoint: POST /invoices/einvoice/cancel?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  push_invoices_einvoice:
    endpoint: POST /invoices/einvoice/push?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  email_invoices:
    endpoint: POST /invoices/email?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_expense_receipt:
    endpoint: DELETE /invoices/expenses/{{ record.expense_id }}/receipt?organization_id={{ config.organization_id }}
    required fields: expense_id
    risk: external destructive mutation in Zoho Books; approval required
  create_invoices_from_estimates:
    endpoint: POST /invoices/fromestimates?organization_id={{ config.organization_id }}&estimate_ids={{ record.estimate_ids }}
    required fields: estimate_ids
    risk: external mutation in Zoho Books accounting data; approval required
  create_invoices_from_projects:
    endpoint: POST /invoices/fromprojects?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  create_invoice_from_salesorder:
    endpoint: POST /invoices/fromsalesorder?organization_id={{ config.organization_id }}&salesorder_id={{ record.salesorder_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  map_invoice_with_salesorder:
    endpoint: POST /invoices/mapwithorder?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  mark_invoices_shipped:
    endpoint: POST /invoices/markasshipped?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_invoice_reminder:
    endpoint: POST /invoices/paymentreminder?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  mark_invoices_sent:
    endpoint: POST /invoices/status/sent?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  void_invoices:
    endpoint: POST /invoices/status/void?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  submit_invoices:
    endpoint: POST /invoices/submit?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  unmap_invoices_from_salesorders:
    endpoint: PUT /invoices/unmap/salesorders?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  unship_invoices:
    endpoint: POST /invoices/unship?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  write_off_invoices:
    endpoint: POST /invoices/writeoff?organization_id={{ config.organization_id }}&invoice_ids={{ record.invoice_ids }}
    required fields: invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  update_invoice:
    endpoint: PUT /invoices/{{ record.invoice_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice:
    endpoint: DELETE /invoices/{{ record.invoice_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  update_invoice_billing_address:
    endpoint: PUT /invoices/{{ record.invoice_id }}/address/billing?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_invoice_shipping_address:
    endpoint: PUT /invoices/{{ record.invoice_id }}/address/shipping?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_invoice_advanced_tracking_details:
    endpoint: PUT /invoices/{{ record.invoice_id }}/advancedtrackingdetails?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  approve_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/approve?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  finalize_invoice_approval:
    endpoint: POST /invoices/{{ record.invoice_id }}/approve/final?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_invoice_attachment:
    endpoint: POST /invoices/{{ record.invoice_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_invoice_attachment_preference:
    endpoint: PUT /invoices/{{ record.invoice_id }}/attachment?organization_id={{ config.organization_id }}&can_send_in_mail={{ record.can_send_in_mail }}
    required fields: invoice_id, can_send_in_mail
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_attachment:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  update_invoice_cfdi_status:
    endpoint: POST /invoices/{{ record.invoice_id }}/cfdi/status?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_invoice_comment:
    endpoint: POST /invoices/{{ record.invoice_id }}/comments?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_invoice_comment:
    endpoint: PUT /invoices/{{ record.invoice_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, comment_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_comment:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  apply_credits_to_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/credits?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_applied_credit:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/creditsapplied/{{ record.creditnotes_invoice_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, creditnotes_invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  add_invoice_document:
    endpoint: POST /invoices/{{ record.invoice_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, document_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_document:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, document_id
    risk: external destructive mutation in Zoho Books; approval required
  upload_invoice_document:
    endpoint: POST /invoices/{{ record.invoice_id }}/documents/{{ record.document_id }}/upload?organization_id={{ config.organization_id }}
    required fields: invoice_id, document_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_invoice_digital_signature:
    endpoint: POST /invoices/{{ record.invoice_id }}/dsign?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  upload_invoice_digital_signature:
    endpoint: POST /invoices/{{ record.invoice_id }}/dsign/upload?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_invoice_einvoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/einvoice/cancel?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  fetch_invoice_einvoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/einvoice/fetch?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_invoice_einvoice_payment_status:
    endpoint: PUT /invoices/{{ record.invoice_id }}/einvoice/paymentstatus?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  push_invoice_einvoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/einvoice/push?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_einvoice_status:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/einvoice/status?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_invoice_einvoice_cancelled:
    endpoint: POST /invoices/{{ record.invoice_id }}/einvoice/status/cancel?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_invoice_einvoice_pushed:
    endpoint: POST /invoices/{{ record.invoice_id }}/einvoice/status/push?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  recall_invoice_einvoice_status:
    endpoint: POST /invoices/{{ record.invoice_id }}/einvoice/status/recall?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  email_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/email?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  schedule_invoice_email:
    endpoint: POST /invoices/{{ record.invoice_id }}/email/schedule?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_scheduled_invoice_email:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/email/schedule?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  force_pay_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/forcepay?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_line_item:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/lineitems/{{ record.line_item_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, line_item_id
    risk: external destructive mutation in Zoho Books; approval required
  mail_invoice_pdf:
    endpoint: POST /invoices/{{ record.invoice_id }}/mailpdf?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_invoice_metadata:
    endpoint: PUT /invoices/{{ record.invoice_id }}/metadata/{{ record.metadata_name }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, metadata_name
    risk: external mutation in Zoho Books accounting data; approval required
  create_invoice_asynchronous_online_payment:
    endpoint: POST /invoices/{{ record.invoice_id }}/onlinepayments/asynchronous?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_invoice_online_payment_bank_account:
    endpoint: POST /invoices/{{ record.invoice_id }}/onlinepayments/bankaccount?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_invoice_synchronous_online_payment:
    endpoint: POST /invoices/{{ record.invoice_id }}/onlinepayments/synchronous?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  remind_customer_for_invoice_payment:
    endpoint: POST /invoices/{{ record.invoice_id }}/paymentreminder?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  disable_invoice_payment_reminder:
    endpoint: POST /invoices/{{ record.invoice_id }}/paymentreminder/disable?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  enable_invoice_payment_reminder:
    endpoint: POST /invoices/{{ record.invoice_id }}/paymentreminder/enable?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_payment:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/payments/{{ record.invoice_payment_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, invoice_payment_id
    risk: external destructive mutation in Zoho Books; approval required
  apply_pricebook_to_invoice:
    endpoint: PUT /invoices/{{ record.invoice_id }}/pricebooks/{{ record.pricebook_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, pricebook_id
    risk: external mutation in Zoho Books accounting data; approval required
  reject_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/reject?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  send_invoice_dunning_notifications:
    endpoint: POST /invoices/{{ record.invoice_id }}/senddunningnotifications?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  send_invoice_retry_sms:
    endpoint: POST /invoices/{{ record.invoice_id }}/sendretrysms?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  send_invoice_sms:
    endpoint: POST /invoices/{{ record.invoice_id }}/sms?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  send_invoice_via_snail_mail:
    endpoint: POST /invoices/{{ record.invoice_id }}/snailmail?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/status/cancel?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_invoice_draft:
    endpoint: POST /invoices/{{ record.invoice_id }}/status/draft?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_invoice_ready_to_push:
    endpoint: POST /invoices/{{ record.invoice_id }}/status/readytopush?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_invoice_sent:
    endpoint: POST /invoices/{{ record.invoice_id }}/status/sent?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_invoice_void:
    endpoint: POST /invoices/{{ record.invoice_id }}/status/void?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  submit_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/submit?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  apply_invoice_substatus:
    endpoint: POST /invoices/{{ record.invoice_id }}/substatus/{{ record.substatus_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, substatus_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_invoice_substatus:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/substatus/{{ record.substatus_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, substatus_id
    risk: external destructive mutation in Zoho Books; approval required
  update_invoice_template:
    endpoint: PUT /invoices/{{ record.invoice_id }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}
    required fields: invoice_id, template_id
    risk: external mutation in Zoho Books accounting data; approval required
  write_off_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/writeoff?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_write_off_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/writeoff/cancel?organization_id={{ config.organization_id }}
    required fields: invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_custom_fields_in_item:
    endpoint: PUT /item/{{ record.item_id }}/customfields?organization_id={{ config.organization_id }}
    required fields: item_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_item:
    endpoint: POST /items?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  add_items_to_portal:
    endpoint: POST /items/addtoportal?organization_id={{ config.organization_id }}&item_ids={{ record.item_ids }}
    required fields: item_ids
    risk: external mutation in Zoho Books accounting data; approval required
  update_item:
    endpoint: PUT /items/{{ record.item_id }}?organization_id={{ config.organization_id }}
    required fields: item_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_item:
    endpoint: DELETE /items/{{ record.item_id }}?organization_id={{ config.organization_id }}
    required fields: item_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_item_active:
    endpoint: POST /items/{{ record.item_id }}/active?organization_id={{ config.organization_id }}
    required fields: item_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_item_to_portal:
    endpoint: POST /items/{{ record.item_id }}/addtoportal?organization_id={{ config.organization_id }}
    required fields: item_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_item_inactive:
    endpoint: POST /items/{{ record.item_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: item_id
    risk: external mutation in Zoho Books accounting data; approval required
  remove_item_from_portal:
    endpoint: POST /items/{{ record.item_id }}/removefromportal?organization_id={{ config.organization_id }}
    required fields: item_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_journal:
    endpoint: POST /journals?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_approve_journals:
    endpoint: POST /journals/approve?organization_id={{ config.organization_id }}&journal_ids={{ record.journal_ids }}
    required fields: journal_ids
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_delete_journals:
    endpoint: DELETE /journals/bulkdelete?organization_id={{ config.organization_id }}&journal_ids={{ record.journal_ids }}
    required fields: journal_ids
    risk: external destructive mutation in Zoho Books; approval required
  bulk_publish_journals:
    endpoint: POST /journals/status/publish?organization_id={{ config.organization_id }}&journal_ids={{ record.journal_ids }}
    required fields: journal_ids
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_submit_journals:
    endpoint: POST /journals/submit?organization_id={{ config.organization_id }}&journal_ids={{ record.journal_ids }}
    required fields: journal_ids
    risk: external mutation in Zoho Books accounting data; approval required
  update_journal:
    endpoint: PUT /journals/{{ record.journal_id }}?organization_id={{ config.organization_id }}
    required fields: journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_journal:
    endpoint: DELETE /journals/{{ record.journal_id }}?organization_id={{ config.organization_id }}
    required fields: journal_id
    risk: external destructive mutation in Zoho Books; approval required
  approve_journal:
    endpoint: POST /journals/{{ record.journal_id }}/approve?organization_id={{ config.organization_id }}
    required fields: journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_journal_attachment:
    endpoint: POST /journals/{{ record.journal_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_journal_comment:
    endpoint: POST /journals/{{ record.journal_id }}/comments?organization_id={{ config.organization_id }}
    required fields: journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_journal_comment:
    endpoint: DELETE /journals/{{ record.journal_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: journal_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  apply_journal_credits_to_bills:
    endpoint: POST /journals/{{ record.journal_id }}/credits/{{ record.journal_line_id }}/bills?organization_id={{ config.organization_id }}
    required fields: journal_id, journal_line_id
    risk: external mutation in Zoho Books accounting data; approval required
  apply_journal_credits_to_invoices:
    endpoint: POST /journals/{{ record.journal_id }}/credits/{{ record.journal_line_id }}/invoices?organization_id={{ config.organization_id }}
    required fields: journal_id, journal_line_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_journal_credits_payables:
    endpoint: DELETE /journals/{{ record.journal_id }}/credits/{{ record.journal_line_id }}/payables?organization_id={{ config.organization_id }}
    required fields: journal_id, journal_line_id
    risk: external destructive mutation in Zoho Books; approval required
  delete_journal_credits_receivables:
    endpoint: DELETE /journals/{{ record.journal_id }}/credits/{{ record.journal_line_id }}/receivables?organization_id={{ config.organization_id }}
    required fields: journal_id, journal_line_id
    risk: external destructive mutation in Zoho Books; approval required
  reject_journal:
    endpoint: POST /journals/{{ record.journal_id }}/reject?organization_id={{ config.organization_id }}
    required fields: journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  reverse_journal:
    endpoint: POST /journals/{{ record.journal_id }}/reverse?organization_id={{ config.organization_id }}&reversal_date={{ record.reversal_date }}&is_reversal_scheduled={{ record.is_reversal_scheduled }}
    required fields: journal_id, reversal_date, is_reversal_scheduled
    risk: external mutation in Zoho Books accounting data; approval required
  mark_journal_published:
    endpoint: POST /journals/{{ record.journal_id }}/status/publish?organization_id={{ config.organization_id }}
    required fields: journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  submit_journal_for_approval:
    endpoint: POST /journals/{{ record.journal_id }}/submit?organization_id={{ config.organization_id }}
    required fields: journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_recurring_journal:
    endpoint: POST /recurringjournals?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_recurring_journal:
    endpoint: PUT /recurringjournals/{{ record.recurring_journal_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_recurring_journal:
    endpoint: DELETE /recurringjournals/{{ record.recurring_journal_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_journal_id
    risk: external destructive mutation in Zoho Books; approval required
  resume_recurring_journal:
    endpoint: POST /recurringjournals/{{ record.recurring_journal_id }}/status/resume?organization_id={{ config.organization_id }}
    required fields: recurring_journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  stop_recurring_journal:
    endpoint: POST /recurringjournals/{{ record.recurring_journal_id }}/status/stop?organization_id={{ config.organization_id }}
    required fields: recurring_journal_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_location:
    endpoint: POST /locations?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_location:
    endpoint: PUT /locations/{{ record.location_id }}?organization_id={{ config.organization_id }}
    required fields: location_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_location:
    endpoint: DELETE /locations/{{ record.location_id }}?organization_id={{ config.organization_id }}
    required fields: location_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_location_active:
    endpoint: POST /locations/{{ record.location_id }}/active?organization_id={{ config.organization_id }}
    required fields: location_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_location_inactive:
    endpoint: POST /locations/{{ record.location_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: location_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_location_primary:
    endpoint: POST /locations/{{ record.location_id }}/markasprimary?organization_id={{ config.organization_id }}
    required fields: location_id
    risk: external mutation in Zoho Books accounting data; approval required
  enable_locations:
    endpoint: POST /settings/locations/enable?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  writeoff_opening_balance:
    endpoint: POST /openingbalances/{{ record.opening_balance_id }}/writeoff?organization_id={{ config.organization_id }}
    required fields: opening_balance_id
    risk: external mutation in Zoho Books accounting data; approval required
  cancel_writeoff_opening_balance:
    endpoint: POST /openingbalances/{{ record.opening_balance_id }}/writeoff/cancel?organization_id={{ config.organization_id }}
    required fields: opening_balance_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_opening_balance:
    endpoint: POST /settings/openingbalances?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_opening_balance:
    endpoint: PUT /settings/openingbalances?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  delete_opening_balance:
    endpoint: DELETE /settings/openingbalances?organization_id={{ config.organization_id }}
    risk: external destructive mutation in Zoho Books; approval required
  create_organization:
    endpoint: POST /organizations
    risk: external mutation in Zoho Books accounting data; approval required
  create_organization_address:
    endpoint: POST /organizations/address
    risk: external mutation in Zoho Books accounting data; approval required
  update_organization_address:
    endpoint: PUT /organizations/address/{{ record.address_id }}
    required fields: address_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_organization_address:
    endpoint: DELETE /organizations/address/{{ record.address_id }}
    required fields: address_id
    risk: external destructive mutation in Zoho Books; approval required
  update_organization:
    endpoint: PUT /organizations/{{ record.organization_id }}
    required fields: organization_id
    risk: external mutation in Zoho Books accounting data; approval required
  copy_organization_settings:
    endpoint: POST /organizations/{{ record.organization_id }}/copysettings?settings_to_copy={{ record.settings_to_copy }}
    required fields: organization_id, settings_to_copy
    risk: external mutation in Zoho Books accounting data; approval required
  downgrade_organization_to_invoice:
    endpoint: POST /organizations/{{ record.organization_id }}/downgradetoinvoice
    required fields: organization_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_organization_inactive:
    endpoint: POST /organizations/{{ record.organization_id }}/inactive
    required fields: organization_id
    risk: external mutation in Zoho Books accounting data; approval required
  upgrade_organization_to_books:
    endpoint: POST /organizations/{{ record.organization_id }}/upgradetobooks
    required fields: organization_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_pricebook:
    endpoint: POST /pricebooks?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_pricebook:
    endpoint: PUT /pricebooks/{{ record.pricebook_id }}?organization_id={{ config.organization_id }}
    required fields: pricebook_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_pricebook:
    endpoint: DELETE /pricebooks/{{ record.pricebook_id }}?organization_id={{ config.organization_id }}
    required fields: pricebook_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_pricebook_active:
    endpoint: POST /pricebooks/{{ record.pricebook_id }}/active?organization_id={{ config.organization_id }}
    required fields: pricebook_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_pricebook_inactive:
    endpoint: POST /pricebooks/{{ record.pricebook_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: pricebook_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_project:
    endpoint: POST /projects?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_project:
    endpoint: PUT /projects/{{ record.project_id }}?organization_id={{ config.organization_id }}
    required fields: project_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_project:
    endpoint: DELETE /projects/{{ record.project_id }}?organization_id={{ config.organization_id }}
    required fields: project_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_project_active:
    endpoint: POST /projects/{{ record.project_id }}/active?organization_id={{ config.organization_id }}
    required fields: project_id
    risk: external mutation in Zoho Books accounting data; approval required
  clone_project:
    endpoint: POST /projects/{{ record.project_id }}/clone?organization_id={{ config.organization_id }}
    required fields: project_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_project_comment:
    endpoint: POST /projects/{{ record.project_id }}/comments?organization_id={{ config.organization_id }}
    required fields: project_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_project_comment:
    endpoint: DELETE /projects/{{ record.project_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: project_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_project_inactive:
    endpoint: POST /projects/{{ record.project_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: project_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_project_task:
    endpoint: POST /projects/{{ record.project_id }}/tasks?organization_id={{ config.organization_id }}
    required fields: project_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_project_task:
    endpoint: PUT /projects/{{ record.project_id }}/tasks/{{ record.task_id }}?organization_id={{ config.organization_id }}
    required fields: project_id, task_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_project_task:
    endpoint: DELETE /projects/{{ record.project_id }}/tasks/{{ record.task_id }}?organization_id={{ config.organization_id }}
    required fields: project_id, task_id
    risk: external destructive mutation in Zoho Books; approval required
  add_project_user:
    endpoint: POST /projects/{{ record.project_id }}/users?organization_id={{ config.organization_id }}
    required fields: project_id
    risk: external mutation in Zoho Books accounting data; approval required
  invite_project_user:
    endpoint: POST /projects/{{ record.project_id }}/users/invite?organization_id={{ config.organization_id }}
    required fields: project_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_project_user:
    endpoint: PUT /projects/{{ record.project_id }}/users/{{ record.user_id }}?organization_id={{ config.organization_id }}
    required fields: project_id, user_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_project_user:
    endpoint: DELETE /projects/{{ record.project_id }}/users/{{ record.user_id }}?organization_id={{ config.organization_id }}
    required fields: project_id, user_id
    risk: external destructive mutation in Zoho Books; approval required
  update_custom_fields_in_purchase_order:
    endpoint: PUT /purchaseorder/{{ record.purchaseorder_id }}/customfields?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_purchase_order:
    endpoint: POST /purchaseorders?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_purchase_order:
    endpoint: PUT /purchaseorders/{{ record.purchaseorder_id }}?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_purchase_order:
    endpoint: DELETE /purchaseorders/{{ record.purchaseorder_id }}?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external destructive mutation in Zoho Books; approval required
  update_purchase_order_billing_address:
    endpoint: PUT /purchaseorders/{{ record.purchaseorder_id }}/address/billing?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  approve_purchase_order:
    endpoint: POST /purchaseorders/{{ record.purchaseorder_id }}/approve?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_purchase_order_attachment:
    endpoint: POST /purchaseorders/{{ record.purchaseorder_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_purchase_order_attachment:
    endpoint: PUT /purchaseorders/{{ record.purchaseorder_id }}/attachment?organization_id={{ config.organization_id }}&can_send_in_mail={{ record.can_send_in_mail }}
    required fields: purchaseorder_id, can_send_in_mail
    risk: external mutation in Zoho Books accounting data; approval required
  delete_purchase_order_attachment:
    endpoint: DELETE /purchaseorders/{{ record.purchaseorder_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external destructive mutation in Zoho Books; approval required
  add_purchase_order_comment:
    endpoint: POST /purchaseorders/{{ record.purchaseorder_id }}/comments?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_purchase_order_comment:
    endpoint: PUT /purchaseorders/{{ record.purchaseorder_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id, comment_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_purchase_order_comment:
    endpoint: DELETE /purchaseorders/{{ record.purchaseorder_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  email_purchase_order:
    endpoint: POST /purchaseorders/{{ record.purchaseorder_id }}/email?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  reject_purchase_orders:
    endpoint: POST /purchaseorders/{{ record.purchaseorder_id }}/reject?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_purchase_order_billed:
    endpoint: POST /purchaseorders/{{ record.purchaseorder_id }}/status/billed?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_purchase_order_cancelled:
    endpoint: POST /purchaseorders/{{ record.purchaseorder_id }}/status/cancelled?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_purchase_order_open:
    endpoint: POST /purchaseorders/{{ record.purchaseorder_id }}/status/open?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  submit_purchase_order:
    endpoint: POST /purchaseorders/{{ record.purchaseorder_id }}/submit?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_purchase_order_template:
    endpoint: PUT /purchaseorders/{{ record.purchaseorder_id }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}
    required fields: purchaseorder_id, template_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_recurring_bill:
    endpoint: DELETE /recurring_bills/{{ record.recurring_bill_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_bill_id
    risk: external destructive mutation in Zoho Books; approval required
  create_recurring_bill:
    endpoint: POST /recurringbills?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_recurring_bill:
    endpoint: PUT /recurringbills/{{ record.recurring_bill_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  resume_recurring_bill:
    endpoint: POST /recurringbills/{{ record.recurring_bill_id }}/status/resume?organization_id={{ config.organization_id }}
    required fields: recurring_bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  stop_recurring_bill:
    endpoint: POST /recurringbills/{{ record.recurring_bill_id }}/status/stop?organization_id={{ config.organization_id }}
    required fields: recurring_bill_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_recurring_expense:
    endpoint: POST /recurringexpenses?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_recurring_expense:
    endpoint: PUT /recurringexpenses/{{ record.recurring_expense_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_expense_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_recurring_expense:
    endpoint: DELETE /recurringexpenses/{{ record.recurring_expense_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_expense_id
    risk: external destructive mutation in Zoho Books; approval required
  resume_recurring_expense:
    endpoint: POST /recurringexpenses/{{ record.recurring_expense_id }}/status/resume?organization_id={{ config.organization_id }}
    required fields: recurring_expense_id
    risk: external mutation in Zoho Books accounting data; approval required
  stop_recurring_expense:
    endpoint: POST /recurringexpenses/{{ record.recurring_expense_id }}/status/stop?organization_id={{ config.organization_id }}
    required fields: recurring_expense_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_recurring_invoice:
    endpoint: POST /recurringinvoices?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  resume_recurring_invoices:
    endpoint: POST /recurringinvoices/status/resume?organization_id={{ config.organization_id }}&recurring_invoice_ids={{ record.recurring_invoice_ids }}
    required fields: recurring_invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  stop_recurring_invoices:
    endpoint: POST /recurringinvoices/status/stop?organization_id={{ config.organization_id }}&recurring_invoice_ids={{ record.recurring_invoice_ids }}
    required fields: recurring_invoice_ids
    risk: external mutation in Zoho Books accounting data; approval required
  update_recurring_invoice:
    endpoint: PUT /recurringinvoices/{{ record.recurring_invoice_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_recurring_invoice:
    endpoint: DELETE /recurringinvoices/{{ record.recurring_invoice_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  disable_recurring_invoice_autobill:
    endpoint: POST /recurringinvoices/{{ record.recurring_invoice_id }}/autobill/disable?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  enable_recurring_invoice_autobill:
    endpoint: POST /recurringinvoices/{{ record.recurring_invoice_id }}/autobill/enable?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_recurring_invoice_bank_account:
    endpoint: DELETE /recurringinvoices/{{ record.recurring_invoice_id }}/bankaccount?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  associate_recurring_invoice_bank_account:
    endpoint: POST /recurringinvoices/{{ record.recurring_invoice_id }}/bankaccount/{{ record.account_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id, account_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_recurring_invoice_card:
    endpoint: DELETE /recurringinvoices/{{ record.recurring_invoice_id }}/card?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  associate_recurring_invoice_card:
    endpoint: POST /recurringinvoices/{{ record.recurring_invoice_id }}/card/{{ record.card_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id, card_id
    risk: external mutation in Zoho Books accounting data; approval required
  resume_recurring_invoice:
    endpoint: POST /recurringinvoices/{{ record.recurring_invoice_id }}/status/resume?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  stop_recurring_invoice:
    endpoint: POST /recurringinvoices/{{ record.recurring_invoice_id }}/status/stop?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_recurring_invoice_template:
    endpoint: PUT /recurringinvoices/{{ record.recurring_invoice_id }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}
    required fields: recurring_invoice_id, template_id
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_delete_register_transactions:
    endpoint: PUT /registers/{{ record.account_id }}/transactions/bulkdelete?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_update_register_transactions:
    endpoint: PUT /registers/{{ record.account_id }}/transactions/bulkupdate?organization_id={{ config.organization_id }}
    required fields: account_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_tag:
    endpoint: POST /reportingtags?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  reorder_tags:
    endpoint: PUT /reportingtags/reorder?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  mark_default_option:
    endpoint: POST /reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id }}&default_option_id={{ record.default_option_id }}
    required fields: tag_id, default_option_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_tag:
    endpoint: PUT /reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_tag:
    endpoint: DELETE /reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: external destructive mutation in Zoho Books; approval required
  active_tag:
    endpoint: POST /reportingtags/{{ record.tag_id }}/active?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_tag_criteria:
    endpoint: PUT /reportingtags/{{ record.tag_id }}/criteria?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: external mutation in Zoho Books accounting data; approval required
  inactive_tag:
    endpoint: POST /reportingtags/{{ record.tag_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: external mutation in Zoho Books accounting data; approval required
  active_tag_option:
    endpoint: POST /reportingtags/{{ record.tag_id }}/option/(\d+)/active?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: external mutation in Zoho Books accounting data; approval required
  inactive_tag_option:
    endpoint: POST /reportingtags/{{ record.tag_id }}/option/(\d+)/inactive?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_tag_options:
    endpoint: PUT /reportingtags/{{ record.tag_id }}/options?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_retainer_invoice:
    endpoint: POST /retainerinvoices?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_retainer_invoice:
    endpoint: PUT /retainerinvoices/{{ record.retainerinvoice_id }}?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_retainer_invoice:
    endpoint: DELETE /retainerinvoices/{{ record.retainerinvoice_id }}?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external destructive mutation in Zoho Books; approval required
  update_retainer_invoice_billing_address:
    endpoint: PUT /retainerinvoices/{{ record.retainerinvoice_id }}/address/billing?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  approve_retainer_invoice:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/approve?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_retainer_invoice_attachment:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_retainer_invoice_comment:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/comments?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_retainer_invoice_comment:
    endpoint: PUT /retainerinvoices/{{ record.retainerinvoice_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id, comment_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_retainer_invoice_comment:
    endpoint: DELETE /retainerinvoices/{{ record.retainerinvoice_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  delete_retainer_invoice_attachment:
    endpoint: DELETE /retainerinvoices/{{ record.retainerinvoice_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id, document_id
    risk: external destructive mutation in Zoho Books; approval required
  email_retainer_invoice:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/email?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  apply_retainer_payments_to_invoices:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/invoices?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_applied_retainer_payment:
    endpoint: DELETE /retainerinvoices/{{ record.retainerinvoice_id }}/invoices/{{ record.invoice_id }}?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id, invoice_id
    risk: external destructive mutation in Zoho Books; approval required
  create_retainer_invoice_async_online_payment:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/onlinepayments/asynchronous?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_retainer_invoice_draft:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/status/draft?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_retainer_invoice_sent:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/status/sent?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_retainer_invoice_void:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/status/void?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  submit_retainer_invoice:
    endpoint: POST /retainerinvoices/{{ record.retainerinvoice_id }}/submit?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_retainer_invoice_template:
    endpoint: PUT /retainerinvoices/{{ record.retainerinvoice_id }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}
    required fields: retainerinvoice_id, template_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_salesorder_customfields:
    endpoint: PUT /salesorder/{{ record.salesorder_id }}/customfields?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_sales_order:
    endpoint: POST /salesorders?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_sales_order:
    endpoint: PUT /salesorders/{{ record.salesorder_id }}?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_sales_order:
    endpoint: DELETE /salesorders/{{ record.salesorder_id }}?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external destructive mutation in Zoho Books; approval required
  update_sales_order_billing_address:
    endpoint: PUT /salesorders/{{ record.salesorder_id }}/address/billing?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_sales_order_shipping_address:
    endpoint: PUT /salesorders/{{ record.salesorder_id }}/address/shipping?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  approve_sales_order:
    endpoint: POST /salesorders/{{ record.salesorder_id }}/approve?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_sales_order_attachment:
    endpoint: POST /salesorders/{{ record.salesorder_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_sales_order_attachment_preference:
    endpoint: PUT /salesorders/{{ record.salesorder_id }}/attachment?organization_id={{ config.organization_id }}&can_send_in_mail={{ record.can_send_in_mail }}
    required fields: salesorder_id, can_send_in_mail
    risk: external mutation in Zoho Books accounting data; approval required
  delete_sales_order_attachment:
    endpoint: DELETE /salesorders/{{ record.salesorder_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external destructive mutation in Zoho Books; approval required
  add_sales_order_comment:
    endpoint: POST /salesorders/{{ record.salesorder_id }}/comments?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_sales_order_comment:
    endpoint: PUT /salesorders/{{ record.salesorder_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: salesorder_id, comment_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_sales_order_comment:
    endpoint: DELETE /salesorders/{{ record.salesorder_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: salesorder_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  email_sales_order:
    endpoint: POST /salesorders/{{ record.salesorder_id }}/email?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_sales_order_as_open:
    endpoint: POST /salesorders/{{ record.salesorder_id }}/status/open?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_sales_order_as_void:
    endpoint: POST /salesorders/{{ record.salesorder_id }}/status/void?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  submit_sales_order:
    endpoint: POST /salesorders/{{ record.salesorder_id }}/submit?organization_id={{ config.organization_id }}
    required fields: salesorder_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_sales_order_sub_status:
    endpoint: POST /salesorders/{{ record.salesorder_id }}/substatus/{{ record.status_code }}?organization_id={{ config.organization_id }}
    required fields: salesorder_id, status_code
    risk: external mutation in Zoho Books accounting data; approval required
  update_sales_order_template:
    endpoint: PUT /salesorders/{{ record.salesorder_id }}/templates/{{ record.template_id }}?organization_id={{ config.organization_id }}
    required fields: salesorder_id, template_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_sales_receipt:
    endpoint: POST /salesreceipts?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_sales_receipt:
    endpoint: PUT /salesreceipts/{{ record.sales_receipt_id }}?organization_id={{ config.organization_id }}
    required fields: sales_receipt_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_sales_receipt:
    endpoint: DELETE /salesreceipts/{{ record.sales_receipt_id }}?organization_id={{ config.organization_id }}
    required fields: sales_receipt_id
    risk: external destructive mutation in Zoho Books; approval required
  email_sales_receipt:
    endpoint: POST /salesreceipts/{{ record.sales_receipt_id }}/email
    required fields: sales_receipt_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_task:
    endpoint: POST /tasks?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_tasks:
    endpoint: PUT /tasks?organization_id={{ config.organization_id }}&bulk_update={{ record.bulk_update }}
    required fields: bulk_update
    risk: external mutation in Zoho Books accounting data; approval required
  delete_tasks:
    endpoint: DELETE /tasks?organization_id={{ config.organization_id }}&task_ids={{ record.task_ids }}
    required fields: task_ids
    risk: external destructive mutation in Zoho Books; approval required
  update_a_task:
    endpoint: PUT /tasks/{{ record.task_id }}?organization_id={{ config.organization_id }}
    required fields: task_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_task:
    endpoint: DELETE /tasks/{{ record.task_id }}?organization_id={{ config.organization_id }}
    required fields: task_id
    risk: external destructive mutation in Zoho Books; approval required
  add_task_attachment:
    endpoint: POST /tasks/{{ record.task_id }}/attachment?organization_id={{ config.organization_id }}
    required fields: task_id
    risk: external mutation in Zoho Books accounting data; approval required
  add_task_comment:
    endpoint: POST /tasks/{{ record.task_id }}/comments?organization_id={{ config.organization_id }}
    required fields: task_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_task_comment:
    endpoint: DELETE /tasks/{{ record.task_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: task_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  delete_task_document:
    endpoint: DELETE /tasks/{{ record.task_id }}/documents/{{ record.document_id }}?organization_id={{ config.organization_id }}
    required fields: task_id, document_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_task_as_completed:
    endpoint: POST /tasks/{{ record.task_id }}/markascompleted?organization_id={{ config.organization_id }}
    required fields: task_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_task_as_ongoing:
    endpoint: POST /tasks/{{ record.task_id }}/markasongoing?organization_id={{ config.organization_id }}
    required fields: task_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_task_as_open:
    endpoint: POST /tasks/{{ record.task_id }}/markasopen?organization_id={{ config.organization_id }}
    required fields: task_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_percentage_task:
    endpoint: POST /tasks/{{ record.task_id }}/percentage?organization_id={{ config.organization_id }}
    required fields: task_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_tax_authority:
    endpoint: POST /settings/taxauthorities?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_tax_authority:
    endpoint: PUT /settings/taxauthorities/{{ record.tax_authority_id }}?organization_id={{ config.organization_id }}
    required fields: tax_authority_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_tax_authority:
    endpoint: DELETE /settings/taxauthorities/{{ record.tax_authority_id }}?organization_id={{ config.organization_id }}
    required fields: tax_authority_id
    risk: external destructive mutation in Zoho Books; approval required
  create_tax:
    endpoint: POST /settings/taxes?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_tax:
    endpoint: PUT /settings/taxes/{{ record.tax_id }}?organization_id={{ config.organization_id }}
    required fields: tax_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_tax:
    endpoint: DELETE /settings/taxes/{{ record.tax_id }}?organization_id={{ config.organization_id }}
    required fields: tax_id
    risk: external destructive mutation in Zoho Books; approval required
  create_tax_exemption:
    endpoint: POST /settings/taxexemptions?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_tax_exemption:
    endpoint: PUT /settings/taxexemptions/{{ record.tax_exemption_id }}?organization_id={{ config.organization_id }}
    required fields: tax_exemption_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_tax_exemption:
    endpoint: DELETE /settings/taxexemptions/{{ record.tax_exemption_id }}?organization_id={{ config.organization_id }}
    required fields: tax_exemption_id
    risk: external destructive mutation in Zoho Books; approval required
  create_tax_group:
    endpoint: POST /settings/taxgroups?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_tax_group:
    endpoint: PUT /settings/taxgroups/{{ record.tax_group_id }}?organization_id={{ config.organization_id }}
    required fields: tax_group_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_tax_group:
    endpoint: DELETE /settings/taxgroups/{{ record.tax_group_id }}?organization_id={{ config.organization_id }}
    required fields: tax_group_id
    risk: external destructive mutation in Zoho Books; approval required
  create_time_entries:
    endpoint: POST /projects/timeentries?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  delete_time_entries:
    endpoint: DELETE /projects/timeentries?organization_id={{ config.organization_id }}
    risk: external destructive mutation in Zoho Books; approval required
  stop_entry_timer:
    endpoint: POST /projects/timeentries/timer/stop?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_time_entry:
    endpoint: PUT /projects/timeentries/{{ record.time_entry_id }}?organization_id={{ config.organization_id }}
    required fields: time_entry_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_time_entry:
    endpoint: DELETE /projects/timeentries/{{ record.time_entry_id }}?organization_id={{ config.organization_id }}
    required fields: time_entry_id
    risk: external destructive mutation in Zoho Books; approval required
  start_entry_timer:
    endpoint: POST /projects/timeentries/{{ record.time_entry_id }}/timer/start?organization_id={{ config.organization_id }}
    required fields: time_entry_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_transaction_lock:
    endpoint: PUT /transactionlock?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  delete_transaction_lock:
    endpoint: DELETE /transactionlock?organization_id={{ config.organization_id }}&transaction_lock_id={{ record.transaction_lock_id }}
    required fields: transaction_lock_id
    risk: external destructive mutation in Zoho Books; approval required
  update_partial_unlock:
    endpoint: PUT /transactionlock/partialunlock?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  create_user:
    endpoint: POST /users?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_user:
    endpoint: PUT /users/{{ record.user_id }}?organization_id={{ config.organization_id }}
    required fields: user_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_user:
    endpoint: DELETE /users/{{ record.user_id }}?organization_id={{ config.organization_id }}
    required fields: user_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_user_active:
    endpoint: POST /users/{{ record.user_id }}/active?organization_id={{ config.organization_id }}
    required fields: user_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_user_inactive:
    endpoint: POST /users/{{ record.user_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: user_id
    risk: external mutation in Zoho Books accounting data; approval required
  invite_user:
    endpoint: POST /users/{{ record.user_id }}/invite?organization_id={{ config.organization_id }}
    required fields: user_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_vendor_credit:
    endpoint: POST /vendorcredits?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  update_vendor_credit:
    endpoint: PUT /vendorcredits/{{ record.vendor_credit_id }}?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_vendor_credit:
    endpoint: DELETE /vendorcredits/{{ record.vendor_credit_id }}?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id
    risk: external destructive mutation in Zoho Books; approval required
  approve_vendor_credit:
    endpoint: POST /vendorcredits/{{ record.vendor_credit_id }}/approve?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id
    risk: external mutation in Zoho Books accounting data; approval required
  apply_credits_to_a_bill:
    endpoint: POST /vendorcredits/{{ record.vendor_credit_id }}/bills?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_vendor_credit_bill:
    endpoint: DELETE /vendorcredits/{{ record.vendor_credit_id }}/bills/{{ record.vendor_credit_bill_id }}?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id, vendor_credit_bill_id
    risk: external destructive mutation in Zoho Books; approval required
  add_vendor_credit_comment:
    endpoint: POST /vendorcredits/{{ record.vendor_credit_id }}/comments?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_vendor_credit_comment:
    endpoint: DELETE /vendorcredits/{{ record.vendor_credit_id }}/comments/{{ record.comment_id }}?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id, comment_id
    risk: external destructive mutation in Zoho Books; approval required
  refund_vendor_credit:
    endpoint: POST /vendorcredits/{{ record.vendor_credit_id }}/refunds?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_vendor_credit_refund:
    endpoint: PUT /vendorcredits/{{ record.vendor_credit_id }}/refunds/{{ record.vendor_credit_refund_id }}?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id, vendor_credit_refund_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_vendor_credit_refund:
    endpoint: DELETE /vendorcredits/{{ record.vendor_credit_id }}/refunds/{{ record.vendor_credit_refund_id }}?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id, vendor_credit_refund_id
    risk: external destructive mutation in Zoho Books; approval required
  mark_vendor_credit_open:
    endpoint: POST /vendorcredits/{{ record.vendor_credit_id }}/status/open?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id
    risk: external mutation in Zoho Books accounting data; approval required
  mark_vendor_credit_void:
    endpoint: POST /vendorcredits/{{ record.vendor_credit_id }}/status/void?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id
    risk: external mutation in Zoho Books accounting data; approval required
  submit_vendor_credit:
    endpoint: POST /vendorcredits/{{ record.vendor_credit_id }}/submit?organization_id={{ config.organization_id }}
    required fields: vendor_credit_id
    risk: external mutation in Zoho Books accounting data; approval required
  create_vendor_payment:
    endpoint: POST /vendorpayments?organization_id={{ config.organization_id }}
    risk: external mutation in Zoho Books accounting data; approval required
  bulk_delete_vendor_payments:
    endpoint: DELETE /vendorpayments?organization_id={{ config.organization_id }}&vendorpayment_id={{ record.vendorpayment_id }}&bulk_delete={{ record.bulk_delete }}
    required fields: vendorpayment_id, bulk_delete
    risk: external destructive mutation in Zoho Books; approval required
  update_vendor_payment:
    endpoint: PUT /vendorpayments/{{ record.payment_id }}?organization_id={{ config.organization_id }}
    required fields: payment_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_vendor_payment:
    endpoint: DELETE /vendorpayments/{{ record.payment_id }}?organization_id={{ config.organization_id }}
    required fields: payment_id
    risk: external destructive mutation in Zoho Books; approval required
  email_vendor_payment:
    endpoint: POST /vendorpayments/{{ record.payment_id }}/email?organization_id={{ config.organization_id }}
    required fields: payment_id
    risk: external mutation in Zoho Books accounting data; approval required
  refund_excess_vendor_payment:
    endpoint: POST /vendorpayments/{{ record.payment_id }}/refunds?organization_id={{ config.organization_id }}
    required fields: payment_id
    risk: external mutation in Zoho Books accounting data; approval required
  update_vendor_payment_refund:
    endpoint: PUT /vendorpayments/{{ record.payment_id }}/refunds/{{ record.vendorpayment_refund_id }}?organization_id={{ config.organization_id }}
    required fields: payment_id, vendorpayment_refund_id
    risk: external mutation in Zoho Books accounting data; approval required
  delete_vendor_payment_refund:
    endpoint: DELETE /vendorpayments/{{ record.payment_id }}/refunds/{{ record.vendorpayment_refund_id }}?organization_id={{ config.organization_id }}
    required fields: payment_id, vendorpayment_refund_id
    risk: external destructive mutation in Zoho Books; approval required

SECURITY
  read risk: external Zoho Books API read of accounting/contact data
  write risk: external Zoho Books API mutation of accounting, contact, inventory, project, banking, tax, and organization-adjacent records
  approval: writes require explicit user approval before execution; destructive deletes are marked destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Zoho Books's declared typed write actions.
  Usage: pm zoho-books <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Reverse ETL writes
  Other Commands
    active tag apply - POST /reportingtags/{tag_id}/active (active_tag) [intent=reverse_etl availability=partial write=active_tag]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tag-id (required), --organization-id
    active tag option apply - POST /reportingtags/{tag_id}/option/(\d+)/active (active_tag_option) [intent=reverse_etl availability=partial write=active_tag_option]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tag-id (required), --organization-id
    add bank reconciliation attachment apply - POST /bankaccounts/{account_id}/reconciliations/{reconciliation_id}/attachment (add_bank_reconciliation_attachment) [intent=reverse_etl availability=partial write=add_bank_reconciliation_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --reconciliation-id (required), --organization-id
    add bill attachment apply - POST /bills/{bill_id}/attachment (add_bill_attachment) [intent=reverse_etl availability=partial write=add_bill_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    add bill comment apply - POST /bills/{bill_id}/comments (add_bill_comment) [intent=reverse_etl availability=partial write=add_bill_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    add contact address apply - POST /contacts/{contact_id}/address (add_contact_address) [intent=reverse_etl availability=partial write=add_contact_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    add contact attachment apply - POST /contacts/{contact_id}/attachment (add_contact_attachment) [intent=reverse_etl availability=partial write=add_contact_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    add contact bank account apply - POST /contacts/{contact_id}/bankaccount (add_contact_bank_account) [intent=reverse_etl availability=partial write=add_contact_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    add contact card apply - POST /contacts/{contact_id}/card (add_contact_card) [intent=reverse_etl availability=partial write=add_contact_card]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    add contact comment apply - POST /contacts/{contact_id}/comments (add_contact_comment) [intent=reverse_etl availability=partial write=add_contact_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    add contact tax info apply - POST /contacts/{contact_id}/taxinfo (add_contact_tax_info) [intent=reverse_etl availability=partial write=add_contact_tax_info]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    add credit note attachment apply - POST /creditnotes/{creditnote_id}/attachment (add_credit_note_attachment) [intent=reverse_etl availability=partial write=add_credit_note_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    add credit note comment apply - POST /creditnotes/{creditnote_id}/comments (add_credit_note_comment) [intent=reverse_etl availability=partial write=add_credit_note_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    add credit note digital signature apply - POST /creditnotes/{creditnote_id}/dsign (add_credit_note_digital_signature) [intent=reverse_etl availability=partial write=add_credit_note_digital_signature]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    add invoice attachment apply - POST /invoices/{invoice_id}/attachment (add_invoice_attachment) [intent=reverse_etl availability=partial write=add_invoice_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    add invoice comment apply - POST /invoices/{invoice_id}/comments (add_invoice_comment) [intent=reverse_etl availability=partial write=add_invoice_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    add invoice digital signature apply - POST /invoices/{invoice_id}/dsign (add_invoice_digital_signature) [intent=reverse_etl availability=partial write=add_invoice_digital_signature]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    add invoice document apply - POST /invoices/{invoice_id}/documents/{document_id} (add_invoice_document) [intent=reverse_etl availability=partial write=add_invoice_document]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --document-id (required), --invoice-id (required), --organization-id
    add invoice online payment bank account apply - POST /invoices/{invoice_id}/onlinepayments/bankaccount (add_invoice_online_payment_bank_account) [intent=reverse_etl availability=partial write=add_invoice_online_payment_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    add item to portal apply - POST /items/{item_id}/addtoportal (add_item_to_portal) [intent=reverse_etl availability=partial write=add_item_to_portal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --item-id (required), --organization-id
    add items to portal apply - POST /items/addtoportal (add_items_to_portal) [intent=reverse_etl availability=partial write=add_items_to_portal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --item-ids (required), --organization-id
    add journal attachment apply - POST /journals/{journal_id}/attachment (add_journal_attachment) [intent=reverse_etl availability=partial write=add_journal_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --organization-id
    add journal comment apply - POST /journals/{journal_id}/comments (add_journal_comment) [intent=reverse_etl availability=partial write=add_journal_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --organization-id
    add project comment apply - POST /projects/{project_id}/comments (add_project_comment) [intent=reverse_etl availability=partial write=add_project_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --organization-id
    add project task apply - POST /projects/{project_id}/tasks (add_project_task) [intent=reverse_etl availability=partial write=add_project_task]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --organization-id
    add project user apply - POST /projects/{project_id}/users (add_project_user) [intent=reverse_etl availability=partial write=add_project_user]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --organization-id
    add purchase order attachment apply - POST /purchaseorders/{purchaseorder_id}/attachment (add_purchase_order_attachment) [intent=reverse_etl availability=partial write=add_purchase_order_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    add purchase order comment apply - POST /purchaseorders/{purchaseorder_id}/comments (add_purchase_order_comment) [intent=reverse_etl availability=partial write=add_purchase_order_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    add retainer invoice attachment apply - POST /retainerinvoices/{retainerinvoice_id}/attachment (add_retainer_invoice_attachment) [intent=reverse_etl availability=partial write=add_retainer_invoice_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    add retainer invoice comment apply - POST /retainerinvoices/{retainerinvoice_id}/comments (add_retainer_invoice_comment) [intent=reverse_etl availability=partial write=add_retainer_invoice_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    add sales order attachment apply - POST /salesorders/{salesorder_id}/attachment (add_sales_order_attachment) [intent=reverse_etl availability=partial write=add_sales_order_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    add sales order comment apply - POST /salesorders/{salesorder_id}/comments (add_sales_order_comment) [intent=reverse_etl availability=partial write=add_sales_order_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    add task apply - POST /tasks (add_task) [intent=reverse_etl availability=partial write=add_task]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    add task attachment apply - POST /tasks/{task_id}/attachment (add_task_attachment) [intent=reverse_etl availability=partial write=add_task_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --task-id (required), --organization-id
    add task comment apply - POST /tasks/{task_id}/comments (add_task_comment) [intent=reverse_etl availability=partial write=add_task_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --task-id (required), --organization-id
    add vendor credit comment apply - POST /vendorcredits/{vendor_credit_id}/comments (add_vendor_credit_comment) [intent=reverse_etl availability=partial write=add_vendor_credit_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --organization-id
    apply credit note substatus apply - POST /creditnotes/{creditnote_id}/substatus/{substatus_id} (apply_credit_note_substatus) [intent=reverse_etl availability=partial write=apply_credit_note_substatus]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --substatus-id (required), --organization-id
    apply credit note to invoice apply - POST /creditnotes/{creditnote_id}/invoices (apply_credit_note_to_invoice) [intent=reverse_etl availability=partial write=apply_credit_note_to_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    apply credits to a bill apply - POST /vendorcredits/{vendor_credit_id}/bills (apply_credits_to_a_bill) [intent=reverse_etl availability=partial write=apply_credits_to_a_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --organization-id
    apply credits to bill apply - POST /bills/{bill_id}/credits (apply_credits_to_bill) [intent=reverse_etl availability=partial write=apply_credits_to_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    apply credits to invoice apply - POST /invoices/{invoice_id}/credits (apply_credits_to_invoice) [intent=reverse_etl availability=partial write=apply_credits_to_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    apply invoice substatus apply - POST /invoices/{invoice_id}/substatus/{substatus_id} (apply_invoice_substatus) [intent=reverse_etl availability=partial write=apply_invoice_substatus]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --substatus-id (required), --organization-id
    apply journal credits to bills apply - POST /journals/{journal_id}/credits/{journal_line_id}/bills (apply_journal_credits_to_bills) [intent=reverse_etl availability=partial write=apply_journal_credits_to_bills]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --journal-line-id (required), --organization-id
    apply journal credits to invoices apply - POST /journals/{journal_id}/credits/{journal_line_id}/invoices (apply_journal_credits_to_invoices) [intent=reverse_etl availability=partial write=apply_journal_credits_to_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --journal-line-id (required), --organization-id
    apply pricebook to invoice apply - PUT /invoices/{invoice_id}/pricebooks/{pricebook_id} (apply_pricebook_to_invoice) [intent=reverse_etl availability=partial write=apply_pricebook_to_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --pricebook-id (required), --organization-id
    apply retainer payments to invoices apply - POST /retainerinvoices/{retainerinvoice_id}/invoices (apply_retainer_payments_to_invoices) [intent=reverse_etl availability=partial write=apply_retainer_payments_to_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    approve bill apply - POST /bills/{bill_id}/approve (approve_bill) [intent=reverse_etl availability=partial write=approve_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    approve contact bank account apply - POST /contacts/{contact_id}/bankaccount/{bank_account_id}/approve (approve_contact_bank_account) [intent=reverse_etl availability=partial write=approve_contact_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bank-account-id (required), --contact-id (required), --organization-id
    approve credit note apply - POST /creditnotes/{creditnote_id}/approve (approve_credit_note) [intent=reverse_etl availability=partial write=approve_credit_note]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    approve credit notes apply - POST /creditnotes/approve (approve_credit_notes) [intent=reverse_etl availability=partial write=approve_credit_notes]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-ids (required), --organization-id
    approve estimate apply - POST /estimates/{estimate_id}/approve (approve_estimate) [intent=reverse_etl availability=partial write=approve_estimate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    approve invoice apply - POST /invoices/{invoice_id}/approve (approve_invoice) [intent=reverse_etl availability=partial write=approve_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    approve invoices apply - POST /invoices/approve (approve_invoices) [intent=reverse_etl availability=partial write=approve_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    approve journal apply - POST /journals/{journal_id}/approve (approve_journal) [intent=reverse_etl availability=partial write=approve_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --organization-id
    approve purchase order apply - POST /purchaseorders/{purchaseorder_id}/approve (approve_purchase_order) [intent=reverse_etl availability=partial write=approve_purchase_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    approve retainer invoice apply - POST /retainerinvoices/{retainerinvoice_id}/approve (approve_retainer_invoice) [intent=reverse_etl availability=partial write=approve_retainer_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    approve sales order apply - POST /salesorders/{salesorder_id}/approve (approve_sales_order) [intent=reverse_etl availability=partial write=approve_sales_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    approve vendor credit apply - POST /vendorcredits/{vendor_credit_id}/approve (approve_vendor_credit) [intent=reverse_etl availability=partial write=approve_vendor_credit]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --organization-id
    assign contact owner apply - POST /contacts/{contact_id}/owner (assign_contact_owner) [intent=reverse_etl availability=partial write=assign_contact_owner]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    assign owner to contacts apply - POST /contacts/owner (assign_owner_to_contacts) [intent=reverse_etl availability=partial write=assign_owner_to_contacts]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-ids (required), --organization-id
    associate recurring invoice bank account apply - POST /recurringinvoices/{recurring_invoice_id}/bankaccount/{account_id} (associate_recurring_invoice_bank_account) [intent=reverse_etl availability=partial write=associate_recurring_invoice_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --recurring-invoice-id (required), --organization-id
    associate recurring invoice card apply - POST /recurringinvoices/{recurring_invoice_id}/card/{card_id} (associate_recurring_invoice_card) [intent=reverse_etl availability=partial write=associate_recurring_invoice_card]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --card-id (required), --recurring-invoice-id (required), --organization-id
    bulk approve journals apply - POST /journals/approve (bulk_approve_journals) [intent=reverse_etl availability=partial write=bulk_approve_journals]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-ids (required), --organization-id
    bulk delete bank account rules apply - DELETE /bankaccounts/rules (bulk_delete_bank_account_rules) [intent=reverse_etl availability=partial write=bulk_delete_bank_account_rules]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --rule-ids (required), --organization-id
    bulk delete base currency adjustments apply - DELETE /basecurrencyadjustment/bulkdelete (bulk_delete_base_currency_adjustments) [intent=reverse_etl availability=partial write=bulk_delete_base_currency_adjustments]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --base-currency-adjustment-ids (required), --organization-id
    bulk delete chart of accounts apply - DELETE /chartofaccounts/bulkdelete (bulk_delete_chart_of_accounts) [intent=reverse_etl availability=partial write=bulk_delete_chart_of_accounts]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-ids (required), --organization-id
    bulk delete customer payments apply - DELETE /customerpayments (bulk_delete_customer_payments) [intent=reverse_etl availability=partial write=bulk_delete_customer_payments]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bulk-delete (required), --payment-ids (required), --organization-id
    bulk delete journals apply - DELETE /journals/bulkdelete (bulk_delete_journals) [intent=reverse_etl availability=partial write=bulk_delete_journals]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-ids (required), --organization-id
    bulk delete register transactions apply - PUT /registers/{account_id}/transactions/bulkdelete (bulk_delete_register_transactions) [intent=reverse_etl availability=partial write=bulk_delete_register_transactions]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    bulk delete vendor payments apply - DELETE /vendorpayments (bulk_delete_vendor_payments) [intent=reverse_etl availability=partial write=bulk_delete_vendor_payments]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bulk-delete (required), --vendorpayment-id (required), --organization-id
    bulk invoice reminder apply - POST /invoices/paymentreminder (bulk_invoice_reminder) [intent=reverse_etl availability=partial write=bulk_invoice_reminder]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    bulk mark chart of accounts active apply - POST /chartofaccounts/active (bulk_mark_chart_of_accounts_active) [intent=reverse_etl availability=partial write=bulk_mark_chart_of_accounts_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-ids (required), --organization-id
    bulk mark chart of accounts inactive apply - POST /chartofaccounts/inactive (bulk_mark_chart_of_accounts_inactive) [intent=reverse_etl availability=partial write=bulk_mark_chart_of_accounts_inactive]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-ids (required), --organization-id
    bulk publish journals apply - POST /journals/status/publish (bulk_publish_journals) [intent=reverse_etl availability=partial write=bulk_publish_journals]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-ids (required), --organization-id
    bulk submit journals apply - POST /journals/submit (bulk_submit_journals) [intent=reverse_etl availability=partial write=bulk_submit_journals]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-ids (required), --organization-id
    bulk update bank account rules apply - PUT /bankaccounts/rules (bulk_update_bank_account_rules) [intent=reverse_etl availability=partial write=bulk_update_bank_account_rules]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    bulk update custom module records apply - PUT /{module_name} (bulk_update_custom_module_records) [intent=reverse_etl availability=partial write=bulk_update_custom_module_records]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --module-name (required), --organization-id
    bulk update register transactions apply - PUT /registers/{account_id}/transactions/bulkupdate (bulk_update_register_transactions) [intent=reverse_etl availability=partial write=bulk_update_register_transactions]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    cancel credit note einvoice apply - POST /creditnotes/{creditnote_id}/einvoice/cancel (cancel_credit_note_einvoice) [intent=reverse_etl availability=partial write=cancel_credit_note_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    cancel credit notes einvoice apply - POST /creditnotes/einvoice/cancel (cancel_credit_notes_einvoice) [intent=reverse_etl availability=partial write=cancel_credit_notes_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    cancel einvoice credit note apply - POST /einvoices/creditnotes/{creditnote_id}/cancel (cancel_einvoice_credit_note) [intent=reverse_etl availability=partial write=cancel_einvoice_credit_note]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    cancel einvoice invoice apply - POST /einvoices/invoices/{invoice_id}/cancel (cancel_einvoice_invoice) [intent=reverse_etl availability=partial write=cancel_einvoice_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    cancel invoice apply - POST /invoices/{invoice_id}/status/cancel (cancel_invoice) [intent=reverse_etl availability=partial write=cancel_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    cancel invoice einvoice apply - POST /invoices/{invoice_id}/einvoice/cancel (cancel_invoice_einvoice) [intent=reverse_etl availability=partial write=cancel_invoice_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    cancel invoices einvoice apply - POST /invoices/einvoice/cancel (cancel_invoices_einvoice) [intent=reverse_etl availability=partial write=cancel_invoices_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    cancel scheduled invoice email apply - DELETE /invoices/{invoice_id}/email/schedule (cancel_scheduled_invoice_email) [intent=reverse_etl availability=partial write=cancel_scheduled_invoice_email]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    cancel write off invoice apply - POST /invoices/{invoice_id}/writeoff/cancel (cancel_write_off_invoice) [intent=reverse_etl availability=partial write=cancel_write_off_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    cancel writeoff opening balance apply - POST /openingbalances/{opening_balance_id}/writeoff/cancel (cancel_writeoff_opening_balance) [intent=reverse_etl availability=partial write=cancel_writeoff_opening_balance]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --opening-balance-id (required), --organization-id
    categorize as credit note refunds apply - POST /banktransactions/uncategorized/{transaction_id}/categorize/creditnoterefunds (categorize_as_credit_note_refunds) [intent=reverse_etl availability=partial write=categorize_as_credit_note_refunds]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    categorize as vendor credit refunds apply - POST /banktransactions/uncategorized/{transaction_id}/categorize/vendorcreditrefunds (categorize_as_vendor_credit_refunds) [intent=reverse_etl availability=partial write=categorize_as_vendor_credit_refunds]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    categorize as vendor payment refund apply - POST /banktransactions/uncategorized/{statement_line_id}/categorize/vendorpaymentrefunds (categorize_as_vendor_payment_refund) [intent=reverse_etl availability=partial write=categorize_as_vendor_payment_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --statement-line-id (required), --organization-id
    categorize bank transaction apply - POST /banktransactions/uncategorized/{transaction_id}/categorize (categorize_bank_transaction) [intent=reverse_etl availability=partial write=categorize_bank_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    categorize bank transaction as customer payment apply - POST /banktransactions/uncategorized/{transaction_id}/categorize/customerpayments (categorize_bank_transaction_as_customer_payment) [intent=reverse_etl availability=partial write=categorize_bank_transaction_as_customer_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    categorize bank transaction as expense apply - POST /banktransactions/uncategorized/{transaction_id}/categorize/expenses (categorize_bank_transaction_as_expense) [intent=reverse_etl availability=partial write=categorize_bank_transaction_as_expense]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    categorize bank transaction as payment refund apply - POST /banktransactions/uncategorized/{statement_line_id}/categorize/paymentrefunds (categorize_bank_transaction_as_payment_refund) [intent=reverse_etl availability=partial write=categorize_bank_transaction_as_payment_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --statement-line-id (required), --organization-id
    categorize bank transaction as vendor payment apply - POST /banktransactions/uncategorized/{transaction_id}/categorize/vendorpayments (categorize_bank_transaction_as_vendor_payment) [intent=reverse_etl availability=partial write=categorize_bank_transaction_as_vendor_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    clone project apply - POST /projects/{project_id}/clone (clone_project) [intent=reverse_etl availability=partial write=clone_project]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --organization-id
    copy organization settings apply - POST /organizations/{organization_id}/copysettings (copy_organization_settings) [intent=reverse_etl availability=implemented write=copy_organization_settings]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --organization-id (required), --settings-to-copy (required)
    create bank account apply - POST /bankaccounts (create_bank_account) [intent=reverse_etl availability=partial write=create_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create bank account match filter apply - POST /bankaccounts/matchfilters (create_bank_account_match_filter) [intent=reverse_etl availability=partial write=create_bank_account_match_filter]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create bank account rule apply - POST /bankaccounts/rules (create_bank_account_rule) [intent=reverse_etl availability=partial write=create_bank_account_rule]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create bank reconciliation apply - POST /bankaccounts/{account_id}/reconciliations (create_bank_reconciliation) [intent=reverse_etl availability=partial write=create_bank_reconciliation]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    create bank transaction apply - POST /banktransactions (create_bank_transaction) [intent=reverse_etl availability=partial write=create_bank_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create base currency adjustment apply - POST /basecurrencyadjustment (create_base_currency_adjustment) [intent=reverse_etl availability=partial write=create_base_currency_adjustment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-ids (required), --organization-id
    create bill apply - POST /bills (create_bill) [intent=reverse_etl availability=partial write=create_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create chart of account apply - POST /chartofaccounts (create_chart_of_account) [intent=reverse_etl availability=partial write=create_chart_of_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create contact apply - POST /contacts (create_contact) [intent=reverse_etl availability=partial write=create_contact]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create contact person 2 apply - Typed action create_contact_person_2 [intent=reverse_etl availability=partial write=create_contact_person_2]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for create_contact_person_2; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for create_contact_person_2; found 0.
    create contact person apply - Typed action create_contact_person [intent=reverse_etl availability=partial write=create_contact_person]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for create_contact_person; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for create_contact_person; found 0.
    create credit note apply - POST /creditnotes (create_credit_note) [intent=reverse_etl availability=partial write=create_credit_note]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create credit note refund apply - POST /creditnotes/{creditnote_id}/refunds (create_credit_note_refund) [intent=reverse_etl availability=partial write=create_credit_note_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    create currency apply - POST /settings/currencies (create_currency) [intent=reverse_etl availability=partial write=create_currency]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create custom module apply - POST /settings/modules (create_custom_module) [intent=reverse_etl availability=partial write=create_custom_module]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create custom module record apply - POST /{module_name} (create_custom_module_record) [intent=reverse_etl availability=partial write=create_custom_module_record]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --module-name (required), --organization-id
    create customer debit note apply - Typed action create_customer_debit_note [intent=reverse_etl availability=partial write=create_customer_debit_note]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for create_customer_debit_note; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for create_customer_debit_note; found 0.
    create customer payment apply - POST /customerpayments (create_customer_payment) [intent=reverse_etl availability=partial write=create_customer_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create customer payment refund apply - POST /customerpayments/{customer_payment_id}/refunds (create_customer_payment_refund) [intent=reverse_etl availability=partial write=create_customer_payment_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --customer-payment-id (required), --organization-id
    create delivery challan apply - POST /deliverychallans (create_delivery_challan) [intent=reverse_etl availability=partial write=create_delivery_challan]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create employee apply - POST /employees (create_employee) [intent=reverse_etl availability=partial write=create_employee]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create estimate apply - POST /estimates (create_estimate) [intent=reverse_etl availability=partial write=create_estimate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create estimate comment apply - POST /estimates/{estimate_id}/comments (create_estimate_comment) [intent=reverse_etl availability=partial write=create_estimate_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    create exchange rate apply - POST /settings/currencies/{currency_id}/exchangerates (create_exchange_rate) [intent=reverse_etl availability=partial write=create_exchange_rate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --currency-id (required), --organization-id
    create expense apply - POST /expenses (create_expense) [intent=reverse_etl availability=partial write=create_expense]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create expense receipt apply - POST /expenses/{expense_id}/receipt (create_expense_receipt) [intent=reverse_etl availability=partial write=create_expense_receipt]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --expense-id (required), --organization-id
    create fixed asset apply - POST /fixedassets (create_fixed_asset) [intent=reverse_etl availability=partial write=create_fixed_asset]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create fixed asset comment apply - POST /fixedassets/{fixed_asset_id}/comments (create_fixed_asset_comment) [intent=reverse_etl availability=partial write=create_fixed_asset_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-id (required), --organization-id
    create fixed asset type apply - POST /fixedassettypes (create_fixed_asset_type) [intent=reverse_etl availability=partial write=create_fixed_asset_type]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create invoice apply - Typed action create_invoice [intent=reverse_etl availability=partial write=create_invoice]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for create_invoice; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for create_invoice; found 0.
    create invoice asynchronous online payment apply - POST /invoices/{invoice_id}/onlinepayments/asynchronous (create_invoice_asynchronous_online_payment) [intent=reverse_etl availability=partial write=create_invoice_asynchronous_online_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    create invoice from salesorder apply - POST /invoices/fromsalesorder (create_invoice_from_salesorder) [intent=reverse_etl availability=partial write=create_invoice_from_salesorder]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    create invoice synchronous online payment apply - POST /invoices/{invoice_id}/onlinepayments/synchronous (create_invoice_synchronous_online_payment) [intent=reverse_etl availability=partial write=create_invoice_synchronous_online_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    create invoices from estimates apply - POST /invoices/fromestimates (create_invoices_from_estimates) [intent=reverse_etl availability=partial write=create_invoices_from_estimates]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-ids (required), --organization-id
    create invoices from projects apply - POST /invoices/fromprojects (create_invoices_from_projects) [intent=reverse_etl availability=partial write=create_invoices_from_projects]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create item apply - POST /items (create_item) [intent=reverse_etl availability=partial write=create_item]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create journal apply - POST /journals (create_journal) [intent=reverse_etl availability=partial write=create_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create location apply - POST /locations (create_location) [intent=reverse_etl availability=partial write=create_location]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create opening balance apply - POST /settings/openingbalances (create_opening_balance) [intent=reverse_etl availability=partial write=create_opening_balance]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create organization address apply - POST /organizations/address (create_organization_address) [intent=reverse_etl availability=implemented write=create_organization_address]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create organization apply - POST /organizations (create_organization) [intent=reverse_etl availability=implemented write=create_organization]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create pricebook apply - POST /pricebooks (create_pricebook) [intent=reverse_etl availability=partial write=create_pricebook]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create project apply - POST /projects (create_project) [intent=reverse_etl availability=partial write=create_project]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create purchase order apply - POST /purchaseorders (create_purchase_order) [intent=reverse_etl availability=partial write=create_purchase_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create recurring bill apply - POST /recurringbills (create_recurring_bill) [intent=reverse_etl availability=partial write=create_recurring_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create recurring expense apply - POST /recurringexpenses (create_recurring_expense) [intent=reverse_etl availability=partial write=create_recurring_expense]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create recurring invoice apply - POST /recurringinvoices (create_recurring_invoice) [intent=reverse_etl availability=partial write=create_recurring_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create recurring journal apply - POST /recurringjournals (create_recurring_journal) [intent=reverse_etl availability=partial write=create_recurring_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create retainer invoice apply - POST /retainerinvoices (create_retainer_invoice) [intent=reverse_etl availability=partial write=create_retainer_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create retainer invoice async online payment apply - POST /retainerinvoices/{retainerinvoice_id}/onlinepayments/asynchronous (create_retainer_invoice_async_online_payment) [intent=reverse_etl availability=partial write=create_retainer_invoice_async_online_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    create sales order apply - POST /salesorders (create_sales_order) [intent=reverse_etl availability=partial write=create_sales_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create sales receipt apply - POST /salesreceipts (create_sales_receipt) [intent=reverse_etl availability=partial write=create_sales_receipt]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create tag apply - POST /reportingtags (create_tag) [intent=reverse_etl availability=partial write=create_tag]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create tax apply - POST /settings/taxes (create_tax) [intent=reverse_etl availability=partial write=create_tax]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create tax authority apply - POST /settings/taxauthorities (create_tax_authority) [intent=reverse_etl availability=partial write=create_tax_authority]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create tax exemption apply - POST /settings/taxexemptions (create_tax_exemption) [intent=reverse_etl availability=partial write=create_tax_exemption]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create tax group apply - POST /settings/taxgroups (create_tax_group) [intent=reverse_etl availability=partial write=create_tax_group]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create time entries apply - POST /projects/timeentries (create_time_entries) [intent=reverse_etl availability=partial write=create_time_entries]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create user apply - POST /users (create_user) [intent=reverse_etl availability=partial write=create_user]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create vendor credit apply - POST /vendorcredits (create_vendor_credit) [intent=reverse_etl availability=partial write=create_vendor_credit]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    create vendor payment apply - POST /vendorpayments (create_vendor_payment) [intent=reverse_etl availability=partial write=create_vendor_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    decline contact bank account apply - POST /contacts/{contact_id}/bankaccount/{bank_account_id}/decline (decline_contact_bank_account) [intent=reverse_etl availability=partial write=decline_contact_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bank-account-id (required), --contact-id (required), --organization-id
    delete applied retainer payment apply - DELETE /retainerinvoices/{retainerinvoice_id}/invoices/{invoice_id} (delete_applied_retainer_payment) [intent=reverse_etl availability=partial write=delete_applied_retainer_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --retainerinvoice-id (required), --organization-id
    delete bank account apply - DELETE /bankaccounts/{account_id} (delete_bank_account) [intent=reverse_etl availability=partial write=delete_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    delete bank account match filter apply - DELETE /bankaccounts/matchfilters/{match_filter_id} (delete_bank_account_match_filter) [intent=reverse_etl availability=partial write=delete_bank_account_match_filter]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --match-filter-id (required), --organization-id
    delete bank account rule apply - DELETE /bankaccounts/rules/{rule_id} (delete_bank_account_rule) [intent=reverse_etl availability=partial write=delete_bank_account_rule]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --rule-id (required), --organization-id
    delete bank reconciliation apply - DELETE /bankaccounts/{account_id}/reconciliations/{reconciliation_id} (delete_bank_reconciliation) [intent=reverse_etl availability=partial write=delete_bank_reconciliation]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --reconciliation-id (required), --organization-id
    delete bank reconciliation document apply - DELETE /bankaccounts/{account_id}/reconciliations/{reconciliation_id}/documents/{document_id} (delete_bank_reconciliation_document) [intent=reverse_etl availability=partial write=delete_bank_reconciliation_document]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --document-id (required), --reconciliation-id (required), --organization-id
    delete bank transaction apply - DELETE /banktransactions/{bank_transaction_id} (delete_bank_transaction) [intent=reverse_etl availability=partial write=delete_bank_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bank-transaction-id (required), --organization-id
    delete base currency adjustment apply - DELETE /basecurrencyadjustment/{base_currency_adjustment_id} (delete_base_currency_adjustment) [intent=reverse_etl availability=partial write=delete_base_currency_adjustment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --base-currency-adjustment-id (required), --organization-id
    delete bill apply - DELETE /bills/{bill_id} (delete_bill) [intent=reverse_etl availability=partial write=delete_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    delete bill attachment apply - DELETE /bills/{bill_id}/attachment (delete_bill_attachment) [intent=reverse_etl availability=partial write=delete_bill_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    delete bill comment apply - DELETE /bills/{bill_id}/comments/{comment_id} (delete_bill_comment) [intent=reverse_etl availability=partial write=delete_bill_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --comment-id (required), --organization-id
    delete bill payment apply - DELETE /bills/{bill_id}/payments/{bill_payment_id} (delete_bill_payment) [intent=reverse_etl availability=partial write=delete_bill_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --bill-payment-id (required), --organization-id
    delete chart of account apply - DELETE /chartofaccounts/{account_id} (delete_chart_of_account) [intent=reverse_etl availability=partial write=delete_chart_of_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    delete chart of account transaction apply - DELETE /chartofaccounts/transactions/{transaction_id} (delete_chart_of_account_transaction) [intent=reverse_etl availability=partial write=delete_chart_of_account_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    delete contact address apply - DELETE /contacts/{contact_id}/address/{address_id} (delete_contact_address) [intent=reverse_etl availability=partial write=delete_contact_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --address-id (required), --contact-id (required), --organization-id
    delete contact apply - DELETE /contacts/{contact_id} (delete_contact) [intent=reverse_etl availability=partial write=delete_contact]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    delete contact bank account apply - DELETE /contacts/{contact_id}/bankaccount/{bank_account_id} (delete_contact_bank_account) [intent=reverse_etl availability=partial write=delete_contact_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bank-account-id (required), --contact-id (required), --organization-id
    delete contact card apply - DELETE /contacts/{contact_id}/card/{card_id} (delete_contact_card) [intent=reverse_etl availability=partial write=delete_contact_card]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --card-id (required), --contact-id (required), --organization-id
    delete contact comment apply - DELETE /contacts/{contact_id}/comments/{comment_id} (delete_contact_comment) [intent=reverse_etl availability=partial write=delete_contact_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --contact-id (required), --organization-id
    delete contact document apply - DELETE /contacts/{contact_id}/documents/{document_id} (delete_contact_document) [intent=reverse_etl availability=partial write=delete_contact_document]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --document-id (required), --organization-id
    delete contact person 2 apply - Typed action delete_contact_person_2 [intent=reverse_etl availability=partial write=delete_contact_person_2]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for delete_contact_person_2; found 0.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for delete_contact_person_2; found 0.; flags: --contactperson-id (required)
    delete contact person apply - Typed action delete_contact_person [intent=reverse_etl availability=partial write=delete_contact_person]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for delete_contact_person; found 0.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for delete_contact_person; found 0.; flags: --contact-person-id (required)
    delete contact tag apply - DELETE /contacts/{contact_id}/tags/{tag_id} (delete_contact_tag) [intent=reverse_etl availability=partial write=delete_contact_tag]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --tag-id (required), --organization-id
    delete contact tax info apply - DELETE /contacts/{contact_id}/taxinfo/{tax_info_id} (delete_contact_tax_info) [intent=reverse_etl availability=partial write=delete_contact_tax_info]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --tax-info-id (required), --organization-id
    delete contacts apply - DELETE /contacts (delete_contacts) [intent=reverse_etl availability=partial write=delete_contacts]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-ids (required), --organization-id
    delete credit note apply - DELETE /creditnotes/{creditnote_id} (delete_credit_note) [intent=reverse_etl availability=partial write=delete_credit_note]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    delete credit note comment apply - DELETE /creditnotes/{creditnote_id}/comments/{comment_id} (delete_credit_note_comment) [intent=reverse_etl availability=partial write=delete_credit_note_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --creditnote-id (required), --organization-id
    delete credit note document apply - DELETE /creditnotes/{creditnote_id}/documents/{document_id} (delete_credit_note_document) [intent=reverse_etl availability=partial write=delete_credit_note_document]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --document-id (required), --organization-id
    delete credit note einvoice status apply - DELETE /creditnotes/{creditnote_id}/einvoice/status (delete_credit_note_einvoice_status) [intent=reverse_etl availability=partial write=delete_credit_note_einvoice_status]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    delete credit note refund apply - DELETE /creditnotes/{creditnote_id}/refunds/{creditnote_refund_id} (delete_credit_note_refund) [intent=reverse_etl availability=partial write=delete_credit_note_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --creditnote-refund-id (required), --organization-id
    delete credit note substatus apply - DELETE /creditnotes/{creditnote_id}/substatus/{substatus_id} (delete_credit_note_substatus) [intent=reverse_etl availability=partial write=delete_credit_note_substatus]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --substatus-id (required), --organization-id
    delete currency apply - DELETE /settings/currencies/{currency_id} (delete_currency) [intent=reverse_etl availability=partial write=delete_currency]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --currency-id (required), --organization-id
    delete custom module apply - DELETE /settings/modules/{module_api_name} (delete_custom_module) [intent=reverse_etl availability=partial write=delete_custom_module]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --module-api-name (required), --organization-id
    delete custom module record apply - DELETE /{module_name}/{module_id} (delete_custom_module_record) [intent=reverse_etl availability=partial write=delete_custom_module_record]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --module-id (required), --module-name (required), --organization-id
    delete custom module records apply - DELETE /{module_name} (delete_custom_module_records) [intent=reverse_etl availability=partial write=delete_custom_module_records]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --module-name (required), --organization-id
    delete customer debit note apply - Typed action delete_customer_debit_note [intent=reverse_etl availability=partial write=delete_customer_debit_note]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for delete_customer_debit_note; found 0.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for delete_customer_debit_note; found 0.; flags: --debit-note-id (required)
    delete customer payment apply - DELETE /customerpayments/{payment_id} (delete_customer_payment) [intent=reverse_etl availability=partial write=delete_customer_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --payment-id (required), --organization-id
    delete customer payment refund apply - DELETE /customerpayments/{customer_payment_id}/refunds/{refund_id} (delete_customer_payment_refund) [intent=reverse_etl availability=partial write=delete_customer_payment_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --customer-payment-id (required), --refund-id (required), --organization-id
    delete delivery challan apply - DELETE /deliverychallans/{deliverychallan_id} (delete_delivery_challan) [intent=reverse_etl availability=partial write=delete_delivery_challan]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-id (required), --organization-id
    delete delivery challan attachment apply - DELETE /deliverychallans/{deliverychallan_id}/documents/{document_id} (delete_delivery_challan_attachment) [intent=reverse_etl availability=partial write=delete_delivery_challan_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-id (required), --document-id (required), --organization-id
    delete employee apply - DELETE /employee/{employee_id} (delete_employee) [intent=reverse_etl availability=partial write=delete_employee]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --employee-id (required), --organization-id
    delete estimate apply - DELETE /estimates/{estimate_id} (delete_estimate) [intent=reverse_etl availability=partial write=delete_estimate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    delete estimate comment apply - DELETE /estimates/{estimate_id}/comments/{comment_id} (delete_estimate_comment) [intent=reverse_etl availability=partial write=delete_estimate_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --estimate-id (required), --organization-id
    delete exchange rate apply - DELETE /settings/currencies/{currency_id}/exchangerates/{exchange_rate_id} (delete_exchange_rate) [intent=reverse_etl availability=partial write=delete_exchange_rate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --currency-id (required), --exchange-rate-id (required), --organization-id
    delete expense apply - DELETE /expenses/{expense_id} (delete_expense) [intent=reverse_etl availability=partial write=delete_expense]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --expense-id (required), --organization-id
    delete expense receipt apply - DELETE /expenses/{expense_id}/receipt (delete_expense_receipt) [intent=reverse_etl availability=partial write=delete_expense_receipt]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --expense-id (required), --organization-id
    delete fixed asset apply - DELETE /fixedassets/{fixed_asset_id} (delete_fixed_asset) [intent=reverse_etl availability=partial write=delete_fixed_asset]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-id (required), --organization-id
    delete fixed asset comment apply - DELETE /fixedassets/{fixed_asset_id}/comments/{comment_id} (delete_fixed_asset_comment) [intent=reverse_etl availability=partial write=delete_fixed_asset_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --fixed-asset-id (required), --organization-id
    delete fixed asset type apply - DELETE /fixedassettypes/{fixed_asset_type_id} (delete_fixed_asset_type) [intent=reverse_etl availability=partial write=delete_fixed_asset_type]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-type-id (required), --organization-id
    delete invoice applied credit apply - DELETE /invoices/{invoice_id}/creditsapplied/{creditnotes_invoice_id} (delete_invoice_applied_credit) [intent=reverse_etl availability=partial write=delete_invoice_applied_credit]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnotes-invoice-id (required), --invoice-id (required), --organization-id
    delete invoice apply - Typed action delete_invoice [intent=reverse_etl availability=partial write=delete_invoice]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for delete_invoice; found 0.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for delete_invoice; found 0.; flags: --invoice-id (required)
    delete invoice attachment apply - DELETE /invoices/{invoice_id}/attachment (delete_invoice_attachment) [intent=reverse_etl availability=partial write=delete_invoice_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    delete invoice comment apply - DELETE /invoices/{invoice_id}/comments/{comment_id} (delete_invoice_comment) [intent=reverse_etl availability=partial write=delete_invoice_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --invoice-id (required), --organization-id
    delete invoice document apply - DELETE /invoices/{invoice_id}/documents/{document_id} (delete_invoice_document) [intent=reverse_etl availability=partial write=delete_invoice_document]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --document-id (required), --invoice-id (required), --organization-id
    delete invoice einvoice status apply - DELETE /invoices/{invoice_id}/einvoice/status (delete_invoice_einvoice_status) [intent=reverse_etl availability=partial write=delete_invoice_einvoice_status]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    delete invoice expense receipt apply - DELETE /invoices/expenses/{expense_id}/receipt (delete_invoice_expense_receipt) [intent=reverse_etl availability=partial write=delete_invoice_expense_receipt]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --expense-id (required), --organization-id
    delete invoice line item apply - DELETE /invoices/{invoice_id}/lineitems/{line_item_id} (delete_invoice_line_item) [intent=reverse_etl availability=partial write=delete_invoice_line_item]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --line-item-id (required), --organization-id
    delete invoice of credit note apply - DELETE /creditnotes/{creditnote_id}/invoices/{creditnote_invoice_id} (delete_invoice_of_credit_note) [intent=reverse_etl availability=partial write=delete_invoice_of_credit_note]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --creditnote-invoice-id (required), --organization-id
    delete invoice payment apply - DELETE /invoices/{invoice_id}/payments/{invoice_payment_id} (delete_invoice_payment) [intent=reverse_etl availability=partial write=delete_invoice_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --invoice-payment-id (required), --organization-id
    delete invoice substatus apply - DELETE /invoices/{invoice_id}/substatus/{substatus_id} (delete_invoice_substatus) [intent=reverse_etl availability=partial write=delete_invoice_substatus]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --substatus-id (required), --organization-id
    delete invoices apply - DELETE /invoices (delete_invoices) [intent=reverse_etl availability=partial write=delete_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    delete item apply - DELETE /items/{item_id} (delete_item) [intent=reverse_etl availability=partial write=delete_item]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --item-id (required), --organization-id
    delete journal apply - DELETE /journals/{journal_id} (delete_journal) [intent=reverse_etl availability=partial write=delete_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --organization-id
    delete journal comment apply - DELETE /journals/{journal_id}/comments/{comment_id} (delete_journal_comment) [intent=reverse_etl availability=partial write=delete_journal_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --journal-id (required), --organization-id
    delete journal credits payables apply - DELETE /journals/{journal_id}/credits/{journal_line_id}/payables (delete_journal_credits_payables) [intent=reverse_etl availability=partial write=delete_journal_credits_payables]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --journal-line-id (required), --organization-id
    delete journal credits receivables apply - DELETE /journals/{journal_id}/credits/{journal_line_id}/receivables (delete_journal_credits_receivables) [intent=reverse_etl availability=partial write=delete_journal_credits_receivables]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --journal-line-id (required), --organization-id
    delete last imported bank statement apply - DELETE /bankaccounts/{account_id}/statement/{statement_id} (delete_last_imported_bank_statement) [intent=reverse_etl availability=partial write=delete_last_imported_bank_statement]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --statement-id (required), --organization-id
    delete location apply - DELETE /locations/{location_id} (delete_location) [intent=reverse_etl availability=partial write=delete_location]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --location-id (required), --organization-id
    delete opening balance apply - DELETE /settings/openingbalances (delete_opening_balance) [intent=reverse_etl availability=partial write=delete_opening_balance]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    delete organization address apply - DELETE /organizations/address/{address_id} (delete_organization_address) [intent=reverse_etl availability=implemented write=delete_organization_address]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --address-id (required)
    delete pricebook apply - DELETE /pricebooks/{pricebook_id} (delete_pricebook) [intent=reverse_etl availability=partial write=delete_pricebook]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --pricebook-id (required), --organization-id
    delete project apply - DELETE /projects/{project_id} (delete_project) [intent=reverse_etl availability=partial write=delete_project]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --organization-id
    delete project comment apply - DELETE /projects/{project_id}/comments/{comment_id} (delete_project_comment) [intent=reverse_etl availability=partial write=delete_project_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --project-id (required), --organization-id
    delete project task apply - DELETE /projects/{project_id}/tasks/{task_id} (delete_project_task) [intent=reverse_etl availability=partial write=delete_project_task]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --task-id (required), --organization-id
    delete project user apply - DELETE /projects/{project_id}/users/{user_id} (delete_project_user) [intent=reverse_etl availability=partial write=delete_project_user]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --user-id (required), --organization-id
    delete purchase order apply - DELETE /purchaseorders/{purchaseorder_id} (delete_purchase_order) [intent=reverse_etl availability=partial write=delete_purchase_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    delete purchase order attachment apply - DELETE /purchaseorders/{purchaseorder_id}/attachment (delete_purchase_order_attachment) [intent=reverse_etl availability=partial write=delete_purchase_order_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    delete purchase order comment apply - DELETE /purchaseorders/{purchaseorder_id}/comments/{comment_id} (delete_purchase_order_comment) [intent=reverse_etl availability=partial write=delete_purchase_order_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --purchaseorder-id (required), --organization-id
    delete recurring bill apply - DELETE /recurring_bills/{recurring_bill_id} (delete_recurring_bill) [intent=reverse_etl availability=partial write=delete_recurring_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-bill-id (required), --organization-id
    delete recurring expense apply - DELETE /recurringexpenses/{recurring_expense_id} (delete_recurring_expense) [intent=reverse_etl availability=partial write=delete_recurring_expense]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-expense-id (required), --organization-id
    delete recurring invoice apply - DELETE /recurringinvoices/{recurring_invoice_id} (delete_recurring_invoice) [intent=reverse_etl availability=partial write=delete_recurring_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-id (required), --organization-id
    delete recurring invoice bank account apply - DELETE /recurringinvoices/{recurring_invoice_id}/bankaccount (delete_recurring_invoice_bank_account) [intent=reverse_etl availability=partial write=delete_recurring_invoice_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-id (required), --organization-id
    delete recurring invoice card apply - DELETE /recurringinvoices/{recurring_invoice_id}/card (delete_recurring_invoice_card) [intent=reverse_etl availability=partial write=delete_recurring_invoice_card]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-id (required), --organization-id
    delete recurring journal apply - DELETE /recurringjournals/{recurring_journal_id} (delete_recurring_journal) [intent=reverse_etl availability=partial write=delete_recurring_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-journal-id (required), --organization-id
    delete retainer invoice apply - DELETE /retainerinvoices/{retainerinvoice_id} (delete_retainer_invoice) [intent=reverse_etl availability=partial write=delete_retainer_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    delete retainer invoice attachment apply - DELETE /retainerinvoices/{retainerinvoice_id}/documents/{document_id} (delete_retainer_invoice_attachment) [intent=reverse_etl availability=partial write=delete_retainer_invoice_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --document-id (required), --retainerinvoice-id (required), --organization-id
    delete retainer invoice comment apply - DELETE /retainerinvoices/{retainerinvoice_id}/comments/{comment_id} (delete_retainer_invoice_comment) [intent=reverse_etl availability=partial write=delete_retainer_invoice_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --retainerinvoice-id (required), --organization-id
    delete sales order apply - DELETE /salesorders/{salesorder_id} (delete_sales_order) [intent=reverse_etl availability=partial write=delete_sales_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    delete sales order attachment apply - DELETE /salesorders/{salesorder_id}/attachment (delete_sales_order_attachment) [intent=reverse_etl availability=partial write=delete_sales_order_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    delete sales order comment apply - DELETE /salesorders/{salesorder_id}/comments/{comment_id} (delete_sales_order_comment) [intent=reverse_etl availability=partial write=delete_sales_order_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --salesorder-id (required), --organization-id
    delete sales receipt apply - DELETE /salesreceipts/{sales_receipt_id} (delete_sales_receipt) [intent=reverse_etl availability=partial write=delete_sales_receipt]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --sales-receipt-id (required), --organization-id
    delete tag apply - DELETE /reportingtags/{tag_id} (delete_tag) [intent=reverse_etl availability=partial write=delete_tag]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tag-id (required), --organization-id
    delete task apply - DELETE /tasks/{task_id} (delete_task) [intent=reverse_etl availability=partial write=delete_task]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --task-id (required), --organization-id
    delete task comment apply - DELETE /tasks/{task_id}/comments/{comment_id} (delete_task_comment) [intent=reverse_etl availability=partial write=delete_task_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --task-id (required), --organization-id
    delete task document apply - DELETE /tasks/{task_id}/documents/{document_id} (delete_task_document) [intent=reverse_etl availability=partial write=delete_task_document]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --document-id (required), --task-id (required), --organization-id
    delete tasks apply - DELETE /tasks (delete_tasks) [intent=reverse_etl availability=partial write=delete_tasks]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --task-ids (required), --organization-id
    delete tax apply - DELETE /settings/taxes/{tax_id} (delete_tax) [intent=reverse_etl availability=partial write=delete_tax]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tax-id (required), --organization-id
    delete tax authority apply - DELETE /settings/taxauthorities/{tax_authority_id} (delete_tax_authority) [intent=reverse_etl availability=partial write=delete_tax_authority]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tax-authority-id (required), --organization-id
    delete tax exemption apply - DELETE /settings/taxexemptions/{tax_exemption_id} (delete_tax_exemption) [intent=reverse_etl availability=partial write=delete_tax_exemption]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tax-exemption-id (required), --organization-id
    delete tax group apply - DELETE /settings/taxgroups/{tax_group_id} (delete_tax_group) [intent=reverse_etl availability=partial write=delete_tax_group]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tax-group-id (required), --organization-id
    delete time entries apply - DELETE /projects/timeentries (delete_time_entries) [intent=reverse_etl availability=partial write=delete_time_entries]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    delete time entry apply - DELETE /projects/timeentries/{time_entry_id} (delete_time_entry) [intent=reverse_etl availability=partial write=delete_time_entry]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --time-entry-id (required), --organization-id
    delete transaction lock apply - DELETE /transactionlock (delete_transaction_lock) [intent=reverse_etl availability=partial write=delete_transaction_lock]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-lock-id (required), --organization-id
    delete user apply - DELETE /users/{user_id} (delete_user) [intent=reverse_etl availability=partial write=delete_user]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --user-id (required), --organization-id
    delete vendor credit apply - DELETE /vendorcredits/{vendor_credit_id} (delete_vendor_credit) [intent=reverse_etl availability=partial write=delete_vendor_credit]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --organization-id
    delete vendor credit bill apply - DELETE /vendorcredits/{vendor_credit_id}/bills/{vendor_credit_bill_id} (delete_vendor_credit_bill) [intent=reverse_etl availability=partial write=delete_vendor_credit_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-bill-id (required), --vendor-credit-id (required), --organization-id
    delete vendor credit comment apply - DELETE /vendorcredits/{vendor_credit_id}/comments/{comment_id} (delete_vendor_credit_comment) [intent=reverse_etl availability=partial write=delete_vendor_credit_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --vendor-credit-id (required), --organization-id
    delete vendor credit refund apply - DELETE /vendorcredits/{vendor_credit_id}/refunds/{vendor_credit_refund_id} (delete_vendor_credit_refund) [intent=reverse_etl availability=partial write=delete_vendor_credit_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --vendor-credit-refund-id (required), --organization-id
    delete vendor payment apply - DELETE /vendorpayments/{payment_id} (delete_vendor_payment) [intent=reverse_etl availability=partial write=delete_vendor_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --payment-id (required), --organization-id
    delete vendor payment refund apply - DELETE /vendorpayments/{payment_id}/refunds/{vendorpayment_refund_id} (delete_vendor_payment_refund) [intent=reverse_etl availability=partial write=delete_vendor_payment_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external destructive mutation in Zoho Books; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --payment-id (required), --vendorpayment-refund-id (required), --organization-id
    disable contact payment reminder apply - POST /contacts/{contact_id}/paymentreminder/disable (disable_contact_payment_reminder) [intent=reverse_etl availability=partial write=disable_contact_payment_reminder]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    disable contact person sms apply - POST /contacts/contactpersons/{contactperson_id}/sms/disable (disable_contact_person_sms) [intent=reverse_etl availability=partial write=disable_contact_person_sms]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contactperson-id (required), --organization-id
    disable contact portal apply - POST /contacts/{contact_id}/portal/disable (disable_contact_portal) [intent=reverse_etl availability=partial write=disable_contact_portal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    disable invoice payment reminder apply - POST /invoices/{invoice_id}/paymentreminder/disable (disable_invoice_payment_reminder) [intent=reverse_etl availability=partial write=disable_invoice_payment_reminder]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    disable recurring invoice autobill apply - POST /recurringinvoices/{recurring_invoice_id}/autobill/disable (disable_recurring_invoice_autobill) [intent=reverse_etl availability=partial write=disable_recurring_invoice_autobill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-id (required), --organization-id
    downgrade organization to invoice apply - POST /organizations/{organization_id}/downgradetoinvoice (downgrade_organization_to_invoice) [intent=reverse_etl availability=implemented write=downgrade_organization_to_invoice]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --organization-id (required)
    email contact apply - POST /contacts/{contact_id}/email (email_contact) [intent=reverse_etl availability=partial write=email_contact]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    email contact statement apply - POST /contacts/{contact_id}/statements/email (email_contact_statement) [intent=reverse_etl availability=partial write=email_contact_statement]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    email credit note apply - POST /creditnotes/{creditnote_id}/email (email_credit_note) [intent=reverse_etl availability=partial write=email_credit_note]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    email estimate apply - POST /estimates/{estimate_id}/email (email_estimate) [intent=reverse_etl availability=partial write=email_estimate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    email invoice apply - POST /invoices/{invoice_id}/email (email_invoice) [intent=reverse_etl availability=partial write=email_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    email invoices apply - POST /invoices/email (email_invoices) [intent=reverse_etl availability=partial write=email_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    email multiple estimates apply - POST /estimates/email (email_multiple_estimates) [intent=reverse_etl availability=partial write=email_multiple_estimates]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-ids (required), --organization-id
    email purchase order apply - POST /purchaseorders/{purchaseorder_id}/email (email_purchase_order) [intent=reverse_etl availability=partial write=email_purchase_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    email retainer invoice apply - POST /retainerinvoices/{retainerinvoice_id}/email (email_retainer_invoice) [intent=reverse_etl availability=partial write=email_retainer_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    email sales order apply - POST /salesorders/{salesorder_id}/email (email_sales_order) [intent=reverse_etl availability=partial write=email_sales_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    email sales receipt apply - POST /salesreceipts/{sales_receipt_id}/email (email_sales_receipt) [intent=reverse_etl availability=implemented write=email_sales_receipt]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --sales-receipt-id (required)
    email vendor payment apply - POST /vendorpayments/{payment_id}/email (email_vendor_payment) [intent=reverse_etl availability=partial write=email_vendor_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --payment-id (required), --organization-id
    enable contact payment reminder apply - POST /contacts/{contact_id}/paymentreminder/enable (enable_contact_payment_reminder) [intent=reverse_etl availability=partial write=enable_contact_payment_reminder]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    enable contact person sms apply - POST /contacts/contactpersons/{contactperson_id}/sms/enable (enable_contact_person_sms) [intent=reverse_etl availability=partial write=enable_contact_person_sms]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contactperson-id (required), --organization-id
    enable contact portal apply - POST /contacts/{contact_id}/portal/enable (enable_contact_portal) [intent=reverse_etl availability=partial write=enable_contact_portal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    enable invoice payment reminder apply - POST /invoices/{invoice_id}/paymentreminder/enable (enable_invoice_payment_reminder) [intent=reverse_etl availability=partial write=enable_invoice_payment_reminder]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    enable locations apply - POST /settings/locations/enable (enable_locations) [intent=reverse_etl availability=partial write=enable_locations]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    enable recurring invoice autobill apply - POST /recurringinvoices/{recurring_invoice_id}/autobill/enable (enable_recurring_invoice_autobill) [intent=reverse_etl availability=partial write=enable_recurring_invoice_autobill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-id (required), --organization-id
    exclude bank transaction apply - POST /banktransactions/uncategorized/{transaction_id}/exclude (exclude_bank_transaction) [intent=reverse_etl availability=partial write=exclude_bank_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    fetch credit note einvoice apply - POST /creditnotes/{creditnote_id}/einvoice/fetch (fetch_credit_note_einvoice) [intent=reverse_etl availability=partial write=fetch_credit_note_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    fetch invoice einvoice apply - POST /invoices/{invoice_id}/einvoice/fetch (fetch_invoice_einvoice) [intent=reverse_etl availability=partial write=fetch_invoice_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    finalize credit note approval apply - POST /creditnotes/{creditnote_id}/approve/final (finalize_credit_note_approval) [intent=reverse_etl availability=partial write=finalize_credit_note_approval]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    finalize invoice approval apply - POST /invoices/{invoice_id}/approve/final (finalize_invoice_approval) [intent=reverse_etl availability=partial write=finalize_invoice_approval]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    force pay invoice apply - POST /invoices/{invoice_id}/forcepay (force_pay_invoice) [intent=reverse_etl availability=partial write=force_pay_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    import bank statements apply - POST /bankstatements (import_bank_statements) [intent=reverse_etl availability=partial write=import_bank_statements]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    import customer using crm account id apply - POST /crm/account/{crm_account_id}/import (import_customer_using_crm_account_id) [intent=reverse_etl availability=partial write=import_customer_using_crm_account_id]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --crm-account-id (required), --organization-id
    import customer using crm contact id apply - POST /crm/contact/{crm_contact_id}/import (import_customer_using_crm_contact_id) [intent=reverse_etl availability=partial write=import_customer_using_crm_contact_id]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --crm-contact-id (required), --organization-id
    import item using crm product id apply - POST /crm/item/{crm_product_id}/import (import_item_using_crm_product_id) [intent=reverse_etl availability=partial write=import_item_using_crm_product_id]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --crm-product-id (required), --organization-id
    import vendor using crm vendor id apply - POST /crm/vendor/{crm_vendor_id}/import (import_vendor_using_crm_vendor_id) [intent=reverse_etl availability=partial write=import_vendor_using_crm_vendor_id]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --crm-vendor-id (required), --organization-id
    inactive tag apply - POST /reportingtags/{tag_id}/inactive (inactive_tag) [intent=reverse_etl availability=partial write=inactive_tag]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tag-id (required), --organization-id
    inactive tag option apply - POST /reportingtags/{tag_id}/option/(\d+)/inactive (inactive_tag_option) [intent=reverse_etl availability=partial write=inactive_tag_option]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tag-id (required), --organization-id
    invite contact person to portal apply - POST /contacts/contactpersons/{contactperson_id}/portal/invite (invite_contact_person_to_portal) [intent=reverse_etl availability=partial write=invite_contact_person_to_portal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contactperson-id (required), --organization-id
    invite project user apply - POST /projects/{project_id}/users/invite (invite_project_user) [intent=reverse_etl availability=partial write=invite_project_user]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --organization-id
    invite user apply - POST /users/{user_id}/invite (invite_user) [intent=reverse_etl availability=partial write=invite_user]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --user-id (required), --organization-id
    mail invoice pdf apply - POST /invoices/{invoice_id}/mailpdf (mail_invoice_pdf) [intent=reverse_etl availability=partial write=mail_invoice_pdf]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    map invoice with salesorder apply - POST /invoices/mapwithorder (map_invoice_with_salesorder) [intent=reverse_etl availability=partial write=map_invoice_with_salesorder]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    mark bank account active apply - POST /bankaccounts/{account_id}/active (mark_bank_account_active) [intent=reverse_etl availability=partial write=mark_bank_account_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    mark bank account inactive apply - POST /bankaccounts/{account_id}/inactive (mark_bank_account_inactive) [intent=reverse_etl availability=partial write=mark_bank_account_inactive]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    mark bill open apply - POST /bills/{bill_id}/status/open (mark_bill_open) [intent=reverse_etl availability=partial write=mark_bill_open]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    mark bill void apply - POST /bills/{bill_id}/status/void (mark_bill_void) [intent=reverse_etl availability=partial write=mark_bill_void]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    mark chart of account active apply - POST /chartofaccounts/{account_id}/active (mark_chart_of_account_active) [intent=reverse_etl availability=partial write=mark_chart_of_account_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    mark chart of account inactive apply - POST /chartofaccounts/{account_id}/inactive (mark_chart_of_account_inactive) [intent=reverse_etl availability=partial write=mark_chart_of_account_inactive]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    mark contact active apply - POST /contacts/{contact_id}/active (mark_contact_active) [intent=reverse_etl availability=partial write=mark_contact_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    mark contact address as billing apply - POST /contacts/{contact_id}/address/{address_id}/markasbilling (mark_contact_address_as_billing) [intent=reverse_etl availability=partial write=mark_contact_address_as_billing]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --address-id (required), --contact-id (required), --organization-id
    mark contact address as shipping apply - POST /contacts/{contact_id}/address/{address_id}/markasshipping (mark_contact_address_as_shipping) [intent=reverse_etl availability=partial write=mark_contact_address_as_shipping]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --address-id (required), --contact-id (required), --organization-id
    mark contact inactive apply - POST /contacts/{contact_id}/inactive (mark_contact_inactive) [intent=reverse_etl availability=partial write=mark_contact_inactive]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    mark contact person primary 2 apply - Typed action mark_contact_person_primary_2 [intent=reverse_etl availability=partial write=mark_contact_person_primary_2]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for mark_contact_person_primary_2; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for mark_contact_person_primary_2; found 0.; flags: --contactperson-id (required)
    mark contact person primary apply - Typed action mark_contact_person_primary [intent=reverse_etl availability=partial write=mark_contact_person_primary]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for mark_contact_person_primary; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for mark_contact_person_primary; found 0.; flags: --contact-person-id (required)
    mark contacts for 1099 tracking apply - POST /contacts/track1099 (mark_contacts_for_1099_tracking) [intent=reverse_etl availability=partial write=mark_contacts_for_1099_tracking]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    mark credit note draft apply - POST /creditnotes/{creditnote_id}/status/draft (mark_credit_note_draft) [intent=reverse_etl availability=partial write=mark_credit_note_draft]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    mark credit note einvoice cancelled apply - POST /creditnotes/{creditnote_id}/einvoice/status/cancel (mark_credit_note_einvoice_cancelled) [intent=reverse_etl availability=partial write=mark_credit_note_einvoice_cancelled]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    mark credit note einvoice pushed apply - POST /creditnotes/{creditnote_id}/einvoice/status/push (mark_credit_note_einvoice_pushed) [intent=reverse_etl availability=partial write=mark_credit_note_einvoice_pushed]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    mark credit note open apply - POST /creditnotes/{creditnote_id}/status/open (mark_credit_note_open) [intent=reverse_etl availability=partial write=mark_credit_note_open]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    mark credit note ready to push apply - POST /creditnotes/{creditnote_id}/status/readytopush (mark_credit_note_ready_to_push) [intent=reverse_etl availability=partial write=mark_credit_note_ready_to_push]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    mark credit note void apply - POST /creditnotes/{creditnote_id}/status/void (mark_credit_note_void) [intent=reverse_etl availability=partial write=mark_credit_note_void]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    mark default option apply - POST /reportingtags/{tag_id} (mark_default_option) [intent=reverse_etl availability=partial write=mark_default_option]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --default-option-id (required), --tag-id (required), --organization-id
    mark delivery challan as delivered apply - POST /deliverychallans/{deliverychallan_id}/status/delivered (mark_delivery_challan_as_delivered) [intent=reverse_etl availability=partial write=mark_delivery_challan_as_delivered]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-id (required), --organization-id
    mark delivery challan as open apply - POST /deliverychallans/{deliverychallan_id}/status/open (mark_delivery_challan_as_open) [intent=reverse_etl availability=partial write=mark_delivery_challan_as_open]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-id (required), --organization-id
    mark delivery challan as returned apply - POST /deliverychallans/{deliverychallan_id}/status/returned (mark_delivery_challan_as_returned) [intent=reverse_etl availability=partial write=mark_delivery_challan_as_returned]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-id (required), --organization-id
    mark delivery challan as undelivered apply - POST /deliverychallans/{deliverychallan_id}/status/undelivered (mark_delivery_challan_as_undelivered) [intent=reverse_etl availability=partial write=mark_delivery_challan_as_undelivered]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-id (required), --organization-id
    mark estimate accepted apply - POST /estimates/{estimate_id}/status/accepted (mark_estimate_accepted) [intent=reverse_etl availability=partial write=mark_estimate_accepted]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    mark estimate declined apply - POST /estimates/{estimate_id}/status/declined (mark_estimate_declined) [intent=reverse_etl availability=partial write=mark_estimate_declined]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    mark estimate sent apply - POST /estimates/{estimate_id}/status/sent (mark_estimate_sent) [intent=reverse_etl availability=partial write=mark_estimate_sent]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    mark fixed asset active apply - POST /fixedassets/{fixed_asset_id}/status/active (mark_fixed_asset_active) [intent=reverse_etl availability=partial write=mark_fixed_asset_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-id (required), --organization-id
    mark fixed asset cancel apply - POST /fixedassets/{fixed_asset_id}/status/cancel (mark_fixed_asset_cancel) [intent=reverse_etl availability=partial write=mark_fixed_asset_cancel]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-id (required), --organization-id
    mark fixed asset draft apply - POST /fixedassets/{fixed_asset_id}/status/draft (mark_fixed_asset_draft) [intent=reverse_etl availability=partial write=mark_fixed_asset_draft]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-id (required), --organization-id
    mark invoice draft apply - POST /invoices/{invoice_id}/status/draft (mark_invoice_draft) [intent=reverse_etl availability=partial write=mark_invoice_draft]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    mark invoice einvoice cancelled apply - POST /invoices/{invoice_id}/einvoice/status/cancel (mark_invoice_einvoice_cancelled) [intent=reverse_etl availability=partial write=mark_invoice_einvoice_cancelled]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    mark invoice einvoice pushed apply - POST /invoices/{invoice_id}/einvoice/status/push (mark_invoice_einvoice_pushed) [intent=reverse_etl availability=partial write=mark_invoice_einvoice_pushed]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    mark invoice ready to push apply - POST /invoices/{invoice_id}/status/readytopush (mark_invoice_ready_to_push) [intent=reverse_etl availability=partial write=mark_invoice_ready_to_push]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    mark invoice sent apply - POST /invoices/{invoice_id}/status/sent (mark_invoice_sent) [intent=reverse_etl availability=partial write=mark_invoice_sent]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    mark invoice void apply - POST /invoices/{invoice_id}/status/void (mark_invoice_void) [intent=reverse_etl availability=partial write=mark_invoice_void]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    mark invoices sent apply - POST /invoices/status/sent (mark_invoices_sent) [intent=reverse_etl availability=partial write=mark_invoices_sent]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    mark invoices shipped apply - POST /invoices/markasshipped (mark_invoices_shipped) [intent=reverse_etl availability=partial write=mark_invoices_shipped]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    mark item active apply - POST /items/{item_id}/active (mark_item_active) [intent=reverse_etl availability=partial write=mark_item_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --item-id (required), --organization-id
    mark item inactive apply - POST /items/{item_id}/inactive (mark_item_inactive) [intent=reverse_etl availability=partial write=mark_item_inactive]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --item-id (required), --organization-id
    mark journal published apply - POST /journals/{journal_id}/status/publish (mark_journal_published) [intent=reverse_etl availability=partial write=mark_journal_published]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --organization-id
    mark location active apply - POST /locations/{location_id}/active (mark_location_active) [intent=reverse_etl availability=partial write=mark_location_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --location-id (required), --organization-id
    mark location inactive apply - POST /locations/{location_id}/inactive (mark_location_inactive) [intent=reverse_etl availability=partial write=mark_location_inactive]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --location-id (required), --organization-id
    mark location primary apply - POST /locations/{location_id}/markasprimary (mark_location_primary) [intent=reverse_etl availability=partial write=mark_location_primary]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --location-id (required), --organization-id
    mark organization inactive apply - POST /organizations/{organization_id}/inactive (mark_organization_inactive) [intent=reverse_etl availability=implemented write=mark_organization_inactive]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --organization-id (required)
    mark pricebook active apply - POST /pricebooks/{pricebook_id}/active (mark_pricebook_active) [intent=reverse_etl availability=partial write=mark_pricebook_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --pricebook-id (required), --organization-id
    mark pricebook inactive apply - POST /pricebooks/{pricebook_id}/inactive (mark_pricebook_inactive) [intent=reverse_etl availability=partial write=mark_pricebook_inactive]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --pricebook-id (required), --organization-id
    mark project active apply - POST /projects/{project_id}/active (mark_project_active) [intent=reverse_etl availability=partial write=mark_project_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --organization-id
    mark project inactive apply - POST /projects/{project_id}/inactive (mark_project_inactive) [intent=reverse_etl availability=partial write=mark_project_inactive]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --organization-id
    mark purchase order billed apply - POST /purchaseorders/{purchaseorder_id}/status/billed (mark_purchase_order_billed) [intent=reverse_etl availability=partial write=mark_purchase_order_billed]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    mark purchase order cancelled apply - POST /purchaseorders/{purchaseorder_id}/status/cancelled (mark_purchase_order_cancelled) [intent=reverse_etl availability=partial write=mark_purchase_order_cancelled]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    mark purchase order open apply - POST /purchaseorders/{purchaseorder_id}/status/open (mark_purchase_order_open) [intent=reverse_etl availability=partial write=mark_purchase_order_open]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    mark retainer invoice draft apply - POST /retainerinvoices/{retainerinvoice_id}/status/draft (mark_retainer_invoice_draft) [intent=reverse_etl availability=partial write=mark_retainer_invoice_draft]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    mark retainer invoice sent apply - POST /retainerinvoices/{retainerinvoice_id}/status/sent (mark_retainer_invoice_sent) [intent=reverse_etl availability=partial write=mark_retainer_invoice_sent]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    mark retainer invoice void apply - POST /retainerinvoices/{retainerinvoice_id}/status/void (mark_retainer_invoice_void) [intent=reverse_etl availability=partial write=mark_retainer_invoice_void]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    mark sales order as open apply - POST /salesorders/{salesorder_id}/status/open (mark_sales_order_as_open) [intent=reverse_etl availability=partial write=mark_sales_order_as_open]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    mark sales order as void apply - POST /salesorders/{salesorder_id}/status/void (mark_sales_order_as_void) [intent=reverse_etl availability=partial write=mark_sales_order_as_void]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    mark task as completed apply - POST /tasks/{task_id}/markascompleted (mark_task_as_completed) [intent=reverse_etl availability=partial write=mark_task_as_completed]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --task-id (required), --organization-id
    mark task as ongoing apply - POST /tasks/{task_id}/markasongoing (mark_task_as_ongoing) [intent=reverse_etl availability=partial write=mark_task_as_ongoing]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --task-id (required), --organization-id
    mark task as open apply - POST /tasks/{task_id}/markasopen (mark_task_as_open) [intent=reverse_etl availability=partial write=mark_task_as_open]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --task-id (required), --organization-id
    mark user active apply - POST /users/{user_id}/active (mark_user_active) [intent=reverse_etl availability=partial write=mark_user_active]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --user-id (required), --organization-id
    mark user inactive apply - POST /users/{user_id}/inactive (mark_user_inactive) [intent=reverse_etl availability=partial write=mark_user_inactive]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --user-id (required), --organization-id
    mark vendor credit open apply - POST /vendorcredits/{vendor_credit_id}/status/open (mark_vendor_credit_open) [intent=reverse_etl availability=partial write=mark_vendor_credit_open]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --organization-id
    mark vendor credit void apply - POST /vendorcredits/{vendor_credit_id}/status/void (mark_vendor_credit_void) [intent=reverse_etl availability=partial write=mark_vendor_credit_void]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --organization-id
    match bank transaction apply - POST /banktransactions/uncategorized/{transaction_id}/match (match_bank_transaction) [intent=reverse_etl availability=partial write=match_bank_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    merge contact apply - POST /contacts/{contact_id}/merge (merge_contact) [intent=reverse_etl availability=partial write=merge_contact]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    preview invoice coupons apply - POST /invoices/coupons/preview (preview_invoice_coupons) [intent=reverse_etl availability=partial write=preview_invoice_coupons]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    push credit note einvoice apply - POST /creditnotes/{creditnote_id}/einvoice/push (push_credit_note_einvoice) [intent=reverse_etl availability=partial write=push_credit_note_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    push credit note refund einvoice apply - POST /creditnotes/{creditnote_id}/refunds/einvoice/push (push_credit_note_refund_einvoice) [intent=reverse_etl availability=partial write=push_credit_note_refund_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    push credit notes einvoice apply - POST /creditnotes/einvoice/push (push_credit_notes_einvoice) [intent=reverse_etl availability=partial write=push_credit_notes_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    push invoice einvoice apply - POST /invoices/{invoice_id}/einvoice/push (push_invoice_einvoice) [intent=reverse_etl availability=partial write=push_invoice_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    push invoices einvoice apply - POST /invoices/einvoice/push (push_invoices_einvoice) [intent=reverse_etl availability=partial write=push_invoices_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    recall credit note einvoice status apply - POST /creditnotes/{creditnote_id}/einvoice/status/recall (recall_credit_note_einvoice_status) [intent=reverse_etl availability=partial write=recall_credit_note_einvoice_status]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    recall invoice einvoice status apply - POST /invoices/{invoice_id}/einvoice/status/recall (recall_invoice_einvoice_status) [intent=reverse_etl availability=partial write=recall_invoice_einvoice_status]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    reevaluate base currency adjustment apply - POST /basecurrencyadjustment/{base_currency_adjustment_id}/reevaluate (reevaluate_base_currency_adjustment) [intent=reverse_etl availability=partial write=reevaluate_base_currency_adjustment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --base-currency-adjustment-id (required), --organization-id
    refund excess vendor payment apply - POST /vendorpayments/{payment_id}/refunds (refund_excess_vendor_payment) [intent=reverse_etl availability=partial write=refund_excess_vendor_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --payment-id (required), --organization-id
    refund vendor credit apply - POST /vendorcredits/{vendor_credit_id}/refunds (refund_vendor_credit) [intent=reverse_etl availability=partial write=refund_vendor_credit]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --organization-id
    reject credit note apply - POST /creditnotes/{creditnote_id}/reject (reject_credit_note) [intent=reverse_etl availability=partial write=reject_credit_note]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    reject invoice apply - POST /invoices/{invoice_id}/reject (reject_invoice) [intent=reverse_etl availability=partial write=reject_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    reject journal apply - POST /journals/{journal_id}/reject (reject_journal) [intent=reverse_etl availability=partial write=reject_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --organization-id
    reject purchase orders apply - POST /purchaseorders/{purchaseorder_id}/reject (reject_purchase_orders) [intent=reverse_etl availability=partial write=reject_purchase_orders]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    remind customer for invoice payment apply - POST /invoices/{invoice_id}/paymentreminder (remind_customer_for_invoice_payment) [intent=reverse_etl availability=partial write=remind_customer_for_invoice_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    remove item from portal apply - POST /items/{item_id}/removefromportal (remove_item_from_portal) [intent=reverse_etl availability=partial write=remove_item_from_portal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --item-id (required), --organization-id
    reorder bank account rules apply - POST /bankaccounts/rules/order (reorder_bank_account_rules) [intent=reverse_etl availability=partial write=reorder_bank_account_rules]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    reorder tags apply - PUT /reportingtags/reorder (reorder_tags) [intent=reverse_etl availability=partial write=reorder_tags]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    resend contact person portal invite apply - POST /contacts/contactpersons/{contactperson_id}/portal/invite/resend (resend_contact_person_portal_invite) [intent=reverse_etl availability=partial write=resend_contact_person_portal_invite]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contactperson-id (required), --organization-id
    restore bank transaction apply - POST /banktransactions/uncategorized/{transaction_id}/restore (restore_bank_transaction) [intent=reverse_etl availability=partial write=restore_bank_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    restore contact documents apply - POST /contacts/documents/restore (restore_contact_documents) [intent=reverse_etl availability=partial write=restore_contact_documents]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    resume recurring bill apply - POST /recurringbills/{recurring_bill_id}/status/resume (resume_recurring_bill) [intent=reverse_etl availability=partial write=resume_recurring_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-bill-id (required), --organization-id
    resume recurring expense apply - POST /recurringexpenses/{recurring_expense_id}/status/resume (resume_recurring_expense) [intent=reverse_etl availability=partial write=resume_recurring_expense]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-expense-id (required), --organization-id
    resume recurring invoice apply - POST /recurringinvoices/{recurring_invoice_id}/status/resume (resume_recurring_invoice) [intent=reverse_etl availability=partial write=resume_recurring_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-id (required), --organization-id
    resume recurring invoices apply - POST /recurringinvoices/status/resume (resume_recurring_invoices) [intent=reverse_etl availability=partial write=resume_recurring_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-ids (required), --organization-id
    resume recurring journal apply - POST /recurringjournals/{recurring_journal_id}/status/resume (resume_recurring_journal) [intent=reverse_etl availability=partial write=resume_recurring_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-journal-id (required), --organization-id
    return delivery challans apply - PUT /deliverychallans/return (return_delivery_challans) [intent=reverse_etl availability=partial write=return_delivery_challans]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-ids (required), --organization-id
    reverse journal apply - POST /journals/{journal_id}/reverse (reverse_journal) [intent=reverse_etl availability=partial write=reverse_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --is-reversal-scheduled (required), --journal-id (required), --reversal-date (required), --organization-id
    save bank reconciliation draft apply - PUT /bankaccounts/{account_id}/reconciliations/{reconciliation_id}/draft (save_bank_reconciliation_draft) [intent=reverse_etl availability=partial write=save_bank_reconciliation_draft]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --reconciliation-id (required), --organization-id
    schedule invoice email apply - POST /invoices/{invoice_id}/email/schedule (schedule_invoice_email) [intent=reverse_etl availability=partial write=schedule_invoice_email]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    sell fixed asset apply - POST /fixedassets/{fixed_asset_id}/sell (sell_fixed_asset) [intent=reverse_etl availability=partial write=sell_fixed_asset]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-id (required), --organization-id
    send contact client review email apply - POST /contacts/{contact_id}/clientreviews/email (send_contact_client_review_email) [intent=reverse_etl availability=partial write=send_contact_client_review_email]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    send contact payment method email apply - POST /contacts/{contact_id}/paymentmethod/email (send_contact_payment_method_email) [intent=reverse_etl availability=partial write=send_contact_payment_method_email]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    send contact sms apply - POST /contacts/{contact_id}/sms (send_contact_sms) [intent=reverse_etl availability=partial write=send_contact_sms]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    send contact vendor statement email apply - POST /contacts/{contact_id}/vendorstatements/email (send_contact_vendor_statement_email) [intent=reverse_etl availability=partial write=send_contact_vendor_statement_email]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    send contacts sms apply - POST /contacts/sms (send_contacts_sms) [intent=reverse_etl availability=partial write=send_contacts_sms]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    send invoice dunning notifications apply - POST /invoices/{invoice_id}/senddunningnotifications (send_invoice_dunning_notifications) [intent=reverse_etl availability=partial write=send_invoice_dunning_notifications]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    send invoice retry sms apply - POST /invoices/{invoice_id}/sendretrysms (send_invoice_retry_sms) [intent=reverse_etl availability=partial write=send_invoice_retry_sms]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    send invoice sms apply - POST /invoices/{invoice_id}/sms (send_invoice_sms) [intent=reverse_etl availability=partial write=send_invoice_sms]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    send invoice via snail mail apply - POST /invoices/{invoice_id}/snailmail (send_invoice_via_snail_mail) [intent=reverse_etl availability=partial write=send_invoice_via_snail_mail]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    skip suggested bank account rule apply - POST /bankaccounts/rules/skipsuggest (skip_suggested_bank_account_rule) [intent=reverse_etl availability=partial write=skip_suggested_bank_account_rule]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    start entry timer apply - POST /projects/timeentries/{time_entry_id}/timer/start (start_entry_timer) [intent=reverse_etl availability=partial write=start_entry_timer]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --time-entry-id (required), --organization-id
    stop entry timer apply - POST /projects/timeentries/timer/stop (stop_entry_timer) [intent=reverse_etl availability=partial write=stop_entry_timer]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    stop recurring bill apply - POST /recurringbills/{recurring_bill_id}/status/stop (stop_recurring_bill) [intent=reverse_etl availability=partial write=stop_recurring_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-bill-id (required), --organization-id
    stop recurring expense apply - POST /recurringexpenses/{recurring_expense_id}/status/stop (stop_recurring_expense) [intent=reverse_etl availability=partial write=stop_recurring_expense]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-expense-id (required), --organization-id
    stop recurring invoice apply - POST /recurringinvoices/{recurring_invoice_id}/status/stop (stop_recurring_invoice) [intent=reverse_etl availability=partial write=stop_recurring_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-id (required), --organization-id
    stop recurring invoices apply - POST /recurringinvoices/status/stop (stop_recurring_invoices) [intent=reverse_etl availability=partial write=stop_recurring_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-ids (required), --organization-id
    stop recurring journal apply - POST /recurringjournals/{recurring_journal_id}/status/stop (stop_recurring_journal) [intent=reverse_etl availability=partial write=stop_recurring_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-journal-id (required), --organization-id
    submit bill apply - POST /bills/{bill_id}/submit (submit_bill) [intent=reverse_etl availability=partial write=submit_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    submit credit note apply - POST /creditnotes/{creditnote_id}/submit (submit_credit_note) [intent=reverse_etl availability=partial write=submit_credit_note]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    submit credit notes apply - POST /creditnotes/submit (submit_credit_notes) [intent=reverse_etl availability=partial write=submit_credit_notes]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-ids (required), --organization-id
    submit estimate apply - POST /estimates/{estimate_id}/submit (submit_estimate) [intent=reverse_etl availability=partial write=submit_estimate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    submit invoice apply - POST /invoices/{invoice_id}/submit (submit_invoice) [intent=reverse_etl availability=partial write=submit_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    submit invoices apply - POST /invoices/submit (submit_invoices) [intent=reverse_etl availability=partial write=submit_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    submit journal for approval apply - POST /journals/{journal_id}/submit (submit_journal_for_approval) [intent=reverse_etl availability=partial write=submit_journal_for_approval]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --organization-id
    submit purchase order apply - POST /purchaseorders/{purchaseorder_id}/submit (submit_purchase_order) [intent=reverse_etl availability=partial write=submit_purchase_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    submit retainer invoice apply - POST /retainerinvoices/{retainerinvoice_id}/submit (submit_retainer_invoice) [intent=reverse_etl availability=partial write=submit_retainer_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    submit sales order apply - POST /salesorders/{salesorder_id}/submit (submit_sales_order) [intent=reverse_etl availability=partial write=submit_sales_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    submit vendor credit apply - POST /vendorcredits/{vendor_credit_id}/submit (submit_vendor_credit) [intent=reverse_etl availability=partial write=submit_vendor_credit]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --organization-id
    track contact 1099 apply - POST /contacts/{contact_id}/track1099 (track_contact_1099) [intent=reverse_etl availability=partial write=track_contact_1099]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    uncategorize bank transaction apply - POST /banktransactions/{transaction_id}/uncategorize (uncategorize_bank_transaction) [intent=reverse_etl availability=partial write=uncategorize_bank_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    undo return delivery challans apply - PUT /deliverychallans/undo/return (undo_return_delivery_challans) [intent=reverse_etl availability=partial write=undo_return_delivery_challans]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-ids (required), --organization-id
    unmap invoices from salesorders apply - PUT /invoices/unmap/salesorders (unmap_invoices_from_salesorders) [intent=reverse_etl availability=partial write=unmap_invoices_from_salesorders]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    unmatch bank transaction apply - POST /banktransactions/{transaction_id}/unmatch (unmatch_bank_transaction) [intent=reverse_etl availability=partial write=unmatch_bank_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --transaction-id (required), --organization-id
    unship invoices apply - POST /invoices/unship (unship_invoices) [intent=reverse_etl availability=partial write=unship_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    untrack contact 1099 apply - POST /contacts/{contact_id}/untrack1099 (untrack_contact_1099) [intent=reverse_etl availability=partial write=untrack_contact_1099]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    update a task apply - PUT /tasks/{task_id} (update_a_task) [intent=reverse_etl availability=partial write=update_a_task]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --task-id (required), --organization-id
    update bank account apply - PUT /bankaccounts/{account_id} (update_bank_account) [intent=reverse_etl availability=partial write=update_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    update bank account match filter apply - PUT /bankaccounts/matchfilters/{match_filter_id} (update_bank_account_match_filter) [intent=reverse_etl availability=partial write=update_bank_account_match_filter]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --match-filter-id (required), --organization-id
    update bank account preferences apply - PUT /bankaccounts/{account_id}/preferences (update_bank_account_preferences) [intent=reverse_etl availability=partial write=update_bank_account_preferences]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    update bank account rule apply - PUT /bankaccounts/rules/{rule_id} (update_bank_account_rule) [intent=reverse_etl availability=partial write=update_bank_account_rule]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --rule-id (required), --organization-id
    update bank reconciliation apply - PUT /bankaccounts/{account_id}/reconciliations/{reconciliation_id} (update_bank_reconciliation) [intent=reverse_etl availability=partial write=update_bank_reconciliation]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --reconciliation-id (required), --organization-id
    update bank transaction apply - PUT /banktransactions/{bank_transaction_id} (update_bank_transaction) [intent=reverse_etl availability=partial write=update_bank_transaction]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bank-transaction-id (required), --organization-id
    update bill apply - PUT /bills/{bill_id} (update_bill) [intent=reverse_etl availability=partial write=update_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    update bill billing address apply - PUT /bills/{bill_id}/address/billing (update_bill_billing_address) [intent=reverse_etl availability=partial write=update_bill_billing_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    update chart of account apply - PUT /chartofaccounts/{account_id} (update_chart_of_account) [intent=reverse_etl availability=partial write=update_chart_of_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --account-id (required), --organization-id
    update contact address apply - PUT /contacts/{contact_id}/address/{address_id} (update_contact_address) [intent=reverse_etl availability=partial write=update_contact_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --address-id (required), --contact-id (required), --organization-id
    update contact apply - PUT /contacts/{contact_id} (update_contact) [intent=reverse_etl availability=partial write=update_contact]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    update contact bank account apply - PUT /contacts/{contact_id}/bankaccount/{bank_account_id} (update_contact_bank_account) [intent=reverse_etl availability=partial write=update_contact_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bank-account-id (required), --contact-id (required), --organization-id
    update contact card apply - PUT /contacts/{contact_id}/card/{card_id} (update_contact_card) [intent=reverse_etl availability=partial write=update_contact_card]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --card-id (required), --contact-id (required), --organization-id
    update contact document apply - PUT /contacts/{contact_id}/documents/{document_id} (update_contact_document) [intent=reverse_etl availability=partial write=update_contact_document]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --document-id (required), --organization-id
    update contact person 2 apply - Typed action update_contact_person_2 [intent=reverse_etl availability=partial write=update_contact_person_2]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for update_contact_person_2; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for update_contact_person_2; found 0.; flags: --contactperson-id (required)
    update contact person apply - Typed action update_contact_person [intent=reverse_etl availability=partial write=update_contact_person]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for update_contact_person; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for update_contact_person; found 0.; flags: --contact-person-id (required)
    update contact tags apply - PUT /contacts/{contact_id}/tags (update_contact_tags) [intent=reverse_etl availability=partial write=update_contact_tags]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    update contact tax info apply - PUT /contacts/{contact_id}/taxinfo/{tax_info_id} (update_contact_tax_info) [intent=reverse_etl availability=partial write=update_contact_tax_info]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --tax-info-id (required), --organization-id
    update contact trn status apply - POST /contacts/{contact_id}/trnstatus (update_contact_trn_status) [intent=reverse_etl availability=partial write=update_contact_trn_status]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    update credit note apply - PUT /creditnotes/{creditnote_id} (update_credit_note) [intent=reverse_etl availability=partial write=update_credit_note]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    update credit note billing address apply - PUT /creditnotes/{creditnote_id}/address/billing (update_credit_note_billing_address) [intent=reverse_etl availability=partial write=update_credit_note_billing_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    update credit note cfdi status apply - POST /creditnotes/{creditnote_id}/cfdi/status (update_credit_note_cfdi_status) [intent=reverse_etl availability=partial write=update_credit_note_cfdi_status]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    update credit note custom fields apply - POST /creditnotes/{creditnote_id}/customfields (update_credit_note_custom_fields) [intent=reverse_etl availability=partial write=update_credit_note_custom_fields]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    update credit note document apply - PUT /creditnotes/{creditnote_id}/documents/{document_id} (update_credit_note_document) [intent=reverse_etl availability=partial write=update_credit_note_document]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --document-id (required), --organization-id
    update credit note refund apply - PUT /creditnotes/{creditnote_id}/refunds/{creditnote_refund_id} (update_credit_note_refund) [intent=reverse_etl availability=partial write=update_credit_note_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --creditnote-refund-id (required), --organization-id
    update credit note shipping address apply - PUT /creditnotes/{creditnote_id}/address/shipping (update_credit_note_shipping_address) [intent=reverse_etl availability=partial write=update_credit_note_shipping_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    update credit note template apply - PUT /creditnotes/{creditnote_id}/templates/{template_id} (update_credit_note_template) [intent=reverse_etl availability=partial write=update_credit_note_template]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --template-id (required), --organization-id
    update currency apply - PUT /settings/currencies/{currency_id} (update_currency) [intent=reverse_etl availability=partial write=update_currency]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --currency-id (required), --organization-id
    update custom fields in bill apply - PUT /bill/{bill_id}/customfields (update_custom_fields_in_bill) [intent=reverse_etl availability=partial write=update_custom_fields_in_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bill-id (required), --organization-id
    update custom fields in customer payment apply - PUT /customerpayment/{customer_payment_id}/customfields (update_custom_fields_in_customer_payment) [intent=reverse_etl availability=partial write=update_custom_fields_in_customer_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --customer-payment-id (required), --organization-id
    update custom fields in estimate apply - PUT /estimate/{estimate_id}/customfields (update_custom_fields_in_estimate) [intent=reverse_etl availability=partial write=update_custom_fields_in_estimate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    update custom fields in invoice apply - PUT /invoice/{invoice_id}/customfields (update_custom_fields_in_invoice) [intent=reverse_etl availability=partial write=update_custom_fields_in_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    update custom fields in item apply - PUT /item/{item_id}/customfields (update_custom_fields_in_item) [intent=reverse_etl availability=partial write=update_custom_fields_in_item]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --item-id (required), --organization-id
    update custom fields in purchase order apply - PUT /purchaseorder/{purchaseorder_id}/customfields (update_custom_fields_in_purchase_order) [intent=reverse_etl availability=partial write=update_custom_fields_in_purchase_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    update custom module apply - PUT /settings/modules/{module_api_name} (update_custom_module) [intent=reverse_etl availability=partial write=update_custom_module]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --module-api-name (required), --organization-id
    update custom module record apply - PUT /{module_name}/{module_id} (update_custom_module_record) [intent=reverse_etl availability=partial write=update_custom_module_record]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --module-id (required), --module-name (required), --organization-id
    update customer debit note apply - Typed action update_customer_debit_note [intent=reverse_etl availability=partial write=update_customer_debit_note]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for update_customer_debit_note; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for update_customer_debit_note; found 0.; flags: --debit-note-id (required)
    update customer payment apply - PUT /customerpayments/{payment_id} (update_customer_payment) [intent=reverse_etl availability=partial write=update_customer_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --payment-id (required), --organization-id
    update customer payment refund apply - PUT /customerpayments/{customer_payment_id}/refunds/{refund_id} (update_customer_payment_refund) [intent=reverse_etl availability=partial write=update_customer_payment_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --customer-payment-id (required), --refund-id (required), --organization-id
    update delivery challan apply - PUT /deliverychallans/{deliverychallan_id} (update_delivery_challan) [intent=reverse_etl availability=partial write=update_delivery_challan]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-id (required), --organization-id
    update delivery challan shipping address apply - PUT /deliverychallans/{deliverychallan_id}/address/shipping (update_delivery_challan_shipping_address) [intent=reverse_etl availability=partial write=update_delivery_challan_shipping_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-id (required), --organization-id
    update delivery challan template apply - PUT /deliverychallans/{deliverychallan_id}/templates/{template_id} (update_delivery_challan_template) [intent=reverse_etl availability=partial write=update_delivery_challan_template]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --deliverychallan-id (required), --template-id (required), --organization-id
    update estimate apply - PUT /estimates/{estimate_id} (update_estimate) [intent=reverse_etl availability=partial write=update_estimate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    update estimate billing address apply - PUT /estimates/{estimate_id}/address/billing (update_estimate_billing_address) [intent=reverse_etl availability=partial write=update_estimate_billing_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    update estimate comment apply - PUT /estimates/{estimate_id}/comments/{comment_id} (update_estimate_comment) [intent=reverse_etl availability=partial write=update_estimate_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --estimate-id (required), --organization-id
    update estimate shipping address apply - PUT /estimates/{estimate_id}/address/shipping (update_estimate_shipping_address) [intent=reverse_etl availability=partial write=update_estimate_shipping_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --organization-id
    update estimate template apply - PUT /estimates/{estimate_id}/templates/{template_id} (update_estimate_template) [intent=reverse_etl availability=partial write=update_estimate_template]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --estimate-id (required), --template-id (required), --organization-id
    update exchange rate apply - PUT /settings/currencies/{currency_id}/exchangerates/{exchange_rate_id} (update_exchange_rate) [intent=reverse_etl availability=partial write=update_exchange_rate]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --currency-id (required), --exchange-rate-id (required), --organization-id
    update expense apply - PUT /expenses/{expense_id} (update_expense) [intent=reverse_etl availability=partial write=update_expense]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --expense-id (required), --organization-id
    update fixed asset apply - PUT /fixedassets/{fixed_asset_id} (update_fixed_asset) [intent=reverse_etl availability=partial write=update_fixed_asset]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-id (required), --organization-id
    update fixed asset type apply - PUT /fixedassettypes/{fixed_asset_type_id} (update_fixed_asset_type) [intent=reverse_etl availability=partial write=update_fixed_asset_type]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-type-id (required), --organization-id
    update invoice advanced tracking details apply - PUT /invoices/{invoice_id}/advancedtrackingdetails (update_invoice_advanced_tracking_details) [intent=reverse_etl availability=partial write=update_invoice_advanced_tracking_details]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    update invoice apply - Typed action update_invoice [intent=reverse_etl availability=partial write=update_invoice]; approval: Blocked pending a faithful CLI record binding: declaration-pending: exact api_surface binding requires one covered_by.write match for update_invoice; found 0.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: exact api_surface binding requires one covered_by.write match for update_invoice; found 0.; flags: --invoice-id (required)
    update invoice attachment preference apply - PUT /invoices/{invoice_id}/attachment (update_invoice_attachment_preference) [intent=reverse_etl availability=partial write=update_invoice_attachment_preference]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --can-send-in-mail (required), --invoice-id (required), --organization-id
    update invoice billing address apply - PUT /invoices/{invoice_id}/address/billing (update_invoice_billing_address) [intent=reverse_etl availability=partial write=update_invoice_billing_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    update invoice cfdi status apply - POST /invoices/{invoice_id}/cfdi/status (update_invoice_cfdi_status) [intent=reverse_etl availability=partial write=update_invoice_cfdi_status]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    update invoice comment apply - PUT /invoices/{invoice_id}/comments/{comment_id} (update_invoice_comment) [intent=reverse_etl availability=partial write=update_invoice_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --invoice-id (required), --organization-id
    update invoice einvoice payment status apply - PUT /invoices/{invoice_id}/einvoice/paymentstatus (update_invoice_einvoice_payment_status) [intent=reverse_etl availability=partial write=update_invoice_einvoice_payment_status]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    update invoice metadata apply - PUT /invoices/{invoice_id}/metadata/{metadata_name} (update_invoice_metadata) [intent=reverse_etl availability=partial write=update_invoice_metadata]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --metadata-name (required), --organization-id
    update invoice shipping address apply - PUT /invoices/{invoice_id}/address/shipping (update_invoice_shipping_address) [intent=reverse_etl availability=partial write=update_invoice_shipping_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    update invoice template apply - PUT /invoices/{invoice_id}/templates/{template_id} (update_invoice_template) [intent=reverse_etl availability=partial write=update_invoice_template]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --template-id (required), --organization-id
    update item apply - PUT /items/{item_id} (update_item) [intent=reverse_etl availability=partial write=update_item]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --item-id (required), --organization-id
    update journal apply - PUT /journals/{journal_id} (update_journal) [intent=reverse_etl availability=partial write=update_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --journal-id (required), --organization-id
    update location apply - PUT /locations/{location_id} (update_location) [intent=reverse_etl availability=partial write=update_location]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --location-id (required), --organization-id
    update opening balance apply - PUT /settings/openingbalances (update_opening_balance) [intent=reverse_etl availability=partial write=update_opening_balance]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    update organization address apply - PUT /organizations/address/{address_id} (update_organization_address) [intent=reverse_etl availability=implemented write=update_organization_address]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --address-id (required)
    update organization apply - PUT /organizations/{organization_id} (update_organization) [intent=reverse_etl availability=implemented write=update_organization]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --organization-id (required)
    update partial unlock apply - PUT /transactionlock/partialunlock (update_partial_unlock) [intent=reverse_etl availability=partial write=update_partial_unlock]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    update percentage task apply - POST /tasks/{task_id}/percentage (update_percentage_task) [intent=reverse_etl availability=partial write=update_percentage_task]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --task-id (required), --organization-id
    update pricebook apply - PUT /pricebooks/{pricebook_id} (update_pricebook) [intent=reverse_etl availability=partial write=update_pricebook]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --pricebook-id (required), --organization-id
    update project apply - PUT /projects/{project_id} (update_project) [intent=reverse_etl availability=partial write=update_project]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --organization-id
    update project task apply - PUT /projects/{project_id}/tasks/{task_id} (update_project_task) [intent=reverse_etl availability=partial write=update_project_task]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --task-id (required), --organization-id
    update project user apply - PUT /projects/{project_id}/users/{user_id} (update_project_user) [intent=reverse_etl availability=partial write=update_project_user]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --project-id (required), --user-id (required), --organization-id
    update purchase order apply - PUT /purchaseorders/{purchaseorder_id} (update_purchase_order) [intent=reverse_etl availability=partial write=update_purchase_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    update purchase order attachment apply - PUT /purchaseorders/{purchaseorder_id}/attachment (update_purchase_order_attachment) [intent=reverse_etl availability=partial write=update_purchase_order_attachment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --can-send-in-mail (required), --purchaseorder-id (required), --organization-id
    update purchase order billing address apply - PUT /purchaseorders/{purchaseorder_id}/address/billing (update_purchase_order_billing_address) [intent=reverse_etl availability=partial write=update_purchase_order_billing_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --organization-id
    update purchase order comment apply - PUT /purchaseorders/{purchaseorder_id}/comments/{comment_id} (update_purchase_order_comment) [intent=reverse_etl availability=partial write=update_purchase_order_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --purchaseorder-id (required), --organization-id
    update purchase order template apply - PUT /purchaseorders/{purchaseorder_id}/templates/{template_id} (update_purchase_order_template) [intent=reverse_etl availability=partial write=update_purchase_order_template]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --purchaseorder-id (required), --template-id (required), --organization-id
    update recurring bill apply - PUT /recurringbills/{recurring_bill_id} (update_recurring_bill) [intent=reverse_etl availability=partial write=update_recurring_bill]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-bill-id (required), --organization-id
    update recurring expense apply - PUT /recurringexpenses/{recurring_expense_id} (update_recurring_expense) [intent=reverse_etl availability=partial write=update_recurring_expense]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-expense-id (required), --organization-id
    update recurring invoice apply - PUT /recurringinvoices/{recurring_invoice_id} (update_recurring_invoice) [intent=reverse_etl availability=partial write=update_recurring_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-id (required), --organization-id
    update recurring invoice template apply - PUT /recurringinvoices/{recurring_invoice_id}/templates/{template_id} (update_recurring_invoice_template) [intent=reverse_etl availability=partial write=update_recurring_invoice_template]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-invoice-id (required), --template-id (required), --organization-id
    update recurring journal apply - PUT /recurringjournals/{recurring_journal_id} (update_recurring_journal) [intent=reverse_etl availability=partial write=update_recurring_journal]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --recurring-journal-id (required), --organization-id
    update retainer invoice apply - PUT /retainerinvoices/{retainerinvoice_id} (update_retainer_invoice) [intent=reverse_etl availability=partial write=update_retainer_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    update retainer invoice billing address apply - PUT /retainerinvoices/{retainerinvoice_id}/address/billing (update_retainer_invoice_billing_address) [intent=reverse_etl availability=partial write=update_retainer_invoice_billing_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --organization-id
    update retainer invoice comment apply - PUT /retainerinvoices/{retainerinvoice_id}/comments/{comment_id} (update_retainer_invoice_comment) [intent=reverse_etl availability=partial write=update_retainer_invoice_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --retainerinvoice-id (required), --organization-id
    update retainer invoice template apply - PUT /retainerinvoices/{retainerinvoice_id}/templates/{template_id} (update_retainer_invoice_template) [intent=reverse_etl availability=partial write=update_retainer_invoice_template]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --retainerinvoice-id (required), --template-id (required), --organization-id
    update sales order apply - PUT /salesorders/{salesorder_id} (update_sales_order) [intent=reverse_etl availability=partial write=update_sales_order]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    update sales order attachment preference apply - PUT /salesorders/{salesorder_id}/attachment (update_sales_order_attachment_preference) [intent=reverse_etl availability=partial write=update_sales_order_attachment_preference]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --can-send-in-mail (required), --salesorder-id (required), --organization-id
    update sales order billing address apply - PUT /salesorders/{salesorder_id}/address/billing (update_sales_order_billing_address) [intent=reverse_etl availability=partial write=update_sales_order_billing_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    update sales order comment apply - PUT /salesorders/{salesorder_id}/comments/{comment_id} (update_sales_order_comment) [intent=reverse_etl availability=partial write=update_sales_order_comment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --comment-id (required), --salesorder-id (required), --organization-id
    update sales order shipping address apply - PUT /salesorders/{salesorder_id}/address/shipping (update_sales_order_shipping_address) [intent=reverse_etl availability=partial write=update_sales_order_shipping_address]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    update sales order sub status apply - POST /salesorders/{salesorder_id}/substatus/{status_code} (update_sales_order_sub_status) [intent=reverse_etl availability=partial write=update_sales_order_sub_status]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --status-code (required), --organization-id
    update sales order template apply - PUT /salesorders/{salesorder_id}/templates/{template_id} (update_sales_order_template) [intent=reverse_etl availability=partial write=update_sales_order_template]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --template-id (required), --organization-id
    update sales receipt apply - PUT /salesreceipts/{sales_receipt_id} (update_sales_receipt) [intent=reverse_etl availability=partial write=update_sales_receipt]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --sales-receipt-id (required), --organization-id
    update salesorder customfields apply - PUT /salesorder/{salesorder_id}/customfields (update_salesorder_customfields) [intent=reverse_etl availability=partial write=update_salesorder_customfields]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --salesorder-id (required), --organization-id
    update tag apply - PUT /reportingtags/{tag_id} (update_tag) [intent=reverse_etl availability=partial write=update_tag]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tag-id (required), --organization-id
    update tag criteria apply - PUT /reportingtags/{tag_id}/criteria (update_tag_criteria) [intent=reverse_etl availability=partial write=update_tag_criteria]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tag-id (required), --organization-id
    update tag options apply - PUT /reportingtags/{tag_id}/options (update_tag_options) [intent=reverse_etl availability=partial write=update_tag_options]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tag-id (required), --organization-id
    update tasks apply - PUT /tasks (update_tasks) [intent=reverse_etl availability=partial write=update_tasks]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bulk-update (required), --organization-id
    update tax apply - PUT /settings/taxes/{tax_id} (update_tax) [intent=reverse_etl availability=partial write=update_tax]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tax-id (required), --organization-id
    update tax authority apply - PUT /settings/taxauthorities/{tax_authority_id} (update_tax_authority) [intent=reverse_etl availability=partial write=update_tax_authority]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tax-authority-id (required), --organization-id
    update tax exemption apply - PUT /settings/taxexemptions/{tax_exemption_id} (update_tax_exemption) [intent=reverse_etl availability=partial write=update_tax_exemption]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tax-exemption-id (required), --organization-id
    update tax group apply - PUT /settings/taxgroups/{tax_group_id} (update_tax_group) [intent=reverse_etl availability=partial write=update_tax_group]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --tax-group-id (required), --organization-id
    update time entry apply - PUT /projects/timeentries/{time_entry_id} (update_time_entry) [intent=reverse_etl availability=partial write=update_time_entry]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --time-entry-id (required), --organization-id
    update transaction lock apply - PUT /transactionlock (update_transaction_lock) [intent=reverse_etl availability=partial write=update_transaction_lock]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --organization-id
    update user apply - PUT /users/{user_id} (update_user) [intent=reverse_etl availability=partial write=update_user]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --user-id (required), --organization-id
    update vendor credit apply - PUT /vendorcredits/{vendor_credit_id} (update_vendor_credit) [intent=reverse_etl availability=partial write=update_vendor_credit]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --organization-id
    update vendor credit refund apply - PUT /vendorcredits/{vendor_credit_id}/refunds/{vendor_credit_refund_id} (update_vendor_credit_refund) [intent=reverse_etl availability=partial write=update_vendor_credit_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --vendor-credit-id (required), --vendor-credit-refund-id (required), --organization-id
    update vendor payment apply - PUT /vendorpayments/{payment_id} (update_vendor_payment) [intent=reverse_etl availability=partial write=update_vendor_payment]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --payment-id (required), --organization-id
    update vendor payment refund apply - PUT /vendorpayments/{payment_id}/refunds/{vendorpayment_refund_id} (update_vendor_payment_refund) [intent=reverse_etl availability=partial write=update_vendor_payment_refund]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --payment-id (required), --vendorpayment-refund-id (required), --organization-id
    upgrade organization to books apply - POST /organizations/{organization_id}/upgradetobooks (upgrade_organization_to_books) [intent=reverse_etl availability=implemented write=upgrade_organization_to_books]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --organization-id (required)
    upload credit note digital signature apply - POST /creditnotes/{creditnote_id}/dsign/upload (upload_credit_note_digital_signature) [intent=reverse_etl availability=partial write=upload_credit_note_digital_signature]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --creditnote-id (required), --organization-id
    upload invoice digital signature apply - POST /invoices/{invoice_id}/dsign/upload (upload_invoice_digital_signature) [intent=reverse_etl availability=partial write=upload_invoice_digital_signature]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    upload invoice document apply - POST /invoices/{invoice_id}/documents/{document_id}/upload (upload_invoice_document) [intent=reverse_etl availability=partial write=upload_invoice_document]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --document-id (required), --invoice-id (required), --organization-id
    verify contact address by id apply - POST /contacts/{contact_id}/address/{address_id}/verify (verify_contact_address_by_id) [intent=reverse_etl availability=partial write=verify_contact_address_by_id]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --address-id (required), --contact-id (required), --organization-id
    verify contact bank account apply - POST /contacts/{contact_id}/bankaccount/{bank_account_id}/verify (verify_contact_bank_account) [intent=reverse_etl availability=partial write=verify_contact_bank_account]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --bank-account-id (required), --contact-id (required), --organization-id
    verify contact einvoice apply - POST /contacts/{contact_id}/einvoice/verify (verify_contact_einvoice) [intent=reverse_etl availability=partial write=verify_contact_einvoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --contact-id (required), --organization-id
    void invoices apply - POST /invoices/status/void (void_invoices) [intent=reverse_etl availability=partial write=void_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    write off fixed asset apply - POST /fixedassets/{fixed_asset_id}/writeoff (write_off_fixed_asset) [intent=reverse_etl availability=partial write=write_off_fixed_asset]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --fixed-asset-id (required), --organization-id
    write off invoice apply - POST /invoices/{invoice_id}/writeoff (write_off_invoice) [intent=reverse_etl availability=partial write=write_off_invoice]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-id (required), --organization-id
    write off invoices apply - POST /invoices/writeoff (write_off_invoices) [intent=reverse_etl availability=partial write=write_off_invoices]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --invoice-ids (required), --organization-id
    writeoff opening balance apply - POST /openingbalances/{opening_balance_id}/writeoff (writeoff_opening_balance) [intent=reverse_etl availability=partial write=writeoff_opening_balance]; approval: Blocked pending a faithful CLI record binding: foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; risk: external mutation in Zoho Books accounting data; approval required; notes: Generated from the connector-owned typed action; foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record.; flags: --opening-balance-id (required), --organization-id

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zoho-books

  # Inspect as structured JSON
  pm connectors inspect zoho-books --json

AGENT WORKFLOW
  - Run pm connectors inspect zoho-books before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
