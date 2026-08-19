# pm connectors inspect outreach

```text
NAME
  pm connectors inspect outreach - Outreach connector manual

SYNOPSIS
  pm connectors inspect outreach
  pm connectors inspect outreach --json
  pm credentials add <name> --connector outreach [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and mutates Outreach REST API v2 JSON:API resources, including standard resources and caller-selected custom objects.

ICON
  id: outreach
  asset: icons/outreach.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.outreach.io/api/v2/docs

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  account_note_id
  base_url
  batch_id
  batch_item_id
  call_disposition_id
  call_id
  call_purpose_id
  compliance_request_id
  content_category_id
  content_category_membership_id
  custom_object_name
  custom_object_record_id
  email_address_id
  event_id
  favorite_id
  import_id
  kaia_recording_id
  mail_alias_id
  mailbox_id
  mailing_id
  mode
  opportunity_id
  opportunity_prospect_role_id
  opportunity_stage_id
  org_setting_id
  persona_id
  phone_number_id
  product_id
  profile_id
  prospect_id
  prospect_note_id
  purchase_id
  recipient_id
  role_id
  ruleset_id
  sequence_id
  sequence_state_id
  sequence_step_id
  sequence_template_id
  snippet_id
  stage_id
  start_date
  task_disposition_id
  task_id
  task_priority_id
  task_purpose_id
  team_id
  template_id
  user_id
  webhook_id
  access_token (secret) (required)

ETL STREAMS
  prospects:
    primary key: id
    cursor: updated_at
    fields: created_at(string), email(string), id(string), name(string), type(string), updated_at(string)
  accounts:
    primary key: id
    cursor: updated_at
    fields: created_at(string), email(string), id(string), name(string), type(string), updated_at(string)
  sequences:
    primary key: id
    cursor: updated_at
    fields: created_at(string), email(string), id(string), name(string), type(string), updated_at(string)
  mailings:
    primary key: id
    cursor: updated_at
    fields: created_at(string), email(string), id(string), name(string), type(string), updated_at(string)
  account_notes:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  account_note:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  account:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  audit_logs:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  batch_items:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  batch_item:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  batches:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  batch:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  call_dispositions:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  call_disposition:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  call_purposes:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  call_purpose:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  calls:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  call:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  compliance_requests:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  compliance_request:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  content_categories:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  content_category:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  content_category_memberships:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  content_category_membership:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  content_category_ownerships:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  duties:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  email_addresses:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  email_address:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  events:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  event:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  favorites:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  favorite:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  import:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  kaia_recordings:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  kaia_recording:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  mail_aliases:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  mail_alias:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  mailboxes:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  mailbox:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  mailing:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  opportunities:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  opportunity:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  opportunity_prospect_roles:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  opportunity_prospect_role:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  opportunity_stages:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  opportunity_stage:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  org_setting:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  personas:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  persona:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  phone_numbers:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  phone_number:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  products:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  product:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  profiles:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  profile:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  prospect_notes:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  prospect_note:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  prospect:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  purchases:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  purchase:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  recipients:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  recipient:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  roles:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  role:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  rulesets:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  ruleset:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  sequence_states:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  sequence_state:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  sequence_steps:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  sequence_step:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  sequence_templates:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  sequence_template:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  sequence:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  snippets:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  snippet:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  stages:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  stage:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  task_dispositions:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  task_disposition:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  task_priorities:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  task_priority:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  task_purposes:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  task_purpose:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  tasks:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  task:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  teams:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  team:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  templates:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  template:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  users:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  user:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  webhooks:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  webhook:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  schema_definitions:
    primary key: name
    fields: definitions(object), name(string), stability(string), type(string)
  custom_object_records:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)
  custom_object_record:
    primary key: id
    fields: attributes(object), id(string), links(object), relationships(object), type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_account_note:
    endpoint: POST /accountNotes
    required fields: data
    risk: external Outreach mutation; create account note; approval required
  update_account_note:
    endpoint: PATCH /accountNotes/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update account note; approval required
  delete_account_note:
    endpoint: DELETE /accountNotes/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete account note in Outreach; approval required
  create_account:
    endpoint: POST /accounts
    required fields: data
    risk: external Outreach mutation; create account; approval required
  update_account:
    endpoint: PATCH /accounts/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update account; approval required
  delete_account:
    endpoint: DELETE /accounts/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete account in Outreach; approval required
  add_account_tags:
    endpoint: POST /batches/actions/accountAddTags
    required fields: data
    risk: external Outreach mutation; add account tags; approval required
  add_account_assignments:
    endpoint: POST /batches/actions/accountsAddAssignments
    required fields: data
    risk: external Outreach mutation; add account assignments; approval required
  assign_account_owner:
    endpoint: POST /batches/actions/accountsAssignOwner
    required fields: data
    risk: external Outreach mutation; assign account owner; approval required
  bulk_modify_accounts:
    endpoint: POST /batches/actions/accountsBulkModify
    required fields: data
    risk: external Outreach mutation; bulk modify accounts; approval required
  destroy_all_accounts:
    endpoint: POST /batches/actions/accountsDestroyAll
    required fields: data
    risk: destructive external mutation; destroy all accounts in Outreach; approval required
  remove_all_account_assignments:
    endpoint: POST /batches/actions/accountsRemoveAllAssignments
    required fields: data
    risk: destructive external mutation; remove all account assignments in Outreach; approval required
  remove_account_assignments:
    endpoint: POST /batches/actions/accountsRemoveAssignments
    required fields: data
    risk: external Outreach mutation; remove account assignments; approval required
  remove_account_tags:
    endpoint: POST /batches/actions/accountsRemoveTags
    required fields: data
    risk: external Outreach mutation; remove account tags; approval required
  bulk_delete_custom_objects:
    endpoint: POST /batches/actions/customObjectBulkDelete?actionParams[objectName]={{ record.objectName }}
    required fields: objectName, data
    risk: destructive external mutation; bulk delete custom objects in Outreach; approval required
  bulk_modify_custom_objects:
    endpoint: POST /batches/actions/customObjectBulkModify?actionParams[objectName]={{ record.objectName }}
    required fields: objectName, data
    risk: external Outreach mutation; bulk modify custom objects; approval required
  add_prospect_assignments:
    endpoint: POST /batches/actions/prospectsAddAssignments
    required fields: data
    risk: external Outreach mutation; add prospect assignments; approval required
  add_prospect_tags:
    endpoint: POST /batches/actions/prospectsAddTags
    required fields: data
    risk: external Outreach mutation; add prospect tags; approval required
  add_prospects_to_sequence:
    endpoint: POST /batches/actions/prospectsAddToSequence
    required fields: data
    risk: external Outreach mutation; add prospects to sequence; approval required
  assign_prospect_account:
    endpoint: POST /batches/actions/prospectsAssignAccount
    required fields: data
    risk: external Outreach mutation; assign prospect account; approval required
  assign_prospect_opportunity:
    endpoint: POST /batches/actions/prospectsAssignOpportunity
    required fields: data
    risk: external Outreach mutation; assign prospect opportunity; approval required
  assign_prospect_owner:
    endpoint: POST /batches/actions/prospectsAssignOwner
    required fields: data
    risk: external Outreach mutation; assign prospect owner; approval required
  bulk_modify_prospects:
    endpoint: POST /batches/actions/prospectsBulkModify
    required fields: data
    risk: external Outreach mutation; bulk modify prospects; approval required
  destroy_all_prospects:
    endpoint: POST /batches/actions/prospectsDestroyAll
    required fields: data
    risk: destructive external mutation; destroy all prospects in Outreach; approval required
  finish_all_prospects:
    endpoint: POST /batches/actions/prospectsFinishAll
    required fields: data
    risk: external Outreach mutation; finish all prospects; approval required
  pause_all_prospects:
    endpoint: POST /batches/actions/prospectsPauseAll
    required fields: data
    risk: external Outreach mutation; pause all prospects; approval required
  remove_all_prospect_assignments:
    endpoint: POST /batches/actions/prospectsRemoveAllAssignments
    required fields: data
    risk: destructive external mutation; remove all prospect assignments in Outreach; approval required
  remove_prospect_assignments:
    endpoint: POST /batches/actions/prospectsRemoveAssignments
    required fields: data
    risk: external Outreach mutation; remove prospect assignments; approval required
  remove_prospect_tags:
    endpoint: POST /batches/actions/prospectsRemoveTags
    required fields: data
    risk: external Outreach mutation; remove prospect tags; approval required
  cancel_batch:
    endpoint: POST /batches/{{ record.id }}/actions/cancel
    required fields: id
    risk: external Outreach mutation; cancel batch; approval required
  confirm_batch:
    endpoint: POST /batches/{{ record.id }}/actions/confirm
    required fields: id
    risk: external Outreach mutation; confirm batch; approval required
  create_call_disposition:
    endpoint: POST /callDispositions
    required fields: data
    risk: external Outreach mutation; create call disposition; approval required
  update_call_disposition:
    endpoint: PATCH /callDispositions/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update call disposition; approval required
  delete_call_disposition:
    endpoint: DELETE /callDispositions/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete call disposition in Outreach; approval required
  create_call_purpose:
    endpoint: POST /callPurposes
    required fields: data
    risk: external Outreach mutation; create call purpose; approval required
  update_call_purpose:
    endpoint: PATCH /callPurposes/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update call purpose; approval required
  delete_call_purpose:
    endpoint: DELETE /callPurposes/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete call purpose in Outreach; approval required
  create_call:
    endpoint: POST /calls
    required fields: data
    risk: external Outreach mutation; create call; approval required
  delete_call:
    endpoint: DELETE /calls/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete call in Outreach; approval required
  create_compliance_request:
    endpoint: POST /complianceRequests
    required fields: data
    risk: external Outreach mutation; create compliance request; approval required
  create_content_category:
    endpoint: POST /contentCategories
    required fields: data
    risk: external Outreach mutation; create content category; approval required
  update_content_category:
    endpoint: PATCH /contentCategories/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update content category; approval required
  delete_content_category:
    endpoint: DELETE /contentCategories/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete content category in Outreach; approval required
  create_content_category_membership:
    endpoint: POST /contentCategoryMemberships
    required fields: data
    risk: external Outreach mutation; create content category membership; approval required
  delete_content_category_membership:
    endpoint: DELETE /contentCategoryMemberships/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete content category membership in Outreach; approval required
  create_content_category_ownership:
    endpoint: POST /contentCategoryOwnerships
    required fields: data
    risk: external Outreach mutation; create content category ownership; approval required
  delete_content_category_ownership:
    endpoint: DELETE /contentCategoryOwnerships/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete content category ownership in Outreach; approval required
  create_custom_duty:
    endpoint: POST /customDuties
    required fields: data
    risk: external Outreach mutation; create custom duty; approval required
  create_email_address:
    endpoint: POST /emailAddresses
    required fields: data
    risk: external Outreach mutation; create email address; approval required
  update_email_address:
    endpoint: PATCH /emailAddresses/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update email address; approval required
  delete_email_address:
    endpoint: DELETE /emailAddresses/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete email address in Outreach; approval required
  create_favorite:
    endpoint: POST /favorites
    required fields: data
    risk: external Outreach mutation; create favorite; approval required
  delete_favorite:
    endpoint: DELETE /favorites/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete favorite in Outreach; approval required
  import_accounts:
    endpoint: POST /imports/actions/accountsImport
    required fields: data
    risk: external Outreach mutation; import accounts; approval required
  bulk_upsert_imports:
    endpoint: POST /imports/actions/bulkUpsert
    required fields: data
    risk: external Outreach mutation; bulk upsert imports; approval required
  bulk_upsert_custom_objects:
    endpoint: POST /imports/actions/customObjectBulkUpsert
    required fields: data
    risk: external Outreach mutation; bulk upsert custom objects; approval required
  generate_import_upload_link:
    endpoint: POST /imports/actions/generateUploadLink
    risk: external Outreach mutation; generate import upload link; approval required
  import_prospects:
    endpoint: POST /imports/actions/prospectsImport
    required fields: data
    risk: external Outreach mutation; import prospects; approval required
  validate_import_upload:
    endpoint: POST /imports/actions/validateUpload?actionParams[hash]={{ record.hash }}&actionParams[storageKey]={{ record.storageKey }}
    required fields: hash, storageKey
    risk: external Outreach mutation; validate import upload; approval required
  create_kaia_voice_import:
    endpoint: POST /kaiaVoiceImports
    required fields: data
    risk: external Outreach mutation; create kaia voice import; approval required
  create_mailbox:
    endpoint: POST /mailboxes
    required fields: data
    risk: external Outreach mutation; create mailbox; approval required
  update_mailbox:
    endpoint: PATCH /mailboxes/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update mailbox; approval required
  delete_mailbox:
    endpoint: DELETE /mailboxes/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete mailbox in Outreach; approval required
  link_ews_master_account_mailbox:
    endpoint: POST /mailboxes/{{ record.id }}/actions/linkEwsMasterAccount
    required fields: id
    risk: external Outreach mutation; link ews master account mailbox; approval required
  test_send_mailbox:
    endpoint: POST /mailboxes/{{ record.id }}/actions/testSend
    required fields: id
    risk: external Outreach mutation; test send mailbox; approval required
  test_sync_mailbox:
    endpoint: POST /mailboxes/{{ record.id }}/actions/testSync
    required fields: id
    risk: external Outreach mutation; test sync mailbox; approval required
  unlink_ews_master_account_mailbox:
    endpoint: POST /mailboxes/{{ record.id }}/actions/unlinkEwsMasterAccount
    required fields: id
    risk: external Outreach mutation; unlink ews master account mailbox; approval required
  create_mailing:
    endpoint: POST /mailings
    required fields: data
    risk: external Outreach mutation; create mailing; approval required
  create_opportunity:
    endpoint: POST /opportunities
    required fields: data
    risk: external Outreach mutation; create opportunity; approval required
  update_opportunity:
    endpoint: PATCH /opportunities/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update opportunity; approval required
  delete_opportunity:
    endpoint: DELETE /opportunities/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete opportunity in Outreach; approval required
  create_opportunity_prospect_role:
    endpoint: POST /opportunityProspectRoles
    required fields: data
    risk: external Outreach mutation; create opportunity prospect role; approval required
  update_opportunity_prospect_role:
    endpoint: PATCH /opportunityProspectRoles/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update opportunity prospect role; approval required
  delete_opportunity_prospect_role:
    endpoint: DELETE /opportunityProspectRoles/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete opportunity prospect role in Outreach; approval required
  create_opportunity_stage:
    endpoint: POST /opportunityStages
    required fields: data
    risk: external Outreach mutation; create opportunity stage; approval required
  update_opportunity_stage:
    endpoint: PATCH /opportunityStages/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update opportunity stage; approval required
  delete_opportunity_stage:
    endpoint: DELETE /opportunityStages/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete opportunity stage in Outreach; approval required
  update_org_setting:
    endpoint: PATCH /orgSettings/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update org setting; approval required
  create_persona:
    endpoint: POST /personas
    required fields: data
    risk: external Outreach mutation; create persona; approval required
  update_persona:
    endpoint: PATCH /personas/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update persona; approval required
  delete_persona:
    endpoint: DELETE /personas/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete persona in Outreach; approval required
  create_phone_number:
    endpoint: POST /phoneNumbers
    required fields: data
    risk: external Outreach mutation; create phone number; approval required
  update_phone_number:
    endpoint: PATCH /phoneNumbers/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update phone number; approval required
  delete_phone_number:
    endpoint: DELETE /phoneNumbers/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete phone number in Outreach; approval required
  create_product:
    endpoint: POST /products
    required fields: data
    risk: external Outreach mutation; create product; approval required
  update_product:
    endpoint: PATCH /products/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update product; approval required
  delete_product:
    endpoint: DELETE /products/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete product in Outreach; approval required
  create_profile:
    endpoint: POST /profiles
    required fields: data
    risk: external Outreach mutation; create profile; approval required
  update_profile:
    endpoint: PATCH /profiles/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update profile; approval required
  delete_profile:
    endpoint: DELETE /profiles/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete profile in Outreach; approval required
  create_prospect_note:
    endpoint: POST /prospectNotes
    required fields: data
    risk: external Outreach mutation; create prospect note; approval required
  update_prospect_note:
    endpoint: PATCH /prospectNotes/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update prospect note; approval required
  delete_prospect_note:
    endpoint: DELETE /prospectNotes/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete prospect note in Outreach; approval required
  create_prospect:
    endpoint: POST /prospects
    required fields: data
    risk: external Outreach mutation; create prospect; approval required
  update_prospect:
    endpoint: PATCH /prospects/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update prospect; approval required
  delete_prospect:
    endpoint: DELETE /prospects/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete prospect in Outreach; approval required
  create_purchase:
    endpoint: POST /purchases
    required fields: data
    risk: external Outreach mutation; create purchase; approval required
  update_purchase:
    endpoint: PATCH /purchases/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update purchase; approval required
  delete_purchase:
    endpoint: DELETE /purchases/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete purchase in Outreach; approval required
  create_recipient:
    endpoint: POST /recipients
    required fields: data
    risk: external Outreach mutation; create recipient; approval required
  update_recipient:
    endpoint: PATCH /recipients/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update recipient; approval required
  delete_recipient:
    endpoint: DELETE /recipients/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete recipient in Outreach; approval required
  create_role:
    endpoint: POST /roles
    required fields: data
    risk: external Outreach mutation; create role; approval required
  update_role:
    endpoint: PATCH /roles/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update role; approval required
  delete_role:
    endpoint: DELETE /roles/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete role in Outreach; approval required
  create_ruleset:
    endpoint: POST /rulesets
    required fields: data
    risk: external Outreach mutation; create ruleset; approval required
  update_ruleset:
    endpoint: PATCH /rulesets/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update ruleset; approval required
  delete_ruleset:
    endpoint: DELETE /rulesets/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete ruleset in Outreach; approval required
  create_sequence_state:
    endpoint: POST /sequenceStates
    required fields: data
    risk: external Outreach mutation; create sequence state; approval required
  delete_sequence_state:
    endpoint: DELETE /sequenceStates/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete sequence state in Outreach; approval required
  finish_sequence_state:
    endpoint: POST /sequenceStates/{{ record.id }}/actions/finish
    required fields: id
    risk: external Outreach mutation; finish sequence state; approval required
  pause_sequence_state:
    endpoint: POST /sequenceStates/{{ record.id }}/actions/pause
    required fields: id
    risk: external Outreach mutation; pause sequence state; approval required
  resume_sequence_state:
    endpoint: POST /sequenceStates/{{ record.id }}/actions/resume
    required fields: id
    risk: external Outreach mutation; resume sequence state; approval required
  create_sequence_step:
    endpoint: POST /sequenceSteps
    required fields: data
    risk: external Outreach mutation; create sequence step; approval required
  update_sequence_step:
    endpoint: PATCH /sequenceSteps/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update sequence step; approval required
  create_sequence_template:
    endpoint: POST /sequenceTemplates
    required fields: data
    risk: external Outreach mutation; create sequence template; approval required
  update_sequence_template:
    endpoint: PATCH /sequenceTemplates/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update sequence template; approval required
  delete_sequence_template:
    endpoint: DELETE /sequenceTemplates/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete sequence template in Outreach; approval required
  activate_sequence_template:
    endpoint: POST /sequenceTemplates/{{ record.id }}/actions/activate
    required fields: id
    risk: external Outreach mutation; activate sequence template; approval required
  deactivate_sequence_template:
    endpoint: POST /sequenceTemplates/{{ record.id }}/actions/deactivate
    required fields: id
    risk: external Outreach mutation; deactivate sequence template; approval required
  create_sequence:
    endpoint: POST /sequences
    required fields: data
    risk: external Outreach mutation; create sequence; approval required
  update_sequence:
    endpoint: PATCH /sequences/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update sequence; approval required
  delete_sequence:
    endpoint: DELETE /sequences/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete sequence in Outreach; approval required
  activate_sequence:
    endpoint: POST /sequences/{{ record.id }}/actions/activate
    required fields: id
    risk: external Outreach mutation; activate sequence; approval required
  deactivate_sequence:
    endpoint: POST /sequences/{{ record.id }}/actions/deactivate
    required fields: id
    risk: external Outreach mutation; deactivate sequence; approval required
  create_snippet:
    endpoint: POST /snippets
    required fields: data
    risk: external Outreach mutation; create snippet; approval required
  update_snippet:
    endpoint: PATCH /snippets/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update snippet; approval required
  delete_snippet:
    endpoint: DELETE /snippets/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete snippet in Outreach; approval required
  create_stage:
    endpoint: POST /stages
    required fields: data
    risk: external Outreach mutation; create stage; approval required
  update_stage:
    endpoint: PATCH /stages/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update stage; approval required
  delete_stage:
    endpoint: DELETE /stages/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete stage in Outreach; approval required
  create_task_disposition:
    endpoint: POST /taskDispositions
    required fields: data
    risk: external Outreach mutation; create task disposition; approval required
  update_task_disposition:
    endpoint: PATCH /taskDispositions/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update task disposition; approval required
  delete_task_disposition:
    endpoint: DELETE /taskDispositions/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete task disposition in Outreach; approval required
  create_task_purpose:
    endpoint: POST /taskPurposes
    required fields: data
    risk: external Outreach mutation; create task purpose; approval required
  update_task_purpose:
    endpoint: PATCH /taskPurposes/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update task purpose; approval required
  delete_task_purpose:
    endpoint: DELETE /taskPurposes/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete task purpose in Outreach; approval required
  create_task:
    endpoint: POST /tasks
    required fields: data
    risk: external Outreach mutation; create task; approval required
  update_task:
    endpoint: PATCH /tasks/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update task; approval required
  delete_task:
    endpoint: DELETE /tasks/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete task in Outreach; approval required
  advance_task:
    endpoint: POST /tasks/{{ record.id }}/actions/advance
    required fields: id
    risk: external Outreach mutation; advance task; approval required
  deliver_task:
    endpoint: POST /tasks/{{ record.id }}/actions/deliver
    required fields: id
    risk: external Outreach mutation; deliver task; approval required
  log_meet_in_person_task:
    endpoint: POST /tasks/{{ record.id }}/actions/logMeetInPerson?actionParams[taskDispositionId]={{ record.taskDispositionId }}
    required fields: id, taskDispositionId
    risk: external Outreach mutation; log meet in person task; approval required
  mark_complete_task:
    endpoint: POST /tasks/{{ record.id }}/actions/markComplete
    required fields: id
    risk: external Outreach mutation; mark complete task; approval required
  reassign_owner_task:
    endpoint: POST /tasks/{{ record.id }}/actions/reassignOwner?actionParams[ownerId]={{ record.ownerId }}
    required fields: id, ownerId
    risk: external Outreach mutation; reassign owner task; approval required
  reschedule_task:
    endpoint: POST /tasks/{{ record.id }}/actions/reschedule?actionParams[dueAt]={{ record.dueAt }}
    required fields: id, dueAt
    risk: external Outreach mutation; reschedule task; approval required
  snooze_task:
    endpoint: POST /tasks/{{ record.id }}/actions/snooze
    required fields: id
    risk: external Outreach mutation; snooze task; approval required
  update_note_task:
    endpoint: POST /tasks/{{ record.id }}/actions/updateNote?actionParams[note]={{ record.note }}
    required fields: id, note
    risk: external Outreach mutation; update note task; approval required
  update_opportunity_association_task:
    endpoint: POST /tasks/{{ record.id }}/actions/updateOpportunityAssociation?actionParams[opportunityAssociation]={{ record.opportunityAssociation }}
    required fields: id, opportunityAssociation
    risk: external Outreach mutation; update opportunity association task; approval required
  create_team:
    endpoint: POST /teams
    required fields: data
    risk: external Outreach mutation; create team; approval required
  update_team:
    endpoint: PATCH /teams/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update team; approval required
  delete_team:
    endpoint: DELETE /teams/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete team in Outreach; approval required
  create_template:
    endpoint: POST /templates
    required fields: data
    risk: external Outreach mutation; create template; approval required
  update_template:
    endpoint: PATCH /templates/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update template; approval required
  delete_template:
    endpoint: DELETE /templates/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete template in Outreach; approval required
  create_user:
    endpoint: POST /users
    required fields: data
    risk: external Outreach mutation; create user; approval required
  update_user:
    endpoint: PATCH /users/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update user; approval required
  create_webhook:
    endpoint: POST /webhooks
    required fields: data
    risk: external Outreach mutation; create webhook; approval required
  update_webhook:
    endpoint: PATCH /webhooks/{{ record.id }}
    required fields: id, data
    risk: external Outreach mutation; update webhook; approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}
    required fields: id
    risk: destructive external mutation; delete webhook in Outreach; approval required
  create_custom_object_record:
    endpoint: POST /customObjects/{{ record.objectName }}
    required fields: objectName, data
    risk: external Outreach mutation; creates a record in the configured custom object type; approval required
  update_custom_object_record:
    endpoint: PATCH /customObjects/{{ record.objectName }}/{{ record.id }}
    required fields: objectName, id, data
    risk: external Outreach mutation; updates a record in the configured custom object type; approval required
  delete_custom_object_record:
    endpoint: DELETE /customObjects/{{ record.objectName }}/{{ record.id }}
    required fields: objectName, id
    risk: destructive external mutation; deletes a record from the configured custom object type; approval required

SECURITY
  read risk: external Outreach API read of JSON:API resources including prospects, accounts, sequences, mailings, tasks, calls, opportunities, users, webhooks, and custom objects
  write risk: external Outreach API mutations for documented JSON:API create/update/delete/action endpoints; destructive delete and bulk-destroy actions require approval
  approval: write actions require explicit reverse-ETL approval; destructive deletes and bulk-destroy actions are marked destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect outreach

  # Inspect as structured JSON
  pm connectors inspect outreach --json

AGENT WORKFLOW
  - Run pm connectors inspect outreach before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
