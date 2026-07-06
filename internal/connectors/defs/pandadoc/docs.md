# Overview

Reads and writes documented PandaDoc public API resources across documents, templates, contacts,
folders, forms, logs, members, webhooks, workspaces, notary, and catalog surfaces.

Readable streams: `documents`, `templates`, `contacts`, `documents_document_id_ai_metadata`,
`documents_document_id_content`, `documents_document_id_docx_export_tasks_task_id`,
`documents_document_id_summary`, `documents_search`, `contacts_id`, `content_library_items`,
`content_library_items_id`, `content_library_items_id_details`,
`documents_document_id_auto_reminders`, `documents_document_id_auto_reminders_status`,
`documents_document_id_esign_disclosure`, `documents_document_id_sections`,
`documents_document_id_sections_section_id`, `documents_document_id_sections_uploads_upload_id`,
`documents_id`, `documents_id_attachments`, `documents_id_attachments_attachment_id`,
`documents_id_details`, `documents_id_fields`, `documents_id_linked_objects`, `documents_folders`,
`documents_linked_objects`, `forms`, `logs`, `logs_id`, `members`, `members_id`, `members_current`,
`sms_opt_outs`, `templates_id`, `templates_id_details`, `templates_folders`, `users`,
`users_user_id`, `webhook_events`, `webhook_events_id`, `webhook_subscriptions`,
`webhook_subscriptions_id`, `workspaces`, `documents_document_id_audit_trail`,
`documents_document_id_settings`, `logs_detail`, `logs_id_detail`, `notary_notaries`,
`notary_notarization_requests`, `notary_notarization_requests_session_request_id`,
`product_catalog_items_item_uuid`, `product_catalog_items_search`, `templates_template_id_settings`.

Write actions: `post_documents_document_id_docx_export_tasks`, `post_documents_ai_metadata`,
`post_contacts`, `delete_contacts_id`, `patch_contacts_id`, `post_content_library_items`,
`delete_documents`, `post_documents`, `patch_documents_document_id_auto_reminders`,
`put_documents_document_id_quotes_quote_id`, `post_documents_document_id_sections`,
`delete_documents_document_id_sections_section_id`, `post_documents_document_id_send_reminder`,
`delete_documents_id`, `patch_documents_id`, `post_documents_id_append_content_library_item`,
`post_documents_id_attachments`, `delete_documents_id_attachments_attachment_id`,
`post_documents_id_draft`, `post_documents_id_editing_sessions`, `patch_documents_id_fields`,
`post_documents_id_fields`, `post_documents_id_linked_objects`,
`delete_documents_id_linked_objects_linked_object_id`, `post_documents_id_move_to_folder_folder_id`,
`patch_documents_id_ownership`, `post_documents_id_recipients`,
`delete_documents_id_recipients_recipient_id`, `post_documents_id_recipients_recipient_id_reassign`,
`patch_documents_id_recipients_recipient_recipient_id`, `post_documents_id_send`,
`post_documents_id_session`, `patch_documents_id_status`, `post_documents_folders`,
`put_documents_folders_id`, `patch_documents_ownership`, `post_templates`, `delete_templates_id`,
`patch_templates_id`, `post_templates_id_editing_sessions`, `post_templates_folders`,
`put_templates_folders_id`, `post_users`, `post_webhook_subscriptions`,
`delete_webhook_subscriptions_id`, `patch_webhook_subscriptions_id`, `post_workspaces`,
`post_workspaces_workspace_id_deactivate`, `post_workspaces_workspace_id_members`,
`delete_workspaces_workspace_id_members_member_id`,
`patch_workspaces_workspace_id_members_member_id_role`, `patch_documents_document_id_settings`,
`post_notary_notarization_requests`, `delete_notary_notarization_requests_session_request_id`,
`post_product_catalog_items`, `delete_product_catalog_items_item_uuid`,
`patch_product_catalog_items_item_uuid`, `patch_templates_template_id_settings`.

Service API documentation: https://developers.pandadoc.com/reference/about.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); PandaDoc API key, sent as the Authorization header with an
  'API-Key ' prefix. Never logged.
- `attachment_id` (optional, string); Optional attachment_id used by detail streams.
- `base_url` (optional, string); default `https://api.pandadoc.com`; format `uri`; PandaDoc API root
  URL override for tests or proxies.
- `count` (optional, string); default `100`; Records per page (1-100).
- `document_id` (optional, string); Optional document_id used by detail streams.
- `id` (optional, string); Optional id used by detail streams.
- `item_uuid` (optional, string); Optional item_uuid used by detail streams.
- `mode` (optional, string).
- `section_id` (optional, string); Optional section_id used by detail streams.
- `session_request_id` (optional, string); Optional session_request_id used by detail streams.
- `task_id` (optional, string); Optional task_id used by detail streams.
- `template_id` (optional, string); Optional template_id used by detail streams.
- `upload_id` (optional, string); Optional upload_id used by detail streams.
- `user_id` (optional, string); Optional user_id used by detail streams.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.pandadoc.com`, `count=100`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/public/v1/documents` with query `count`=`1`; `page`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next`; next URLs stay
on the configured API host.

Pagination by stream: next_url: `documents`, `templates`, `contacts`, `documents_search`,
`content_library_items`, `documents_document_id_sections`, `documents_id_attachments`,
`documents_id_fields`, `documents_id_linked_objects`, `documents_folders`,
`documents_linked_objects`, `forms`, `logs`, `members`, `members_current`, `sms_opt_outs`,
`templates_folders`, `users`, `webhook_events`, `webhook_subscriptions`, `workspaces`,
`documents_document_id_audit_trail`, `logs_detail`, `notary_notaries`,
`notary_notarization_requests`, `product_catalog_items_search`; none:
`documents_document_id_ai_metadata`, `documents_document_id_content`,
`documents_document_id_docx_export_tasks_task_id`, `documents_document_id_summary`, `contacts_id`,
`content_library_items_id`, `content_library_items_id_details`,
`documents_document_id_auto_reminders`, `documents_document_id_auto_reminders_status`,
`documents_document_id_esign_disclosure`, `documents_document_id_sections_section_id`,
`documents_document_id_sections_uploads_upload_id`, `documents_id`,
`documents_id_attachments_attachment_id`, `documents_id_details`, `logs_id`, `members_id`,
`templates_id`, `templates_id_details`, `users_user_id`, `webhook_events_id`,
`webhook_subscriptions_id`, `documents_document_id_settings`, `logs_id_detail`,
`notary_notarization_requests_session_request_id`, `product_catalog_items_item_uuid`,
`templates_template_id_settings`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `documents`: GET `/public/v1/documents` - records path `results`; query `count`=`{{ config.count
  }}`; `page`=`1`; follows a next-page URL from the response body; URL path `next`; next URLs stay
  on the configured API host; incremental cursor `date_created`; formatted as `rfc3339`.
- `templates`: GET `/public/v1/templates` - records path `results`; query `count`=`{{ config.count
  }}`; `page`=`1`; follows a next-page URL from the response body; URL path `next`; next URLs stay
  on the configured API host; incremental cursor `date_created`; formatted as `rfc3339`.
- `contacts`: GET `/public/v1/contacts` - records path `results`; query `count`=`{{ config.count
  }}`; `page`=`1`; follows a next-page URL from the response body; URL path `next`; next URLs stay
  on the configured API host; incremental cursor `created_date`; formatted as `rfc3339`.
- `documents_document_id_ai_metadata`: GET `/public/beta/documents/{{ config.document_id
  }}/ai-metadata` - single-object response; records at response root; emits passthrough records.
- `documents_document_id_content`: GET `/public/beta/documents/{{ config.document_id }}/content` -
  single-object response; records at response root; emits passthrough records.
- `documents_document_id_docx_export_tasks_task_id`: GET `/public/beta/documents/{{
  config.document_id }}/docx-export-tasks/{{ config.task_id }}` - single-object response; records at
  response root; emits passthrough records.
- `documents_document_id_summary`: GET `/public/beta/documents/{{ config.document_id }}/summary` -
  single-object response; records at response root; emits passthrough records.
- `documents_search`: GET `/public/beta/documents/search` - records path `results`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host;
  emits passthrough records.
- `contacts_id`: GET `/public/v1/contacts/{{ config.id }}` - single-object response; records at
  response root; emits passthrough records.
- `content_library_items`: GET `/public/v1/content-library-items` - records path `results`; follows
  a next-page URL from the response body; URL path `next`; next URLs stay on the configured API
  host; emits passthrough records.
- `content_library_items_id`: GET `/public/v1/content-library-items/{{ config.id }}` - single-object
  response; records at response root; emits passthrough records.
- `content_library_items_id_details`: GET `/public/v1/content-library-items/{{ config.id }}/details`
  - single-object response; records at response root; emits passthrough records.
- `documents_document_id_auto_reminders`: GET `/public/v1/documents/{{ config.document_id
  }}/auto-reminders` - single-object response; records at response root; emits passthrough records.
- `documents_document_id_auto_reminders_status`: GET `/public/v1/documents/{{ config.document_id
  }}/auto-reminders/status` - single-object response; records at response root; emits passthrough
  records.
- `documents_document_id_esign_disclosure`: GET `/public/v1/documents/{{ config.document_id
  }}/esign-disclosure` - single-object response; records at response root; emits passthrough
  records.
- `documents_document_id_sections`: GET `/public/v1/documents/{{ config.document_id }}/sections` -
  records path `results`; follows a next-page URL from the response body; URL path `next`; next URLs
  stay on the configured API host; emits passthrough records.
- `documents_document_id_sections_section_id`: GET `/public/v1/documents/{{ config.document_id
  }}/sections/{{ config.section_id }}` - single-object response; records at response root; emits
  passthrough records.
- `documents_document_id_sections_uploads_upload_id`: GET `/public/v1/documents/{{
  config.document_id }}/sections/uploads/{{ config.upload_id }}` - single-object response; records
  at response root; emits passthrough records.
- `documents_id`: GET `/public/v1/documents/{{ config.id }}` - single-object response; records at
  response root; emits passthrough records.
- `documents_id_attachments`: GET `/public/v1/documents/{{ config.id }}/attachments` - records path
  `results`; follows a next-page URL from the response body; URL path `next`; next URLs stay on the
  configured API host; emits passthrough records.
- `documents_id_attachments_attachment_id`: GET `/public/v1/documents/{{ config.id }}/attachments/{{
  config.attachment_id }}` - single-object response; records at response root; emits passthrough
  records.
- `documents_id_details`: GET `/public/v1/documents/{{ config.id }}/details` - single-object
  response; records at response root; emits passthrough records.
- `documents_id_fields`: GET `/public/v1/documents/{{ config.id }}/fields` - records path `results`;
  follows a next-page URL from the response body; URL path `next`; next URLs stay on the configured
  API host; emits passthrough records.
- `documents_id_linked_objects`: GET `/public/v1/documents/{{ config.id }}/linked-objects` - records
  path `results`; follows a next-page URL from the response body; URL path `next`; next URLs stay on
  the configured API host; emits passthrough records.
- `documents_folders`: GET `/public/v1/documents/folders` - records path `results`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host;
  emits passthrough records.
- `documents_linked_objects`: GET `/public/v1/documents/linked-objects` - records path `results`;
  follows a next-page URL from the response body; URL path `next`; next URLs stay on the configured
  API host; emits passthrough records.
- `forms`: GET `/public/v1/forms` - records path `results`; follows a next-page URL from the
  response body; URL path `next`; next URLs stay on the configured API host; emits passthrough
  records.
- `logs`: GET `/public/v1/logs` - records path `results`; follows a next-page URL from the response
  body; URL path `next`; next URLs stay on the configured API host; emits passthrough records.
- `logs_id`: GET `/public/v1/logs/{{ config.id }}` - single-object response; records at response
  root; emits passthrough records.
- `members`: GET `/public/v1/members` - records path `results`; follows a next-page URL from the
  response body; URL path `next`; next URLs stay on the configured API host; emits passthrough
  records.
- `members_id`: GET `/public/v1/members/{{ config.id }}` - single-object response; records at
  response root; emits passthrough records.
- `members_current`: GET `/public/v1/members/current` - records path `results`; follows a next-page
  URL from the response body; URL path `next`; next URLs stay on the configured API host; emits
  passthrough records.
- `sms_opt_outs`: GET `/public/v1/sms-opt-outs` - records path `results`; follows a next-page URL
  from the response body; URL path `next`; next URLs stay on the configured API host; emits
  passthrough records.
- `templates_id`: GET `/public/v1/templates/{{ config.id }}` - single-object response; records at
  response root; emits passthrough records.
- `templates_id_details`: GET `/public/v1/templates/{{ config.id }}/details` - single-object
  response; records at response root; emits passthrough records.
- `templates_folders`: GET `/public/v1/templates/folders` - records path `results`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host;
  emits passthrough records.
- `users`: GET `/public/v1/users` - records path `results`; follows a next-page URL from the
  response body; URL path `next`; next URLs stay on the configured API host; emits passthrough
  records.
- `users_user_id`: GET `/public/v1/users/{{ config.user_id }}` - single-object response; records at
  response root; emits passthrough records.
- `webhook_events`: GET `/public/v1/webhook-events` - records path `results`; follows a next-page
  URL from the response body; URL path `next`; next URLs stay on the configured API host; emits
  passthrough records.
- `webhook_events_id`: GET `/public/v1/webhook-events/{{ config.id }}` - single-object response;
  records at response root; emits passthrough records.
- `webhook_subscriptions`: GET `/public/v1/webhook-subscriptions` - records path `results`; follows
  a next-page URL from the response body; URL path `next`; next URLs stay on the configured API
  host; emits passthrough records.
- `webhook_subscriptions_id`: GET `/public/v1/webhook-subscriptions/{{ config.id }}` - single-object
  response; records at response root; emits passthrough records.
- `workspaces`: GET `/public/v1/workspaces` - records path `results`; follows a next-page URL from
  the response body; URL path `next`; next URLs stay on the configured API host; emits passthrough
  records.
- `documents_document_id_audit_trail`: GET `/public/v2/documents/{{ config.document_id
  }}/audit-trail` - records path `results`; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host; emits passthrough records.
- `documents_document_id_settings`: GET `/public/v2/documents/{{ config.document_id }}/settings` -
  single-object response; records at response root; emits passthrough records.
- `logs_detail`: GET `/public/v2/logs` - records path `results`; follows a next-page URL from the
  response body; URL path `next`; next URLs stay on the configured API host; emits passthrough
  records.
- `logs_id_detail`: GET `/public/v2/logs/{{ config.id }}` - single-object response; records at
  response root; emits passthrough records.
- `notary_notaries`: GET `/public/v2/notary/notaries` - records path `results`; follows a next-page
  URL from the response body; URL path `next`; next URLs stay on the configured API host; emits
  passthrough records.
- `notary_notarization_requests`: GET `/public/v2/notary/notarization-requests` - records path
  `results`; follows a next-page URL from the response body; URL path `next`; next URLs stay on the
  configured API host; emits passthrough records.
- `notary_notarization_requests_session_request_id`: GET `/public/v2/notary/notarization-requests/{{
  config.session_request_id }}` - single-object response; records at response root; emits
  passthrough records.
- `product_catalog_items_item_uuid`: GET `/public/v2/product-catalog/items/{{ config.item_uuid }}` -
  single-object response; records at response root; emits passthrough records.
- `product_catalog_items_search`: GET `/public/v2/product-catalog/items/search` - records path
  `results`; follows a next-page URL from the response body; URL path `next`; next URLs stay on the
  configured API host; emits passthrough records.
- `templates_template_id_settings`: GET `/public/v2/templates/{{ config.template_id }}/settings` -
  single-object response; records at response root; emits passthrough records.

## Write actions & risks

Overall write risk: external PandaDoc API mutations including document sending/status changes,
contact/template/folder/webhook/admin updates, and destructive deletes.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `post_documents_document_id_docx_export_tasks`: POST `/public/beta/documents/{{ record.document_id
  }}/docx-export-tasks` - kind `create`; body type `none`; path fields `document_id`; required
  record fields `document_id`; accepted fields `document_id`; risk: POST
  /public/beta/documents/{document_id}/docx-export-tasks mutates PandaDoc data or workflow state.
- `post_documents_ai_metadata`: POST `/public/beta/documents/ai-metadata` - kind `create`; body type
  `json`; risk: POST /public/beta/documents/ai-metadata mutates PandaDoc data or workflow state.
- `post_contacts`: POST `/public/v1/contacts` - kind `create`; body type `json`; risk: POST
  /public/v1/contacts mutates PandaDoc data or workflow state.
- `delete_contacts_id`: DELETE `/public/v1/contacts/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes PandaDoc resource
  via /public/v1/contacts/{id}; destructive external mutation.
- `patch_contacts_id`: PATCH `/public/v1/contacts/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /public/v1/contacts/{id} mutates PandaDoc data or workflow state.
- `post_content_library_items`: POST `/public/v1/content-library-items` - kind `create`; body type
  `json`; risk: POST /public/v1/content-library-items mutates PandaDoc data or workflow state.
- `delete_documents`: DELETE `/public/v1/documents` - kind `delete`; body type `none`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: deletes PandaDoc
  resource via /public/v1/documents; destructive external mutation.
- `post_documents`: POST `/public/v1/documents` - kind `create`; body type `json`; risk: POST
  /public/v1/documents mutates PandaDoc data or workflow state.
- `patch_documents_document_id_auto_reminders`: PATCH `/public/v1/documents/{{ record.document_id
  }}/auto-reminders` - kind `update`; body type `json`; path fields `document_id`; required record
  fields `document_id`; accepted fields `document_id`; risk: PATCH
  /public/v1/documents/{document_id}/auto-reminders mutates PandaDoc data or workflow state.
- `put_documents_document_id_quotes_quote_id`: PUT `/public/v1/documents/{{ record.document_id
  }}/quotes/{{ record.quote_id }}` - kind `update`; body type `json`; path fields `document_id`,
  `quote_id`; required record fields `document_id`, `quote_id`; accepted fields `document_id`,
  `quote_id`; risk: PUT /public/v1/documents/{document_id}/quotes/{quote_id} mutates PandaDoc data
  or workflow state.
- `post_documents_document_id_sections`: POST `/public/v1/documents/{{ record.document_id
  }}/sections` - kind `create`; body type `json`; path fields `document_id`; required record fields
  `document_id`; accepted fields `document_id`; risk: POST
  /public/v1/documents/{document_id}/sections mutates PandaDoc data or workflow state.
- `delete_documents_document_id_sections_section_id`: DELETE `/public/v1/documents/{{
  record.document_id }}/sections/{{ record.section_id }}` - kind `delete`; body type `none`; path
  fields `document_id`, `section_id`; required record fields `document_id`, `section_id`; accepted
  fields `document_id`, `section_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes PandaDoc resource via
  /public/v1/documents/{document_id}/sections/{section_id}; destructive external mutation.
- `post_documents_document_id_send_reminder`: POST `/public/v1/documents/{{ record.document_id
  }}/send-reminder` - kind `create`; body type `json`; path fields `document_id`; required record
  fields `document_id`; accepted fields `document_id`; risk: POST
  /public/v1/documents/{document_id}/send-reminder mutates PandaDoc data or workflow state.
- `delete_documents_id`: DELETE `/public/v1/documents/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes PandaDoc resource
  via /public/v1/documents/{id}; destructive external mutation.
- `patch_documents_id`: PATCH `/public/v1/documents/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /public/v1/documents/{id} mutates PandaDoc data or workflow state.
- `post_documents_id_append_content_library_item`: POST `/public/v1/documents/{{ record.id
  }}/append-content-library-item` - kind `create`; body type `json`; path fields `id`; required
  record fields `id`; accepted fields `id`; risk: POST
  /public/v1/documents/{id}/append-content-library-item mutates PandaDoc data or workflow state.
- `post_documents_id_attachments`: POST `/public/v1/documents/{{ record.id }}/attachments` - kind
  `create`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: POST /public/v1/documents/{id}/attachments mutates PandaDoc data or workflow state.
- `delete_documents_id_attachments_attachment_id`: DELETE `/public/v1/documents/{{ record.id
  }}/attachments/{{ record.attachment_id }}` - kind `delete`; body type `none`; path fields `id`,
  `attachment_id`; required record fields `id`, `attachment_id`; accepted fields `attachment_id`,
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  deletes PandaDoc resource via /public/v1/documents/{id}/attachments/{attachment_id}; destructive
  external mutation.
- `post_documents_id_draft`: POST `/public/v1/documents/{{ record.id }}/draft` - kind `create`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /public/v1/documents/{id}/draft mutates PandaDoc data or workflow state.
- `post_documents_id_editing_sessions`: POST `/public/v1/documents/{{ record.id }}/editing-sessions`
  - kind `create`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: POST /public/v1/documents/{id}/editing-sessions mutates PandaDoc data or workflow
  state.
- `patch_documents_id_fields`: PATCH `/public/v1/documents/{{ record.id }}/fields` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /public/v1/documents/{id}/fields mutates PandaDoc data or workflow state.
- `post_documents_id_fields`: POST `/public/v1/documents/{{ record.id }}/fields` - kind `create`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /public/v1/documents/{id}/fields mutates PandaDoc data or workflow state.
- `post_documents_id_linked_objects`: POST `/public/v1/documents/{{ record.id }}/linked-objects` -
  kind `create`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: POST /public/v1/documents/{id}/linked-objects mutates PandaDoc data or workflow state.
- `delete_documents_id_linked_objects_linked_object_id`: DELETE `/public/v1/documents/{{ record.id
  }}/linked-objects/{{ record.linked_object_id }}` - kind `delete`; body type `none`; path fields
  `id`, `linked_object_id`; required record fields `id`, `linked_object_id`; accepted fields `id`,
  `linked_object_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes PandaDoc resource via
  /public/v1/documents/{id}/linked-objects/{linked_object_id}; destructive external mutation.
- `post_documents_id_move_to_folder_folder_id`: POST `/public/v1/documents/{{ record.id
  }}/move-to-folder/{{ record.folder_id }}` - kind `create`; body type `none`; path fields `id`,
  `folder_id`; required record fields `id`, `folder_id`; accepted fields `folder_id`, `id`; risk:
  POST /public/v1/documents/{id}/move-to-folder/{folder_id} mutates PandaDoc data or workflow state.
- `patch_documents_id_ownership`: PATCH `/public/v1/documents/{{ record.id }}/ownership` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: PATCH /public/v1/documents/{id}/ownership mutates PandaDoc data or workflow state.
- `post_documents_id_recipients`: POST `/public/v1/documents/{{ record.id }}/recipients` - kind
  `create`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: POST /public/v1/documents/{id}/recipients mutates PandaDoc data or workflow state.
- `delete_documents_id_recipients_recipient_id`: DELETE `/public/v1/documents/{{ record.id
  }}/recipients/{{ record.recipient_id }}` - kind `delete`; body type `none`; path fields `id`,
  `recipient_id`; required record fields `id`, `recipient_id`; accepted fields `id`, `recipient_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: deletes
  PandaDoc resource via /public/v1/documents/{id}/recipients/{recipient_id}; destructive external
  mutation.
- `post_documents_id_recipients_recipient_id_reassign`: POST `/public/v1/documents/{{ record.id
  }}/recipients/{{ record.recipient_id }}/reassign` - kind `create`; body type `json`; path fields
  `id`, `recipient_id`; required record fields `id`, `recipient_id`; accepted fields `id`,
  `recipient_id`; risk: POST /public/v1/documents/{id}/recipients/{recipient_id}/reassign mutates
  PandaDoc data or workflow state.
- `patch_documents_id_recipients_recipient_recipient_id`: PATCH `/public/v1/documents/{{ record.id
  }}/recipients/recipient/{{ record.recipient_id }}` - kind `update`; body type `json`; path fields
  `id`, `recipient_id`; required record fields `id`, `recipient_id`; accepted fields `id`,
  `recipient_id`; risk: PATCH /public/v1/documents/{id}/recipients/recipient/{recipient_id} mutates
  PandaDoc data or workflow state.
- `post_documents_id_send`: POST `/public/v1/documents/{{ record.id }}/send` - kind `create`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /public/v1/documents/{id}/send mutates PandaDoc data or workflow state.
- `post_documents_id_session`: POST `/public/v1/documents/{{ record.id }}/session` - kind `create`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: POST
  /public/v1/documents/{id}/session mutates PandaDoc data or workflow state.
- `patch_documents_id_status`: PATCH `/public/v1/documents/{{ record.id }}/status` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /public/v1/documents/{id}/status mutates PandaDoc data or workflow state.
- `post_documents_folders`: POST `/public/v1/documents/folders` - kind `create`; body type `json`;
  risk: POST /public/v1/documents/folders mutates PandaDoc data or workflow state.
- `put_documents_folders_id`: PUT `/public/v1/documents/folders/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PUT
  /public/v1/documents/folders/{id} mutates PandaDoc data or workflow state.
- `patch_documents_ownership`: PATCH `/public/v1/documents/ownership` - kind `update`; body type
  `json`; risk: PATCH /public/v1/documents/ownership mutates PandaDoc data or workflow state.
- `post_templates`: POST `/public/v1/templates` - kind `create`; body type `json`; risk: POST
  /public/v1/templates mutates PandaDoc data or workflow state.
- `delete_templates_id`: DELETE `/public/v1/templates/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes PandaDoc resource
  via /public/v1/templates/{id}; destructive external mutation.
- `patch_templates_id`: PATCH `/public/v1/templates/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /public/v1/templates/{id} mutates PandaDoc data or workflow state.
- `post_templates_id_editing_sessions`: POST `/public/v1/templates/{{ record.id }}/editing-sessions`
  - kind `create`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: POST /public/v1/templates/{id}/editing-sessions mutates PandaDoc data or workflow
  state.
- `post_templates_folders`: POST `/public/v1/templates/folders` - kind `create`; body type `json`;
  risk: POST /public/v1/templates/folders mutates PandaDoc data or workflow state.
- `put_templates_folders_id`: PUT `/public/v1/templates/folders/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PUT
  /public/v1/templates/folders/{id} mutates PandaDoc data or workflow state.
- `post_users`: POST `/public/v1/users` - kind `create`; body type `json`; risk: POST
  /public/v1/users mutates PandaDoc data or workflow state.
- `post_webhook_subscriptions`: POST `/public/v1/webhook-subscriptions` - kind `create`; body type
  `json`; risk: POST /public/v1/webhook-subscriptions mutates PandaDoc data or workflow state.
- `delete_webhook_subscriptions_id`: DELETE `/public/v1/webhook-subscriptions/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  deletes PandaDoc resource via /public/v1/webhook-subscriptions/{id}; destructive external
  mutation.
- `patch_webhook_subscriptions_id`: PATCH `/public/v1/webhook-subscriptions/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: PATCH /public/v1/webhook-subscriptions/{id} mutates PandaDoc data or workflow state.
- `post_workspaces`: POST `/public/v1/workspaces` - kind `create`; body type `json`; risk: POST
  /public/v1/workspaces mutates PandaDoc data or workflow state.
- `post_workspaces_workspace_id_deactivate`: POST `/public/v1/workspaces/{{ record.workspace_id
  }}/deactivate` - kind `create`; body type `json`; path fields `workspace_id`; required record
  fields `workspace_id`; accepted fields `workspace_id`; risk: POST
  /public/v1/workspaces/{workspace_id}/deactivate mutates PandaDoc data or workflow state.
- `post_workspaces_workspace_id_members`: POST `/public/v1/workspaces/{{ record.workspace_id
  }}/members` - kind `create`; body type `json`; path fields `workspace_id`; required record fields
  `workspace_id`; accepted fields `workspace_id`; risk: POST
  /public/v1/workspaces/{workspace_id}/members mutates PandaDoc data or workflow state.
- `delete_workspaces_workspace_id_members_member_id`: DELETE `/public/v1/workspaces/{{
  record.workspace_id }}/members/{{ record.member_id }}` - kind `delete`; body type `none`; path
  fields `workspace_id`, `member_id`; required record fields `workspace_id`, `member_id`; accepted
  fields `member_id`, `workspace_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes PandaDoc resource via
  /public/v1/workspaces/{workspace_id}/members/{member_id}; destructive external mutation.
- `patch_workspaces_workspace_id_members_member_id_role`: PATCH `/public/v1/workspaces/{{
  record.workspace_id }}/members/{{ record.member_id }}/role` - kind `update`; body type `json`;
  path fields `workspace_id`, `member_id`; required record fields `workspace_id`, `member_id`;
  accepted fields `member_id`, `workspace_id`; risk: PATCH
  /public/v1/workspaces/{workspace_id}/members/{member_id}/role mutates PandaDoc data or workflow
  state.
- `patch_documents_document_id_settings`: PATCH `/public/v2/documents/{{ record.document_id
  }}/settings` - kind `update`; body type `json`; path fields `document_id`; required record fields
  `document_id`; accepted fields `document_id`; risk: PATCH
  /public/v2/documents/{document_id}/settings mutates PandaDoc data or workflow state.
- `post_notary_notarization_requests`: POST `/public/v2/notary/notarization-requests` - kind
  `create`; body type `json`; risk: POST /public/v2/notary/notarization-requests mutates PandaDoc
  data or workflow state.
- `delete_notary_notarization_requests_session_request_id`: DELETE
  `/public/v2/notary/notarization-requests/{{ record.session_request_id }}` - kind `delete`; body
  type `none`; path fields `session_request_id`; required record fields `session_request_id`;
  accepted fields `session_request_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes PandaDoc resource via
  /public/v2/notary/notarization-requests/{session_request_id}; destructive external mutation.
- `post_product_catalog_items`: POST `/public/v2/product-catalog/items` - kind `create`; body type
  `json`; risk: POST /public/v2/product-catalog/items mutates PandaDoc data or workflow state.
- `delete_product_catalog_items_item_uuid`: DELETE `/public/v2/product-catalog/items/{{
  record.item_uuid }}` - kind `delete`; body type `none`; path fields `item_uuid`; required record
  fields `item_uuid`; accepted fields `item_uuid`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: deletes PandaDoc resource via
  /public/v2/product-catalog/items/{item_uuid}; destructive external mutation.
- `patch_product_catalog_items_item_uuid`: PATCH `/public/v2/product-catalog/items/{{
  record.item_uuid }}` - kind `update`; body type `json`; path fields `item_uuid`; required record
  fields `item_uuid`; accepted fields `item_uuid`; risk: PATCH
  /public/v2/product-catalog/items/{item_uuid} mutates PandaDoc data or workflow state.
- `patch_templates_template_id_settings`: PATCH `/public/v2/templates/{{ record.template_id
  }}/settings` - kind `update`; body type `json`; path fields `template_id`; required record fields
  `template_id`; accepted fields `template_id`; risk: PATCH
  /public/v2/templates/{template_id}/settings mutates PandaDoc data or workflow state.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 53 stream-backed endpoint group(s), 58 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=9, non_data_endpoint=1, requires_elevated_scope=3.
