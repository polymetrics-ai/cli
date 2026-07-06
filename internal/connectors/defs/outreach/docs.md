# Overview

Reads and mutates Outreach REST API v2 JSON:API resources, including standard resources and
caller-selected custom objects.

Readable streams: `prospects`, `accounts`, `sequences`, `mailings`, `account_notes`, `account_note`,
`account`, `audit_logs`, `batch_items`, `batch_item`, `batches`, `batch`, `call_dispositions`,
`call_disposition`, `call_purposes`, `call_purpose`, `calls`, `call`, `compliance_requests`,
`compliance_request`, `content_categories`, `content_category`, `content_category_memberships`,
`content_category_membership`, `content_category_ownerships`, `duties`, `email_addresses`,
`email_address`, `events`, `event`, `favorites`, `favorite`, `import`, `kaia_recordings`,
`kaia_recording`, `mail_aliases`, `mail_alias`, `mailboxes`, `mailbox`, `mailing`, `opportunities`,
`opportunity`, `opportunity_prospect_roles`, `opportunity_prospect_role`, `opportunity_stages`,
`opportunity_stage`, `org_setting`, `personas`, `persona`, `phone_numbers`, `phone_number`,
`products`, `product`, `profiles`, `profile`, `prospect_notes`, `prospect_note`, `prospect`,
`purchases`, `purchase`, `recipients`, `recipient`, `roles`, `role`, `rulesets`, `ruleset`,
`sequence_states`, `sequence_state`, `sequence_steps`, `sequence_step`, `sequence_templates`,
`sequence_template`, `sequence`, `snippets`, `snippet`, `stages`, `stage`, `task_dispositions`,
`task_disposition`, `task_priorities`, `task_priority`, `task_purposes`, `task_purpose`, `tasks`,
`task`, `teams`, `team`, `templates`, `template`, `users`, `user`, `webhooks`, `webhook`,
`schema_definitions`, `custom_object_records`, `custom_object_record`.

Write actions: `create_account_note`, `update_account_note`, `delete_account_note`,
`create_account`, `update_account`, `delete_account`, `add_account_tags`, `add_account_assignments`,
`assign_account_owner`, `bulk_modify_accounts`, `destroy_all_accounts`,
`remove_all_account_assignments`, `remove_account_assignments`, `remove_account_tags`,
`bulk_delete_custom_objects`, `bulk_modify_custom_objects`, `add_prospect_assignments`,
`add_prospect_tags`, `add_prospects_to_sequence`, `assign_prospect_account`,
`assign_prospect_opportunity`, `assign_prospect_owner`, `bulk_modify_prospects`,
`destroy_all_prospects`, `finish_all_prospects`, `pause_all_prospects`,
`remove_all_prospect_assignments`, `remove_prospect_assignments`, `remove_prospect_tags`,
`cancel_batch`, `confirm_batch`, `create_call_disposition`, `update_call_disposition`,
`delete_call_disposition`, `create_call_purpose`, `update_call_purpose`, `delete_call_purpose`,
`create_call`, `delete_call`, `create_compliance_request`, `create_content_category`,
`update_content_category`, `delete_content_category`, `create_content_category_membership`,
`delete_content_category_membership`, `create_content_category_ownership`,
`delete_content_category_ownership`, `create_custom_duty`, `create_email_address`,
`update_email_address`, `delete_email_address`, `create_favorite`, `delete_favorite`,
`import_accounts`, `bulk_upsert_imports`, `bulk_upsert_custom_objects`,
`generate_import_upload_link`, `import_prospects`, `validate_import_upload`,
`create_kaia_voice_import`, `create_mailbox`, `update_mailbox`, `delete_mailbox`,
`link_ews_master_account_mailbox`, `test_send_mailbox`, `test_sync_mailbox`,
`unlink_ews_master_account_mailbox`, `create_mailing`, `create_opportunity`, `update_opportunity`,
`delete_opportunity`, `create_opportunity_prospect_role`, `update_opportunity_prospect_role`,
`delete_opportunity_prospect_role`, `create_opportunity_stage`, `update_opportunity_stage`,
`delete_opportunity_stage`, `update_org_setting`, `create_persona`, `update_persona`, and 83 more.

Service API documentation: https://developers.outreach.io/api.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Sent as a Bearer token; never logged. The
  client_id/client_secret/refresh_token exchange is performed outside this connector.
- `account_id` (optional, string); Optional Outreach identifier used by the account detail stream.
- `account_note_id` (optional, string); Optional Outreach identifier used by the account_note detail
  stream.
- `base_url` (optional, string); default `https://api.outreach.io/api/v2`; format `uri`; Outreach
  API base URL override for tests or proxies. Defaults to https://api.outreach.io/api/v2.
- `batch_id` (optional, string); Optional Outreach identifier used by the batch detail stream.
- `batch_item_id` (optional, string); Optional Outreach identifier used by the batch_item detail
  stream.
- `call_disposition_id` (optional, string); Optional Outreach identifier used by the
  call_disposition detail stream.
- `call_id` (optional, string); Optional Outreach identifier used by the call detail stream.
- `call_purpose_id` (optional, string); Optional Outreach identifier used by the call_purpose detail
  stream.
- `compliance_request_id` (optional, string); Optional Outreach identifier used by the
  compliance_request detail stream.
- `content_category_id` (optional, string); Optional Outreach identifier used by the
  content_category detail stream.
- `content_category_membership_id` (optional, string); Optional Outreach identifier used by the
  content_category_membership detail stream.
- `custom_object_name` (optional, string); Optional Outreach identifier used by the
  custom_object_records detail stream.
- `custom_object_record_id` (optional, string); Optional Outreach identifier used by the
  custom_object_record detail stream.
- `email_address_id` (optional, string); Optional Outreach identifier used by the email_address
  detail stream.
- `event_id` (optional, string); Optional Outreach identifier used by the event detail stream.
- `favorite_id` (optional, string); Optional Outreach identifier used by the favorite detail stream.
- `import_id` (optional, string); Optional Outreach identifier used by the import detail stream.
- `kaia_recording_id` (optional, string); Optional Outreach identifier used by the kaia_recording
  detail stream.
- `mail_alias_id` (optional, string); Optional Outreach identifier used by the mail_alias detail
  stream.
- `mailbox_id` (optional, string); Optional Outreach identifier used by the mailbox detail stream.
- `mailing_id` (optional, string); Optional Outreach identifier used by the mailing detail stream.
- `mode` (optional, string).
- `opportunity_id` (optional, string); Optional Outreach identifier used by the opportunity detail
  stream.
- `opportunity_prospect_role_id` (optional, string); Optional Outreach identifier used by the
  opportunity_prospect_role detail stream.
- `opportunity_stage_id` (optional, string); Optional Outreach identifier used by the
  opportunity_stage detail stream.
- `org_setting_id` (optional, string); Optional Outreach identifier used by the org_setting detail
  stream.
- `persona_id` (optional, string); Optional Outreach identifier used by the persona detail stream.
- `phone_number_id` (optional, string); Optional Outreach identifier used by the phone_number detail
  stream.
- `product_id` (optional, string); Optional Outreach identifier used by the product detail stream.
- `profile_id` (optional, string); Optional Outreach identifier used by the profile detail stream.
- `prospect_id` (optional, string); Optional Outreach identifier used by the prospect detail stream.
- `prospect_note_id` (optional, string); Optional Outreach identifier used by the prospect_note
  detail stream.
- `purchase_id` (optional, string); Optional Outreach identifier used by the purchase detail stream.
- `recipient_id` (optional, string); Optional Outreach identifier used by the recipient detail
  stream.
- `role_id` (optional, string); Optional Outreach identifier used by the role detail stream.
- `ruleset_id` (optional, string); Optional Outreach identifier used by the ruleset detail stream.
- `sequence_id` (optional, string); Optional Outreach identifier used by the sequence detail stream.
- `sequence_state_id` (optional, string); Optional Outreach identifier used by the sequence_state
  detail stream.
- `sequence_step_id` (optional, string); Optional Outreach identifier used by the sequence_step
  detail stream.
- `sequence_template_id` (optional, string); Optional Outreach identifier used by the
  sequence_template detail stream.
- `snippet_id` (optional, string); Optional Outreach identifier used by the snippet detail stream.
- `stage_id` (optional, string); Optional Outreach identifier used by the stage detail stream.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only records updated at
  or after this time are read (sent as filter[updatedAt]).
- `task_disposition_id` (optional, string); Optional Outreach identifier used by the
  task_disposition detail stream.
- `task_id` (optional, string); Optional Outreach identifier used by the task detail stream.
- `task_priority_id` (optional, string); Optional Outreach identifier used by the task_priority
  detail stream.
- `task_purpose_id` (optional, string); Optional Outreach identifier used by the task_purpose detail
  stream.
- `team_id` (optional, string); Optional Outreach identifier used by the team detail stream.
- `template_id` (optional, string); Optional Outreach identifier used by the template detail stream.
- `user_id` (optional, string); Optional Outreach identifier used by the user detail stream.
- `webhook_id` (optional, string); Optional Outreach identifier used by the webhook detail stream.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.outreach.io/api/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/prospects` with query `page[size]`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `links.next`; next URLs
stay on the configured API host.

Pagination by stream: next_url: `prospects`, `accounts`, `sequences`, `mailings`, `account_notes`,
`audit_logs`, `batch_items`, `batches`, `call_dispositions`, `call_purposes`, `calls`,
`compliance_requests`, `content_categories`, `content_category_memberships`,
`content_category_ownerships`, `duties`, `email_addresses`, `events`, `favorites`,
`kaia_recordings`, `mail_aliases`, `mailboxes`, `opportunities`, `opportunity_prospect_roles`,
`opportunity_stages`, `personas`, `phone_numbers`, `products`, `profiles`, `prospect_notes`,
`purchases`, `recipients`, `roles`, `rulesets`, `sequence_states`, `sequence_steps`,
`sequence_templates`, `snippets`, `stages`, `task_dispositions`, `task_priorities`, `task_purposes`,
`tasks`, `teams`, `templates`, `users`, `webhooks`, `custom_object_records`; none: `account_note`,
`account`, `batch_item`, `batch`, `call_disposition`, `call_purpose`, `call`, `compliance_request`,
`content_category`, `content_category_membership`, `email_address`, `event`, `favorite`, `import`,
`kaia_recording`, `mail_alias`, `mailbox`, `mailing`, `opportunity`, `opportunity_prospect_role`,
`opportunity_stage`, `org_setting`, `persona`, `phone_number`, `product`, `profile`,
`prospect_note`, `prospect`, `purchase`, `recipient`, `role`, `ruleset`, `sequence_state`,
`sequence_step`, `sequence_template`, `sequence`, `snippet`, `stage`, `task_disposition`,
`task_priority`, `task_purpose`, `task`, `team`, `template`, `user`, `webhook`,
`schema_definitions`, `custom_object_record`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `prospects`: GET `/prospects` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  incremental cursor `updated_at`; sent as `filter[updatedAt]`; formatted as `rfc3339`; initial
  lower bound from `start_date`; computed output fields `created_at`, `email`, `id`, `name`, `type`,
  `updated_at`.
- `accounts`: GET `/accounts` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  incremental cursor `updated_at`; sent as `filter[updatedAt]`; formatted as `rfc3339`; initial
  lower bound from `start_date`; computed output fields `created_at`, `email`, `id`, `name`, `type`,
  `updated_at`.
- `sequences`: GET `/sequences` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  incremental cursor `updated_at`; sent as `filter[updatedAt]`; formatted as `rfc3339`; initial
  lower bound from `start_date`; computed output fields `created_at`, `email`, `id`, `name`, `type`,
  `updated_at`.
- `mailings`: GET `/mailings` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  incremental cursor `updated_at`; sent as `filter[updatedAt]`; formatted as `rfc3339`; initial
  lower bound from `start_date`; computed output fields `created_at`, `email`, `id`, `name`, `type`,
  `updated_at`.
- `account_notes`: GET `/accountNotes` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `account_note`: GET `/accountNotes/{{ config.account_note_id }}` - single-object response; records
  path `data`; emits passthrough records.
- `account`: GET `/accounts/{{ config.account_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `audit_logs`: GET `/auditLogs` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `batch_items`: GET `/batchItems` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `batch_item`: GET `/batchItems/{{ config.batch_item_id }}` - single-object response; records path
  `data`; emits passthrough records.
- `batches`: GET `/batches` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `batch`: GET `/batches/{{ config.batch_id }}` - single-object response; records path `data`; emits
  passthrough records.
- `call_dispositions`: GET `/callDispositions` - records path `data`; query `page[size]`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `call_disposition`: GET `/callDispositions/{{ config.call_disposition_id }}` - single-object
  response; records path `data`; emits passthrough records.
- `call_purposes`: GET `/callPurposes` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `call_purpose`: GET `/callPurposes/{{ config.call_purpose_id }}` - single-object response; records
  path `data`; emits passthrough records.
- `calls`: GET `/calls` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `call`: GET `/calls/{{ config.call_id }}` - single-object response; records path `data`; emits
  passthrough records.
- `compliance_requests`: GET `/complianceRequests` - records path `data`; query `page[size]`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `compliance_request`: GET `/complianceRequests/{{ config.compliance_request_id }}` - single-object
  response; records path `data`; emits passthrough records.
- `content_categories`: GET `/contentCategories` - records path `data`; query `page[size]`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `content_category`: GET `/contentCategories/{{ config.content_category_id }}` - single-object
  response; records path `data`; emits passthrough records.
- `content_category_memberships`: GET `/contentCategoryMemberships` - records path `data`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host; emits passthrough records.
- `content_category_membership`: GET `/contentCategoryMemberships/{{
  config.content_category_membership_id }}` - single-object response; records path `data`; emits
  passthrough records.
- `content_category_ownerships`: GET `/contentCategoryOwnerships` - records path `data`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host; emits passthrough records.
- `duties`: GET `/duties` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `email_addresses`: GET `/emailAddresses` - records path `data`; query `page[size]`=`100`; follows
  a next-page URL from the response body; URL path `links.next`; next URLs stay on the configured
  API host; emits passthrough records.
- `email_address`: GET `/emailAddresses/{{ config.email_address_id }}` - single-object response;
  records path `data`; emits passthrough records.
- `events`: GET `/events` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `event`: GET `/events/{{ config.event_id }}` - single-object response; records path `data`; emits
  passthrough records.
- `favorites`: GET `/favorites` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `favorite`: GET `/favorites/{{ config.favorite_id }}` - single-object response; records path
  `data`; emits passthrough records.
- `import`: GET `/imports/{{ config.import_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `kaia_recordings`: GET `/kaiaRecordings` - records path `data`; query `page[size]`=`100`; follows
  a next-page URL from the response body; URL path `links.next`; next URLs stay on the configured
  API host; emits passthrough records.
- `kaia_recording`: GET `/kaiaRecordings/{{ config.kaia_recording_id }}` - single-object response;
  records path `data`; emits passthrough records.
- `mail_aliases`: GET `/mailAliases` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `mail_alias`: GET `/mailAliases/{{ config.mail_alias_id }}` - single-object response; records path
  `data`; emits passthrough records.
- `mailboxes`: GET `/mailboxes` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `mailbox`: GET `/mailboxes/{{ config.mailbox_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `mailing`: GET `/mailings/{{ config.mailing_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `opportunities`: GET `/opportunities` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `opportunity`: GET `/opportunities/{{ config.opportunity_id }}` - single-object response; records
  path `data`; emits passthrough records.
- `opportunity_prospect_roles`: GET `/opportunityProspectRoles` - records path `data`; query
  `page[size]`=`100`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host; emits passthrough records.
- `opportunity_prospect_role`: GET `/opportunityProspectRoles/{{ config.opportunity_prospect_role_id
  }}` - single-object response; records path `data`; emits passthrough records.
- `opportunity_stages`: GET `/opportunityStages` - records path `data`; query `page[size]`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `opportunity_stage`: GET `/opportunityStages/{{ config.opportunity_stage_id }}` - single-object
  response; records path `data`; emits passthrough records.
- `org_setting`: GET `/orgSettings/{{ config.org_setting_id }}` - single-object response; records
  path `data`; emits passthrough records.
- `personas`: GET `/personas` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `persona`: GET `/personas/{{ config.persona_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `phone_numbers`: GET `/phoneNumbers` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `phone_number`: GET `/phoneNumbers/{{ config.phone_number_id }}` - single-object response; records
  path `data`; emits passthrough records.
- `products`: GET `/products` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `product`: GET `/products/{{ config.product_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `profiles`: GET `/profiles` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `profile`: GET `/profiles/{{ config.profile_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `prospect_notes`: GET `/prospectNotes` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `prospect_note`: GET `/prospectNotes/{{ config.prospect_note_id }}` - single-object response;
  records path `data`; emits passthrough records.
- `prospect`: GET `/prospects/{{ config.prospect_id }}` - single-object response; records path
  `data`; emits passthrough records.
- `purchases`: GET `/purchases` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `purchase`: GET `/purchases/{{ config.purchase_id }}` - single-object response; records path
  `data`; emits passthrough records.
- `recipients`: GET `/recipients` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `recipient`: GET `/recipients/{{ config.recipient_id }}` - single-object response; records path
  `data`; emits passthrough records.
- `roles`: GET `/roles` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `role`: GET `/roles/{{ config.role_id }}` - single-object response; records path `data`; emits
  passthrough records.
- `rulesets`: GET `/rulesets` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `ruleset`: GET `/rulesets/{{ config.ruleset_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `sequence_states`: GET `/sequenceStates` - records path `data`; query `page[size]`=`100`; follows
  a next-page URL from the response body; URL path `links.next`; next URLs stay on the configured
  API host; emits passthrough records.
- `sequence_state`: GET `/sequenceStates/{{ config.sequence_state_id }}` - single-object response;
  records path `data`; emits passthrough records.
- `sequence_steps`: GET `/sequenceSteps` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `sequence_step`: GET `/sequenceSteps/{{ config.sequence_step_id }}` - single-object response;
  records path `data`; emits passthrough records.
- `sequence_templates`: GET `/sequenceTemplates` - records path `data`; query `page[size]`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `sequence_template`: GET `/sequenceTemplates/{{ config.sequence_template_id }}` - single-object
  response; records path `data`; emits passthrough records.
- `sequence`: GET `/sequences/{{ config.sequence_id }}` - single-object response; records path
  `data`; emits passthrough records.
- `snippets`: GET `/snippets` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `snippet`: GET `/snippets/{{ config.snippet_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `stages`: GET `/stages` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `stage`: GET `/stages/{{ config.stage_id }}` - single-object response; records path `data`; emits
  passthrough records.
- `task_dispositions`: GET `/taskDispositions` - records path `data`; query `page[size]`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `task_disposition`: GET `/taskDispositions/{{ config.task_disposition_id }}` - single-object
  response; records path `data`; emits passthrough records.
- `task_priorities`: GET `/taskPriorities` - records path `data`; query `page[size]`=`100`; follows
  a next-page URL from the response body; URL path `links.next`; next URLs stay on the configured
  API host; emits passthrough records.
- `task_priority`: GET `/taskPriorities/{{ config.task_priority_id }}` - single-object response;
  records path `data`; emits passthrough records.
- `task_purposes`: GET `/taskPurposes` - records path `data`; query `page[size]`=`100`; follows a
  next-page URL from the response body; URL path `links.next`; next URLs stay on the configured API
  host; emits passthrough records.
- `task_purpose`: GET `/taskPurposes/{{ config.task_purpose_id }}` - single-object response; records
  path `data`; emits passthrough records.
- `tasks`: GET `/tasks` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `task`: GET `/tasks/{{ config.task_id }}` - single-object response; records path `data`; emits
  passthrough records.
- `teams`: GET `/teams` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `team`: GET `/teams/{{ config.team_id }}` - single-object response; records path `data`; emits
  passthrough records.
- `templates`: GET `/templates` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `template`: GET `/templates/{{ config.template_id }}` - single-object response; records path
  `data`; emits passthrough records.
- `users`: GET `/users` - records path `data`; query `page[size]`=`100`; follows a next-page URL
  from the response body; URL path `links.next`; next URLs stay on the configured API host; emits
  passthrough records.
- `user`: GET `/users/{{ config.user_id }}` - single-object response; records path `data`; emits
  passthrough records.
- `webhooks`: GET `/webhooks` - records path `data`; query `page[size]`=`100`; follows a next-page
  URL from the response body; URL path `links.next`; next URLs stay on the configured API host;
  emits passthrough records.
- `webhook`: GET `/webhooks/{{ config.webhook_id }}` - single-object response; records path `data`;
  emits passthrough records.
- `schema_definitions`: GET `/schema` - records path `definitions`; flattens keyed objects; key
  field `name`; emits passthrough records.
- `custom_object_records`: GET `/customObjects/{{ config.custom_object_name }}` - records path
  `data`; query `page[size]`=`100`; follows a next-page URL from the response body; URL path
  `links.next`; next URLs stay on the configured API host; emits passthrough records.
- `custom_object_record`: GET `/customObjects/{{ config.custom_object_name }}/{{
  config.custom_object_record_id }}` - single-object response; records path `data`; emits
  passthrough records.

## Write actions & risks

Overall write risk: external Outreach API mutations for documented JSON:API
create/update/delete/action endpoints; destructive delete and bulk-destroy actions require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_account_note`: POST `/accountNotes` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create account note;
  approval required.
- `update_account_note`: PATCH `/accountNotes/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update account note; approval required.
- `delete_account_note`: DELETE `/accountNotes/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation; delete
  account note in Outreach; approval required.
- `create_account`: POST `/accounts` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create account; approval
  required.
- `update_account`: PATCH `/accounts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update account; approval required.
- `delete_account`: DELETE `/accounts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete account
  in Outreach; approval required.
- `add_account_tags`: POST `/batches/actions/accountAddTags` - kind `update`; body type `json`;
  required record fields `data`; accepted fields `data`; risk: external Outreach mutation; add
  account tags; approval required.
- `add_account_assignments`: POST `/batches/actions/accountsAddAssignments` - kind `update`; body
  type `json`; required record fields `data`; accepted fields `data`; risk: external Outreach
  mutation; add account assignments; approval required.
- `assign_account_owner`: POST `/batches/actions/accountsAssignOwner` - kind `update`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  assign account owner; approval required.
- `bulk_modify_accounts`: POST `/batches/actions/accountsBulkModify` - kind `update`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  bulk modify accounts; approval required.
- `destroy_all_accounts`: POST `/batches/actions/accountsDestroyAll` - kind `delete`; body type
  `json`; required record fields `data`; accepted fields `data`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; destroy all
  accounts in Outreach; approval required.
- `remove_all_account_assignments`: POST `/batches/actions/accountsRemoveAllAssignments` - kind
  `update`; body type `json`; required record fields `data`; accepted fields `data`; confirmation
  `destructive`; risk: destructive external mutation; remove all account assignments in Outreach;
  approval required.
- `remove_account_assignments`: POST `/batches/actions/accountsRemoveAssignments` - kind `update`;
  body type `json`; required record fields `data`; accepted fields `data`; risk: external Outreach
  mutation; remove account assignments; approval required.
- `remove_account_tags`: POST `/batches/actions/accountsRemoveTags` - kind `update`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  remove account tags; approval required.
- `bulk_delete_custom_objects`: POST
  `/batches/actions/customObjectBulkDelete?actionParams[objectName]={{ record.objectName }}` - kind
  `delete`; body type `json`; path fields `objectName`; required record fields `objectName`, `data`;
  accepted fields `data`, `objectName`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: destructive external mutation; bulk delete custom objects in
  Outreach; approval required.
- `bulk_modify_custom_objects`: POST
  `/batches/actions/customObjectBulkModify?actionParams[objectName]={{ record.objectName }}` - kind
  `update`; body type `json`; path fields `objectName`; required record fields `objectName`, `data`;
  accepted fields `data`, `objectName`; risk: external Outreach mutation; bulk modify custom
  objects; approval required.
- `add_prospect_assignments`: POST `/batches/actions/prospectsAddAssignments` - kind `update`; body
  type `json`; required record fields `data`; accepted fields `data`; risk: external Outreach
  mutation; add prospect assignments; approval required.
- `add_prospect_tags`: POST `/batches/actions/prospectsAddTags` - kind `update`; body type `json`;
  required record fields `data`; accepted fields `data`; risk: external Outreach mutation; add
  prospect tags; approval required.
- `add_prospects_to_sequence`: POST `/batches/actions/prospectsAddToSequence` - kind `update`; body
  type `json`; required record fields `data`; accepted fields `data`; risk: external Outreach
  mutation; add prospects to sequence; approval required.
- `assign_prospect_account`: POST `/batches/actions/prospectsAssignAccount` - kind `update`; body
  type `json`; required record fields `data`; accepted fields `data`; risk: external Outreach
  mutation; assign prospect account; approval required.
- `assign_prospect_opportunity`: POST `/batches/actions/prospectsAssignOpportunity` - kind `update`;
  body type `json`; required record fields `data`; accepted fields `data`; risk: external Outreach
  mutation; assign prospect opportunity; approval required.
- `assign_prospect_owner`: POST `/batches/actions/prospectsAssignOwner` - kind `update`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  assign prospect owner; approval required.
- `bulk_modify_prospects`: POST `/batches/actions/prospectsBulkModify` - kind `update`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  bulk modify prospects; approval required.
- `destroy_all_prospects`: POST `/batches/actions/prospectsDestroyAll` - kind `delete`; body type
  `json`; required record fields `data`; accepted fields `data`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; destroy all
  prospects in Outreach; approval required.
- `finish_all_prospects`: POST `/batches/actions/prospectsFinishAll` - kind `update`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  finish all prospects; approval required.
- `pause_all_prospects`: POST `/batches/actions/prospectsPauseAll` - kind `update`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  pause all prospects; approval required.
- `remove_all_prospect_assignments`: POST `/batches/actions/prospectsRemoveAllAssignments` - kind
  `update`; body type `json`; required record fields `data`; accepted fields `data`; confirmation
  `destructive`; risk: destructive external mutation; remove all prospect assignments in Outreach;
  approval required.
- `remove_prospect_assignments`: POST `/batches/actions/prospectsRemoveAssignments` - kind `update`;
  body type `json`; required record fields `data`; accepted fields `data`; risk: external Outreach
  mutation; remove prospect assignments; approval required.
- `remove_prospect_tags`: POST `/batches/actions/prospectsRemoveTags` - kind `update`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  remove prospect tags; approval required.
- `cancel_batch`: POST `/batches/{{ record.id }}/actions/cancel` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external Outreach
  mutation; cancel batch; approval required.
- `confirm_batch`: POST `/batches/{{ record.id }}/actions/confirm` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  Outreach mutation; confirm batch; approval required.
- `create_call_disposition`: POST `/callDispositions` - kind `create`; body type `json`; required
  record fields `data`; accepted fields `data`; risk: external Outreach mutation; create call
  disposition; approval required.
- `update_call_disposition`: PATCH `/callDispositions/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update call disposition; approval required.
- `delete_call_disposition`: DELETE `/callDispositions/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; delete call disposition in Outreach; approval required.
- `create_call_purpose`: POST `/callPurposes` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create call purpose;
  approval required.
- `update_call_purpose`: PATCH `/callPurposes/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update call purpose; approval required.
- `delete_call_purpose`: DELETE `/callPurposes/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation; delete
  call purpose in Outreach; approval required.
- `create_call`: POST `/calls` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: external Outreach mutation; create call; approval required.
- `delete_call`: DELETE `/calls/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external mutation; delete call in
  Outreach; approval required.
- `create_compliance_request`: POST `/complianceRequests` - kind `create`; body type `json`;
  required record fields `data`; accepted fields `data`; risk: external Outreach mutation; create
  compliance request; approval required.
- `create_content_category`: POST `/contentCategories` - kind `create`; body type `json`; required
  record fields `data`; accepted fields `data`; risk: external Outreach mutation; create content
  category; approval required.
- `update_content_category`: PATCH `/contentCategories/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update content category; approval required.
- `delete_content_category`: DELETE `/contentCategories/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; delete content category in Outreach; approval required.
- `create_content_category_membership`: POST `/contentCategoryMemberships` - kind `create`; body
  type `json`; required record fields `data`; accepted fields `data`; risk: external Outreach
  mutation; create content category membership; approval required.
- `delete_content_category_membership`: DELETE `/contentCategoryMemberships/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external mutation; delete content category membership in Outreach; approval required.
- `create_content_category_ownership`: POST `/contentCategoryOwnerships` - kind `create`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  create content category ownership; approval required.
- `delete_content_category_ownership`: DELETE `/contentCategoryOwnerships/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external mutation; delete content category ownership in Outreach; approval required.
- `create_custom_duty`: POST `/customDuties` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create custom duty;
  approval required.
- `create_email_address`: POST `/emailAddresses` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create email address;
  approval required.
- `update_email_address`: PATCH `/emailAddresses/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update email address; approval required.
- `delete_email_address`: DELETE `/emailAddresses/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; delete email address in Outreach; approval required.
- `create_favorite`: POST `/favorites` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create favorite; approval
  required.
- `delete_favorite`: DELETE `/favorites/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete favorite
  in Outreach; approval required.
- `import_accounts`: POST `/imports/actions/accountsImport` - kind `update`; body type `json`;
  required record fields `data`; accepted fields `data`; risk: external Outreach mutation; import
  accounts; approval required.
- `bulk_upsert_imports`: POST `/imports/actions/bulkUpsert` - kind `upsert`; body type `json`;
  required record fields `data`; accepted fields `data`; risk: external Outreach mutation; bulk
  upsert imports; approval required.
- `bulk_upsert_custom_objects`: POST `/imports/actions/customObjectBulkUpsert` - kind `upsert`; body
  type `json`; required record fields `data`; accepted fields `data`; risk: external Outreach
  mutation; bulk upsert custom objects; approval required.
- `generate_import_upload_link`: POST `/imports/actions/generateUploadLink` - kind `update`; body
  type `none`; risk: external Outreach mutation; generate import upload link; approval required.
- `import_prospects`: POST `/imports/actions/prospectsImport` - kind `update`; body type `json`;
  required record fields `data`; accepted fields `data`; risk: external Outreach mutation; import
  prospects; approval required.
- `validate_import_upload`: POST `/imports/actions/validateUpload?actionParams[hash]={{ record.hash
  }}&actionParams[storageKey]={{ record.storageKey }}` - kind `update`; body type `none`; path
  fields `hash`, `storageKey`; required record fields `hash`, `storageKey`; accepted fields `hash`,
  `storageKey`; risk: external Outreach mutation; validate import upload; approval required.
- `create_kaia_voice_import`: POST `/kaiaVoiceImports` - kind `create`; body type `json`; required
  record fields `data`; accepted fields `data`; risk: external Outreach mutation; create kaia voice
  import; approval required.
- `create_mailbox`: POST `/mailboxes` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create mailbox; approval
  required.
- `update_mailbox`: PATCH `/mailboxes/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external
  Outreach mutation; update mailbox; approval required.
- `delete_mailbox`: DELETE `/mailboxes/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete mailbox
  in Outreach; approval required.
- `link_ews_master_account_mailbox`: POST `/mailboxes/{{ record.id }}/actions/linkEwsMasterAccount`
  - kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: external Outreach mutation; link ews master account mailbox; approval required.
- `test_send_mailbox`: POST `/mailboxes/{{ record.id }}/actions/testSend` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  Outreach mutation; test send mailbox; approval required.
- `test_sync_mailbox`: POST `/mailboxes/{{ record.id }}/actions/testSync` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  Outreach mutation; test sync mailbox; approval required.
- `unlink_ews_master_account_mailbox`: POST `/mailboxes/{{ record.id
  }}/actions/unlinkEwsMasterAccount` - kind `update`; body type `none`; path fields `id`; required
  record fields `id`; accepted fields `id`; risk: external Outreach mutation; unlink ews master
  account mailbox; approval required.
- `create_mailing`: POST `/mailings` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create mailing; approval
  required.
- `create_opportunity`: POST `/opportunities` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create opportunity;
  approval required.
- `update_opportunity`: PATCH `/opportunities/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update opportunity; approval required.
- `delete_opportunity`: DELETE `/opportunities/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation; delete
  opportunity in Outreach; approval required.
- `create_opportunity_prospect_role`: POST `/opportunityProspectRoles` - kind `create`; body type
  `json`; required record fields `data`; accepted fields `data`; risk: external Outreach mutation;
  create opportunity prospect role; approval required.
- `update_opportunity_prospect_role`: PATCH `/opportunityProspectRoles/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`, `data`; accepted fields
  `data`, `id`; risk: external Outreach mutation; update opportunity prospect role; approval
  required.
- `delete_opportunity_prospect_role`: DELETE `/opportunityProspectRoles/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: destructive
  external mutation; delete opportunity prospect role in Outreach; approval required.
- `create_opportunity_stage`: POST `/opportunityStages` - kind `create`; body type `json`; required
  record fields `data`; accepted fields `data`; risk: external Outreach mutation; create opportunity
  stage; approval required.
- `update_opportunity_stage`: PATCH `/opportunityStages/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update opportunity stage; approval required.
- `delete_opportunity_stage`: DELETE `/opportunityStages/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; delete opportunity stage in Outreach; approval required.
- `update_org_setting`: PATCH `/orgSettings/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external
  Outreach mutation; update org setting; approval required.
- `create_persona`: POST `/personas` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create persona; approval
  required.
- `update_persona`: PATCH `/personas/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update persona; approval required.
- `delete_persona`: DELETE `/personas/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete persona
  in Outreach; approval required.
- `create_phone_number`: POST `/phoneNumbers` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create phone number;
  approval required.
- `update_phone_number`: PATCH `/phoneNumbers/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update phone number; approval required.
- `delete_phone_number`: DELETE `/phoneNumbers/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation; delete
  phone number in Outreach; approval required.
- `create_product`: POST `/products` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create product; approval
  required.
- `update_product`: PATCH `/products/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update product; approval required.
- `delete_product`: DELETE `/products/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete product
  in Outreach; approval required.
- `create_profile`: POST `/profiles` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create profile; approval
  required.
- `update_profile`: PATCH `/profiles/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update profile; approval required.
- `delete_profile`: DELETE `/profiles/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete profile
  in Outreach; approval required.
- `create_prospect_note`: POST `/prospectNotes` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create prospect note;
  approval required.
- `update_prospect_note`: PATCH `/prospectNotes/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update prospect note; approval required.
- `delete_prospect_note`: DELETE `/prospectNotes/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation; delete
  prospect note in Outreach; approval required.
- `create_prospect`: POST `/prospects` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create prospect; approval
  required.
- `update_prospect`: PATCH `/prospects/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external
  Outreach mutation; update prospect; approval required.
- `delete_prospect`: DELETE `/prospects/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete prospect
  in Outreach; approval required.
- `create_purchase`: POST `/purchases` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create purchase; approval
  required.
- `update_purchase`: PATCH `/purchases/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external
  Outreach mutation; update purchase; approval required.
- `delete_purchase`: DELETE `/purchases/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete purchase
  in Outreach; approval required.
- `create_recipient`: POST `/recipients` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create recipient; approval
  required.
- `update_recipient`: PATCH `/recipients/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external
  Outreach mutation; update recipient; approval required.
- `delete_recipient`: DELETE `/recipients/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete
  recipient in Outreach; approval required.
- `create_role`: POST `/roles` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: external Outreach mutation; create role; approval required.
- `update_role`: PATCH `/roles/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update role; approval required.
- `delete_role`: DELETE `/roles/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external mutation; delete role in
  Outreach; approval required.
- `create_ruleset`: POST `/rulesets` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create ruleset; approval
  required.
- `update_ruleset`: PATCH `/rulesets/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update ruleset; approval required.
- `delete_ruleset`: DELETE `/rulesets/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete ruleset
  in Outreach; approval required.
- `create_sequence_state`: POST `/sequenceStates` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create sequence state;
  approval required.
- `delete_sequence_state`: DELETE `/sequenceStates/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; delete sequence state in Outreach; approval required.
- `finish_sequence_state`: POST `/sequenceStates/{{ record.id }}/actions/finish` - kind `update`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  external Outreach mutation; finish sequence state; approval required.
- `pause_sequence_state`: POST `/sequenceStates/{{ record.id }}/actions/pause` - kind `update`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  Outreach mutation; pause sequence state; approval required.
- `resume_sequence_state`: POST `/sequenceStates/{{ record.id }}/actions/resume` - kind `update`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  external Outreach mutation; resume sequence state; approval required.
- `create_sequence_step`: POST `/sequenceSteps` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create sequence step;
  approval required.
- `update_sequence_step`: PATCH `/sequenceSteps/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update sequence step; approval required.
- `create_sequence_template`: POST `/sequenceTemplates` - kind `create`; body type `json`; required
  record fields `data`; accepted fields `data`; risk: external Outreach mutation; create sequence
  template; approval required.
- `update_sequence_template`: PATCH `/sequenceTemplates/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update sequence template; approval required.
- `delete_sequence_template`: DELETE `/sequenceTemplates/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; delete sequence template in Outreach; approval required.
- `activate_sequence_template`: POST `/sequenceTemplates/{{ record.id }}/actions/activate` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: external Outreach mutation; activate sequence template; approval required.
- `deactivate_sequence_template`: POST `/sequenceTemplates/{{ record.id }}/actions/deactivate` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: external Outreach mutation; deactivate sequence template; approval required.
- `create_sequence`: POST `/sequences` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create sequence; approval
  required.
- `update_sequence`: PATCH `/sequences/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external
  Outreach mutation; update sequence; approval required.
- `delete_sequence`: DELETE `/sequences/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete sequence
  in Outreach; approval required.
- `activate_sequence`: POST `/sequences/{{ record.id }}/actions/activate` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  Outreach mutation; activate sequence; approval required.
- `deactivate_sequence`: POST `/sequences/{{ record.id }}/actions/deactivate` - kind `update`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  Outreach mutation; deactivate sequence; approval required.
- `create_snippet`: POST `/snippets` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create snippet; approval
  required.
- `update_snippet`: PATCH `/snippets/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update snippet; approval required.
- `delete_snippet`: DELETE `/snippets/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete snippet
  in Outreach; approval required.
- `create_stage`: POST `/stages` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: external Outreach mutation; create stage; approval required.
- `update_stage`: PATCH `/stages/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update stage; approval required.
- `delete_stage`: DELETE `/stages/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external mutation; delete stage in
  Outreach; approval required.
- `create_task_disposition`: POST `/taskDispositions` - kind `create`; body type `json`; required
  record fields `data`; accepted fields `data`; risk: external Outreach mutation; create task
  disposition; approval required.
- `update_task_disposition`: PATCH `/taskDispositions/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update task disposition; approval required.
- `delete_task_disposition`: DELETE `/taskDispositions/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: destructive external
  mutation; delete task disposition in Outreach; approval required.
- `create_task_purpose`: POST `/taskPurposes` - kind `create`; body type `json`; required record
  fields `data`; accepted fields `data`; risk: external Outreach mutation; create task purpose;
  approval required.
- `update_task_purpose`: PATCH `/taskPurposes/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk:
  external Outreach mutation; update task purpose; approval required.
- `delete_task_purpose`: DELETE `/taskPurposes/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: destructive external mutation; delete
  task purpose in Outreach; approval required.
- `create_task`: POST `/tasks` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: external Outreach mutation; create task; approval required.
- `update_task`: PATCH `/tasks/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update task; approval required.
- `delete_task`: DELETE `/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external mutation; delete task in
  Outreach; approval required.
- `advance_task`: POST `/tasks/{{ record.id }}/actions/advance` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external Outreach
  mutation; advance task; approval required.
- `deliver_task`: POST `/tasks/{{ record.id }}/actions/deliver` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external Outreach
  mutation; deliver task; approval required.
- `log_meet_in_person_task`: POST `/tasks/{{ record.id
  }}/actions/logMeetInPerson?actionParams[taskDispositionId]={{ record.taskDispositionId }}` - kind
  `update`; body type `none`; path fields `id`, `taskDispositionId`; required record fields `id`,
  `taskDispositionId`; accepted fields `id`, `taskDispositionId`; risk: external Outreach mutation;
  log meet in person task; approval required.
- `mark_complete_task`: POST `/tasks/{{ record.id }}/actions/markComplete` - kind `update`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  Outreach mutation; mark complete task; approval required.
- `reassign_owner_task`: POST `/tasks/{{ record.id }}/actions/reassignOwner?actionParams[ownerId]={{
  record.ownerId }}` - kind `update`; body type `none`; path fields `id`, `ownerId`; required record
  fields `id`, `ownerId`; accepted fields `id`, `ownerId`; risk: external Outreach mutation;
  reassign owner task; approval required.
- `reschedule_task`: POST `/tasks/{{ record.id }}/actions/reschedule?actionParams[dueAt]={{
  record.dueAt }}` - kind `update`; body type `none`; path fields `id`, `dueAt`; required record
  fields `id`, `dueAt`; accepted fields `dueAt`, `id`; risk: external Outreach mutation; reschedule
  task; approval required.
- `snooze_task`: POST `/tasks/{{ record.id }}/actions/snooze` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external Outreach
  mutation; snooze task; approval required.
- `update_note_task`: POST `/tasks/{{ record.id }}/actions/updateNote?actionParams[note]={{
  record.note }}` - kind `update`; body type `none`; path fields `id`, `note`; required record
  fields `id`, `note`; accepted fields `id`, `note`; risk: external Outreach mutation; update note
  task; approval required.
- `update_opportunity_association_task`: POST `/tasks/{{ record.id
  }}/actions/updateOpportunityAssociation?actionParams[opportunityAssociation]={{
  record.opportunityAssociation }}` - kind `update`; body type `none`; path fields `id`,
  `opportunityAssociation`; required record fields `id`, `opportunityAssociation`; accepted fields
  `id`, `opportunityAssociation`; risk: external Outreach mutation; update opportunity association
  task; approval required.
- `create_team`: POST `/teams` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: external Outreach mutation; create team; approval required.
- `update_team`: PATCH `/teams/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update team; approval required.
- `delete_team`: DELETE `/teams/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external mutation; delete team in
  Outreach; approval required.
- `create_template`: POST `/templates` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create template; approval
  required.
- `update_template`: PATCH `/templates/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external
  Outreach mutation; update template; approval required.
- `delete_template`: DELETE `/templates/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete template
  in Outreach; approval required.
- `create_user`: POST `/users` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: external Outreach mutation; create user; approval required.
- `update_user`: PATCH `/users/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update user; approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: external Outreach mutation; create webhook; approval
  required.
- `update_webhook`: PATCH `/webhooks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`, `data`; accepted fields `data`, `id`; risk: external Outreach
  mutation; update webhook; approval required.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: destructive external mutation; delete webhook
  in Outreach; approval required.
- `create_custom_object_record`: POST `/customObjects/{{ record.objectName }}` - kind `create`; body
  type `json`; path fields `objectName`; required record fields `objectName`, `data`; accepted
  fields `data`, `objectName`; risk: external Outreach mutation; creates a record in the configured
  custom object type; approval required.
- `update_custom_object_record`: PATCH `/customObjects/{{ record.objectName }}/{{ record.id }}` -
  kind `update`; body type `json`; path fields `objectName`, `id`; required record fields
  `objectName`, `id`, `data`; accepted fields `data`, `id`, `objectName`; risk: external Outreach
  mutation; updates a record in the configured custom object type; approval required.
- `delete_custom_object_record`: DELETE `/customObjects/{{ record.objectName }}/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `objectName`, `id`; required record fields
  `objectName`, `id`; accepted fields `id`, `objectName`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: destructive external mutation; deletes a record
  from the configured custom object type; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 96 stream-backed endpoint group(s), 163 write-backed endpoint group(s).
