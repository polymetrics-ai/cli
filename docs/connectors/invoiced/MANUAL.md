# pm connectors inspect invoiced

```text
NAME
  pm connectors inspect invoiced - Invoiced connector manual

SYNOPSIS
  pm connectors inspect invoiced
  pm connectors inspect invoiced --json
  pm credentials add <name> --connector invoiced [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes the documented Invoiced REST API surface for billing, payments, subscriptions, events, and related resources.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  contact_id
  coupon_id
  credit_balance_adjustment_id
  credit_note_id
  customer_id
  estimate_id
  event_id
  file_id
  invoice_id
  item_id
  line_item_id
  max_pages
  mode
  page_size
  payment_id
  plan_id
  subscription_id
  task_id
  tax_rate_id
  api_key (secret)

ETL STREAMS
  customers:
    primary key: id
    cursor: updated_at
    fields: balance(), country(), created_at(), currency(), email(), id(), name(), number(), object(), phone(), type(), updated_at()
  invoices:
    primary key: id
    cursor: updated_at
    fields: balance(), closed(), created_at(), currency(), customer(), due_date(), id(), number(), object(), paid(), status(), total(), updated_at()
  payments:
    primary key: id
    cursor: updated_at
    fields: amount(), created_at(), currency(), customer(), date(), id(), invoice(), method(), object(), status(), updated_at()
  subscriptions:
    primary key: id
    cursor: updated_at
    fields: canceled_at(), created_at(), customer(), id(), object(), period_end(), period_start(), plan(), quantity(), start_date(), status(), updated_at()
  estimates:
    primary key: id
    cursor: updated_at
    fields: approved(), closed(), created_at(), currency(), customer(), expiration_date(), id(), number(), object(), status(), total(), updated_at()
  customer_contacts:
    primary key: id
    fields: id()
  customer_contact:
    primary key: id
    fields: id()
  coupons:
    primary key: id
    fields: id()
  coupon:
    primary key: id
    fields: id()
  credit_balance_adjustments:
    primary key: id
    fields: id()
  credit_balance_adjustment:
    primary key: id
    fields: id()
  credit_notes:
    primary key: id
    fields: id()
  credit_note:
    primary key: id
    fields: id()
  credit_note_attachments:
    primary key: id
    fields: id()
  customer:
    primary key: id
    fields: id()
  customer_balance:
  estimate:
    primary key: id
    fields: id()
  estimate_attachments:
    primary key: id
    fields: id()
  events:
    primary key: id
    fields: id()
  event:
    primary key: id
    fields: id()
  file:
    primary key: id
    fields: id()
  invoice:
    primary key: id
    fields: id()
  invoice_attachments:
    primary key: id
    fields: id()
  items:
    primary key: id
    fields: id()
  item:
    primary key: id
    fields: id()
  customer_line_items:
    primary key: id
    fields: id()
  customer_line_item:
    primary key: id
    fields: id()
  notes:
    primary key: id
    fields: id()
  customer_notes:
    primary key: id
    fields: id()
  invoice_notes:
    primary key: id
    fields: id()
  invoice_payment_plan:
    primary key: id
    fields: id()
  customer_payment_sources:
    primary key: id
    fields: id()
  payment:
    primary key: id
    fields: id()
  plans:
    primary key: id
    fields: id()
  plan:
    primary key: id
    fields: id()
  subscription:
    primary key: id
    fields: id()
  tasks:
    primary key: id
    fields: id()
  task:
    primary key: id
    fields: id()
  tax_rates:
    primary key: id
    fields: id()
  tax_rate:
    primary key: id
    fields: id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_charge:
    endpoint: POST /charges
    risk: High-impact Invoiced mutation: Create a charge.
  create_contact:
    endpoint: POST /customers/{{ record.customer_id }}/contacts
    required fields: customer_id
    risk: Invoiced mutation: Create a contact.
  update_contact:
    endpoint: PATCH /customers/{{ record.customer_id }}/contacts/{{ record.contact_id }}
    required fields: customer_id, contact_id
    risk: Invoiced mutation: Update a contact.
  delete_contact:
    endpoint: DELETE /customers/{{ record.customer_id }}/contacts/{{ record.contact_id }}
    required fields: customer_id, contact_id
    risk: Deletes or cancels Invoiced data: Delete a contact.
  create_coupon:
    endpoint: POST /coupons
    risk: Invoiced mutation: Create a coupon.
  update_coupon:
    endpoint: PATCH /coupons/{{ record.coupon_id }}
    required fields: coupon_id
    risk: Invoiced mutation: Update a coupon.
  delete_coupon:
    endpoint: DELETE /coupons/{{ record.coupon_id }}
    required fields: coupon_id
    risk: Deletes or cancels Invoiced data: Delete a coupon.
  create_credit_balance_adjustment:
    endpoint: POST /credit_balance_adjustments
    risk: Invoiced mutation: Create a credit balance adjustment.
  update_credit_balance_adjustment:
    endpoint: PATCH /credit_balance_adjustments/{{ record.credit_balance_adjustment_id }}
    required fields: credit_balance_adjustment_id
    risk: Invoiced mutation: Update a credit balance adjustment.
  delete_credit_balance_adjustment:
    endpoint: DELETE /credit_balance_adjustments/{{ record.credit_balance_adjustment_id }}
    required fields: credit_balance_adjustment_id
    risk: Deletes or cancels Invoiced data: Delete a credit balance adjustment.
  create_credit_note:
    endpoint: POST /credit_notes
    risk: Invoiced mutation: Create a credit note.
  update_credit_note:
    endpoint: PATCH /credit_notes/{{ record.credit_note_id }}
    required fields: credit_note_id
    risk: Invoiced mutation: Update a credit note.
  send_credit_note_email:
    endpoint: POST /credit_notes/{{ record.credit_note_id }}/emails
    required fields: credit_note_id
    risk: Invoiced mutation: Send a credit note email.
  void_credit_note:
    endpoint: POST /credit_notes/{{ record.credit_note_id }}/void
    required fields: credit_note_id
    risk: Destructive Invoiced mutation: Void a credit note.
  delete_credit_note:
    endpoint: DELETE /credit_notes/{{ record.credit_note_id }}
    required fields: credit_note_id
    risk: Deletes or cancels Invoiced data: Delete a credit note.
  create_customer:
    endpoint: POST /customers
    risk: Invoiced mutation: Create a customer.
  update_customer:
    endpoint: PATCH /customers/{{ record.customer_id }}
    required fields: customer_id
    risk: Invoiced mutation: Update a customer.
  send_statement_email:
    endpoint: POST /customers/{{ record.customer_id }}/emails
    required fields: customer_id
    risk: Invoiced mutation: Send a statement email.
  send_statement_sms:
    endpoint: POST /customers/{{ record.customer_id }}/text_messages
    required fields: customer_id
    risk: Invoiced mutation: Send a statement SMS.
  send_statement_letter:
    endpoint: POST /customers/{{ record.customer_id }}/letters
    required fields: customer_id
    risk: Invoiced mutation: Send a statement letter.
  delete_customer:
    endpoint: DELETE /customers/{{ record.customer_id }}
    required fields: customer_id
    risk: Deletes or cancels Invoiced data: Delete a customer.
  create_an_estimate:
    endpoint: POST /estimates
    risk: Invoiced mutation: Create an estimate.
  update_an_estimate:
    endpoint: PATCH /estimates/{{ record.estimate_id }}
    required fields: estimate_id
    risk: Invoiced mutation: Update an estimate.
  send_estimate_email:
    endpoint: POST /estimates/{{ record.estimate_id }}/emails
    required fields: estimate_id
    risk: Invoiced mutation: Send an estimate email.
  void_estimate:
    endpoint: POST /estimates/{{ record.estimate_id }}/void
    required fields: estimate_id
    risk: Destructive Invoiced mutation: Void an estimate.
  delete_an_estimate:
    endpoint: DELETE /estimates/{{ record.estimate_id }}
    required fields: estimate_id
    risk: Deletes or cancels Invoiced data: Delete an estimate.
  convert_estimate_to_invoice:
    endpoint: POST /estimates/{{ record.estimate_id }}/invoice
    required fields: estimate_id
    risk: Invoiced mutation: Convert an estimate to an invoice.
  delete_file:
    endpoint: DELETE /files/{{ record.file_id }}
    required fields: file_id
    risk: Deletes or cancels Invoiced data: Delete a file.
  create_an_invoice:
    endpoint: POST /invoices
    risk: Invoiced mutation: Create an invoice.
  update_an_invoice:
    endpoint: PATCH /invoices/{{ record.invoice_id }}
    required fields: invoice_id
    risk: Invoiced mutation: Update an invoice.
  send_invoice_email:
    endpoint: POST /invoices/{{ record.invoice_id }}/emails
    required fields: invoice_id
    risk: Invoiced mutation: Send an invoice email.
  send_invoice_sms:
    endpoint: POST /invoices/{{ record.invoice_id }}/text_messages
    required fields: invoice_id
    risk: Invoiced mutation: Send an invoice SMS.
  send_invoice_letter:
    endpoint: POST /invoices/{{ record.invoice_id }}/letters
    required fields: invoice_id
    risk: Invoiced mutation: Send an invoice letter.
  pay_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/pay
    required fields: invoice_id
    risk: High-impact Invoiced mutation: Pay an invoice.
  create_consolidated_invoice:
    endpoint: POST /customers/{{ record.customer_id }}/consolidate_invoices
    required fields: customer_id
    risk: Invoiced mutation: Create a consolidated invoice.
  void_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/void
    required fields: invoice_id
    risk: Destructive Invoiced mutation: Void an invoice.
  delete_an_invoice:
    endpoint: DELETE /invoices/{{ record.invoice_id }}
    required fields: invoice_id
    risk: Deletes or cancels Invoiced data: Delete an invoice.
  create_an_item:
    endpoint: POST /items
    risk: Invoiced mutation: Create an item.
  update_an_item:
    endpoint: PATCH /items/{{ record.item_id }}
    required fields: item_id
    risk: Invoiced mutation: Update an item.
  delete_an_item:
    endpoint: DELETE /items/{{ record.item_id }}
    required fields: item_id
    risk: Deletes or cancels Invoiced data: Delete an item.
  create_customer_line_item:
    endpoint: POST /customers/{{ record.customer_id }}/line_items
    required fields: customer_id
    risk: Invoiced mutation: Create a metered line item.
  update_customer_line_item:
    endpoint: PATCH /customers/{{ record.customer_id }}/line_items/{{ record.line_item_id }}
    required fields: customer_id, line_item_id
    risk: Invoiced mutation: Update a metered line item.
  delete_customer_line_item:
    endpoint: DELETE /customers/{{ record.customer_id }}/line_items/{{ record.line_item_id }}
    required fields: customer_id, line_item_id
    risk: Deletes or cancels Invoiced data: Delete a metered line item.
  create_customer_invoice:
    endpoint: POST /customers/{{ record.customer_id }}/invoices
    required fields: customer_id
    risk: Invoiced mutation: Create a metered invoice.
  create_note:
    endpoint: POST /notes
    risk: Invoiced mutation: Create a note.
  update_note:
    endpoint: PATCH /notes/{{ record.note_id }}
    required fields: note_id
    risk: Invoiced mutation: Update a note.
  delete_note:
    endpoint: DELETE /notes/{{ record.note_id }}
    required fields: note_id
    risk: Deletes or cancels Invoiced data: Delete a note.
  create_payment_plan:
    endpoint: PUT /invoices/{{ record.invoice_id }}/payment_plan
    required fields: invoice_id
    risk: Invoiced mutation: Create a payment plan.
  cancel_payment_plan:
    endpoint: DELETE /invoices/{{ record.invoice_id }}/payment_plan
    required fields: invoice_id
    risk: Deletes or cancels Invoiced data: Cancel a payment plan.
  create_payment_source:
    endpoint: POST /customers/{{ record.customer_id }}/payment_sources
    required fields: customer_id
    risk: Invoiced mutation: Create a payment source.
  delete_card_payment_source:
    endpoint: DELETE /customers/{{ record.customer_id }}/cards/{{ record.card_id }}
    required fields: customer_id, card_id
    risk: Deletes or cancels Invoiced data: Delete card payment source.
  delete_bank_account_payment_source:
    endpoint: DELETE /customers/{{ record.customer_id }}/bank_accounts/{{ record.bank_account_id }}
    required fields: customer_id, bank_account_id
    risk: Deletes or cancels Invoiced data: Delete bank account payment source.
  create_payment:
    endpoint: POST /payments
    risk: Invoiced mutation: Create a payment.
  update_payment:
    endpoint: PATCH /payments/{{ record.payment_id }}
    required fields: payment_id
    risk: Invoiced mutation: Update a payment.
  send_a_payment_receipt_email:
    endpoint: POST /payments/{{ record.payment_id }}/emails
    required fields: payment_id
    risk: Invoiced mutation: Send a payment receipt email.
  delete_payment:
    endpoint: DELETE /payments/{{ record.payment_id }}
    required fields: payment_id
    risk: Deletes or cancels Invoiced data: Delete a payment.
  create_plan:
    endpoint: POST /plans
    risk: Invoiced mutation: Create a plan.
  update_plan:
    endpoint: PATCH /plans/{{ record.plan_id }}
    required fields: plan_id
    risk: Invoiced mutation: Update a plan.
  delete_plan:
    endpoint: DELETE /plans/{{ record.plan_id }}
    required fields: plan_id
    risk: Deletes or cancels Invoiced data: Delete a plan.
  refund_charge:
    endpoint: POST /charges/{{ record.charge_id }}/refunds
    required fields: charge_id
    risk: High-impact Invoiced mutation: Refund a charge.
  create_subscription:
    endpoint: POST /subscriptions
    risk: Invoiced mutation: Create a subscription.
  preview_subscription:
    endpoint: POST /subscriptions/preview
    risk: Invoiced mutation: Preview a subscription.
  update_subscription:
    endpoint: PATCH /subscriptions/{{ record.subscription_id }}
    required fields: subscription_id
    risk: Invoiced mutation: Update a subscription.
  cancel_subscription:
    endpoint: DELETE /subscriptions/{{ record.subscription_id }}
    required fields: subscription_id
    risk: Deletes or cancels Invoiced data: Cancel a subscription.
  create_task:
    endpoint: POST /tasks
    risk: Invoiced mutation: Create a task.
  update_task:
    endpoint: PATCH /tasks/{{ record.task_id }}
    required fields: task_id
    risk: Invoiced mutation: Update a task.
  delete_task:
    endpoint: DELETE /tasks/{{ record.task_id }}
    required fields: task_id
    risk: Deletes or cancels Invoiced data: Delete a task.
  create_tax_rate:
    endpoint: POST /tax_rates
    risk: Invoiced mutation: Create a tax rate.
  update_tax_rate:
    endpoint: PATCH /tax_rates/{{ record.tax_rate_id }}
    required fields: tax_rate_id
    risk: Invoiced mutation: Update a tax rate.
  delete_tax_rate:
    endpoint: DELETE /tax_rates/{{ record.tax_rate_id }}
    required fields: tax_rate_id
    risk: Deletes or cancels Invoiced data: Delete a tax rate.

SECURITY
  read risk: external Invoiced API read of billing, customer, payment, subscription, and event data
  write risk: live Invoiced API mutations can create, update, send, charge, refund, void, cancel, or delete billing records
  approval: reverse ETL writes require plan, preview, approval token, and destructive confirmation for delete/void/cancel/payment operations
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect invoiced

  # Inspect as structured JSON
  pm connectors inspect invoiced --json

AGENT WORKFLOW
  - Run pm connectors inspect invoiced before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
