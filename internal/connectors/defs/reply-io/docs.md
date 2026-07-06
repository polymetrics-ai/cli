# Overview

Reply.io reads 85 stream(s), and writes through 189 action(s).

Readable streams: `people`, `campaigns`, `tasks`, `email_accounts`, `list_knowledge_bases`,
`get_knowledge_base`, `list_knowledge_base_links`, `list_knowledge_base_documents`,
`list_reply_handlers`, `get_reply_handler`, `list_reengagement_cards`, `get_reengagement_card`,
`list_offers`, `get_offer`, `list_playbooks`, `get_playbook`, `get_contact_accounts`,
`get_contact_account_by_id`, `get_contact_account_contacts`, `get_contact_account_lists`,
`get_contact_account_list_by_id`, `get_background_jobs`, `get_background_job_by_id`, `get_contacts`,
`get_contact_by_id`, `get_contact_sequences`, `get_contact_activities`, `get_contact_statuses`,
`get_blacklist_domain_rules`, `get_blacklist_domain_rule_by_id`, `get_blacklist_email_rules`,
`get_blacklist_email_rule_by_id`, `get_blacklist_email_exception_rules`,
`get_blacklist_email_exception_rule_by_id`, `get_contact_lists`, `get_contact_list_by_id`,
`get_contact_lists_for_contact`, `get_custom_fields`, `get_custom_field_by_id`,
`list_email_accounts`, `get_email_account`, `connect_gmail_account`, `connect_office365_account`,
`list_email_account_tags`, `list_email_templates`, `get_email_template`,
`get_email_template_variables`, `list_email_template_folders`, `get_email_template_folder`,
`list_holiday_calendars`, `get_holiday_calendar`, `list_inbox_threads`, `get_inbox_thread`,
`list_inbox_thread_messages`, `list_inbox_categories`, `get_inbox_category`,
`list_linked_in_accounts`, `get_linked_in_account`, `list_pending_linked_in_accounts`,
`list_schedules`, `get_schedule`, `list_schedule_holiday_calendars`, `get_sequences`,
`get_sequence_by_id`, `get_sequence_contacts`, `get_sequence_contact_by_id`,
`get_sequence_contacts_state`, `get_sequence_email_accounts`, `get_sequence_folders`,
`get_sequence_folder_by_id`, `get_folder_sequences`, `get_sequence_linked_in_accounts`,
`get_sequence_steps`, `get_sequence_step_by_id`, `get_sequence_step_variants`,
`get_condition_properties`, `get_sequence_templates`, `get_settings`, `get_tasks`, `get_task_by_id`,
`whoami`, `list_webhooks`, `get_webhook_events`, `get_webhook_by_id`, `get_webhook_logs`.

Write actions: `create_knowledge_base`, `update_knowledge_base`, `delete_knowledge_base`,
`duplicate_knowledge_base`, `add_knowledge_base_link`, `delete_knowledge_base_link`,
`delete_knowledge_base_document`, `create_reply_handler`, `update_reply_handler`,
`delete_reply_handler`, `delete_reply_handler_media`, `create_reengagement_card`,
`update_reengagement_card`, `delete_reengagement_card`, `delete_reengagement_card_media`,
`create_offer`, `update_offer`, `delete_offer`, `create_playbook`, `update_playbook`,
`delete_playbook`, `delete_playbook_style_file`, `create_contact_account`, `update_contact_account`,
`delete_contact_account`, `bulk_create_contact_accounts`, `bulk_delete_contact_accounts`,
`update_contact_account_owner`, `bulk_add_contacts_to_contact_account`,
`bulk_remove_contacts_from_contact_account`, `create_contact_account_list`,
`update_contact_account_list`, `delete_contact_account_list`,
`move_accounts_to_contact_account_list`, `add_accounts_to_contact_account_list`,
`cancel_background_job`, `create_contact`, `update_contact`, `delete_contact`, `import_contacts`,
`bulk_delete_contacts`, `set_contacts_replied`, `set_contacts_bounced`,
`set_contacts_status_in_sequence`, `change_contacts_owner`, `add_contact_note`,
`move_contact_to_sequence`, `create_blacklist_domain_rule`, `update_blacklist_domain_rule`,
`delete_blacklist_domain_rule`, `bulk_delete_blacklist_domain_rules`, `create_blacklist_email_rule`,
`update_blacklist_email_rule`, `delete_blacklist_email_rule`, `bulk_delete_blacklist_email_rules`,
`create_blacklist_email_exception_rule`, `update_blacklist_email_exception_rule`,
`delete_blacklist_email_exception_rule`, `bulk_delete_blacklist_email_exception_rules`,
`create_contact_list`, `update_contact_list`, `delete_contact_list`, `share_contact_list`,
`unshare_contact_list`, `move_contacts_to_contact_list`, `add_contacts_to_contact_list`,
`create_custom_field`, `update_custom_field`, `delete_custom_field`, `send_direct_email`,
`send_direct_linked_in_connect`, `send_direct_linked_in_in_mail`, `send_direct_linked_in_message`,
`send_direct_linked_in_voice`, `send_direct_linked_in_ai_voice`, `create_email_account`,
`update_email_account`, `delete_email_account`, `set_default_email_account`, `resume_sending`, and
109 more.

Service API documentation: https://docs.reply.io/api-reference/introduction.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Never logged.
- `base_url` (optional, string); default `https://api.reply.io`; format `uri`; Reply.io API base URL
  override for tests or proxies.
- `bearer_token` (optional, secret, string); Optional Reply.io v3 bearer token/API key. When
  provided, v3 streams and writes use Authorization: Bearer.
- `calendarId` (optional, string); Path parameter calendarId for documented Reply.io v3 endpoints.
- `contact_id` (optional, string); Path parameter contact_id for documented Reply.io v3 endpoints.
- `created_after` (optional, string); Optional created-after filter applied to every stream's
  request.
- `document_id` (optional, string); Path parameter document_id for documented Reply.io v3 endpoints.
- `email` (optional, string); Optional email filter applied to every stream's request.
- `email_account_id` (optional, string); Path parameter email_account_id for documented Reply.io v3
  endpoints.
- `id` (optional, string); Path parameter id for documented Reply.io v3 endpoints.
- `jobId` (optional, string); Path parameter jobId for documented Reply.io v3 endpoints.
- `knowledge_base_id` (optional, string); Path parameter knowledge_base_id for documented Reply.io
  v3 endpoints.
- `link_id` (optional, string); Path parameter link_id for documented Reply.io v3 endpoints.
- `linkedInAccountId` (optional, string); Path parameter linkedInAccountId for documented Reply.io
  v3 endpoints.
- `media_id` (optional, string); Path parameter media_id for documented Reply.io v3 endpoints.
- `playbook_id` (optional, string); Path parameter playbook_id for documented Reply.io v3 endpoints.
- `reengagement_card_id` (optional, string); Path parameter reengagement_card_id for documented
  Reply.io v3 endpoints.
- `reply_handler_id` (optional, string); Path parameter reply_handler_id for documented Reply.io v3
  endpoints.
- `sequence_id` (optional, string); Path parameter sequence_id for documented Reply.io v3 endpoints.
- `status` (optional, string); Optional status filter applied to every stream's request.
- `step_id` (optional, string); Path parameter step_id for documented Reply.io v3 endpoints.
- `style_file_id` (optional, string); Path parameter style_file_id for documented Reply.io v3
  endpoints.
- `tagId` (optional, string); Path parameter tagId for documented Reply.io v3 endpoints.
- `updated_after` (optional, string).
- `variant_id` (optional, string); Path parameter variant_id for documented Reply.io v3 endpoints.

Secret fields are redacted in logs and write previews: `api_key`, `bearer_token`.

Default configuration values: `base_url=https://api.reply.io`.

Authentication behavior:

- Bearer token authentication using `secrets.bearer_token` when `{{ secrets.bearer_token }}`.
- API key authentication in `X-Api-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/people` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Pagination by stream: offset_limit: `list_knowledge_bases`, `list_knowledge_base_links`,
`list_knowledge_base_documents`, `list_reply_handlers`, `list_reengagement_cards`, `list_offers`,
`list_playbooks`, `get_contact_accounts`, `get_contact_account_contacts`,
`get_contact_account_lists`, `get_background_jobs`, `get_contacts`, `get_contact_activities`,
`get_blacklist_domain_rules`, `get_blacklist_email_rules`, `get_blacklist_email_exception_rules`,
`get_contact_lists`, `list_email_accounts`, `list_email_templates`, `list_inbox_threads`,
`list_inbox_thread_messages`, `get_sequences`, `get_sequence_contacts`,
`get_sequence_contacts_state`, `get_folder_sequences`, `get_tasks`, `list_webhooks`,
`get_webhook_logs`; page_number: `people`, `campaigns`, `tasks`, `email_accounts`,
`get_knowledge_base`, `get_reply_handler`, `get_reengagement_card`, `get_offer`, `get_playbook`,
`get_contact_account_by_id`, `get_contact_account_list_by_id`, `get_background_job_by_id`,
`get_contact_by_id`, `get_contact_sequences`, `get_contact_statuses`,
`get_blacklist_domain_rule_by_id`, `get_blacklist_email_rule_by_id`,
`get_blacklist_email_exception_rule_by_id`, `get_contact_list_by_id`,
`get_contact_lists_for_contact`, `get_custom_fields`, `get_custom_field_by_id`, `get_email_account`,
`connect_gmail_account`, `connect_office365_account`, `list_email_account_tags`,
`get_email_template`, `get_email_template_variables`, `list_email_template_folders`,
`get_email_template_folder`, `list_holiday_calendars`, `get_holiday_calendar`, `get_inbox_thread`,
`list_inbox_categories`, `get_inbox_category`, `list_linked_in_accounts`, `get_linked_in_account`,
`list_pending_linked_in_accounts`, `list_schedules`, `get_schedule`,
`list_schedule_holiday_calendars`, `get_sequence_by_id`, `get_sequence_contact_by_id`,
`get_sequence_email_accounts`, `get_sequence_folders`, `get_sequence_folder_by_id`,
`get_sequence_linked_in_accounts`, `get_sequence_steps`, `get_sequence_step_by_id`,
`get_sequence_step_variants`, `get_condition_properties`, `get_sequence_templates`, `get_settings`,
`get_task_by_id`, `whoami`, `get_webhook_events`, `get_webhook_by_id`.

- `people`: GET `/v1/people` - records at response root; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `email` from template `{{ config.email }}`, omitted
  when absent; `status` from template `{{ config.status }}`, omitted when absent; `updated_after`
  from template `{{ config.updated_after }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `stream`.
- `campaigns`: GET `/v1/campaigns` - records at response root; query `created_after` from template
  `{{ config.created_after }}`, omitted when absent; `email` from template `{{ config.email }}`,
  omitted when absent; `status` from template `{{ config.status }}`, omitted when absent;
  `updated_after` from template `{{ config.updated_after }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed
  output fields `stream`.
- `tasks`: GET `/v1/tasks` - records at response root; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `email` from template `{{ config.email }}`, omitted
  when absent; `status` from template `{{ config.status }}`, omitted when absent; `updated_after`
  from template `{{ config.updated_after }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; computed output fields
  `stream`.
- `email_accounts`: GET `/v1/emailAccounts` - records at response root; query `created_after` from
  template `{{ config.created_after }}`, omitted when absent; `email` from template `{{ config.email
  }}`, omitted when absent; `status` from template `{{ config.status }}`, omitted when absent;
  `updated_after` from template `{{ config.updated_after }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; computed
  output fields `stream`.
- `list_knowledge_bases`: GET `/v3/ai-sdr/knowledge-bases` - records path `items`; offset/limit
  pagination; offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough
  records.
- `get_knowledge_base`: GET `/v3/ai-sdr/knowledge-bases/{{ config.id }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `list_knowledge_base_links`: GET `/v3/ai-sdr/knowledge-bases/{{ config.knowledge_base_id }}/links`
  - records path `items`; offset/limit pagination; offset parameter `skip`; limit parameter `top`;
  page size 100; emits passthrough records.
- `list_knowledge_base_documents`: GET `/v3/ai-sdr/knowledge-bases/{{ config.knowledge_base_id
  }}/documents` - records path `items`; offset/limit pagination; offset parameter `skip`; limit
  parameter `top`; page size 100; emits passthrough records.
- `list_reply_handlers`: GET `/v3/ai-sdr/knowledge-bases/{{ config.knowledge_base_id
  }}/reply-handlers` - records path `items`; offset/limit pagination; offset parameter `skip`; limit
  parameter `top`; page size 100; emits passthrough records.
- `get_reply_handler`: GET `/v3/ai-sdr/knowledge-bases/{{ config.knowledge_base_id
  }}/reply-handlers/{{ config.reply_handler_id }}` - single-object response; records path `.`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `list_reengagement_cards`: GET `/v3/ai-sdr/knowledge-bases/{{ config.knowledge_base_id
  }}/reengagement-cards` - records path `items`; offset/limit pagination; offset parameter `skip`;
  limit parameter `top`; page size 100; emits passthrough records.
- `get_reengagement_card`: GET `/v3/ai-sdr/knowledge-bases/{{ config.knowledge_base_id
  }}/reengagement-cards/{{ config.reengagement_card_id }}` - single-object response; records path
  `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size
  100; emits passthrough records.
- `list_offers`: GET `/v3/ai-sdr/offers` - records path `items`; offset/limit pagination; offset
  parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_offer`: GET `/v3/ai-sdr/offers/{{ config.id }}` - single-object response; records path `.`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `list_playbooks`: GET `/v3/ai-sdr/playbooks` - records path `items`; offset/limit pagination;
  offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_playbook`: GET `/v3/ai-sdr/playbooks/{{ config.id }}` - single-object response; records path
  `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size
  100; emits passthrough records.
- `get_contact_accounts`: GET `/v3/contact-accounts` - records path `items`; offset/limit
  pagination; offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough
  records.
- `get_contact_account_by_id`: GET `/v3/contact-accounts/{{ config.id }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `get_contact_account_contacts`: GET `/v3/contact-accounts/{{ config.id }}/contacts` - records path
  `items`; offset/limit pagination; offset parameter `skip`; limit parameter `top`; page size 100;
  emits passthrough records.
- `get_contact_account_lists`: GET `/v3/contact-account-lists` - records path `items`; offset/limit
  pagination; offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough
  records.
- `get_contact_account_list_by_id`: GET `/v3/contact-account-lists/{{ config.id }}` - single-object
  response; records path `.`; page-number pagination; page parameter `page`; size parameter `limit`;
  starts at 1; page size 100; emits passthrough records.
- `get_background_jobs`: GET `/v3/background-jobs` - records path `items`; offset/limit pagination;
  offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_background_job_by_id`: GET `/v3/background-jobs/{{ config.jobId }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `get_contacts`: GET `/v3/contacts` - records path `items`; offset/limit pagination; offset
  parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_contact_by_id`: GET `/v3/contacts/{{ config.id }}` - single-object response; records path
  `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size
  100; emits passthrough records.
- `get_contact_sequences`: GET `/v3/contacts/{{ config.id }}/sequences` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `get_contact_activities`: GET `/v3/contacts/{{ config.id }}/activities` - records path `items`;
  offset/limit pagination; offset parameter `skip`; limit parameter `top`; page size 100; emits
  passthrough records.
- `get_contact_statuses`: GET `/v3/contacts/{{ config.id }}/statuses` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `get_blacklist_domain_rules`: GET `/v3/contact-blacklist-rules/domains` - records path `items`;
  offset/limit pagination; offset parameter `skip`; limit parameter `top`; page size 100; emits
  passthrough records.
- `get_blacklist_domain_rule_by_id`: GET `/v3/contact-blacklist-rules/domains/{{ config.id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `get_blacklist_email_rules`: GET `/v3/contact-blacklist-rules/emails` - records path `items`;
  offset/limit pagination; offset parameter `skip`; limit parameter `top`; page size 100; emits
  passthrough records.
- `get_blacklist_email_rule_by_id`: GET `/v3/contact-blacklist-rules/emails/{{ config.id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `get_blacklist_email_exception_rules`: GET `/v3/contact-blacklist-rules/email-exceptions` -
  records path `items`; offset/limit pagination; offset parameter `skip`; limit parameter `top`;
  page size 100; emits passthrough records.
- `get_blacklist_email_exception_rule_by_id`: GET `/v3/contact-blacklist-rules/email-exceptions/{{
  config.id }}` - single-object response; records path `.`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `get_contact_lists`: GET `/v3/contact-lists` - records path `items`; offset/limit pagination;
  offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_contact_list_by_id`: GET `/v3/contact-lists/{{ config.id }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `get_contact_lists_for_contact`: GET `/v3/contacts/{{ config.id }}/lists` - records at response
  root; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; emits passthrough records.
- `get_custom_fields`: GET `/v3/custom-fields` - records at response root; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough
  records.
- `get_custom_field_by_id`: GET `/v3/custom-fields/{{ config.id }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `list_email_accounts`: GET `/v3/email-accounts` - records path `items`; offset/limit pagination;
  offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_email_account`: GET `/v3/email-accounts/{{ config.id }}` - single-object response; records
  path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; emits passthrough records.
- `connect_gmail_account`: GET `/v3/email-accounts/connect/gmail` - single-object response; records
  path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; emits passthrough records.
- `connect_office365_account`: GET `/v3/email-accounts/connect/office-365` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `list_email_account_tags`: GET `/v3/email-accounts/tags` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits
  passthrough records.
- `list_email_templates`: GET `/v3/email-templates` - records path `items`; offset/limit pagination;
  offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_email_template`: GET `/v3/email-templates/{{ config.id }}` - single-object response; records
  path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; emits passthrough records.
- `get_email_template_variables`: GET `/v3/email-templates/variables` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `list_email_template_folders`: GET `/v3/email-template-folders` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `get_email_template_folder`: GET `/v3/email-template-folders/{{ config.id }}` - single-object
  response; records path `.`; page-number pagination; page parameter `page`; size parameter `limit`;
  starts at 1; page size 100; emits passthrough records.
- `list_holiday_calendars`: GET `/v3/holiday-calendars` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits
  passthrough records.
- `get_holiday_calendar`: GET `/v3/holiday-calendars/{{ config.id }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `list_inbox_threads`: GET `/v3/inbox/threads` - records path `items`; offset/limit pagination;
  offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_inbox_thread`: GET `/v3/inbox/threads/{{ config.id }}` - single-object response; records path
  `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size
  100; emits passthrough records.
- `list_inbox_thread_messages`: GET `/v3/inbox/threads/{{ config.id }}/messages` - records path
  `items`; offset/limit pagination; offset parameter `skip`; limit parameter `top`; page size 100;
  emits passthrough records.
- `list_inbox_categories`: GET `/v3/inbox/threads/categories` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `get_inbox_category`: GET `/v3/inbox/threads/categories/{{ config.id }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `list_linked_in_accounts`: GET `/v3/linkedin-accounts` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits
  passthrough records.
- `get_linked_in_account`: GET `/v3/linkedin-accounts/{{ config.id }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `list_pending_linked_in_accounts`: GET `/v3/linkedin-accounts/pending` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `list_schedules`: GET `/v3/schedules` - records at response root; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `get_schedule`: GET `/v3/schedules/{{ config.id }}` - single-object response; records path `.`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `list_schedule_holiday_calendars`: GET `/v3/schedules/{{ config.id }}/holiday-calendars` - records
  at response root; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `get_sequences`: GET `/v3/sequences` - records path `items`; offset/limit pagination; offset
  parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_sequence_by_id`: GET `/v3/sequences/{{ config.id }}` - single-object response; records path
  `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size
  100; emits passthrough records.
- `get_sequence_contacts`: GET `/v3/sequences/{{ config.id }}/contacts` - records path `items`;
  offset/limit pagination; offset parameter `skip`; limit parameter `top`; page size 100; emits
  passthrough records.
- `get_sequence_contact_by_id`: GET `/v3/sequences/{{ config.id }}/contacts/{{ config.contact_id }}`
  - single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `get_sequence_contacts_state`: GET `/v3/sequences/{{ config.id }}/contacts/state` - records path
  `items`; offset/limit pagination; offset parameter `skip`; limit parameter `top`; page size 100;
  emits passthrough records.
- `get_sequence_email_accounts`: GET `/v3/sequences/{{ config.id }}/email-accounts` - records at
  response root; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1;
  page size 100; emits passthrough records.
- `get_sequence_folders`: GET `/v3/sequence-folders` - records at response root; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits
  passthrough records.
- `get_sequence_folder_by_id`: GET `/v3/sequence-folders/{{ config.id }}` - single-object response;
  records path `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `get_folder_sequences`: GET `/v3/sequence-folders/{{ config.id }}/sequences` - records path
  `items`; offset/limit pagination; offset parameter `skip`; limit parameter `top`; page size 100;
  emits passthrough records.
- `get_sequence_linked_in_accounts`: GET `/v3/sequences/{{ config.id }}/linkedin-accounts` - records
  at response root; page-number pagination; page parameter `page`; size parameter `limit`; starts at
  1; page size 100; emits passthrough records.
- `get_sequence_steps`: GET `/v3/sequences/{{ config.id }}/steps` - records at response root;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `get_sequence_step_by_id`: GET `/v3/sequences/{{ config.id }}/steps/{{ config.step_id }}` -
  single-object response; records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `get_sequence_step_variants`: GET `/v3/sequences/{{ config.id }}/steps/{{ config.step_id
  }}/variants` - records at response root; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `get_condition_properties`: GET `/v3/sequences/steps/contact-filter-properties` - single-object
  response; records path `.`; page-number pagination; page parameter `page`; size parameter `limit`;
  starts at 1; page size 100; emits passthrough records.
- `get_sequence_templates`: GET `/v3/sequence-templates` - single-object response; records path `.`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `get_settings`: GET `/v3/settings` - single-object response; records path `.`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits
  passthrough records.
- `get_tasks`: GET `/v3/tasks` - records path `items`; offset/limit pagination; offset parameter
  `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_task_by_id`: GET `/v3/tasks/{{ config.id }}` - single-object response; records path `.`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `whoami`: GET `/v3/whoami` - single-object response; records path `.`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough
  records.
- `list_webhooks`: GET `/v3/webhooks` - records path `items`; offset/limit pagination; offset
  parameter `skip`; limit parameter `top`; page size 100; emits passthrough records.
- `get_webhook_events`: GET `/v3/webhooks/events` - single-object response; records path `.`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `get_webhook_by_id`: GET `/v3/webhooks/{{ config.id }}` - single-object response; records path
  `.`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size
  100; emits passthrough records.
- `get_webhook_logs`: GET `/v3/webhooks/{{ config.id }}/logs` - records path `items`; offset/limit
  pagination; offset parameter `skip`; limit parameter `top`; page size 100; emits passthrough
  records.

## Write actions & risks

Overall write risk: external Reply.io API mutations for contacts, accounts, sequences, templates,
tasks, settings, webhooks, and related v3 resources.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_knowledge_base`: POST `/v3/ai-sdr/knowledge-bases` - kind `create`; body type `json`; body
  fields `name`, `instructions`; accepted fields `instructions`, `name`; risk: POST
  /v3/ai-sdr/knowledge-bases (Create a knowledge base) mutates Reply.io data; review records before
  execution.
- `update_knowledge_base`: PATCH `/v3/ai-sdr/knowledge-bases/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; body fields `name`, `instructions`; required record fields `id`;
  accepted fields `id`, `instructions`, `name`; risk: PATCH /v3/ai-sdr/knowledge-bases/{id} (Update
  a knowledge base) mutates Reply.io data; review records before execution.
- `delete_knowledge_base`: DELETE `/v3/ai-sdr/knowledge-bases/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/ai-sdr/knowledge-bases/{id} (Delete a knowledge base) mutates Reply.io data; review records
  before execution.
- `duplicate_knowledge_base`: POST `/v3/ai-sdr/knowledge-bases/{{ record.id }}/duplicate` - kind
  `create`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: POST /v3/ai-sdr/knowledge-bases/{id}/duplicate (Duplicate a knowledge base) mutates Reply.io
  data; review records before execution.
- `add_knowledge_base_link`: POST `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id }}/links`
  - kind `create`; body type `json`; path fields `knowledge_base_id`; body fields `url`; required
  record fields `knowledge_base_id`; accepted fields `knowledge_base_id`, `url`; risk: POST
  /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/links (Add a link) mutates Reply.io data; review
  records before execution.
- `delete_knowledge_base_link`: DELETE `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/links/{{ record.link_id }}` - kind `delete`; body type `none`; path fields `knowledge_base_id`,
  `link_id`; required record fields `knowledge_base_id`, `link_id`; accepted fields
  `knowledge_base_id`, `link_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/links/{link_id} (Delete
  a link) mutates Reply.io data; review records before execution.
- `delete_knowledge_base_document`: DELETE `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/documents/{{ record.document_id }}` - kind `delete`; body type `none`; path fields
  `knowledge_base_id`, `document_id`; required record fields `knowledge_base_id`, `document_id`;
  accepted fields `document_id`, `knowledge_base_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: DELETE
  /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/documents/{document_id} (Delete a document) mutates
  Reply.io data; review records before execution.
- `create_reply_handler`: POST `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/reply-handlers` - kind `create`; body type `json`; path fields `knowledge_base_id`; body fields
  `typeOfQuestion`, `instructions`, `sampleAnswer`, `toneOfVoice`, `responseLength`, `links`,
  `isAutoSend`; required record fields `knowledge_base_id`; accepted fields `instructions`,
  `isAutoSend`, `knowledge_base_id`, `links`, `responseLength`, `sampleAnswer`, `toneOfVoice`,
  `typeOfQuestion`; risk: POST /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/reply-handlers (Create
  a reply handler) mutates Reply.io data; review records before execution.
- `update_reply_handler`: PATCH `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/reply-handlers/{{ record.reply_handler_id }}` - kind `update`; body type `json`; path fields
  `knowledge_base_id`, `reply_handler_id`; body fields `typeOfQuestion`, `instructions`,
  `sampleAnswer`, `toneOfVoice`, `responseLength`, `links`, `isAutoSend`; required record fields
  `knowledge_base_id`, `reply_handler_id`; accepted fields `instructions`, `isAutoSend`,
  `knowledge_base_id`, `links`, `reply_handler_id`, `responseLength`, `sampleAnswer`, `toneOfVoice`,
  `typeOfQuestion`; risk: PATCH
  /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/reply-handlers/{reply_handler_id} (Update a reply
  handler) mutates Reply.io data; review records before execution.
- `delete_reply_handler`: DELETE `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/reply-handlers/{{ record.reply_handler_id }}` - kind `delete`; body type `none`; path fields
  `knowledge_base_id`, `reply_handler_id`; required record fields `knowledge_base_id`,
  `reply_handler_id`; accepted fields `knowledge_base_id`, `reply_handler_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/reply-handlers/{reply_handler_id} (Delete a reply
  handler) mutates Reply.io data; review records before execution.
- `delete_reply_handler_media`: DELETE `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/reply-handlers/{{ record.reply_handler_id }}/media/{{ record.media_id }}` - kind `delete`; body
  type `none`; path fields `knowledge_base_id`, `reply_handler_id`, `media_id`; required record
  fields `knowledge_base_id`, `reply_handler_id`, `media_id`; accepted fields `knowledge_base_id`,
  `media_id`, `reply_handler_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE
  /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/reply-handlers/{reply_handler_id}/media/{media_id}
  (Delete reply handler media) mutates Reply.io data; review records before execution.
- `create_reengagement_card`: POST `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/reengagement-cards` - kind `create`; body type `json`; path fields `knowledge_base_id`; body
  fields `name`, `instructions`, `sendAfter`, `sampleAnswer`, `toneOfVoice`, `responseLength`,
  `links`, `isEnabled`; required record fields `knowledge_base_id`; accepted fields `instructions`,
  `isEnabled`, `knowledge_base_id`, `links`, `name`, `responseLength`, `sampleAnswer`, `sendAfter`,
  `toneOfVoice`; risk: POST /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/reengagement-cards
  (Create a reengagement card) mutates Reply.io data; review records before execution.
- `update_reengagement_card`: PATCH `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/reengagement-cards/{{ record.reengagement_card_id }}` - kind `update`; body type `json`; path
  fields `knowledge_base_id`, `reengagement_card_id`; body fields `name`, `instructions`,
  `sampleAnswer`, `sendAfter`, `toneOfVoice`, `responseLength`, `links`, `isEnabled`; required
  record fields `knowledge_base_id`, `reengagement_card_id`; accepted fields `instructions`,
  `isEnabled`, `knowledge_base_id`, `links`, `name`, `reengagement_card_id`, `responseLength`,
  `sampleAnswer`, `sendAfter`, `toneOfVoice`; risk: PATCH
  /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/reengagement-cards/{reengagement_card_id} (Update a
  reengagement card) mutates Reply.io data; review records before execution.
- `delete_reengagement_card`: DELETE `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/reengagement-cards/{{ record.reengagement_card_id }}` - kind `delete`; body type `none`; path
  fields `knowledge_base_id`, `reengagement_card_id`; required record fields `knowledge_base_id`,
  `reengagement_card_id`; accepted fields `knowledge_base_id`, `reengagement_card_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/reengagement-cards/{reengagement_card_id} (Delete a
  reengagement card) mutates Reply.io data; review records before execution.
- `delete_reengagement_card_media`: DELETE `/v3/ai-sdr/knowledge-bases/{{ record.knowledge_base_id
  }}/reengagement-cards/{{ record.reengagement_card_id }}/media/{{ record.media_id }}` - kind
  `delete`; body type `none`; path fields `knowledge_base_id`, `reengagement_card_id`, `media_id`;
  required record fields `knowledge_base_id`, `reengagement_card_id`, `media_id`; accepted fields
  `knowledge_base_id`, `media_id`, `reengagement_card_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: DELETE
  /v3/ai-sdr/knowledge-bases/{knowledge_base_id}/reengagement-cards/{reengagement_card_id}/media/{media_id}
  (Delete reengagement card media) mutates Reply.io data; review records before execution.
- `create_offer`: POST `/v3/ai-sdr/offers` - kind `create`; body type `json`; body fields `name`,
  `companyName`, `companyDescription`, `icp`, `reasonForOutreach`, `caseStudies`, `painPoints`,
  `proofPoints`; accepted fields `caseStudies`, `companyDescription`, `companyName`, `icp`, `name`,
  `painPoints`, `proofPoints`, `reasonForOutreach`; risk: POST /v3/ai-sdr/offers (Create an offer)
  mutates Reply.io data; review records before execution.
- `update_offer`: PATCH `/v3/ai-sdr/offers/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `name`, `companyName`, `companyDescription`, `icp`, `reasonForOutreach`,
  `caseStudies`, `painPoints`, `proofPoints`; required record fields `id`; accepted fields
  `caseStudies`, `companyDescription`, `companyName`, `icp`, `id`, `name`, `painPoints`,
  `proofPoints`, `reasonForOutreach`; risk: PATCH /v3/ai-sdr/offers/{id} (Update an offer) mutates
  Reply.io data; review records before execution.
- `delete_offer`: DELETE `/v3/ai-sdr/offers/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: DELETE /v3/ai-sdr/offers/{id} (Delete an
  offer) mutates Reply.io data; review records before execution.
- `create_playbook`: POST `/v3/ai-sdr/playbooks` - kind `create`; body type `json`; body fields
  `name`, `description`, `body`, `type`; accepted fields `body`, `description`, `name`, `type`;
  risk: POST /v3/ai-sdr/playbooks (Create a playbook) mutates Reply.io data; review records before
  execution.
- `update_playbook`: PATCH `/v3/ai-sdr/playbooks/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; body fields `name`, `description`, `body`; required record fields `id`; accepted
  fields `body`, `description`, `id`, `name`; risk: PATCH /v3/ai-sdr/playbooks/{id} (Update a
  playbook) mutates Reply.io data; review records before execution.
- `delete_playbook`: DELETE `/v3/ai-sdr/playbooks/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/ai-sdr/playbooks/{id} (Delete a playbook) mutates Reply.io data; review records before
  execution.
- `delete_playbook_style_file`: DELETE `/v3/ai-sdr/playbooks/{{ record.playbook_id }}/style-files/{{
  record.style_file_id }}` - kind `delete`; body type `none`; path fields `playbook_id`,
  `style_file_id`; required record fields `playbook_id`, `style_file_id`; accepted fields
  `playbook_id`, `style_file_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /v3/ai-sdr/playbooks/{playbook_id}/style-files/{style_file_id} (Delete
  a style file) mutates Reply.io data; review records before execution.
- `create_contact_account`: POST `/v3/contact-accounts` - kind `create`; body type `json`; body
  fields `name`, `description`, `domainName`, `domainSecondary`, `industry`, `companySize`,
  `country`, `state`; accepted fields `companySize`, `country`, `description`, `domainName`,
  `domainSecondary`, `industry`, `name`, `state`; risk: POST /v3/contact-accounts (Create an
  account) mutates Reply.io data; review records before execution.
- `update_contact_account`: PUT `/v3/contact-accounts/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; body fields `name`, `description`, `domainName`, `domainSecondary`,
  `industry`, `companySize`, `country`, `state`; required record fields `id`; accepted fields
  `companySize`, `country`, `description`, `domainName`, `domainSecondary`, `id`, `industry`,
  `name`, `state`; risk: PUT /v3/contact-accounts/{id} (Update an account) mutates Reply.io data;
  review records before execution.
- `delete_contact_account`: DELETE `/v3/contact-accounts/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/contact-accounts/{id} (Delete an account) mutates Reply.io data; review records before
  execution.
- `bulk_create_contact_accounts`: POST `/v3/contact-accounts/bulk` - kind `create`; body type
  `json`; body fields `items`; accepted fields `items`; risk: POST /v3/contact-accounts/bulk (Bulk
  create accounts) mutates Reply.io data; review records before execution.
- `bulk_delete_contact_accounts`: POST `/v3/contact-accounts/bulk-delete` - kind `custom`; body type
  `json`; body fields `ids`; accepted fields `ids`; risk: POST /v3/contact-accounts/bulk-delete
  (Bulk delete accounts) mutates Reply.io data; review records before execution.
- `update_contact_account_owner`: PUT `/v3/contact-accounts/{{ record.id }}/owner` - kind `update`;
  body type `json`; path fields `id`; body fields `userId`; required record fields `id`; accepted
  fields `id`, `userId`; risk: PUT /v3/contact-accounts/{id}/owner (Update account owner) mutates
  Reply.io data; review records before execution.
- `bulk_add_contacts_to_contact_account`: POST `/v3/contact-accounts/{{ record.id
  }}/contact-links/bulk` - kind `create`; body type `json`; path fields `id`; body fields
  `contactIds`; required record fields `id`; accepted fields `contactIds`, `id`; risk: POST
  /v3/contact-accounts/{id}/contact-links/bulk (Bulk add contacts to an account) mutates Reply.io
  data; review records before execution.
- `bulk_remove_contacts_from_contact_account`: POST `/v3/contact-accounts/{{ record.id
  }}/contact-links/bulk-delete` - kind `update`; body type `json`; path fields `id`; body fields
  `contactIds`; required record fields `id`; accepted fields `contactIds`, `id`; risk: POST
  /v3/contact-accounts/{id}/contact-links/bulk-delete (Bulk remove contacts from an account) mutates
  Reply.io data; review records before execution.
- `create_contact_account_list`: POST `/v3/contact-account-lists` - kind `create`; body type `json`;
  body fields `name`; accepted fields `name`; risk: POST /v3/contact-account-lists (Create an
  account list) mutates Reply.io data; review records before execution.
- `update_contact_account_list`: PUT `/v3/contact-account-lists/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; body fields `name`; required record fields `id`; accepted
  fields `id`, `name`; risk: PUT /v3/contact-account-lists/{id} (Update an account list) mutates
  Reply.io data; review records before execution.
- `delete_contact_account_list`: DELETE `/v3/contact-account-lists/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/contact-account-lists/{id} (Delete an account list) mutates Reply.io data; review records
  before execution.
- `move_accounts_to_contact_account_list`: POST `/v3/contact-account-lists/{{ record.id
  }}/move-accounts` - kind `update`; body type `json`; path fields `id`; body fields `accountIds`;
  required record fields `id`; accepted fields `accountIds`, `id`; risk: POST
  /v3/contact-account-lists/{id}/move-accounts (Move accounts to an account list) mutates Reply.io
  data; review records before execution.
- `add_accounts_to_contact_account_list`: POST `/v3/contact-account-lists/{{ record.id
  }}/add-accounts` - kind `create`; body type `json`; path fields `id`; body fields `accountIds`;
  required record fields `id`; accepted fields `accountIds`, `id`; risk: POST
  /v3/contact-account-lists/{id}/add-accounts (Add accounts to an account list) mutates Reply.io
  data; review records before execution.
- `cancel_background_job`: POST `/v3/background-jobs/{{ record.jobId }}/cancel` - kind `custom`;
  body type `json`; path fields `jobId`; body fields `reason`; required record fields `jobId`;
  accepted fields `jobId`, `reason`; risk: POST /v3/background-jobs/{jobId}/cancel (Cancel a
  background job) mutates Reply.io data; review records before execution.
- `create_contact`: POST `/v3/contacts` - kind `create`; body type `json`; body fields `email`,
  `firstName`, `lastName`, `phone`, `phone2`, `title`, `company`, `companySize`; accepted fields
  `company`, `companySize`, `email`, `firstName`, `lastName`, `phone`, `phone2`, `title`; risk: POST
  /v3/contacts (Create a contact) mutates Reply.io data; review records before execution.
- `update_contact`: PATCH `/v3/contacts/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `email`, `firstName`, `lastName`, `phone`, `phone2`, `title`, `company`,
  `companySize`; required record fields `id`; accepted fields `company`, `companySize`, `email`,
  `firstName`, `id`, `lastName`, `phone`, `phone2`, `title`; risk: PATCH /v3/contacts/{id} (Update a
  contact) mutates Reply.io data; review records before execution.
- `delete_contact`: DELETE `/v3/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: DELETE /v3/contacts/{id} (Delete a contact)
  mutates Reply.io data; review records before execution.
- `import_contacts`: POST `/v3/contacts/import` - kind `create`; body type `json`; body fields
  `items`, `options`; accepted fields `items`, `options`; risk: POST /v3/contacts/import (Import
  contacts) mutates Reply.io data; review records before execution.
- `bulk_delete_contacts`: POST `/v3/contacts/bulk-delete` - kind `custom`; body type `json`; body
  fields `ids`; accepted fields `ids`; risk: POST /v3/contacts/bulk-delete (Bulk delete contacts)
  mutates Reply.io data; review records before execution.
- `set_contacts_replied`: POST `/v3/contacts/set-replied` - kind `update`; body type `json`; body
  fields `contactIds`, `isReplied`; accepted fields `contactIds`, `isReplied`; risk: POST
  /v3/contacts/set-replied (Mark or unmark contacts as replied) mutates Reply.io data; review
  records before execution.
- `set_contacts_bounced`: POST `/v3/contacts/set-bounced` - kind `update`; body type `json`; body
  fields `contactIds`, `isBounced`, `resendEmails`; accepted fields `contactIds`, `isBounced`,
  `resendEmails`; risk: POST /v3/contacts/set-bounced (Mark or unmark contacts as bounced) mutates
  Reply.io data; review records before execution.
- `set_contacts_status_in_sequence`: POST `/v3/contacts/set-status-in-sequence` - kind `update`;
  body type `json`; body fields `contactIds`, `statusInSequence`; accepted fields `contactIds`,
  `statusInSequence`; risk: POST /v3/contacts/set-status-in-sequence (Set contacts' in-sequence
  status) mutates Reply.io data; review records before execution.
- `change_contacts_owner`: PUT `/v3/contacts/owner` - kind `update`; body type `json`; body fields
  `contactIds`, `userId`, `reassignTasks`; accepted fields `contactIds`, `reassignTasks`, `userId`;
  risk: PUT /v3/contacts/owner (Change contacts owner) mutates Reply.io data; review records before
  execution.
- `add_contact_note`: POST `/v3/contacts/{{ record.id }}/notes` - kind `create`; body type `json`;
  path fields `id`; body fields `notes`; required record fields `id`; accepted fields `id`, `notes`;
  risk: POST /v3/contacts/{id}/notes (Add a note to a contact) mutates Reply.io data; review records
  before execution.
- `move_contact_to_sequence`: POST `/v3/contacts/{{ record.id }}/move-to-sequence` - kind `update`;
  body type `json`; path fields `id`; body fields `sequenceId`, `removeFromExisting`, `startStepId`,
  `ignoreStepDelay`, `startFrom`; required record fields `id`; accepted fields `id`,
  `ignoreStepDelay`, `removeFromExisting`, `sequenceId`, `startFrom`, `startStepId`; risk: POST
  /v3/contacts/{id}/move-to-sequence (Move a contact to a sequence) mutates Reply.io data; review
  records before execution.
- `create_blacklist_domain_rule`: POST `/v3/contact-blacklist-rules/domains` - kind `create`; body
  type `json`; body fields `pattern`; accepted fields `pattern`; risk: POST
  /v3/contact-blacklist-rules/domains (Create a domain blacklist rule) mutates Reply.io data; review
  records before execution.
- `update_blacklist_domain_rule`: PUT `/v3/contact-blacklist-rules/domains/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; body fields `pattern`; required record fields `id`;
  accepted fields `id`, `pattern`; risk: PUT /v3/contact-blacklist-rules/domains/{id} (Update a
  domain blacklist rule) mutates Reply.io data; review records before execution.
- `delete_blacklist_domain_rule`: DELETE `/v3/contact-blacklist-rules/domains/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  DELETE /v3/contact-blacklist-rules/domains/{id} (Delete a domain blacklist rule) mutates Reply.io
  data; review records before execution.
- `bulk_delete_blacklist_domain_rules`: POST `/v3/contact-blacklist-rules/domains/bulk-delete` -
  kind `custom`; body type `json`; body fields `ids`; accepted fields `ids`; risk: POST
  /v3/contact-blacklist-rules/domains/bulk-delete (Bulk delete domain blacklist rules) mutates
  Reply.io data; review records before execution.
- `create_blacklist_email_rule`: POST `/v3/contact-blacklist-rules/emails` - kind `create`; body
  type `json`; body fields `pattern`; accepted fields `pattern`; risk: POST
  /v3/contact-blacklist-rules/emails (Create an email blacklist rule) mutates Reply.io data; review
  records before execution.
- `update_blacklist_email_rule`: PUT `/v3/contact-blacklist-rules/emails/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; body fields `pattern`; required record fields `id`;
  accepted fields `id`, `pattern`; risk: PUT /v3/contact-blacklist-rules/emails/{id} (Update an
  email blacklist rule) mutates Reply.io data; review records before execution.
- `delete_blacklist_email_rule`: DELETE `/v3/contact-blacklist-rules/emails/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/contact-blacklist-rules/emails/{id} (Delete an email blacklist rule) mutates Reply.io data;
  review records before execution.
- `bulk_delete_blacklist_email_rules`: POST `/v3/contact-blacklist-rules/emails/bulk-delete` - kind
  `custom`; body type `json`; body fields `ids`; accepted fields `ids`; risk: POST
  /v3/contact-blacklist-rules/emails/bulk-delete (Bulk delete email blacklist rules) mutates
  Reply.io data; review records before execution.
- `create_blacklist_email_exception_rule`: POST `/v3/contact-blacklist-rules/email-exceptions` -
  kind `create`; body type `json`; body fields `pattern`; accepted fields `pattern`; risk: POST
  /v3/contact-blacklist-rules/email-exceptions (Create an email exception blacklist rule) mutates
  Reply.io data; review records before execution.
- `update_blacklist_email_exception_rule`: PUT `/v3/contact-blacklist-rules/email-exceptions/{{
  record.id }}` - kind `update`; body type `json`; path fields `id`; body fields `pattern`; required
  record fields `id`; accepted fields `id`, `pattern`; risk: PUT
  /v3/contact-blacklist-rules/email-exceptions/{id} (Update an email exception blacklist rule)
  mutates Reply.io data; review records before execution.
- `delete_blacklist_email_exception_rule`: DELETE `/v3/contact-blacklist-rules/email-exceptions/{{
  record.id }}` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /v3/contact-blacklist-rules/email-exceptions/{id} (Delete an email
  exception blacklist rule) mutates Reply.io data; review records before execution.
- `bulk_delete_blacklist_email_exception_rules`: POST
  `/v3/contact-blacklist-rules/email-exceptions/bulk-delete` - kind `custom`; body type `json`; body
  fields `ids`; accepted fields `ids`; risk: POST
  /v3/contact-blacklist-rules/email-exceptions/bulk-delete (Bulk delete email exception blacklist
  rules) mutates Reply.io data; review records before execution.
- `create_contact_list`: POST `/v3/contact-lists` - kind `create`; body type `json`; body fields
  `name`, `isShared`; accepted fields `isShared`, `name`; risk: POST /v3/contact-lists (Create a
  contact list) mutates Reply.io data; review records before execution.
- `update_contact_list`: PUT `/v3/contact-lists/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; body fields `name`; required record fields `id`; accepted fields `id`, `name`;
  risk: PUT /v3/contact-lists/{id} (Update a contact list) mutates Reply.io data; review records
  before execution.
- `delete_contact_list`: DELETE `/v3/contact-lists/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/contact-lists/{id} (Delete a contact list) mutates Reply.io data; review records before
  execution.
- `share_contact_list`: POST `/v3/contact-lists/{{ record.id }}/share` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/contact-lists/{id}/share (Share a contact list) mutates Reply.io data; review records before
  execution.
- `unshare_contact_list`: POST `/v3/contact-lists/{{ record.id }}/unshare` - kind `update`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/contact-lists/{id}/unshare (Unshare a contact list) mutates Reply.io data; review records
  before execution.
- `move_contacts_to_contact_list`: POST `/v3/contact-lists/{{ record.id }}/move-contacts` - kind
  `update`; body type `json`; path fields `id`; body fields `contactIds`; required record fields
  `id`; accepted fields `contactIds`, `id`; risk: POST /v3/contact-lists/{id}/move-contacts (Move
  contacts to a contact list) mutates Reply.io data; review records before execution.
- `add_contacts_to_contact_list`: POST `/v3/contact-lists/{{ record.id }}/add-contacts` - kind
  `create`; body type `json`; path fields `id`; body fields `contactIds`; required record fields
  `id`; accepted fields `contactIds`, `id`; risk: POST /v3/contact-lists/{id}/add-contacts (Add
  contacts to a contact list) mutates Reply.io data; review records before execution.
- `create_custom_field`: POST `/v3/custom-fields` - kind `create`; body type `json`; body fields
  `title`, `fieldType`, `metadata`, `orgWide`; accepted fields `fieldType`, `metadata`, `orgWide`,
  `title`; risk: POST /v3/custom-fields (Create a custom field) mutates Reply.io data; review
  records before execution.
- `update_custom_field`: PUT `/v3/custom-fields/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; body fields `title`, `fieldType`, `metadata`; required record fields `id`;
  accepted fields `fieldType`, `id`, `metadata`, `title`; risk: PUT /v3/custom-fields/{id} (Update a
  custom field) mutates Reply.io data; review records before execution.
- `delete_custom_field`: DELETE `/v3/custom-fields/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/custom-fields/{id} (Delete a custom field) mutates Reply.io data; review records before
  execution.
- `send_direct_email`: POST `/v3/contacts/{{ record.id }}/send-direct-email` - kind `create`; body
  type `json`; path fields `id`; body fields `subject`, `body`, `emailAccountId`; required record
  fields `id`; accepted fields `body`, `emailAccountId`, `id`, `subject`; risk: POST
  /v3/contacts/{id}/send-direct-email (Send a direct email to a contact) mutates Reply.io data;
  review records before execution.
- `send_direct_linked_in_connect`: POST `/v3/contacts/{{ record.id }}/send-direct-linkedin-connect`
  - kind `create`; body type `json`; path fields `id`; body fields `linkedInAccountId`, `message`;
  required record fields `id`; accepted fields `id`, `linkedInAccountId`, `message`; risk: POST
  /v3/contacts/{id}/send-direct-linkedin-connect (Send a LinkedIn connection request to a contact)
  mutates Reply.io data; review records before execution.
- `send_direct_linked_in_in_mail`: POST `/v3/contacts/{{ record.id }}/send-direct-linkedin-inmail` -
  kind `create`; body type `json`; path fields `id`; body fields `linkedInAccountId`, `subject`,
  `body`; required record fields `id`; accepted fields `body`, `id`, `linkedInAccountId`, `subject`;
  risk: POST /v3/contacts/{id}/send-direct-linkedin-inmail (Send a LinkedIn InMail to a contact)
  mutates Reply.io data; review records before execution.
- `send_direct_linked_in_message`: POST `/v3/contacts/{{ record.id }}/send-direct-linkedin-message`
  - kind `create`; body type `json`; path fields `id`; body fields `linkedInAccountId`, `message`;
  required record fields `id`; accepted fields `id`, `linkedInAccountId`, `message`; risk: POST
  /v3/contacts/{id}/send-direct-linkedin-message (Send a LinkedIn message to a contact) mutates
  Reply.io data; review records before execution.
- `send_direct_linked_in_voice`: POST `/v3/contacts/{{ record.id }}/send-direct-linkedin-voice` -
  kind `create`; body type `json`; path fields `id`; body fields `linkedInAccountId`,
  `voiceAttachmentId`; required record fields `id`; accepted fields `id`, `linkedInAccountId`,
  `voiceAttachmentId`; risk: POST /v3/contacts/{id}/send-direct-linkedin-voice (Send a LinkedIn
  voice message to a contact) mutates Reply.io data; review records before execution.
- `send_direct_linked_in_ai_voice`: POST `/v3/contacts/{{ record.id
  }}/send-direct-linkedin-ai-voice` - kind `create`; body type `json`; path fields `id`; body fields
  `linkedInAccountId`, `script`; required record fields `id`; accepted fields `id`,
  `linkedInAccountId`, `script`; risk: POST /v3/contacts/{id}/send-direct-linkedin-ai-voice (Send an
  AI-generated LinkedIn voice message to a contact) mutates Reply.io data; review records before
  execution.
- `create_email_account`: POST `/v3/email-accounts` - kind `create`; body type `json`; body fields
  `connection`, `safety`, `signature`, `optOut`, `rampUp`, `tags`; accepted fields `connection`,
  `optOut`, `rampUp`, `safety`, `signature`, `tags`; risk: POST /v3/email-accounts (Create an email
  account) mutates Reply.io data; review records before execution.
- `update_email_account`: PATCH `/v3/email-accounts/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; body fields `connection`, `safety`, `signature`, `optOut`, `rampUp`,
  `tags`; required record fields `id`; accepted fields `connection`, `id`, `optOut`, `rampUp`,
  `safety`, `signature`, `tags`; risk: PATCH /v3/email-accounts/{id} (Update an email account)
  mutates Reply.io data; review records before execution.
- `delete_email_account`: DELETE `/v3/email-accounts/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/email-accounts/{id} (Delete an email account) mutates Reply.io data; review records before
  execution.
- `set_default_email_account`: POST `/v3/email-accounts/{{ record.id }}/set-default` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: POST /v3/email-accounts/{id}/set-default (Set default email account) mutates Reply.io data;
  review records before execution.
- `resume_sending`: POST `/v3/email-accounts/{{ record.id }}/resume-sending` - kind `create`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/email-accounts/{id}/resume-sending (Resume sending) mutates Reply.io data; review records
  before execution.
- `bulk_delete_email_accounts`: POST `/v3/email-accounts/bulk-delete` - kind `custom`; body type
  `json`; body fields `ids`; accepted fields `ids`; risk: POST /v3/email-accounts/bulk-delete (Bulk
  delete email accounts) mutates Reply.io data; review records before execution.
- `update_email_account_tag`: PUT `/v3/email-accounts/tags/{{ record.tagId }}` - kind `update`; body
  type `json`; path fields `tagId`; body fields `name`, `colorId`; required record fields `tagId`;
  accepted fields `colorId`, `name`, `tagId`; risk: PUT /v3/email-accounts/tags/{tagId} (Update a
  tag) mutates Reply.io data; review records before execution.
- `add_tags_to_email_account`: POST `/v3/email-accounts/{{ record.id }}/tags` - kind `create`; body
  type `json`; path fields `id`; body fields `tags`; required record fields `id`; accepted fields
  `id`, `tags`; risk: POST /v3/email-accounts/{id}/tags (Add tags to an email account) mutates
  Reply.io data; review records before execution.
- `remove_tags_from_email_account`: DELETE `/v3/email-accounts/{{ record.id }}/tags` - kind
  `delete`; body type `json`; path fields `id`; body fields `tags`; required record fields `id`;
  accepted fields `id`, `tags`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /v3/email-accounts/{id}/tags (Remove tags from an email account)
  mutates Reply.io data; review records before execution.
- `create_email_template`: POST `/v3/email-templates` - kind `create`; body type `json`; body fields
  `name`, `body`, `folderType`, `subject`, `folderId`, `attachmentIds`; accepted fields
  `attachmentIds`, `body`, `folderId`, `folderType`, `name`, `subject`; risk: POST
  /v3/email-templates (Create an email template) mutates Reply.io data; review records before
  execution.
- `update_email_template`: PUT `/v3/email-templates/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; body fields `name`, `body`, `subject`, `attachmentIds`; required record
  fields `id`; accepted fields `attachmentIds`, `body`, `id`, `name`, `subject`; risk: PUT
  /v3/email-templates/{id} (Update an email template) mutates Reply.io data; review records before
  execution.
- `delete_email_template`: DELETE `/v3/email-templates/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/email-templates/{id} (Delete an email template) mutates Reply.io data; review records before
  execution.
- `clone_email_template`: POST `/v3/email-templates/{{ record.id }}/clone` - kind `create`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/email-templates/{id}/clone (Clone an email template) mutates Reply.io data; review records
  before execution.
- `move_email_template`: POST `/v3/email-templates/{{ record.id }}/move` - kind `update`; body type
  `json`; path fields `id`; body fields `folderId`, `folderType`; required record fields `id`;
  accepted fields `folderId`, `folderType`, `id`; risk: POST /v3/email-templates/{id}/move (Move an
  email template) mutates Reply.io data; review records before execution.
- `render_email_template`: POST `/v3/email-templates/{{ record.id }}/render` - kind `custom`; body
  type `json`; path fields `id`; body fields `contactId`, `sequenceId`, `emailAccountId`; required
  record fields `id`; accepted fields `contactId`, `emailAccountId`, `id`, `sequenceId`; risk: POST
  /v3/email-templates/{id}/render (Render an email template) mutates Reply.io data; review records
  before execution.
- `send_test_email_template`: POST `/v3/email-templates/{{ record.id }}/send-test` - kind `create`;
  body type `json`; path fields `id`; body fields `email`, `emailAccountId`; required record fields
  `id`; accepted fields `email`, `emailAccountId`, `id`; risk: POST
  /v3/email-templates/{id}/send-test (Send a test email) mutates Reply.io data; review records
  before execution.
- `create_email_template_folder`: POST `/v3/email-template-folders` - kind `create`; body type
  `json`; body fields `name`, `folderType`; accepted fields `folderType`, `name`; risk: POST
  /v3/email-template-folders (Create an email template folder) mutates Reply.io data; review records
  before execution.
- `share_email_template_folder`: POST `/v3/email-template-folders/{{ record.id }}/share` - kind
  `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: POST /v3/email-template-folders/{id}/share (Share an email template folder) mutates Reply.io
  data; review records before execution.
- `estimate_email_validation`: POST `/v3/email-validations/estimate` - kind `custom`; body type
  `json`; body fields `contactIds`, `acceptPartial`; accepted fields `acceptPartial`, `contactIds`;
  risk: POST /v3/email-validations/estimate (Estimate email validation) mutates Reply.io data;
  review records before execution.
- `schedule_email_validation`: POST `/v3/email-validations/schedule` - kind `create`; body type
  `json`; body fields `contactIds`, `acceptPartial`; accepted fields `acceptPartial`, `contactIds`;
  risk: POST /v3/email-validations/schedule (Schedule email validation) mutates Reply.io data;
  review records before execution.
- `create_holiday_calendar`: POST `/v3/holiday-calendars` - kind `create`; body type `json`; body
  fields `name`, `repeatEveryYear`, `holidays`; accepted fields `holidays`, `name`,
  `repeatEveryYear`; risk: POST /v3/holiday-calendars (Create a holiday calendar) mutates Reply.io
  data; review records before execution.
- `update_holiday_calendar`: PUT `/v3/holiday-calendars/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; body fields `name`, `repeatEveryYear`, `holidays`; required record
  fields `id`; accepted fields `holidays`, `id`, `name`, `repeatEveryYear`; risk: PUT
  /v3/holiday-calendars/{id} (Update a holiday calendar) mutates Reply.io data; review records
  before execution.
- `delete_holiday_calendar`: DELETE `/v3/holiday-calendars/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/holiday-calendars/{id} (Delete a holiday calendar) mutates Reply.io data; review records
  before execution.
- `delete_inbox_thread`: DELETE `/v3/inbox/threads/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/inbox/threads/{id} (Delete inbox thread) mutates Reply.io data; review records before
  execution.
- `bulk_delete_inbox_threads`: POST `/v3/inbox/threads/bulk-delete` - kind `custom`; body type
  `json`; body fields `threadIds`; accepted fields `threadIds`; risk: POST
  /v3/inbox/threads/bulk-delete (Bulk-delete inbox threads) mutates Reply.io data; review records
  before execution.
- `mark_inbox_threads_as_read`: POST `/v3/inbox/threads/mark-as-read` - kind `update`; body type
  `json`; body fields `threadIds`; accepted fields `threadIds`; risk: POST
  /v3/inbox/threads/mark-as-read (Mark threads as read) mutates Reply.io data; review records before
  execution.
- `mark_inbox_threads_as_unread`: POST `/v3/inbox/threads/mark-as-unread` - kind `update`; body type
  `json`; body fields `threadIds`; accepted fields `threadIds`; risk: POST
  /v3/inbox/threads/mark-as-unread (Mark threads as unread) mutates Reply.io data; review records
  before execution.
- `send_inbox_thread_message`: POST `/v3/inbox/threads/{{ record.id }}/messages` - kind `create`;
  body type `json`; path fields `id`; body fields `channel`, `message`, `attachmentIds`, `cc`,
  `bcc`, `applySignature`; required record fields `id`; accepted fields `applySignature`,
  `attachmentIds`, `bcc`, `cc`, `channel`, `id`, `message`; risk: POST
  /v3/inbox/threads/{id}/messages (Send a reply within a thread) mutates Reply.io data; review
  records before execution.
- `set_inbox_thread_category`: PUT `/v3/inbox/threads/{{ record.id }}/category` - kind `update`;
  body type `json`; path fields `id`; body fields `categoryId`; required record fields `id`;
  accepted fields `categoryId`, `id`; risk: PUT /v3/inbox/threads/{id}/category (Assign or clear a
  thread's category) mutates Reply.io data; review records before execution.
- `set_inbox_thread_meeting_intent`: PUT `/v3/inbox/threads/{{ record.id }}/meeting-intent` - kind
  `update`; body type `json`; path fields `id`; body fields `hasMeetingIntent`; required record
  fields `id`; accepted fields `hasMeetingIntent`, `id`; risk: PUT
  /v3/inbox/threads/{id}/meeting-intent (Toggle thread meeting-intent) mutates Reply.io data; review
  records before execution.
- `create_inbox_category`: POST `/v3/inbox/threads/categories` - kind `create`; body type `json`;
  body fields `name`, `color`; accepted fields `color`, `name`; risk: POST
  /v3/inbox/threads/categories (Create inbox category) mutates Reply.io data; review records before
  execution.
- `update_inbox_category`: PUT `/v3/inbox/threads/categories/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; body fields `name`, `color`; required record fields `id`; accepted
  fields `color`, `id`, `name`; risk: PUT /v3/inbox/threads/categories/{id} (Update inbox category)
  mutates Reply.io data; review records before execution.
- `delete_inbox_category`: DELETE `/v3/inbox/threads/categories/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/inbox/threads/categories/{id} (Delete inbox category) mutates Reply.io data; review records
  before execution.
- `assign_threads_to_inbox_category`: POST `/v3/inbox/threads/categories/{{ record.id
  }}/thread-links/bulk` - kind `create`; body type `json`; path fields `id`; body fields
  `threadIds`; required record fields `id`; accepted fields `id`, `threadIds`; risk: POST
  /v3/inbox/threads/categories/{id}/thread-links/bulk (Assign threads to a category) mutates
  Reply.io data; review records before execution.
- `unassign_threads_from_inbox_category`: POST `/v3/inbox/threads/categories/{{ record.id
  }}/thread-links/bulk-delete` - kind `create`; body type `json`; path fields `id`; body fields
  `threadIds`; required record fields `id`; accepted fields `id`, `threadIds`; risk: POST
  /v3/inbox/threads/categories/{id}/thread-links/bulk-delete (Unassign threads from a category)
  mutates Reply.io data; review records before execution.
- `delete_linked_in_account`: DELETE `/v3/linkedin-accounts/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/linkedin-accounts/{id} (Delete a LinkedIn account) mutates Reply.io data; review records
  before execution.
- `bulk_delete_linked_in_accounts`: POST `/v3/linkedin-accounts/bulk-delete` - kind `custom`; body
  type `json`; body fields `ids`, `force`; accepted fields `force`, `ids`; risk: POST
  /v3/linkedin-accounts/bulk-delete (Bulk delete LinkedIn accounts) mutates Reply.io data; review
  records before execution.
- `toggle_linked_in_account_status`: POST `/v3/linkedin-accounts/{{ record.id }}/toggle-status` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: POST /v3/linkedin-accounts/{id}/toggle-status (Toggle LinkedIn account status) mutates
  Reply.io data; review records before execution.
- `update_linked_in_account_limits`: PUT `/v3/linkedin-accounts/{{ record.id }}/limits` - kind
  `update`; body type `json`; path fields `id`; body fields `limitsMode`, `rangeMin`, `rangeMax`,
  `fixedMax`; required record fields `id`; accepted fields `fixedMax`, `id`, `limitsMode`,
  `rangeMax`, `rangeMin`; risk: PUT /v3/linkedin-accounts/{id}/limits (Update LinkedIn account
  limits) mutates Reply.io data; review records before execution.
- `update_linked_in_account_revoke_settings`: PUT `/v3/linkedin-accounts/{{ record.id
  }}/revoke-settings` - kind `update`; body type `json`; path fields `id`; body fields `enabled`,
  `periodDays`; required record fields `id`; accepted fields `enabled`, `id`, `periodDays`; risk:
  PUT /v3/linkedin-accounts/{id}/revoke-settings (Update LinkedIn account revoke settings) mutates
  Reply.io data; review records before execution.
- `create_linked_in_connection_link`: POST `/v3/linkedin-accounts/connection-link` - kind `create`;
  body type `json`; body fields `name`; accepted fields `name`; risk: POST
  /v3/linkedin-accounts/connection-link (Create a connection link) mutates Reply.io data; review
  records before execution.
- `create_direct_linked_in_connection_link`: POST `/v3/linkedin-accounts/connect` - kind `create`;
  body type `none`; risk: POST /v3/linkedin-accounts/connect (Create a direct connection link)
  mutates Reply.io data; review records before execution.
- `reconnect_linked_in_account`: POST `/v3/linkedin-accounts/{{ record.id }}/reconnect` - kind
  `create`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: POST /v3/linkedin-accounts/{id}/reconnect (Reconnect a LinkedIn account) mutates Reply.io
  data; review records before execution.
- `delete_pending_linked_in_account`: DELETE `/v3/linkedin-accounts/pending/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/linkedin-accounts/pending/{id} (Delete a pending LinkedIn account) mutates Reply.io data;
  review records before execution.
- `get_team_performance_overview`: POST `/v3/reporting/team-performance/overview` - kind `custom`;
  body type `json`; body fields `filters`; accepted fields `filters`; risk: POST
  /v3/reporting/team-performance/overview (Get team performance overview) mutates Reply.io data;
  review records before execution.
- `get_channel_efficiency_overview`: POST `/v3/reporting/channel-efficiency/overview` - kind
  `custom`; body type `json`; body fields `filters`; accepted fields `filters`; risk: POST
  /v3/reporting/channel-efficiency/overview (Get channel efficiency overview) mutates Reply.io data;
  review records before execution.
- `get_emails_list`: POST `/v3/reporting/emails` - kind `custom`; body type `json`; body fields
  `filters`; accepted fields `filters`; risk: POST /v3/reporting/emails (List email activity)
  mutates Reply.io data; review records before execution.
- `get_calls_list`: POST `/v3/reporting/calls` - kind `custom`; body type `json`; body fields
  `filters`; accepted fields `filters`; risk: POST /v3/reporting/calls (List call activity) mutates
  Reply.io data; review records before execution.
- `get_tasks_list`: POST `/v3/reporting/tasks` - kind `custom`; body type `json`; body fields
  `filters`; accepted fields `filters`; risk: POST /v3/reporting/tasks (List task activity) mutates
  Reply.io data; review records before execution.
- `get_linked_in_list`: POST `/v3/reporting/linkedin` - kind `custom`; body type `json`; body fields
  `filters`; accepted fields `filters`; risk: POST /v3/reporting/linkedin (List LinkedIn activity)
  mutates Reply.io data; review records before execution.
- `get_meetings_list`: POST `/v3/reporting/team-performance/meetings` - kind `custom`; body type
  `json`; body fields `filters`; accepted fields `filters`; risk: POST
  /v3/reporting/team-performance/meetings (List meetings) mutates Reply.io data; review records
  before execution.
- `create_schedule`: POST `/v3/schedules` - kind `create`; body type `json`; body fields `name`,
  `timezoneId`, `excludeHolidays`, `useProspectTimezone`, `useFollowUpSchedule`, `mainTimings`,
  `followUpTimings`; accepted fields `excludeHolidays`, `followUpTimings`, `mainTimings`, `name`,
  `timezoneId`, `useFollowUpSchedule`, `useProspectTimezone`; risk: POST /v3/schedules (Create a
  schedule) mutates Reply.io data; review records before execution.
- `update_schedule`: PUT `/v3/schedules/{{ record.id }}` - kind `create`; body type `json`; path
  fields `id`; body fields `name`, `timezoneId`, `excludeHolidays`, `useProspectTimezone`,
  `useFollowUpSchedule`, `mainTimings`, `followUpTimings`; required record fields `id`; accepted
  fields `excludeHolidays`, `followUpTimings`, `id`, `mainTimings`, `name`, `timezoneId`,
  `useFollowUpSchedule`, `useProspectTimezone`; risk: PUT /v3/schedules/{id} (Update a schedule)
  mutates Reply.io data; review records before execution.
- `delete_schedule`: DELETE `/v3/schedules/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: DELETE /v3/schedules/{id} (Delete a schedule)
  mutates Reply.io data; review records before execution.
- `set_default_schedule`: POST `/v3/schedules/{{ record.id }}/set-default` - kind `create`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/schedules/{id}/set-default (Set default schedule) mutates Reply.io data; review records before
  execution.
- `link_holiday_calendar_to_schedule`: POST `/v3/schedules/{{ record.id }}/holiday-calendar-links` -
  kind `create`; body type `json`; path fields `id`; body fields `calendarId`; required record
  fields `id`; accepted fields `calendarId`, `id`; risk: POST
  /v3/schedules/{id}/holiday-calendar-links (Link a holiday calendar to a schedule) mutates Reply.io
  data; review records before execution.
- `unlink_holiday_calendar_from_schedule`: DELETE `/v3/schedules/{{ record.id
  }}/holiday-calendar-links/{{ record.calendarId }}` - kind `delete`; body type `none`; path fields
  `id`, `calendarId`; required record fields `id`, `calendarId`; accepted fields `calendarId`, `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/schedules/{id}/holiday-calendar-links/{calendarId} (Unlink a holiday calendar from a schedule)
  mutates Reply.io data; review records before execution.
- `create_sequence`: POST `/v3/sequences` - kind `create`; body type `json`; body fields `name`,
  `scheduleId`, `settings`, `emailAccounts`, `linkedInAccounts`, `steps`; accepted fields
  `emailAccounts`, `linkedInAccounts`, `name`, `scheduleId`, `settings`, `steps`; risk: POST
  /v3/sequences (Create a sequence) mutates Reply.io data; review records before execution.
- `update_sequence`: PATCH `/v3/sequences/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `name`, `scheduleId`, `settings`, `emailAccounts`, `linkedInAccounts`;
  required record fields `id`; accepted fields `emailAccounts`, `id`, `linkedInAccounts`, `name`,
  `scheduleId`, `settings`; risk: PATCH /v3/sequences/{id} (Update a sequence) mutates Reply.io
  data; review records before execution.
- `delete_sequence`: DELETE `/v3/sequences/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: DELETE /v3/sequences/{id} (Delete a sequence)
  mutates Reply.io data; review records before execution.
- `start_sequence`: POST `/v3/sequences/{{ record.id }}/start` - kind `create`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/sequences/{id}/start (Start a sequence) mutates Reply.io data; review records before
  execution.
- `pause_sequence`: POST `/v3/sequences/{{ record.id }}/pause` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/sequences/{id}/pause (Pause a sequence) mutates Reply.io data; review records before
  execution.
- `archive_sequence`: POST `/v3/sequences/{{ record.id }}/archive` - kind `update`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/sequences/{id}/archive (Archive a sequence) mutates Reply.io data; review records before
  execution.
- `update_sequence_owner`: PUT `/v3/sequences/{{ record.id }}/owner` - kind `update`; body type
  `json`; path fields `id`; body fields `userId`; required record fields `id`; accepted fields `id`,
  `userId`; risk: PUT /v3/sequences/{id}/owner (Change sequence owner) mutates Reply.io data; review
  records before execution.
- `bulk_start_sequences`: POST `/v3/sequences/start` - kind `create`; body type `json`; body fields
  `ids`; accepted fields `ids`; risk: POST /v3/sequences/start (Bulk start sequences) mutates
  Reply.io data; review records before execution.
- `bulk_pause_sequences`: POST `/v3/sequences/pause` - kind `update`; body type `json`; body fields
  `ids`; accepted fields `ids`; risk: POST /v3/sequences/pause (Bulk pause sequences) mutates
  Reply.io data; review records before execution.
- `bulk_archive_sequences`: POST `/v3/sequences/archive` - kind `update`; body type `json`; body
  fields `ids`; accepted fields `ids`; risk: POST /v3/sequences/archive (Bulk archive sequences)
  mutates Reply.io data; review records before execution.
- `bulk_delete_sequences`: POST `/v3/sequences/bulk-delete` - kind `custom`; body type `json`; body
  fields `ids`; accepted fields `ids`; risk: POST /v3/sequences/bulk-delete (Bulk delete sequences)
  mutates Reply.io data; review records before execution.
- `bulk_change_sequence_owner`: POST `/v3/sequences/batch/owner` - kind `update`; body type `json`;
  body fields `ids`, `userId`; accepted fields `ids`, `userId`; risk: POST /v3/sequences/batch/owner
  (Bulk change sequence owner) mutates Reply.io data; review records before execution.
- `save_sequence_as_template`: POST `/v3/sequences/{{ record.id }}/save-as-template` - kind
  `custom`; body type `json`; path fields `id`; body fields `name`, `description`, `scope`; required
  record fields `id`; accepted fields `description`, `id`, `name`, `scope`; risk: POST
  /v3/sequences/{id}/save-as-template (Save a sequence as a template) mutates Reply.io data; review
  records before execution.
- `create_sequence_from_template`: POST `/v3/sequences/create-from-template` - kind `create`; body
  type `json`; body fields `templateId`, `sequenceFolderId`; accepted fields `sequenceFolderId`,
  `templateId`; risk: POST /v3/sequences/create-from-template (Create a sequence from a template)
  mutates Reply.io data; review records before execution.
- `remove_contact_from_sequence`: DELETE `/v3/sequences/{{ record.id }}/contact-links/{{
  record.contact_id }}` - kind `delete`; body type `none`; path fields `id`, `contact_id`; required
  record fields `id`, `contact_id`; accepted fields `contact_id`, `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/sequences/{id}/contact-links/{contact_id} (Remove contact from sequence) mutates Reply.io
  data; review records before execution.
- `bulk_add_contacts_to_sequence`: POST `/v3/sequences/{{ record.id }}/contact-links/bulk` - kind
  `create`; body type `json`; path fields `id`; body fields `contactIds`, `removeFromExisting`,
  `startStepId`, `ignoreStepDelay`, `startFrom`; required record fields `id`; accepted fields
  `contactIds`, `id`, `ignoreStepDelay`, `removeFromExisting`, `startFrom`, `startStepId`; risk:
  POST /v3/sequences/{id}/contact-links/bulk (Bulk add contacts to sequence) mutates Reply.io data;
  review records before execution.
- `bulk_remove_contacts_from_sequence`: POST `/v3/sequences/{{ record.id
  }}/contact-links/bulk-delete` - kind `update`; body type `json`; path fields `id`; body fields
  `contactIds`; required record fields `id`; accepted fields `contactIds`, `id`; risk: POST
  /v3/sequences/{id}/contact-links/bulk-delete (Bulk remove contacts from sequence) mutates Reply.io
  data; review records before execution.
- `set_sequence_contacts_status_in_sequence`: POST `/v3/sequences/{{ record.id
  }}/contacts/set-status-in-sequence` - kind `update`; body type `json`; path fields `id`; body
  fields `contactIds`, `statusInSequence`; required record fields `id`; accepted fields
  `contactIds`, `id`, `statusInSequence`; risk: POST
  /v3/sequences/{id}/contacts/set-status-in-sequence (Set contacts' status in this sequence) mutates
  Reply.io data; review records before execution.
- `set_sequence_contacts_replied`: POST `/v3/sequences/{{ record.id }}/contacts/set-replied` - kind
  `update`; body type `json`; path fields `id`; body fields `contactIds`, `isReplied`; required
  record fields `id`; accepted fields `contactIds`, `id`, `isReplied`; risk: POST
  /v3/sequences/{id}/contacts/set-replied (Mark or unmark contacts as replied in this sequence)
  mutates Reply.io data; review records before execution.
- `set_sequence_contacts_bounced`: POST `/v3/sequences/{{ record.id }}/contacts/set-bounced` - kind
  `update`; body type `json`; path fields `id`; body fields `contactIds`, `isBounced`,
  `resendEmails`; required record fields `id`; accepted fields `contactIds`, `id`, `isBounced`,
  `resendEmails`; risk: POST /v3/sequences/{id}/contacts/set-bounced (Mark or unmark contacts as
  bounced in this sequence) mutates Reply.io data; review records before execution.
- `assign_email_account_to_sequence`: POST `/v3/sequences/{{ record.id }}/email-account-links` -
  kind `create`; body type `json`; path fields `id`; body fields `emailAccountId`; required record
  fields `id`; accepted fields `emailAccountId`, `id`; risk: POST
  /v3/sequences/{id}/email-account-links (Assign email account to sequence) mutates Reply.io data;
  review records before execution.
- `set_sequence_email_accounts`: PUT `/v3/sequences/{{ record.id }}/email-account-links` - kind
  `update`; body type `json`; path fields `id`; body fields `emailAccountIds`; required record
  fields `id`; accepted fields `emailAccountIds`, `id`; risk: PUT
  /v3/sequences/{id}/email-account-links (Set sequence email accounts) mutates Reply.io data; review
  records before execution.
- `remove_email_account_from_sequence`: DELETE `/v3/sequences/{{ record.id }}/email-account-links/{{
  record.email_account_id }}` - kind `delete`; body type `none`; path fields `id`,
  `email_account_id`; required record fields `id`, `email_account_id`; accepted fields
  `email_account_id`, `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /v3/sequences/{id}/email-account-links/{email_account_id} (Remove
  email account from sequence) mutates Reply.io data; review records before execution.
- `create_sequence_folder`: POST `/v3/sequence-folders` - kind `create`; body type `json`; body
  fields `name`; accepted fields `name`; risk: POST /v3/sequence-folders (Create a sequence folder)
  mutates Reply.io data; review records before execution.
- `update_sequence_folder`: PUT `/v3/sequence-folders/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; body fields `name`; required record fields `id`; accepted fields `id`,
  `name`; risk: PUT /v3/sequence-folders/{id} (Update a sequence folder) mutates Reply.io data;
  review records before execution.
- `delete_sequence_folder`: DELETE `/v3/sequence-folders/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/sequence-folders/{id} (Delete a sequence folder) mutates Reply.io data; review records before
  execution.
- `bulk_assign_sequences_to_folder`: POST `/v3/sequence-folders/{{ record.id }}/sequence-links/bulk`
  - kind `create`; body type `json`; path fields `id`; body fields `sequenceIds`; required record
  fields `id`; accepted fields `id`, `sequenceIds`; risk: POST
  /v3/sequence-folders/{id}/sequence-links/bulk (Bulk assign sequences to a folder) mutates Reply.io
  data; review records before execution.
- `bulk_unassign_sequences_from_folder`: POST `/v3/sequence-folders/{{ record.id
  }}/sequence-links/bulk-delete` - kind `create`; body type `json`; path fields `id`; body fields
  `sequenceIds`; required record fields `id`; accepted fields `id`, `sequenceIds`; risk: POST
  /v3/sequence-folders/{id}/sequence-links/bulk-delete (Bulk unassign sequences from a folder)
  mutates Reply.io data; review records before execution.
- `assign_linked_in_account_to_sequence`: POST `/v3/sequences/{{ record.id
  }}/linkedin-account-links` - kind `create`; body type `json`; path fields `id`; body fields
  `linkedInAccountId`; required record fields `id`; accepted fields `id`, `linkedInAccountId`; risk:
  POST /v3/sequences/{id}/linkedin-account-links (Assign a LinkedIn account to a sequence) mutates
  Reply.io data; review records before execution.
- `remove_linked_in_account_from_sequence`: DELETE `/v3/sequences/{{ record.id
  }}/linkedin-account-links/{{ record.linkedInAccountId }}` - kind `delete`; body type `none`; path
  fields `id`, `linkedInAccountId`; required record fields `id`, `linkedInAccountId`; accepted
  fields `id`, `linkedInAccountId`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: DELETE
  /v3/sequences/{id}/linkedin-account-links/{linkedInAccountId} (Remove a LinkedIn account from a
  sequence) mutates Reply.io data; review records before execution.
- `create_sequence_step`: POST `/v3/sequences/{{ record.id }}/steps` - kind `create`; body type
  `json`; path fields `id`; body fields `type`, `delayInMinutes`, `executionMode`, `variants`,
  `parentId`, `ifConditionPositive`; required record fields `id`; accepted fields `delayInMinutes`,
  `executionMode`, `id`, `ifConditionPositive`, `parentId`, `type`, `variants`; risk: POST
  /v3/sequences/{id}/steps (Create a sequence step) mutates Reply.io data; review records before
  execution.
- `update_sequence_step`: PUT `/v3/sequences/{{ record.id }}/steps/{{ record.step_id }}` - kind
  `update`; body type `json`; path fields `id`, `step_id`; body fields `type`, `delayInMinutes`,
  `executionMode`, `variants`, `parentId`, `ifConditionPositive`; required record fields `id`,
  `step_id`; accepted fields `delayInMinutes`, `executionMode`, `id`, `ifConditionPositive`,
  `parentId`, `step_id`, `type`, `variants`; risk: PUT /v3/sequences/{id}/steps/{step_id} (Update a
  sequence step) mutates Reply.io data; review records before execution.
- `delete_sequence_step`: DELETE `/v3/sequences/{{ record.id }}/steps/{{ record.step_id }}` - kind
  `delete`; body type `none`; path fields `id`, `step_id`; required record fields `id`, `step_id`;
  accepted fields `id`, `step_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /v3/sequences/{id}/steps/{step_id} (Delete a sequence step) mutates
  Reply.io data; review records before execution.
- `bulk_delete_sequence_steps`: POST `/v3/sequences/{{ record.id }}/steps/bulk-delete` - kind
  `custom`; body type `json`; path fields `id`; body fields `ids`; required record fields `id`;
  accepted fields `id`, `ids`; risk: POST /v3/sequences/{id}/steps/bulk-delete (Bulk delete sequence
  steps) mutates Reply.io data; review records before execution.
- `enable_sequence_step_variants`: POST `/v3/sequences/{{ record.id }}/steps/{{ record.step_id
  }}/enable` - kind `update`; body type `json`; path fields `id`, `step_id`; body fields
  `variantIds`; required record fields `id`, `step_id`; accepted fields `id`, `step_id`,
  `variantIds`; risk: POST /v3/sequences/{id}/steps/{step_id}/enable (Enable step variants) mutates
  Reply.io data; review records before execution.
- `disable_sequence_step_variants`: POST `/v3/sequences/{{ record.id }}/steps/{{ record.step_id
  }}/disable` - kind `update`; body type `json`; path fields `id`, `step_id`; body fields
  `variantIds`; required record fields `id`, `step_id`; accepted fields `id`, `step_id`,
  `variantIds`; risk: POST /v3/sequences/{id}/steps/{step_id}/disable (Disable step variants)
  mutates Reply.io data; review records before execution.
- `create_sequence_step_variant`: POST `/v3/sequences/{{ record.id }}/steps/{{ record.step_id
  }}/variants` - kind `create`; body type `json`; path fields `id`, `step_id`; body fields
  `message`, `subject`, `isEnabled`, `attachmentIds`; required record fields `id`, `step_id`;
  accepted fields `attachmentIds`, `id`, `isEnabled`, `message`, `step_id`, `subject`; risk: POST
  /v3/sequences/{id}/steps/{step_id}/variants (Create a step variant) mutates Reply.io data; review
  records before execution.
- `bulk_delete_sequence_step_variants`: POST `/v3/sequences/{{ record.id }}/steps/{{ record.step_id
  }}/variants/bulk-delete` - kind `custom`; body type `json`; path fields `id`, `step_id`; body
  fields `ids`; required record fields `id`, `step_id`; accepted fields `id`, `ids`, `step_id`;
  risk: POST /v3/sequences/{id}/steps/{step_id}/variants/bulk-delete (Bulk delete step variants)
  mutates Reply.io data; review records before execution.
- `update_sequence_step_variant`: PUT `/v3/sequences/{{ record.id }}/steps/{{ record.step_id
  }}/variants/{{ record.variant_id }}` - kind `update`; body type `json`; path fields `id`,
  `step_id`, `variant_id`; body fields `message`, `subject`, `isEnabled`, `attachmentIds`; required
  record fields `id`, `step_id`, `variant_id`; accepted fields `attachmentIds`, `id`, `isEnabled`,
  `message`, `step_id`, `subject`, `variant_id`; risk: PUT
  /v3/sequences/{id}/steps/{step_id}/variants/{variant_id} (Update a step variant) mutates Reply.io
  data; review records before execution.
- `delete_sequence_step_variant`: DELETE `/v3/sequences/{{ record.id }}/steps/{{ record.step_id
  }}/variants/{{ record.variant_id }}` - kind `delete`; body type `none`; path fields `id`,
  `step_id`, `variant_id`; required record fields `id`, `step_id`, `variant_id`; accepted fields
  `id`, `step_id`, `variant_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: DELETE /v3/sequences/{id}/steps/{step_id}/variants/{variant_id} (Delete a
  step variant) mutates Reply.io data; review records before execution.
- `delete_sequence_template`: DELETE `/v3/sequence-templates/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /v3/sequence-templates/{id} (Delete a sequence template) mutates Reply.io data; review records
  before execution.
- `update_settings`: PATCH `/v3/settings` - kind `update`; body type `json`; body fields `account`,
  `emails`, `linkedIn`, `calls`, `contacts`, `beta`; accepted fields `account`, `beta`, `calls`,
  `contacts`, `emails`, `linkedIn`; risk: PATCH /v3/settings (Update settings) mutates Reply.io
  data; review records before execution.
- `create_task`: POST `/v3/tasks` - kind `create`; body type `json`; body fields `taskType`,
  `startAt`, `dueTo`, `template`, `contactId`, `linkedInTaskType`; accepted fields `contactId`,
  `dueTo`, `linkedInTaskType`, `startAt`, `taskType`, `template`; risk: POST /v3/tasks (Create a
  task) mutates Reply.io data; review records before execution.
- `update_task`: PUT `/v3/tasks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; body fields `taskType`, `startAt`, `dueTo`, `template`, `contactId`, `linkedInTaskType`;
  required record fields `id`; accepted fields `contactId`, `dueTo`, `id`, `linkedInTaskType`,
  `startAt`, `taskType`, `template`; risk: PUT /v3/tasks/{id} (Update a task) mutates Reply.io data;
  review records before execution.
- `delete_task`: DELETE `/v3/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: DELETE /v3/tasks/{id} (Delete a task) mutates
  Reply.io data; review records before execution.
- `assign_task`: PUT `/v3/tasks/{{ record.id }}/assigned-user` - kind `create`; body type `json`;
  path fields `id`; body fields `userId`; required record fields `id`; accepted fields `id`,
  `userId`; risk: PUT /v3/tasks/{id}/assigned-user (Reassign a task) mutates Reply.io data; review
  records before execution.
- `complete_task`: POST `/v3/tasks/{{ record.id }}/complete` - kind `update`; body type `json`; path
  fields `id`; body fields `callResolution`, `finishProspectInSequence`; required record fields
  `id`; accepted fields `callResolution`, `finishProspectInSequence`, `id`; risk: POST
  /v3/tasks/{id}/complete (Complete a task) mutates Reply.io data; review records before execution.
- `execute_task`: POST `/v3/tasks/{{ record.id }}/execute` - kind `custom`; body type `json`; path
  fields `id`; body fields `content`, `emailAccountId`; required record fields `id`; accepted fields
  `content`, `emailAccountId`, `id`; risk: POST /v3/tasks/{id}/execute (Execute and complete a task)
  mutates Reply.io data; review records before execution.
- `bulk_delete_tasks`: POST `/v3/tasks/bulk-delete` - kind `custom`; body type `json`; body fields
  `ids`; accepted fields `ids`; risk: POST /v3/tasks/bulk-delete (Bulk delete tasks) mutates
  Reply.io data; review records before execution.
- `batch_assign_tasks`: POST `/v3/tasks/batch/assign` - kind `create`; body type `json`; body fields
  `ids`, `userId`; accepted fields `ids`, `userId`; risk: POST /v3/tasks/batch/assign (Batch
  reassign tasks) mutates Reply.io data; review records before execution.
- `batch_complete_tasks`: POST `/v3/tasks/batch/complete` - kind `update`; body type `json`; body
  fields `ids`; accepted fields `ids`; risk: POST /v3/tasks/batch/complete (Batch complete tasks)
  mutates Reply.io data; review records before execution.
- `create_webhook`: POST `/v3/webhooks` - kind `create`; body type `json`; body fields `eventType`,
  `url`, `scope`, `enabled`, `payloadConfig`; accepted fields `enabled`, `eventType`,
  `payloadConfig`, `scope`, `url`; risk: POST /v3/webhooks (Create a webhook subscription) mutates
  Reply.io data; review records before execution.
- `update_webhook`: PUT `/v3/webhooks/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; body fields `eventType`, `url`, `scope`, `payloadConfig`; required record fields
  `id`; accepted fields `eventType`, `id`, `payloadConfig`, `scope`, `url`; risk: PUT
  /v3/webhooks/{id} (Update a webhook subscription) mutates Reply.io data; review records before
  execution.
- `delete_webhook`: DELETE `/v3/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: DELETE /v3/webhooks/{id} (Delete a webhook
  subscription) mutates Reply.io data; review records before execution.
- `test_webhook`: POST `/v3/webhooks/{{ record.id }}/test` - kind `custom`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: POST /v3/webhooks/{id}/test
  (Send a test payload) mutates Reply.io data; review records before execution.
- `enable_webhook`: POST `/v3/webhooks/{{ record.id }}/enable` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/webhooks/{id}/enable (Enable a webhook subscription) mutates Reply.io data; review records
  before execution.
- `disable_webhook`: POST `/v3/webhooks/{{ record.id }}/disable` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /v3/webhooks/{id}/disable (Disable a webhook subscription) mutates Reply.io data; review records
  before execution.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=100.
- API coverage includes 85 stream-backed endpoint group(s), 189 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=7, out_of_scope=70, requires_elevated_scope=2.
