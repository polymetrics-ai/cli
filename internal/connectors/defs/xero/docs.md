# Overview

Reads and writes Xero Accounting API resources through data streams and typed
write actions.

Readable streams: `invoices`, `contacts`, `accounts`, `bank_transactions`, `items`, `payments`,
`account`, `account_attachments`, `bank_transaction`, `bank_transaction_attachments`,
`bank_transactions_history`, `bank_transfer`, `bank_transfer_attachments`, `bank_transfer_history`,
`bank_transfers`, `batch_payment`, `batch_payment_history`, `batch_payments`, `branding_theme`,
`branding_theme_payment_services`, `branding_themes`, `budget`, `budgets`, `contact`,
`contact_attachments`, `contact_by_contact_number`, `contact_cis_settings`, `contact_group`,
`contact_groups`, `contact_history`, `credit_note`, `credit_note_attachments`,
`credit_note_history`, `credit_notes`, `currencies`, `expense_claim`, `expense_claim_history`,
`expense_claims`, `invoice`, `invoice_attachments`, `invoice_history`, `invoice_reminders`, `item`,
`item_history`, `journal`, `journal_by_number`, `journals`, `linked_transaction`,
`linked_transactions`, `manual_journal`, `manual_journal_attachments`, `manual_journals`,
`manual_journals_history`, `online_invoice`, `organisation_actions`, `organisation_cis_settings`,
`organisations`, `overpayment`, `overpayment_history`, `overpayments`, `payment`, `payment_history`,
`payment_services`, `prepayment`, `prepayment_history`, `prepayments`, `purchase_order`,
`purchase_order_attachments`, `purchase_order_by_number`, `purchase_order_history`,
`purchase_orders`, `quote`, `quote_attachments`, `quote_history`, `quotes`, `receipt`,
`receipt_attachments`, `receipt_history`, `receipts`, `repeating_invoice`,
`repeating_invoice_attachments`, `repeating_invoice_history`, `repeating_invoices`,
`report_aged_payables_by_contact`, `report_aged_receivables_by_contact`, `report_balance_sheet`,
`report_bank_summary`, `report_budget_summary`, `report_executive_summary`, `report_from_id`,
`report_profit_and_loss`, `report_ten_ninety_nine`, `report_trial_balance`, `reports_list`,
`tax_rate_by_tax_type`, `tax_rates`, `tracking_categories`, `tracking_category`, `user`, `users`.

Write actions: `create_account`, `delete_account`, `update_account`, `upsert_bank_transactions`,
`create_bank_transactions`, `update_bank_transaction`, `create_bank_transaction_history_record`,
`create_bank_transfer`, `create_bank_transfer_history_record`, `delete_batch_payment`,
`create_batch_payment`, `delete_batch_payment_by_url_param`, `create_batch_payment_history_record`,
`create_branding_theme_payment_services`, `create_contact_group`, `update_contact_group`,
`delete_contact_group_contacts`, `create_contact_group_contacts`, `delete_contact_group_contact`,
`upsert_contacts`, `create_contacts`, `update_contact`, `create_contact_history`,
`upsert_credit_notes`, `create_credit_notes`, `update_credit_note`, `create_credit_note_allocation`,
`delete_credit_note_allocations`, `create_credit_note_history`, `create_currency`,
`create_expense_claims`, `update_expense_claim`, `create_expense_claim_history`, `upsert_invoices`,
`create_invoices`, `update_invoice`, `email_invoice`, `create_invoice_history`, `upsert_items`,
`create_items`, `delete_item`, `update_item`, `create_item_history`, `create_linked_transaction`,
`delete_linked_transaction`, `update_linked_transaction`, `upsert_manual_journals`,
`create_manual_journals`, `update_manual_journal`, `create_manual_journal_history_record`,
`create_overpayment_allocations`, `delete_overpayment_allocations`, `create_overpayment_history`,
`create_payment`, `create_payments`, `delete_payment`, `create_payment_history`,
`create_payment_service`, `create_prepayment_allocations`, `delete_prepayment_allocations`,
`create_prepayment_history`, `upsert_purchase_orders`, `create_purchase_orders`,
`update_purchase_order`, `create_purchase_order_history`, `upsert_quotes`, `create_quotes`,
`update_quote`, `create_quote_history`, `create_receipt`, `update_receipt`,
`create_receipt_history`, `upsert_repeating_invoices`, `create_repeating_invoices`,
`update_repeating_invoice`, `create_repeating_invoice_history`, `setup_organisation`,
`update_tax_rate`, `create_tax_rates`, `create_tracking_category`, and 5 more.

Service API documentation: https://developer.xero.com/documentation/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Xero OAuth2 access token. Sent as a Bearer token; never
  logged.
- `account_id` (optional, string); Xero account id path parameter for docs-derived streams.
- `bank_transaction_id` (optional, string); Xero bank transaction id path parameter for docs-derived
  streams.
- `bank_transfer_id` (optional, string); Xero bank transfer id path parameter for docs-derived
  streams.
- `base_url` (optional, string); default `https://api.xero.com/api.xro/2.0`; format `uri`; Xero
  Accounting API base URL override for tests or proxies.
- `batch_payment_id` (optional, string); Xero batch payment id path parameter for docs-derived
  streams.
- `branding_theme_id` (optional, string); Xero branding theme id path parameter for docs-derived
  streams.
- `budget_id` (optional, string); Xero budget id path parameter for docs-derived streams.
- `contact_group_id` (optional, string); Xero contact group id path parameter for docs-derived
  streams.
- `contact_id` (optional, string); Xero contact id path parameter for docs-derived streams.
- `contact_number` (optional, string); Xero contact number path parameter for docs-derived streams.
- `credit_note_id` (optional, string); Xero credit note id path parameter for docs-derived streams.
- `expense_claim_id` (optional, string); Xero expense claim id path parameter for docs-derived
  streams.
- `invoice_id` (optional, string); Xero invoice id path parameter for docs-derived streams.
- `item_id` (optional, string); Xero item id path parameter for docs-derived streams.
- `journal_id` (optional, string); Xero journal id path parameter for docs-derived streams.
- `journal_number` (optional, string); Xero journal number path parameter for docs-derived streams.
- `linked_transaction_id` (optional, string); Xero linked transaction id path parameter for
  docs-derived streams.
- `manual_journal_id` (optional, string); Xero manual journal id path parameter for docs-derived
  streams.
- `organisation_id` (optional, string); Xero organisation id path parameter for docs-derived
  streams.
- `overpayment_id` (optional, string); Xero overpayment id path parameter for docs-derived streams.
- `payment_id` (optional, string); Xero payment id path parameter for docs-derived streams.
- `prepayment_id` (optional, string); Xero prepayment id path parameter for docs-derived streams.
- `purchase_order_id` (optional, string); Xero purchase order id path parameter for docs-derived
  streams.
- `purchase_order_number` (optional, string); Xero purchase order number path parameter for
  docs-derived streams.
- `quote_id` (optional, string); Xero quote id path parameter for docs-derived streams.
- `receipt_id` (optional, string); Xero receipt id path parameter for docs-derived streams.
- `repeating_invoice_id` (optional, string); Xero repeating invoice id path parameter for
  docs-derived streams.
- `report_id` (optional, string); Xero report id path parameter for docs-derived streams.
- `tax_type` (optional, string); Xero tax type path parameter for docs-derived streams.
- `tenant_id` (required, secret, string); Xero tenant (organisation) id. Sent as the Xero-tenant-id
  header; never logged.
- `tracking_category_id` (optional, string); Xero tracking category id path parameter for
  docs-derived streams.
- `user_id` (optional, string); Xero user id path parameter for docs-derived streams.

Secret fields are redacted in logs and write previews: `access_token`, `tenant_id`.

Default configuration values: `base_url=https://api.xero.com/api.xro/2.0`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `Organisation`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; no page-size parameter; starts at
1; page size 100.

Pagination by stream: none: `account`, `account_attachments`, `bank_transaction`,
`bank_transaction_attachments`, `bank_transactions_history`, `bank_transfer`,
`bank_transfer_attachments`, `bank_transfer_history`, `bank_transfers`, `batch_payment`,
`batch_payment_history`, `batch_payments`, `branding_theme`, `branding_theme_payment_services`,
`branding_themes`, `budget`, `budgets`, `contact`, `contact_attachments`,
`contact_by_contact_number`, `contact_cis_settings`, `contact_group`, `contact_groups`,
`contact_history`, `credit_note`, `credit_note_attachments`, `credit_note_history`, `currencies`,
`expense_claim`, `expense_claim_history`, `expense_claims`, `invoice`, `invoice_attachments`,
`invoice_history`, `invoice_reminders`, `item`, `item_history`, `journal`, `journal_by_number`,
`journals`, `linked_transaction`, `manual_journal`, `manual_journal_attachments`,
`manual_journals_history`, `online_invoice`, `organisation_actions`, `organisation_cis_settings`,
`organisations`, `overpayment`, `overpayment_history`, `payment`, `payment_history`,
`payment_services`, `prepayment`, `prepayment_history`, `purchase_order`,
`purchase_order_attachments`, `purchase_order_by_number`, `purchase_order_history`, `quote`, and 27
more; page_number: `invoices`, `contacts`, `accounts`, `bank_transactions`, `items`, `payments`,
`credit_notes`, `linked_transactions`, `manual_journals`, `overpayments`, `prepayments`,
`purchase_orders`, `quotes`.

- `invoices`: GET `Invoices` - records path `Invoices`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 100; computed output fields `ContactID`,
  `id`, `status`, `type`, `updated_at`.
- `contacts`: GET `Contacts` - records path `Contacts`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 100; computed output fields `id`, `status`,
  `updated_at`.
- `accounts`: GET `Accounts` - records path `Accounts`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 100; computed output fields `id`, `status`,
  `type`, `updated_at`.
- `bank_transactions`: GET `BankTransactions` - records path `BankTransactions`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 100; computed
  output fields `ContactID`, `id`, `status`, `type`, `updated_at`.
- `items`: GET `Items` - records path `Items`; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 100; computed output fields `id`, `updated_at`.
- `payments`: GET `Payments` - records path `Payments`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 100; computed output fields `id`, `status`,
  `updated_at`.
- `account`: GET `Accounts/{{ config.account_id }}` - records path `Accounts`; emits passthrough
  records.
- `account_attachments`: GET `Accounts/{{ config.account_id }}/Attachments` - records path
  `Attachments`; emits passthrough records.
- `bank_transaction`: GET `BankTransactions/{{ config.bank_transaction_id }}` - records path
  `BankTransactions`; emits passthrough records.
- `bank_transaction_attachments`: GET `BankTransactions/{{ config.bank_transaction_id
  }}/Attachments` - records path `Attachments`; emits passthrough records.
- `bank_transactions_history`: GET `BankTransactions/{{ config.bank_transaction_id }}/History` -
  records path `HistoryRecords`; emits passthrough records.
- `bank_transfer`: GET `BankTransfers/{{ config.bank_transfer_id }}` - records path `BankTransfers`;
  emits passthrough records.
- `bank_transfer_attachments`: GET `BankTransfers/{{ config.bank_transfer_id }}/Attachments` -
  records path `Attachments`; emits passthrough records.
- `bank_transfer_history`: GET `BankTransfers/{{ config.bank_transfer_id }}/History` - records path
  `HistoryRecords`; emits passthrough records.
- `bank_transfers`: GET `BankTransfers` - records path `BankTransfers`; emits passthrough records.
- `batch_payment`: GET `BatchPayments/{{ config.batch_payment_id }}` - records path `BatchPayments`;
  emits passthrough records.
- `batch_payment_history`: GET `BatchPayments/{{ config.batch_payment_id }}/History` - records path
  `HistoryRecords`; emits passthrough records.
- `batch_payments`: GET `BatchPayments` - records path `BatchPayments`; emits passthrough records.
- `branding_theme`: GET `BrandingThemes/{{ config.branding_theme_id }}` - records path
  `BrandingThemes`; emits passthrough records.
- `branding_theme_payment_services`: GET `BrandingThemes/{{ config.branding_theme_id
  }}/PaymentServices` - records path `PaymentServices`; emits passthrough records.
- `branding_themes`: GET `BrandingThemes` - records path `BrandingThemes`; emits passthrough
  records.
- `budget`: GET `Budgets/{{ config.budget_id }}` - records path `Budgets`; emits passthrough
  records.
- `budgets`: GET `Budgets` - records path `Budgets`; emits passthrough records.
- `contact`: GET `Contacts/{{ config.contact_id }}` - records path `Contacts`; emits passthrough
  records.
- `contact_attachments`: GET `Contacts/{{ config.contact_id }}/Attachments` - records path
  `Attachments`; emits passthrough records.
- `contact_by_contact_number`: GET `Contacts/{{ config.contact_number }}` - records path `Contacts`;
  emits passthrough records.
- `contact_cis_settings`: GET `Contacts/{{ config.contact_id }}/CISSettings` - records path
  `CISSettings`; emits passthrough records.
- `contact_group`: GET `ContactGroups/{{ config.contact_group_id }}` - records path `ContactGroups`;
  emits passthrough records.
- `contact_groups`: GET `ContactGroups` - records path `ContactGroups`; emits passthrough records.
- `contact_history`: GET `Contacts/{{ config.contact_id }}/History` - records path `HistoryRecords`;
  emits passthrough records.
- `credit_note`: GET `CreditNotes/{{ config.credit_note_id }}` - records path `CreditNotes`; emits
  passthrough records.
- `credit_note_attachments`: GET `CreditNotes/{{ config.credit_note_id }}/Attachments` - records
  path `Attachments`; emits passthrough records.
- `credit_note_history`: GET `CreditNotes/{{ config.credit_note_id }}/History` - records path
  `HistoryRecords`; emits passthrough records.
- `credit_notes`: GET `CreditNotes` - records path `CreditNotes`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 100; emits passthrough records.
- `currencies`: GET `Currencies` - records path `Currencies`; emits passthrough records.
- `expense_claim`: GET `ExpenseClaims/{{ config.expense_claim_id }}` - records path `ExpenseClaims`;
  emits passthrough records.
- `expense_claim_history`: GET `ExpenseClaims/{{ config.expense_claim_id }}/History` - records path
  `HistoryRecords`; emits passthrough records.
- `expense_claims`: GET `ExpenseClaims` - records path `ExpenseClaims`; emits passthrough records.
- `invoice`: GET `Invoices/{{ config.invoice_id }}` - records path `Invoices`; emits passthrough
  records.
- `invoice_attachments`: GET `Invoices/{{ config.invoice_id }}/Attachments` - records path
  `Attachments`; emits passthrough records.
- `invoice_history`: GET `Invoices/{{ config.invoice_id }}/History` - records path `HistoryRecords`;
  emits passthrough records.
- `invoice_reminders`: GET `InvoiceReminders/Settings` - records path `InvoiceReminders`; emits
  passthrough records.
- `item`: GET `Items/{{ config.item_id }}` - records path `Items`; emits passthrough records.
- `item_history`: GET `Items/{{ config.item_id }}/History` - records path `HistoryRecords`; emits
  passthrough records.
- `journal`: GET `Journals/{{ config.journal_id }}` - records path `Journals`; emits passthrough
  records.
- `journal_by_number`: GET `Journals/{{ config.journal_number }}` - records path `Journals`; emits
  passthrough records.
- `journals`: GET `Journals` - records path `Journals`; emits passthrough records.
- `linked_transaction`: GET `LinkedTransactions/{{ config.linked_transaction_id }}` - records path
  `LinkedTransactions`; emits passthrough records.
- `linked_transactions`: GET `LinkedTransactions` - records path `LinkedTransactions`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 100; emits
  passthrough records.
- `manual_journal`: GET `ManualJournals/{{ config.manual_journal_id }}` - records path
  `ManualJournals`; emits passthrough records.
- `manual_journal_attachments`: GET `ManualJournals/{{ config.manual_journal_id }}/Attachments` -
  records path `Attachments`; emits passthrough records.
- `manual_journals`: GET `ManualJournals` - records path `ManualJournals`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 100; emits passthrough
  records.
- `manual_journals_history`: GET `ManualJournals/{{ config.manual_journal_id }}/History` - records
  path `HistoryRecords`; emits passthrough records.
- `online_invoice`: GET `Invoices/{{ config.invoice_id }}/OnlineInvoice` - records path
  `OnlineInvoices`; emits passthrough records.
- `organisation_actions`: GET `Organisation/Actions` - records path `Actions`; emits passthrough
  records.
- `organisation_cis_settings`: GET `Organisation/{{ config.organisation_id }}/CISSettings` - records
  path `CISSettings`; emits passthrough records.
- `organisations`: GET `Organisation` - records path `Organisations`; emits passthrough records.
- `overpayment`: GET `Overpayments/{{ config.overpayment_id }}` - records path `Overpayments`; emits
  passthrough records.
- `overpayment_history`: GET `Overpayments/{{ config.overpayment_id }}/History` - records path
  `HistoryRecords`; emits passthrough records.
- `overpayments`: GET `Overpayments` - records path `Overpayments`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 100; emits passthrough records.
- `payment`: GET `Payments/{{ config.payment_id }}` - records path `Payments`; emits passthrough
  records.
- `payment_history`: GET `Payments/{{ config.payment_id }}/History` - records path `HistoryRecords`;
  emits passthrough records.
- `payment_services`: GET `PaymentServices` - records path `PaymentServices`; emits passthrough
  records.
- `prepayment`: GET `Prepayments/{{ config.prepayment_id }}` - records path `Prepayments`; emits
  passthrough records.
- `prepayment_history`: GET `Prepayments/{{ config.prepayment_id }}/History` - records path
  `HistoryRecords`; emits passthrough records.
- `prepayments`: GET `Prepayments` - records path `Prepayments`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 100; emits passthrough records.
- `purchase_order`: GET `PurchaseOrders/{{ config.purchase_order_id }}` - records path
  `PurchaseOrders`; emits passthrough records.
- `purchase_order_attachments`: GET `PurchaseOrders/{{ config.purchase_order_id }}/Attachments` -
  records path `Attachments`; emits passthrough records.
- `purchase_order_by_number`: GET `PurchaseOrders/{{ config.purchase_order_number }}` - records path
  `PurchaseOrders`; emits passthrough records.
- `purchase_order_history`: GET `PurchaseOrders/{{ config.purchase_order_id }}/History` - records
  path `HistoryRecords`; emits passthrough records.
- `purchase_orders`: GET `PurchaseOrders` - records path `PurchaseOrders`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 100; emits passthrough
  records.
- `quote`: GET `Quotes/{{ config.quote_id }}` - records path `Quotes`; emits passthrough records.
- `quote_attachments`: GET `Quotes/{{ config.quote_id }}/Attachments` - records path `Attachments`;
  emits passthrough records.
- `quote_history`: GET `Quotes/{{ config.quote_id }}/History` - records path `HistoryRecords`; emits
  passthrough records.
- `quotes`: GET `Quotes` - records path `Quotes`; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 100; emits passthrough records.
- `receipt`: GET `Receipts/{{ config.receipt_id }}` - records path `Receipts`; emits passthrough
  records.
- `receipt_attachments`: GET `Receipts/{{ config.receipt_id }}/Attachments` - records path
  `Attachments`; emits passthrough records.
- `receipt_history`: GET `Receipts/{{ config.receipt_id }}/History` - records path `HistoryRecords`;
  emits passthrough records.
- `receipts`: GET `Receipts` - records path `Receipts`; emits passthrough records.
- `repeating_invoice`: GET `RepeatingInvoices/{{ config.repeating_invoice_id }}` - records path
  `RepeatingInvoices`; emits passthrough records.
- `repeating_invoice_attachments`: GET `RepeatingInvoices/{{ config.repeating_invoice_id
  }}/Attachments` - records path `Attachments`; emits passthrough records.
- `repeating_invoice_history`: GET `RepeatingInvoices/{{ config.repeating_invoice_id }}/History` -
  records path `HistoryRecords`; emits passthrough records.
- `repeating_invoices`: GET `RepeatingInvoices` - records path `RepeatingInvoices`; emits
  passthrough records.
- `report_aged_payables_by_contact`: GET `Reports/AgedPayablesByContact` - records path `Reports`;
  emits passthrough records.
- `report_aged_receivables_by_contact`: GET `Reports/AgedReceivablesByContact` - records path
  `Reports`; emits passthrough records.
- `report_balance_sheet`: GET `Reports/BalanceSheet` - records path `Reports`; emits passthrough
  records.
- `report_bank_summary`: GET `Reports/BankSummary` - records path `Reports`; emits passthrough
  records.
- `report_budget_summary`: GET `Reports/BudgetSummary` - records path `Reports`; emits passthrough
  records.
- `report_executive_summary`: GET `Reports/ExecutiveSummary` - records path `Reports`; emits
  passthrough records.
- `report_from_id`: GET `Reports/{{ config.report_id }}` - records path `Reports`; emits passthrough
  records.
- `report_profit_and_loss`: GET `Reports/ProfitAndLoss` - records path `Reports`; emits passthrough
  records.
- `report_ten_ninety_nine`: GET `Reports/TenNinetyNine` - records path `Reports`; emits passthrough
  records.
- `report_trial_balance`: GET `Reports/TrialBalance` - records path `Reports`; emits passthrough
  records.
- `reports_list`: GET `Reports` - records path `Reports`; emits passthrough records.
- `tax_rate_by_tax_type`: GET `TaxRates/{{ config.tax_type }}` - records path `TaxRates`; emits
  passthrough records.
- `tax_rates`: GET `TaxRates` - records path `TaxRates`; emits passthrough records.
- `tracking_categories`: GET `TrackingCategories` - records path `TrackingCategories`; emits
  passthrough records.
- `tracking_category`: GET `TrackingCategories/{{ config.tracking_category_id }}` - records path
  `TrackingCategories`; emits passthrough records.
- `user`: GET `Users/{{ config.user_id }}` - records path `Users`; emits passthrough records.
- `users`: GET `Users` - records path `Users`; emits passthrough records.

## Write actions & risks

Overall write risk: creates, updates, emails, sets up, and deletes Xero Accounting API resources in
the connected tenant.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_account`: PUT `Accounts` - kind `create`; body type `json`; accepted fields `AccountID`,
  `AddToWatchlist`, `BankAccountNumber`, `BankAccountType`, `Class`, `Code`, `CurrencyCode`,
  `Description`, `EnablePaymentsToAccount`, `HasAttachments`, `Name`, `ReportingCode`,
  `ReportingCodeName`, `ShowInExpenseClaims`, `Status`, `SystemAccount`, `TaxType`, `Type`, and 2
  more; risk: creates Xero account resources in the connected tenant; approval required.
- `delete_account`: DELETE `Accounts/{{ record.account_id }}` - kind `delete`; body type `none`;
  path fields `account_id`; required record fields `account_id`; accepted fields `account_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: deletes
  Xero account resources in the connected tenant; approval required.
- `update_account`: POST `Accounts/{{ record.account_id }}` - kind `update`; body type `json`; path
  fields `account_id`; required record fields `account_id`; accepted fields `Accounts`,
  `account_id`; risk: mutates Xero account resources in the connected tenant; approval required.
- `upsert_bank_transactions`: POST `BankTransactions` - kind `update`; body type `json`; accepted
  fields `BankTransactions`, `Warnings`, `pagination`; risk: mutates Xero bank transactions
  resources in the connected tenant; approval required.
- `create_bank_transactions`: PUT `BankTransactions` - kind `create`; body type `json`; accepted
  fields `BankTransactions`, `Warnings`, `pagination`; risk: creates Xero bank transactions
  resources in the connected tenant; approval required.
- `update_bank_transaction`: POST `BankTransactions/{{ record.bank_transaction_id }}` - kind
  `update`; body type `json`; path fields `bank_transaction_id`; required record fields
  `bank_transaction_id`; accepted fields `BankTransactions`, `Warnings`, `bank_transaction_id`,
  `pagination`; risk: mutates Xero bank transaction resources in the connected tenant; approval
  required.
- `create_bank_transaction_history_record`: PUT `BankTransactions/{{ record.bank_transaction_id
  }}/History` - kind `create`; body type `none`; path fields `bank_transaction_id`; required record
  fields `bank_transaction_id`; accepted fields `bank_transaction_id`; risk: creates Xero bank
  transaction history record resources in the connected tenant; approval required.
- `create_bank_transfer`: PUT `BankTransfers` - kind `create`; body type `json`; accepted fields
  `BankTransfers`; risk: creates Xero bank transfer resources in the connected tenant; approval
  required.
- `create_bank_transfer_history_record`: PUT `BankTransfers/{{ record.bank_transfer_id }}/History` -
  kind `create`; body type `none`; path fields `bank_transfer_id`; required record fields
  `bank_transfer_id`; accepted fields `bank_transfer_id`; risk: creates Xero bank transfer history
  record resources in the connected tenant; approval required.
- `delete_batch_payment`: POST `BatchPayments` - kind `delete`; body type `json`; required record
  fields `Status`, `BatchPaymentID`; accepted fields `BatchPaymentID`, `Status`; confirmation
  `destructive`; risk: deletes Xero batch payment resources in the connected tenant; approval
  required.
- `create_batch_payment`: PUT `BatchPayments` - kind `create`; body type `json`; accepted fields
  `BatchPayments`; risk: creates Xero batch payment resources in the connected tenant; approval
  required.
- `delete_batch_payment_by_url_param`: POST `BatchPayments/{{ record.batch_payment_id }}` - kind
  `delete`; body type `json`; path fields `batch_payment_id`; required record fields
  `batch_payment_id`, `Status`; accepted fields `Status`, `batch_payment_id`; confirmation
  `destructive`; risk: deletes Xero batch payment by url param resources in the connected tenant;
  approval required.
- `create_batch_payment_history_record`: PUT `BatchPayments/{{ record.batch_payment_id }}/History` -
  kind `create`; body type `none`; path fields `batch_payment_id`; required record fields
  `batch_payment_id`; accepted fields `batch_payment_id`; risk: creates Xero batch payment history
  record resources in the connected tenant; approval required.
- `create_branding_theme_payment_services`: POST `BrandingThemes/{{ record.branding_theme_id
  }}/PaymentServices` - kind `create`; body type `json`; path fields `branding_theme_id`; required
  record fields `branding_theme_id`; accepted fields `PaymentServices`, `branding_theme_id`; risk:
  creates Xero branding theme payment services resources in the connected tenant; approval required.
- `create_contact_group`: PUT `ContactGroups` - kind `create`; body type `json`; accepted fields
  `ContactGroups`; risk: creates Xero contact group resources in the connected tenant; approval
  required.
- `update_contact_group`: POST `ContactGroups/{{ record.contact_group_id }}` - kind `update`; body
  type `json`; path fields `contact_group_id`; required record fields `contact_group_id`; accepted
  fields `ContactGroups`, `contact_group_id`; risk: mutates Xero contact group resources in the
  connected tenant; approval required.
- `delete_contact_group_contacts`: DELETE `ContactGroups/{{ record.contact_group_id }}/Contacts` -
  kind `delete`; body type `none`; path fields `contact_group_id`; required record fields
  `contact_group_id`; accepted fields `contact_group_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes Xero contact group contacts resources in
  the connected tenant; approval required.
- `create_contact_group_contacts`: PUT `ContactGroups/{{ record.contact_group_id }}/Contacts` - kind
  `create`; body type `json`; path fields `contact_group_id`; required record fields
  `contact_group_id`; accepted fields `Contacts`, `Warnings`, `contact_group_id`, `pagination`;
  risk: creates Xero contact group contacts resources in the connected tenant; approval required.
- `delete_contact_group_contact`: DELETE `ContactGroups/{{ record.contact_group_id }}/Contacts/{{
  record.contact_id }}` - kind `delete`; body type `none`; path fields `contact_group_id`,
  `contact_id`; required record fields `contact_group_id`, `contact_id`; accepted fields
  `contact_group_id`, `contact_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes Xero contact group contact resources in the connected
  tenant; approval required.
- `upsert_contacts`: POST `Contacts` - kind `update`; body type `json`; accepted fields `Contacts`,
  `Warnings`, `pagination`; risk: mutates Xero contacts resources in the connected tenant; approval
  required.
- `create_contacts`: PUT `Contacts` - kind `create`; body type `json`; accepted fields `Contacts`,
  `Warnings`, `pagination`; risk: creates Xero contacts resources in the connected tenant; approval
  required.
- `update_contact`: POST `Contacts/{{ record.contact_id }}` - kind `update`; body type `json`; path
  fields `contact_id`; required record fields `contact_id`; accepted fields `Contacts`, `Warnings`,
  `contact_id`, `pagination`; risk: mutates Xero contact resources in the connected tenant; approval
  required.
- `create_contact_history`: PUT `Contacts/{{ record.contact_id }}/History` - kind `create`; body
  type `none`; path fields `contact_id`; required record fields `contact_id`; accepted fields
  `contact_id`; risk: creates Xero contact history resources in the connected tenant; approval
  required.
- `upsert_credit_notes`: POST `CreditNotes` - kind `update`; body type `json`; accepted fields
  `CreditNotes`, `Warnings`, `pagination`; risk: mutates Xero credit notes resources in the
  connected tenant; approval required.
- `create_credit_notes`: PUT `CreditNotes` - kind `create`; body type `json`; accepted fields
  `CreditNotes`, `Warnings`, `pagination`; risk: creates Xero credit notes resources in the
  connected tenant; approval required.
- `update_credit_note`: POST `CreditNotes/{{ record.credit_note_id }}` - kind `update`; body type
  `json`; path fields `credit_note_id`; required record fields `credit_note_id`; accepted fields
  `CreditNotes`, `Warnings`, `credit_note_id`, `pagination`; risk: mutates Xero credit note
  resources in the connected tenant; approval required.
- `create_credit_note_allocation`: PUT `CreditNotes/{{ record.credit_note_id }}/Allocations` - kind
  `create`; body type `json`; path fields `credit_note_id`; required record fields `credit_note_id`;
  accepted fields `Allocations`, `credit_note_id`; risk: creates Xero credit note allocation
  resources in the connected tenant; approval required.
- `delete_credit_note_allocations`: DELETE `CreditNotes/{{ record.credit_note_id }}/Allocations/{{
  record.allocation_id }}` - kind `delete`; body type `none`; path fields `credit_note_id`,
  `allocation_id`; required record fields `credit_note_id`, `allocation_id`; accepted fields
  `allocation_id`, `credit_note_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes Xero credit note allocations resources in the connected
  tenant; approval required.
- `create_credit_note_history`: PUT `CreditNotes/{{ record.credit_note_id }}/History` - kind
  `create`; body type `none`; path fields `credit_note_id`; required record fields `credit_note_id`;
  accepted fields `credit_note_id`; risk: creates Xero credit note history resources in the
  connected tenant; approval required.
- `create_currency`: PUT `Currencies` - kind `create`; body type `json`; accepted fields `Code`,
  `Description`; risk: creates Xero currency resources in the connected tenant; approval required.
- `create_expense_claims`: PUT `ExpenseClaims` - kind `create`; body type `json`; accepted fields
  `ExpenseClaims`; risk: creates Xero expense claims resources in the connected tenant; approval
  required.
- `update_expense_claim`: POST `ExpenseClaims/{{ record.expense_claim_id }}` - kind `update`; body
  type `json`; path fields `expense_claim_id`; required record fields `expense_claim_id`; accepted
  fields `ExpenseClaims`, `expense_claim_id`; risk: mutates Xero expense claim resources in the
  connected tenant; approval required.
- `create_expense_claim_history`: PUT `ExpenseClaims/{{ record.expense_claim_id }}/History` - kind
  `create`; body type `none`; path fields `expense_claim_id`; required record fields
  `expense_claim_id`; accepted fields `expense_claim_id`; risk: creates Xero expense claim history
  resources in the connected tenant; approval required.
- `upsert_invoices`: POST `Invoices` - kind `update`; body type `json`; accepted fields `Invoices`,
  `Warnings`, `pagination`; risk: mutates Xero invoices resources in the connected tenant; approval
  required.
- `create_invoices`: PUT `Invoices` - kind `create`; body type `json`; accepted fields `Invoices`,
  `Warnings`, `pagination`; risk: creates Xero invoices resources in the connected tenant; approval
  required.
- `update_invoice`: POST `Invoices/{{ record.invoice_id }}` - kind `update`; body type `json`; path
  fields `invoice_id`; required record fields `invoice_id`; accepted fields `Invoices`, `Warnings`,
  `invoice_id`, `pagination`; risk: mutates Xero invoice resources in the connected tenant; approval
  required.
- `email_invoice`: POST `Invoices/{{ record.invoice_id }}/Email` - kind `custom`; body type `json`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `Status`,
  `invoice_id`; risk: executes Xero email invoice resources in the connected tenant; approval
  required.
- `create_invoice_history`: PUT `Invoices/{{ record.invoice_id }}/History` - kind `create`; body
  type `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: creates Xero invoice history resources in the connected tenant; approval
  required.
- `upsert_items`: POST `Items` - kind `update`; body type `json`; accepted fields `Items`; risk:
  mutates Xero items resources in the connected tenant; approval required.
- `create_items`: PUT `Items` - kind `create`; body type `json`; accepted fields `Items`; risk:
  creates Xero items resources in the connected tenant; approval required.
- `delete_item`: DELETE `Items/{{ record.item_id }}` - kind `delete`; body type `none`; path fields
  `item_id`; required record fields `item_id`; accepted fields `item_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes Xero item resources in the
  connected tenant; approval required.
- `update_item`: POST `Items/{{ record.item_id }}` - kind `update`; body type `json`; path fields
  `item_id`; required record fields `item_id`; accepted fields `Items`, `item_id`; risk: mutates
  Xero item resources in the connected tenant; approval required.
- `create_item_history`: PUT `Items/{{ record.item_id }}/History` - kind `create`; body type `none`;
  path fields `item_id`; required record fields `item_id`; accepted fields `item_id`; risk: creates
  Xero item history resources in the connected tenant; approval required.
- `create_linked_transaction`: PUT `LinkedTransactions` - kind `create`; body type `json`; accepted
  fields `ContactID`, `LinkedTransactionID`, `SourceLineItemID`, `SourceTransactionID`,
  `SourceTransactionTypeCode`, `Status`, `TargetLineItemID`, `TargetTransactionID`, `Type`,
  `UpdatedDateUTC`, `ValidationErrors`; risk: creates Xero linked transaction resources in the
  connected tenant; approval required.
- `delete_linked_transaction`: DELETE `LinkedTransactions/{{ record.linked_transaction_id }}` - kind
  `delete`; body type `none`; path fields `linked_transaction_id`; required record fields
  `linked_transaction_id`; accepted fields `linked_transaction_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes Xero linked transaction
  resources in the connected tenant; approval required.
- `update_linked_transaction`: POST `LinkedTransactions/{{ record.linked_transaction_id }}` - kind
  `update`; body type `json`; path fields `linked_transaction_id`; required record fields
  `linked_transaction_id`; accepted fields `LinkedTransactions`, `linked_transaction_id`; risk:
  mutates Xero linked transaction resources in the connected tenant; approval required.
- `upsert_manual_journals`: POST `ManualJournals` - kind `update`; body type `json`; accepted fields
  `ManualJournals`, `Warnings`, `pagination`; risk: mutates Xero manual journals resources in the
  connected tenant; approval required.
- `create_manual_journals`: PUT `ManualJournals` - kind `create`; body type `json`; accepted fields
  `ManualJournals`, `Warnings`, `pagination`; risk: creates Xero manual journals resources in the
  connected tenant; approval required.
- `update_manual_journal`: POST `ManualJournals/{{ record.manual_journal_id }}` - kind `update`;
  body type `json`; path fields `manual_journal_id`; required record fields `manual_journal_id`;
  accepted fields `ManualJournals`, `Warnings`, `manual_journal_id`, `pagination`; risk: mutates
  Xero manual journal resources in the connected tenant; approval required.
- `create_manual_journal_history_record`: PUT `ManualJournals/{{ record.manual_journal_id
  }}/History` - kind `create`; body type `none`; path fields `manual_journal_id`; required record
  fields `manual_journal_id`; accepted fields `manual_journal_id`; risk: creates Xero manual journal
  history record resources in the connected tenant; approval required.
- `create_overpayment_allocations`: PUT `Overpayments/{{ record.overpayment_id }}/Allocations` -
  kind `create`; body type `json`; path fields `overpayment_id`; required record fields
  `overpayment_id`; accepted fields `Allocations`, `overpayment_id`; risk: creates Xero overpayment
  allocations resources in the connected tenant; approval required.
- `delete_overpayment_allocations`: DELETE `Overpayments/{{ record.overpayment_id }}/Allocations/{{
  record.allocation_id }}` - kind `delete`; body type `none`; path fields `overpayment_id`,
  `allocation_id`; required record fields `overpayment_id`, `allocation_id`; accepted fields
  `allocation_id`, `overpayment_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes Xero overpayment allocations resources in the connected
  tenant; approval required.
- `create_overpayment_history`: PUT `Overpayments/{{ record.overpayment_id }}/History` - kind
  `create`; body type `none`; path fields `overpayment_id`; required record fields `overpayment_id`;
  accepted fields `overpayment_id`; risk: creates Xero overpayment history resources in the
  connected tenant; approval required.
- `create_payment`: POST `Payments` - kind `create`; body type `json`; accepted fields `Account`,
  `Amount`, `BankAccountNumber`, `BankAmount`, `BatchPayment`, `BatchPaymentID`, `Code`,
  `CreditNote`, `CreditNoteNumber`, `CurrencyRate`, `Date`, `Details`, `HasAccount`,
  `HasValidationErrors`, `Invoice`, `InvoiceNumber`, `IsReconciled`, `Overpayment`, and 10 more;
  risk: creates Xero payment resources in the connected tenant; approval required.
- `create_payments`: PUT `Payments` - kind `create`; body type `json`; accepted fields `Payments`,
  `Warnings`, `pagination`; risk: creates Xero payments resources in the connected tenant; approval
  required.
- `delete_payment`: POST `Payments/{{ record.payment_id }}` - kind `delete`; body type `json`; path
  fields `payment_id`; required record fields `payment_id`, `Status`; accepted fields `Status`,
  `payment_id`; confirmation `destructive`; risk: deletes Xero payment resources in the connected
  tenant; approval required.
- `create_payment_history`: PUT `Payments/{{ record.payment_id }}/History` - kind `create`; body
  type `none`; path fields `payment_id`; required record fields `payment_id`; accepted fields
  `payment_id`; risk: creates Xero payment history resources in the connected tenant; approval
  required.
- `create_payment_service`: PUT `PaymentServices` - kind `create`; body type `json`; accepted fields
  `PaymentServices`; risk: creates Xero payment service resources in the connected tenant; approval
  required.
- `create_prepayment_allocations`: PUT `Prepayments/{{ record.prepayment_id }}/Allocations` - kind
  `create`; body type `json`; path fields `prepayment_id`; required record fields `prepayment_id`;
  accepted fields `Allocations`, `prepayment_id`; risk: creates Xero prepayment allocations
  resources in the connected tenant; approval required.
- `delete_prepayment_allocations`: DELETE `Prepayments/{{ record.prepayment_id }}/Allocations/{{
  record.allocation_id }}` - kind `delete`; body type `none`; path fields `prepayment_id`,
  `allocation_id`; required record fields `prepayment_id`, `allocation_id`; accepted fields
  `allocation_id`, `prepayment_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes Xero prepayment allocations resources in the connected
  tenant; approval required.
- `create_prepayment_history`: PUT `Prepayments/{{ record.prepayment_id }}/History` - kind `create`;
  body type `none`; path fields `prepayment_id`; required record fields `prepayment_id`; accepted
  fields `prepayment_id`; risk: creates Xero prepayment history resources in the connected tenant;
  approval required.
- `upsert_purchase_orders`: POST `PurchaseOrders` - kind `update`; body type `json`; accepted fields
  `PurchaseOrders`, `Warnings`, `pagination`; risk: mutates Xero purchase orders resources in the
  connected tenant; approval required.
- `create_purchase_orders`: PUT `PurchaseOrders` - kind `create`; body type `json`; accepted fields
  `PurchaseOrders`, `Warnings`, `pagination`; risk: creates Xero purchase orders resources in the
  connected tenant; approval required.
- `update_purchase_order`: POST `PurchaseOrders/{{ record.purchase_order_id }}` - kind `update`;
  body type `json`; path fields `purchase_order_id`; required record fields `purchase_order_id`;
  accepted fields `PurchaseOrders`, `Warnings`, `pagination`, `purchase_order_id`; risk: mutates
  Xero purchase order resources in the connected tenant; approval required.
- `create_purchase_order_history`: PUT `PurchaseOrders/{{ record.purchase_order_id }}/History` -
  kind `create`; body type `none`; path fields `purchase_order_id`; required record fields
  `purchase_order_id`; accepted fields `purchase_order_id`; risk: creates Xero purchase order
  history resources in the connected tenant; approval required.
- `upsert_quotes`: POST `Quotes` - kind `update`; body type `json`; accepted fields `Quotes`; risk:
  mutates Xero quotes resources in the connected tenant; approval required.
- `create_quotes`: PUT `Quotes` - kind `create`; body type `json`; accepted fields `Quotes`; risk:
  creates Xero quotes resources in the connected tenant; approval required.
- `update_quote`: POST `Quotes/{{ record.quote_id }}` - kind `update`; body type `json`; path fields
  `quote_id`; required record fields `quote_id`; accepted fields `Quotes`, `quote_id`; risk: mutates
  Xero quote resources in the connected tenant; approval required.
- `create_quote_history`: PUT `Quotes/{{ record.quote_id }}/History` - kind `create`; body type
  `none`; path fields `quote_id`; required record fields `quote_id`; accepted fields `quote_id`;
  risk: creates Xero quote history resources in the connected tenant; approval required.
- `create_receipt`: PUT `Receipts` - kind `create`; body type `json`; accepted fields `Receipts`;
  risk: creates Xero receipt resources in the connected tenant; approval required.
- `update_receipt`: POST `Receipts/{{ record.receipt_id }}` - kind `update`; body type `json`; path
  fields `receipt_id`; required record fields `receipt_id`; accepted fields `Receipts`,
  `receipt_id`; risk: mutates Xero receipt resources in the connected tenant; approval required.
- `create_receipt_history`: PUT `Receipts/{{ record.receipt_id }}/History` - kind `create`; body
  type `none`; path fields `receipt_id`; required record fields `receipt_id`; accepted fields
  `receipt_id`; risk: creates Xero receipt history resources in the connected tenant; approval
  required.
- `upsert_repeating_invoices`: POST `RepeatingInvoices` - kind `update`; body type `json`; accepted
  fields `RepeatingInvoices`; risk: mutates Xero repeating invoices resources in the connected
  tenant; approval required.
- `create_repeating_invoices`: PUT `RepeatingInvoices` - kind `create`; body type `json`; accepted
  fields `RepeatingInvoices`; risk: creates Xero repeating invoices resources in the connected
  tenant; approval required.
- `update_repeating_invoice`: POST `RepeatingInvoices/{{ record.repeating_invoice_id }}` - kind
  `update`; body type `json`; path fields `repeating_invoice_id`; required record fields
  `repeating_invoice_id`; accepted fields `RepeatingInvoices`, `repeating_invoice_id`; risk: mutates
  Xero repeating invoice resources in the connected tenant; approval required.
- `create_repeating_invoice_history`: PUT `RepeatingInvoices/{{ record.repeating_invoice_id
  }}/History` - kind `create`; body type `none`; path fields `repeating_invoice_id`; required record
  fields `repeating_invoice_id`; accepted fields `repeating_invoice_id`; risk: creates Xero
  repeating invoice history resources in the connected tenant; approval required.
- `setup_organisation`: POST `Setup` - kind `update`; body type `json`; accepted fields `Accounts`,
  `ConversionBalances`, `ConversionDate`; risk: mutates Xero setup resources in the connected
  tenant; approval required.
- `update_tax_rate`: POST `TaxRates` - kind `update`; body type `json`; accepted fields `TaxRates`;
  risk: mutates Xero tax rate resources in the connected tenant; approval required.
- `create_tax_rates`: PUT `TaxRates` - kind `create`; body type `json`; accepted fields `TaxRates`;
  risk: creates Xero tax rates resources in the connected tenant; approval required.
- `create_tracking_category`: PUT `TrackingCategories` - kind `create`; body type `json`; accepted
  fields `Name`, `Option`, `Options`, `Status`, `TrackingCategoryID`, `TrackingOptionID`; risk:
  creates Xero tracking category resources in the connected tenant; approval required.
- `delete_tracking_category`: DELETE `TrackingCategories/{{ record.tracking_category_id }}` - kind
  `delete`; body type `none`; path fields `tracking_category_id`; required record fields
  `tracking_category_id`; accepted fields `tracking_category_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes Xero tracking category resources in
  the connected tenant; approval required.
- `update_tracking_category`: POST `TrackingCategories/{{ record.tracking_category_id }}` - kind
  `update`; body type `json`; path fields `tracking_category_id`; required record fields
  `tracking_category_id`; accepted fields `Name`, `Option`, `Options`, `Status`,
  `TrackingCategoryID`, `TrackingOptionID`, `tracking_category_id`; risk: mutates Xero tracking
  category resources in the connected tenant; approval required.
- `create_tracking_options`: PUT `TrackingCategories/{{ record.tracking_category_id }}/Options` -
  kind `create`; body type `json`; path fields `tracking_category_id`; required record fields
  `tracking_category_id`; accepted fields `Name`, `Status`, `TrackingCategoryID`,
  `TrackingOptionID`, `tracking_category_id`; risk: creates Xero tracking options resources in the
  connected tenant; approval required.
- `delete_tracking_options`: DELETE `TrackingCategories/{{ record.tracking_category_id }}/Options/{{
  record.tracking_option_id }}` - kind `delete`; body type `none`; path fields
  `tracking_category_id`, `tracking_option_id`; required record fields `tracking_category_id`,
  `tracking_option_id`; accepted fields `tracking_category_id`, `tracking_option_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: deletes Xero
  tracking options resources in the connected tenant; approval required.
- `update_tracking_options`: POST `TrackingCategories/{{ record.tracking_category_id }}/Options/{{
  record.tracking_option_id }}` - kind `update`; body type `json`; path fields
  `tracking_category_id`, `tracking_option_id`; required record fields `tracking_category_id`,
  `tracking_option_id`; accepted fields `Name`, `Status`, `TrackingCategoryID`, `TrackingOptionID`,
  `tracking_category_id`, `tracking_option_id`; risk: mutates Xero tracking options resources in the
  connected tenant; approval required.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 100 stream-backed endpoint group(s), 85 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=48.
