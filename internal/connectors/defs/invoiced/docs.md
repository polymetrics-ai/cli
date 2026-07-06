# Overview

Reads and writes the documented Invoiced REST API surface for billing, payments, subscriptions,
events, and related resources.

Readable streams: `customers`, `invoices`, `payments`, `subscriptions`, `estimates`,
`customer_contacts`, `customer_contact`, `coupons`, `coupon`, `credit_balance_adjustments`,
`credit_balance_adjustment`, `credit_notes`, `credit_note`, `credit_note_attachments`, `customer`,
`customer_balance`, `estimate`, `estimate_attachments`, `events`, `event`, `file`, `invoice`,
`invoice_attachments`, `items`, `item`, `customer_line_items`, `customer_line_item`, `notes`,
`customer_notes`, `invoice_notes`, `invoice_payment_plan`, `customer_payment_sources`, `payment`,
`plans`, `plan`, `subscription`, `tasks`, `task`, `tax_rates`, `tax_rate`.

Write actions: `create_charge`, `create_contact`, `update_contact`, `delete_contact`,
`create_coupon`, `update_coupon`, `delete_coupon`, `create_credit_balance_adjustment`,
`update_credit_balance_adjustment`, `delete_credit_balance_adjustment`, `create_credit_note`,
`update_credit_note`, `send_credit_note_email`, `void_credit_note`, `delete_credit_note`,
`create_customer`, `update_customer`, `send_statement_email`, `send_statement_sms`,
`send_statement_letter`, `delete_customer`, `create_an_estimate`, `update_an_estimate`,
`send_estimate_email`, `void_estimate`, `delete_an_estimate`, `convert_estimate_to_invoice`,
`delete_file`, `create_an_invoice`, `update_an_invoice`, `send_invoice_email`, `send_invoice_sms`,
`send_invoice_letter`, `pay_invoice`, `create_consolidated_invoice`, `void_invoice`,
`delete_an_invoice`, `create_an_item`, `update_an_item`, `delete_an_item`,
`create_customer_line_item`, `update_customer_line_item`, `delete_customer_line_item`,
`create_customer_invoice`, `create_note`, `update_note`, `delete_note`, `create_payment_plan`,
`cancel_payment_plan`, `create_payment_source`, `delete_card_payment_source`,
`delete_bank_account_payment_source`, `create_payment`, `update_payment`,
`send_a_payment_receipt_email`, `delete_payment`, `create_plan`, `update_plan`, `delete_plan`,
`refund_charge`, `create_subscription`, `preview_subscription`, `update_subscription`,
`cancel_subscription`, `create_task`, `update_task`, `delete_task`, `create_tax_rate`,
`update_tax_rate`, `delete_tax_rate`.

Service API documentation: https://developer.invoiced.com/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Invoiced API key. Sent as the HTTP Basic username with a
  blank password; never logged.
- `base_url` (optional, string); default `https://api.invoiced.com`; format `uri`; Invoiced API base
  URL. Use https://api.sandbox.invoiced.com for sandbox accounts or an override for tests/proxies.
- `contact_id` (optional, string); Invoiced contact id used for detail or nested read streams.
- `coupon_id` (optional, string); Invoiced coupon id used for detail or nested read streams.
- `credit_balance_adjustment_id` (optional, string); Invoiced credit balance adjustment id used for
  detail or nested read streams.
- `credit_note_id` (optional, string); Invoiced credit note id used for detail or nested read
  streams.
- `customer_id` (optional, string); Invoiced customer id used for detail or nested read streams.
- `estimate_id` (optional, string); Invoiced estimate id used for detail or nested read streams.
- `event_id` (optional, string); Invoiced event id used for detail or nested read streams.
- `file_id` (optional, string); Invoiced file id used for detail or nested read streams.
- `invoice_id` (optional, string); Invoiced invoice id used for detail or nested read streams.
- `item_id` (optional, string); Invoiced item id used for detail or nested read streams.
- `line_item_id` (optional, string); Invoiced line item id used for detail or nested read streams.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-100).
- `payment_id` (optional, string); Invoiced payment id used for detail or nested read streams.
- `plan_id` (optional, string); Invoiced plan id used for detail or nested read streams.
- `subscription_id` (optional, string); Invoiced subscription id used for detail or nested read
  streams.
- `task_id` (optional, string); Invoiced task id used for detail or nested read streams.
- `tax_rate_id` (optional, string); Invoiced tax rate id used for detail or nested read streams.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.invoiced.com`, `max_pages=0`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/customers` with query `page`=`1`; `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `customer_contact`, `coupon`, `credit_balance_adjustment`,
`credit_note`, `customer`, `customer_balance`, `estimate`, `event`, `file`, `invoice`, `item`,
`customer_line_item`, `invoice_payment_plan`, `payment`, `plan`, `subscription`, `task`, `tax_rate`;
page_number: `customers`, `invoices`, `payments`, `subscriptions`, `estimates`, `customer_contacts`,
`coupons`, `credit_balance_adjustments`, `credit_notes`, `credit_note_attachments`,
`estimate_attachments`, `events`, `invoice_attachments`, `items`, `customer_line_items`, `notes`,
`customer_notes`, `invoice_notes`, `customer_payment_sources`, `plans`, `tasks`, `tax_rates`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `customers`: GET `/customers` - records at response root; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `invoices`: GET `/invoices` - records at response root; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `payments`: GET `/payments` - records at response root; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `subscriptions`: GET `/subscriptions` - records at response root; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor
  `updated_at`; formatted as `rfc3339`.
- `estimates`: GET `/estimates` - records at response root; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor `updated_at`;
  formatted as `rfc3339`.
- `customer_contacts`: GET `/customers/{{ config.customer_id }}/contacts` - records at response
  root; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page
  size 100; emits passthrough records.
- `customer_contact`: GET `/customers/{{ config.customer_id }}/contacts/{{ config.contact_id }}` -
  records at response root; emits passthrough records.
- `coupons`: GET `/coupons` - records at response root; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `coupon`: GET `/coupons/{{ config.coupon_id }}` - records at response root; emits passthrough
  records.
- `credit_balance_adjustments`: GET `/credit_balance_adjustments` - records at response root;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `credit_balance_adjustment`: GET `/credit_balance_adjustments/{{
  config.credit_balance_adjustment_id }}` - records at response root; emits passthrough records.
- `credit_notes`: GET `/credit_notes` - records at response root; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough
  records.
- `credit_note`: GET `/credit_notes/{{ config.credit_note_id }}` - records at response root; emits
  passthrough records.
- `credit_note_attachments`: GET `/credit_notes/{{ config.credit_note_id }}/attachments` - records
  at response root; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; emits passthrough records.
- `customer`: GET `/customers/{{ config.customer_id }}` - records at response root; emits
  passthrough records.
- `customer_balance`: GET `/customers/{{ config.customer_id }}/balance` - records at response root;
  emits passthrough records.
- `estimate`: GET `/estimates/{{ config.estimate_id }}` - records at response root; emits
  passthrough records.
- `estimate_attachments`: GET `/estimates/{{ config.estimate_id }}/attachments` - records at
  response root; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `events`: GET `/events` - records at response root; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `event`: GET `/events/{{ config.event_id }}` - records at response root; emits passthrough
  records.
- `file`: GET `/files/{{ config.file_id }}` - records at response root; emits passthrough records.
- `invoice`: GET `/invoices/{{ config.invoice_id }}` - records at response root; emits passthrough
  records.
- `invoice_attachments`: GET `/invoices/{{ config.invoice_id }}/attachments` - records at response
  root; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page
  size 100; emits passthrough records.
- `items`: GET `/items` - records at response root; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `item`: GET `/items/{{ config.item_id }}` - records at response root; emits passthrough records.
- `customer_line_items`: GET `/customers/{{ config.customer_id }}/line_items` - records at response
  root; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page
  size 100; emits passthrough records.
- `customer_line_item`: GET `/customers/{{ config.customer_id }}/line_items/{{ config.line_item_id
  }}` - records at response root; emits passthrough records.
- `notes`: GET `/notes` - records at response root; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `customer_notes`: GET `/customer/{{ config.customer_id }}/notes` - records at response root;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `invoice_notes`: GET `/invoice/{{ config.invoice_id }}/notes` - records at response root;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; emits passthrough records.
- `invoice_payment_plan`: GET `/invoices/{{ config.invoice_id }}/payment_plan` - records at response
  root; emits passthrough records.
- `customer_payment_sources`: GET `/customers/{{ config.customer_id }}/payment_sources` - records at
  response root; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; emits passthrough records.
- `payment`: GET `/payments/{{ config.payment_id }}` - records at response root; emits passthrough
  records.
- `plans`: GET `/plans` - records at response root; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `plan`: GET `/plans/{{ config.plan_id }}` - records at response root; emits passthrough records.
- `subscription`: GET `/subscriptions/{{ config.subscription_id }}` - records at response root;
  emits passthrough records.
- `tasks`: GET `/tasks` - records at response root; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `task`: GET `/tasks/{{ config.task_id }}` - records at response root; emits passthrough records.
- `tax_rates`: GET `/tax_rates` - records at response root; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; emits passthrough records.
- `tax_rate`: GET `/tax_rates/{{ config.tax_rate_id }}` - records at response root; emits
  passthrough records.

## Write actions & risks

Overall write risk: live Invoiced API mutations can create, update, send, charge, refund, void,
cancel, or delete billing records.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_charge`: POST `/charges` - kind `create`; body type `json`; confirmation `destructive`;
  risk: High-impact Invoiced mutation: Create a charge.
- `create_contact`: POST `/customers/{{ record.customer_id }}/contacts` - kind `create`; body type
  `json`; path fields `customer_id`; required record fields `customer_id`; accepted fields
  `customer_id`; risk: Invoiced mutation: Create a contact.
- `update_contact`: PATCH `/customers/{{ record.customer_id }}/contacts/{{ record.contact_id }}` -
  kind `update`; body type `json`; path fields `customer_id`, `contact_id`; required record fields
  `customer_id`, `contact_id`; accepted fields `contact_id`, `customer_id`; risk: Invoiced mutation:
  Update a contact.
- `delete_contact`: DELETE `/customers/{{ record.customer_id }}/contacts/{{ record.contact_id }}` -
  kind `delete`; body type `none`; path fields `customer_id`, `contact_id`; required record fields
  `customer_id`, `contact_id`; accepted fields `contact_id`, `customer_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: Deletes or cancels Invoiced data:
  Delete a contact.
- `create_coupon`: POST `/coupons` - kind `create`; body type `json`; risk: Invoiced mutation:
  Create a coupon.
- `update_coupon`: PATCH `/coupons/{{ record.coupon_id }}` - kind `update`; body type `json`; path
  fields `coupon_id`; required record fields `coupon_id`; accepted fields `coupon_id`; risk:
  Invoiced mutation: Update a coupon.
- `delete_coupon`: DELETE `/coupons/{{ record.coupon_id }}` - kind `delete`; body type `none`; path
  fields `coupon_id`; required record fields `coupon_id`; accepted fields `coupon_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes or cancels
  Invoiced data: Delete a coupon.
- `create_credit_balance_adjustment`: POST `/credit_balance_adjustments` - kind `create`; body type
  `json`; risk: Invoiced mutation: Create a credit balance adjustment.
- `update_credit_balance_adjustment`: PATCH `/credit_balance_adjustments/{{
  record.credit_balance_adjustment_id }}` - kind `update`; body type `json`; path fields
  `credit_balance_adjustment_id`; required record fields `credit_balance_adjustment_id`; accepted
  fields `credit_balance_adjustment_id`; risk: Invoiced mutation: Update a credit balance
  adjustment.
- `delete_credit_balance_adjustment`: DELETE `/credit_balance_adjustments/{{
  record.credit_balance_adjustment_id }}` - kind `delete`; body type `none`; path fields
  `credit_balance_adjustment_id`; required record fields `credit_balance_adjustment_id`; accepted
  fields `credit_balance_adjustment_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Deletes or cancels Invoiced data: Delete a credit balance
  adjustment.
- `create_credit_note`: POST `/credit_notes` - kind `create`; body type `json`; risk: Invoiced
  mutation: Create a credit note.
- `update_credit_note`: PATCH `/credit_notes/{{ record.credit_note_id }}` - kind `update`; body type
  `json`; path fields `credit_note_id`; required record fields `credit_note_id`; accepted fields
  `credit_note_id`; risk: Invoiced mutation: Update a credit note.
- `send_credit_note_email`: POST `/credit_notes/{{ record.credit_note_id }}/emails` - kind `create`;
  body type `json`; path fields `credit_note_id`; required record fields `credit_note_id`; accepted
  fields `credit_note_id`; risk: Invoiced mutation: Send a credit note email.
- `void_credit_note`: POST `/credit_notes/{{ record.credit_note_id }}/void` - kind `update`; body
  type `none`; path fields `credit_note_id`; required record fields `credit_note_id`; accepted
  fields `credit_note_id`; confirmation `destructive`; risk: Destructive Invoiced mutation: Void a
  credit note.
- `delete_credit_note`: DELETE `/credit_notes/{{ record.credit_note_id }}` - kind `delete`; body
  type `none`; path fields `credit_note_id`; required record fields `credit_note_id`; accepted
  fields `credit_note_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes or cancels Invoiced data: Delete a credit note.
- `create_customer`: POST `/customers` - kind `create`; body type `json`; risk: Invoiced mutation:
  Create a customer.
- `update_customer`: PATCH `/customers/{{ record.customer_id }}` - kind `update`; body type `json`;
  path fields `customer_id`; required record fields `customer_id`; accepted fields `customer_id`;
  risk: Invoiced mutation: Update a customer.
- `send_statement_email`: POST `/customers/{{ record.customer_id }}/emails` - kind `create`; body
  type `json`; path fields `customer_id`; required record fields `customer_id`; accepted fields
  `customer_id`; risk: Invoiced mutation: Send a statement email.
- `send_statement_sms`: POST `/customers/{{ record.customer_id }}/text_messages` - kind `create`;
  body type `json`; path fields `customer_id`; required record fields `customer_id`; accepted fields
  `customer_id`; risk: Invoiced mutation: Send a statement SMS.
- `send_statement_letter`: POST `/customers/{{ record.customer_id }}/letters` - kind `create`; body
  type `json`; path fields `customer_id`; required record fields `customer_id`; accepted fields
  `customer_id`; risk: Invoiced mutation: Send a statement letter.
- `delete_customer`: DELETE `/customers/{{ record.customer_id }}` - kind `delete`; body type `none`;
  path fields `customer_id`; required record fields `customer_id`; accepted fields `customer_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes or
  cancels Invoiced data: Delete a customer.
- `create_an_estimate`: POST `/estimates` - kind `create`; body type `json`; risk: Invoiced
  mutation: Create an estimate.
- `update_an_estimate`: PATCH `/estimates/{{ record.estimate_id }}` - kind `update`; body type
  `json`; path fields `estimate_id`; required record fields `estimate_id`; accepted fields
  `estimate_id`; risk: Invoiced mutation: Update an estimate.
- `send_estimate_email`: POST `/estimates/{{ record.estimate_id }}/emails` - kind `create`; body
  type `json`; path fields `estimate_id`; required record fields `estimate_id`; accepted fields
  `estimate_id`; risk: Invoiced mutation: Send an estimate email.
- `void_estimate`: POST `/estimates/{{ record.estimate_id }}/void` - kind `update`; body type
  `none`; path fields `estimate_id`; required record fields `estimate_id`; accepted fields
  `estimate_id`; confirmation `destructive`; risk: Destructive Invoiced mutation: Void an estimate.
- `delete_an_estimate`: DELETE `/estimates/{{ record.estimate_id }}` - kind `delete`; body type
  `none`; path fields `estimate_id`; required record fields `estimate_id`; accepted fields
  `estimate_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes or cancels Invoiced data: Delete an estimate.
- `convert_estimate_to_invoice`: POST `/estimates/{{ record.estimate_id }}/invoice` - kind `create`;
  body type `json`; path fields `estimate_id`; required record fields `estimate_id`; accepted fields
  `estimate_id`; risk: Invoiced mutation: Convert an estimate to an invoice.
- `delete_file`: DELETE `/files/{{ record.file_id }}` - kind `delete`; body type `none`; path fields
  `file_id`; required record fields `file_id`; accepted fields `file_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes or cancels Invoiced data:
  Delete a file.
- `create_an_invoice`: POST `/invoices` - kind `create`; body type `json`; risk: Invoiced mutation:
  Create an invoice.
- `update_an_invoice`: PATCH `/invoices/{{ record.invoice_id }}` - kind `update`; body type `json`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`; risk:
  Invoiced mutation: Update an invoice.
- `send_invoice_email`: POST `/invoices/{{ record.invoice_id }}/emails` - kind `create`; body type
  `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: Invoiced mutation: Send an invoice email.
- `send_invoice_sms`: POST `/invoices/{{ record.invoice_id }}/text_messages` - kind `create`; body
  type `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: Invoiced mutation: Send an invoice SMS.
- `send_invoice_letter`: POST `/invoices/{{ record.invoice_id }}/letters` - kind `create`; body type
  `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: Invoiced mutation: Send an invoice letter.
- `pay_invoice`: POST `/invoices/{{ record.invoice_id }}/pay` - kind `update`; body type `json`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`;
  confirmation `destructive`; risk: High-impact Invoiced mutation: Pay an invoice.
- `create_consolidated_invoice`: POST `/customers/{{ record.customer_id }}/consolidate_invoices` -
  kind `create`; body type `json`; path fields `customer_id`; required record fields `customer_id`;
  accepted fields `customer_id`; risk: Invoiced mutation: Create a consolidated invoice.
- `void_invoice`: POST `/invoices/{{ record.invoice_id }}/void` - kind `update`; body type `none`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`;
  confirmation `destructive`; risk: Destructive Invoiced mutation: Void an invoice.
- `delete_an_invoice`: DELETE `/invoices/{{ record.invoice_id }}` - kind `delete`; body type `none`;
  path fields `invoice_id`; required record fields `invoice_id`; accepted fields `invoice_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes or
  cancels Invoiced data: Delete an invoice.
- `create_an_item`: POST `/items` - kind `create`; body type `json`; risk: Invoiced mutation: Create
  an item.
- `update_an_item`: PATCH `/items/{{ record.item_id }}` - kind `update`; body type `json`; path
  fields `item_id`; required record fields `item_id`; accepted fields `item_id`; risk: Invoiced
  mutation: Update an item.
- `delete_an_item`: DELETE `/items/{{ record.item_id }}` - kind `delete`; body type `none`; path
  fields `item_id`; required record fields `item_id`; accepted fields `item_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes or cancels Invoiced
  data: Delete an item.
- `create_customer_line_item`: POST `/customers/{{ record.customer_id }}/line_items` - kind
  `create`; body type `json`; path fields `customer_id`; required record fields `customer_id`;
  accepted fields `customer_id`; risk: Invoiced mutation: Create a metered line item.
- `update_customer_line_item`: PATCH `/customers/{{ record.customer_id }}/line_items/{{
  record.line_item_id }}` - kind `update`; body type `json`; path fields `customer_id`,
  `line_item_id`; required record fields `customer_id`, `line_item_id`; accepted fields
  `customer_id`, `line_item_id`; risk: Invoiced mutation: Update a metered line item.
- `delete_customer_line_item`: DELETE `/customers/{{ record.customer_id }}/line_items/{{
  record.line_item_id }}` - kind `delete`; body type `none`; path fields `customer_id`,
  `line_item_id`; required record fields `customer_id`, `line_item_id`; accepted fields
  `customer_id`, `line_item_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes or cancels Invoiced data: Delete a metered line item.
- `create_customer_invoice`: POST `/customers/{{ record.customer_id }}/invoices` - kind `create`;
  body type `json`; path fields `customer_id`; required record fields `customer_id`; accepted fields
  `customer_id`; risk: Invoiced mutation: Create a metered invoice.
- `create_note`: POST `/notes` - kind `create`; body type `json`; risk: Invoiced mutation: Create a
  note.
- `update_note`: PATCH `/notes/{{ record.note_id }}` - kind `update`; body type `json`; path fields
  `note_id`; required record fields `note_id`; accepted fields `note_id`; risk: Invoiced mutation:
  Update a note.
- `delete_note`: DELETE `/notes/{{ record.note_id }}` - kind `delete`; body type `none`; path fields
  `note_id`; required record fields `note_id`; accepted fields `note_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes or cancels Invoiced data:
  Delete a note.
- `create_payment_plan`: PUT `/invoices/{{ record.invoice_id }}/payment_plan` - kind `upsert`; body
  type `json`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; risk: Invoiced mutation: Create a payment plan.
- `cancel_payment_plan`: DELETE `/invoices/{{ record.invoice_id }}/payment_plan` - kind `delete`;
  body type `none`; path fields `invoice_id`; required record fields `invoice_id`; accepted fields
  `invoice_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Deletes or cancels Invoiced data: Cancel a payment plan.
- `create_payment_source`: POST `/customers/{{ record.customer_id }}/payment_sources` - kind
  `create`; body type `json`; path fields `customer_id`; required record fields `customer_id`;
  accepted fields `customer_id`; risk: Invoiced mutation: Create a payment source.
- `delete_card_payment_source`: DELETE `/customers/{{ record.customer_id }}/cards/{{ record.card_id
  }}` - kind `delete`; body type `none`; path fields `customer_id`, `card_id`; required record
  fields `customer_id`, `card_id`; accepted fields `card_id`, `customer_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: Deletes or cancels Invoiced data:
  Delete card payment source.
- `delete_bank_account_payment_source`: DELETE `/customers/{{ record.customer_id }}/bank_accounts/{{
  record.bank_account_id }}` - kind `delete`; body type `none`; path fields `customer_id`,
  `bank_account_id`; required record fields `customer_id`, `bank_account_id`; accepted fields
  `bank_account_id`, `customer_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Deletes or cancels Invoiced data: Delete bank account payment
  source.
- `create_payment`: POST `/payments` - kind `create`; body type `json`; risk: Invoiced mutation:
  Create a payment.
- `update_payment`: PATCH `/payments/{{ record.payment_id }}` - kind `update`; body type `json`;
  path fields `payment_id`; required record fields `payment_id`; accepted fields `payment_id`; risk:
  Invoiced mutation: Update a payment.
- `send_a_payment_receipt_email`: POST `/payments/{{ record.payment_id }}/emails` - kind `create`;
  body type `json`; path fields `payment_id`; required record fields `payment_id`; accepted fields
  `payment_id`; risk: Invoiced mutation: Send a payment receipt email.
- `delete_payment`: DELETE `/payments/{{ record.payment_id }}` - kind `delete`; body type `none`;
  path fields `payment_id`; required record fields `payment_id`; accepted fields `payment_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes or
  cancels Invoiced data: Delete a payment.
- `create_plan`: POST `/plans` - kind `create`; body type `json`; risk: Invoiced mutation: Create a
  plan.
- `update_plan`: PATCH `/plans/{{ record.plan_id }}` - kind `update`; body type `json`; path fields
  `plan_id`; required record fields `plan_id`; accepted fields `plan_id`; risk: Invoiced mutation:
  Update a plan.
- `delete_plan`: DELETE `/plans/{{ record.plan_id }}` - kind `delete`; body type `none`; path fields
  `plan_id`; required record fields `plan_id`; accepted fields `plan_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes or cancels Invoiced data:
  Delete a plan.
- `refund_charge`: POST `/charges/{{ record.charge_id }}/refunds` - kind `update`; body type `json`;
  path fields `charge_id`; required record fields `charge_id`; accepted fields `charge_id`;
  confirmation `destructive`; risk: High-impact Invoiced mutation: Refund a charge.
- `create_subscription`: POST `/subscriptions` - kind `create`; body type `json`; risk: Invoiced
  mutation: Create a subscription.
- `preview_subscription`: POST `/subscriptions/preview` - kind `custom`; body type `json`; risk:
  Invoiced mutation: Preview a subscription.
- `update_subscription`: PATCH `/subscriptions/{{ record.subscription_id }}` - kind `update`; body
  type `json`; path fields `subscription_id`; required record fields `subscription_id`; accepted
  fields `subscription_id`; risk: Invoiced mutation: Update a subscription.
- `cancel_subscription`: DELETE `/subscriptions/{{ record.subscription_id }}` - kind `delete`; body
  type `none`; path fields `subscription_id`; required record fields `subscription_id`; accepted
  fields `subscription_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Deletes or cancels Invoiced data: Cancel a subscription.
- `create_task`: POST `/tasks` - kind `create`; body type `json`; risk: Invoiced mutation: Create a
  task.
- `update_task`: PATCH `/tasks/{{ record.task_id }}` - kind `update`; body type `json`; path fields
  `task_id`; required record fields `task_id`; accepted fields `task_id`; risk: Invoiced mutation:
  Update a task.
- `delete_task`: DELETE `/tasks/{{ record.task_id }}` - kind `delete`; body type `none`; path fields
  `task_id`; required record fields `task_id`; accepted fields `task_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes or cancels Invoiced data:
  Delete a task.
- `create_tax_rate`: POST `/tax_rates` - kind `create`; body type `json`; risk: Invoiced mutation:
  Create a tax rate.
- `update_tax_rate`: PATCH `/tax_rates/{{ record.tax_rate_id }}` - kind `update`; body type `json`;
  path fields `tax_rate_id`; required record fields `tax_rate_id`; accepted fields `tax_rate_id`;
  risk: Invoiced mutation: Update a tax rate.
- `delete_tax_rate`: DELETE `/tax_rates/{{ record.tax_rate_id }}` - kind `delete`; body type `none`;
  path fields `tax_rate_id`; required record fields `tax_rate_id`; accepted fields `tax_rate_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes or
  cancels Invoiced data: Delete a tax rate.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 40 stream-backed endpoint group(s), 70 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1.
