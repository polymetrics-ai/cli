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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
    fields: company_name(), created_at(), first_name(), id(), job_title(), last_name(), owner_id(), starred(), status_id(), updated_at()
  deals:
    primary key: id
    cursor: updated_at
    fields: amount(), contact_id(), created_at(), currency(), expected_close_date(), id(), name(), owner_id(), stage(), status(), updated_at()
  actions:
    primary key: id
    cursor: updated_at
    fields: assignee_id(), contact_id(), created_at(), date(), done(), id(), status(), text(), updated_at()
  companies:
    primary key: id
    cursor: updated_at
    fields: created_at(), description(), id(), name(), phone(), updated_at(), url()
  users:
    primary key: id
    fields: email(), first_name(), id(), last_name(), role(), status()
  bootstrap:
    primary key: id
    fields: id()
  user:
    primary key: id
    fields: id()
  lead_sources:
    primary key: id
    fields: id()
  lead_source:
    primary key: id
    fields: id()
  statuses:
    primary key: id
    fields: id()
  status:
    primary key: id
    fields: id()
  deal_fields:
    primary key: id
    fields: id()
  deal_field:
    primary key: id
    fields: id()
  custom_fields:
    primary key: id
    fields: id()
  custom_field:
    primary key: id
    fields: id()
  company_fields:
    primary key: id
    fields: id()
  company_field:
    primary key: id
    fields: id()
  predefined_actions:
    primary key: id
    fields: id()
  predefined_action:
    primary key: id
    fields: id()
  predefined_action_groups:
    primary key: id
    fields: id()
  predefined_action_group:
    primary key: id
    fields: id()
  predefined_items:
    primary key: id
    fields: id()
  predefined_item:
    primary key: id
    fields: id()
  predefined_item_groups:
    primary key: id
    fields: id()
  predefined_item_group:
    primary key: id
    fields: id()
  notes:
    primary key: id
    fields: id()
  note:
    primary key: id
    fields: id()
  calls:
    primary key: id
    fields: id()
  call:
    primary key: id
    fields: id()
  call_results:
    primary key: id
    fields: id()
  meetings:
    primary key: id
    fields: id()
  meeting:
    primary key: id
    fields: id()
  deal:
    primary key: id
    fields: id()
  relationship_types:
    primary key: id
    fields: id()
  relationship_type:
    primary key: id
    fields: id()
  countries:
    primary key: id
    fields: id()
  action:
    primary key: id
    fields: id()
  filters:
    primary key: id
    fields: id()
  filter:
    primary key: id
    fields: id()
  company:
    primary key: id
    fields: id()
  company_actions:
    primary key: id
    fields: id()
  company_deals:
    primary key: id
    fields: id()
  company_notes:
    primary key: id
    fields: id()
  company_calls:
    primary key: id
    fields: id()
  company_meetings:
    primary key: id
    fields: id()
  company_linked_contacts:
    primary key: id
    fields: id()
  company_pinned_attachments:
    primary key: id
    fields: id()
  contact:
    primary key: id
    fields: id()
  filtered_contacts:
    primary key: id
    fields: id()
  contact_actions:
    primary key: id
    fields: id()
  contact_deals:
    primary key: id
    fields: id()
  contact_notes:
    primary key: id
    fields: id()
  contact_calls:
    primary key: id
    fields: id()
  contact_meetings:
    primary key: id
    fields: id()
  contact_relationships:
    primary key: id
    fields: id()
  contact_relationship:
    primary key: id
    fields: id()
  contact_pinned_attachments:
    primary key: id
    fields: id()
  contacts_cascade:
    primary key: id
    fields: id()
  contacts_cascade_after:
    primary key: id
    fields: id()
  action_stream:
    primary key: id
    fields: id()
  team_stream:
    primary key: id
    fields: id()
  notifications:
    primary key: id
    fields: id()
  notification:
    primary key: id
    fields: id()
  webhooks:
    primary key: id
    fields: id()
  webhook:
    primary key: id
    fields: id()
  pipelines:
    primary key: id
    fields: id()
  pipeline:
    primary key: id
    fields: id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  update_user:
    endpoint: PUT /users/{{ record.user_id }}
    required fields: user_id
    risk: Update a specific user; external OnePageCRM mutation, approval required
  create_lead_source:
    endpoint: POST /lead_sources
    risk: Create a new lead source; external OnePageCRM mutation, approval required
  update_lead_source:
    endpoint: PUT /lead_sources/{{ record.lead_source_id }}
    required fields: lead_source_id
    risk: Update a specific lead source; external OnePageCRM mutation, approval required
  delete_lead_source:
    endpoint: DELETE /lead_sources/{{ record.lead_source_id }}
    required fields: lead_source_id
    risk: Delete a specific lead source; external OnePageCRM mutation, approval required
  create_status:
    endpoint: POST /statuses
    risk: Create a new status; external OnePageCRM mutation, approval required
  update_status:
    endpoint: PUT /statuses/{{ record.status_id }}
    required fields: status_id
    risk: Update a specific status; external OnePageCRM mutation, approval required
  delete_status:
    endpoint: DELETE /statuses/{{ record.status_id }}
    required fields: status_id
    risk: Delete a specific status; external OnePageCRM mutation, approval required
  create_deal_field:
    endpoint: POST /deal_fields
    risk: Create a new deal field; external OnePageCRM mutation, approval required
  update_deal_field:
    endpoint: PUT /deal_fields/{{ record.deal_field_id }}
    required fields: deal_field_id
    risk: Update a specific deal field; external OnePageCRM mutation, approval required
  delete_deal_field:
    endpoint: DELETE /deal_fields/{{ record.deal_field_id }}
    required fields: deal_field_id
    risk: Delete a specific deal field; external OnePageCRM mutation, approval required
  create_custom_field:
    endpoint: POST /custom_fields
    risk: Create a new custom field; external OnePageCRM mutation, approval required
  update_custom_field:
    endpoint: PUT /custom_fields/{{ record.custom_field_id }}
    required fields: custom_field_id
    risk: Update a specific custom field; external OnePageCRM mutation, approval required
  delete_custom_field:
    endpoint: DELETE /custom_fields/{{ record.custom_field_id }}
    required fields: custom_field_id
    risk: Delete a specific custom field; external OnePageCRM mutation, approval required
  create_company_field:
    endpoint: POST /company_fields
    risk: Create a new company field; external OnePageCRM mutation, approval required
  update_company_field:
    endpoint: PUT /company_fields/{{ record.company_field_id }}
    required fields: company_field_id
    risk: Update a specific company field; external OnePageCRM mutation, approval required
  delete_company_field:
    endpoint: DELETE /company_fields/{{ record.company_field_id }}
    required fields: company_field_id
    risk: Delete a specific company field; external OnePageCRM mutation, approval required
  create_predefined_action:
    endpoint: POST /predefined_actions
    risk: Create a new predefined action; external OnePageCRM mutation, approval required
  update_predefined_action:
    endpoint: PUT /predefined_actions/{{ record.predefined_action_id }}
    required fields: predefined_action_id
    risk: Update a specific predefined action; external OnePageCRM mutation, approval required
  delete_predefined_action:
    endpoint: DELETE /predefined_actions/{{ record.predefined_action_id }}
    required fields: predefined_action_id
    risk: Delete a specific predefined action; external OnePageCRM mutation, approval required
  create_predefined_action_group:
    endpoint: POST /predefined_action_groups
    risk: Create a new predefined action group; external OnePageCRM mutation, approval required
  update_predefined_action_group:
    endpoint: PUT /predefined_action_groups/{{ record.predefined_action_group_id }}
    required fields: predefined_action_group_id
    risk: Update a specific predefined action group; external OnePageCRM mutation, approval required
  delete_predefined_action_group:
    endpoint: DELETE /predefined_action_groups/{{ record.predefined_action_group_id }}
    required fields: predefined_action_group_id
    risk: Delete a specific predefined action group; external OnePageCRM mutation, approval required
  create_predefined_item:
    endpoint: POST /predefined_items
    risk: Create a new predefined item; external OnePageCRM mutation, approval required
  update_predefined_item:
    endpoint: PUT /predefined_items/{{ record.predefined_item_id }}
    required fields: predefined_item_id
    risk: Update a specific predefined item; external OnePageCRM mutation, approval required
  delete_predefined_item:
    endpoint: DELETE /predefined_items/{{ record.predefined_item_id }}
    required fields: predefined_item_id
    risk: Delete a specific predefined item; external OnePageCRM mutation, approval required
  create_predefined_item_group:
    endpoint: POST /predefined_item_groups
    risk: Create a new predefined item group; external OnePageCRM mutation, approval required
  delete_predefined_item_group:
    endpoint: DELETE /predefined_item_groups/{{ record.predefined_item_group_id }}
    required fields: predefined_item_group_id
    risk: Delete a specific predefined item group; external OnePageCRM mutation, approval required
  create_note:
    endpoint: POST /notes
    risk: Create a new note; external OnePageCRM mutation, approval required
  update_note:
    endpoint: PUT /notes/{{ record.note_id }}
    required fields: note_id
    risk: Update a specific note; external OnePageCRM mutation, approval required
  delete_note:
    endpoint: DELETE /notes/{{ record.note_id }}
    required fields: note_id
    risk: Delete a specific note; external OnePageCRM mutation, approval required
  create_note_attachment:
    endpoint: POST /notes/{{ record.note_id }}/attachments
    required fields: note_id
    risk: Create attachment and assign it to an existing note; external OnePageCRM mutation, approval required
  create_call:
    endpoint: POST /calls
    risk: Create a call; external OnePageCRM mutation, approval required
  update_call:
    endpoint: PUT /calls/{{ record.call_id }}
    required fields: call_id
    risk: Update a specific call; external OnePageCRM mutation, approval required
  delete_call:
    endpoint: DELETE /calls/{{ record.call_id }}
    required fields: call_id
    risk: Delete a specific call; external OnePageCRM mutation, approval required
  create_call_attachment:
    endpoint: POST /calls/{{ record.call_id }}/attachments
    required fields: call_id
    risk: Create attachment and assign it to an existing call; external OnePageCRM mutation, approval required
  create_meeting:
    endpoint: POST /meetings
    risk: Create a meeting; external OnePageCRM mutation, approval required
  update_meeting:
    endpoint: PUT /meetings/{{ record.meeting_id }}
    required fields: meeting_id
    risk: Update a specific meeting; external OnePageCRM mutation, approval required
  delete_meeting:
    endpoint: DELETE /meetings/{{ record.meeting_id }}
    required fields: meeting_id
    risk: Delete a specific meeting; external OnePageCRM mutation, approval required
  create_meeting_attachment:
    endpoint: POST /meetings/{{ record.meeting_id }}/attachments
    required fields: meeting_id
    risk: Create attachment and assign it to an existing meeting; external OnePageCRM mutation, approval required
  create_deal:
    endpoint: POST /deals
    risk: Create a new deal; external OnePageCRM mutation, approval required
  update_deal:
    endpoint: PUT /deals/{{ record.deal_id }}
    required fields: deal_id
    risk: Update a specific deal; external OnePageCRM mutation, approval required
  delete_deal:
    endpoint: DELETE /deals/{{ record.deal_id }}
    required fields: deal_id
    risk: Delete a specific deal; external OnePageCRM mutation, approval required
  create_deal_attachment:
    endpoint: POST /deals/{{ record.deal_id }}/attachments
    required fields: deal_id
    risk: Create attachment and assign it to an existing deal; external OnePageCRM mutation, approval required
  create_attachment:
    endpoint: POST /attachments
    risk: Create a new attachment; external OnePageCRM mutation, approval required
  update_attachment:
    endpoint: PATCH /attachments/{{ record.attachment_id }}
    required fields: attachment_id
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
    risk: Create a new relationship type; external OnePageCRM mutation, approval required
  update_relationship_type:
    endpoint: PUT /relationship_types/{{ record.relationship_type_id }}
    required fields: relationship_type_id
    risk: Update a specific relationship type; external OnePageCRM mutation, approval required
  delete_relationship_type:
    endpoint: DELETE /relationship_types/{{ record.relationship_type_id }}
    required fields: relationship_type_id
    risk: Delete a relationship type; external OnePageCRM mutation, approval required
  create_action:
    endpoint: POST /actions
    risk: Create a new action; external OnePageCRM mutation, approval required
  update_action:
    endpoint: PUT /actions/{{ record.action_id }}
    required fields: action_id
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
    required fields: company_id
    risk: Update a specific company; external OnePageCRM mutation, approval required
  create_company_linked_contact:
    endpoint: POST /companies/{{ record.company_id }}/linked_contacts
    required fields: company_id
    risk: Link a contact to a specific company; external OnePageCRM mutation, approval required
  delete_company_linked_contact:
    endpoint: DELETE /companies/{{ record.company_id }}/linked_contacts/{{ record.contact_id }}
    required fields: company_id, contact_id
    risk: Unlink a contact from a company; external OnePageCRM mutation, approval required
  enable_company_synced_status:
    endpoint: POST /companies/{{ record.company_id }}/synced_status
    required fields: company_id
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
    risk: Create a contact; external OnePageCRM mutation, approval required
  update_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}
    required fields: contact_id
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
    required fields: contact_id
    risk: Create an action for a specific contact; external OnePageCRM mutation, approval required
  create_contact_deal:
    endpoint: POST /contacts/{{ record.contact_id }}/deals
    required fields: contact_id
    risk: Create a deal for a specific contact; external OnePageCRM mutation, approval required
  create_contact_note:
    endpoint: POST /contacts/{{ record.contact_id }}/notes
    required fields: contact_id
    risk: Create a note for a specific contact; external OnePageCRM mutation, approval required
  create_contact_call:
    endpoint: POST /contacts/{{ record.contact_id }}/calls
    required fields: contact_id
    risk: Create a call for a specific contact; external OnePageCRM mutation, approval required
  create_contact_meeting:
    endpoint: POST /contacts/{{ record.contact_id }}/meetings
    required fields: contact_id
    risk: Create a meeting for a specific contact; external OnePageCRM mutation, approval required
  create_contact_relationship:
    endpoint: POST /contacts/{{ record.contact_id }}/relationships
    required fields: contact_id
    risk: Create a relationships for a specific contact; external OnePageCRM mutation, approval required
  update_relationship:
    endpoint: PUT /contacts/{{ record.contact_id }}/relationships/{{ record.relationship_id }}
    required fields: contact_id, relationship_id
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
    required fields: contact_id
    risk: Close the sales cycle for a specific contact; external OnePageCRM mutation, approval required
  force_close_sales_cycle_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}/force_close_sales_cycle
    required fields: contact_id
    risk: Force close the sales cycle for a specific contact; external OnePageCRM mutation, approval required
  reopen_sales_cycle_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}/reopen_sales_cycle
    required fields: contact_id
    risk: Reopen the sales cycle for a specific contact; external OnePageCRM mutation, approval required
  split_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}/split
    required fields: contact_id
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
