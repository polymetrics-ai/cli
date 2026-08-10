# pm connectors inspect lever-hiring

```text
NAME
  pm connectors inspect lever-hiring - Lever Hiring connector manual

SYNOPSIS
  pm connectors inspect lever-hiring
  pm connectors inspect lever-hiring --json
  pm credentials add <name> --connector lever-hiring [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Lever Hiring opportunities, postings, users, requisitions, stages, and related hiring resources; exposes bounded direct reads and selected typed reverse-ETL write plans.

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
  base_url
  mode
  access_token (secret)
  api_key (secret)

ETL STREAMS
  opportunities:
    primary key: id
    cursor: createdAt
    fields: archivedAt(integer), createdAt(integer), emails(array), headline(string), id(string), lastInteractionAt(integer), name(string), origin(string), sources(array), stage(string), tags(array), updatedAt(integer)
  postings:
    primary key: id
    cursor: createdAt
    fields: categories(object), createdAt(integer), hiringManager(string), id(string), owner(string), state(string), text(string), updatedAt(integer), user(string)
  users:
    primary key: id
    cursor: createdAt
    fields: accessRole(string), createdAt(integer), deactivatedAt(integer), email(string), id(string), name(string), username(string)
  requisitions:
    primary key: id
    cursor: createdAt
    fields: createdAt(integer), headcountHired(integer), headcountTotal(integer), id(string), name(string), owner(string), requisitionCode(string), status(string), updatedAt(integer)
  stages:
    primary key: id
    fields: id(string), text(string)
  deleted_applications:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  archive_reasons:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  audit_events:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  disposition_stages:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  feedback_templates:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  deleted_opportunities:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  deleted_postings:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  form_templates:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  requisition_fields:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  sources:
    primary key: id
    fields: createdAt(integer), id(string), name(string), text(string), updatedAt(integer)
  applications:
    primary key: id
    fields: createdAt(integer), id(string), name(string), opportunity_id(string), text(string), updatedAt(integer)
  opportunity_feedback:
    primary key: id
    fields: createdAt(integer), id(string), name(string), opportunity_id(string), text(string), updatedAt(integer)
  opportunity_interviews:
    primary key: id
    fields: createdAt(integer), id(string), name(string), opportunity_id(string), text(string), updatedAt(integer)
  opportunity_notes:
    primary key: id
    fields: createdAt(integer), id(string), name(string), opportunity_id(string), text(string), updatedAt(integer)
  opportunity_offers:
    primary key: id
    fields: createdAt(integer), id(string), name(string), opportunity_id(string), text(string), updatedAt(integer)
  opportunity_file_actions:
    primary key: id
    fields: createdAt(integer), id(string), name(string), opportunity_id(string), text(string), updatedAt(integer)
  opportunity_panels:
    primary key: id
    fields: createdAt(integer), id(string), name(string), opportunity_id(string), text(string), updatedAt(integer)
  opportunity_forms:
    primary key: id
    fields: createdAt(integer), id(string), name(string), opportunity_id(string), text(string), updatedAt(integer)
  opportunity_referrals:
    primary key: id
    fields: createdAt(integer), id(string), name(string), opportunity_id(string), text(string), updatedAt(integer)
  posting_users:
    primary key: id
    fields: createdAt(integer), id(string), name(string), posting_id(string), text(string), updatedAt(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  update_feedback:
    endpoint: PUT /opportunities/{{ record.opportunity }}/feedback/{{ record.feedback }}
    required fields: opportunity, feedback
    risk: Updates a feedback form for an opportunity. Reverse ETL writes require plan, preview, explicit approval, and execute.
  delete_feedback:
    endpoint: DELETE /opportunities/{{ record.opportunity }}/feedback/{{ record.feedback }}
    required fields: opportunity, feedback
    risk: Deletes a feedback form from an opportunity. Reverse ETL writes require plan, preview, explicit approval, and execute.
  create_feedback_template:
    endpoint: POST /feedback_templates
    required fields: text, fields
    risk: Creates a feedback template. Reverse ETL writes require plan, preview, explicit approval, and execute.
  update_feedback_template:
    endpoint: PUT /feedback_templates/{{ record.feedback_template }}
    required fields: feedback_template
    risk: Updates a feedback template. Reverse ETL writes require plan, preview, explicit approval, and execute.
  delete_feedback_template:
    endpoint: DELETE /feedback_templates/{{ record.feedback_template }}
    required fields: feedback_template
    risk: Deletes a feedback template. Reverse ETL writes require plan, preview, explicit approval, and execute.
  create_form_template:
    endpoint: POST /form_templates
    required fields: text, fields
    risk: Creates a profile form template. Reverse ETL writes require plan, preview, explicit approval, and execute.
  update_form_template:
    endpoint: PUT /form_templates/{{ record.form_template }}
    required fields: form_template
    risk: Updates a profile form template. Reverse ETL writes require plan, preview, explicit approval, and execute.
  delete_form_template:
    endpoint: DELETE /form_templates/{{ record.form_template }}
    required fields: form_template
    risk: Deletes a profile form template. Reverse ETL writes require plan, preview, explicit approval, and execute.
  delete_note:
    endpoint: DELETE /opportunities/{{ record.opportunity }}/notes/{{ record.noteId }}
    required fields: opportunity, noteId
    risk: Deletes a note from an opportunity. Reverse ETL writes require plan, preview, explicit approval, and execute.
  delete_requisition:
    endpoint: DELETE /requisitions/{{ record.requisition }}
    required fields: requisition
    risk: Deletes a requisition. Reverse ETL writes require plan, preview, explicit approval, and execute.
  delete_requisition_field_options:
    endpoint: DELETE /requisition_fields/{{ record.requisition_field }}/options
    required fields: requisition_field
    risk: Deletes dropdown options for a requisition field. Reverse ETL writes require plan, preview, explicit approval, and execute.
  delete_requisition_field:
    endpoint: DELETE /requisition_fields/{{ record.requisition_field }}
    required fields: requisition_field
    risk: Deletes a requisition field. Reverse ETL writes require plan, preview, explicit approval, and execute.
  deactivate_user:
    endpoint: POST /users/{{ record.user }}/deactivate
    required fields: user
    risk: Deactivates a Lever user. Reverse ETL writes require plan, preview, explicit approval, and execute.
  reactivate_user:
    endpoint: POST /users/{{ record.user }}/reactivate
    required fields: user
    risk: Reactivates a Lever user. Reverse ETL writes require plan, preview, explicit approval, and execute.

SECURITY
  read risk: external Lever API read of candidate and hiring pipeline data
  write risk: selected Lever mutations are exposed only through reverse ETL plan, preview, explicit approval, and execute; destructive actions require typed confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Lever Hiring's declared streams and reverse-ETL actions.
  Usage: pm lever-hiring <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete opportunities opportunity files file - Documented DELETE /opportunities/{opportunity}/files/{file} (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.delete.opportunities-opportunity-files-file]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api delete opportunities opportunity interviews interview - Documented DELETE /opportunities/{opportunity}/interviews/{interview} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.delete.opportunities-opportunity-interviews-interview]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: critical; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete opportunities opportunity panels panel - Documented DELETE /opportunities/{opportunity}/panels/{panel} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.delete.opportunities-opportunity-panels-panel]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: critical; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete webhooks webhookid - Documented DELETE /webhooks/{webhookId} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.delete.webhooks-webhookid]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api get opportunities opportunity files - Documented GET /opportunities/{opportunity}/files (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.get.opportunities-opportunity-files]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api get opportunities opportunity files file - Documented GET /opportunities/{opportunity}/files/{file} (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.get.opportunities-opportunity-files-file]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api get opportunities opportunity files file download - Documented GET /opportunities/{opportunity}/files/{file}/download (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.get.opportunities-opportunity-files-file-download]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api get opportunities opportunity offers offer download - Documented GET /opportunities/{opportunity}/offers/{offer}/download (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.get.opportunities-opportunity-offers-offer-download]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api get opportunities opportunity resumes - Documented GET /opportunities/{opportunity}/resumes (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.get.opportunities-opportunity-resumes]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api get opportunities opportunity resumes resume - Documented GET /opportunities/{opportunity}/resumes/{resume} (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.get.opportunities-opportunity-resumes-resume]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api get opportunities opportunity resumes resume download - Documented GET /opportunities/{opportunity}/resumes/{resume}/download (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.get.opportunities-opportunity-resumes-resume-download]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api get v1 eeo responses - Documented GET /v1/eeo/responses (not implemented) [intent=direct_read availability=not_implemented operation=lever-hiring.get.v1-eeo-responses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: high; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get webhooks - Documented GET /webhooks (not implemented) [intent=direct_read availability=not_implemented operation=lever-hiring.get.webhooks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: high; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post opportunities - Documented POST /opportunities (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity addlinks - Documented POST /opportunities/{opportunity}/addLinks (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-addlinks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity addsources - Documented POST /opportunities/{opportunity}/addSources (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-addsources]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity addtags - Documented POST /opportunities/{opportunity}/addTags (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-addtags]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity feedback - Documented POST /opportunities/{opportunity}/feedback (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-feedback]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity files - Documented POST /opportunities/{opportunity}/files (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-files]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api post opportunities opportunity forms - Documented POST /opportunities/{opportunity}/forms (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-forms]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity interviews - Documented POST /opportunities/{opportunity}/interviews (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-interviews]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity notes - Documented POST /opportunities/{opportunity}/notes (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-notes]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity panels - Documented POST /opportunities/{opportunity}/panels (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-panels]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity removelinks - Documented POST /opportunities/{opportunity}/removeLinks (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-removelinks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity removesources - Documented POST /opportunities/{opportunity}/removeSources (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-removesources]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post opportunities opportunity removetags - Documented POST /opportunities/{opportunity}/removeTags (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.opportunities-opportunity-removetags]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post postings - Documented POST /postings (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.postings]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post postings posting - Documented POST /postings/{posting} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.postings-posting]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post postings posting apply - Documented POST /postings/{posting}/apply (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.postings-posting-apply]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post requisition-fields - Documented POST /requisition_fields (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.requisition-fields]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post requisition-fields requisition-field options - Documented POST /requisition_fields/{requisition_field}/options (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.requisition-fields-requisition-field-options]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post requisitions - Documented POST /requisitions (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.requisitions]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post uploads - Documented POST /uploads (not implemented) [intent=binary_download availability=not_implemented operation=lever-hiring.post.uploads]; approval: not implemented: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; risk: high; notes: named_dependency=engine.binary_download_operation_contract: the binary-download executor lacks a reviewed operation-specific destination and CLI contract; flags: --dest-root (required), --file-name, --max-bytes
    api post users - Documented POST /users (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.users]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post webhooks - Documented POST /webhooks (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.post.webhooks]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api put contacts contact - Documented PUT /contacts/{contact} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.contacts-contact]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put opportunities opportunity archived - Documented PUT /opportunities/{opportunity}/archived (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.opportunities-opportunity-archived]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put opportunities opportunity interviews interview - Documented PUT /opportunities/{opportunity}/interviews/{interview} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.opportunities-opportunity-interviews-interview]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put opportunities opportunity notes note - Documented PUT /opportunities/{opportunity}/notes/{note} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.opportunities-opportunity-notes-note]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put opportunities opportunity panels panel - Documented PUT /opportunities/{opportunity}/panels/{panel} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.opportunities-opportunity-panels-panel]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put opportunities opportunity stage - Documented PUT /opportunities/{opportunity}/stage (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.opportunities-opportunity-stage]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put requisition-fields requisition-field - Documented PUT /requisition_fields/{requisition_field} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.requisition-fields-requisition-field]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put requisition-fields requisition-field options - Documented PUT /requisition_fields/{requisition_field}/options (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.requisition-fields-requisition-field-options]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put requisitions requisition - Documented PUT /requisitions/{requisition} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.requisitions-requisition]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put users user - Documented PUT /users/{user} (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.users-user]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put webhooks - Documented PUT /webhooks/ (not implemented) [intent=direct_write availability=not_implemented operation=lever-hiring.put.webhooks]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    applications list - Read Lever Hiring applications as ETL records. [intent=etl availability=implemented stream=applications]
    archive-reasons list - Read Lever Hiring archive reasons as ETL records. [intent=etl availability=implemented stream=archive_reasons]
    audit-events list - Read Lever Hiring audit events as ETL records. [intent=etl availability=implemented stream=audit_events]
    create-feedback-template plan - Creates a feedback template. [intent=reverse_etl availability=not_implemented write=create_feedback_template]; approval: requires plan, preview, approval, and execute; risk: Creates a feedback template. Reverse ETL writes require plan, preview, explicit approval, and execute.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-form-template plan - Creates a profile form template. [intent=reverse_etl availability=not_implemented write=create_form_template]; approval: requires plan, preview, approval, and execute; risk: Creates a profile form template. Reverse ETL writes require plan, preview, explicit approval, and execute.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    deactivate-user plan - Deactivates a Lever user. [intent=reverse_etl availability=implemented write=deactivate_user]; approval: requires plan, preview, approval, and execute; risk: Deactivates a Lever user. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --user (required)
    delete-feedback plan - Deletes a feedback form from an opportunity. [intent=reverse_etl availability=implemented write=delete_feedback]; approval: requires plan, preview, approval, and execute; risk: Deletes a feedback form from an opportunity. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --feedback (required), --opportunity (required)
    delete-feedback-template plan - Deletes a feedback template. [intent=reverse_etl availability=implemented write=delete_feedback_template]; approval: requires plan, preview, approval, and execute; risk: Deletes a feedback template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --feedback_template (required)
    delete-form-template plan - Deletes a profile form template. [intent=reverse_etl availability=implemented write=delete_form_template]; approval: requires plan, preview, approval, and execute; risk: Deletes a profile form template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --form_template (required)
    delete-note plan - Deletes a note from an opportunity. [intent=reverse_etl availability=implemented write=delete_note]; approval: requires plan, preview, approval, and execute; risk: Deletes a note from an opportunity. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --noteId (required), --opportunity (required)
    delete-requisition plan - Deletes a requisition. [intent=reverse_etl availability=implemented write=delete_requisition]; approval: requires plan, preview, approval, and execute; risk: Deletes a requisition. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --requisition (required)
    delete-requisition-field plan - Deletes a requisition field. [intent=reverse_etl availability=implemented write=delete_requisition_field]; approval: requires plan, preview, approval, and execute; risk: Deletes a requisition field. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --requisition_field (required)
    delete-requisition-field-options plan - Deletes dropdown options for a requisition field. [intent=reverse_etl availability=implemented write=delete_requisition_field_options]; approval: requires plan, preview, approval, and execute; risk: Deletes dropdown options for a requisition field. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --requisition_field (required)
    deleted-applications list - Read Lever Hiring deleted applications as ETL records. [intent=etl availability=implemented stream=deleted_applications]
    deleted-opportunities list - Read Lever Hiring deleted opportunities as ETL records. [intent=etl availability=implemented stream=deleted_opportunities]
    deleted-postings list - Read Lever Hiring deleted postings as ETL records. [intent=etl availability=implemented stream=deleted_postings]
    direct list-all-tags - List all tags [intent=direct_read availability=implemented operation=lever-hiring.list_all_tags]; flags: --page, --page-cursor
    direct retrieve-a-diversity-survey - Retrieve a diversity survey [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_diversity_survey]; flags: --posting, --page, --page-cursor
    direct retrieve-a-feedback-form - Retrieve a feedback form [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_feedback_form]; flags: --opportunity, --feedback, --page, --page-cursor
    direct retrieve-a-feedback-template - Retrieve a feedback template [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_feedback_template]; flags: --feedback-template, --page, --page-cursor
    direct retrieve-a-profile-form - Retrieve a profile form [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_profile_form]; flags: --opportunity, --form, --page, --page-cursor
    direct retrieve-a-profile-form-template - Retrieve a profile form template [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_profile_form_template]; flags: --form-template, --page, --page-cursor
    direct retrieve-a-single-application - Retrieve a single application [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_application]; flags: --opportunity, --application, --page, --page-cursor
    direct retrieve-a-single-archive-reason - Retrieve a single archive reason [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_archive_reason]; flags: --archive-reason, --page, --page-cursor
    direct retrieve-a-single-contact - Retrieve a single contact [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_contact]; flags: --contact, --page, --page-cursor
    direct retrieve-a-single-interview - Retrieve a single interview [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_interview]; flags: --opportunity, --interview, --page, --page-cursor
    direct retrieve-a-single-note - Retrieve a single note [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_note]; flags: --opportunity, --note, --page, --page-cursor
    direct retrieve-a-single-opportunity - Retrieve a single opportunity [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_opportunity]; flags: --opportunity, --page, --page-cursor
    direct retrieve-a-single-panel - Retrieve a single panel [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_panel]; flags: --opportunity, --panel, --page, --page-cursor
    direct retrieve-a-single-posting - Retrieve a single posting [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_posting]; flags: --posting, --page, --page-cursor
    direct retrieve-a-single-referral - Retrieve a single referral [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_referral]; flags: --opportunity, --referral, --page, --page-cursor
    direct retrieve-a-single-requisition - Retrieve a single requisition [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_requisition]; flags: --requisition, --page, --page-cursor
    direct retrieve-a-single-requisition-field - Retrieve a single requisition field [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_requisition_field]; flags: --requisition-field, --page, --page-cursor
    direct retrieve-a-single-stage - Retrieve a single stage [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_stage]; flags: --stage, --page, --page-cursor
    direct retrieve-a-single-user - Retrieve a single user [intent=direct_read availability=implemented operation=lever-hiring.retrieve_a_single_user]; flags: --user, --page, --page-cursor
    direct retrieve-eeo-responses-with-pii - Retrieve EEO responses with PII [intent=direct_read availability=implemented operation=lever-hiring.retrieve_eeo_responses_with_pii]; flags: --page, --page-cursor
    direct retrieve-posting-application-questions - Retrieve posting application questions [intent=direct_read availability=implemented operation=lever-hiring.retrieve_posting_application_questions]; flags: --posting, --page, --page-cursor
    disposition-stages list - Read Lever Hiring disposition stages as ETL records. [intent=etl availability=implemented stream=disposition_stages]
    feedback-templates list - Read Lever Hiring feedback templates as ETL records. [intent=etl availability=implemented stream=feedback_templates]
    form-templates list - Read Lever Hiring form templates as ETL records. [intent=etl availability=implemented stream=form_templates]
    opportunities list - Read Lever Hiring opportunities as ETL records. [intent=etl availability=implemented stream=opportunities]
    opportunity-feedback list - Read Lever Hiring opportunity feedback as ETL records. [intent=etl availability=implemented stream=opportunity_feedback]
    opportunity-file-actions list - Read Lever Hiring opportunity file actions as ETL records. [intent=etl availability=implemented stream=opportunity_file_actions]
    opportunity-forms list - Read Lever Hiring opportunity forms as ETL records. [intent=etl availability=implemented stream=opportunity_forms]
    opportunity-interviews list - Read Lever Hiring opportunity interviews as ETL records. [intent=etl availability=implemented stream=opportunity_interviews]
    opportunity-notes list - Read Lever Hiring opportunity notes as ETL records. [intent=etl availability=implemented stream=opportunity_notes]
    opportunity-offers list - Read Lever Hiring opportunity offers as ETL records. [intent=etl availability=implemented stream=opportunity_offers]
    opportunity-panels list - Read Lever Hiring opportunity panels as ETL records. [intent=etl availability=implemented stream=opportunity_panels]
    opportunity-referrals list - Read Lever Hiring opportunity referrals as ETL records. [intent=etl availability=implemented stream=opportunity_referrals]
    posting-users list - Read Lever Hiring posting users as ETL records. [intent=etl availability=implemented stream=posting_users]
    postings list - Read Lever Hiring postings as ETL records. [intent=etl availability=implemented stream=postings]
    reactivate-user plan - Reactivates a Lever user. [intent=reverse_etl availability=implemented write=reactivate_user]; approval: requires plan, preview, approval, and execute; risk: Reactivates a Lever user. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --user (required)
    requisition-fields list - Read Lever Hiring requisition fields as ETL records. [intent=etl availability=implemented stream=requisition_fields]
    requisitions list - Read Lever Hiring requisitions as ETL records. [intent=etl availability=implemented stream=requisitions]
    sources list - Read Lever Hiring sources as ETL records. [intent=etl availability=implemented stream=sources]
    stages list - Read Lever Hiring stages as ETL records. [intent=etl availability=implemented stream=stages]
    update-feedback plan - Updates a feedback form for an opportunity. [intent=reverse_etl availability=implemented write=update_feedback]; approval: requires plan, preview, approval, and execute; risk: Updates a feedback form for an opportunity. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --feedback (required), --opportunity (required)
    update-feedback-template plan - Updates a feedback template. [intent=reverse_etl availability=implemented write=update_feedback_template]; approval: requires plan, preview, approval, and execute; risk: Updates a feedback template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --feedback_template (required)
    update-form-template plan - Updates a profile form template. [intent=reverse_etl availability=implemented write=update_form_template]; approval: requires plan, preview, approval, and execute; risk: Updates a profile form template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --form_template (required)
    users list - Read Lever Hiring users as ETL records. [intent=etl availability=implemented stream=users]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect lever-hiring

  # Inspect as structured JSON
  pm connectors inspect lever-hiring --json

AGENT WORKFLOW
  - Run pm connectors inspect lever-hiring before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
