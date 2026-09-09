# pm connectors inspect callrail

```text
NAME
  pm connectors inspect callrail - CallRail connector manual

SYNOPSIS
  pm connectors inspect callrail
  pm connectors inspect callrail --json
  pm credentials add <name> --connector callrail [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes CallRail call tracking data (calls, companies, users, tags, trackers, form submissions, text messages, notifications, integrations, and more) through the CallRail v3 REST API.

ICON
  id: callrail
  asset: icons/callrail.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://apidocs.callrail.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id (required)
  base_url
  company_id
  start_date
  api_key (secret) (required)

ETL STREAMS
  calls:
    primary key: id
    cursor: start_time
    fields: answered(boolean), business_phone_number(string), company_id(string), customer_city(string), customer_country(string), customer_name(string), customer_phone_number(string), customer_state(string), direction(string), duration(integer), id(string), recording(string), start_time(string), tracking_phone_number(string), voicemail(boolean)
  companies:
    primary key: id
    cursor: created_at
    fields: callscore_enabled(boolean), created_at(string), disabled_at(string), dni_active(boolean), id(string), name(string), status(string), time_zone(string)
  users:
    primary key: id
    cursor: created_at
    fields: created_at(string), email(string), first_name(string), id(string), last_name(string), name(string), role(string)
  text_messages:
    primary key: id
    cursor: last_message_at
    fields: company_id(string), customer_name(string), customer_phone_number(string), id(string), initial_tracker_id(string), last_message_at(string), state(string), tracking_phone_number(string)
  accounts:
    primary key: id
    fields: hipaa_account(boolean), id(string), name(string), outbound_recording_enabled(boolean)
  tags:
    primary key: id
    cursor: created_at
    fields: background_color(string), color(string), company_id(string), created_at(string), id(string), name(string), status(string), tag_level(string)
  trackers:
    primary key: id
    cursor: created_at
    fields: company_id(string), company_name(string), created_at(string), destination_number(string), disabled_at(string), id(string), name(string), sms_enabled(boolean), sms_supported(boolean), status(string), tracking_numbers(array), type(string), whisper_message(string)
  form_submissions:
    primary key: id
    cursor: submitted_at
    fields: campaign(string), company_id(string), customer_email(string), customer_name(string), customer_phone_number(string), first_form(boolean), form_url(string), id(string), keywords(string), landing_page_url(string), medium(string), person_id(string), referrer(string), referring_url(string), source(string), submitted_at(string)
  integrations:
    primary key: id
    fields: config(object), id(integer), state(string), type(string)
  integration_filters:
    primary key: id
    fields: call_type(string), company_id(string), id(integer), integration_id(integer), integration_type(string), lead_status(string), max_duration(integer), min_duration(integer), tracker_ids(array)
  notifications:
    primary key: id
    fields: alert_type(string), call_enabled(boolean), company_id(string), company_name(string), id(integer), name(string), send_desktop(boolean), send_email(boolean), send_push(boolean), sms_enabled(boolean), tracker_id(string), tracker_name(string), user_id(string)
  caller_ids:
    primary key: id
    cursor: created_at
    fields: company_id(string), created_at(string), id(integer), name(string), phone_number(string), validation_code(string), verified(boolean)
  sms_threads:
    primary key: id
    fields: company_id(string), company_time_zone(string), current_tracker_id(string), current_tracking_number(string), customer_name(string), customer_phone_number(string), id(string), initial_tracker_id(string), initial_tracking_number(string), lead_qualification(string), notes(string), state(string), tags(array), value(number)
  message_flows:
    primary key: id
    fields: id(string), initial_step_id(string), name(string), steps(object), tracker_ids(array), updated_at(string)
  leads:
    primary key: id
    cursor: created_at
    fields: company_id(string), company_name(string), created_at(string), email(string), id(string), name(string), phone(string)
  page_views:
    primary key: call_id, created_at
    cursor: created_at
    fields: call_id(string), created_at(string), page_url(string), referrer_url(string)
  lead_timeline:
    primary key: lead_id
    fields: campaign(string), customer_name(string), customer_phone_number(string), first_touch(object), last_touch(object), lead_creation(object), lead_id(string), lead_qualification(object), medium(string), source(string), tags(array), total_interactions(integer), transcript(boolean), voice_assist(boolean)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_tag:
    endpoint: POST /a/{{ config.account_id }}/tags.json
    required fields: name
    risk: creates a new call/text tag definition visible account- or company-wide; low-risk external mutation, no approval required
  update_tag:
    endpoint: PUT /a/{{ config.account_id }}/tags/{{ record.id }}.json
    required fields: id
    risk: renames/recolors/disables a tag; renaming changes the tag everywhere it is currently assigned; low-risk external mutation, no approval required
  delete_tag:
    endpoint: DELETE /a/{{ config.account_id }}/tags/{{ record.id }}.json
    required fields: id
    risk: permanently removes a tag, including from every call/text interaction it has been applied to; irreversible, approval recommended
  create_company:
    endpoint: POST /a/{{ config.account_id }}/companies.json
    required fields: name
    risk: creates a new company (a billable tracking entity) within the account; approval recommended
  update_company:
    endpoint: PUT /a/{{ config.account_id }}/companies/{{ record.id }}.json
    required fields: id
    risk: updates company configuration; setting status to disabled deactivates all of the company's tracking numbers and its dynamic-number-insertion script — approval recommended for status changes
  create_user:
    endpoint: POST /a/{{ config.account_id }}/users.json
    required fields: first_name, last_name, email, role
    risk: creates a new CallRail user and emails them a password-setup prompt; requires an administrator-scoped API key; approval recommended
  update_user:
    endpoint: PUT /a/{{ config.account_id }}/users/{{ record.id }}.json
    required fields: id
    risk: updates a user's profile/role/company access; name/email changes are restricted to the API key's own owning user by CallRail; approval recommended for role/company changes
  delete_user:
    endpoint: DELETE /a/{{ config.account_id }}/users/{{ record.id }}.json
    required fields: id
    risk: permanently removes a user's access to the account; requires an administrator-scoped API key; irreversible, approval required
  update_call:
    endpoint: PUT /a/{{ config.account_id }}/calls/{{ record.id }}.json
    required fields: id
    risk: applies tags/notes/lead-status/value/customer-name metadata to an existing call record; low-risk external mutation, no approval required
  create_outbound_call:
    endpoint: POST /a/{{ config.account_id }}/calls.json
    required fields: caller_id, business_phone_number, customer_phone_number
    risk: places a real outbound phone call connecting a business and a customer number (US/Canada only); a real-world side effect outside the CallRail account itself, approval required
  send_text_message:
    endpoint: POST /a/{{ config.account_id }}/text-messages.json
    required fields: company_id, customer_phone_number, tracking_number, content
    risk: sends a real SMS/MMS text message to a customer's phone (subject to 10DLC business-registration compliance rules); a real-world side effect outside the CallRail account itself, approval required. Direct file-upload MMS (multipart media_file) is out of scope — see execution bundle/docs.md; the media_url variant covers publicly-hosted-image MMS instead.
  create_integration:
    endpoint: POST /a/{{ config.account_id }}/integrations.json
    required fields: type, company_id
    risk: creates and activates a Webhooks or Custom-cookie-capture integration for a company (the only 2 integration types the API can create); approval recommended since Webhooks integrations push call data to an external URL
  update_integration:
    endpoint: PUT /a/{{ config.account_id }}/integrations/{{ record.id }}.json
    required fields: id
    risk: updates an integration's active/disabled state or its webhook/cookie-capture configuration; approval recommended
  disable_integration:
    endpoint: DELETE /a/{{ config.account_id }}/integrations/{{ record.id }}.json
    required fields: id
    risk: disables (the docs' own term; not a hard delete) an integration; stops any external data flow it previously drove; approval recommended
  create_integration_filter:
    endpoint: POST /a/{{ config.account_id }}/integration_triggers.json
    required fields: company_id, integration_id
    risk: adds a filter narrowing which calls trigger an existing integration; low-risk external mutation, no approval required
  update_integration_filter:
    endpoint: PUT /a/{{ config.account_id }}/integration_triggers/{{ record.id }}.json
    required fields: id
    risk: updates an integration filter's trigger criteria; low-risk external mutation, no approval required
  delete_integration_filter:
    endpoint: DELETE /a/{{ config.account_id }}/integration_triggers/{{ record.id }}.json
    required fields: id
    risk: removes a filter; the parent integration keeps firing for every call, unfiltered, once this is removed; low-risk, no approval required
  create_notification:
    endpoint: POST /a/{{ config.account_id }}/notifications.json
    risk: creates a call/text alert subscription for a user; low-risk external mutation, no approval required
  update_notification:
    endpoint: PUT /a/{{ config.account_id }}/notifications/{{ record.id }}.json
    required fields: id
    risk: updates an existing notification's scope/channel settings; low-risk external mutation, no approval required
  delete_notification:
    endpoint: DELETE /a/{{ config.account_id }}/notifications/{{ record.id }}.json
    required fields: id
    risk: permanently removes a notification subscription (restricted to notifications managed by the current user); irreversible, low-risk, no approval required
  create_caller_id:
    endpoint: POST /a/{{ config.account_id }}/caller_ids.json
    required fields: company_id, phone_number, name
    risk: registers an outbound caller-id number and immediately triggers a real verification phone call to it; a real-world side effect, approval required
  delete_caller_id:
    endpoint: DELETE /a/{{ config.account_id }}/caller_ids/{{ record.id }}.json
    required fields: id
    risk: removes an outbound caller id from the company; irreversible, low-risk, no approval required
  update_sms_thread:
    endpoint: PUT /a/{{ config.account_id }}/sms-threads/{{ record.id }}.json
    required fields: id
    risk: applies notes/value/tags/lead-qualification metadata to an existing SMS thread; low-risk external mutation, no approval required
  update_tracker:
    endpoint: PUT /a/{{ config.account_id }}/trackers/{{ record.id }}.json
    required fields: id
    risk: reconfigures an existing (already-provisioned) session or source tracker's call flow, whisper message, SMS setting, or source rules; does not provision/deprovision a phone number itself, unlike create/disable; low-risk external mutation, no approval required
  create_message_flow:
    endpoint: POST /a/{{ config.account_id }}/message-flows.json
    required fields: company_id, name, initial_step_id, steps
    risk: creates a new automated SMS message flow (a step-graph of tag/response actions) for a company; low-risk external mutation, no approval required
  update_message_flow:
    endpoint: PUT /a/{{ config.account_id }}/message-flows.json
    required fields: id, initial_step_id, steps
    risk: replaces an existing message flow's step graph; the docs' own endpoint takes no {message_flow_id} path segment, identifying the flow purely via the body's id field; low-risk external mutation, no approval required
  delete_message_flow:
    endpoint: DELETE /a/{{ config.account_id }}/message-flows/{{ record.id }}.json
    required fields: id
    risk: permanently removes a message flow; any tracker still referencing it stops running the automated SMS steps; irreversible, approval recommended

SECURITY
  read risk: external CallRail API read of call tracking analytics, contact, and configuration data
  write risk: external mutation of CallRail account configuration (tags, companies, users, notifications, outbound caller ids, message flows, integration filters), call/lead metadata (call tags, lead status, value), and outbound communications (placing outbound calls, sending SMS)
  approval: required for outbound-communication and account-configuration writes (placing calls, sending texts, disabling companies, deleting users/caller-ids); tag/metadata-only writes are lower risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect callrail

  # Inspect as structured JSON
  pm connectors inspect callrail --json

AGENT WORKFLOW
  - Run pm connectors inspect callrail before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
