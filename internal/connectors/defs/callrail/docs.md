# Overview

Reads and writes CallRail call tracking data (calls, companies, users, tags, trackers, form
submissions, text messages, notifications, integrations, and more) through the CallRail v3 REST API.

Readable streams: `calls`, `companies`, `users`, `text_messages`, `accounts`, `tags`, `trackers`,
`form_submissions`, `integrations`, `integration_filters`, `notifications`, `caller_ids`,
`sms_threads`, `message_flows`, `leads`, `page_views`, `lead_timeline`.

Write actions: `create_tag`, `update_tag`, `delete_tag`, `create_company`, `update_company`,
`create_user`, `update_user`, `delete_user`, `update_call`, `create_outbound_call`,
`send_text_message`, `create_integration`, `update_integration`, `disable_integration`,
`create_integration_filter`, `update_integration_filter`, `delete_integration_filter`,
`create_notification`, `update_notification`, `delete_notification`, `create_caller_id`,
`delete_caller_id`, `update_sms_thread`, `update_tracker`, `create_message_flow`,
`update_message_flow`, `delete_message_flow`.

Service API documentation: https://apidocs.callrail.com/.

## Auth setup

Connection fields:

- `account_id` (required, string); CallRail account id; substituted into every request path (a/{{
  config.account_id }}/...).
- `api_key` (required, secret, string); CallRail API key. Sent as Authorization: Token
  token="<api_key>". Never logged.
- `base_url` (optional, string); default `https://api.callrail.com/v3`; format `uri`; CallRail API
  base URL override for tests or proxies.
- `company_id` (optional, string); Optional CallRail company id filter, applied to the integrations,
  integration_filters, caller_ids, message_flows, and leads streams' company_id query parameter.
  Omitted from each request entirely when unset (stream.Query's omit_when_absent dialect).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only records at or after
  this date are read on a fresh sync.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.callrail.com/v3`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token token=` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/a/{{ config.account_id }}/companies.json` with query `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `page_views`, `lead_timeline`; page_number: `calls`, `companies`,
`users`, `text_messages`, `accounts`, `tags`, `trackers`, `form_submissions`, `integrations`,
`integration_filters`, `notifications`, `caller_ids`, `sms_threads`, `message_flows`, `leads`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `calls`: GET `/a/{{ config.account_id }}/calls.json` - records path `calls`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100;
  incremental cursor `start_time`; sent as `start_date`; formatted as YYYY-MM-DD date; initial lower
  bound from `start_date`.
- `companies`: GET `/a/{{ config.account_id }}/companies.json` - records path `companies`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; incremental cursor `created_at`; sent as `start_date`; formatted as YYYY-MM-DD date; initial
  lower bound from `start_date`.
- `users`: GET `/a/{{ config.account_id }}/users.json` - records path `users`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100;
  incremental cursor `created_at`; sent as `start_date`; formatted as YYYY-MM-DD date; initial lower
  bound from `start_date`.
- `text_messages`: GET `/a/{{ config.account_id }}/text-messages.json` - records path
  `text_messages`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; incremental cursor `last_message_at`; sent as `start_date`; formatted as
  YYYY-MM-DD date; initial lower bound from `start_date`.
- `accounts`: GET `/a.json` - records path `accounts`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `tags`: GET `/a/{{ config.account_id }}/tags.json` - records path `tags`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor
  `created_at`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `trackers`: GET `/a/{{ config.account_id }}/trackers.json` - records path `trackers`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100;
  incremental cursor `created_at`; formatted as `rfc3339`; initial lower bound from `start_date`;
  computed output fields `company_id`, `company_name`.
- `form_submissions`: GET `/a/{{ config.account_id }}/form_submissions.json` - records path
  `form_submissions`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; incremental cursor `submitted_at`; sent as `start_date`; formatted as
  YYYY-MM-DD date; initial lower bound from `start_date`.
- `integrations`: GET `/a/{{ config.account_id }}/integrations.json` - records path `integrations`;
  query `company_id` from template `{{ config.company_id }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `integration_filters`: GET `/a/{{ config.account_id }}/integration_triggers.json` - records path
  `integration_criteria`; query `company_id` from template `{{ config.company_id }}`, omitted when
  absent; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; computed output fields `integration_type`.
- `notifications`: GET `/a/{{ config.account_id }}/notifications.json` - records path
  `notifications`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100.
- `caller_ids`: GET `/a/{{ config.account_id }}/caller_ids.json` - records path `caller_ids`; query
  `company_id` from template `{{ config.company_id }}`, omitted when absent; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `sms_threads`: GET `/a/{{ config.account_id }}/sms-threads.json` - records path `sms_threads`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100.
- `message_flows`: GET `/a/{{ config.account_id }}/message-flows.json` - records path
  `message-flows`; query `company_id` from template `{{ config.company_id }}`, omitted when absent;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100.
- `leads`: GET `/a/{{ config.account_id }}/leads.json` - records path `leads`; query `company_id`
  from template `{{ config.company_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor
  `created_at`; formatted as `rfc3339`; initial lower bound from `start_date`.
- `page_views`: GET `/a/{{ config.account_id }}/calls/{{ fanout.id }}/page_views.json` - records
  path `page_views`; fan-out; ids from request `/a/{{ config.account_id }}/calls.json`; id-list
  records path `calls`; id field `id`; id inserted into the request path; stamps `call_id`.
- `lead_timeline`: GET `/a/{{ config.account_id }}/leads/{{ fanout.id }}/timeline.json` - records
  path `lead`; fan-out; ids from request `/a/{{ config.account_id }}/leads.json`; id-list records
  path `leads`; id field `id`; id inserted into the request path; stamps `lead_id`.

## Write actions & risks

Overall write risk: external mutation of CallRail account configuration (tags, companies, users,
notifications, outbound caller ids, message flows, integration filters), call/lead metadata (call
tags, lead status, value), and outbound communications (placing outbound calls, sending SMS).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_tag`: POST `/a/{{ config.account_id }}/tags.json` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `color`, `company_id`, `name`, `tag_level`; risk:
  creates a new call/text tag definition visible account- or company-wide; low-risk external
  mutation, no approval required.
- `update_tag`: PUT `/a/{{ config.account_id }}/tags/{{ record.id }}.json` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `color`, `disabled`,
  `id`, `name`; risk: renames/recolors/disables a tag; renaming changes the tag everywhere it is
  currently assigned; low-risk external mutation, no approval required.
- `delete_tag`: DELETE `/a/{{ config.account_id }}/tags/{{ record.id }}.json` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently removes a tag, including from every
  call/text interaction it has been applied to; irreversible, approval recommended.
- `create_company`: POST `/a/{{ config.account_id }}/companies.json` - kind `create`; body type
  `json`; required record fields `name`; accepted fields `name`, `time_zone`; risk: creates a new
  company (a billable tracking entity) within the account; approval recommended.
- `update_company`: PUT `/a/{{ config.account_id }}/companies/{{ record.id }}.json` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields
  `callscore_enabled`, `callscribe_enabled`, `id`, `name`, `status`, `swap_landing_override`,
  `swap_ppc_override`, `time_zone`; risk: updates company configuration; setting status to disabled
  deactivates all of the company's tracking numbers and its dynamic-number-insertion script -
  approval recommended for status changes.
- `create_user`: POST `/a/{{ config.account_id }}/users.json` - kind `create`; body type `json`;
  required record fields `first_name`, `last_name`, `email`, `role`; accepted fields `companies`,
  `email`, `first_name`, `last_name`, `role`; risk: creates a new CallRail user and emails them a
  password-setup prompt; requires an administrator-scoped API key; approval recommended.
- `update_user`: PUT `/a/{{ config.account_id }}/users/{{ record.id }}.json` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `companies`, `email`,
  `first_name`, `id`, `last_name`, `role`; risk: updates a user's profile/role/company access;
  name/email changes are restricted to the API key's own owning user by CallRail; approval
  recommended for role/company changes.
- `delete_user`: DELETE `/a/{{ config.account_id }}/users/{{ record.id }}.json` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; risk: permanently removes a user's access to the
  account; requires an administrator-scoped API key; irreversible, approval required.
- `update_call`: PUT `/a/{{ config.account_id }}/calls/{{ record.id }}.json` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `append_tags`,
  `customer_name`, `id`, `lead_status`, `note`, `tags`, `value`; risk: applies
  tags/notes/lead-status/value/customer-name metadata to an existing call record; low-risk external
  mutation, no approval required.
- `create_outbound_call`: POST `/a/{{ config.account_id }}/calls.json` - kind `create`; body type
  `json`; required record fields `caller_id`, `business_phone_number`, `customer_phone_number`;
  accepted fields `business_phone_number`, `caller_id`, `customer_phone_number`,
  `outbound_greeting_recording_url`, `outbound_greeting_text`, `recording_enabled`; risk: places a
  real outbound phone call connecting a business and a customer number (US/Canada only); a
  real-world side effect outside the CallRail account itself, approval required.
- `send_text_message`: POST `/a/{{ config.account_id }}/text-messages.json` - kind `create`; body
  type `json`; required record fields `company_id`, `customer_phone_number`, `tracking_number`,
  `content`; accepted fields `company_id`, `content`, `customer_phone_number`, `media_url`,
  `tracking_number`; risk: sends a real SMS/MMS text message to a customer's phone (subject to 10DLC
  business-registration compliance rules); a real-world side effect outside the CallRail account
  itself, approval required. Direct file-upload MMS (multipart media_file) is out of scope - see
  api_surface.json/docs.md; the media_url variant covers publicly-hosted-image MMS instead.
- `create_integration`: POST `/a/{{ config.account_id }}/integrations.json` - kind `create`; body
  type `json`; required record fields `type`, `company_id`; accepted fields `company_id`, `config`,
  `type`; risk: creates and activates a Webhooks or Custom-cookie-capture integration for a company
  (the only 2 integration types the API can create); approval recommended since Webhooks
  integrations push call data to an external URL.
- `update_integration`: PUT `/a/{{ config.account_id }}/integrations/{{ record.id }}.json` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `config`, `id`, `state`; risk: updates an integration's active/disabled state or its
  webhook/cookie-capture configuration; approval recommended.
- `disable_integration`: DELETE `/a/{{ config.account_id }}/integrations/{{ record.id }}.json` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: disables (the docs' own term; not
  a hard delete) an integration; stops any external data flow it previously drove; approval
  recommended.
- `create_integration_filter`: POST `/a/{{ config.account_id }}/integration_triggers.json` - kind
  `create`; body type `json`; required record fields `company_id`, `integration_id`; accepted fields
  `call_type`, `company_id`, `integration_id`, `lead_status`, `max_duration`, `min_duration`,
  `tracker_ids`; risk: adds a filter narrowing which calls trigger an existing integration; low-risk
  external mutation, no approval required.
- `update_integration_filter`: PUT `/a/{{ config.account_id }}/integration_triggers/{{ record.id
  }}.json` - kind `update`; body type `json`; path fields `id`; required record fields `id`;
  accepted fields `call_type`, `id`, `lead_status`, `max_duration`, `min_duration`, `tracker_ids`;
  risk: updates an integration filter's trigger criteria; low-risk external mutation, no approval
  required.
- `delete_integration_filter`: DELETE `/a/{{ config.account_id }}/integration_triggers/{{ record.id
  }}.json` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; risk: removes a filter;
  the parent integration keeps firing for every call, unfiltered, once this is removed; low-risk, no
  approval required.
- `create_notification`: POST `/a/{{ config.account_id }}/notifications.json` - kind `create`; body
  type `json`; accepted fields `alert_type`, `call_enabled`, `company_id`, `email`, `send_desktop`,
  `send_email`, `send_push`, `sms_enabled`, `tracker_id`, `user_id`; risk: creates a call/text alert
  subscription for a user; low-risk external mutation, no approval required.
- `update_notification`: PUT `/a/{{ config.account_id }}/notifications/{{ record.id }}.json` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `alert_type`, `call_enabled`, `company_id`, `id`, `sms_enabled`, `tracker_id`; risk: updates an
  existing notification's scope/channel settings; low-risk external mutation, no approval required.
- `delete_notification`: DELETE `/a/{{ config.account_id }}/notifications/{{ record.id }}.json` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: permanently removes a
  notification subscription (restricted to notifications managed by the current user); irreversible,
  low-risk, no approval required.
- `create_caller_id`: POST `/a/{{ config.account_id }}/caller_ids.json` - kind `create`; body type
  `json`; required record fields `company_id`, `phone_number`, `name`; accepted fields `company_id`,
  `name`, `phone_number`; risk: registers an outbound caller-id number and immediately triggers a
  real verification phone call to it; a real-world side effect, approval required.
- `delete_caller_id`: DELETE `/a/{{ config.account_id }}/caller_ids/{{ record.id }}.json` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; risk: removes an outbound caller id from the
  company; irreversible, low-risk, no approval required.
- `update_sms_thread`: PUT `/a/{{ config.account_id }}/sms-threads/{{ record.id }}.json` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `append_tags`, `id`, `lead_qualification`, `notes`, `tags`, `value`; risk: applies
  notes/value/tags/lead-qualification metadata to an existing SMS thread; low-risk external
  mutation, no approval required.
- `update_tracker`: PUT `/a/{{ config.account_id }}/trackers/{{ record.id }}.json` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `call_flow`,
  `campaign_name`, `id`, `name`, `replace_tracking_number`, `sms_enabled`, `source`, `swap_targets`,
  `whisper_message`; risk: reconfigures an existing (already-provisioned) session or source
  tracker's call flow, whisper message, SMS setting, or source rules; does not provision/deprovision
  a phone number itself, unlike create/disable; low-risk external mutation, no approval required.
- `create_message_flow`: POST `/a/{{ config.account_id }}/message-flows.json` - kind `create`; body
  type `json`; required record fields `company_id`, `name`, `initial_step_id`, `steps`; accepted
  fields `company_id`, `initial_step_id`, `name`, `steps`; risk: creates a new automated SMS message
  flow (a step-graph of tag/response actions) for a company; low-risk external mutation, no approval
  required.
- `update_message_flow`: PUT `/a/{{ config.account_id }}/message-flows.json` - kind `update`; body
  type `json`; required record fields `id`, `initial_step_id`, `steps`; accepted fields `id`,
  `initial_step_id`, `steps`; risk: replaces an existing message flow's step graph; the docs' own
  endpoint takes no {message_flow_id} path segment, identifying the flow purely via the body's id
  field; low-risk external mutation, no approval required.
- `delete_message_flow`: DELETE `/a/{{ config.account_id }}/message-flows/{{ record.id }}.json` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; risk: permanently removes a message
  flow; any tracker still referencing it stops running the automated SMS steps; irreversible,
  approval recommended.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 17 stream-backed endpoint group(s), 27 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=1, destructive_admin=1, duplicate_of=13, non_data_endpoint=5, out_of_scope=7,
  requires_elevated_scope=4.
