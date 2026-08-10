---
name: pm-survey-sparrow
description: SurveySparrow connector knowledge and safe action guide.
---

# pm-survey-sparrow

## Purpose

Reads and manages SurveySparrow surveys, contacts, responses, questions, channels, contact lists/properties, reminders, reputation platforms/reviews, survey folders, ticket fields, tickets, teams, roles, variables, webhooks, users, templates, email themes, and expressions through the SurveySparrow API.

## Icon

- id: surveysparrow
- asset: icons/surveysparrow.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.surveysparrow.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- survey_id
- access_token (secret)

## ETL Streams

- surveys:
  - primary key: id
  - fields: id(integer), name(string), survey_type(string)
- contacts:
  - primary key: id
  - fields: email(string), id(integer), name(string)
- responses:
  - primary key: id
  - cursor: completed_time
  - fields: completed_time(string), id(integer), survey_id(integer)
- questions:
  - primary key: id
  - fields: id(integer), question(string), survey_id(integer)
- channels:
  - primary key: id
  - fields: id(integer), name(string), properties(object), status(string), type(string)
- contact_lists:
  - primary key: id
  - fields: description(string), id(integer), name(string)
- contact_properties:
  - primary key: id
  - fields: contact_property_group_id(integer), description(string), group(string), id(integer), label(string), name(string), type(string)
- reminders:
  - primary key: id
  - fields: account_id(integer), after_days(integer), created_at(string), frequency(string), id(integer), message(string), sent_count(integer), subject(string), survey_id(integer), type(string), updated_at(string)
- reputation_platforms:
  - primary key: id
  - fields: id(integer), label(string), logo_url(string), type(string)
- reputation_app_platforms:
  - primary key: id
  - fields: created_at(string), data_fetch_address(string), id(integer), is_active(boolean), location(string), platform_id(integer), updated_at(string)
- reputation_reviews:
  - primary key: id
  - fields: app_platform_id(integer), id(integer), rating(number), review_content(string), review_date(string), review_title(string), reviewer_name(string), reviewer_photo_url(string)
- survey_folders:
  - primary key: id
  - fields: auto_created(boolean), description(string), id(integer), name(string), parent_survey_folder_id(integer), teams(array), users(array), visibility(string)
- ticket_fields:
  - primary key: id
  - fields: created_at(string), description(string), id(integer), internal_name(string), is_default(boolean), mandatory(boolean), name(string), options(array), type(string), updated_at(string)
- tickets:
  - primary key: id
  - fields: agent(object), created_at(string), custom_fields(object), deleted_at(string), description(string), description_html(string), id(integer), priority(object), requester(object), source(object), status(object), subject(string), team(object), template_id(integer), updated_at(string)
- teams:
  - primary key: id
  - fields: account_id(integer), business_hour_id(integer), created_at(string), deleted_at(string), description(string), id(integer), name(string), round_robin_enabled(boolean), type(string), updated_at(string)
- roles:
  - primary key: id
  - fields: account_id(integer), created_at(string), deleted_at(string), description(string), id(integer), label(string), name(string), updated_at(string)
- variables:
  - primary key: id
  - fields: description(string), id(integer), label(string), name(string), type(string)
- webhooks:
  - primary key: id
  - fields: description(string), eventType(string), httpMethod(string), id(integer), name(string), objectType(string), url(string)
- users:
  - primary key: id
  - fields: admin(boolean), agency_owner(boolean), email(string), id(integer), name(string), owner(boolean), phone(string), role_id(integer), verified(boolean)
- templates:
  - primary key: id
  - fields: created_at(string), deleted_at(string), description(string), id(integer), name(string), updated_at(string)
- email_themes:
  - primary key: id
  - fields: created_at(string), id(integer), is_public(boolean), name(string), properties(object), updated_at(string)
- expressions:
  - primary key: id
  - fields: id(integer), name(string), representation(array)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_survey:
  - endpoint: POST /surveys
  - required fields: name, survey_type
  - risk: external mutation; approval required
- update_survey:
  - endpoint: PATCH /surveys/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- create_contact:
  - endpoint: POST /contacts
  - risk: external mutation; approval required
- update_contact:
  - endpoint: PUT /contacts/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_contact:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_question:
  - endpoint: POST /questions
  - required fields: survey_id, text, type
  - risk: external mutation; approval required
- update_question:
  - endpoint: PUT /questions/{{ record.question_id }}
  - required fields: question_id, survey_id
  - risk: external mutation; approval required
- delete_question:
  - endpoint: DELETE /questions/{{ record.question_id }}
  - required fields: question_id
  - risk: irreversible external deletion; approval required
- create_contact_list:
  - endpoint: POST /contact_lists
  - required fields: name
  - risk: external mutation; approval required
- update_contact_list:
  - endpoint: PATCH /contact_lists/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_contact_list:
  - endpoint: DELETE /contact_lists/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_contact_property:
  - endpoint: POST /contact_properties
  - required fields: type, label
  - risk: external mutation; approval required
- update_contact_property:
  - endpoint: PATCH /contact_properties/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_contact_property:
  - endpoint: DELETE /contact_properties/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_survey_folder:
  - endpoint: POST /survey_folders
  - required fields: name
  - risk: external mutation; approval required
- update_survey_folder:
  - endpoint: PATCH /survey_folders/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_survey_folder:
  - endpoint: DELETE /survey_folders/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_team:
  - endpoint: POST /teams
  - required fields: name
  - risk: external mutation; approval required
- create_ticket:
  - endpoint: POST /tickets
  - required fields: subject, priority, status
  - risk: external mutation; approval required
- update_ticket:
  - endpoint: PUT /tickets/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_ticket:
  - endpoint: DELETE /tickets/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_webhook:
  - endpoint: POST /webhooks
  - required fields: url, survey_id, http_method
  - risk: external mutation; approval required
- update_webhook:
  - endpoint: PUT /webhooks/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_webhook:
  - endpoint: DELETE /webhooks/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_user:
  - endpoint: POST /users
  - required fields: name, email, role_id
  - risk: external mutation creating a live user account with console access; approval required
- update_user:
  - endpoint: PATCH /users/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_user:
  - endpoint: DELETE /users/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a user account; approval required
- create_reminder:
  - endpoint: POST /reminders
  - required fields: channel_id, survey_id, frequency, type, interval, embed_first_question, custom_footer
  - risk: external mutation; approval required
- delete_reminder:
  - endpoint: DELETE /reminders/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_variable:
  - endpoint: POST /variables
  - required fields: survey_id, label, name, type
  - risk: external mutation; approval required
- delete_variable:
  - endpoint: DELETE /variables/{{ record.variable_id }}
  - required fields: variable_id
  - risk: irreversible external deletion; approval required
- create_channel:
  - endpoint: POST /channels
  - required fields: type
  - risk: external mutation; approval required
- delete_channel:
  - endpoint: DELETE /channels/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required

## Security

- read risk: external SurveySparrow API read of survey, contact, response, question, and 18 additional catalog resource types
- write risk: external mutation of SurveySparrow surveys, contacts, questions, contact lists/properties, survey folders, teams, tickets, webhooks, users, reminders, variables, and channels, including irreversible deletes and live-user-account creation/deletion
- approval: read: none; write: required for all mutation actions
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run SurveySparrow's declared streams and reverse-ETL actions.
- Usage: pm survey-sparrow <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete v3 audit-logs events id - Documented DELETE /v3/audit_logs/events/:id (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.delete.v3-audit-logs-events-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v3 responses id - Documented DELETE /v3/responses/:id (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.delete.v3-responses-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v3 sections section-id - Documented DELETE /v3/sections/:section_id (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.delete.v3-sections-section-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v3 translation - Documented DELETE /v3/translation (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.delete.v3-translation]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get v3 audit-logs - Documented GET /v3/audit_logs (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-audit-logs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 audit-logs events - Documented GET /v3/audit_logs/events (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-audit-logs-events]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 audit-logs id - Documented GET /v3/audit_logs/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-audit-logs-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 channels id - Documented GET /v3/channels/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-channels-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 contact-lists id - Documented GET /v3/contact_lists/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-contact-lists-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 contacts id - Documented GET /v3/contacts/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-contacts-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 contacts status token - Documented GET /v3/contacts/status/:token (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-contacts-status-token]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 languages - Documented GET /v3/languages (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-languages]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 metrics - Documented GET /v3/metrics (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-metrics]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 metrics responses - Documented GET /v3/metrics/responses (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-metrics-responses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 reminders id - Documented GET /v3/reminders/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-reminders-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 reports - Documented GET /v3/reports (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-reports]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 reputation app-platforms id - Documented GET /v3/reputation/app_platforms/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-reputation-app-platforms-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 reputation platforms id - Documented GET /v3/reputation/platforms/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-reputation-platforms-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 reputation reviews id - Documented GET /v3/reputation/reviews/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-reputation-reviews-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 response-properties - Documented GET /v3/response_properties (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-response-properties]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 responses id - Documented GET /v3/responses/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-responses-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 responses status token - Documented GET /v3/responses/status/:token (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-responses-status-token]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 survey approvers - Documented GET /v3/survey/approvers (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-survey-approvers]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 survey subject evaluators - Documented GET /v3/survey/subject/evaluators (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-survey-subject-evaluators]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 survey subject report - Documented GET /v3/survey/subject/report (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-survey-subject-report]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 survey subjects - Documented GET /v3/survey/subjects (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-survey-subjects]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 survey-folders id - Documented GET /v3/survey_folders/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-survey-folders-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 surveys id - Documented GET /v3/surveys/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-surveys-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 targets - Documented GET /v3/targets (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-targets]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 ticket-fields id - Documented GET /v3/ticket_fields/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-ticket-fields-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 tickets batch status token - Documented GET /v3/tickets/batch/status/:token (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-tickets-batch-status-token]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 tickets id - Documented GET /v3/tickets/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-tickets-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 tickets id comments - Documented GET /v3/tickets/:id/comments (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-tickets-id-comments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 translation export - Documented GET /v3/translation/export (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-translation-export]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 users id - Documented GET /v3/users/:id (not implemented) [intent=direct_read availability=not_implemented operation=survey-sparrow.get.v3-users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post v3 audit-logs events - Documented POST /v3/audit_logs/events (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-audit-logs-events]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 channels create-unique-links - Documented POST /v3/channels/create_unique_links (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-channels-create-unique-links]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 channels id clone - Documented POST /v3/channels/:id/clone (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-channels-id-clone]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 channels id summary - Documented POST /v3/channels/:id/summary (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-channels-id-summary]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 contacts batch - Documented POST /v3/contacts/batch (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-contacts-batch]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 language - Documented POST /v3/language (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-language]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 responses - Documented POST /v3/responses (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-responses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 responses batch - Documented POST /v3/responses/batch (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-responses-batch]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 responses new - Documented POST /v3/responses/new (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-responses-new]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 sections - Documented POST /v3/sections (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-sections]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 survey invite - Documented POST /v3/survey/invite (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-survey-invite]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 surveys id clone - Documented POST /v3/surveys/:id/clone (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-surveys-id-clone]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 tickets batch - Documented POST /v3/tickets/batch (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-tickets-batch]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 tickets id comments - Documented POST /v3/tickets/:id/comments (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-tickets-id-comments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 translation - Documented POST /v3/translation (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-translation]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 variables batch - Documented POST /v3/variables/batch (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.post.v3-variables-batch]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 channels id - Documented PUT /v3/channels/:id (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.put.v3-channels-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 responses id - Documented PUT /v3/responses/:id (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.put.v3-responses-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 responses response-id complete - Documented PUT /v3/responses/:response_id/complete (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.put.v3-responses-response-id-complete]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 responses response-id update - Documented PUT /v3/responses/:response_id/update (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.put.v3-responses-response-id-update]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 sections section-id - Documented PUT /v3/sections/:section_id (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.put.v3-sections-section-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 survey subject updateinvite - Documented PUT /v3/survey/subject/updateinvite (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.put.v3-survey-subject-updateinvite]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 translation - Documented PUT /v3/translation (not implemented) [intent=direct_write availability=not_implemented operation=survey-sparrow.put.v3-translation]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - channels list - Run the channels ETL stream [intent=etl availability=implemented stream=channels]
  - contact lists list - Run the contact lists ETL stream [intent=etl availability=implemented stream=contact_lists]
  - contact properties list - Run the contact properties ETL stream [intent=etl availability=implemented stream=contact_properties]
  - contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
  - create channel apply - Plan and execute the create channel reverse-ETL action [intent=reverse_etl availability=implemented write=create_channel]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --type (required)
  - create contact apply - Plan and execute the create contact reverse-ETL action [intent=reverse_etl availability=implemented write=create_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required
  - create contact list apply - Plan and execute the create contact list reverse-ETL action [intent=reverse_etl availability=implemented write=create_contact_list]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required)
  - create contact property apply - Plan and execute the create contact property reverse-ETL action [intent=reverse_etl availability=implemented write=create_contact_property]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --label (required), --type (required)
  - create question apply - Plan and execute the create question reverse-ETL action [intent=reverse_etl availability=implemented write=create_question]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --survey_id (required), --text (required), --type (required)
  - create reminder apply - Plan and execute the create reminder reverse-ETL action [intent=reverse_etl availability=implemented write=create_reminder]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --channel_id (required), --custom_footer (required), --embed_first_question (required), --frequency (required), --interval (required), --survey_id (required), --type (required)
  - create survey apply - Plan and execute the create survey reverse-ETL action [intent=reverse_etl availability=implemented write=create_survey]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required), --survey_type (required)
  - create survey folder apply - Plan and execute the create survey folder reverse-ETL action [intent=reverse_etl availability=implemented write=create_survey_folder]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required)
  - create team apply - Plan and execute the create team reverse-ETL action [intent=reverse_etl availability=implemented write=create_team]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --name (required)
  - create ticket apply - Plan and execute the create ticket reverse-ETL action [intent=reverse_etl availability=implemented write=create_ticket]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --priority (required), --status (required), --subject (required)
  - create user apply - Plan and execute the create user reverse-ETL action [intent=reverse_etl availability=implemented write=create_user]; approval: requires plan, preview, approval, and execute; risk: external mutation creating a live user account with console access; approval required; flags: --email (required), --name (required), --role_id (required)
  - create variable apply - Plan and execute the create variable reverse-ETL action [intent=reverse_etl availability=implemented write=create_variable]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --label (required), --name (required), --survey_id (required), --type (required)
  - create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --http_method (required), --survey_id (required), --url (required)
  - delete channel apply - Plan and execute the delete channel reverse-ETL action [intent=reverse_etl availability=implemented write=delete_channel]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --id (required)
  - delete contact apply - Plan and execute the delete contact reverse-ETL action [intent=reverse_etl availability=implemented write=delete_contact]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --id (required)
  - delete contact list apply - Plan and execute the delete contact list reverse-ETL action [intent=reverse_etl availability=implemented write=delete_contact_list]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --id (required)
  - delete contact property apply - Plan and execute the delete contact property reverse-ETL action [intent=reverse_etl availability=implemented write=delete_contact_property]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --id (required)
  - delete question apply - Plan and execute the delete question reverse-ETL action [intent=reverse_etl availability=implemented write=delete_question]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --question_id (required)
  - delete reminder apply - Plan and execute the delete reminder reverse-ETL action [intent=reverse_etl availability=implemented write=delete_reminder]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --id (required)
  - delete survey folder apply - Plan and execute the delete survey folder reverse-ETL action [intent=reverse_etl availability=implemented write=delete_survey_folder]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --id (required)
  - delete ticket apply - Plan and execute the delete ticket reverse-ETL action [intent=reverse_etl availability=implemented write=delete_ticket]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --id (required)
  - delete user apply - Plan and execute the delete user reverse-ETL action [intent=reverse_etl availability=implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion of a user account; approval required; flags: --id (required)
  - delete variable apply - Plan and execute the delete variable reverse-ETL action [intent=reverse_etl availability=implemented write=delete_variable]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --variable_id (required)
  - delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion; approval required; flags: --id (required)
  - email themes list - Run the email themes ETL stream [intent=etl availability=implemented stream=email_themes]
  - expressions list - Run the expressions ETL stream [intent=etl availability=implemented stream=expressions]
  - questions list - Run the questions ETL stream [intent=etl availability=implemented stream=questions]
  - reminders list - Run the reminders ETL stream [intent=etl availability=implemented stream=reminders]
  - reputation app platforms list - Run the reputation app platforms ETL stream [intent=etl availability=implemented stream=reputation_app_platforms]
  - reputation platforms list - Run the reputation platforms ETL stream [intent=etl availability=implemented stream=reputation_platforms]
  - reputation reviews list - Run the reputation reviews ETL stream [intent=etl availability=implemented stream=reputation_reviews]
  - responses list - Run the responses ETL stream [intent=etl availability=implemented stream=responses]
  - roles list - Run the roles ETL stream [intent=etl availability=implemented stream=roles]
  - survey folders list - Run the survey folders ETL stream [intent=etl availability=implemented stream=survey_folders]
  - surveys list - Run the surveys ETL stream [intent=etl availability=implemented stream=surveys]
  - teams list - Run the teams ETL stream [intent=etl availability=implemented stream=teams]
  - templates list - Run the templates ETL stream [intent=etl availability=implemented stream=templates]
  - ticket fields list - Run the ticket fields ETL stream [intent=etl availability=implemented stream=ticket_fields]
  - tickets list - Run the tickets ETL stream [intent=etl availability=implemented stream=tickets]
  - update contact apply - Plan and execute the update contact reverse-ETL action [intent=reverse_etl availability=implemented write=update_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
  - update contact list apply - Plan and execute the update contact list reverse-ETL action [intent=reverse_etl availability=implemented write=update_contact_list]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
  - update contact property apply - Plan and execute the update contact property reverse-ETL action [intent=reverse_etl availability=implemented write=update_contact_property]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
  - update question apply - Plan and execute the update question reverse-ETL action [intent=reverse_etl availability=implemented write=update_question]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --question_id (required), --survey_id (required)
  - update survey apply - Plan and execute the update survey reverse-ETL action [intent=reverse_etl availability=implemented write=update_survey]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
  - update survey folder apply - Plan and execute the update survey folder reverse-ETL action [intent=reverse_etl availability=implemented write=update_survey_folder]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
  - update ticket apply - Plan and execute the update ticket reverse-ETL action [intent=reverse_etl availability=implemented write=update_ticket]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
  - update user apply - Plan and execute the update user reverse-ETL action [intent=reverse_etl availability=implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
  - update webhook apply - Plan and execute the update webhook reverse-ETL action [intent=reverse_etl availability=implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
  - variables list - Run the variables ETL stream [intent=etl availability=implemented stream=variables]
  - webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]

## Commands

### Inspect as a manual

```bash
pm connectors inspect survey-sparrow
```

### Inspect as structured JSON

```bash
pm connectors inspect survey-sparrow --json
```

## Agent Rules

- Run pm connectors inspect survey-sparrow before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
