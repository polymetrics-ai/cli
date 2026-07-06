# Overview

Reads and writes Zoho Expense API v1 resources through the connector engine.

Readable streams: `reports`, `expenses`, `users`, `currencies`, `currency`, `customers`, `customer`,
`advance_payments`, `advance_payment`, `expense_categories`, `expense_category`, `expense_report`,
`expense_report_approval_history`, `expense_report_validation`, `expense_report_budget_summary`,
`expense_report_reimbursement`, `organizations`, `organization`, `projects`, `project`,
`reporting_tags`, `tag_options`, `reporting_tag_options_all`, `taxes`, `tax`, `tax_group`, `trips`,
`trip`, `user`.

Write actions: `create_currency`, `update_currency`, `delete_currency`, `create_customer`,
`update_customer`, `delete_customer`, `create_advance_payment`, `update_advance_payment`,
`delete_advance_payment`, `create_expense_category`, `update_expense_category`,
`delete_expense_category`, `enable_expense_category`, `disable_expense_category`,
`create_expense_report`, `bulk_delete_expense_reports`, `update_expense_report`,
`delete_expense_report`, `submit_expense_report`, `approve_expense_report`, `reject_expense_report`,
`reimburse_expense_report`, `takeback_expense_report`, `archive_expense_report`,
`unarchive_expense_report`, `share_expense_report`, `forward_approval_expense_report`,
`reject_expense_in_report`, `bulk_reject_expenses_in_report`, `add_comment_to_expense_report`,
`delete_comment_of_expense_report`, `associate_tags_to_expense_report`,
`add_expenses_to_expense_report`, `remove_expenses_from_expense_report`,
`delete_expense_report_attachment`, `reupload_expense_report_attachment`,
`remove_advance_payment_from_expense_report`, `change_status_of_expense_report`,
`export_expense_report`, `cancel_reimbursement_of_expense_report`, `add_expense_to_report`,
`bulk_submit_expense_reports`, `bulk_approve_expense_reports`, `bulk_reject_expense_reports`,
`bulk_reimburse_expense_reports`, `bulk_archive_expense_reports`, `bulk_unarchive_expense_reports`,
`bulk_forward_approval_expense_reports`, `bulk_change_status_of_expense_reports`,
`bulk_reset_substatus_of_expense_reports`, `bulk_field_update_expense_reports`, `create_expense`,
`update_expense`, `merge_expenses`, `create_organization`, `update_organization`, `create_project`,
`update_project`, `delete_project`, `activate_project`, `deactivate_project`, `create_tag`,
`mark_default_option`, `update_tag`, `delete_tag`, `update_tag_options`, `update_tag_criteria`,
`active_tag`, `inactive_tag`, `active_tag_option`, `inactive_tag_option`, `reorder_tags`,
`create_tax`, `update_tax`, `delete_tax`, `create_trip`, `update_trip`, `delete_trip`,
`approve_trip`, `reject_trip`, and 8 more.

Service API documentation: https://www.zoho.com/expense/api/v1/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Zoho OAuth access token. Sent as the Authorization
  header with a 'Zoho-oauthtoken ' prefix; never logged.
- `advance_payment_id` (optional, string); Optional advance_payment_id used by Zoho Expense detail
  streams.
- `base_url` (optional, string); default `https://www.zohoapis.com/expense/v1`; format `uri`; Zoho
  Expense API base URL override for tests or region-specific data centers.
- `currency_id` (optional, string); Optional currency_id used by Zoho Expense detail streams.
- `customer_id` (optional, string); Optional customer_id used by Zoho Expense detail streams.
- `expense_category_id` (optional, string); Optional expense_category_id used by Zoho Expense detail
  streams.
- `expense_report_id` (optional, string); Optional expense_report_id used by Zoho Expense detail
  streams.
- `mode` (optional, string).
- `organization_id` (optional, string); Zoho Expense organization ID. Sent as the
  X-com-zoho-expense-organizationid header for organization-scoped endpoints and as the documented
  organization_id query parameter for reporting tag endpoints.
- `project_id` (optional, string); Optional project_id used by Zoho Expense detail streams.
- `tag_id` (optional, string); Optional tag_id used by Zoho Expense detail streams.
- `tax_group_id` (optional, string); Optional tax_group_id used by Zoho Expense detail streams.
- `tax_id` (optional, string); Optional tax_id used by Zoho Expense detail streams.
- `trip_id` (optional, string); Optional trip_id used by Zoho Expense detail streams.
- `user_id` (optional, string); Optional user_id used by Zoho Expense detail streams.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://www.zohoapis.com/expense/v1`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Zoho-oauthtoken` using
  `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/expensereports`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 200.

Pagination by stream: none: `currency`, `customer`, `advance_payment`, `expense_category`,
`expense_report`, `expense_report_approval_history`, `expense_report_validation`,
`expense_report_budget_summary`, `expense_report_reimbursement`, `organizations`, `organization`,
`project`, `reporting_tags`, `tag_options`, `reporting_tag_options_all`, `tax`, `tax_group`, `trip`,
`user`; page_number: `reports`, `expenses`, `users`, `currencies`, `customers`, `advance_payments`,
`expense_categories`, `projects`, `taxes`, `trips`.

- `reports`: GET `/expensereports` - records path `expense_reports`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields
  `id`, `name`, `updated_at`; emits passthrough records.
- `expenses`: GET `/expenses` - records path `expenses`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; computed output fields `id`,
  `name`, `updated_at`; emits passthrough records.
- `users`: GET `/users` - records path `users`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 200; computed output fields `id`, `name`,
  `updated_at`; emits passthrough records.
- `currencies`: GET `/settings/currencies` - records path `currencies`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 200; emits passthrough
  records.
- `currency`: GET `/settings/currencies/{{ config.currency_id }}` - single-object response; records
  path `currency`; emits passthrough records.
- `customers`: GET `/contacts` - records path `customers`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `customer`: GET `/contacts/{{ config.customer_id }}` - single-object response; records path
  `contact`; emits passthrough records.
- `advance_payments`: GET `/advancepayments` - records path `advance_payments`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; emits
  passthrough records.
- `advance_payment`: GET `/advancepayments/{{ config.advance_payment_id }}` - single-object
  response; records path `advance_payment`; emits passthrough records.
- `expense_categories`: GET `/expensecategories` - records path `expense_categories`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 200; emits
  passthrough records.
- `expense_category`: GET `/expensecategories/{{ config.expense_category_id }}` - single-object
  response; records path `expense_category`; emits passthrough records.
- `expense_report`: GET `/expensereports/{{ config.expense_report_id }}` - single-object response;
  records path `expense_report`; emits passthrough records.
- `expense_report_approval_history`: GET `/expensereports/{{ config.expense_report_id
  }}/approvalhistory` - records path `approval_history`; emits passthrough records.
- `expense_report_validation`: GET `/expensereports/{{ config.expense_report_id }}/validate` -
  single-object response; records path `.`; emits passthrough records.
- `expense_report_budget_summary`: GET `/expensereports/{{ config.expense_report_id
  }}/budgetsummary` - single-object response; records path `budget_summary`; emits passthrough
  records.
- `expense_report_reimbursement`: GET `/expensereports/{{ config.expense_report_id }}/reimbursement`
  - single-object response; records path `reimbursement`; emits passthrough records.
- `organizations`: GET `/organizations` - records path `organizations`; emits passthrough records.
- `organization`: GET `/organizations/{{ config.organization_id }}` - single-object response;
  records path `organization`; emits passthrough records.
- `projects`: GET `/projects` - records path `projects`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `project`: GET `/projects/{{ config.project_id }}` - single-object response; records path
  `project`; emits passthrough records.
- `reporting_tags`: GET `/reportingtags` - records path `reporting_tags`; query
  `organization_id`=`{{ config.organization_id }}`; emits passthrough records.
- `tag_options`: GET `/reportingtags/options` - records path `options`; query `organization_id`=`{{
  config.organization_id }}`; `tag_id`=`{{ config.tag_id }}`; emits passthrough records.
- `reporting_tag_options_all`: GET `/reportingtags/{{ config.tag_id }}/options/all` - records path
  `results`; query `organization_id`=`{{ config.organization_id }}`; `tag_id`=`{{ config.tag_id }}`;
  emits passthrough records.
- `taxes`: GET `/settings/taxes` - records path `taxes`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `tax`: GET `/settings/taxes/{{ config.tax_id }}` - single-object response; records path `tax`;
  emits passthrough records.
- `tax_group`: GET `/settings/taxgroups/{{ config.tax_group_id }}` - single-object response; records
  path `tax_group`; emits passthrough records.
- `trips`: GET `/trips` - records path `trips`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 200; emits passthrough records.
- `trip`: GET `/trips/{{ config.trip_id }}` - single-object response; records path `trip`; emits
  passthrough records.
- `user`: GET `/users/{{ config.user_id }}` - single-object response; records path `user`; emits
  passthrough records.

## Write actions & risks

Overall write risk: creates, updates, submits, approves, reimburses, archives, deletes, and
otherwise mutates Zoho Expense resources; requires explicit approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_currency`: POST `/settings/currencies` - kind `create`; body type `json`; accepted fields
  `currency_code`, `currency_format`, `currency_name`, `currency_symbol`, `price_precision`; risk:
  creates Zoho Expense resources through create_currency; approval required.
- `update_currency`: PUT `/settings/currencies/{{ record.currency_id }}` - kind `update`; body type
  `json`; path fields `currency_id`; required record fields `currency_id`; accepted fields
  `currency_format`, `currency_id`, `currency_symbol`, `price_precision`; risk: mutates Zoho Expense
  resources through update_currency; approval required.
- `delete_currency`: DELETE `/settings/currencies/{{ record.currency_id }}` - kind `delete`; body
  type `none`; path fields `currency_id`; required record fields `currency_id`; accepted fields
  `currency_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: deletes Zoho Expense resources through delete_currency; approval required.
- `create_customer`: POST `/contacts` - kind `create`; body type `json`; accepted fields
  `billing_address`, `company_name`, `contact_name`, `contact_persons`, `custom_fields`, `email`,
  `facebook`, `mobile`, `phone`, `shipping_address`, `twitter`, `website`; risk: creates Zoho
  Expense resources through create_customer; approval required.
- `update_customer`: PUT `/contacts/{{ record.customer_id }}` - kind `update`; body type `json`;
  path fields `customer_id`; required record fields `customer_id`; accepted fields
  `billing_address`, `company_name`, `contact_name`, `custom_fields`, `customer_id`, `email`,
  `facebook`, `mobile`, `notes`, `phone`, `website`; risk: mutates Zoho Expense resources through
  update_customer; approval required.
- `delete_customer`: DELETE `/contacts/{{ record.customer_id }}` - kind `delete`; body type `none`;
  path fields `customer_id`; required record fields `customer_id`; accepted fields `customer_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: deletes
  Zoho Expense resources through delete_customer; approval required.
- `create_advance_payment`: POST `/advancepayments` - kind `create`; body type `json`; accepted
  fields `account_id`, `amount`, `currency_id`, `custom_fields`, `customer_id`, `date`,
  `exchange_rate`, `location_id`, `notes`, `project_id`, `rcy_exchange_rate`, `reference_number`,
  `trip_id`, `user_id`; risk: creates Zoho Expense resources through create_advance_payment;
  approval required.
- `update_advance_payment`: PUT `/advancepayments/{{ record.advance_payment_id }}` - kind `update`;
  body type `json`; path fields `advance_payment_id`; required record fields `advance_payment_id`;
  accepted fields `account_id`, `advance_payment_id`, `amount`, `currency_id`, `custom_fields`,
  `customer_id`, `date`, `exchange_rate`, `location_id`, `notes`, `project_id`, `rcy_exchange_rate`,
  `reference_number`, `trip_id`, `user_id`; risk: mutates Zoho Expense resources through
  update_advance_payment; approval required.
- `delete_advance_payment`: DELETE `/advancepayments/{{ record.advance_payment_id }}` - kind
  `delete`; body type `none`; path fields `advance_payment_id`; required record fields
  `advance_payment_id`; accepted fields `advance_payment_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes Zoho Expense resources through
  delete_advance_payment; approval required.
- `create_expense_category`: POST `/expensecategories` - kind `create`; body type `json`; accepted
  fields `category_name`, `description`, `flat_amount`, `gl_code`, `maximum_allowed_amount`,
  `parent_account_id`, `receipt_required_amount`; risk: creates Zoho Expense resources through
  create_expense_category; approval required.
- `update_expense_category`: PUT `/expensecategories/{{ record.expense_category_id }}` - kind
  `update`; body type `json`; path fields `expense_category_id`; required record fields
  `expense_category_id`; accepted fields `category_name`, `description`, `expense_category_id`,
  `flat_amount`, `gl_code`, `maximum_allowed_amount`, `receipt_required_amount`; risk: mutates Zoho
  Expense resources through update_expense_category; approval required.
- `delete_expense_category`: DELETE `/expensecategories/{{ record.expense_category_id }}` - kind
  `delete`; body type `none`; path fields `expense_category_id`; required record fields
  `expense_category_id`; accepted fields `expense_category_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes Zoho Expense resources through
  delete_expense_category; approval required.
- `enable_expense_category`: POST `/expensecategories/{{ record.expense_category_id }}/show` - kind
  `update`; body type `none`; path fields `expense_category_id`; required record fields
  `expense_category_id`; accepted fields `expense_category_id`; risk: mutates Zoho Expense resources
  through enable_expense_category; approval required.
- `disable_expense_category`: POST `/expensecategories/{{ record.expense_category_id }}/hide` - kind
  `update`; body type `none`; path fields `expense_category_id`; required record fields
  `expense_category_id`; accepted fields `expense_category_id`; risk: mutates Zoho Expense resources
  through disable_expense_category; approval required.
- `create_expense_report`: POST `/expensereports` - kind `create`; body type `json`; accepted fields
  `custom_fields`, `customer_id`, `description`, `end_date`, `expenses`, `project_id`,
  `report_name`, `start_date`, `tags`; risk: creates Zoho Expense resources through
  create_expense_report; approval required.
- `bulk_delete_expense_reports`: DELETE `/expensereports` - kind `delete`; body type `json`;
  accepted fields `report_ids`; confirmation `destructive`; risk: deletes Zoho Expense resources
  through bulk_delete_expense_reports; approval required.
- `update_expense_report`: PUT `/expensereports/{{ record.expense_report_id }}` - kind `update`;
  body type `json`; path fields `expense_report_id`; required record fields `expense_report_id`;
  accepted fields `custom_fields`, `customer_id`, `customfield_id`, `description`, `end_date`,
  `expense_id`, `expense_report_id`, `expenses`, `project_id`, `report_name`, `start_date`,
  `tag_id`, `tag_option_id`, `tags`, `value`; risk: mutates Zoho Expense resources through
  update_expense_report; approval required.
- `delete_expense_report`: DELETE `/expensereports/{{ record.expense_report_id }}` - kind `delete`;
  body type `none`; path fields `expense_report_id`; required record fields `expense_report_id`;
  accepted fields `expense_report_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes Zoho Expense resources through delete_expense_report;
  approval required.
- `submit_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/submit` - kind
  `update`; body type `json`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `approver_id`, `cc_mail_ids`, `expense_report_id`; risk:
  mutates Zoho Expense resources through submit_expense_report; approval required.
- `approve_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/approve` - kind
  `update`; body type `none`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `expense_report_id`; risk: mutates Zoho Expense resources
  through approve_expense_report; approval required.
- `reject_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/reject` - kind
  `update`; body type `json`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `comments`, `expense_report_id`; risk: mutates Zoho Expense
  resources through reject_expense_report; approval required.
- `reimburse_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/reimburse` - kind
  `update`; body type `json`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `amount`, `currency_id`, `date`, `expense_report_id`,
  `notes`, `reference_number`; risk: mutates Zoho Expense resources through
  reimburse_expense_report; approval required.
- `takeback_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/takeback` - kind
  `update`; body type `json`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `comments`, `expense_report_id`; risk: mutates Zoho Expense
  resources through takeback_expense_report; approval required.
- `archive_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/archive` - kind
  `update`; body type `none`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `expense_report_id`; risk: mutates Zoho Expense resources
  through archive_expense_report; approval required.
- `unarchive_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/unarchive` - kind
  `update`; body type `none`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `expense_report_id`; risk: mutates Zoho Expense resources
  through unarchive_expense_report; approval required.
- `share_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/share` - kind
  `update`; body type `json`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `expense_report_id`, `user_ids`; risk: mutates Zoho Expense
  resources through share_expense_report; approval required.
- `forward_approval_expense_report`: POST `/expensereports/{{ record.expense_report_id
  }}/forwardapproval` - kind `update`; body type `json`; path fields `expense_report_id`; required
  record fields `expense_report_id`; accepted fields `approval_reason`, `approver_id`, `comments`,
  `expense_report_id`; risk: mutates Zoho Expense resources through forward_approval_expense_report;
  approval required.
- `reject_expense_in_report`: POST `/expensereports/{{ record.expense_report_id }}/expense/{{
  record.expense_id }}/reject` - kind `update`; body type `json`; path fields `expense_report_id`,
  `expense_id`; required record fields `expense_report_id`, `expense_id`; accepted fields
  `comments`, `expense_id`, `expense_report_id`; risk: mutates Zoho Expense resources through
  reject_expense_in_report; approval required.
- `bulk_reject_expenses_in_report`: POST `/expensereports/{{ record.expense_report_id
  }}/expensebulkreject` - kind `update`; body type `json`; path fields `expense_report_id`; required
  record fields `expense_report_id`; accepted fields `comments`, `expense_ids`, `expense_report_id`;
  risk: mutates Zoho Expense resources through bulk_reject_expenses_in_report; approval required.
- `add_comment_to_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/comments` -
  kind `create`; body type `json`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `comments`, `expense_id`, `expense_report_id`; risk: creates
  Zoho Expense resources through add_comment_to_expense_report; approval required.
- `delete_comment_of_expense_report`: DELETE `/expensereports/{{ record.expense_report_id
  }}/comments/{{ record.comment_id }}` - kind `delete`; body type `none`; path fields
  `expense_report_id`, `comment_id`; required record fields `expense_report_id`, `comment_id`;
  accepted fields `comment_id`, `expense_report_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: deletes Zoho Expense resources through
  delete_comment_of_expense_report; approval required.
- `associate_tags_to_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/tags` -
  kind `update`; body type `json`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `expense_report_id`, `tags`; risk: mutates Zoho Expense
  resources through associate_tags_to_expense_report; approval required.
- `add_expenses_to_expense_report`: POST `/expensereports/{{ record.expense_report_id
  }}/addexpenses` - kind `create`; body type `json`; path fields `expense_report_id`; required
  record fields `expense_report_id`; accepted fields `expense_ids`, `expense_report_id`; risk:
  creates Zoho Expense resources through add_expenses_to_expense_report; approval required.
- `remove_expenses_from_expense_report`: POST `/expensereports/{{ record.expense_report_id
  }}/removeexpenses` - kind `update`; body type `json`; path fields `expense_report_id`; required
  record fields `expense_report_id`; accepted fields `expense_ids`, `expense_report_id`; risk:
  mutates Zoho Expense resources through remove_expenses_from_expense_report; approval required.
- `delete_expense_report_attachment`: DELETE `/expensereports/{{ record.expense_report_id
  }}/documents/{{ record.document_id }}` - kind `delete`; body type `none`; path fields
  `expense_report_id`, `document_id`; required record fields `expense_report_id`, `document_id`;
  accepted fields `document_id`, `expense_report_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: deletes Zoho Expense resources through
  delete_expense_report_attachment; approval required.
- `reupload_expense_report_attachment`: POST `/expensereports/{{ record.expense_report_id
  }}/documents/{{ record.document_id }}/upload` - kind `update`; body type `json`; path fields
  `expense_report_id`, `document_id`; required record fields `expense_report_id`, `document_id`;
  accepted fields `doc`, `docName`, `document_id`, `expense_report_id`, `folderId`, `source`; risk:
  mutates Zoho Expense resources through reupload_expense_report_attachment; approval required.
- `remove_advance_payment_from_expense_report`: POST `/expensereports/{{ record.expense_report_id
  }}/removeadvancepayment` - kind `update`; body type `json`; path fields `expense_report_id`;
  required record fields `expense_report_id`; accepted fields `advance_payment_ids`,
  `expense_report_id`; risk: mutates Zoho Expense resources through
  remove_advance_payment_from_expense_report; approval required.
- `change_status_of_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/status/{{
  record.action }}` - kind `update`; body type `none`; path fields `expense_report_id`, `action`;
  required record fields `expense_report_id`, `action`; accepted fields `action`,
  `expense_report_id`; risk: mutates Zoho Expense resources through change_status_of_expense_report;
  approval required.
- `export_expense_report`: POST `/expensereports/{{ record.expense_report_id }}/export` - kind
  `update`; body type `none`; path fields `expense_report_id`; required record fields
  `expense_report_id`; accepted fields `expense_report_id`; risk: mutates Zoho Expense resources
  through export_expense_report; approval required.
- `cancel_reimbursement_of_expense_report`: POST `/expensereports/{{ record.expense_report_id
  }}/reimburse/cancel` - kind `update`; body type `none`; path fields `expense_report_id`; required
  record fields `expense_report_id`; accepted fields `expense_report_id`; risk: mutates Zoho Expense
  resources through cancel_reimbursement_of_expense_report; approval required.
- `add_expense_to_report`: POST `/expenses/{{ record.expense_id }}/addtoreport` - kind `create`;
  body type `json`; path fields `expense_id`; required record fields `expense_id`; accepted fields
  `expense_id`, `report_id`; risk: creates Zoho Expense resources through add_expense_to_report;
  approval required.
- `bulk_submit_expense_reports`: POST `/expensereports/submit` - kind `update`; body type `json`;
  accepted fields `approver_id`, `report_ids`; risk: mutates Zoho Expense resources through
  bulk_submit_expense_reports; approval required.
- `bulk_approve_expense_reports`: POST `/expensereports/approve` - kind `update`; body type `json`;
  accepted fields `report_ids`; risk: mutates Zoho Expense resources through
  bulk_approve_expense_reports; approval required.
- `bulk_reject_expense_reports`: POST `/expensereports/reject` - kind `update`; body type `json`;
  accepted fields `comments`, `report_ids`; risk: mutates Zoho Expense resources through
  bulk_reject_expense_reports; approval required.
- `bulk_reimburse_expense_reports`: POST `/expensereports/reimburse` - kind `update`; body type
  `json`; accepted fields `reference_number`, `reimbursed_date`, `report_ids`; risk: mutates Zoho
  Expense resources through bulk_reimburse_expense_reports; approval required.
- `bulk_archive_expense_reports`: POST `/expensereports/archive` - kind `update`; body type `json`;
  accepted fields `report_ids`; risk: mutates Zoho Expense resources through
  bulk_archive_expense_reports; approval required.
- `bulk_unarchive_expense_reports`: POST `/expensereports/unarchive` - kind `update`; body type
  `json`; accepted fields `report_ids`; risk: mutates Zoho Expense resources through
  bulk_unarchive_expense_reports; approval required.
- `bulk_forward_approval_expense_reports`: POST `/expensereports/forwardapproval` - kind `update`;
  body type `json`; accepted fields `approver_id`, `comments`, `report_ids`; risk: mutates Zoho
  Expense resources through bulk_forward_approval_expense_reports; approval required.
- `bulk_change_status_of_expense_reports`: POST `/expensereports/status/{{ record.action }}` - kind
  `update`; body type `json`; path fields `action`; required record fields `action`; accepted fields
  `action`, `report_ids`; risk: mutates Zoho Expense resources through
  bulk_change_status_of_expense_reports; approval required.
- `bulk_reset_substatus_of_expense_reports`: POST `/expensereports/substatus/reset` - kind `update`;
  body type `json`; accepted fields `report_ids`; risk: mutates Zoho Expense resources through
  bulk_reset_substatus_of_expense_reports; approval required.
- `bulk_field_update_expense_reports`: POST `/expensereports/bulkfieldupdate` - kind `update`; body
  type `json`; accepted fields `fields`, `report_ids`; risk: mutates Zoho Expense resources through
  bulk_field_update_expense_reports; approval required.
- `create_expense`: POST `/expenses` - kind `create`; body type `json`; accepted fields `attendees`,
  `currency_id`, `custom_fields`, `customer_id`, `date`, `distance`, `is_billable`,
  `is_inclusive_tax`, `is_reimbursable`, `line_items`, `merchant_id`, `paid_through_account_id`,
  `payment_mode`, `project_id`, `report_id`; risk: creates Zoho Expense resources through
  create_expense; approval required.
- `update_expense`: PUT `/expenses/{{ record.expense_id }}` - kind `update`; body type `json`; path
  fields `expense_id`; required record fields `expense_id`; accepted fields `attendees`,
  `currency_id`, `custom_fields`, `customer_id`, `date`, `expense_id`, `is_billable`,
  `is_inclusive_tax`, `is_reimbursable`, `line_items`, `merchant_id`, `paid_through_account_id`,
  `payment_mode`, `project_id`, `report_id`; risk: mutates Zoho Expense resources through
  update_expense; approval required.
- `merge_expenses`: POST `/expenses/{{ record.expense_id }}/merge?duplicate_expense_id={{
  record.duplicate_expense_id }}` - kind `update`; body type `none`; path fields `expense_id`,
  `duplicate_expense_id`; required record fields `expense_id`, `duplicate_expense_id`; accepted
  fields `duplicate_expense_id`, `expense_id`; risk: mutates Zoho Expense resources through
  merge_expenses; approval required.
- `create_organization`: POST `/organizations` - kind `create`; body type `json`; accepted fields
  `address`, `currency_code`, `date_format`, `field_separator`, `fiscal_year_start_month`,
  `industry_size`, `industry_type`, `language_code`, `name`, `org_address`, `portal_name`,
  `remit_to_address`, `time_zone`; risk: creates Zoho Expense resources through create_organization;
  approval required.
- `update_organization`: PUT `/organizations/{{ record.organization_id }}` - kind `update`; body
  type `json`; path fields `organization_id`; required record fields `organization_id`; accepted
  fields `address`, `companyid_label`, `companyid_value`, `contact_name`, `currency_id`,
  `custom_fields`, `date_format`, `email`, `fax`, `field_separator`, `fiscal_year_start_month`,
  `is_logo_uploaded`, `language_code`, `name`, `org_address`, `organization_id`, `phone`,
  `remit_to_address`, and 4 more; risk: mutates Zoho Expense resources through update_organization;
  approval required.
- `create_project`: POST `/projects` - kind `create`; body type `json`; accepted fields
  `billing_type`, `custom_fields`, `customer_id`, `description`, `project_head_id`, `project_name`,
  `show_to_all_users`, `users`; risk: creates Zoho Expense resources through create_project;
  approval required.
- `update_project`: PUT `/projects/{{ record.project_id }}` - kind `update`; body type `json`; path
  fields `project_id`; required record fields `project_id`; accepted fields `billing_type`,
  `custom_fields`, `customer_id`, `description`, `project_id`, `project_name`; risk: mutates Zoho
  Expense resources through update_project; approval required.
- `delete_project`: DELETE `/projects/{{ record.project_id }}` - kind `delete`; body type `none`;
  path fields `project_id`; required record fields `project_id`; accepted fields `project_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: deletes
  Zoho Expense resources through delete_project; approval required.
- `activate_project`: POST `/projects/{{ record.project_id }}/active` - kind `update`; body type
  `none`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: mutates Zoho Expense resources through activate_project; approval required.
- `deactivate_project`: POST `/projects/{{ record.project_id }}/inactive` - kind `update`; body type
  `none`; path fields `project_id`; required record fields `project_id`; accepted fields
  `project_id`; risk: mutates Zoho Expense resources through deactivate_project; approval required.
- `create_tag`: POST `/reportingtags?organization_id={{ config.organization_id }}` - kind `create`;
  body type `json`; accepted fields `description`, `entities`, `is_mandatory`,
  `multi_preference_entities`, `tag_name`; risk: creates Zoho Expense resources through create_tag;
  approval required.
- `mark_default_option`: POST `/reportingtags/{{ record.tag_id }}?organization_id={{
  config.organization_id }}&default_option_id={{ record.default_option_id }}` - kind `update`; body
  type `none`; path fields `tag_id`, `default_option_id`; required record fields `tag_id`,
  `default_option_id`; accepted fields `default_option_id`, `tag_id`; risk: mutates Zoho Expense
  resources through mark_default_option; approval required.
- `update_tag`: PUT `/reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id
  }}` - kind `update`; body type `json`; path fields `tag_id`; required record fields `tag_id`;
  accepted fields `description`, `entities`, `is_mandatory`, `multi_preference_entities`, `tag_id`,
  `tag_name`; risk: mutates Zoho Expense resources through update_tag; approval required.
- `delete_tag`: DELETE `/reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id
  }}` - kind `delete`; body type `none`; path fields `tag_id`; required record fields `tag_id`;
  accepted fields `tag_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Zoho Expense resources through delete_tag; approval required.
- `update_tag_options`: PUT `/reportingtags/{{ record.tag_id }}/options?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `tag_id`; required
  record fields `tag_id`; accepted fields `options`, `tag_id`; risk: mutates Zoho Expense resources
  through update_tag_options; approval required.
- `update_tag_criteria`: PUT `/reportingtags/{{ record.tag_id }}/criteria?organization_id={{
  config.organization_id }}` - kind `update`; body type `json`; path fields `tag_id`; required
  record fields `tag_id`; accepted fields `criteria_tags`, `options`, `tag_id`; risk: mutates Zoho
  Expense resources through update_tag_criteria; approval required.
- `active_tag`: POST `/reportingtags/{{ record.tag_id }}/active?organization_id={{
  config.organization_id }}` - kind `update`; body type `none`; path fields `tag_id`; required
  record fields `tag_id`; accepted fields `tag_id`; risk: mutates Zoho Expense resources through
  active_tag; approval required.
- `inactive_tag`: POST `/reportingtags/{{ record.tag_id }}/inactive?organization_id={{
  config.organization_id }}` - kind `update`; body type `none`; path fields `tag_id`; required
  record fields `tag_id`; accepted fields `tag_id`; risk: mutates Zoho Expense resources through
  inactive_tag; approval required.
- `active_tag_option`: POST `/reportingtags/{{ record.tag_id }}/option/{{ record.option_id
  }}/active?organization_id={{ config.organization_id }}` - kind `update`; body type `none`; path
  fields `tag_id`, `option_id`; required record fields `tag_id`, `option_id`; accepted fields
  `option_id`, `tag_id`; risk: mutates Zoho Expense resources through active_tag_option; approval
  required.
- `inactive_tag_option`: POST `/reportingtags/{{ record.tag_id }}/option/{{ record.option_id
  }}/inactive?organization_id={{ config.organization_id }}` - kind `update`; body type `none`; path
  fields `tag_id`, `option_id`; required record fields `tag_id`, `option_id`; accepted fields
  `option_id`, `tag_id`; risk: mutates Zoho Expense resources through inactive_tag_option; approval
  required.
- `reorder_tags`: PUT `/reportingtags/reorder?organization_id={{ config.organization_id }}` - kind
  `update`; body type `json`; accepted fields `tag_ids`; risk: mutates Zoho Expense resources
  through reorder_tags; approval required.
- `create_tax`: POST `/settings/taxes` - kind `create`; body type `json`; accepted fields
  `is_value_added`, `tax_authority_name`, `tax_name`, `tax_percentage`, `tax_type`; risk: creates
  Zoho Expense resources through create_tax; approval required.
- `update_tax`: PUT `/settings/taxes/{{ record.tax_id }}` - kind `update`; body type `json`; path
  fields `tax_id`; required record fields `tax_id`; accepted fields `is_value_added`, `tax_id`,
  `tax_name`, `tax_percentage`, `tax_type`; risk: mutates Zoho Expense resources through update_tax;
  approval required.
- `delete_tax`: DELETE `/settings/taxes/{{ record.tax_id }}` - kind `delete`; body type `none`; path
  fields `tax_id`; required record fields `tax_id`; accepted fields `tax_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes Zoho Expense
  resources through delete_tax; approval required.
- `create_trip`: POST `/trips` - kind `create`; body type `json`; accepted fields `budget`,
  `business_purpose`, `custom_fields`, `customer_id`, `destination_country_code`, `is_billable`,
  `is_visa_required`, `other_travellers`, `policy_id`, `project_id`, `travel_type`, `trip_name`;
  risk: creates Zoho Expense resources through create_trip; approval required.
- `update_trip`: PUT `/trips/{{ record.trip_id }}` - kind `update`; body type `json`; path fields
  `trip_id`; required record fields `trip_id`; accepted fields `budget`, `business_purpose`,
  `custom_fields`, `customer_id`, `destination_country_code`, `is_billable`, `is_visa_required`,
  `itineraries`, `other_travellers`, `policy_id`, `project_id`, `travel_type`, `trip_id`,
  `trip_name`; risk: mutates Zoho Expense resources through update_trip; approval required.
- `delete_trip`: DELETE `/trips/{{ record.trip_id }}` - kind `delete`; body type `none`; path fields
  `trip_id`; required record fields `trip_id`; accepted fields `trip_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes Zoho Expense resources through
  delete_trip; approval required.
- `approve_trip`: POST `/trips/{{ record.trip_id }}/approve` - kind `update`; body type `none`; path
  fields `trip_id`; required record fields `trip_id`; accepted fields `trip_id`; risk: mutates Zoho
  Expense resources through approve_trip; approval required.
- `reject_trip`: POST `/trips/{{ record.trip_id }}/reject` - kind `update`; body type `json`; path
  fields `trip_id`; required record fields `trip_id`; accepted fields `comments`, `trip_id`; risk:
  mutates Zoho Expense resources through reject_trip; approval required.
- `cancel_trip`: POST `/trips/{{ record.trip_id }}/cancel` - kind `update`; body type `json`; path
  fields `trip_id`; required record fields `trip_id`; accepted fields `comments`, `trip_id`; risk:
  mutates Zoho Expense resources through cancel_trip; approval required.
- `close_trip`: POST `/trips/{{ record.trip_id }}/close` - kind `update`; body type `none`; path
  fields `trip_id`; required record fields `trip_id`; accepted fields `trip_id`; risk: mutates Zoho
  Expense resources through close_trip; approval required.
- `create_user`: POST `/users` - kind `create`; body type `json`; accepted fields
  `approval_amount_limit`, `approves_to_email`, `custom_fields`, `date_of_birth`, `date_of_joining`,
  `default_approver_email`, `department_name`, `designation_name`, `email`, `employee_number`,
  `gender`, `mobile`, `name`, `policy_name`, `submission_amount_limit`, `user_role`; risk: creates
  Zoho Expense resources through create_user; approval required.
- `update_user`: PUT `/users/{{ record.user_id }}` - kind `update`; body type `json`; path fields
  `user_id`; required record fields `user_id`; accepted fields `approval_amount_limit`,
  `custom_fields`, `department_id`, `department_name`, `designation_name`, `mobile`, `name`,
  `submission_amount_limit`, `user_id`, `user_role`; risk: mutates Zoho Expense resources through
  update_user; approval required.
- `delete_user`: DELETE `/users/{{ record.user_id }}` - kind `delete`; body type `none`; path fields
  `user_id`; required record fields `user_id`; accepted fields `user_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes Zoho Expense resources through
  delete_user; approval required.
- `deactivate_user`: POST `/users/{{ record.user_id }}/inactive` - kind `update`; body type `none`;
  path fields `user_id`; required record fields `user_id`; accepted fields `user_id`; risk: mutates
  Zoho Expense resources through deactivate_user; approval required.
- `activate_user`: POST `/users/{{ record.user_id }}/active` - kind `update`; body type `none`; path
  fields `user_id`; required record fields `user_id`; accepted fields `user_id`; risk: mutates Zoho
  Expense resources through activate_user; approval required.
- `assign_role_to_user`: POST `/users/{{ record.user_id }}/role/{{ record.role_id }}` - kind
  `update`; body type `none`; path fields `user_id`, `role_id`; required record fields `user_id`,
  `role_id`; accepted fields `role_id`, `user_id`; risk: mutates Zoho Expense resources through
  assign_role_to_user; approval required.

## Known limits

- Batch defaults: read_page_size=200, write_batch_size=1.
- API coverage includes 29 stream-backed endpoint group(s), 88 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4.
