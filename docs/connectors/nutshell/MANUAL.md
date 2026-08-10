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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

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
    fields: accountTypeId(integer), createdTime(string), entityType(string), id(integer), industryId(integer), isHotLead(boolean), modifiedTime(string), name(string), url(string)
  contacts:
    primary key: id
    cursor: modifiedTime
    fields: createdTime(string), description(string), entityType(string), htmlUrl(string), id(integer), modifiedTime(string), name(string)
  leads:
    primary key: id
    cursor: modifiedTime
    fields: closedTime(string), confidence(integer), createdTime(string), entityType(string), id(integer), isOverdue(boolean), modifiedTime(string), name(string), status(integer), value(string)
  activities:
    primary key: id
    cursor: modifiedTime
    fields: activityTypeId(integer), createdTime(string), description(string), entityType(string), id(integer), isFlagged(boolean), logNote(string), modifiedTime(string), name(string), status(integer)
  users:
    primary key: id
    fields: createdTime(string), emails(string), entityType(string), id(integer), isAdministrator(boolean), isEnabled(boolean), modifiedTime(string), name(string)
  account:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  account_custom_fields:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  account_custom_field_attributes:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  account_list_items:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  account_list_fields:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  account_types:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  activity:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  activity_types:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  audiences:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  competitor_maps:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  competitor_map:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  competitors:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  competitor:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  contact:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  contact_custom_fields:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  contact_custom_field_attributes:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  contact_list_items:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  contact_list_fields:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  editions:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  edition:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  email:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  events:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  deleted_events:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  saved_filters:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  forms:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  form_field:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  form:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  industries:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  industry:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  invoices:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  invoice:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead_custom_fields:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead_installments:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead_stages:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead_custom_field_attributes:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead_list_items:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead_list_fields:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead_reports:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  markets:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  market:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  notes:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  note:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead_outcomes:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  lead_outcome:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  product_categories:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  product_category:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  product_maps:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  product_map:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  products:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  product:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  quotes:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  quote:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  sources:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  stages:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  pipelines:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  tags:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  tasks:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  task:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  territories:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)
  user:
    primary key: id
    fields: createdTime(string), deletedTime(string), href(string), htmlUrl(string), id(string), modifiedTime(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_account:
    endpoint: POST /accounts
    required fields: accounts
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
    required fields: name, type
    risk: creates an account custom field definition
  create_activity:
    endpoint: POST /activities
    required fields: activities
    risk: creates Nutshell activity records
  update_activity:
    endpoint: PUT /activities/{{ record.id }}
    required fields: id, activities
    risk: updates an existing Nutshell activity
  create_audience:
    endpoint: POST /audiences
    required fields: emAudiences
    risk: creates a Nutshell email marketing audience
  delete_competitor_map:
    endpoint: DELETE /competitormaps/{{ record.id }}
    required fields: id
    risk: deletes a lead-competitor relationship
  create_contact:
    endpoint: POST /contacts
    required fields: contacts
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
    required fields: name, type
    risk: creates a contact custom field definition
  create_lead:
    endpoint: POST /leads
    required fields: leads
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
    required fields: id, stageset
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
    required fields: name, type
    risk: creates a lead custom field definition
  create_note:
    endpoint: POST /notes
    required fields: data
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
    required fields: productCategories
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
    required fields: sources
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
    required fields: tags
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
    required fields: title
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

COMMAND SURFACE
  Run Nutshell's declared streams and reverse-ETL actions.
  Usage: pm nutshell <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    account custom field attributes list - Run the account custom field attributes ETL stream [intent=etl availability=implemented stream=account_custom_field_attributes]
    account custom fields list - Run the account custom fields ETL stream [intent=etl availability=implemented stream=account_custom_fields]
    account list - Run the account ETL stream [intent=etl availability=implemented stream=account]
    account list fields list - Run the account list fields ETL stream [intent=etl availability=implemented stream=account_list_fields]
    account list items list - Run the account list items ETL stream [intent=etl availability=implemented stream=account_list_items]
    account types list - Run the account types ETL stream [intent=etl availability=implemented stream=account_types]
    accounts list - Run the accounts ETL stream [intent=etl availability=implemented stream=accounts]
    activities list - Run the activities ETL stream [intent=etl availability=implemented stream=activities]
    activity list - Run the activity ETL stream [intent=etl availability=implemented stream=activity]
    activity types list - Run the activity types ETL stream [intent=etl availability=implemented stream=activity_types]
    api get stagesets id export - Documented GET /stagesets/{id}/export (not implemented) [intent=direct_read availability=not_implemented operation=nutshell.get.stagesets-id-export]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch accounts id - Documented PATCH /accounts/{id} (not implemented) [intent=direct_write availability=not_implemented operation=nutshell.patch.accounts-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch competitormaps id - Documented PATCH /competitormaps/{id} (not implemented) [intent=direct_write availability=not_implemented operation=nutshell.patch.competitormaps-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch contacts id - Documented PATCH /contacts/{id} (not implemented) [intent=direct_write availability=not_implemented operation=nutshell.patch.contacts-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch leads id - Documented PATCH /leads/{id} (not implemented) [intent=direct_write availability=not_implemented operation=nutshell.patch.leads-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch productmaps id - Documented PATCH /productmaps/{id} (not implemented) [intent=direct_write availability=not_implemented operation=nutshell.patch.productmaps-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch tasks id - Documented PATCH /tasks/{id} (not implemented) [intent=direct_write availability=not_implemented operation=nutshell.patch.tasks-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post leads id installments - Documented POST /leads/{id}/installments (not implemented) [intent=direct_write availability=not_implemented operation=nutshell.post.leads-id-installments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    audiences list - Run the audiences ETL stream [intent=etl availability=implemented stream=audiences]
    competitor list - Run the competitor ETL stream [intent=etl availability=implemented stream=competitor]
    competitor map list - Run the competitor map ETL stream [intent=etl availability=implemented stream=competitor_map]
    competitor maps list - Run the competitor maps ETL stream [intent=etl availability=implemented stream=competitor_maps]
    competitors list - Run the competitors ETL stream [intent=etl availability=implemented stream=competitors]
    contact custom field attributes list - Run the contact custom field attributes ETL stream [intent=etl availability=implemented stream=contact_custom_field_attributes]
    contact custom fields list - Run the contact custom fields ETL stream [intent=etl availability=implemented stream=contact_custom_fields]
    contact list - Run the contact ETL stream [intent=etl availability=implemented stream=contact]
    contact list fields list - Run the contact list fields ETL stream [intent=etl availability=implemented stream=contact_list_fields]
    contact list items list - Run the contact list items ETL stream [intent=etl availability=implemented stream=contact_list_items]
    contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
    create account apply - Plan and execute the create account reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_account]; approval: requires plan, preview, approval, and execute; risk: creates company/account records in Nutshell; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create account custom field apply - Plan and execute the create account custom field reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_account_custom_field]; approval: requires plan, preview, approval, and execute; risk: creates an account custom field definition; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create activity apply - Plan and execute the create activity reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_activity]; approval: requires plan, preview, approval, and execute; risk: creates Nutshell activity records; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create audience apply - Plan and execute the create audience reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_audience]; approval: requires plan, preview, approval, and execute; risk: creates a Nutshell email marketing audience; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact apply - Plan and execute the create contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact]; approval: requires plan, preview, approval, and execute; risk: creates person/contact records in Nutshell; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact custom field apply - Plan and execute the create contact custom field reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact_custom_field]; approval: requires plan, preview, approval, and execute; risk: creates a contact custom field definition; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create lead apply - Plan and execute the create lead reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_lead]; approval: requires plan, preview, approval, and execute; risk: creates Nutshell lead records; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create lead custom field apply - Plan and execute the create lead custom field reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_lead_custom_field]; approval: requires plan, preview, approval, and execute; risk: creates a lead custom field definition; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create note apply - Plan and execute the create note reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_note]; approval: requires plan, preview, approval, and execute; risk: creates a note attached to a Nutshell entity; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create product category apply - Plan and execute the create product category reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_product_category]; approval: requires plan, preview, approval, and execute; risk: creates a Nutshell product category; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create source apply - Plan and execute the create source reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_source]; approval: requires plan, preview, approval, and execute; risk: creates a lead source in Nutshell; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create tag apply - Plan and execute the create tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_tag]; approval: requires plan, preview, approval, and execute; risk: creates a Nutshell tag and optionally links entities; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create task apply - Plan and execute the create task reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_task]; approval: requires plan, preview, approval, and execute; risk: creates a task in Nutshell; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete account apply - Plan and execute the delete account reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_account]; approval: requires plan, preview, approval, and execute; risk: deletes a Nutshell account/company; recoverable only through the undelete endpoint for a limited period; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete competitor map apply - Plan and execute the delete competitor map reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_competitor_map]; approval: requires plan, preview, approval, and execute; risk: deletes a lead-competitor relationship; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete contact apply - Plan and execute the delete contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_contact]; approval: requires plan, preview, approval, and execute; risk: deletes a Nutshell contact/person; recoverable only through the undelete endpoint for a limited period; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete lead apply - Plan and execute the delete lead reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_lead]; approval: requires plan, preview, approval, and execute; risk: deletes a Nutshell lead; recoverable only through the undelete endpoint for a limited period; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete note apply - Plan and execute the delete note reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_note]; approval: requires plan, preview, approval, and execute; risk: deletes a Nutshell note; recoverable only through the undelete endpoint for a limited period; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete product apply - Plan and execute the delete product reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_product]; approval: requires plan, preview, approval, and execute; risk: deletes a Nutshell product; recoverable only through the undelete endpoint for a limited period; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete product map apply - Plan and execute the delete product map reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_product_map]; approval: requires plan, preview, approval, and execute; risk: deletes a product mapping from a lead; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete source apply - Plan and execute the delete source reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_source]; approval: requires plan, preview, approval, and execute; risk: deletes a lead source; recoverable only through the undelete endpoint for a limited period; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete tag apply - Plan and execute the delete tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_tag]; approval: requires plan, preview, approval, and execute; risk: deletes a Nutshell tag; recoverable only through the undelete endpoint for a limited period; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete task apply - Plan and execute the delete task reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_task]; approval: requires plan, preview, approval, and execute; risk: deletes a Nutshell task; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    deleted events list - Run the deleted events ETL stream [intent=etl availability=implemented stream=deleted_events]
    edition list - Run the edition ETL stream [intent=etl availability=implemented stream=edition]
    editions list - Run the editions ETL stream [intent=etl availability=implemented stream=editions]
    email list - Run the email ETL stream [intent=etl availability=implemented stream=email]
    events list - Run the events ETL stream [intent=etl availability=implemented stream=events]
    form field list - Run the form field ETL stream [intent=etl availability=implemented stream=form_field]
    form list - Run the form ETL stream [intent=etl availability=implemented stream=form]
    forms list - Run the forms ETL stream [intent=etl availability=implemented stream=forms]
    industries list - Run the industries ETL stream [intent=etl availability=implemented stream=industries]
    industry list - Run the industry ETL stream [intent=etl availability=implemented stream=industry]
    invoice list - Run the invoice ETL stream [intent=etl availability=implemented stream=invoice]
    invoices list - Run the invoices ETL stream [intent=etl availability=implemented stream=invoices]
    lead custom field attributes list - Run the lead custom field attributes ETL stream [intent=etl availability=implemented stream=lead_custom_field_attributes]
    lead custom fields list - Run the lead custom fields ETL stream [intent=etl availability=implemented stream=lead_custom_fields]
    lead installments list - Run the lead installments ETL stream [intent=etl availability=implemented stream=lead_installments]
    lead list - Run the lead ETL stream [intent=etl availability=implemented stream=lead]
    lead list fields list - Run the lead list fields ETL stream [intent=etl availability=implemented stream=lead_list_fields]
    lead list items list - Run the lead list items ETL stream [intent=etl availability=implemented stream=lead_list_items]
    lead outcome list - Run the lead outcome ETL stream [intent=etl availability=implemented stream=lead_outcome]
    lead outcomes list - Run the lead outcomes ETL stream [intent=etl availability=implemented stream=lead_outcomes]
    lead reports list - Run the lead reports ETL stream [intent=etl availability=implemented stream=lead_reports]
    lead stages list - Run the lead stages ETL stream [intent=etl availability=implemented stream=lead_stages]
    leads list - Run the leads ETL stream [intent=etl availability=implemented stream=leads]
    market list - Run the market ETL stream [intent=etl availability=implemented stream=market]
    markets list - Run the markets ETL stream [intent=etl availability=implemented stream=markets]
    note list - Run the note ETL stream [intent=etl availability=implemented stream=note]
    notes list - Run the notes ETL stream [intent=etl availability=implemented stream=notes]
    pipelines list - Run the pipelines ETL stream [intent=etl availability=implemented stream=pipelines]
    product categories list - Run the product categories ETL stream [intent=etl availability=implemented stream=product_categories]
    product category list - Run the product category ETL stream [intent=etl availability=implemented stream=product_category]
    product list - Run the product ETL stream [intent=etl availability=implemented stream=product]
    product map list - Run the product map ETL stream [intent=etl availability=implemented stream=product_map]
    product maps list - Run the product maps ETL stream [intent=etl availability=implemented stream=product_maps]
    products list - Run the products ETL stream [intent=etl availability=implemented stream=products]
    quote list - Run the quote ETL stream [intent=etl availability=implemented stream=quote]
    quotes list - Run the quotes ETL stream [intent=etl availability=implemented stream=quotes]
    reopen lead apply - Plan and execute the reopen lead reverse-ETL action [intent=reverse_etl availability=not_implemented write=reopen_lead]; approval: requires plan, preview, approval, and execute; risk: reopens a previously closed Nutshell lead; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    saved filters list - Run the saved filters ETL stream [intent=etl availability=implemented stream=saved_filters]
    set lead pipeline apply - Plan and execute the set lead pipeline reverse-ETL action [intent=reverse_etl availability=not_implemented write=set_lead_pipeline]; approval: requires plan, preview, approval, and execute; risk: changes the pipeline/stageset assigned to a lead; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    sources list - Run the sources ETL stream [intent=etl availability=implemented stream=sources]
    stages list - Run the stages ETL stream [intent=etl availability=implemented stream=stages]
    tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]
    task list - Run the task ETL stream [intent=etl availability=implemented stream=task]
    tasks list - Run the tasks ETL stream [intent=etl availability=implemented stream=tasks]
    territories list - Run the territories ETL stream [intent=etl availability=implemented stream=territories]
    undelete account apply - Plan and execute the undelete account reverse-ETL action [intent=reverse_etl availability=not_implemented write=undelete_account]; approval: requires plan, preview, approval, and execute; risk: restores a deleted Nutshell account/company; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    undelete contact apply - Plan and execute the undelete contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=undelete_contact]; approval: requires plan, preview, approval, and execute; risk: restores a deleted Nutshell contact/person; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    undelete lead apply - Plan and execute the undelete lead reverse-ETL action [intent=reverse_etl availability=not_implemented write=undelete_lead]; approval: requires plan, preview, approval, and execute; risk: restores a deleted Nutshell lead; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    undelete note apply - Plan and execute the undelete note reverse-ETL action [intent=reverse_etl availability=not_implemented write=undelete_note]; approval: requires plan, preview, approval, and execute; risk: restores a deleted Nutshell note; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    undelete product apply - Plan and execute the undelete product reverse-ETL action [intent=reverse_etl availability=not_implemented write=undelete_product]; approval: requires plan, preview, approval, and execute; risk: restores a deleted Nutshell product; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    undelete source apply - Plan and execute the undelete source reverse-ETL action [intent=reverse_etl availability=not_implemented write=undelete_source]; approval: requires plan, preview, approval, and execute; risk: restores a deleted lead source; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    undelete tag apply - Plan and execute the undelete tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=undelete_tag]; approval: requires plan, preview, approval, and execute; risk: restores a deleted Nutshell tag; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update activity apply - Plan and execute the update activity reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_activity]; approval: requires plan, preview, approval, and execute; risk: updates an existing Nutshell activity; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update lead status apply - Plan and execute the update lead status reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_lead_status]; approval: requires plan, preview, approval, and execute; risk: updates a lead status/outcome and optional competitor/product maps; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    user list - Run the user ETL stream [intent=etl availability=implemented stream=user]
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
    watch lead apply - Plan and execute the watch lead reverse-ETL action [intent=reverse_etl availability=not_implemented write=watch_lead]; approval: requires plan, preview, approval, and execute; risk: toggles watch notifications for the authenticated user on a lead; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags

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
