# pm connectors inspect pandadoc

```text
NAME
  pm connectors inspect pandadoc - PandaDoc connector manual

SYNOPSIS
  pm connectors inspect pandadoc
  pm connectors inspect pandadoc --json
  pm credentials add <name> --connector pandadoc [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes documented PandaDoc public API resources across documents, templates, contacts, folders, forms, logs, members, webhooks, workspaces, notary, and catalog surfaces.

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
  attachment_id
  base_url
  count
  document_id
  id
  item_uuid
  mode
  section_id
  session_request_id
  task_id
  template_id
  upload_id
  user_id
  api_key (secret)

ETL STREAMS
  documents:
    primary key: id
    cursor: date_created
    fields: date_created(string), date_modified(string), id(string), name(string), status(string)
  templates:
    primary key: id
    cursor: date_created
    fields: date_created(string), date_modified(string), id(string), name(string)
  contacts:
    primary key: id
    cursor: created_date
    fields: created_date(string), email(string), first_name(string), id(string), last_name(string)
  documents_document_id_ai_metadata:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_content:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_docx_export_tasks_task_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_summary:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_search:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  contacts_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  content_library_items:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  content_library_items_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  content_library_items_id_details:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_auto_reminders:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_auto_reminders_status:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_esign_disclosure:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_sections:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_sections_section_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_sections_uploads_upload_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_id_attachments:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_id_attachments_attachment_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_id_details:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_id_fields:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_id_linked_objects:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_folders:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_linked_objects:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  forms:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  logs:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  logs_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  members:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  members_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  members_current:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  sms_opt_outs:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  templates_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  templates_id_details:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  templates_folders:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  users:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  users_user_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  webhook_events:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  webhook_events_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  webhook_subscriptions:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  webhook_subscriptions_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  workspaces:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_audit_trail:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  documents_document_id_settings:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  logs_detail:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  logs_id_detail:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  notary_notaries:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  notary_notarization_requests:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  notary_notarization_requests_session_request_id:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  product_catalog_items_item_uuid:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  product_catalog_items_search:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)
  templates_template_id_settings:
    fields: date_created(string), date_modified(string), id(string), name(string), status(string), uuid(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  post_documents_document_id_docx_export_tasks:
    endpoint: POST /public/beta/documents/{{ record.document_id }}/docx-export-tasks
    required fields: document_id
    risk: POST /public/beta/documents/{document_id}/docx-export-tasks mutates PandaDoc data or workflow state
  post_documents_ai_metadata:
    endpoint: POST /public/beta/documents/ai-metadata
    risk: POST /public/beta/documents/ai-metadata mutates PandaDoc data or workflow state
  post_contacts:
    endpoint: POST /public/v1/contacts
    risk: POST /public/v1/contacts mutates PandaDoc data or workflow state
  delete_contacts_id:
    endpoint: DELETE /public/v1/contacts/{{ record.id }}
    required fields: id
    risk: deletes PandaDoc resource via /public/v1/contacts/{id}; destructive external mutation
  patch_contacts_id:
    endpoint: PATCH /public/v1/contacts/{{ record.id }}
    required fields: id
    risk: PATCH /public/v1/contacts/{id} mutates PandaDoc data or workflow state
  post_content_library_items:
    endpoint: POST /public/v1/content-library-items
    risk: POST /public/v1/content-library-items mutates PandaDoc data or workflow state
  delete_documents:
    endpoint: DELETE /public/v1/documents
    risk: deletes PandaDoc resource via /public/v1/documents; destructive external mutation
  post_documents:
    endpoint: POST /public/v1/documents
    risk: POST /public/v1/documents mutates PandaDoc data or workflow state
  patch_documents_document_id_auto_reminders:
    endpoint: PATCH /public/v1/documents/{{ record.document_id }}/auto-reminders
    required fields: document_id
    risk: PATCH /public/v1/documents/{document_id}/auto-reminders mutates PandaDoc data or workflow state
  put_documents_document_id_quotes_quote_id:
    endpoint: PUT /public/v1/documents/{{ record.document_id }}/quotes/{{ record.quote_id }}
    required fields: document_id, quote_id
    risk: PUT /public/v1/documents/{document_id}/quotes/{quote_id} mutates PandaDoc data or workflow state
  post_documents_document_id_sections:
    endpoint: POST /public/v1/documents/{{ record.document_id }}/sections
    required fields: document_id
    risk: POST /public/v1/documents/{document_id}/sections mutates PandaDoc data or workflow state
  delete_documents_document_id_sections_section_id:
    endpoint: DELETE /public/v1/documents/{{ record.document_id }}/sections/{{ record.section_id }}
    required fields: document_id, section_id
    risk: deletes PandaDoc resource via /public/v1/documents/{document_id}/sections/{section_id}; destructive external mutation
  post_documents_document_id_send_reminder:
    endpoint: POST /public/v1/documents/{{ record.document_id }}/send-reminder
    required fields: document_id
    risk: POST /public/v1/documents/{document_id}/send-reminder mutates PandaDoc data or workflow state
  delete_documents_id:
    endpoint: DELETE /public/v1/documents/{{ record.id }}
    required fields: id
    risk: deletes PandaDoc resource via /public/v1/documents/{id}; destructive external mutation
  patch_documents_id:
    endpoint: PATCH /public/v1/documents/{{ record.id }}
    required fields: id
    risk: PATCH /public/v1/documents/{id} mutates PandaDoc data or workflow state
  post_documents_id_append_content_library_item:
    endpoint: POST /public/v1/documents/{{ record.id }}/append-content-library-item
    required fields: id
    risk: POST /public/v1/documents/{id}/append-content-library-item mutates PandaDoc data or workflow state
  post_documents_id_attachments:
    endpoint: POST /public/v1/documents/{{ record.id }}/attachments
    required fields: id
    risk: POST /public/v1/documents/{id}/attachments mutates PandaDoc data or workflow state
  delete_documents_id_attachments_attachment_id:
    endpoint: DELETE /public/v1/documents/{{ record.id }}/attachments/{{ record.attachment_id }}
    required fields: id, attachment_id
    risk: deletes PandaDoc resource via /public/v1/documents/{id}/attachments/{attachment_id}; destructive external mutation
  post_documents_id_draft:
    endpoint: POST /public/v1/documents/{{ record.id }}/draft
    required fields: id
    risk: POST /public/v1/documents/{id}/draft mutates PandaDoc data or workflow state
  post_documents_id_editing_sessions:
    endpoint: POST /public/v1/documents/{{ record.id }}/editing-sessions
    required fields: id
    risk: POST /public/v1/documents/{id}/editing-sessions mutates PandaDoc data or workflow state
  patch_documents_id_fields:
    endpoint: PATCH /public/v1/documents/{{ record.id }}/fields
    required fields: id
    risk: PATCH /public/v1/documents/{id}/fields mutates PandaDoc data or workflow state
  post_documents_id_fields:
    endpoint: POST /public/v1/documents/{{ record.id }}/fields
    required fields: id
    risk: POST /public/v1/documents/{id}/fields mutates PandaDoc data or workflow state
  post_documents_id_linked_objects:
    endpoint: POST /public/v1/documents/{{ record.id }}/linked-objects
    required fields: id
    risk: POST /public/v1/documents/{id}/linked-objects mutates PandaDoc data or workflow state
  delete_documents_id_linked_objects_linked_object_id:
    endpoint: DELETE /public/v1/documents/{{ record.id }}/linked-objects/{{ record.linked_object_id }}
    required fields: id, linked_object_id
    risk: deletes PandaDoc resource via /public/v1/documents/{id}/linked-objects/{linked_object_id}; destructive external mutation
  post_documents_id_move_to_folder_folder_id:
    endpoint: POST /public/v1/documents/{{ record.id }}/move-to-folder/{{ record.folder_id }}
    required fields: id, folder_id
    risk: POST /public/v1/documents/{id}/move-to-folder/{folder_id} mutates PandaDoc data or workflow state
  patch_documents_id_ownership:
    endpoint: PATCH /public/v1/documents/{{ record.id }}/ownership
    required fields: id
    risk: PATCH /public/v1/documents/{id}/ownership mutates PandaDoc data or workflow state
  post_documents_id_recipients:
    endpoint: POST /public/v1/documents/{{ record.id }}/recipients
    required fields: id
    risk: POST /public/v1/documents/{id}/recipients mutates PandaDoc data or workflow state
  delete_documents_id_recipients_recipient_id:
    endpoint: DELETE /public/v1/documents/{{ record.id }}/recipients/{{ record.recipient_id }}
    required fields: id, recipient_id
    risk: deletes PandaDoc resource via /public/v1/documents/{id}/recipients/{recipient_id}; destructive external mutation
  post_documents_id_recipients_recipient_id_reassign:
    endpoint: POST /public/v1/documents/{{ record.id }}/recipients/{{ record.recipient_id }}/reassign
    required fields: id, recipient_id
    risk: POST /public/v1/documents/{id}/recipients/{recipient_id}/reassign mutates PandaDoc data or workflow state
  patch_documents_id_recipients_recipient_recipient_id:
    endpoint: PATCH /public/v1/documents/{{ record.id }}/recipients/recipient/{{ record.recipient_id }}
    required fields: id, recipient_id
    risk: PATCH /public/v1/documents/{id}/recipients/recipient/{recipient_id} mutates PandaDoc data or workflow state
  post_documents_id_send:
    endpoint: POST /public/v1/documents/{{ record.id }}/send
    required fields: id
    risk: POST /public/v1/documents/{id}/send mutates PandaDoc data or workflow state
  post_documents_id_session:
    endpoint: POST /public/v1/documents/{{ record.id }}/session
    required fields: id
    risk: POST /public/v1/documents/{id}/session mutates PandaDoc data or workflow state
  patch_documents_id_status:
    endpoint: PATCH /public/v1/documents/{{ record.id }}/status
    required fields: id
    risk: PATCH /public/v1/documents/{id}/status mutates PandaDoc data or workflow state
  post_documents_folders:
    endpoint: POST /public/v1/documents/folders
    risk: POST /public/v1/documents/folders mutates PandaDoc data or workflow state
  put_documents_folders_id:
    endpoint: PUT /public/v1/documents/folders/{{ record.id }}
    required fields: id
    risk: PUT /public/v1/documents/folders/{id} mutates PandaDoc data or workflow state
  patch_documents_ownership:
    endpoint: PATCH /public/v1/documents/ownership
    risk: PATCH /public/v1/documents/ownership mutates PandaDoc data or workflow state
  post_templates:
    endpoint: POST /public/v1/templates
    risk: POST /public/v1/templates mutates PandaDoc data or workflow state
  delete_templates_id:
    endpoint: DELETE /public/v1/templates/{{ record.id }}
    required fields: id
    risk: deletes PandaDoc resource via /public/v1/templates/{id}; destructive external mutation
  patch_templates_id:
    endpoint: PATCH /public/v1/templates/{{ record.id }}
    required fields: id
    risk: PATCH /public/v1/templates/{id} mutates PandaDoc data or workflow state
  post_templates_id_editing_sessions:
    endpoint: POST /public/v1/templates/{{ record.id }}/editing-sessions
    required fields: id
    risk: POST /public/v1/templates/{id}/editing-sessions mutates PandaDoc data or workflow state
  post_templates_folders:
    endpoint: POST /public/v1/templates/folders
    risk: POST /public/v1/templates/folders mutates PandaDoc data or workflow state
  put_templates_folders_id:
    endpoint: PUT /public/v1/templates/folders/{{ record.id }}
    required fields: id
    risk: PUT /public/v1/templates/folders/{id} mutates PandaDoc data or workflow state
  post_users:
    endpoint: POST /public/v1/users
    risk: POST /public/v1/users mutates PandaDoc data or workflow state
  post_webhook_subscriptions:
    endpoint: POST /public/v1/webhook-subscriptions
    risk: POST /public/v1/webhook-subscriptions mutates PandaDoc data or workflow state
  delete_webhook_subscriptions_id:
    endpoint: DELETE /public/v1/webhook-subscriptions/{{ record.id }}
    required fields: id
    risk: deletes PandaDoc resource via /public/v1/webhook-subscriptions/{id}; destructive external mutation
  patch_webhook_subscriptions_id:
    endpoint: PATCH /public/v1/webhook-subscriptions/{{ record.id }}
    required fields: id
    risk: PATCH /public/v1/webhook-subscriptions/{id} mutates PandaDoc data or workflow state
  post_workspaces:
    endpoint: POST /public/v1/workspaces
    risk: POST /public/v1/workspaces mutates PandaDoc data or workflow state
  post_workspaces_workspace_id_deactivate:
    endpoint: POST /public/v1/workspaces/{{ record.workspace_id }}/deactivate
    required fields: workspace_id
    risk: POST /public/v1/workspaces/{workspace_id}/deactivate mutates PandaDoc data or workflow state
  post_workspaces_workspace_id_members:
    endpoint: POST /public/v1/workspaces/{{ record.workspace_id }}/members
    required fields: workspace_id
    risk: POST /public/v1/workspaces/{workspace_id}/members mutates PandaDoc data or workflow state
  delete_workspaces_workspace_id_members_member_id:
    endpoint: DELETE /public/v1/workspaces/{{ record.workspace_id }}/members/{{ record.member_id }}
    required fields: workspace_id, member_id
    risk: deletes PandaDoc resource via /public/v1/workspaces/{workspace_id}/members/{member_id}; destructive external mutation
  patch_workspaces_workspace_id_members_member_id_role:
    endpoint: PATCH /public/v1/workspaces/{{ record.workspace_id }}/members/{{ record.member_id }}/role
    required fields: workspace_id, member_id
    risk: PATCH /public/v1/workspaces/{workspace_id}/members/{member_id}/role mutates PandaDoc data or workflow state
  patch_documents_document_id_settings:
    endpoint: PATCH /public/v2/documents/{{ record.document_id }}/settings
    required fields: document_id
    risk: PATCH /public/v2/documents/{document_id}/settings mutates PandaDoc data or workflow state
  post_notary_notarization_requests:
    endpoint: POST /public/v2/notary/notarization-requests
    risk: POST /public/v2/notary/notarization-requests mutates PandaDoc data or workflow state
  delete_notary_notarization_requests_session_request_id:
    endpoint: DELETE /public/v2/notary/notarization-requests/{{ record.session_request_id }}
    required fields: session_request_id
    risk: deletes PandaDoc resource via /public/v2/notary/notarization-requests/{session_request_id}; destructive external mutation
  post_product_catalog_items:
    endpoint: POST /public/v2/product-catalog/items
    risk: POST /public/v2/product-catalog/items mutates PandaDoc data or workflow state
  delete_product_catalog_items_item_uuid:
    endpoint: DELETE /public/v2/product-catalog/items/{{ record.item_uuid }}
    required fields: item_uuid
    risk: deletes PandaDoc resource via /public/v2/product-catalog/items/{item_uuid}; destructive external mutation
  patch_product_catalog_items_item_uuid:
    endpoint: PATCH /public/v2/product-catalog/items/{{ record.item_uuid }}
    required fields: item_uuid
    risk: PATCH /public/v2/product-catalog/items/{item_uuid} mutates PandaDoc data or workflow state
  patch_templates_template_id_settings:
    endpoint: PATCH /public/v2/templates/{{ record.template_id }}/settings
    required fields: template_id
    risk: PATCH /public/v2/templates/{template_id}/settings mutates PandaDoc data or workflow state

SECURITY
  read risk: external PandaDoc API read of document, template, contact, member, workspace, webhook, notary, and catalog data
  write risk: external PandaDoc API mutations including document sending/status changes, contact/template/folder/webhook/admin updates, and destructive deletes
  approval: required for write actions; destructive and notification-sending actions carry explicit risk metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run PandaDoc's declared streams and reverse-ETL actions.
  Usage: pm pandadoc <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api get public v1 documents id attachments attachment-id download - Documented GET /public/v1/documents/{id}/attachments/{attachment_id}/download (not implemented) [intent=direct_read availability=not_implemented operation=pandadoc.get.public-v1-documents-id-attachments-attachment-id-download]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get public v1 documents id download - Documented GET /public/v1/documents/{id}/download (not implemented) [intent=direct_read availability=not_implemented operation=pandadoc.get.public-v1-documents-id-download]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get public v1 documents id download-protected - Documented GET /public/v1/documents/{id}/download-protected (not implemented) [intent=direct_read availability=not_implemented operation=pandadoc.get.public-v1-documents-id-download-protected]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get public v1 templates template-id shares - Documented GET /public/v1/templates/{template_id}/shares (not implemented) [intent=direct_read availability=not_implemented operation=pandadoc.get.public-v1-templates-template-id-shares]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch public v1 documents id status-upload - Documented PATCH /public/v1/documents/{id}/status?upload (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.patch.public-v1-documents-id-status-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch public v1 webhook-subscriptions id shared-key - Documented PATCH /public/v1/webhook-subscriptions/{id}/shared-key (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.patch.public-v1-webhook-subscriptions-id-shared-key]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api post oauth2 access-token - Documented POST /oauth2/access_token (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.oauth2-access-token]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 content-library-items-upload - Documented POST /public/v1/content-library-items?upload (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-content-library-items-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 documents document-id sections uploads - Documented POST /public/v1/documents/{document_id}/sections/uploads (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-documents-document-id-sections-uploads]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 documents document-id sections uploads-upload - Documented POST /public/v1/documents/{document_id}/sections/uploads?upload (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-documents-document-id-sections-uploads-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 documents document-id sections-upload - Documented POST /public/v1/documents/{document_id}/sections?upload (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-documents-document-id-sections-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 documents id attachments-upload - Documented POST /public/v1/documents/{id}/attachments?upload (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-documents-id-attachments-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 documents-upload - Documented POST /public/v1/documents?upload (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-documents-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 documents-upload-markdown - Documented POST /public/v1/documents?upload-markdown (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-documents-upload-markdown]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 members member-id token - Documented POST /public/v1/members/{member_id}/token (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-members-member-id-token]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 templates id duplicate - Documented POST /public/v1/templates/{id}/duplicate (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-templates-id-duplicate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 templates-upload - Documented POST /public/v1/templates?upload (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-templates-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v1 workspaces workspace-id api-keys - Documented POST /public/v1/workspaces/{workspace_id}/api-keys (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v1-workspaces-workspace-id-api-keys]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post public v2 dsv document-id add-named-items - Documented POST /public/v2/dsv/{document_id}/add-named-items (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.post.public-v2-dsv-document-id-add-named-items]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put public v1 templates template-id shares - Documented PUT /public/v1/templates/{template_id}/shares (not implemented) [intent=direct_write availability=not_implemented operation=pandadoc.put.public-v1-templates-template-id-shares]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    contacts id list - Run the contacts id ETL stream [intent=etl availability=implemented stream=contacts_id]
    contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
    content library items id details list - Run the content library items id details ETL stream [intent=etl availability=implemented stream=content_library_items_id_details]
    content library items id list - Run the content library items id ETL stream [intent=etl availability=implemented stream=content_library_items_id]
    content library items list - Run the content library items ETL stream [intent=etl availability=implemented stream=content_library_items]
    delete contacts id apply - Plan and execute the delete contacts id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_contacts_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/contacts/{id}; destructive external mutation; flags: --id (required)
    delete documents apply - Plan and execute the delete documents reverse-ETL action [intent=reverse_etl availability=implemented write=delete_documents]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/documents; destructive external mutation
    delete documents document id sections section id apply - Plan and execute the delete documents document id sections section id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_documents_document_id_sections_section_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/documents/{document_id}/sections/{section_id}; destructive external mutation; flags: --document_id (required), --section_id (required)
    delete documents id apply - Plan and execute the delete documents id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_documents_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/documents/{id}; destructive external mutation; flags: --id (required)
    delete documents id attachments attachment id apply - Plan and execute the delete documents id attachments attachment id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_documents_id_attachments_attachment_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/documents/{id}/attachments/{attachment_id}; destructive external mutation; flags: --attachment_id (required), --id (required)
    delete documents id linked objects linked object id apply - Plan and execute the delete documents id linked objects linked object id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_documents_id_linked_objects_linked_object_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/documents/{id}/linked-objects/{linked_object_id}; destructive external mutation; flags: --id (required), --linked_object_id (required)
    delete documents id recipients recipient id apply - Plan and execute the delete documents id recipients recipient id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_documents_id_recipients_recipient_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/documents/{id}/recipients/{recipient_id}; destructive external mutation; flags: --id (required), --recipient_id (required)
    delete notary notarization requests session request id apply - Plan and execute the delete notary notarization requests session request id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_notary_notarization_requests_session_request_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v2/notary/notarization-requests/{session_request_id}; destructive external mutation; flags: --session_request_id (required)
    delete product catalog items item uuid apply - Plan and execute the delete product catalog items item uuid reverse-ETL action [intent=reverse_etl availability=implemented write=delete_product_catalog_items_item_uuid]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v2/product-catalog/items/{item_uuid}; destructive external mutation; flags: --item_uuid (required)
    delete templates id apply - Plan and execute the delete templates id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_templates_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/templates/{id}; destructive external mutation; flags: --id (required)
    delete webhook subscriptions id apply - Plan and execute the delete webhook subscriptions id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook_subscriptions_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/webhook-subscriptions/{id}; destructive external mutation; flags: --id (required)
    delete workspaces workspace id members member id apply - Plan and execute the delete workspaces workspace id members member id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_workspaces_workspace_id_members_member_id]; approval: requires plan, preview, approval, and execute; risk: deletes PandaDoc resource via /public/v1/workspaces/{workspace_id}/members/{member_id}; destructive external mutation; flags: --member_id (required), --workspace_id (required)
    documents document id ai metadata list - Run the documents document id ai metadata ETL stream [intent=etl availability=implemented stream=documents_document_id_ai_metadata]
    documents document id audit trail list - Run the documents document id audit trail ETL stream [intent=etl availability=implemented stream=documents_document_id_audit_trail]
    documents document id auto reminders list - Run the documents document id auto reminders ETL stream [intent=etl availability=implemented stream=documents_document_id_auto_reminders]
    documents document id auto reminders status list - Run the documents document id auto reminders status ETL stream [intent=etl availability=implemented stream=documents_document_id_auto_reminders_status]
    documents document id content list - Run the documents document id content ETL stream [intent=etl availability=implemented stream=documents_document_id_content]
    documents document id docx export tasks task id list - Run the documents document id docx export tasks task id ETL stream [intent=etl availability=implemented stream=documents_document_id_docx_export_tasks_task_id]
    documents document id esign disclosure list - Run the documents document id esign disclosure ETL stream [intent=etl availability=implemented stream=documents_document_id_esign_disclosure]
    documents document id sections list - Run the documents document id sections ETL stream [intent=etl availability=implemented stream=documents_document_id_sections]
    documents document id sections section id list - Run the documents document id sections section id ETL stream [intent=etl availability=implemented stream=documents_document_id_sections_section_id]
    documents document id sections uploads upload id list - Run the documents document id sections uploads upload id ETL stream [intent=etl availability=implemented stream=documents_document_id_sections_uploads_upload_id]
    documents document id settings list - Run the documents document id settings ETL stream [intent=etl availability=implemented stream=documents_document_id_settings]
    documents document id summary list - Run the documents document id summary ETL stream [intent=etl availability=implemented stream=documents_document_id_summary]
    documents folders list - Run the documents folders ETL stream [intent=etl availability=implemented stream=documents_folders]
    documents id attachments attachment id list - Run the documents id attachments attachment id ETL stream [intent=etl availability=implemented stream=documents_id_attachments_attachment_id]
    documents id attachments list - Run the documents id attachments ETL stream [intent=etl availability=implemented stream=documents_id_attachments]
    documents id details list - Run the documents id details ETL stream [intent=etl availability=implemented stream=documents_id_details]
    documents id fields list - Run the documents id fields ETL stream [intent=etl availability=implemented stream=documents_id_fields]
    documents id linked objects list - Run the documents id linked objects ETL stream [intent=etl availability=implemented stream=documents_id_linked_objects]
    documents id list - Run the documents id ETL stream [intent=etl availability=implemented stream=documents_id]
    documents linked objects list - Run the documents linked objects ETL stream [intent=etl availability=implemented stream=documents_linked_objects]
    documents list - Run the documents ETL stream [intent=etl availability=implemented stream=documents]
    documents search list - Run the documents search ETL stream [intent=etl availability=implemented stream=documents_search]
    forms list - Run the forms ETL stream [intent=etl availability=implemented stream=forms]
    logs detail list - Run the logs detail ETL stream [intent=etl availability=implemented stream=logs_detail]
    logs id detail list - Run the logs id detail ETL stream [intent=etl availability=implemented stream=logs_id_detail]
    logs id list - Run the logs id ETL stream [intent=etl availability=implemented stream=logs_id]
    logs list - Run the logs ETL stream [intent=etl availability=implemented stream=logs]
    members current list - Run the members current ETL stream [intent=etl availability=implemented stream=members_current]
    members id list - Run the members id ETL stream [intent=etl availability=implemented stream=members_id]
    members list - Run the members ETL stream [intent=etl availability=implemented stream=members]
    notary notaries list - Run the notary notaries ETL stream [intent=etl availability=implemented stream=notary_notaries]
    notary notarization requests list - Run the notary notarization requests ETL stream [intent=etl availability=implemented stream=notary_notarization_requests]
    notary notarization requests session request id list - Run the notary notarization requests session request id ETL stream [intent=etl availability=implemented stream=notary_notarization_requests_session_request_id]
    patch contacts id apply - Plan and execute the patch contacts id reverse-ETL action [intent=reverse_etl availability=implemented write=patch_contacts_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/contacts/{id} mutates PandaDoc data or workflow state; flags: --id (required)
    patch documents document id auto reminders apply - Plan and execute the patch documents document id auto reminders reverse-ETL action [intent=reverse_etl availability=implemented write=patch_documents_document_id_auto_reminders]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/documents/{document_id}/auto-reminders mutates PandaDoc data or workflow state; flags: --document_id (required)
    patch documents document id settings apply - Plan and execute the patch documents document id settings reverse-ETL action [intent=reverse_etl availability=implemented write=patch_documents_document_id_settings]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v2/documents/{document_id}/settings mutates PandaDoc data or workflow state; flags: --document_id (required)
    patch documents id apply - Plan and execute the patch documents id reverse-ETL action [intent=reverse_etl availability=implemented write=patch_documents_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/documents/{id} mutates PandaDoc data or workflow state; flags: --id (required)
    patch documents id fields apply - Plan and execute the patch documents id fields reverse-ETL action [intent=reverse_etl availability=implemented write=patch_documents_id_fields]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/documents/{id}/fields mutates PandaDoc data or workflow state; flags: --id (required)
    patch documents id ownership apply - Plan and execute the patch documents id ownership reverse-ETL action [intent=reverse_etl availability=implemented write=patch_documents_id_ownership]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/documents/{id}/ownership mutates PandaDoc data or workflow state; flags: --id (required)
    patch documents id recipients recipient recipient id apply - Plan and execute the patch documents id recipients recipient recipient id reverse-ETL action [intent=reverse_etl availability=implemented write=patch_documents_id_recipients_recipient_recipient_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/documents/{id}/recipients/recipient/{recipient_id} mutates PandaDoc data or workflow state; flags: --id (required), --recipient_id (required)
    patch documents id status apply - Plan and execute the patch documents id status reverse-ETL action [intent=reverse_etl availability=implemented write=patch_documents_id_status]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/documents/{id}/status mutates PandaDoc data or workflow state; flags: --id (required)
    patch documents ownership apply - Plan and execute the patch documents ownership reverse-ETL action [intent=reverse_etl availability=implemented write=patch_documents_ownership]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/documents/ownership mutates PandaDoc data or workflow state
    patch product catalog items item uuid apply - Plan and execute the patch product catalog items item uuid reverse-ETL action [intent=reverse_etl availability=implemented write=patch_product_catalog_items_item_uuid]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v2/product-catalog/items/{item_uuid} mutates PandaDoc data or workflow state; flags: --item_uuid (required)
    patch templates id apply - Plan and execute the patch templates id reverse-ETL action [intent=reverse_etl availability=implemented write=patch_templates_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/templates/{id} mutates PandaDoc data or workflow state; flags: --id (required)
    patch templates template id settings apply - Plan and execute the patch templates template id settings reverse-ETL action [intent=reverse_etl availability=implemented write=patch_templates_template_id_settings]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v2/templates/{template_id}/settings mutates PandaDoc data or workflow state; flags: --template_id (required)
    patch webhook subscriptions id apply - Plan and execute the patch webhook subscriptions id reverse-ETL action [intent=reverse_etl availability=implemented write=patch_webhook_subscriptions_id]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/webhook-subscriptions/{id} mutates PandaDoc data or workflow state; flags: --id (required)
    patch workspaces workspace id members member id role apply - Plan and execute the patch workspaces workspace id members member id role reverse-ETL action [intent=reverse_etl availability=implemented write=patch_workspaces_workspace_id_members_member_id_role]; approval: requires plan, preview, approval, and execute; risk: PATCH /public/v1/workspaces/{workspace_id}/members/{member_id}/role mutates PandaDoc data or workflow state; flags: --member_id (required), --workspace_id (required)
    post contacts apply - Plan and execute the post contacts reverse-ETL action [intent=reverse_etl availability=implemented write=post_contacts]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/contacts mutates PandaDoc data or workflow state
    post content library items apply - Plan and execute the post content library items reverse-ETL action [intent=reverse_etl availability=implemented write=post_content_library_items]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/content-library-items mutates PandaDoc data or workflow state
    post documents ai metadata apply - Plan and execute the post documents ai metadata reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_ai_metadata]; approval: requires plan, preview, approval, and execute; risk: POST /public/beta/documents/ai-metadata mutates PandaDoc data or workflow state
    post documents apply - Plan and execute the post documents reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents mutates PandaDoc data or workflow state
    post documents document id docx export tasks apply - Plan and execute the post documents document id docx export tasks reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_document_id_docx_export_tasks]; approval: requires plan, preview, approval, and execute; risk: POST /public/beta/documents/{document_id}/docx-export-tasks mutates PandaDoc data or workflow state; flags: --document_id (required)
    post documents document id sections apply - Plan and execute the post documents document id sections reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_document_id_sections]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{document_id}/sections mutates PandaDoc data or workflow state; flags: --document_id (required)
    post documents document id send reminder apply - Plan and execute the post documents document id send reminder reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_document_id_send_reminder]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{document_id}/send-reminder mutates PandaDoc data or workflow state; flags: --document_id (required)
    post documents folders apply - Plan and execute the post documents folders reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_folders]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/folders mutates PandaDoc data or workflow state
    post documents id append content library item apply - Plan and execute the post documents id append content library item reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_append_content_library_item]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/append-content-library-item mutates PandaDoc data or workflow state; flags: --id (required)
    post documents id attachments apply - Plan and execute the post documents id attachments reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_attachments]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/attachments mutates PandaDoc data or workflow state; flags: --id (required)
    post documents id draft apply - Plan and execute the post documents id draft reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_draft]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/draft mutates PandaDoc data or workflow state; flags: --id (required)
    post documents id editing sessions apply - Plan and execute the post documents id editing sessions reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_editing_sessions]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/editing-sessions mutates PandaDoc data or workflow state; flags: --id (required)
    post documents id fields apply - Plan and execute the post documents id fields reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_fields]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/fields mutates PandaDoc data or workflow state; flags: --id (required)
    post documents id linked objects apply - Plan and execute the post documents id linked objects reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_linked_objects]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/linked-objects mutates PandaDoc data or workflow state; flags: --id (required)
    post documents id move to folder folder id apply - Plan and execute the post documents id move to folder folder id reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_move_to_folder_folder_id]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/move-to-folder/{folder_id} mutates PandaDoc data or workflow state; flags: --folder_id (required), --id (required)
    post documents id recipients apply - Plan and execute the post documents id recipients reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_recipients]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/recipients mutates PandaDoc data or workflow state; flags: --id (required)
    post documents id recipients recipient id reassign apply - Plan and execute the post documents id recipients recipient id reassign reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_recipients_recipient_id_reassign]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/recipients/{recipient_id}/reassign mutates PandaDoc data or workflow state; flags: --id (required), --recipient_id (required)
    post documents id send apply - Plan and execute the post documents id send reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_send]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/send mutates PandaDoc data or workflow state; flags: --id (required)
    post documents id session apply - Plan and execute the post documents id session reverse-ETL action [intent=reverse_etl availability=implemented write=post_documents_id_session]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/documents/{id}/session mutates PandaDoc data or workflow state; flags: --id (required)
    post notary notarization requests apply - Plan and execute the post notary notarization requests reverse-ETL action [intent=reverse_etl availability=implemented write=post_notary_notarization_requests]; approval: requires plan, preview, approval, and execute; risk: POST /public/v2/notary/notarization-requests mutates PandaDoc data or workflow state
    post product catalog items apply - Plan and execute the post product catalog items reverse-ETL action [intent=reverse_etl availability=implemented write=post_product_catalog_items]; approval: requires plan, preview, approval, and execute; risk: POST /public/v2/product-catalog/items mutates PandaDoc data or workflow state
    post templates apply - Plan and execute the post templates reverse-ETL action [intent=reverse_etl availability=implemented write=post_templates]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/templates mutates PandaDoc data or workflow state
    post templates folders apply - Plan and execute the post templates folders reverse-ETL action [intent=reverse_etl availability=implemented write=post_templates_folders]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/templates/folders mutates PandaDoc data or workflow state
    post templates id editing sessions apply - Plan and execute the post templates id editing sessions reverse-ETL action [intent=reverse_etl availability=implemented write=post_templates_id_editing_sessions]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/templates/{id}/editing-sessions mutates PandaDoc data or workflow state; flags: --id (required)
    post users apply - Plan and execute the post users reverse-ETL action [intent=reverse_etl availability=implemented write=post_users]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/users mutates PandaDoc data or workflow state
    post webhook subscriptions apply - Plan and execute the post webhook subscriptions reverse-ETL action [intent=reverse_etl availability=implemented write=post_webhook_subscriptions]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/webhook-subscriptions mutates PandaDoc data or workflow state
    post workspaces apply - Plan and execute the post workspaces reverse-ETL action [intent=reverse_etl availability=implemented write=post_workspaces]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/workspaces mutates PandaDoc data or workflow state
    post workspaces workspace id deactivate apply - Plan and execute the post workspaces workspace id deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=post_workspaces_workspace_id_deactivate]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/workspaces/{workspace_id}/deactivate mutates PandaDoc data or workflow state; flags: --workspace_id (required)
    post workspaces workspace id members apply - Plan and execute the post workspaces workspace id members reverse-ETL action [intent=reverse_etl availability=implemented write=post_workspaces_workspace_id_members]; approval: requires plan, preview, approval, and execute; risk: POST /public/v1/workspaces/{workspace_id}/members mutates PandaDoc data or workflow state; flags: --workspace_id (required)
    product catalog items item uuid list - Run the product catalog items item uuid ETL stream [intent=etl availability=implemented stream=product_catalog_items_item_uuid]
    product catalog items search list - Run the product catalog items search ETL stream [intent=etl availability=implemented stream=product_catalog_items_search]
    put documents document id quotes quote id apply - Plan and execute the put documents document id quotes quote id reverse-ETL action [intent=reverse_etl availability=implemented write=put_documents_document_id_quotes_quote_id]; approval: requires plan, preview, approval, and execute; risk: PUT /public/v1/documents/{document_id}/quotes/{quote_id} mutates PandaDoc data or workflow state; flags: --document_id (required), --quote_id (required)
    put documents folders id apply - Plan and execute the put documents folders id reverse-ETL action [intent=reverse_etl availability=implemented write=put_documents_folders_id]; approval: requires plan, preview, approval, and execute; risk: PUT /public/v1/documents/folders/{id} mutates PandaDoc data or workflow state; flags: --id (required)
    put templates folders id apply - Plan and execute the put templates folders id reverse-ETL action [intent=reverse_etl availability=implemented write=put_templates_folders_id]; approval: requires plan, preview, approval, and execute; risk: PUT /public/v1/templates/folders/{id} mutates PandaDoc data or workflow state; flags: --id (required)
    sms opt outs list - Run the sms opt outs ETL stream [intent=etl availability=implemented stream=sms_opt_outs]
    templates folders list - Run the templates folders ETL stream [intent=etl availability=implemented stream=templates_folders]
    templates id details list - Run the templates id details ETL stream [intent=etl availability=implemented stream=templates_id_details]
    templates id list - Run the templates id ETL stream [intent=etl availability=implemented stream=templates_id]
    templates list - Run the templates ETL stream [intent=etl availability=implemented stream=templates]
    templates template id settings list - Run the templates template id settings ETL stream [intent=etl availability=implemented stream=templates_template_id_settings]
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
    users user id list - Run the users user id ETL stream [intent=etl availability=implemented stream=users_user_id]
    webhook events id list - Run the webhook events id ETL stream [intent=etl availability=implemented stream=webhook_events_id]
    webhook events list - Run the webhook events ETL stream [intent=etl availability=implemented stream=webhook_events]
    webhook subscriptions id list - Run the webhook subscriptions id ETL stream [intent=etl availability=implemented stream=webhook_subscriptions_id]
    webhook subscriptions list - Run the webhook subscriptions ETL stream [intent=etl availability=implemented stream=webhook_subscriptions]
    workspaces list - Run the workspaces ETL stream [intent=etl availability=implemented stream=workspaces]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pandadoc

  # Inspect as structured JSON
  pm connectors inspect pandadoc --json

AGENT WORKFLOW
  - Run pm connectors inspect pandadoc before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
