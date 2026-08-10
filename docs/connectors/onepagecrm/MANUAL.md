# pm connectors inspect onepagecrm

```text
NAME
  pm connectors inspect onepagecrm - OnePageCRM connector manual

SYNOPSIS
  pm connectors inspect onepagecrm
  pm connectors inspect onepagecrm --json
  pm credentials add <name> --connector onepagecrm [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads the documented OnePageCRM API v3 CRM surface and exposes declarative write actions for supported JSON/path mutations.

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
  action_id
  attachment_id
  base_url
  call_id
  company_field_id
  company_id
  contact_id
  custom_field_id
  deal_field_id
  deal_id
  filter_id
  last_id
  lead_source_id
  meeting_id
  mode
  note_id
  notification_id
  owner_id
  pipeline_id
  predefined_action_group_id
  predefined_action_id
  predefined_item_group_id
  predefined_item_id
  relationship_id
  relationship_type_id
  status_id
  tag_name
  user_id
  username
  webhook_id
  password (secret)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: company_name(string), created_at(string), first_name(string), id(string), job_title(string), last_name(string), owner_id(string), starred(boolean), status_id(string), updated_at(string)
  deals:
    primary key: id
    cursor: updated_at
    fields: amount(number), contact_id(string), created_at(string), currency(string), expected_close_date(string), id(string), name(string), owner_id(string), stage(string), status(string), updated_at(string)
  actions:
    primary key: id
    cursor: updated_at
    fields: assignee_id(string), contact_id(string), created_at(string), date(string), done(boolean), id(string), status(string), text(string), updated_at(string)
  companies:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), phone(string), updated_at(string), url(string)
  users:
    primary key: id
    fields: email(string), first_name(string), id(string), last_name(string), role(string), status(string)
  bootstrap:
    primary key: id
    fields: id(string)
  user:
    primary key: id
    fields: id(string)
  lead_sources:
    primary key: id
    fields: id(string)
  lead_source:
    primary key: id
    fields: id(string)
  statuses:
    primary key: id
    fields: id(string)
  status:
    primary key: id
    fields: id(string)
  deal_fields:
    primary key: id
    fields: id(string)
  deal_field:
    primary key: id
    fields: id(string)
  custom_fields:
    primary key: id
    fields: id(string)
  custom_field:
    primary key: id
    fields: id(string)
  company_fields:
    primary key: id
    fields: id(string)
  company_field:
    primary key: id
    fields: id(string)
  predefined_actions:
    primary key: id
    fields: id(string)
  predefined_action:
    primary key: id
    fields: id(string)
  predefined_action_groups:
    primary key: id
    fields: id(string)
  predefined_action_group:
    primary key: id
    fields: id(string)
  predefined_items:
    primary key: id
    fields: id(string)
  predefined_item:
    primary key: id
    fields: id(string)
  predefined_item_groups:
    primary key: id
    fields: id(string)
  predefined_item_group:
    primary key: id
    fields: id(string)
  notes:
    primary key: id
    fields: id(string)
  note:
    primary key: id
    fields: id(string)
  calls:
    primary key: id
    fields: id(string)
  call:
    primary key: id
    fields: id(string)
  call_results:
    primary key: id
    fields: id(string)
  meetings:
    primary key: id
    fields: id(string)
  meeting:
    primary key: id
    fields: id(string)
  deal:
    primary key: id
    fields: id(string)
  relationship_types:
    primary key: id
    fields: id(string)
  relationship_type:
    primary key: id
    fields: id(string)
  countries:
    primary key: id
    fields: id(string)
  action:
    primary key: id
    fields: id(string)
  filters:
    primary key: id
    fields: id(string)
  filter:
    primary key: id
    fields: id(string)
  company:
    primary key: id
    fields: id(string)
  company_actions:
    primary key: id
    fields: id(string)
  company_deals:
    primary key: id
    fields: id(string)
  company_notes:
    primary key: id
    fields: id(string)
  company_calls:
    primary key: id
    fields: id(string)
  company_meetings:
    primary key: id
    fields: id(string)
  company_linked_contacts:
    primary key: id
    fields: id(string)
  company_pinned_attachments:
    primary key: id
    fields: id(string)
  contact:
    primary key: id
    fields: id(string)
  filtered_contacts:
    primary key: id
    fields: id(string)
  contact_actions:
    primary key: id
    fields: id(string)
  contact_deals:
    primary key: id
    fields: id(string)
  contact_notes:
    primary key: id
    fields: id(string)
  contact_calls:
    primary key: id
    fields: id(string)
  contact_meetings:
    primary key: id
    fields: id(string)
  contact_relationships:
    primary key: id
    fields: id(string)
  contact_relationship:
    primary key: id
    fields: id(string)
  contact_pinned_attachments:
    primary key: id
    fields: id(string)
  contacts_cascade:
    primary key: id
    fields: id(string)
  contacts_cascade_after:
    primary key: id
    fields: id(string)
  action_stream:
    primary key: id
    fields: id(string)
  team_stream:
    primary key: id
    fields: id(string)
  notifications:
    primary key: id
    fields: id(string)
  notification:
    primary key: id
    fields: id(string)
  webhooks:
    primary key: id
    fields: id(string)
  webhook:
    primary key: id
    fields: id(string)
  pipelines:
    primary key: id
    fields: id(string)
  pipeline:
    primary key: id
    fields: id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  update_user:
    endpoint: PUT /users/{{ record.user_id }}
    required fields: user_id, first_name
    risk: Update a specific user; external OnePageCRM mutation, approval required
  create_lead_source:
    endpoint: POST /lead_sources
    required fields: id
    risk: Create a new lead source; external OnePageCRM mutation, approval required
  update_lead_source:
    endpoint: PUT /lead_sources/{{ record.lead_source_id }}
    required fields: lead_source_id, id
    risk: Update a specific lead source; external OnePageCRM mutation, approval required
  delete_lead_source:
    endpoint: DELETE /lead_sources/{{ record.lead_source_id }}
    required fields: lead_source_id
    risk: Delete a specific lead source; external OnePageCRM mutation, approval required
  create_status:
    endpoint: POST /statuses
    required fields: id
    risk: Create a new status; external OnePageCRM mutation, approval required
  update_status:
    endpoint: PUT /statuses/{{ record.status_id }}
    required fields: status_id, id
    risk: Update a specific status; external OnePageCRM mutation, approval required
  delete_status:
    endpoint: DELETE /statuses/{{ record.status_id }}
    required fields: status_id
    risk: Delete a specific status; external OnePageCRM mutation, approval required
  create_deal_field:
    endpoint: POST /deal_fields
    required fields: id
    risk: Create a new deal field; external OnePageCRM mutation, approval required
  update_deal_field:
    endpoint: PUT /deal_fields/{{ record.deal_field_id }}
    required fields: deal_field_id, id
    risk: Update a specific deal field; external OnePageCRM mutation, approval required
  delete_deal_field:
    endpoint: DELETE /deal_fields/{{ record.deal_field_id }}
    required fields: deal_field_id
    risk: Delete a specific deal field; external OnePageCRM mutation, approval required
  create_custom_field:
    endpoint: POST /custom_fields
    required fields: id
    risk: Create a new custom field; external OnePageCRM mutation, approval required
  update_custom_field:
    endpoint: PUT /custom_fields/{{ record.custom_field_id }}
    required fields: custom_field_id, id
    risk: Update a specific custom field; external OnePageCRM mutation, approval required
  delete_custom_field:
    endpoint: DELETE /custom_fields/{{ record.custom_field_id }}
    required fields: custom_field_id
    risk: Delete a specific custom field; external OnePageCRM mutation, approval required
  create_company_field:
    endpoint: POST /company_fields
    required fields: id
    risk: Create a new company field; external OnePageCRM mutation, approval required
  update_company_field:
    endpoint: PUT /company_fields/{{ record.company_field_id }}
    required fields: company_field_id, id
    risk: Update a specific company field; external OnePageCRM mutation, approval required
  delete_company_field:
    endpoint: DELETE /company_fields/{{ record.company_field_id }}
    required fields: company_field_id
    risk: Delete a specific company field; external OnePageCRM mutation, approval required
  create_predefined_action:
    endpoint: POST /predefined_actions
    required fields: id
    risk: Create a new predefined action; external OnePageCRM mutation, approval required
  update_predefined_action:
    endpoint: PUT /predefined_actions/{{ record.predefined_action_id }}
    required fields: predefined_action_id, id
    risk: Update a specific predefined action; external OnePageCRM mutation, approval required
  delete_predefined_action:
    endpoint: DELETE /predefined_actions/{{ record.predefined_action_id }}
    required fields: predefined_action_id
    risk: Delete a specific predefined action; external OnePageCRM mutation, approval required
  create_predefined_action_group:
    endpoint: POST /predefined_action_groups
    required fields: text
    risk: Create a new predefined action group; external OnePageCRM mutation, approval required
  update_predefined_action_group:
    endpoint: PUT /predefined_action_groups/{{ record.predefined_action_group_id }}
    required fields: predefined_action_group_id, text
    risk: Update a specific predefined action group; external OnePageCRM mutation, approval required
  delete_predefined_action_group:
    endpoint: DELETE /predefined_action_groups/{{ record.predefined_action_group_id }}
    required fields: predefined_action_group_id
    risk: Delete a specific predefined action group; external OnePageCRM mutation, approval required
  create_predefined_item:
    endpoint: POST /predefined_items
    required fields: name
    risk: Create a new predefined item; external OnePageCRM mutation, approval required
  update_predefined_item:
    endpoint: PUT /predefined_items/{{ record.predefined_item_id }}
    required fields: predefined_item_id, name
    risk: Update a specific predefined item; external OnePageCRM mutation, approval required
  delete_predefined_item:
    endpoint: DELETE /predefined_items/{{ record.predefined_item_id }}
    required fields: predefined_item_id
    risk: Delete a specific predefined item; external OnePageCRM mutation, approval required
  create_predefined_item_group:
    endpoint: POST /predefined_item_groups
    required fields: name
    risk: Create a new predefined item group; external OnePageCRM mutation, approval required
  delete_predefined_item_group:
    endpoint: DELETE /predefined_item_groups/{{ record.predefined_item_group_id }}
    required fields: predefined_item_group_id
    risk: Delete a specific predefined item group; external OnePageCRM mutation, approval required
  create_note:
    endpoint: POST /notes
    required fields: contact_id
    risk: Create a new note; external OnePageCRM mutation, approval required
  update_note:
    endpoint: PUT /notes/{{ record.note_id }}
    required fields: note_id, contact_id
    risk: Update a specific note; external OnePageCRM mutation, approval required
  delete_note:
    endpoint: DELETE /notes/{{ record.note_id }}
    required fields: note_id
    risk: Delete a specific note; external OnePageCRM mutation, approval required
  create_note_attachment:
    endpoint: POST /notes/{{ record.note_id }}/attachments
    required fields: note_id, reference_id
    risk: Create attachment and assign it to an existing note; external OnePageCRM mutation, approval required
  create_call:
    endpoint: POST /calls
    required fields: contact_id
    risk: Create a call; external OnePageCRM mutation, approval required
  update_call:
    endpoint: PUT /calls/{{ record.call_id }}
    required fields: call_id, contact_id
    risk: Update a specific call; external OnePageCRM mutation, approval required
  delete_call:
    endpoint: DELETE /calls/{{ record.call_id }}
    required fields: call_id
    risk: Delete a specific call; external OnePageCRM mutation, approval required
  create_call_attachment:
    endpoint: POST /calls/{{ record.call_id }}/attachments
    required fields: call_id, reference_id
    risk: Create attachment and assign it to an existing call; external OnePageCRM mutation, approval required
  create_meeting:
    endpoint: POST /meetings
    required fields: contact_id
    risk: Create a meeting; external OnePageCRM mutation, approval required
  update_meeting:
    endpoint: PUT /meetings/{{ record.meeting_id }}
    required fields: meeting_id, contact_id
    risk: Update a specific meeting; external OnePageCRM mutation, approval required
  delete_meeting:
    endpoint: DELETE /meetings/{{ record.meeting_id }}
    required fields: meeting_id
    risk: Delete a specific meeting; external OnePageCRM mutation, approval required
  create_meeting_attachment:
    endpoint: POST /meetings/{{ record.meeting_id }}/attachments
    required fields: meeting_id, reference_id
    risk: Create attachment and assign it to an existing meeting; external OnePageCRM mutation, approval required
  create_deal:
    endpoint: POST /deals
    required fields: contact_id
    risk: Create a new deal; external OnePageCRM mutation, approval required
  update_deal:
    endpoint: PUT /deals/{{ record.deal_id }}
    required fields: deal_id, contact_id
    risk: Update a specific deal; external OnePageCRM mutation, approval required
  delete_deal:
    endpoint: DELETE /deals/{{ record.deal_id }}
    required fields: deal_id
    risk: Delete a specific deal; external OnePageCRM mutation, approval required
  create_deal_attachment:
    endpoint: POST /deals/{{ record.deal_id }}/attachments
    required fields: deal_id, reference_id
    risk: Create attachment and assign it to an existing deal; external OnePageCRM mutation, approval required
  create_attachment:
    endpoint: POST /attachments
    required fields: reference_id
    risk: Create a new attachment; external OnePageCRM mutation, approval required
  update_attachment:
    endpoint: PATCH /attachments/{{ record.attachment_id }}
    required fields: attachment_id, attachment
    risk: Sets/updates attachment custom file name; external OnePageCRM mutation, approval required
  delete_attachment:
    endpoint: DELETE /attachments/{{ record.attachment_id }}
    required fields: attachment_id
    risk: Delete a specific attachment; external OnePageCRM mutation, approval required
  pin_attachment:
    endpoint: PATCH /attachments/{{ record.attachment_id }}/pin
    required fields: attachment_id
    risk: Pin attachment to its owner contact through its note/call/deal; external OnePageCRM mutation, approval required
  unpin_attachment:
    endpoint: PATCH /attachments/{{ record.attachment_id }}/unpin
    required fields: attachment_id
    risk: Unpin attachment from its owner contact through its note/call/deal; external OnePageCRM mutation, approval required
  create_relationship_type:
    endpoint: POST /relationship_types
    required fields: variants
    risk: Create a new relationship type; external OnePageCRM mutation, approval required
  update_relationship_type:
    endpoint: PUT /relationship_types/{{ record.relationship_type_id }}
    required fields: relationship_type_id, variants
    risk: Update a specific relationship type; external OnePageCRM mutation, approval required
  delete_relationship_type:
    endpoint: DELETE /relationship_types/{{ record.relationship_type_id }}
    required fields: relationship_type_id
    risk: Delete a relationship type; external OnePageCRM mutation, approval required
  create_action:
    endpoint: POST /actions
    required fields: contact_id
    risk: Create a new action; external OnePageCRM mutation, approval required
  update_action:
    endpoint: PUT /actions/{{ record.action_id }}
    required fields: action_id, contact_id
    risk: Update a specific action; external OnePageCRM mutation, approval required
  delete_action:
    endpoint: DELETE /actions/{{ record.action_id }}
    required fields: action_id
    risk: Delete a specific action; external OnePageCRM mutation, approval required
  unassign_action:
    endpoint: PUT /actions/{{ record.action_id }}/unassign
    required fields: action_id
    risk: Unassign a specific action (from the currently assigned user); external OnePageCRM mutation, approval required
  mark_as_done_action:
    endpoint: PUT /actions/{{ record.action_id }}/mark_as_done
    required fields: action_id
    risk: Mark a specific action as complete; external OnePageCRM mutation, approval required
  undo_completion_action:
    endpoint: PUT /actions/{{ record.action_id }}/undo_completion
    required fields: action_id
    risk: Undo action completion; external OnePageCRM mutation, approval required
  promote_action:
    endpoint: PUT /actions/{{ record.action_id }}/promote
    required fields: action_id
    risk: Specify action to be promoted as the logged API users next action; external OnePageCRM mutation, approval required
  revert_promotion_action:
    endpoint: PUT /actions/{{ record.action_id }}/revert_promotion
    required fields: action_id
    risk: Undo action promotion; external OnePageCRM mutation, approval required
  swap_action:
    endpoint: PUT /actions/{{ record.action_id }}/swap
    required fields: action_id
    risk: Specify action to be swapped in as the logged API users next action; external OnePageCRM mutation, approval required
  update_company:
    endpoint: PUT /companies/{{ record.company_id }}
    required fields: company_id, name
    risk: Update a specific company; external OnePageCRM mutation, approval required
  create_company_linked_contact:
    endpoint: POST /companies/{{ record.company_id }}/linked_contacts
    required fields: company_id, contact_id
    risk: Link a contact to a specific company; external OnePageCRM mutation, approval required
  delete_company_linked_contact:
    endpoint: DELETE /companies/{{ record.company_id }}/linked_contacts/{{ record.contact_id }}
    required fields: company_id, contact_id
    risk: Unlink a contact from a company; external OnePageCRM mutation, approval required
  enable_company_synced_status:
    endpoint: POST /companies/{{ record.company_id }}/synced_status
    required fields: company_id, status_id
    risk: Enable company status sync; external OnePageCRM mutation, approval required
  delete_company_synced_status:
    endpoint: DELETE /companies/{{ record.company_id }}/synced_status
    required fields: company_id
    risk: Disable company status sync; external OnePageCRM mutation, approval required
  delete_company_logo:
    endpoint: DELETE /companies/{{ record.company_id }}/logo
    required fields: company_id
    risk: Delete logo in then given company; external OnePageCRM mutation, approval required
  create_contact:
    endpoint: POST /contacts
    required fields: title
    risk: Create a contact; external OnePageCRM mutation, approval required
  update_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}
    required fields: contact_id, title
    risk: Update a specific contact; external OnePageCRM mutation, approval required
  delete_contact:
    endpoint: DELETE /contacts/{{ record.contact_id }}
    required fields: contact_id
    risk: Delete a specific contact; external OnePageCRM mutation, approval required
  delete_contact_contact_photo:
    endpoint: DELETE /contacts/{{ record.contact_id }}/contact_photo
    required fields: contact_id
    risk: Remove a contact's photo; external OnePageCRM mutation, approval required
  save_contact_to_google_contacts:
    endpoint: POST /contacts/{{ record.contact_id }}/google_contacts
    required fields: contact_id
    risk: Save a specific OnePageCRM contact to Google Contacts; external OnePageCRM mutation, approval required
  create_contact_action:
    endpoint: POST /contacts/{{ record.contact_id }}/actions
    required fields: contact_id, assignee_id
    risk: Create an action for a specific contact; external OnePageCRM mutation, approval required
  create_contact_deal:
    endpoint: POST /contacts/{{ record.contact_id }}/deals
    required fields: contact_id, owner_id
    risk: Create a deal for a specific contact; external OnePageCRM mutation, approval required
  create_contact_note:
    endpoint: POST /contacts/{{ record.contact_id }}/notes
    required fields: contact_id, text
    risk: Create a note for a specific contact; external OnePageCRM mutation, approval required
  create_contact_call:
    endpoint: POST /contacts/{{ record.contact_id }}/calls
    required fields: contact_id, call_time_int
    risk: Create a call for a specific contact; external OnePageCRM mutation, approval required
  create_contact_meeting:
    endpoint: POST /contacts/{{ record.contact_id }}/meetings
    required fields: contact_id, meeting_time_int
    risk: Create a meeting for a specific contact; external OnePageCRM mutation, approval required
  create_contact_relationship:
    endpoint: POST /contacts/{{ record.contact_id }}/relationships
    required fields: contact_id, relationship_type_id
    risk: Create a relationships for a specific contact; external OnePageCRM mutation, approval required
  update_relationship:
    endpoint: PUT /contacts/{{ record.contact_id }}/relationships/{{ record.relationship_id }}
    required fields: contact_id, relationship_id, relationship_type_id
    risk: Update a specific relationship; external OnePageCRM mutation, approval required
  delete_contact_relationship:
    endpoint: DELETE /contacts/{{ record.contact_id }}/relationships/{{ record.relationship_id }}
    required fields: contact_id, relationship_id
    risk: Delete a relationship; external OnePageCRM mutation, approval required
  assign_contact_tag:
    endpoint: PUT /contacts/{{ record.contact_id }}/assign_tag/{{ record.tag_name }}
    required fields: contact_id, tag_name
    risk: Assign a tag to a specific contact; external OnePageCRM mutation, approval required
  unassign_contact_tag:
    endpoint: PUT /contacts/{{ record.contact_id }}/unassign_tag/{{ record.tag_name }}
    required fields: contact_id, tag_name
    risk: Remove a tag from a specific contact; external OnePageCRM mutation, approval required
  change_contact_status:
    endpoint: PUT /contacts/{{ record.contact_id }}/change_status/{{ record.status_id }}
    required fields: contact_id, status_id
    risk: Change the status of a specific contact; external OnePageCRM mutation, approval required
  change_contact_owner:
    endpoint: PUT /contacts/{{ record.contact_id }}/change_owner/{{ record.owner_id }}
    required fields: contact_id, owner_id
    risk: Change the owner of a specific contact; external OnePageCRM mutation, approval required
  star_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}/star
    required fields: contact_id
    risk: Apply a star to a specific contact; external OnePageCRM mutation, approval required
  unstar_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}/unstar
    required fields: contact_id
    risk: Remove star from a specific contact; external OnePageCRM mutation, approval required
  close_sales_cycle_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}/close_sales_cycle
    required fields: contact_id, comment
    risk: Close the sales cycle for a specific contact; external OnePageCRM mutation, approval required
  force_close_sales_cycle_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}/force_close_sales_cycle
    required fields: contact_id, comment
    risk: Force close the sales cycle for a specific contact; external OnePageCRM mutation, approval required
  reopen_sales_cycle_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}/reopen_sales_cycle
    required fields: contact_id
    risk: Reopen the sales cycle for a specific contact; external OnePageCRM mutation, approval required
  split_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}/split
    required fields: contact_id, company_name
    risk: Split a contact from their current company (and potentially to a new company); external OnePageCRM mutation, approval required
  mark_as_read_notification:
    endpoint: POST /notifications/{{ record.notification_id }}/mark_as_read
    required fields: notification_id
    risk: Marks given notification as read; external OnePageCRM mutation, approval required
  mark_all_notifications_as_read:
    endpoint: POST /notifications/mark_all_as_read
    risk: Marks all users' notifications as read; external OnePageCRM mutation, approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: Delete a specific webhook; external OnePageCRM mutation, approval required

SECURITY
  read risk: external OnePageCRM API read of CRM contact, deal, task, account, and configuration data
  write risk: external OnePageCRM API mutations can create, update, complete, tag, export, disable, or delete live CRM records and account configuration
  approval: write actions require explicit approval; reads require none
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run OnePageCRM's declared streams and reverse-ETL actions.
  Usage: pm onepagecrm <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    action list - Run the action ETL stream [intent=etl availability=implemented stream=action]
    action stream list - Run the action stream ETL stream [intent=etl availability=implemented stream=action_stream]
    actions list - Run the actions ETL stream [intent=etl availability=implemented stream=actions]
    api delete contacts delete - Documented DELETE /contacts/delete (not implemented) [intent=direct_write availability=not_implemented operation=onepagecrm.delete.contacts-delete]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get attachments s3-form - Documented GET /attachments/s3_form (not implemented) [intent=direct_read availability=not_implemented operation=onepagecrm.get.attachments-s3-form]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch companies company-id logo - Documented PATCH /companies/{company_id}/logo (not implemented) [intent=direct_write availability=not_implemented operation=onepagecrm.patch.companies-company-id-logo]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post change-auth-key - Documented POST /change_auth_key (not implemented) [intent=direct_write availability=not_implemented operation=onepagecrm.post.change-auth-key]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post companies company-id logo - Documented POST /companies/{company_id}/logo (not implemented) [intent=direct_write availability=not_implemented operation=onepagecrm.post.companies-company-id-logo]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post contacts contact-id contact-photo - Documented POST /contacts/{contact_id}/contact_photo (not implemented) [intent=direct_write availability=not_implemented operation=onepagecrm.post.contacts-contact-id-contact-photo]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put contacts contact-id contact-photo - Documented PUT /contacts/{contact_id}/contact_photo (not implemented) [intent=direct_write availability=not_implemented operation=onepagecrm.put.contacts-contact-id-contact-photo]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    assign contact tag apply - Plan and execute the assign contact tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=assign_contact_tag]; approval: requires plan, preview, approval, and execute; risk: Assign a tag to a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    bootstrap list - Run the bootstrap ETL stream [intent=etl availability=implemented stream=bootstrap]
    call list - Run the call ETL stream [intent=etl availability=implemented stream=call]
    call results list - Run the call results ETL stream [intent=etl availability=implemented stream=call_results]
    calls list - Run the calls ETL stream [intent=etl availability=implemented stream=calls]
    change contact owner apply - Plan and execute the change contact owner reverse-ETL action [intent=reverse_etl availability=not_implemented write=change_contact_owner]; approval: requires plan, preview, approval, and execute; risk: Change the owner of a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    change contact status apply - Plan and execute the change contact status reverse-ETL action [intent=reverse_etl availability=not_implemented write=change_contact_status]; approval: requires plan, preview, approval, and execute; risk: Change the status of a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    close sales cycle contact apply - Plan and execute the close sales cycle contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=close_sales_cycle_contact]; approval: requires plan, preview, approval, and execute; risk: Close the sales cycle for a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    companies list - Run the companies ETL stream [intent=etl availability=implemented stream=companies]
    company actions list - Run the company actions ETL stream [intent=etl availability=implemented stream=company_actions]
    company calls list - Run the company calls ETL stream [intent=etl availability=implemented stream=company_calls]
    company deals list - Run the company deals ETL stream [intent=etl availability=implemented stream=company_deals]
    company field list - Run the company field ETL stream [intent=etl availability=implemented stream=company_field]
    company fields list - Run the company fields ETL stream [intent=etl availability=implemented stream=company_fields]
    company linked contacts list - Run the company linked contacts ETL stream [intent=etl availability=implemented stream=company_linked_contacts]
    company list - Run the company ETL stream [intent=etl availability=implemented stream=company]
    company meetings list - Run the company meetings ETL stream [intent=etl availability=implemented stream=company_meetings]
    company notes list - Run the company notes ETL stream [intent=etl availability=implemented stream=company_notes]
    company pinned attachments list - Run the company pinned attachments ETL stream [intent=etl availability=implemented stream=company_pinned_attachments]
    contact actions list - Run the contact actions ETL stream [intent=etl availability=implemented stream=contact_actions]
    contact calls list - Run the contact calls ETL stream [intent=etl availability=implemented stream=contact_calls]
    contact deals list - Run the contact deals ETL stream [intent=etl availability=implemented stream=contact_deals]
    contact list - Run the contact ETL stream [intent=etl availability=implemented stream=contact]
    contact meetings list - Run the contact meetings ETL stream [intent=etl availability=implemented stream=contact_meetings]
    contact notes list - Run the contact notes ETL stream [intent=etl availability=implemented stream=contact_notes]
    contact pinned attachments list - Run the contact pinned attachments ETL stream [intent=etl availability=implemented stream=contact_pinned_attachments]
    contact relationship list - Run the contact relationship ETL stream [intent=etl availability=implemented stream=contact_relationship]
    contact relationships list - Run the contact relationships ETL stream [intent=etl availability=implemented stream=contact_relationships]
    contacts cascade after list - Run the contacts cascade after ETL stream [intent=etl availability=implemented stream=contacts_cascade_after]
    contacts cascade list - Run the contacts cascade ETL stream [intent=etl availability=implemented stream=contacts_cascade]
    contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
    countries list - Run the countries ETL stream [intent=etl availability=implemented stream=countries]
    create action apply - Plan and execute the create action reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_action]; approval: requires plan, preview, approval, and execute; risk: Create a new action; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create attachment apply - Plan and execute the create attachment reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_attachment]; approval: requires plan, preview, approval, and execute; risk: Create a new attachment; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create call apply - Plan and execute the create call reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_call]; approval: requires plan, preview, approval, and execute; risk: Create a call; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create call attachment apply - Plan and execute the create call attachment reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_call_attachment]; approval: requires plan, preview, approval, and execute; risk: Create attachment and assign it to an existing call; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create company field apply - Plan and execute the create company field reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_company_field]; approval: requires plan, preview, approval, and execute; risk: Create a new company field; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create company linked contact apply - Plan and execute the create company linked contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_company_linked_contact]; approval: requires plan, preview, approval, and execute; risk: Link a contact to a specific company; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact action apply - Plan and execute the create contact action reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact_action]; approval: requires plan, preview, approval, and execute; risk: Create an action for a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact apply - Plan and execute the create contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact]; approval: requires plan, preview, approval, and execute; risk: Create a contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact call apply - Plan and execute the create contact call reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact_call]; approval: requires plan, preview, approval, and execute; risk: Create a call for a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact deal apply - Plan and execute the create contact deal reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact_deal]; approval: requires plan, preview, approval, and execute; risk: Create a deal for a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact meeting apply - Plan and execute the create contact meeting reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact_meeting]; approval: requires plan, preview, approval, and execute; risk: Create a meeting for a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact note apply - Plan and execute the create contact note reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact_note]; approval: requires plan, preview, approval, and execute; risk: Create a note for a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create contact relationship apply - Plan and execute the create contact relationship reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_contact_relationship]; approval: requires plan, preview, approval, and execute; risk: Create a relationships for a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create custom field apply - Plan and execute the create custom field reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_custom_field]; approval: requires plan, preview, approval, and execute; risk: Create a new custom field; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create deal apply - Plan and execute the create deal reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_deal]; approval: requires plan, preview, approval, and execute; risk: Create a new deal; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create deal attachment apply - Plan and execute the create deal attachment reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_deal_attachment]; approval: requires plan, preview, approval, and execute; risk: Create attachment and assign it to an existing deal; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create deal field apply - Plan and execute the create deal field reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_deal_field]; approval: requires plan, preview, approval, and execute; risk: Create a new deal field; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create lead source apply - Plan and execute the create lead source reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_lead_source]; approval: requires plan, preview, approval, and execute; risk: Create a new lead source; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create meeting apply - Plan and execute the create meeting reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_meeting]; approval: requires plan, preview, approval, and execute; risk: Create a meeting; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create meeting attachment apply - Plan and execute the create meeting attachment reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_meeting_attachment]; approval: requires plan, preview, approval, and execute; risk: Create attachment and assign it to an existing meeting; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create note apply - Plan and execute the create note reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_note]; approval: requires plan, preview, approval, and execute; risk: Create a new note; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create note attachment apply - Plan and execute the create note attachment reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_note_attachment]; approval: requires plan, preview, approval, and execute; risk: Create attachment and assign it to an existing note; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create predefined action apply - Plan and execute the create predefined action reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_predefined_action]; approval: requires plan, preview, approval, and execute; risk: Create a new predefined action; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create predefined action group apply - Plan and execute the create predefined action group reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_predefined_action_group]; approval: requires plan, preview, approval, and execute; risk: Create a new predefined action group; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create predefined item apply - Plan and execute the create predefined item reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_predefined_item]; approval: requires plan, preview, approval, and execute; risk: Create a new predefined item; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create predefined item group apply - Plan and execute the create predefined item group reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_predefined_item_group]; approval: requires plan, preview, approval, and execute; risk: Create a new predefined item group; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create relationship type apply - Plan and execute the create relationship type reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_relationship_type]; approval: requires plan, preview, approval, and execute; risk: Create a new relationship type; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create status apply - Plan and execute the create status reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_status]; approval: requires plan, preview, approval, and execute; risk: Create a new status; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    custom field list - Run the custom field ETL stream [intent=etl availability=implemented stream=custom_field]
    custom fields list - Run the custom fields ETL stream [intent=etl availability=implemented stream=custom_fields]
    deal field list - Run the deal field ETL stream [intent=etl availability=implemented stream=deal_field]
    deal fields list - Run the deal fields ETL stream [intent=etl availability=implemented stream=deal_fields]
    deal list - Run the deal ETL stream [intent=etl availability=implemented stream=deal]
    deals list - Run the deals ETL stream [intent=etl availability=implemented stream=deals]
    delete action apply - Plan and execute the delete action reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_action]; approval: requires plan, preview, approval, and execute; risk: Delete a specific action; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete attachment apply - Plan and execute the delete attachment reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_attachment]; approval: requires plan, preview, approval, and execute; risk: Delete a specific attachment; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete call apply - Plan and execute the delete call reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_call]; approval: requires plan, preview, approval, and execute; risk: Delete a specific call; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete company field apply - Plan and execute the delete company field reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_company_field]; approval: requires plan, preview, approval, and execute; risk: Delete a specific company field; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete company linked contact apply - Plan and execute the delete company linked contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_company_linked_contact]; approval: requires plan, preview, approval, and execute; risk: Unlink a contact from a company; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete company logo apply - Plan and execute the delete company logo reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_company_logo]; approval: requires plan, preview, approval, and execute; risk: Delete logo in then given company; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete company synced status apply - Plan and execute the delete company synced status reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_company_synced_status]; approval: requires plan, preview, approval, and execute; risk: Disable company status sync; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete contact apply - Plan and execute the delete contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_contact]; approval: requires plan, preview, approval, and execute; risk: Delete a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete contact contact photo apply - Plan and execute the delete contact contact photo reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_contact_contact_photo]; approval: requires plan, preview, approval, and execute; risk: Remove a contact's photo; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete contact relationship apply - Plan and execute the delete contact relationship reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_contact_relationship]; approval: requires plan, preview, approval, and execute; risk: Delete a relationship; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete custom field apply - Plan and execute the delete custom field reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_custom_field]; approval: requires plan, preview, approval, and execute; risk: Delete a specific custom field; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete deal apply - Plan and execute the delete deal reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_deal]; approval: requires plan, preview, approval, and execute; risk: Delete a specific deal; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete deal field apply - Plan and execute the delete deal field reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_deal_field]; approval: requires plan, preview, approval, and execute; risk: Delete a specific deal field; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete lead source apply - Plan and execute the delete lead source reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_lead_source]; approval: requires plan, preview, approval, and execute; risk: Delete a specific lead source; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete meeting apply - Plan and execute the delete meeting reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_meeting]; approval: requires plan, preview, approval, and execute; risk: Delete a specific meeting; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete note apply - Plan and execute the delete note reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_note]; approval: requires plan, preview, approval, and execute; risk: Delete a specific note; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete predefined action apply - Plan and execute the delete predefined action reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_predefined_action]; approval: requires plan, preview, approval, and execute; risk: Delete a specific predefined action; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete predefined action group apply - Plan and execute the delete predefined action group reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_predefined_action_group]; approval: requires plan, preview, approval, and execute; risk: Delete a specific predefined action group; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete predefined item apply - Plan and execute the delete predefined item reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_predefined_item]; approval: requires plan, preview, approval, and execute; risk: Delete a specific predefined item; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete predefined item group apply - Plan and execute the delete predefined item group reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_predefined_item_group]; approval: requires plan, preview, approval, and execute; risk: Delete a specific predefined item group; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete relationship type apply - Plan and execute the delete relationship type reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_relationship_type]; approval: requires plan, preview, approval, and execute; risk: Delete a relationship type; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete status apply - Plan and execute the delete status reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_status]; approval: requires plan, preview, approval, and execute; risk: Delete a specific status; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: Delete a specific webhook; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    enable company synced status apply - Plan and execute the enable company synced status reverse-ETL action [intent=reverse_etl availability=not_implemented write=enable_company_synced_status]; approval: requires plan, preview, approval, and execute; risk: Enable company status sync; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    filter list - Run the filter ETL stream [intent=etl availability=implemented stream=filter]
    filtered contacts list - Run the filtered contacts ETL stream [intent=etl availability=implemented stream=filtered_contacts]
    filters list - Run the filters ETL stream [intent=etl availability=implemented stream=filters]
    force close sales cycle contact apply - Plan and execute the force close sales cycle contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=force_close_sales_cycle_contact]; approval: requires plan, preview, approval, and execute; risk: Force close the sales cycle for a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    lead source list - Run the lead source ETL stream [intent=etl availability=implemented stream=lead_source]
    lead sources list - Run the lead sources ETL stream [intent=etl availability=implemented stream=lead_sources]
    mark all notifications as read apply - Plan and execute the mark all notifications as read reverse-ETL action [intent=reverse_etl availability=implemented write=mark_all_notifications_as_read]; approval: requires plan, preview, approval, and execute; risk: Marks all users' notifications as read; external OnePageCRM mutation, approval required
    mark as done action apply - Plan and execute the mark as done action reverse-ETL action [intent=reverse_etl availability=not_implemented write=mark_as_done_action]; approval: requires plan, preview, approval, and execute; risk: Mark a specific action as complete; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    mark as read notification apply - Plan and execute the mark as read notification reverse-ETL action [intent=reverse_etl availability=not_implemented write=mark_as_read_notification]; approval: requires plan, preview, approval, and execute; risk: Marks given notification as read; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    meeting list - Run the meeting ETL stream [intent=etl availability=implemented stream=meeting]
    meetings list - Run the meetings ETL stream [intent=etl availability=implemented stream=meetings]
    note list - Run the note ETL stream [intent=etl availability=implemented stream=note]
    notes list - Run the notes ETL stream [intent=etl availability=implemented stream=notes]
    notification list - Run the notification ETL stream [intent=etl availability=implemented stream=notification]
    notifications list - Run the notifications ETL stream [intent=etl availability=implemented stream=notifications]
    pin attachment apply - Plan and execute the pin attachment reverse-ETL action [intent=reverse_etl availability=not_implemented write=pin_attachment]; approval: requires plan, preview, approval, and execute; risk: Pin attachment to its owner contact through its note/call/deal; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    pipeline list - Run the pipeline ETL stream [intent=etl availability=implemented stream=pipeline]
    pipelines list - Run the pipelines ETL stream [intent=etl availability=implemented stream=pipelines]
    predefined action group list - Run the predefined action group ETL stream [intent=etl availability=implemented stream=predefined_action_group]
    predefined action groups list - Run the predefined action groups ETL stream [intent=etl availability=implemented stream=predefined_action_groups]
    predefined action list - Run the predefined action ETL stream [intent=etl availability=implemented stream=predefined_action]
    predefined actions list - Run the predefined actions ETL stream [intent=etl availability=implemented stream=predefined_actions]
    predefined item group list - Run the predefined item group ETL stream [intent=etl availability=implemented stream=predefined_item_group]
    predefined item groups list - Run the predefined item groups ETL stream [intent=etl availability=implemented stream=predefined_item_groups]
    predefined item list - Run the predefined item ETL stream [intent=etl availability=implemented stream=predefined_item]
    predefined items list - Run the predefined items ETL stream [intent=etl availability=implemented stream=predefined_items]
    promote action apply - Plan and execute the promote action reverse-ETL action [intent=reverse_etl availability=not_implemented write=promote_action]; approval: requires plan, preview, approval, and execute; risk: Specify action to be promoted as the logged API users next action; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    relationship type list - Run the relationship type ETL stream [intent=etl availability=implemented stream=relationship_type]
    relationship types list - Run the relationship types ETL stream [intent=etl availability=implemented stream=relationship_types]
    reopen sales cycle contact apply - Plan and execute the reopen sales cycle contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=reopen_sales_cycle_contact]; approval: requires plan, preview, approval, and execute; risk: Reopen the sales cycle for a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    revert promotion action apply - Plan and execute the revert promotion action reverse-ETL action [intent=reverse_etl availability=not_implemented write=revert_promotion_action]; approval: requires plan, preview, approval, and execute; risk: Undo action promotion; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    save contact to google contacts apply - Plan and execute the save contact to google contacts reverse-ETL action [intent=reverse_etl availability=not_implemented write=save_contact_to_google_contacts]; approval: requires plan, preview, approval, and execute; risk: Save a specific OnePageCRM contact to Google Contacts; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    split contact apply - Plan and execute the split contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=split_contact]; approval: requires plan, preview, approval, and execute; risk: Split a contact from their current company (and potentially to a new company); external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    star contact apply - Plan and execute the star contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=star_contact]; approval: requires plan, preview, approval, and execute; risk: Apply a star to a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    status list - Run the status ETL stream [intent=etl availability=implemented stream=status]
    statuses list - Run the statuses ETL stream [intent=etl availability=implemented stream=statuses]
    swap action apply - Plan and execute the swap action reverse-ETL action [intent=reverse_etl availability=not_implemented write=swap_action]; approval: requires plan, preview, approval, and execute; risk: Specify action to be swapped in as the logged API users next action; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    team stream list - Run the team stream ETL stream [intent=etl availability=implemented stream=team_stream]
    unassign action apply - Plan and execute the unassign action reverse-ETL action [intent=reverse_etl availability=not_implemented write=unassign_action]; approval: requires plan, preview, approval, and execute; risk: Unassign a specific action (from the currently assigned user); external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    unassign contact tag apply - Plan and execute the unassign contact tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=unassign_contact_tag]; approval: requires plan, preview, approval, and execute; risk: Remove a tag from a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    undo completion action apply - Plan and execute the undo completion action reverse-ETL action [intent=reverse_etl availability=not_implemented write=undo_completion_action]; approval: requires plan, preview, approval, and execute; risk: Undo action completion; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    unpin attachment apply - Plan and execute the unpin attachment reverse-ETL action [intent=reverse_etl availability=not_implemented write=unpin_attachment]; approval: requires plan, preview, approval, and execute; risk: Unpin attachment from its owner contact through its note/call/deal; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    unstar contact apply - Plan and execute the unstar contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=unstar_contact]; approval: requires plan, preview, approval, and execute; risk: Remove star from a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update action apply - Plan and execute the update action reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_action]; approval: requires plan, preview, approval, and execute; risk: Update a specific action; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update attachment apply - Plan and execute the update attachment reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_attachment]; approval: requires plan, preview, approval, and execute; risk: Sets/updates attachment custom file name; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update call apply - Plan and execute the update call reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_call]; approval: requires plan, preview, approval, and execute; risk: Update a specific call; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update company apply - Plan and execute the update company reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_company]; approval: requires plan, preview, approval, and execute; risk: Update a specific company; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update company field apply - Plan and execute the update company field reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_company_field]; approval: requires plan, preview, approval, and execute; risk: Update a specific company field; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update contact apply - Plan and execute the update contact reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_contact]; approval: requires plan, preview, approval, and execute; risk: Update a specific contact; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update custom field apply - Plan and execute the update custom field reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_custom_field]; approval: requires plan, preview, approval, and execute; risk: Update a specific custom field; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update deal apply - Plan and execute the update deal reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_deal]; approval: requires plan, preview, approval, and execute; risk: Update a specific deal; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update deal field apply - Plan and execute the update deal field reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_deal_field]; approval: requires plan, preview, approval, and execute; risk: Update a specific deal field; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update lead source apply - Plan and execute the update lead source reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_lead_source]; approval: requires plan, preview, approval, and execute; risk: Update a specific lead source; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update meeting apply - Plan and execute the update meeting reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_meeting]; approval: requires plan, preview, approval, and execute; risk: Update a specific meeting; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update note apply - Plan and execute the update note reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_note]; approval: requires plan, preview, approval, and execute; risk: Update a specific note; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update predefined action apply - Plan and execute the update predefined action reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_predefined_action]; approval: requires plan, preview, approval, and execute; risk: Update a specific predefined action; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update predefined action group apply - Plan and execute the update predefined action group reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_predefined_action_group]; approval: requires plan, preview, approval, and execute; risk: Update a specific predefined action group; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update predefined item apply - Plan and execute the update predefined item reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_predefined_item]; approval: requires plan, preview, approval, and execute; risk: Update a specific predefined item; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update relationship apply - Plan and execute the update relationship reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_relationship]; approval: requires plan, preview, approval, and execute; risk: Update a specific relationship; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update relationship type apply - Plan and execute the update relationship type reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_relationship_type]; approval: requires plan, preview, approval, and execute; risk: Update a specific relationship type; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update status apply - Plan and execute the update status reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_status]; approval: requires plan, preview, approval, and execute; risk: Update a specific status; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update user apply - Plan and execute the update user reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: Update a specific user; external OnePageCRM mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    user list - Run the user ETL stream [intent=etl availability=implemented stream=user]
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
    webhook list - Run the webhook ETL stream [intent=etl availability=implemented stream=webhook]
    webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect onepagecrm

  # Inspect as structured JSON
  pm connectors inspect onepagecrm --json

AGENT WORKFLOW
  - Run pm connectors inspect onepagecrm before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
