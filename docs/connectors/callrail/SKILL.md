---
name: pm-callrail
description: CallRail connector knowledge and safe action guide.
---

# pm-callrail

## Purpose

Reads and writes CallRail call tracking data (calls, companies, users, tags, trackers, form submissions, text messages, notifications, integrations, and more) through the CallRail v3 REST API.

## Icon

- id: callrail
- asset: icons/callrail.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://apidocs.callrail.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_id
- base_url
- company_id
- start_date
- api_key (secret)

## ETL Streams

- calls:
  - primary key: id
  - cursor: start_time
  - fields: answered(boolean), business_phone_number(string), company_id(string), customer_city(string), customer_country(string), customer_name(string), customer_phone_number(string), customer_state(string), direction(string), duration(integer), id(string), recording(string), start_time(string), tracking_phone_number(string), voicemail(boolean)
- companies:
  - primary key: id
  - cursor: created_at
  - fields: callscore_enabled(boolean), created_at(string), disabled_at(string), dni_active(boolean), id(string), name(string), status(string), time_zone(string)
- users:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), email(string), first_name(string), id(string), last_name(string), name(string), role(string)
- text_messages:
  - primary key: id
  - cursor: last_message_at
  - fields: company_id(string), customer_name(string), customer_phone_number(string), id(string), initial_tracker_id(string), last_message_at(string), state(string), tracking_phone_number(string)
- accounts:
  - primary key: id
  - fields: hipaa_account(boolean), id(string), name(string), outbound_recording_enabled(boolean)
- tags:
  - primary key: id
  - cursor: created_at
  - fields: background_color(string), color(string), company_id(string), created_at(string), id(string), name(string), status(string), tag_level(string)
- trackers:
  - primary key: id
  - cursor: created_at
  - fields: company_id(string), company_name(string), created_at(string), destination_number(string), disabled_at(string), id(string), name(string), sms_enabled(boolean), sms_supported(boolean), status(string), tracking_numbers(array), type(string), whisper_message(string)
- form_submissions:
  - primary key: id
  - cursor: submitted_at
  - fields: campaign(string), company_id(string), customer_email(string), customer_name(string), customer_phone_number(string), first_form(boolean), form_url(string), id(string), keywords(string), landing_page_url(string), medium(string), person_id(string), referrer(string), referring_url(string), source(string), submitted_at(string)
- integrations:
  - primary key: id
  - fields: config(object), id(integer), state(string), type(string)
- integration_filters:
  - primary key: id
  - fields: call_type(string), company_id(string), id(integer), integration_id(integer), integration_type(string), lead_status(string), max_duration(integer), min_duration(integer), tracker_ids(array)
- notifications:
  - primary key: id
  - fields: alert_type(string), call_enabled(boolean), company_id(string), company_name(string), id(integer), name(string), send_desktop(boolean), send_email(boolean), send_push(boolean), sms_enabled(boolean), tracker_id(string), tracker_name(string), user_id(string)
- caller_ids:
  - primary key: id
  - cursor: created_at
  - fields: company_id(string), created_at(string), id(integer), name(string), phone_number(string), validation_code(string), verified(boolean)
- sms_threads:
  - primary key: id
  - fields: company_id(string), company_time_zone(string), current_tracker_id(string), current_tracking_number(string), customer_name(string), customer_phone_number(string), id(string), initial_tracker_id(string), initial_tracking_number(string), lead_qualification(string), notes(string), state(string), tags(array), value(number)
- message_flows:
  - primary key: id
  - fields: id(string), initial_step_id(string), name(string), steps(object), tracker_ids(array), updated_at(string)
- leads:
  - primary key: id
  - cursor: created_at
  - fields: company_id(string), company_name(string), created_at(string), email(string), id(string), name(string), phone(string)
- page_views:
  - primary key: call_id, created_at
  - cursor: created_at
  - fields: call_id(string), created_at(string), page_url(string), referrer_url(string)
- lead_timeline:
  - primary key: lead_id
  - fields: campaign(string), customer_name(string), customer_phone_number(string), first_touch(object), last_touch(object), lead_creation(object), lead_id(string), lead_qualification(object), medium(string), source(string), tags(array), total_interactions(integer), transcript(boolean), voice_assist(boolean)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_tag:
  - endpoint: POST /a/{{ config.account_id }}/tags.json
  - required fields: name
  - risk: creates a new call/text tag definition visible account- or company-wide; low-risk external mutation, no approval required
- update_tag:
  - endpoint: PUT /a/{{ config.account_id }}/tags/{{ record.id }}.json
  - required fields: id
  - risk: renames/recolors/disables a tag; renaming changes the tag everywhere it is currently assigned; low-risk external mutation, no approval required
- delete_tag:
  - endpoint: DELETE /a/{{ config.account_id }}/tags/{{ record.id }}.json
  - required fields: id
  - risk: permanently removes a tag, including from every call/text interaction it has been applied to; irreversible, approval recommended
- create_company:
  - endpoint: POST /a/{{ config.account_id }}/companies.json
  - required fields: name
  - risk: creates a new company (a billable tracking entity) within the account; approval recommended
- update_company:
  - endpoint: PUT /a/{{ config.account_id }}/companies/{{ record.id }}.json
  - required fields: id
  - risk: updates company configuration; setting status to disabled deactivates all of the company's tracking numbers and its dynamic-number-insertion script — approval recommended for status changes
- create_user:
  - endpoint: POST /a/{{ config.account_id }}/users.json
  - required fields: first_name, last_name, email, role
  - risk: creates a new CallRail user and emails them a password-setup prompt; requires an administrator-scoped API key; approval recommended
- update_user:
  - endpoint: PUT /a/{{ config.account_id }}/users/{{ record.id }}.json
  - required fields: id
  - risk: updates a user's profile/role/company access; name/email changes are restricted to the API key's own owning user by CallRail; approval recommended for role/company changes
- delete_user:
  - endpoint: DELETE /a/{{ config.account_id }}/users/{{ record.id }}.json
  - required fields: id
  - risk: permanently removes a user's access to the account; requires an administrator-scoped API key; irreversible, approval required
- update_call:
  - endpoint: PUT /a/{{ config.account_id }}/calls/{{ record.id }}.json
  - required fields: id
  - risk: applies tags/notes/lead-status/value/customer-name metadata to an existing call record; low-risk external mutation, no approval required
- create_outbound_call:
  - endpoint: POST /a/{{ config.account_id }}/calls.json
  - required fields: caller_id, business_phone_number, customer_phone_number
  - risk: places a real outbound phone call connecting a business and a customer number (US/Canada only); a real-world side effect outside the CallRail account itself, approval required
- send_text_message:
  - endpoint: POST /a/{{ config.account_id }}/text-messages.json
  - required fields: company_id, customer_phone_number, tracking_number, content
  - risk: sends a real SMS/MMS text message to a customer's phone (subject to 10DLC business-registration compliance rules); a real-world side effect outside the CallRail account itself, approval required. Direct file-upload MMS (multipart media_file) is out of scope — see api_surface.json/docs.md; the media_url variant covers publicly-hosted-image MMS instead.
- create_integration:
  - endpoint: POST /a/{{ config.account_id }}/integrations.json
  - required fields: type, company_id
  - risk: creates and activates a Webhooks or Custom-cookie-capture integration for a company (the only 2 integration types the API can create); approval recommended since Webhooks integrations push call data to an external URL
- update_integration:
  - endpoint: PUT /a/{{ config.account_id }}/integrations/{{ record.id }}.json
  - required fields: id
  - risk: updates an integration's active/disabled state or its webhook/cookie-capture configuration; approval recommended
- disable_integration:
  - endpoint: DELETE /a/{{ config.account_id }}/integrations/{{ record.id }}.json
  - required fields: id
  - risk: disables (the docs' own term; not a hard delete) an integration; stops any external data flow it previously drove; approval recommended
- create_integration_filter:
  - endpoint: POST /a/{{ config.account_id }}/integration_triggers.json
  - required fields: company_id, integration_id
  - risk: adds a filter narrowing which calls trigger an existing integration; low-risk external mutation, no approval required
- update_integration_filter:
  - endpoint: PUT /a/{{ config.account_id }}/integration_triggers/{{ record.id }}.json
  - required fields: id
  - risk: updates an integration filter's trigger criteria; low-risk external mutation, no approval required
- delete_integration_filter:
  - endpoint: DELETE /a/{{ config.account_id }}/integration_triggers/{{ record.id }}.json
  - required fields: id
  - risk: removes a filter; the parent integration keeps firing for every call, unfiltered, once this is removed; low-risk, no approval required
- create_notification:
  - endpoint: POST /a/{{ config.account_id }}/notifications.json
  - risk: creates a call/text alert subscription for a user; low-risk external mutation, no approval required
- update_notification:
  - endpoint: PUT /a/{{ config.account_id }}/notifications/{{ record.id }}.json
  - required fields: id
  - risk: updates an existing notification's scope/channel settings; low-risk external mutation, no approval required
- delete_notification:
  - endpoint: DELETE /a/{{ config.account_id }}/notifications/{{ record.id }}.json
  - required fields: id
  - risk: permanently removes a notification subscription (restricted to notifications managed by the current user); irreversible, low-risk, no approval required
- create_caller_id:
  - endpoint: POST /a/{{ config.account_id }}/caller_ids.json
  - required fields: company_id, phone_number, name
  - risk: registers an outbound caller-id number and immediately triggers a real verification phone call to it; a real-world side effect, approval required
- delete_caller_id:
  - endpoint: DELETE /a/{{ config.account_id }}/caller_ids/{{ record.id }}.json
  - required fields: id
  - risk: removes an outbound caller id from the company; irreversible, low-risk, no approval required
- update_sms_thread:
  - endpoint: PUT /a/{{ config.account_id }}/sms-threads/{{ record.id }}.json
  - required fields: id
  - risk: applies notes/value/tags/lead-qualification metadata to an existing SMS thread; low-risk external mutation, no approval required
- update_tracker:
  - endpoint: PUT /a/{{ config.account_id }}/trackers/{{ record.id }}.json
  - required fields: id
  - risk: reconfigures an existing (already-provisioned) session or source tracker's call flow, whisper message, SMS setting, or source rules; does not provision/deprovision a phone number itself, unlike create/disable; low-risk external mutation, no approval required
- create_message_flow:
  - endpoint: POST /a/{{ config.account_id }}/message-flows.json
  - required fields: company_id, name, initial_step_id, steps
  - risk: creates a new automated SMS message flow (a step-graph of tag/response actions) for a company; low-risk external mutation, no approval required
- update_message_flow:
  - endpoint: PUT /a/{{ config.account_id }}/message-flows.json
  - required fields: id, initial_step_id, steps
  - risk: replaces an existing message flow's step graph; the docs' own endpoint takes no {message_flow_id} path segment, identifying the flow purely via the body's id field; low-risk external mutation, no approval required
- delete_message_flow:
  - endpoint: DELETE /a/{{ config.account_id }}/message-flows/{{ record.id }}.json
  - required fields: id
  - risk: permanently removes a message flow; any tracker still referencing it stops running the automated SMS steps; irreversible, approval recommended

## Security

- read risk: external CallRail API read of call tracking analytics, contact, and configuration data
- write risk: external mutation of CallRail account configuration (tags, companies, users, notifications, outbound caller ids, message flows, integration filters), call/lead metadata (call tags, lead status, value), and outbound communications (placing outbound calls, sending SMS)
- approval: required for outbound-communication and account-configuration writes (placing calls, sending texts, disabling companies, deleting users/caller-ids); tag/metadata-only writes are lower risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run CallRail's declared streams and reverse-ETL actions.
- Usage: pm callrail <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - accounts list - Run the accounts ETL stream [intent=etl availability=implemented stream=accounts]
  - api delete a account-id companies company-id json - Documented DELETE /a/{account_id}/companies/{company_id}.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.delete.a-account-id-companies-company-id-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete a account-id summary-emails summary-email-id json - Documented DELETE /a/{account_id}/summary_emails/{summary_email_id}.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.delete.a-account-id-summary-emails-summary-email-id-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete a account-id trackers tracker-id json - Documented DELETE /a/{account_id}/trackers/{tracker_id}.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.delete.a-account-id-trackers-tracker-id-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v3 a account-id integration-triggers integration-triggers-id json - Documented DELETE /v3/a/{account_id}/integration_triggers/{integration_triggers_id}.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.delete.v3-a-account-id-integration-triggers-integration-triggers-id-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get a account-id caller-ids caller-ids-id json - Documented GET /a/{account_id}/caller_ids/{caller_ids_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-caller-ids-caller-ids-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id calls call-id json - Documented GET /a/{account_id}/calls/{call_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-calls-call-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id calls call-id recording-json - Documented GET /a/{account_id}/calls/{call_id}/recording.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-calls-call-id-recording-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id calls summary-json - Documented GET /a/{account_id}/calls/summary.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-calls-summary-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id calls timeseries-json - Documented GET /a/{account_id}/calls/timeseries.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-calls-timeseries-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id companies company-id json - Documented GET /a/{account_id}/companies/{company_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-companies-company-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id form-submissions ignored-fields-json - Documented GET /a/{account_id}/form_submissions/ignored_fields.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-form-submissions-ignored-fields-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id forms summary-json - Documented GET /a/{account_id}/forms/summary.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-forms-summary-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id integration-triggers integration-trigger-id json - Documented GET /a/{account_id}/integration_triggers/{integration_trigger_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-integration-triggers-integration-trigger-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id integrations configurations - Documented GET /a/{account_id}/integrations/configurations (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-integrations-configurations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id integrations integration-id json - Documented GET /a/{account_id}/integrations/{integration_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-integrations-integration-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id json - Documented GET /a/{account_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id message-flows configurations - Documented GET /a/{account_id}/message-flows/configurations (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-message-flows-configurations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id message-flows message-flow-id json - Documented GET /a/{account_id}/message-flows/{message_flow_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-message-flows-message-flow-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id sms-threads thread-id json - Documented GET /a/{account_id}/sms-threads/{thread_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-sms-threads-thread-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id summary-emails summary-email-id json - Documented GET /a/{account_id}/summary_emails/{summary_email_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-summary-emails-summary-email-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id summary-emails-json - Documented GET /a/{account_id}/summary_emails.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-summary-emails-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id text-messages conversation-id json - Documented GET /a/{account_id}/text-messages/{conversation_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-text-messages-conversation-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id trackers request-number-json - Documented GET /a/{account_id}/trackers/request_number.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-trackers-request-number-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id trackers tracker-id json - Documented GET /a/{account_id}/trackers/{tracker_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-trackers-tracker-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get a account-id users user-id json - Documented GET /a/{account_id}/users/{user_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.a-account-id-users-user-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 a account-id calls 444941612-json - Documented GET /v3/a/{account_id}/calls/444941612.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.v3-a-account-id-calls-444941612-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 a account-id integration-triggers integration-triggers-id json - Documented GET /v3/a/{account_id}/integration_triggers/{integration_triggers_id}.json (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.v3-a-account-id-integration-triggers-integration-triggers-id-json]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 a account-id summary-emails - Documented GET /v3/a/{account_id}/summary_emails (not implemented) [intent=direct_read availability=not_implemented operation=callrail.get.v3-a-account-id-summary-emails]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post a account-id form-submissions-json - Documented POST /a/{account_id}/form_submissions.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.post.a-account-id-form-submissions-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post a account-id summary-emails-json - Documented POST /a/{account_id}/summary_emails.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.post.a-account-id-summary-emails-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post a account-id trackers-json - Documented POST /a/{account_id}/trackers.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.post.a-account-id-trackers-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post a agency-id trackers-json - Documented POST /a/{agency_id}/trackers.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.post.a-agency-id-trackers-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 a account-id companies bulk-update-json - Documented POST /v3/a/{account_id}/companies/bulk_update.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.post.v3-a-account-id-companies-bulk-update-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 a account-id form-submissions ignored-fields-json - Documented POST /v3/a/{account_id}/form_submissions/ignored_fields.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.post.v3-a-account-id-form-submissions-ignored-fields-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put a account-id companies bulk-update-json - Documented PUT /a/{account_id}/companies/bulk_update.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.put.a-account-id-companies-bulk-update-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put a account-id form-submissions form-submission-id json - Documented PUT /a/{account_id}/form_submissions/{form_submission_id}.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.put.a-account-id-form-submissions-form-submission-id-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put a account-id summary-emails summary-email-id json - Documented PUT /a/{account_id}/summary_emails/{summary_email_id}.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.put.a-account-id-summary-emails-summary-email-id-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 a account-id message-flows message-flow-id json - Documented PUT /v3/a/{account_id}/message-flows/{message_flow_id}.json (not implemented) [intent=direct_write availability=not_implemented operation=callrail.put.v3-a-account-id-message-flows-message-flow-id-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - caller ids list - Run the caller ids ETL stream [intent=etl availability=implemented stream=caller_ids]
  - calls list - Run the calls ETL stream [intent=etl availability=implemented stream=calls]
  - companies list - Run the companies ETL stream [intent=etl availability=implemented stream=companies]
  - create caller id apply - Plan and execute the create caller id reverse-ETL action [intent=reverse_etl availability=implemented write=create_caller_id]; approval: requires plan, preview, approval, and execute; risk: registers an outbound caller-id number and immediately triggers a real verification phone call to it; a real-world side effect, approval required; flags: --company_id (required), --name (required), --phone_number (required)
  - create company apply - Plan and execute the create company reverse-ETL action [intent=reverse_etl availability=implemented write=create_company]; approval: requires plan, preview, approval, and execute; risk: creates a new company (a billable tracking entity) within the account; approval recommended; flags: --name (required)
  - create integration apply - Plan and execute the create integration reverse-ETL action [intent=reverse_etl availability=implemented write=create_integration]; approval: requires plan, preview, approval, and execute; risk: creates and activates a Webhooks or Custom-cookie-capture integration for a company (the only 2 integration types the API can create); approval recommended since Webhooks integrations push call data to an external URL; flags: --company_id (required), --type (required)
  - create integration filter apply - Plan and execute the create integration filter reverse-ETL action [intent=reverse_etl availability=implemented write=create_integration_filter]; approval: requires plan, preview, approval, and execute; risk: adds a filter narrowing which calls trigger an existing integration; low-risk external mutation, no approval required; flags: --company_id (required), --integration_id (required)
  - create message flow apply - Plan and execute the create message flow reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_message_flow]; approval: requires plan, preview, approval, and execute; risk: creates a new automated SMS message flow (a step-graph of tag/response actions) for a company; low-risk external mutation, no approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create notification apply - Plan and execute the create notification reverse-ETL action [intent=reverse_etl availability=implemented write=create_notification]; approval: requires plan, preview, approval, and execute; risk: creates a call/text alert subscription for a user; low-risk external mutation, no approval required
  - create outbound call apply - Plan and execute the create outbound call reverse-ETL action [intent=reverse_etl availability=implemented write=create_outbound_call]; approval: requires plan, preview, approval, and execute; risk: places a real outbound phone call connecting a business and a customer number (US/Canada only); a real-world side effect outside the CallRail account itself, approval required; flags: --business_phone_number (required), --caller_id (required), --customer_phone_number (required)
  - create tag apply - Plan and execute the create tag reverse-ETL action [intent=reverse_etl availability=implemented write=create_tag]; approval: requires plan, preview, approval, and execute; risk: creates a new call/text tag definition visible account- or company-wide; low-risk external mutation, no approval required; flags: --name (required)
  - create user apply - Plan and execute the create user reverse-ETL action [intent=reverse_etl availability=implemented write=create_user]; approval: requires plan, preview, approval, and execute; risk: creates a new CallRail user and emails them a password-setup prompt; requires an administrator-scoped API key; approval recommended; flags: --email (required), --first_name (required), --last_name (required), --role (required)
  - delete caller id apply - Plan and execute the delete caller id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_caller_id]; approval: requires plan, preview, approval, and execute; risk: removes an outbound caller id from the company; irreversible, low-risk, no approval required; flags: --id (required)
  - delete integration filter apply - Plan and execute the delete integration filter reverse-ETL action [intent=reverse_etl availability=implemented write=delete_integration_filter]; approval: requires plan, preview, approval, and execute; risk: removes a filter; the parent integration keeps firing for every call, unfiltered, once this is removed; low-risk, no approval required; flags: --id (required)
  - delete message flow apply - Plan and execute the delete message flow reverse-ETL action [intent=reverse_etl availability=implemented write=delete_message_flow]; approval: requires plan, preview, approval, and execute; risk: permanently removes a message flow; any tracker still referencing it stops running the automated SMS steps; irreversible, approval recommended; flags: --id (required)
  - delete notification apply - Plan and execute the delete notification reverse-ETL action [intent=reverse_etl availability=implemented write=delete_notification]; approval: requires plan, preview, approval, and execute; risk: permanently removes a notification subscription (restricted to notifications managed by the current user); irreversible, low-risk, no approval required; flags: --id (required)
  - delete tag apply - Plan and execute the delete tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_tag]; approval: requires plan, preview, approval, and execute; risk: permanently removes a tag, including from every call/text interaction it has been applied to; irreversible, approval recommended; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete user apply - Plan and execute the delete user reverse-ETL action [intent=reverse_etl availability=implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: permanently removes a user's access to the account; requires an administrator-scoped API key; irreversible, approval required; flags: --id (required)
  - disable integration apply - Plan and execute the disable integration reverse-ETL action [intent=reverse_etl availability=implemented write=disable_integration]; approval: requires plan, preview, approval, and execute; risk: disables (the docs' own term; not a hard delete) an integration; stops any external data flow it previously drove; approval recommended; flags: --id (required)
  - form submissions list - Run the form submissions ETL stream [intent=etl availability=implemented stream=form_submissions]
  - integration filters list - Run the integration filters ETL stream [intent=etl availability=implemented stream=integration_filters]
  - integrations list - Run the integrations ETL stream [intent=etl availability=implemented stream=integrations]
  - lead timeline list - Run the lead timeline ETL stream [intent=etl availability=implemented stream=lead_timeline]
  - leads list - Run the leads ETL stream [intent=etl availability=implemented stream=leads]
  - message flows list - Run the message flows ETL stream [intent=etl availability=implemented stream=message_flows]
  - notifications list - Run the notifications ETL stream [intent=etl availability=implemented stream=notifications]
  - page views list - Run the page views ETL stream [intent=etl availability=implemented stream=page_views]
  - send text message apply - Plan and execute the send text message reverse-ETL action [intent=reverse_etl availability=implemented write=send_text_message]; approval: requires plan, preview, approval, and execute; risk: sends a real SMS/MMS text message to a customer's phone (subject to 10DLC business-registration compliance rules); a real-world side effect outside the CallRail account itself, approval required. Direct file-upload MMS (multipart media_file) is out of scope — see api_surface.json/docs.md; the media_url variant covers publicly-hosted-image MMS instead.; flags: --company_id (required), --content (required), --customer_phone_number (required), --tracking_number (required)
  - sms threads list - Run the sms threads ETL stream [intent=etl availability=implemented stream=sms_threads]
  - tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]
  - text messages list - Run the text messages ETL stream [intent=etl availability=implemented stream=text_messages]
  - trackers list - Run the trackers ETL stream [intent=etl availability=implemented stream=trackers]
  - update call apply - Plan and execute the update call reverse-ETL action [intent=reverse_etl availability=implemented write=update_call]; approval: requires plan, preview, approval, and execute; risk: applies tags/notes/lead-status/value/customer-name metadata to an existing call record; low-risk external mutation, no approval required; flags: --id (required)
  - update company apply - Plan and execute the update company reverse-ETL action [intent=reverse_etl availability=implemented write=update_company]; approval: requires plan, preview, approval, and execute; risk: updates company configuration; setting status to disabled deactivates all of the company's tracking numbers and its dynamic-number-insertion script — approval recommended for status changes; flags: --id (required)
  - update integration apply - Plan and execute the update integration reverse-ETL action [intent=reverse_etl availability=implemented write=update_integration]; approval: requires plan, preview, approval, and execute; risk: updates an integration's active/disabled state or its webhook/cookie-capture configuration; approval recommended; flags: --id (required)
  - update integration filter apply - Plan and execute the update integration filter reverse-ETL action [intent=reverse_etl availability=implemented write=update_integration_filter]; approval: requires plan, preview, approval, and execute; risk: updates an integration filter's trigger criteria; low-risk external mutation, no approval required; flags: --id (required)
  - update message flow apply - Plan and execute the update message flow reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_message_flow]; approval: requires plan, preview, approval, and execute; risk: replaces an existing message flow's step graph; the docs' own endpoint takes no {message_flow_id} path segment, identifying the flow purely via the body's id field; low-risk external mutation, no approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update notification apply - Plan and execute the update notification reverse-ETL action [intent=reverse_etl availability=implemented write=update_notification]; approval: requires plan, preview, approval, and execute; risk: updates an existing notification's scope/channel settings; low-risk external mutation, no approval required; flags: --id (required)
  - update sms thread apply - Plan and execute the update sms thread reverse-ETL action [intent=reverse_etl availability=implemented write=update_sms_thread]; approval: requires plan, preview, approval, and execute; risk: applies notes/value/tags/lead-qualification metadata to an existing SMS thread; low-risk external mutation, no approval required; flags: --id (required)
  - update tag apply - Plan and execute the update tag reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_tag]; approval: requires plan, preview, approval, and execute; risk: renames/recolors/disables a tag; renaming changes the tag everywhere it is currently assigned; low-risk external mutation, no approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update tracker apply - Plan and execute the update tracker reverse-ETL action [intent=reverse_etl availability=implemented write=update_tracker]; approval: requires plan, preview, approval, and execute; risk: reconfigures an existing (already-provisioned) session or source tracker's call flow, whisper message, SMS setting, or source rules; does not provision/deprovision a phone number itself, unlike create/disable; low-risk external mutation, no approval required; flags: --id (required)
  - update user apply - Plan and execute the update user reverse-ETL action [intent=reverse_etl availability=implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: updates a user's profile/role/company access; name/email changes are restricted to the API key's own owning user by CallRail; approval recommended for role/company changes; flags: --id (required)
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]

## Commands

### Inspect as a manual

```bash
pm connectors inspect callrail
```

### Inspect as structured JSON

```bash
pm connectors inspect callrail --json
```

## Agent Rules

- Run pm connectors inspect callrail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
