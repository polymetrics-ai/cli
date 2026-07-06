---
name: pm-callrail
description: CallRail connector knowledge and safe action guide.
---

# pm-callrail

## Purpose

Reads and writes CallRail call tracking data (calls, companies, users, tags, trackers, form submissions, text messages, notifications, integrations, and more) through the CallRail v3 REST API.

## Icon

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
  - fields: answered(), business_phone_number(), company_id(), customer_city(), customer_country(), customer_name(), customer_phone_number(), customer_state(), direction(), duration(), id(), recording(), start_time(), tracking_phone_number(), voicemail()
- companies:
  - primary key: id
  - cursor: created_at
  - fields: callscore_enabled(), created_at(), disabled_at(), dni_active(), id(), name(), status(), time_zone()
- users:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), email(), first_name(), id(), last_name(), name(), role()
- text_messages:
  - primary key: id
  - cursor: last_message_at
  - fields: company_id(), customer_name(), customer_phone_number(), id(), initial_tracker_id(), last_message_at(), state(), tracking_phone_number()
- accounts:
  - primary key: id
  - fields: hipaa_account(), id(), name(), outbound_recording_enabled()
- tags:
  - primary key: id
  - cursor: created_at
  - fields: background_color(), color(), company_id(), created_at(), id(), name(), status(), tag_level()
- trackers:
  - primary key: id
  - cursor: created_at
  - fields: company_id(), company_name(), created_at(), destination_number(), disabled_at(), id(), name(), sms_enabled(), sms_supported(), status(), tracking_numbers(), type(), whisper_message()
- form_submissions:
  - primary key: id
  - cursor: submitted_at
  - fields: campaign(), company_id(), customer_email(), customer_name(), customer_phone_number(), first_form(), form_url(), id(), keywords(), landing_page_url(), medium(), person_id(), referrer(), referring_url(), source(), submitted_at()
- integrations:
  - primary key: id
  - fields: config(), id(), state(), type()
- integration_filters:
  - primary key: id
  - fields: call_type(), company_id(), id(), integration_id(), integration_type(), lead_status(), max_duration(), min_duration(), tracker_ids()
- notifications:
  - primary key: id
  - fields: alert_type(), call_enabled(), company_id(), company_name(), id(), name(), send_desktop(), send_email(), send_push(), sms_enabled(), tracker_id(), tracker_name(), user_id()
- caller_ids:
  - primary key: id
  - cursor: created_at
  - fields: company_id(), created_at(), id(), name(), phone_number(), validation_code(), verified()
- sms_threads:
  - primary key: id
  - fields: company_id(), company_time_zone(), current_tracker_id(), current_tracking_number(), customer_name(), customer_phone_number(), id(), initial_tracker_id(), initial_tracking_number(), lead_qualification(), notes(), state(), tags(), value()
- message_flows:
  - primary key: id
  - fields: id(), initial_step_id(), name(), steps(), tracker_ids(), updated_at()
- leads:
  - primary key: id
  - cursor: created_at
  - fields: company_id(), company_name(), created_at(), email(), id(), name(), phone()
- page_views:
  - primary key: call_id, created_at
  - cursor: created_at
  - fields: call_id(), created_at(), page_url(), referrer_url()
- lead_timeline:
  - primary key: lead_id
  - fields: campaign(), customer_name(), customer_phone_number(), first_touch(), last_touch(), lead_creation(), lead_id(), lead_qualification(), medium(), source(), tags(), total_interactions(), transcript(), voice_assist()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_tag:
  - endpoint: POST /a/{{ config.account_id }}/tags.json
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
  - risk: creates a new company (a billable tracking entity) within the account; approval recommended
- update_company:
  - endpoint: PUT /a/{{ config.account_id }}/companies/{{ record.id }}.json
  - required fields: id
  - risk: updates company configuration; setting status to disabled deactivates all of the company's tracking numbers and its dynamic-number-insertion script — approval recommended for status changes
- create_user:
  - endpoint: POST /a/{{ config.account_id }}/users.json
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
  - risk: places a real outbound phone call connecting a business and a customer number (US/Canada only); a real-world side effect outside the CallRail account itself, approval required
- send_text_message:
  - endpoint: POST /a/{{ config.account_id }}/text-messages.json
  - risk: sends a real SMS/MMS text message to a customer's phone (subject to 10DLC business-registration compliance rules); a real-world side effect outside the CallRail account itself, approval required. Direct file-upload MMS (multipart media_file) is out of scope — see api_surface.json/docs.md; the media_url variant covers publicly-hosted-image MMS instead.
- create_integration:
  - endpoint: POST /a/{{ config.account_id }}/integrations.json
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
  - risk: creates a new automated SMS message flow (a step-graph of tag/response actions) for a company; low-risk external mutation, no approval required
- update_message_flow:
  - endpoint: PUT /a/{{ config.account_id }}/message-flows.json
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
