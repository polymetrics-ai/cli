# pm connectors inspect zoho-expense

```text
NAME
  pm connectors inspect zoho-expense - Zoho Expense connector manual

SYNOPSIS
  pm connectors inspect zoho-expense
  pm connectors inspect zoho-expense --json
  pm credentials add <name> --connector zoho-expense [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Zoho Expense API v1 resources through the declarative connector engine.

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
  advance_payment_id
  base_url
  currency_id
  customer_id
  expense_category_id
  expense_report_id
  mode
  organization_id
  project_id
  tag_id
  tax_group_id
  tax_id
  trip_id
  user_id
  access_token (secret)

ETL STREAMS
  reports:
    primary key: report_id
    cursor: modified_time
    fields: id(), name(), report_id(), updated_at()
  expenses:
    primary key: expense_id
    cursor: modified_time
    fields: expense_id(), id(), name(), updated_at()
  users:
    primary key: user_id
    cursor: modified_time
    fields: id(), name(), updated_at(), user_id()
  currencies:
    primary key: currency_id
    fields: currency_id()
  currency:
    primary key: currency_id
    fields: currency_id()
  customers:
    primary key: customer_id
    fields: customer_id()
  customer:
    primary key: customer_id
    fields: customer_id()
  advance_payments:
    primary key: advance_payment_id
    fields: advance_payment_id()
  advance_payment:
    primary key: advance_payment_id
    fields: advance_payment_id()
  expense_categories:
    primary key: expense_category_id
    fields: expense_category_id()
  expense_category:
    primary key: expense_category_id
    fields: expense_category_id()
  expense_report:
    primary key: expense_report_id
    fields: expense_report_id()
  expense_report_approval_history:
    primary key: approval_id
    fields: approval_id()
  expense_report_validation:
    primary key: expense_report_id
    fields: expense_report_id()
  expense_report_budget_summary:
    primary key: expense_report_id
    fields: expense_report_id()
  expense_report_reimbursement:
    primary key: expense_report_id
    fields: expense_report_id()
  organizations:
    primary key: organization_id
    fields: organization_id()
  organization:
    primary key: organization_id
    fields: organization_id()
  projects:
    primary key: project_id
    fields: project_id()
  project:
    primary key: project_id
    fields: project_id()
  reporting_tags:
    primary key: tag_id
    fields: tag_id()
  tag_options:
    primary key: option_id
    fields: option_id()
  reporting_tag_options_all:
    primary key: option_id
    fields: option_id()
  taxes:
    primary key: tax_id
    fields: tax_id()
  tax:
    primary key: tax_id
    fields: tax_id()
  tax_group:
    primary key: tax_group_id
    fields: tax_group_id()
  trips:
    primary key: trip_id
    fields: trip_id()
  trip:
    primary key: trip_id
    fields: trip_id()
  user:
    primary key: user_id
    fields: user_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_currency:
    endpoint: POST /settings/currencies
    risk: creates Zoho Expense resources through create_currency; approval required
  update_currency:
    endpoint: PUT /settings/currencies/{{ record.currency_id }}
    required fields: currency_id
    risk: mutates Zoho Expense resources through update_currency; approval required
  delete_currency:
    endpoint: DELETE /settings/currencies/{{ record.currency_id }}
    required fields: currency_id
    risk: deletes Zoho Expense resources through delete_currency; approval required
  create_customer:
    endpoint: POST /contacts
    risk: creates Zoho Expense resources through create_customer; approval required
  update_customer:
    endpoint: PUT /contacts/{{ record.customer_id }}
    required fields: customer_id
    risk: mutates Zoho Expense resources through update_customer; approval required
  delete_customer:
    endpoint: DELETE /contacts/{{ record.customer_id }}
    required fields: customer_id
    risk: deletes Zoho Expense resources through delete_customer; approval required
  create_advance_payment:
    endpoint: POST /advancepayments
    risk: creates Zoho Expense resources through create_advance_payment; approval required
  update_advance_payment:
    endpoint: PUT /advancepayments/{{ record.advance_payment_id }}
    required fields: advance_payment_id
    risk: mutates Zoho Expense resources through update_advance_payment; approval required
  delete_advance_payment:
    endpoint: DELETE /advancepayments/{{ record.advance_payment_id }}
    required fields: advance_payment_id
    risk: deletes Zoho Expense resources through delete_advance_payment; approval required
  create_expense_category:
    endpoint: POST /expensecategories
    risk: creates Zoho Expense resources through create_expense_category; approval required
  update_expense_category:
    endpoint: PUT /expensecategories/{{ record.expense_category_id }}
    required fields: expense_category_id
    risk: mutates Zoho Expense resources through update_expense_category; approval required
  delete_expense_category:
    endpoint: DELETE /expensecategories/{{ record.expense_category_id }}
    required fields: expense_category_id
    risk: deletes Zoho Expense resources through delete_expense_category; approval required
  enable_expense_category:
    endpoint: POST /expensecategories/{{ record.expense_category_id }}/show
    required fields: expense_category_id
    risk: mutates Zoho Expense resources through enable_expense_category; approval required
  disable_expense_category:
    endpoint: POST /expensecategories/{{ record.expense_category_id }}/hide
    required fields: expense_category_id
    risk: mutates Zoho Expense resources through disable_expense_category; approval required
  create_expense_report:
    endpoint: POST /expensereports
    risk: creates Zoho Expense resources through create_expense_report; approval required
  bulk_delete_expense_reports:
    endpoint: DELETE /expensereports
    risk: deletes Zoho Expense resources through bulk_delete_expense_reports; approval required
  update_expense_report:
    endpoint: PUT /expensereports/{{ record.expense_report_id }}
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through update_expense_report; approval required
  delete_expense_report:
    endpoint: DELETE /expensereports/{{ record.expense_report_id }}
    required fields: expense_report_id
    risk: deletes Zoho Expense resources through delete_expense_report; approval required
  submit_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/submit
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through submit_expense_report; approval required
  approve_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/approve
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through approve_expense_report; approval required
  reject_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/reject
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through reject_expense_report; approval required
  reimburse_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/reimburse
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through reimburse_expense_report; approval required
  takeback_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/takeback
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through takeback_expense_report; approval required
  archive_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/archive
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through archive_expense_report; approval required
  unarchive_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/unarchive
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through unarchive_expense_report; approval required
  share_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/share
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through share_expense_report; approval required
  forward_approval_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/forwardapproval
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through forward_approval_expense_report; approval required
  reject_expense_in_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/expense/{{ record.expense_id }}/reject
    required fields: expense_report_id, expense_id
    risk: mutates Zoho Expense resources through reject_expense_in_report; approval required
  bulk_reject_expenses_in_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/expensebulkreject
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through bulk_reject_expenses_in_report; approval required
  add_comment_to_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/comments
    required fields: expense_report_id
    risk: creates Zoho Expense resources through add_comment_to_expense_report; approval required
  delete_comment_of_expense_report:
    endpoint: DELETE /expensereports/{{ record.expense_report_id }}/comments/{{ record.comment_id }}
    required fields: expense_report_id, comment_id
    risk: deletes Zoho Expense resources through delete_comment_of_expense_report; approval required
  associate_tags_to_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/tags
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through associate_tags_to_expense_report; approval required
  add_expenses_to_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/addexpenses
    required fields: expense_report_id
    risk: creates Zoho Expense resources through add_expenses_to_expense_report; approval required
  remove_expenses_from_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/removeexpenses
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through remove_expenses_from_expense_report; approval required
  delete_expense_report_attachment:
    endpoint: DELETE /expensereports/{{ record.expense_report_id }}/documents/{{ record.document_id }}
    required fields: expense_report_id, document_id
    risk: deletes Zoho Expense resources through delete_expense_report_attachment; approval required
  reupload_expense_report_attachment:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/documents/{{ record.document_id }}/upload
    required fields: expense_report_id, document_id
    risk: mutates Zoho Expense resources through reupload_expense_report_attachment; approval required
  remove_advance_payment_from_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/removeadvancepayment
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through remove_advance_payment_from_expense_report; approval required
  change_status_of_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/status/{{ record.action }}
    required fields: expense_report_id, action
    risk: mutates Zoho Expense resources through change_status_of_expense_report; approval required
  export_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/export
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through export_expense_report; approval required
  cancel_reimbursement_of_expense_report:
    endpoint: POST /expensereports/{{ record.expense_report_id }}/reimburse/cancel
    required fields: expense_report_id
    risk: mutates Zoho Expense resources through cancel_reimbursement_of_expense_report; approval required
  add_expense_to_report:
    endpoint: POST /expenses/{{ record.expense_id }}/addtoreport
    required fields: expense_id
    risk: creates Zoho Expense resources through add_expense_to_report; approval required
  bulk_submit_expense_reports:
    endpoint: POST /expensereports/submit
    risk: mutates Zoho Expense resources through bulk_submit_expense_reports; approval required
  bulk_approve_expense_reports:
    endpoint: POST /expensereports/approve
    risk: mutates Zoho Expense resources through bulk_approve_expense_reports; approval required
  bulk_reject_expense_reports:
    endpoint: POST /expensereports/reject
    risk: mutates Zoho Expense resources through bulk_reject_expense_reports; approval required
  bulk_reimburse_expense_reports:
    endpoint: POST /expensereports/reimburse
    risk: mutates Zoho Expense resources through bulk_reimburse_expense_reports; approval required
  bulk_archive_expense_reports:
    endpoint: POST /expensereports/archive
    risk: mutates Zoho Expense resources through bulk_archive_expense_reports; approval required
  bulk_unarchive_expense_reports:
    endpoint: POST /expensereports/unarchive
    risk: mutates Zoho Expense resources through bulk_unarchive_expense_reports; approval required
  bulk_forward_approval_expense_reports:
    endpoint: POST /expensereports/forwardapproval
    risk: mutates Zoho Expense resources through bulk_forward_approval_expense_reports; approval required
  bulk_change_status_of_expense_reports:
    endpoint: POST /expensereports/status/{{ record.action }}
    required fields: action
    risk: mutates Zoho Expense resources through bulk_change_status_of_expense_reports; approval required
  bulk_reset_substatus_of_expense_reports:
    endpoint: POST /expensereports/substatus/reset
    risk: mutates Zoho Expense resources through bulk_reset_substatus_of_expense_reports; approval required
  bulk_field_update_expense_reports:
    endpoint: POST /expensereports/bulkfieldupdate
    risk: mutates Zoho Expense resources through bulk_field_update_expense_reports; approval required
  create_expense:
    endpoint: POST /expenses
    risk: creates Zoho Expense resources through create_expense; approval required
  update_expense:
    endpoint: PUT /expenses/{{ record.expense_id }}
    required fields: expense_id
    risk: mutates Zoho Expense resources through update_expense; approval required
  merge_expenses:
    endpoint: POST /expenses/{{ record.expense_id }}/merge?duplicate_expense_id={{ record.duplicate_expense_id }}
    required fields: expense_id, duplicate_expense_id
    risk: mutates Zoho Expense resources through merge_expenses; approval required
  create_organization:
    endpoint: POST /organizations
    risk: creates Zoho Expense resources through create_organization; approval required
  update_organization:
    endpoint: PUT /organizations/{{ record.organization_id }}
    required fields: organization_id
    risk: mutates Zoho Expense resources through update_organization; approval required
  create_project:
    endpoint: POST /projects
    risk: creates Zoho Expense resources through create_project; approval required
  update_project:
    endpoint: PUT /projects/{{ record.project_id }}
    required fields: project_id
    risk: mutates Zoho Expense resources through update_project; approval required
  delete_project:
    endpoint: DELETE /projects/{{ record.project_id }}
    required fields: project_id
    risk: deletes Zoho Expense resources through delete_project; approval required
  activate_project:
    endpoint: POST /projects/{{ record.project_id }}/active
    required fields: project_id
    risk: mutates Zoho Expense resources through activate_project; approval required
  deactivate_project:
    endpoint: POST /projects/{{ record.project_id }}/inactive
    required fields: project_id
    risk: mutates Zoho Expense resources through deactivate_project; approval required
  create_tag:
    endpoint: POST /reportingtags?organization_id={{ config.organization_id }}
    risk: creates Zoho Expense resources through create_tag; approval required
  mark_default_option:
    endpoint: POST /reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id }}&default_option_id={{ record.default_option_id }}
    required fields: tag_id, default_option_id
    risk: mutates Zoho Expense resources through mark_default_option; approval required
  update_tag:
    endpoint: PUT /reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: mutates Zoho Expense resources through update_tag; approval required
  delete_tag:
    endpoint: DELETE /reportingtags/{{ record.tag_id }}?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: deletes Zoho Expense resources through delete_tag; approval required
  update_tag_options:
    endpoint: PUT /reportingtags/{{ record.tag_id }}/options?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: mutates Zoho Expense resources through update_tag_options; approval required
  update_tag_criteria:
    endpoint: PUT /reportingtags/{{ record.tag_id }}/criteria?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: mutates Zoho Expense resources through update_tag_criteria; approval required
  active_tag:
    endpoint: POST /reportingtags/{{ record.tag_id }}/active?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: mutates Zoho Expense resources through active_tag; approval required
  inactive_tag:
    endpoint: POST /reportingtags/{{ record.tag_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: tag_id
    risk: mutates Zoho Expense resources through inactive_tag; approval required
  active_tag_option:
    endpoint: POST /reportingtags/{{ record.tag_id }}/option/{{ record.option_id }}/active?organization_id={{ config.organization_id }}
    required fields: tag_id, option_id
    risk: mutates Zoho Expense resources through active_tag_option; approval required
  inactive_tag_option:
    endpoint: POST /reportingtags/{{ record.tag_id }}/option/{{ record.option_id }}/inactive?organization_id={{ config.organization_id }}
    required fields: tag_id, option_id
    risk: mutates Zoho Expense resources through inactive_tag_option; approval required
  reorder_tags:
    endpoint: PUT /reportingtags/reorder?organization_id={{ config.organization_id }}
    risk: mutates Zoho Expense resources through reorder_tags; approval required
  create_tax:
    endpoint: POST /settings/taxes
    risk: creates Zoho Expense resources through create_tax; approval required
  update_tax:
    endpoint: PUT /settings/taxes/{{ record.tax_id }}
    required fields: tax_id
    risk: mutates Zoho Expense resources through update_tax; approval required
  delete_tax:
    endpoint: DELETE /settings/taxes/{{ record.tax_id }}
    required fields: tax_id
    risk: deletes Zoho Expense resources through delete_tax; approval required
  create_trip:
    endpoint: POST /trips
    risk: creates Zoho Expense resources through create_trip; approval required
  update_trip:
    endpoint: PUT /trips/{{ record.trip_id }}
    required fields: trip_id
    risk: mutates Zoho Expense resources through update_trip; approval required
  delete_trip:
    endpoint: DELETE /trips/{{ record.trip_id }}
    required fields: trip_id
    risk: deletes Zoho Expense resources through delete_trip; approval required
  approve_trip:
    endpoint: POST /trips/{{ record.trip_id }}/approve
    required fields: trip_id
    risk: mutates Zoho Expense resources through approve_trip; approval required
  reject_trip:
    endpoint: POST /trips/{{ record.trip_id }}/reject
    required fields: trip_id
    risk: mutates Zoho Expense resources through reject_trip; approval required
  cancel_trip:
    endpoint: POST /trips/{{ record.trip_id }}/cancel
    required fields: trip_id
    risk: mutates Zoho Expense resources through cancel_trip; approval required
  close_trip:
    endpoint: POST /trips/{{ record.trip_id }}/close
    required fields: trip_id
    risk: mutates Zoho Expense resources through close_trip; approval required
  create_user:
    endpoint: POST /users
    risk: creates Zoho Expense resources through create_user; approval required
  update_user:
    endpoint: PUT /users/{{ record.user_id }}
    required fields: user_id
    risk: mutates Zoho Expense resources through update_user; approval required
  delete_user:
    endpoint: DELETE /users/{{ record.user_id }}
    required fields: user_id
    risk: deletes Zoho Expense resources through delete_user; approval required
  deactivate_user:
    endpoint: POST /users/{{ record.user_id }}/inactive
    required fields: user_id
    risk: mutates Zoho Expense resources through deactivate_user; approval required
  activate_user:
    endpoint: POST /users/{{ record.user_id }}/active
    required fields: user_id
    risk: mutates Zoho Expense resources through activate_user; approval required
  assign_role_to_user:
    endpoint: POST /users/{{ record.user_id }}/role/{{ record.role_id }}
    required fields: user_id, role_id
    risk: mutates Zoho Expense resources through assign_role_to_user; approval required

SECURITY
  read risk: external Zoho Expense API read of expense, report, organization, user, project, tax, currency, customer, trip, advance, and reporting tag data
  write risk: creates, updates, submits, approves, reimburses, archives, deletes, and otherwise mutates Zoho Expense resources; requires explicit approval
  approval: writes require explicit operator approval through the reverse ETL approval flow
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zoho-expense

  # Inspect as structured JSON
  pm connectors inspect zoho-expense --json

AGENT WORKFLOW
  - Run pm connectors inspect zoho-expense before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
