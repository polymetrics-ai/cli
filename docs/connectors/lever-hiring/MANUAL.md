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
    fields: archivedAt(), createdAt(), emails(), headline(), id(), lastInteractionAt(), name(), origin(), sources(), stage(), tags(), updatedAt()
  postings:
    primary key: id
    cursor: createdAt
    fields: categories(), createdAt(), hiringManager(), id(), owner(), state(), text(), updatedAt(), user()
  users:
    primary key: id
    cursor: createdAt
    fields: accessRole(), createdAt(), deactivatedAt(), email(), id(), name(), username()
  requisitions:
    primary key: id
    cursor: createdAt
    fields: createdAt(), headcountHired(), headcountTotal(), id(), name(), owner(), requisitionCode(), status(), updatedAt()
  stages:
    primary key: id
    fields: id(), text()
  deleted_applications:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  archive_reasons:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  audit_events:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  disposition_stages:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  feedback_templates:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  deleted_opportunities:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  deleted_postings:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  form_templates:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  requisition_fields:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  sources:
    primary key: id
    fields: createdAt(), id(), name(), text(), updatedAt()
  applications:
    primary key: id
    fields: createdAt(), id(), name(), opportunity_id(), text(), updatedAt()
  opportunity_feedback:
    primary key: id
    fields: createdAt(), id(), name(), opportunity_id(), text(), updatedAt()
  opportunity_interviews:
    primary key: id
    fields: createdAt(), id(), name(), opportunity_id(), text(), updatedAt()
  opportunity_notes:
    primary key: id
    fields: createdAt(), id(), name(), opportunity_id(), text(), updatedAt()
  opportunity_offers:
    primary key: id
    fields: createdAt(), id(), name(), opportunity_id(), text(), updatedAt()
  opportunity_file_actions:
    primary key: id
    fields: createdAt(), id(), name(), opportunity_id(), text(), updatedAt()
  opportunity_panels:
    primary key: id
    fields: createdAt(), id(), name(), opportunity_id(), text(), updatedAt()
  opportunity_forms:
    primary key: id
    fields: createdAt(), id(), name(), opportunity_id(), text(), updatedAt()
  opportunity_referrals:
    primary key: id
    fields: createdAt(), id(), name(), opportunity_id(), text(), updatedAt()
  posting_users:
    primary key: id
    fields: createdAt(), id(), name(), posting_id(), text(), updatedAt()

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
  Read Lever Hiring records and safely plan typed Lever mutations.
  Usage: pm lever-hiring <command> [flags]
  Source CLI: Lever API (Lever Developer documentation fetched 2026-08-01)
  Global flags:
    --credential (string): Credential name to use for the Lever request.
    --connection (string): Alias for --credential.
    --config (string_array): Connector config override as key=value; never pass secret values here.
    --json (boolean): Emit machine-readable JSON output.
    --limit (integer): Maximum ETL records to emit.
    --max-bytes (integer): Maximum bounded direct-read response bytes.
    --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
    --approve (string): Approval token required to execute a reverse-ETL plan.
    --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
  ETL streams
    applications list - Read Lever Hiring applications as ETL records. [intent=etl availability=implemented stream=applications]
    archive-reasons list - Read Lever Hiring archive reasons as ETL records. [intent=etl availability=implemented stream=archive_reasons]
    audit-events list - Read Lever Hiring audit events as ETL records. [intent=etl availability=implemented stream=audit_events]
    deleted-applications list - Read Lever Hiring deleted applications as ETL records. [intent=etl availability=implemented stream=deleted_applications]
    deleted-opportunities list - Read Lever Hiring deleted opportunities as ETL records. [intent=etl availability=implemented stream=deleted_opportunities]
    deleted-postings list - Read Lever Hiring deleted postings as ETL records. [intent=etl availability=implemented stream=deleted_postings]
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
    requisition-fields list - Read Lever Hiring requisition fields as ETL records. [intent=etl availability=implemented stream=requisition_fields]
    requisitions list - Read Lever Hiring requisitions as ETL records. [intent=etl availability=implemented stream=requisitions]
    sources list - Read Lever Hiring sources as ETL records. [intent=etl availability=implemented stream=sources]
    stages list - Read Lever Hiring stages as ETL records. [intent=etl availability=implemented stream=stages]
    users list - Read Lever Hiring users as ETL records. [intent=etl availability=implemented stream=users]
  Bounded direct reads
    direct retrieve-a-single-application - Retrieve a single application [intent=direct_read availability=implemented]; flags: --opportunity, --application
    direct retrieve-a-single-archive-reason - Retrieve a single archive reason [intent=direct_read availability=implemented]; flags: --archive-reason
    direct retrieve-a-single-contact - Retrieve a single contact [intent=direct_read availability=implemented]; flags: --contact
    direct retrieve-eeo-responses-with-pii - Retrieve EEO responses with PII [intent=direct_read availability=implemented]
    direct retrieve-a-feedback-form - Retrieve a feedback form [intent=direct_read availability=implemented]; flags: --opportunity, --feedback
    direct retrieve-a-feedback-template - Retrieve a feedback template [intent=direct_read availability=implemented]; flags: --feedback-template
    direct retrieve-a-single-interview - Retrieve a single interview [intent=direct_read availability=implemented]; flags: --opportunity, --interview
    direct retrieve-a-single-note - Retrieve a single note [intent=direct_read availability=implemented]; flags: --opportunity, --note
    direct retrieve-a-single-opportunity - Retrieve a single opportunity [intent=direct_read availability=implemented]; flags: --opportunity
    direct retrieve-a-single-panel - Retrieve a single panel [intent=direct_read availability=implemented]; flags: --opportunity, --panel
    direct retrieve-a-single-posting - Retrieve a single posting [intent=direct_read availability=implemented]; flags: --posting
    direct list-all-postings-content - List all postings with content expansion [intent=direct_read availability=implemented]
    direct list-all-postings-followers - List all postings with followers expansion [intent=direct_read availability=implemented]
    direct retrieve-posting-application-questions - Retrieve posting application questions [intent=direct_read availability=implemented]; flags: --posting
    direct retrieve-a-profile-form - Retrieve a profile form [intent=direct_read availability=implemented]; flags: --opportunity, --form
    direct retrieve-a-profile-form-template - Retrieve a profile form template [intent=direct_read availability=implemented]; flags: --form-template
    direct retrieve-a-single-referral - Retrieve a single referral [intent=direct_read availability=implemented]; flags: --opportunity, --referral
    direct retrieve-a-single-requisition - Retrieve a single requisition [intent=direct_read availability=implemented]; flags: --requisition
    direct retrieve-a-single-requisition-field - Retrieve a single requisition field [intent=direct_read availability=implemented]; flags: --requisition-field
    direct retrieve-a-single-stage - Retrieve a single stage [intent=direct_read availability=implemented]; flags: --stage
    direct retrieve-a-diversity-survey - Retrieve a diversity survey [intent=direct_read availability=implemented]; flags: --posting
    direct list-all-tags - List all tags [intent=direct_read availability=implemented]
    direct retrieve-a-single-user - Retrieve a single user [intent=direct_read availability=implemented]; flags: --user
  Reverse ETL write plans
    create-feedback-template plan - Creates a feedback template. [intent=reverse_etl availability=implemented write=create_feedback_template]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Creates a feedback template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --text, --instructions, --group, --field-id, --field-type, --field-text, --field-required, --field-description, --field-prompt
    create-form-template plan - Creates a profile form template. [intent=reverse_etl availability=implemented write=create_form_template]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Creates a profile form template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --text, --instructions, --group, --secretbydefault, --field-id, --field-type, --field-text, --field-required, --field-description, --field-prompt
    deactivate-user plan - Deactivates a Lever user. [intent=reverse_etl availability=implemented write=deactivate_user]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Deactivates a Lever user. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --user
    delete-feedback plan - Deletes a feedback form from an opportunity. [intent=reverse_etl availability=implemented write=delete_feedback]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Deletes a feedback form from an opportunity. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --opportunity, --feedback
    delete-feedback-template plan - Deletes a feedback template. [intent=reverse_etl availability=implemented write=delete_feedback_template]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Deletes a feedback template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --feedback-template
    delete-form-template plan - Deletes a profile form template. [intent=reverse_etl availability=implemented write=delete_form_template]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Deletes a profile form template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --form-template
    delete-note plan - Deletes a note from an opportunity. [intent=reverse_etl availability=implemented write=delete_note]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Deletes a note from an opportunity. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --opportunity, --noteid
    delete-requisition plan - Deletes a requisition. [intent=reverse_etl availability=implemented write=delete_requisition]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Deletes a requisition. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --requisition
    delete-requisition-field plan - Deletes a requisition field. [intent=reverse_etl availability=implemented write=delete_requisition_field]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Deletes a requisition field. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --requisition-field
    delete-requisition-field-options plan - Deletes dropdown options for a requisition field. [intent=reverse_etl availability=implemented write=delete_requisition_field_options]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Deletes dropdown options for a requisition field. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --requisition-field
    reactivate-user plan - Reactivates a Lever user. [intent=reverse_etl availability=implemented write=reactivate_user]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Reactivates a Lever user. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --user
    update-feedback plan - Updates a feedback form for an opportunity. [intent=reverse_etl availability=implemented write=update_feedback]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Updates a feedback form for an opportunity. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --opportunity, --feedback, --completedat, --fieldvalue-id, --fieldvalue-value
    update-feedback-template plan - Updates a feedback template. [intent=reverse_etl availability=implemented write=update_feedback_template]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Updates a feedback template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --feedback-template, --text, --instructions, --group, --field-id, --field-type, --field-text, --field-required, --field-description, --field-prompt
    update-form-template plan - Updates a profile form template. [intent=reverse_etl availability=implemented write=update_form_template]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Updates a profile form template. Reverse ETL writes require plan, preview, explicit approval, and execute.; flags: --form-template, --text, --instructions, --group, --secretbydefault, --field-id, --field-type, --field-text, --field-required, --field-description, --field-prompt
  Help topics:
    operation-ledger - Lever operation ledger covers 107 HTTP operations and 10 webhook events from the official documentation.
    write-safety - Lever write commands use reverse ETL plan -> preview -> approval -> execute; destructive actions require typed confirmation.

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
