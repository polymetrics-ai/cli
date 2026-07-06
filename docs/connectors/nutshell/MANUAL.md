# pm connectors inspect nutshell

```text
NAME
  pm connectors inspect nutshell - Nutshell connector manual

SYNOPSIS
  pm connectors inspect nutshell
  pm connectors inspect nutshell --json
  pm credentials add <name> --connector nutshell [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes documented Nutshell CRM REST resources through the Nutshell REST API.

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
  account_id
  activity_id
  base_url
  competitor_id
  competitor_map_id
  contact_id
  edition_id
  email_id
  form_field_id
  form_id
  industry_id
  invoice_id
  lead_id
  market_id
  mode
  note_id
  outcome_id
  page_size
  product_category_id
  product_id
  product_map_id
  quote_id
  task_id
  user_id
  username
  password (secret)

ETL STREAMS
  accounts:
    primary key: id
    cursor: modifiedTime
    fields: accountTypeId(), createdTime(), entityType(), id(), industryId(), isHotLead(), modifiedTime(), name(), url()
  contacts:
    primary key: id
    cursor: modifiedTime
    fields: createdTime(), description(), entityType(), htmlUrl(), id(), modifiedTime(), name()
  leads:
    primary key: id
    cursor: modifiedTime
    fields: closedTime(), confidence(), createdTime(), entityType(), id(), isOverdue(), modifiedTime(), name(), status(), value()
  activities:
    primary key: id
    cursor: modifiedTime
    fields: activityTypeId(), createdTime(), description(), entityType(), id(), isFlagged(), logNote(), modifiedTime(), name(), status()
  users:
    primary key: id
    fields: createdTime(), emails(), entityType(), id(), isAdministrator(), isEnabled(), modifiedTime(), name()
  account:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  account_custom_fields:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  account_custom_field_attributes:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  account_list_items:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  account_list_fields:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  account_types:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  activity:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  activity_types:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  audiences:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  competitor_maps:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  competitor_map:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  competitors:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  competitor:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  contact:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  contact_custom_fields:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  contact_custom_field_attributes:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  contact_list_items:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  contact_list_fields:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  editions:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  edition:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  email:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  events:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  deleted_events:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  saved_filters:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  forms:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  form_field:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  form:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  industries:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  industry:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  invoices:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  invoice:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead_custom_fields:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead_installments:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead_stages:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead_custom_field_attributes:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead_list_items:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead_list_fields:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead_reports:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  markets:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  market:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  notes:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  note:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead_outcomes:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  lead_outcome:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  product_categories:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  product_category:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  product_maps:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  product_map:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  products:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  product:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  quotes:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  quote:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  sources:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  stages:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  pipelines:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  tags:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  tasks:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  task:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  territories:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()
  user:
    primary key: id
    fields: createdTime(), deletedTime(), href(), htmlUrl(), id(), modifiedTime(), name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_account:
    endpoint: POST /accounts
    risk: creates company/account records in Nutshell
  delete_account:
    endpoint: DELETE /accounts/{{ record.id }}
    required fields: id
    risk: deletes a Nutshell account/company; recoverable only through the undelete endpoint for a limited period
  undelete_account:
    endpoint: POST /accounts/{{ record.id }}/undelete
    required fields: id
    risk: restores a deleted Nutshell account/company
  create_account_custom_field:
    endpoint: POST /accounts/customfield
    risk: creates an account custom field definition
  create_activity:
    endpoint: POST /activities
    risk: creates Nutshell activity records
  update_activity:
    endpoint: PUT /activities/{{ record.id }}
    required fields: id
    risk: updates an existing Nutshell activity
  create_audience:
    endpoint: POST /audiences
    risk: creates a Nutshell email marketing audience
  delete_competitor_map:
    endpoint: DELETE /competitormaps/{{ record.id }}
    required fields: id
    risk: deletes a lead-competitor relationship
  create_contact:
    endpoint: POST /contacts
    risk: creates person/contact records in Nutshell
  delete_contact:
    endpoint: DELETE /contacts/{{ record.id }}
    required fields: id
    risk: deletes a Nutshell contact/person; recoverable only through the undelete endpoint for a limited period
  undelete_contact:
    endpoint: POST /contacts/{{ record.id }}/undelete
    required fields: id
    risk: restores a deleted Nutshell contact/person
  create_contact_custom_field:
    endpoint: POST /contacts/customfield
    risk: creates a contact custom field definition
  create_lead:
    endpoint: POST /leads
    risk: creates Nutshell lead records
  delete_lead:
    endpoint: DELETE /leads/{{ record.id }}
    required fields: id
    risk: deletes a Nutshell lead; recoverable only through the undelete endpoint for a limited period
  reopen_lead:
    endpoint: POST /leads/{{ record.id }}/reopen
    required fields: id
    risk: reopens a previously closed Nutshell lead
  set_lead_pipeline:
    endpoint: POST /leads/{{ record.id }}/stageset
    required fields: id
    risk: changes the pipeline/stageset assigned to a lead
  update_lead_status:
    endpoint: POST /leads/{{ record.id }}/status
    required fields: id
    risk: updates a lead status/outcome and optional competitor/product maps
  undelete_lead:
    endpoint: POST /leads/{{ record.id }}/undelete
    required fields: id
    risk: restores a deleted Nutshell lead
  watch_lead:
    endpoint: POST /leads/{{ record.id }}/watch
    required fields: id
    risk: toggles watch notifications for the authenticated user on a lead
  create_lead_custom_field:
    endpoint: POST /leads/customfield
    risk: creates a lead custom field definition
  create_note:
    endpoint: POST /notes
    risk: creates a note attached to a Nutshell entity
  delete_note:
    endpoint: DELETE /notes/{{ record.id }}
    required fields: id
    risk: deletes a Nutshell note; recoverable only through the undelete endpoint for a limited period
  undelete_note:
    endpoint: POST /notes/{{ record.id }}/undelete
    required fields: id
    risk: restores a deleted Nutshell note
  create_product_category:
    endpoint: POST /productcategories
    risk: creates a Nutshell product category
  delete_product_map:
    endpoint: DELETE /productMaps/{{ record.id }}
    required fields: id
    risk: deletes a product mapping from a lead
  delete_product:
    endpoint: DELETE /products/{{ record.id }}
    required fields: id
    risk: deletes a Nutshell product; recoverable only through the undelete endpoint for a limited period
  undelete_product:
    endpoint: POST /products/{{ record.id }}/undelete
    required fields: id
    risk: restores a deleted Nutshell product
  create_source:
    endpoint: POST /sources
    risk: creates a lead source in Nutshell
  delete_source:
    endpoint: DELETE /sources/{{ record.id }}
    required fields: id
    risk: deletes a lead source; recoverable only through the undelete endpoint for a limited period
  undelete_source:
    endpoint: POST /sources/{{ record.id }}/undelete
    required fields: id
    risk: restores a deleted lead source
  create_tag:
    endpoint: POST /tags
    risk: creates a Nutshell tag and optionally links entities
  delete_tag:
    endpoint: DELETE /tags/{{ record.id }}
    required fields: id
    risk: deletes a Nutshell tag; recoverable only through the undelete endpoint for a limited period
  undelete_tag:
    endpoint: POST /tags/{{ record.id }}/undelete
    required fields: id
    risk: restores a deleted Nutshell tag
  create_task:
    endpoint: POST /tasks
    risk: creates a task in Nutshell
  delete_task:
    endpoint: DELETE /tasks/{{ record.id }}
    required fields: id
    risk: deletes a Nutshell task

SECURITY
  read risk: external Nutshell CRM read of account/contact/lead/activity/user and related CRM data
  write risk: external Nutshell CRM mutations including creates, updates, undeletes, watches, and destructive deletes
  approval: required for write actions; destructive delete actions carry destructive confirmation metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect nutshell

  # Inspect as structured JSON
  pm connectors inspect nutshell --json

AGENT WORKFLOW
  - Run pm connectors inspect nutshell before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
