---
name: pm-pandadoc
description: PandaDoc connector knowledge and safe action guide.
---

# pm-pandadoc

## Purpose

Reads and writes documented PandaDoc public API resources across documents, templates, contacts, folders, forms, logs, members, webhooks, workspaces, notary, and catalog surfaces.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- attachment_id
- base_url
- count
- document_id
- id
- item_uuid
- mode
- section_id
- session_request_id
- task_id
- template_id
- upload_id
- user_id
- api_key (secret)

## ETL Streams

- documents:
  - primary key: id
  - cursor: date_created
  - fields: date_created(), date_modified(), id(), name(), status()
- templates:
  - primary key: id
  - cursor: date_created
  - fields: date_created(), date_modified(), id(), name()
- contacts:
  - primary key: id
  - cursor: created_date
  - fields: created_date(), email(), first_name(), id(), last_name()
- documents_document_id_ai_metadata:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_content:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_docx_export_tasks_task_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_summary:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_search:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- contacts_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- content_library_items:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- content_library_items_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- content_library_items_id_details:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_auto_reminders:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_auto_reminders_status:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_esign_disclosure:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_sections:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_sections_section_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_sections_uploads_upload_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_id_attachments:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_id_attachments_attachment_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_id_details:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_id_fields:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_id_linked_objects:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_folders:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_linked_objects:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- forms:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- logs:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- logs_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- members:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- members_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- members_current:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- sms_opt_outs:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- templates_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- templates_id_details:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- templates_folders:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- users:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- users_user_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- webhook_events:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- webhook_events_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- webhook_subscriptions:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- webhook_subscriptions_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- workspaces:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_audit_trail:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- documents_document_id_settings:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- logs_detail:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- logs_id_detail:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- notary_notaries:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- notary_notarization_requests:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- notary_notarization_requests_session_request_id:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- product_catalog_items_item_uuid:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- product_catalog_items_search:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()
- templates_template_id_settings:
  - fields: date_created(), date_modified(), id(), name(), status(), uuid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- post_documents_document_id_docx_export_tasks:
  - endpoint: POST /public/beta/documents/{{ record.document_id }}/docx-export-tasks
  - required fields: document_id
  - risk: POST /public/beta/documents/{document_id}/docx-export-tasks mutates PandaDoc data or workflow state
- post_documents_ai_metadata:
  - endpoint: POST /public/beta/documents/ai-metadata
  - risk: POST /public/beta/documents/ai-metadata mutates PandaDoc data or workflow state
- post_contacts:
  - endpoint: POST /public/v1/contacts
  - risk: POST /public/v1/contacts mutates PandaDoc data or workflow state
- delete_contacts_id:
  - endpoint: DELETE /public/v1/contacts/{{ record.id }}
  - required fields: id
  - risk: deletes PandaDoc resource via /public/v1/contacts/{id}; destructive external mutation
- patch_contacts_id:
  - endpoint: PATCH /public/v1/contacts/{{ record.id }}
  - required fields: id
  - risk: PATCH /public/v1/contacts/{id} mutates PandaDoc data or workflow state
- post_content_library_items:
  - endpoint: POST /public/v1/content-library-items
  - risk: POST /public/v1/content-library-items mutates PandaDoc data or workflow state
- delete_documents:
  - endpoint: DELETE /public/v1/documents
  - risk: deletes PandaDoc resource via /public/v1/documents; destructive external mutation
- post_documents:
  - endpoint: POST /public/v1/documents
  - risk: POST /public/v1/documents mutates PandaDoc data or workflow state
- patch_documents_document_id_auto_reminders:
  - endpoint: PATCH /public/v1/documents/{{ record.document_id }}/auto-reminders
  - required fields: document_id
  - risk: PATCH /public/v1/documents/{document_id}/auto-reminders mutates PandaDoc data or workflow state
- put_documents_document_id_quotes_quote_id:
  - endpoint: PUT /public/v1/documents/{{ record.document_id }}/quotes/{{ record.quote_id }}
  - required fields: document_id, quote_id
  - risk: PUT /public/v1/documents/{document_id}/quotes/{quote_id} mutates PandaDoc data or workflow state
- post_documents_document_id_sections:
  - endpoint: POST /public/v1/documents/{{ record.document_id }}/sections
  - required fields: document_id
  - risk: POST /public/v1/documents/{document_id}/sections mutates PandaDoc data or workflow state
- delete_documents_document_id_sections_section_id:
  - endpoint: DELETE /public/v1/documents/{{ record.document_id }}/sections/{{ record.section_id }}
  - required fields: document_id, section_id
  - risk: deletes PandaDoc resource via /public/v1/documents/{document_id}/sections/{section_id}; destructive external mutation
- post_documents_document_id_send_reminder:
  - endpoint: POST /public/v1/documents/{{ record.document_id }}/send-reminder
  - required fields: document_id
  - risk: POST /public/v1/documents/{document_id}/send-reminder mutates PandaDoc data or workflow state
- delete_documents_id:
  - endpoint: DELETE /public/v1/documents/{{ record.id }}
  - required fields: id
  - risk: deletes PandaDoc resource via /public/v1/documents/{id}; destructive external mutation
- patch_documents_id:
  - endpoint: PATCH /public/v1/documents/{{ record.id }}
  - required fields: id
  - risk: PATCH /public/v1/documents/{id} mutates PandaDoc data or workflow state
- post_documents_id_append_content_library_item:
  - endpoint: POST /public/v1/documents/{{ record.id }}/append-content-library-item
  - required fields: id
  - risk: POST /public/v1/documents/{id}/append-content-library-item mutates PandaDoc data or workflow state
- post_documents_id_attachments:
  - endpoint: POST /public/v1/documents/{{ record.id }}/attachments
  - required fields: id
  - risk: POST /public/v1/documents/{id}/attachments mutates PandaDoc data or workflow state
- delete_documents_id_attachments_attachment_id:
  - endpoint: DELETE /public/v1/documents/{{ record.id }}/attachments/{{ record.attachment_id }}
  - required fields: id, attachment_id
  - risk: deletes PandaDoc resource via /public/v1/documents/{id}/attachments/{attachment_id}; destructive external mutation
- post_documents_id_draft:
  - endpoint: POST /public/v1/documents/{{ record.id }}/draft
  - required fields: id
  - risk: POST /public/v1/documents/{id}/draft mutates PandaDoc data or workflow state
- post_documents_id_editing_sessions:
  - endpoint: POST /public/v1/documents/{{ record.id }}/editing-sessions
  - required fields: id
  - risk: POST /public/v1/documents/{id}/editing-sessions mutates PandaDoc data or workflow state
- patch_documents_id_fields:
  - endpoint: PATCH /public/v1/documents/{{ record.id }}/fields
  - required fields: id
  - risk: PATCH /public/v1/documents/{id}/fields mutates PandaDoc data or workflow state
- post_documents_id_fields:
  - endpoint: POST /public/v1/documents/{{ record.id }}/fields
  - required fields: id
  - risk: POST /public/v1/documents/{id}/fields mutates PandaDoc data or workflow state
- post_documents_id_linked_objects:
  - endpoint: POST /public/v1/documents/{{ record.id }}/linked-objects
  - required fields: id
  - risk: POST /public/v1/documents/{id}/linked-objects mutates PandaDoc data or workflow state
- delete_documents_id_linked_objects_linked_object_id:
  - endpoint: DELETE /public/v1/documents/{{ record.id }}/linked-objects/{{ record.linked_object_id }}
  - required fields: id, linked_object_id
  - risk: deletes PandaDoc resource via /public/v1/documents/{id}/linked-objects/{linked_object_id}; destructive external mutation
- post_documents_id_move_to_folder_folder_id:
  - endpoint: POST /public/v1/documents/{{ record.id }}/move-to-folder/{{ record.folder_id }}
  - required fields: id, folder_id
  - risk: POST /public/v1/documents/{id}/move-to-folder/{folder_id} mutates PandaDoc data or workflow state
- patch_documents_id_ownership:
  - endpoint: PATCH /public/v1/documents/{{ record.id }}/ownership
  - required fields: id
  - risk: PATCH /public/v1/documents/{id}/ownership mutates PandaDoc data or workflow state
- post_documents_id_recipients:
  - endpoint: POST /public/v1/documents/{{ record.id }}/recipients
  - required fields: id
  - risk: POST /public/v1/documents/{id}/recipients mutates PandaDoc data or workflow state
- delete_documents_id_recipients_recipient_id:
  - endpoint: DELETE /public/v1/documents/{{ record.id }}/recipients/{{ record.recipient_id }}
  - required fields: id, recipient_id
  - risk: deletes PandaDoc resource via /public/v1/documents/{id}/recipients/{recipient_id}; destructive external mutation
- post_documents_id_recipients_recipient_id_reassign:
  - endpoint: POST /public/v1/documents/{{ record.id }}/recipients/{{ record.recipient_id }}/reassign
  - required fields: id, recipient_id
  - risk: POST /public/v1/documents/{id}/recipients/{recipient_id}/reassign mutates PandaDoc data or workflow state
- patch_documents_id_recipients_recipient_recipient_id:
  - endpoint: PATCH /public/v1/documents/{{ record.id }}/recipients/recipient/{{ record.recipient_id }}
  - required fields: id, recipient_id
  - risk: PATCH /public/v1/documents/{id}/recipients/recipient/{recipient_id} mutates PandaDoc data or workflow state
- post_documents_id_send:
  - endpoint: POST /public/v1/documents/{{ record.id }}/send
  - required fields: id
  - risk: POST /public/v1/documents/{id}/send mutates PandaDoc data or workflow state
- post_documents_id_session:
  - endpoint: POST /public/v1/documents/{{ record.id }}/session
  - required fields: id
  - risk: POST /public/v1/documents/{id}/session mutates PandaDoc data or workflow state
- patch_documents_id_status:
  - endpoint: PATCH /public/v1/documents/{{ record.id }}/status
  - required fields: id
  - risk: PATCH /public/v1/documents/{id}/status mutates PandaDoc data or workflow state
- post_documents_folders:
  - endpoint: POST /public/v1/documents/folders
  - risk: POST /public/v1/documents/folders mutates PandaDoc data or workflow state
- put_documents_folders_id:
  - endpoint: PUT /public/v1/documents/folders/{{ record.id }}
  - required fields: id
  - risk: PUT /public/v1/documents/folders/{id} mutates PandaDoc data or workflow state
- patch_documents_ownership:
  - endpoint: PATCH /public/v1/documents/ownership
  - risk: PATCH /public/v1/documents/ownership mutates PandaDoc data or workflow state
- post_templates:
  - endpoint: POST /public/v1/templates
  - risk: POST /public/v1/templates mutates PandaDoc data or workflow state
- delete_templates_id:
  - endpoint: DELETE /public/v1/templates/{{ record.id }}
  - required fields: id
  - risk: deletes PandaDoc resource via /public/v1/templates/{id}; destructive external mutation
- patch_templates_id:
  - endpoint: PATCH /public/v1/templates/{{ record.id }}
  - required fields: id
  - risk: PATCH /public/v1/templates/{id} mutates PandaDoc data or workflow state
- post_templates_id_editing_sessions:
  - endpoint: POST /public/v1/templates/{{ record.id }}/editing-sessions
  - required fields: id
  - risk: POST /public/v1/templates/{id}/editing-sessions mutates PandaDoc data or workflow state
- post_templates_folders:
  - endpoint: POST /public/v1/templates/folders
  - risk: POST /public/v1/templates/folders mutates PandaDoc data or workflow state
- put_templates_folders_id:
  - endpoint: PUT /public/v1/templates/folders/{{ record.id }}
  - required fields: id
  - risk: PUT /public/v1/templates/folders/{id} mutates PandaDoc data or workflow state
- post_users:
  - endpoint: POST /public/v1/users
  - risk: POST /public/v1/users mutates PandaDoc data or workflow state
- post_webhook_subscriptions:
  - endpoint: POST /public/v1/webhook-subscriptions
  - risk: POST /public/v1/webhook-subscriptions mutates PandaDoc data or workflow state
- delete_webhook_subscriptions_id:
  - endpoint: DELETE /public/v1/webhook-subscriptions/{{ record.id }}
  - required fields: id
  - risk: deletes PandaDoc resource via /public/v1/webhook-subscriptions/{id}; destructive external mutation
- patch_webhook_subscriptions_id:
  - endpoint: PATCH /public/v1/webhook-subscriptions/{{ record.id }}
  - required fields: id
  - risk: PATCH /public/v1/webhook-subscriptions/{id} mutates PandaDoc data or workflow state
- post_workspaces:
  - endpoint: POST /public/v1/workspaces
  - risk: POST /public/v1/workspaces mutates PandaDoc data or workflow state
- post_workspaces_workspace_id_deactivate:
  - endpoint: POST /public/v1/workspaces/{{ record.workspace_id }}/deactivate
  - required fields: workspace_id
  - risk: POST /public/v1/workspaces/{workspace_id}/deactivate mutates PandaDoc data or workflow state
- post_workspaces_workspace_id_members:
  - endpoint: POST /public/v1/workspaces/{{ record.workspace_id }}/members
  - required fields: workspace_id
  - risk: POST /public/v1/workspaces/{workspace_id}/members mutates PandaDoc data or workflow state
- delete_workspaces_workspace_id_members_member_id:
  - endpoint: DELETE /public/v1/workspaces/{{ record.workspace_id }}/members/{{ record.member_id }}
  - required fields: workspace_id, member_id
  - risk: deletes PandaDoc resource via /public/v1/workspaces/{workspace_id}/members/{member_id}; destructive external mutation
- patch_workspaces_workspace_id_members_member_id_role:
  - endpoint: PATCH /public/v1/workspaces/{{ record.workspace_id }}/members/{{ record.member_id }}/role
  - required fields: workspace_id, member_id
  - risk: PATCH /public/v1/workspaces/{workspace_id}/members/{member_id}/role mutates PandaDoc data or workflow state
- patch_documents_document_id_settings:
  - endpoint: PATCH /public/v2/documents/{{ record.document_id }}/settings
  - required fields: document_id
  - risk: PATCH /public/v2/documents/{document_id}/settings mutates PandaDoc data or workflow state
- post_notary_notarization_requests:
  - endpoint: POST /public/v2/notary/notarization-requests
  - risk: POST /public/v2/notary/notarization-requests mutates PandaDoc data or workflow state
- delete_notary_notarization_requests_session_request_id:
  - endpoint: DELETE /public/v2/notary/notarization-requests/{{ record.session_request_id }}
  - required fields: session_request_id
  - risk: deletes PandaDoc resource via /public/v2/notary/notarization-requests/{session_request_id}; destructive external mutation
- post_product_catalog_items:
  - endpoint: POST /public/v2/product-catalog/items
  - risk: POST /public/v2/product-catalog/items mutates PandaDoc data or workflow state
- delete_product_catalog_items_item_uuid:
  - endpoint: DELETE /public/v2/product-catalog/items/{{ record.item_uuid }}
  - required fields: item_uuid
  - risk: deletes PandaDoc resource via /public/v2/product-catalog/items/{item_uuid}; destructive external mutation
- patch_product_catalog_items_item_uuid:
  - endpoint: PATCH /public/v2/product-catalog/items/{{ record.item_uuid }}
  - required fields: item_uuid
  - risk: PATCH /public/v2/product-catalog/items/{item_uuid} mutates PandaDoc data or workflow state
- patch_templates_template_id_settings:
  - endpoint: PATCH /public/v2/templates/{{ record.template_id }}/settings
  - required fields: template_id
  - risk: PATCH /public/v2/templates/{template_id}/settings mutates PandaDoc data or workflow state

## Security

- read risk: external PandaDoc API read of document, template, contact, member, workspace, webhook, notary, and catalog data
- write risk: external PandaDoc API mutations including document sending/status changes, contact/template/folder/webhook/admin updates, and destructive deletes
- approval: required for write actions; destructive and notification-sending actions carry explicit risk metadata
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pandadoc
```

### Inspect as structured JSON

```bash
pm connectors inspect pandadoc --json
```

## Agent Rules

- Run pm connectors inspect pandadoc before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
