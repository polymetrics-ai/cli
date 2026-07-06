# Overview

Reads and writes Twilio public REST API v2010 resources through data streams and typed write
actions.

Readable streams: `messages`, `calls`, `recordings`, `conferences`, `usage_records`, `account`,
`accounts`, `address`, `address_2`, `application`, `applications`, `authorized_connect_app`,
`authorized_connect_apps`, `available_phone_number_countries`, `available_phone_number_country`,
`available_phone_number_locals`, `available_phone_number_machine_to_machines`,
`available_phone_number_mobiles`, `available_phone_number_nationals`,
`available_phone_number_shared_costs`, `available_phone_number_toll_frees`,
`available_phone_number_voips`, `balance`, `call`, `call_events`, `call_notification`,
`call_notifications`, `call_recording`, `call_recordings`, `conference`, `conference_recording`,
`conference_recordings`, `connect_app`, `connect_apps`, `dependent_phone_numbers`,
`incoming_phone_number`, `incoming_phone_number_assigned_add_on`,
`incoming_phone_number_assigned_add_on_extension`,
`incoming_phone_number_assigned_add_on_extensions`, `incoming_phone_number_assigned_add_ons`,
`incoming_phone_number_locals`, `incoming_phone_number_mobiles`, `incoming_phone_number_toll_frees`,
`incoming_phone_numbers`, `key`, `keys`, `media`, `medias`, `member`, `members`, `message`,
`notification`, `notifications`, `outgoing_caller_id`, `outgoing_caller_ids`, `participant`,
`participants`, `queue`, `queues`, `recording`, `recording_add_on_result`,
`recording_add_on_result_payload`, `recording_add_on_result_payload_data`,
`recording_add_on_result_payloads`, `recording_add_on_results`, `recording_transcription`,
`recording_transcriptions`, `short_code`, `short_codes`, `signing_key`, `signing_keys`,
`sip_auth_calls_credential_list_mapping`, `sip_auth_calls_credential_list_mappings`,
`sip_auth_calls_ip_access_control_list_mapping`, `sip_auth_calls_ip_access_control_list_mappings`,
`sip_auth_registrations_credential_list_mapping`, `sip_auth_registrations_credential_list_mappings`,
`sip_credential`, `sip_credential_list`, `sip_credential_list_mapping`,
`sip_credential_list_mappings`, `sip_credential_lists`, `sip_credentials`, `sip_domain`,
`sip_domains`, `sip_ip_access_control_list`, `sip_ip_access_control_list_mapping`,
`sip_ip_access_control_list_mappings`, `sip_ip_access_control_lists`, `sip_ip_address`,
`sip_ip_address_2`, `transcription`, `transcriptions`, `usage_record_all_times`,
`usage_record_dailies`, `usage_record_last_months`, `usage_record_monthlies`,
`usage_record_this_months`, `usage_record_todays`, `usage_record_yearlies`,
`usage_record_yesterdays`, `usage_trigger`, `usage_triggers`.

Write actions: `create_account`, `create_address`, `delete_address`, `update_address`,
`create_application`, `delete_application`, `update_application`, `create_call`, `create_payments`,
`update_payments`, `create_call_recording`, `delete_call_recording`, `update_call_recording`,
`create_siprec`, `update_siprec`, `create_stream`, `update_stream`, `create_realtime_transcription`,
`update_realtime_transcription`, `create_user_defined_message`,
`create_user_defined_message_subscription`, `delete_user_defined_message_subscription`,
`delete_call`, `update_call`, `create_participant`, `delete_participant`, `update_participant`,
`delete_conference_recording`, `update_conference_recording`, `update_conference`,
`delete_connect_app`, `update_connect_app`, `create_incoming_phone_number`,
`create_incoming_phone_number_assigned_add_on`, `delete_incoming_phone_number_assigned_add_on`,
`delete_incoming_phone_number`, `update_incoming_phone_number`,
`create_incoming_phone_number_local`, `create_incoming_phone_number_mobile`,
`create_incoming_phone_number_toll_free`, `create_key`, `delete_key`, `update_key`,
`create_message`, `create_message_feedback`, `delete_media`, `delete_message`, `update_message`,
`create_outgoing_caller_id_validation_request`, `delete_outgoing_caller_id`,
`update_outgoing_caller_id`, `create_queue`, `update_member`, `delete_queue`, `update_queue`,
`delete_recording_transcription`, `delete_recording_add_on_result_payload`,
`delete_recording_add_on_result`, `delete_recording`, `create_signing_key`, `delete_signing_key`,
`update_signing_key`, `create_sip_credential_list`, `create_sip_credential`,
`delete_sip_credential`, `update_sip_credential`, `delete_sip_credential_list`,
`update_sip_credential_list`, `create_sip_domain`, `create_sip_auth_calls_credential_list_mapping`,
`delete_sip_auth_calls_credential_list_mapping`,
`create_sip_auth_calls_ip_access_control_list_mapping`,
`delete_sip_auth_calls_ip_access_control_list_mapping`,
`create_sip_auth_registrations_credential_list_mapping`,
`delete_sip_auth_registrations_credential_list_mapping`, `create_sip_credential_list_mapping`,
`delete_sip_credential_list_mapping`, `create_sip_ip_access_control_list_mapping`,
`delete_sip_ip_access_control_list_mapping`, `delete_sip_domain`, and 14 more.

Service API documentation: https://www.twilio.com/docs/usage/api.

## Auth setup

Connection fields:

- `account_sid` (required, secret, string); Twilio Account SID. Used as the HTTP Basic auth username
  and to scope every account-relative resource path; never logged.
- `add_on_result_sid` (optional, string); Twilio AddOnResult SID used by docs-derived detail
  streams.
- `address_sid` (optional, string); Twilio Address SID used by docs-derived nested streams.
- `assigned_add_on_sid` (optional, string); Twilio Assigned Add-on SID used by docs-derived nested
  streams.
- `auth_token` (required, secret, string); Twilio Auth Token. Used only as the HTTP Basic auth
  password; never logged.
- `base_url` (optional, string); default `https://api.twilio.com/2010-04-01`; format `uri`; Twilio
  REST API base URL override for tests or proxies.
- `call_sid` (optional, string); Twilio Call SID used by docs-derived nested streams.
- `conference_sid` (optional, string); Twilio Conference SID used by docs-derived nested streams.
- `connect_app_sid` (optional, string); Twilio Connect App SID used by docs-derived detail streams.
- `country_code` (optional, string); ISO country code used by available phone number streams.
- `credential_list_sid` (optional, string); Twilio SIP Credential List SID used by docs-derived
  nested streams.
- `domain_sid` (optional, string); Twilio SIP Domain SID used by docs-derived nested streams.
- `ip_access_control_list_sid` (optional, string); Twilio SIP IP Access Control List SID used by
  docs-derived nested streams.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `message_sid` (optional, string); Twilio Message SID used by docs-derived media and feedback
  streams.
- `mode` (optional, string).
- `page_size` (optional, string); default `50`; Records per page (1-1000, PageSize).
- `payload_sid` (optional, string); Twilio Add-on payload SID used by docs-derived payload-data
  streams.
- `queue_sid` (optional, string); Twilio Queue SID used by docs-derived queue member streams.
- `recording_sid` (optional, string); Twilio Recording SID used by docs-derived transcription
  streams.
- `reference_sid` (optional, string); Twilio recording reference SID used by docs-derived Add-on
  result streams.
- `resource_sid` (optional, string); Twilio resource SID used by assigned Add-on streams.
- `sid` (optional, string); Generic Twilio resource SID for docs-derived detail streams and
  account-level resources.

Secret fields are redacted in logs and write previews: `account_sid`, `auth_token`.

Default configuration values: `base_url=https://api.twilio.com/2010-04-01`, `max_pages=0`,
`page_size=50`.

Authentication behavior:

- HTTP Basic authentication using `secrets.account_sid`, `secrets.auth_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/Accounts/{{ secrets.account_sid }}/Messages.json`.

## Streams notes

Default pagination: single request; no pagination.

Streams derived from the REST API reference use passthrough projection with minimal schemas, so
fields Twilio adds to responses flow through without schema changes.

Pagination by stream: none: `messages`, `calls`, `recordings`, `conferences`, `usage_records`,
`account`, `address_2`, `application`, `authorized_connect_app`, `available_phone_number_country`,
`balance`, `call`, `call_notification`, `call_recording`, `conference`, `conference_recording`,
`connect_app`, `incoming_phone_number`, `incoming_phone_number_assigned_add_on`,
`incoming_phone_number_assigned_add_on_extension`, `key`, `media`, `member`, `message`,
`notification`, `outgoing_caller_id`, `participant`, `queue`, `recording`,
`recording_add_on_result`, `recording_add_on_result_payload`,
`recording_add_on_result_payload_data`, `recording_transcription`, `short_code`, `signing_key`,
`sip_auth_calls_credential_list_mapping`, `sip_auth_calls_ip_access_control_list_mapping`,
`sip_auth_registrations_credential_list_mapping`, `sip_credential`, `sip_credential_list`,
`sip_credential_list_mapping`, `sip_domain`, `sip_ip_access_control_list`,
`sip_ip_access_control_list_mapping`, `sip_ip_address_2`, `transcription`, `usage_trigger`;
page_number: `accounts`, `address`, `applications`, `authorized_connect_apps`,
`available_phone_number_countries`, `available_phone_number_locals`,
`available_phone_number_machine_to_machines`, `available_phone_number_mobiles`,
`available_phone_number_nationals`, `available_phone_number_shared_costs`,
`available_phone_number_toll_frees`, `available_phone_number_voips`, `call_events`,
`call_notifications`, `call_recordings`, `conference_recordings`, `connect_apps`,
`dependent_phone_numbers`, `incoming_phone_number_assigned_add_on_extensions`,
`incoming_phone_number_assigned_add_ons`, `incoming_phone_number_locals`,
`incoming_phone_number_mobiles`, `incoming_phone_number_toll_frees`, `incoming_phone_numbers`,
`keys`, `medias`, `members`, `notifications`, `outgoing_caller_ids`, `participants`, `queues`,
`recording_add_on_result_payloads`, `recording_add_on_results`, `recording_transcriptions`,
`short_codes`, `signing_keys`, `sip_auth_calls_credential_list_mappings`,
`sip_auth_calls_ip_access_control_list_mappings`, `sip_auth_registrations_credential_list_mappings`,
`sip_credential_list_mappings`, `sip_credential_lists`, `sip_credentials`, `sip_domains`,
`sip_ip_access_control_list_mappings`, `sip_ip_access_control_lists`, `sip_ip_address`,
`transcriptions`, `usage_record_all_times`, `usage_record_dailies`, `usage_record_last_months`,
`usage_record_monthlies`, `usage_record_this_months`, `usage_record_todays`,
`usage_record_yearlies`, `usage_record_yesterdays`, `usage_triggers`.

- `messages`: GET `/Accounts/{{ secrets.account_sid }}/Messages.json` - records path `messages`.
- `calls`: GET `/Accounts/{{ secrets.account_sid }}/Calls.json` - records path `calls`.
- `recordings`: GET `/Accounts/{{ secrets.account_sid }}/Recordings.json` - records path
  `recordings`.
- `conferences`: GET `/Accounts/{{ secrets.account_sid }}/Conferences.json` - records path
  `conferences`.
- `usage_records`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Records.json` - records path
  `usage_records`.
- `account`: GET `/Accounts/{{ config.sid }}.json` - single-object response; records path `.`; emits
  passthrough records.
- `accounts`: GET `/Accounts.json` - records path `accounts`; page-number pagination; page parameter
  `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits
  passthrough records.
- `address`: GET `/Accounts/{{ secrets.account_sid }}/Addresses.json` - records path `addresses`;
  page-number pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size
  50; maximum 25 page(s); emits passthrough records.
- `address_2`: GET `/Accounts/{{ secrets.account_sid }}/Addresses/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `application`: GET `/Accounts/{{ secrets.account_sid }}/Applications/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `applications`: GET `/Accounts/{{ secrets.account_sid }}/Applications.json` - records path
  `applications`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts
  at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `authorized_connect_app`: GET `/Accounts/{{ secrets.account_sid }}/AuthorizedConnectApps/{{
  config.connect_app_sid }}.json` - single-object response; records path `.`; emits passthrough
  records.
- `authorized_connect_apps`: GET `/Accounts/{{ secrets.account_sid }}/AuthorizedConnectApps.json` -
  records path `authorized_connect_apps`; page-number pagination; page parameter `Page`; size
  parameter `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `available_phone_number_countries`: GET `/Accounts/{{ secrets.account_sid
  }}/AvailablePhoneNumbers.json` - records path `countries`; page-number pagination; page parameter
  `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits
  passthrough records.
- `available_phone_number_country`: GET `/Accounts/{{ secrets.account_sid
  }}/AvailablePhoneNumbers/{{ config.country_code }}.json` - single-object response; records path
  `.`; emits passthrough records.
- `available_phone_number_locals`: GET `/Accounts/{{ secrets.account_sid }}/AvailablePhoneNumbers/{{
  config.country_code }}/Local.json` - records path `available_phone_numbers`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `available_phone_number_machine_to_machines`: GET `/Accounts/{{ secrets.account_sid
  }}/AvailablePhoneNumbers/{{ config.country_code }}/MachineToMachine.json` - records path
  `available_phone_numbers`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `available_phone_number_mobiles`: GET `/Accounts/{{ secrets.account_sid
  }}/AvailablePhoneNumbers/{{ config.country_code }}/Mobile.json` - records path
  `available_phone_numbers`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `available_phone_number_nationals`: GET `/Accounts/{{ secrets.account_sid
  }}/AvailablePhoneNumbers/{{ config.country_code }}/National.json` - records path
  `available_phone_numbers`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `available_phone_number_shared_costs`: GET `/Accounts/{{ secrets.account_sid
  }}/AvailablePhoneNumbers/{{ config.country_code }}/SharedCost.json` - records path
  `available_phone_numbers`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `available_phone_number_toll_frees`: GET `/Accounts/{{ secrets.account_sid
  }}/AvailablePhoneNumbers/{{ config.country_code }}/TollFree.json` - records path
  `available_phone_numbers`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `available_phone_number_voips`: GET `/Accounts/{{ secrets.account_sid }}/AvailablePhoneNumbers/{{
  config.country_code }}/Voip.json` - records path `available_phone_numbers`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `balance`: GET `/Accounts/{{ secrets.account_sid }}/Balance.json` - single-object response;
  records path `.`; emits passthrough records.
- `call`: GET `/Accounts/{{ secrets.account_sid }}/Calls/{{ config.sid }}.json` - single-object
  response; records path `.`; emits passthrough records.
- `call_events`: GET `/Accounts/{{ secrets.account_sid }}/Calls/{{ config.call_sid }}/Events.json` -
  records path `events`; page-number pagination; page parameter `Page`; size parameter `PageSize`;
  starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `call_notification`: GET `/Accounts/{{ secrets.account_sid }}/Calls/{{ config.call_sid
  }}/Notifications/{{ config.sid }}.json` - single-object response; records path `.`; emits
  passthrough records.
- `call_notifications`: GET `/Accounts/{{ secrets.account_sid }}/Calls/{{ config.call_sid
  }}/Notifications.json` - records path `notifications`; page-number pagination; page parameter
  `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits
  passthrough records.
- `call_recording`: GET `/Accounts/{{ secrets.account_sid }}/Calls/{{ config.call_sid
  }}/Recordings/{{ config.sid }}.json` - single-object response; records path `.`; emits passthrough
  records.
- `call_recordings`: GET `/Accounts/{{ secrets.account_sid }}/Calls/{{ config.call_sid
  }}/Recordings.json` - records path `recordings`; page-number pagination; page parameter `Page`;
  size parameter `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough
  records.
- `conference`: GET `/Accounts/{{ secrets.account_sid }}/Conferences/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `conference_recording`: GET `/Accounts/{{ secrets.account_sid }}/Conferences/{{
  config.conference_sid }}/Recordings/{{ config.sid }}.json` - single-object response; records path
  `.`; emits passthrough records.
- `conference_recordings`: GET `/Accounts/{{ secrets.account_sid }}/Conferences/{{
  config.conference_sid }}/Recordings.json` - records path `recordings`; page-number pagination;
  page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum 25 page(s);
  emits passthrough records.
- `connect_app`: GET `/Accounts/{{ secrets.account_sid }}/ConnectApps/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `connect_apps`: GET `/Accounts/{{ secrets.account_sid }}/ConnectApps.json` - records path
  `connect_apps`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts
  at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `dependent_phone_numbers`: GET `/Accounts/{{ secrets.account_sid }}/Addresses/{{
  config.address_sid }}/DependentPhoneNumbers.json` - records path `dependent_phone_numbers`;
  page-number pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size
  50; maximum 25 page(s); emits passthrough records.
- `incoming_phone_number`: GET `/Accounts/{{ secrets.account_sid }}/IncomingPhoneNumbers/{{
  config.sid }}.json` - single-object response; records path `.`; emits passthrough records.
- `incoming_phone_number_assigned_add_on`: GET `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/{{ config.resource_sid }}/AssignedAddOns/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `incoming_phone_number_assigned_add_on_extension`: GET `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/{{ config.resource_sid }}/AssignedAddOns/{{ config.assigned_add_on_sid
  }}/Extensions/{{ config.sid }}.json` - single-object response; records path `.`; emits passthrough
  records.
- `incoming_phone_number_assigned_add_on_extensions`: GET `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/{{ config.resource_sid }}/AssignedAddOns/{{ config.assigned_add_on_sid
  }}/Extensions.json` - records path `extensions`; page-number pagination; page parameter `Page`;
  size parameter `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough
  records.
- `incoming_phone_number_assigned_add_ons`: GET `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/{{ config.resource_sid }}/AssignedAddOns.json` - records path
  `assigned_add_ons`; page-number pagination; page parameter `Page`; size parameter `PageSize`;
  starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `incoming_phone_number_locals`: GET `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/Local.json` - records path `incoming_phone_numbers`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `incoming_phone_number_mobiles`: GET `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/Mobile.json` - records path `incoming_phone_numbers`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `incoming_phone_number_toll_frees`: GET `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/TollFree.json` - records path `incoming_phone_numbers`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `incoming_phone_numbers`: GET `/Accounts/{{ secrets.account_sid }}/IncomingPhoneNumbers.json` -
  records path `incoming_phone_numbers`; page-number pagination; page parameter `Page`; size
  parameter `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `key`: GET `/Accounts/{{ secrets.account_sid }}/Keys/{{ config.sid }}.json` - single-object
  response; records path `.`; emits passthrough records.
- `keys`: GET `/Accounts/{{ secrets.account_sid }}/Keys.json` - records path `keys`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `media`: GET `/Accounts/{{ secrets.account_sid }}/Messages/{{ config.message_sid }}/Media/{{
  config.sid }}.json` - single-object response; records path `.`; emits passthrough records.
- `medias`: GET `/Accounts/{{ secrets.account_sid }}/Messages/{{ config.message_sid }}/Media.json` -
  records path `media_list`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `member`: GET `/Accounts/{{ secrets.account_sid }}/Queues/{{ config.queue_sid }}/Members/{{
  config.call_sid }}.json` - single-object response; records path `.`; emits passthrough records.
- `members`: GET `/Accounts/{{ secrets.account_sid }}/Queues/{{ config.queue_sid }}/Members.json` -
  records path `queue_members`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `message`: GET `/Accounts/{{ secrets.account_sid }}/Messages/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `notification`: GET `/Accounts/{{ secrets.account_sid }}/Notifications/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `notifications`: GET `/Accounts/{{ secrets.account_sid }}/Notifications.json` - records path
  `notifications`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts
  at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `outgoing_caller_id`: GET `/Accounts/{{ secrets.account_sid }}/OutgoingCallerIds/{{ config.sid
  }}.json` - single-object response; records path `.`; emits passthrough records.
- `outgoing_caller_ids`: GET `/Accounts/{{ secrets.account_sid }}/OutgoingCallerIds.json` - records
  path `outgoing_caller_ids`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `participant`: GET `/Accounts/{{ secrets.account_sid }}/Conferences/{{ config.conference_sid
  }}/Participants/{{ config.call_sid }}.json` - single-object response; records path `.`; emits
  passthrough records.
- `participants`: GET `/Accounts/{{ secrets.account_sid }}/Conferences/{{ config.conference_sid
  }}/Participants.json` - records path `participants`; page-number pagination; page parameter
  `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits
  passthrough records.
- `queue`: GET `/Accounts/{{ secrets.account_sid }}/Queues/{{ config.sid }}.json` - single-object
  response; records path `.`; emits passthrough records.
- `queues`: GET `/Accounts/{{ secrets.account_sid }}/Queues.json` - records path `queues`;
  page-number pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size
  50; maximum 25 page(s); emits passthrough records.
- `recording`: GET `/Accounts/{{ secrets.account_sid }}/Recordings/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `recording_add_on_result`: GET `/Accounts/{{ secrets.account_sid }}/Recordings/{{
  config.reference_sid }}/AddOnResults/{{ config.sid }}.json` - single-object response; records path
  `.`; emits passthrough records.
- `recording_add_on_result_payload`: GET `/Accounts/{{ secrets.account_sid }}/Recordings/{{
  config.reference_sid }}/AddOnResults/{{ config.add_on_result_sid }}/Payloads/{{ config.sid
  }}.json` - single-object response; records path `.`; emits passthrough records.
- `recording_add_on_result_payload_data`: GET `/Accounts/{{ secrets.account_sid }}/Recordings/{{
  config.reference_sid }}/AddOnResults/{{ config.add_on_result_sid }}/Payloads/{{ config.payload_sid
  }}/Data.json` - single-object response; records path `.`; emits passthrough records.
- `recording_add_on_result_payloads`: GET `/Accounts/{{ secrets.account_sid }}/Recordings/{{
  config.reference_sid }}/AddOnResults/{{ config.add_on_result_sid }}/Payloads.json` - records path
  `payloads`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts at 0;
  page size 50; maximum 25 page(s); emits passthrough records.
- `recording_add_on_results`: GET `/Accounts/{{ secrets.account_sid }}/Recordings/{{
  config.reference_sid }}/AddOnResults.json` - records path `add_on_results`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `recording_transcription`: GET `/Accounts/{{ secrets.account_sid }}/Recordings/{{
  config.recording_sid }}/Transcriptions/{{ config.sid }}.json` - single-object response; records
  path `.`; emits passthrough records.
- `recording_transcriptions`: GET `/Accounts/{{ secrets.account_sid }}/Recordings/{{
  config.recording_sid }}/Transcriptions.json` - records path `transcriptions`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `short_code`: GET `/Accounts/{{ secrets.account_sid }}/SMS/ShortCodes/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `short_codes`: GET `/Accounts/{{ secrets.account_sid }}/SMS/ShortCodes.json` - records path
  `short_codes`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts at
  0; page size 50; maximum 25 page(s); emits passthrough records.
- `signing_key`: GET `/Accounts/{{ secrets.account_sid }}/SigningKeys/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `signing_keys`: GET `/Accounts/{{ secrets.account_sid }}/SigningKeys.json` - records path
  `signing_keys`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts
  at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `sip_auth_calls_credential_list_mapping`: GET `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{
  config.domain_sid }}/Auth/Calls/CredentialListMappings/{{ config.sid }}.json` - single-object
  response; records path `.`; emits passthrough records.
- `sip_auth_calls_credential_list_mappings`: GET `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{
  config.domain_sid }}/Auth/Calls/CredentialListMappings.json` - records path `contents`;
  page-number pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size
  50; maximum 25 page(s); emits passthrough records.
- `sip_auth_calls_ip_access_control_list_mapping`: GET `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ config.domain_sid }}/Auth/Calls/IpAccessControlListMappings/{{ config.sid
  }}.json` - single-object response; records path `.`; emits passthrough records.
- `sip_auth_calls_ip_access_control_list_mappings`: GET `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ config.domain_sid }}/Auth/Calls/IpAccessControlListMappings.json` - records path
  `contents`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts at 0;
  page size 50; maximum 25 page(s); emits passthrough records.
- `sip_auth_registrations_credential_list_mapping`: GET `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ config.domain_sid }}/Auth/Registrations/CredentialListMappings/{{ config.sid
  }}.json` - single-object response; records path `.`; emits passthrough records.
- `sip_auth_registrations_credential_list_mappings`: GET `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ config.domain_sid }}/Auth/Registrations/CredentialListMappings.json` - records
  path `contents`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts
  at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `sip_credential`: GET `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists/{{
  config.credential_list_sid }}/Credentials/{{ config.sid }}.json` - single-object response; records
  path `.`; emits passthrough records.
- `sip_credential_list`: GET `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists/{{ config.sid
  }}.json` - single-object response; records path `.`; emits passthrough records.
- `sip_credential_list_mapping`: GET `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{
  config.domain_sid }}/CredentialListMappings/{{ config.sid }}.json` - single-object response;
  records path `.`; emits passthrough records.
- `sip_credential_list_mappings`: GET `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{
  config.domain_sid }}/CredentialListMappings.json` - records path `credential_list_mappings`;
  page-number pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size
  50; maximum 25 page(s); emits passthrough records.
- `sip_credential_lists`: GET `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists.json` -
  records path `credential_lists`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `sip_credentials`: GET `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists/{{
  config.credential_list_sid }}/Credentials.json` - records path `credentials`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `sip_domain`: GET `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `sip_domains`: GET `/Accounts/{{ secrets.account_sid }}/SIP/Domains.json` - records path
  `domains`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts at 0;
  page size 50; maximum 25 page(s); emits passthrough records.
- `sip_ip_access_control_list`: GET `/Accounts/{{ secrets.account_sid }}/SIP/IpAccessControlLists/{{
  config.sid }}.json` - single-object response; records path `.`; emits passthrough records.
- `sip_ip_access_control_list_mapping`: GET `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{
  config.domain_sid }}/IpAccessControlListMappings/{{ config.sid }}.json` - single-object response;
  records path `.`; emits passthrough records.
- `sip_ip_access_control_list_mappings`: GET `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{
  config.domain_sid }}/IpAccessControlListMappings.json` - records path
  `ip_access_control_list_mappings`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `sip_ip_access_control_lists`: GET `/Accounts/{{ secrets.account_sid
  }}/SIP/IpAccessControlLists.json` - records path `ip_access_control_lists`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `sip_ip_address`: GET `/Accounts/{{ secrets.account_sid }}/SIP/IpAccessControlLists/{{
  config.ip_access_control_list_sid }}/IpAddresses.json` - records path `ip_addresses`; page-number
  pagination; page parameter `Page`; size parameter `PageSize`; starts at 0; page size 50; maximum
  25 page(s); emits passthrough records.
- `sip_ip_address_2`: GET `/Accounts/{{ secrets.account_sid }}/SIP/IpAccessControlLists/{{
  config.ip_access_control_list_sid }}/IpAddresses/{{ config.sid }}.json` - single-object response;
  records path `.`; emits passthrough records.
- `transcription`: GET `/Accounts/{{ secrets.account_sid }}/Transcriptions/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `transcriptions`: GET `/Accounts/{{ secrets.account_sid }}/Transcriptions.json` - records path
  `transcriptions`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts
  at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `usage_record_all_times`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Records/AllTime.json` -
  records path `usage_records`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `usage_record_dailies`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Records/Daily.json` -
  records path `usage_records`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `usage_record_last_months`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Records/LastMonth.json`
  - records path `usage_records`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `usage_record_monthlies`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Records/Monthly.json` -
  records path `usage_records`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `usage_record_this_months`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Records/ThisMonth.json`
  - records path `usage_records`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `usage_record_todays`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Records/Today.json` -
  records path `usage_records`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `usage_record_yearlies`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Records/Yearly.json` -
  records path `usage_records`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `usage_record_yesterdays`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Records/Yesterday.json`
  - records path `usage_records`; page-number pagination; page parameter `Page`; size parameter
  `PageSize`; starts at 0; page size 50; maximum 25 page(s); emits passthrough records.
- `usage_trigger`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Triggers/{{ config.sid }}.json` -
  single-object response; records path `.`; emits passthrough records.
- `usage_triggers`: GET `/Accounts/{{ secrets.account_sid }}/Usage/Triggers.json` - records path
  `usage_triggers`; page-number pagination; page parameter `Page`; size parameter `PageSize`; starts
  at 0; page size 50; maximum 25 page(s); emits passthrough records.

## Write actions & risks

Overall write risk: creates, updates, and deletes Twilio v2010 resources including messages, calls,
phone numbers, recordings, queues, SIP resources, keys, and usage triggers.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_account`: POST `/Accounts.json` - kind `create`; body type `form`; accepted fields
  `FriendlyName`; risk: creates Twilio account resources in the connected account; approval
  required.
- `create_address`: POST `/Accounts/{{ secrets.account_sid }}/Addresses.json` - kind `create`; body
  type `form`; required record fields `CustomerName`, `Street`, `City`, `Region`, `PostalCode`,
  `IsoCountry`; accepted fields `AutoCorrectAddress`, `City`, `CustomerName`, `EmergencyEnabled`,
  `FriendlyName`, `IsoCountry`, `PostalCode`, `Region`, `Street`, `StreetSecondary`; risk: creates
  Twilio address resources in the connected account; approval required.
- `delete_address`: DELETE `/Accounts/{{ secrets.account_sid }}/Addresses/{{ record.sid }}.json` -
  kind `delete`; body type `none`; path fields `sid`; required record fields `sid`; accepted fields
  `sid`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  deletes Twilio address resources in the connected account; approval required.
- `update_address`: POST `/Accounts/{{ secrets.account_sid }}/Addresses/{{ record.sid }}.json` -
  kind `update`; body type `form`; path fields `sid`; required record fields `sid`; accepted fields
  `AutoCorrectAddress`, `City`, `CustomerName`, `EmergencyEnabled`, `FriendlyName`, `PostalCode`,
  `Region`, `Street`, `StreetSecondary`, `sid`; risk: mutates Twilio address resources in the
  connected account; approval required.
- `create_application`: POST `/Accounts/{{ secrets.account_sid }}/Applications.json` - kind
  `create`; body type `form`; accepted fields `ApiVersion`, `FriendlyName`, `MessageStatusCallback`,
  `PublicApplicationConnectEnabled`, `SmsFallbackMethod`, `SmsFallbackUrl`, `SmsMethod`,
  `SmsStatusCallback`, `SmsUrl`, `StatusCallback`, `StatusCallbackMethod`, `VoiceCallerIdLookup`,
  `VoiceFallbackMethod`, `VoiceFallbackUrl`, `VoiceMethod`, `VoiceUrl`; risk: creates Twilio
  application resources in the connected account; approval required.
- `delete_application`: DELETE `/Accounts/{{ secrets.account_sid }}/Applications/{{ record.sid
  }}.json` - kind `delete`; body type `none`; path fields `sid`; required record fields `sid`;
  accepted fields `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio application resources in the connected account; approval
  required.
- `update_application`: POST `/Accounts/{{ secrets.account_sid }}/Applications/{{ record.sid
  }}.json` - kind `update`; body type `form`; path fields `sid`; required record fields `sid`;
  accepted fields `ApiVersion`, `FriendlyName`, `MessageStatusCallback`,
  `PublicApplicationConnectEnabled`, `SmsFallbackMethod`, `SmsFallbackUrl`, `SmsMethod`,
  `SmsStatusCallback`, `SmsUrl`, `StatusCallback`, `StatusCallbackMethod`, `VoiceCallerIdLookup`,
  `VoiceFallbackMethod`, `VoiceFallbackUrl`, `VoiceMethod`, `VoiceUrl`, `sid`; risk: mutates Twilio
  application resources in the connected account; approval required.
- `create_call`: POST `/Accounts/{{ secrets.account_sid }}/Calls.json` - kind `create`; body type
  `form`; required record fields `To`, `From`; accepted fields `ApplicationSid`, `AsyncAmd`,
  `AsyncAmdStatusCallback`, `AsyncAmdStatusCallbackMethod`, `Byoc`, `CallReason`, `CallToken`,
  `CallerId`, `ClientNotificationUrl`, `FallbackMethod`, `FallbackUrl`, `From`, `MachineDetection`,
  `MachineDetectionSilenceTimeout`, `MachineDetectionSpeechEndThreshold`,
  `MachineDetectionSpeechThreshold`, `MachineDetectionTimeout`, `Method`, and 19 more; risk: creates
  Twilio call resources in the connected account; approval required.
- `create_payments`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid
  }}/Payments.json` - kind `create`; body type `form`; path fields `call_sid`; required record
  fields `call_sid`, `IdempotencyKey`, `StatusCallback`; accepted fields `BankAccountType`,
  `ChargeAmount`, `Confirmation`, `Currency`, `Description`, `IdempotencyKey`, `Input`,
  `MinPostalCodeLength`, `Parameter`, `PaymentConnector`, `PaymentMethod`, `PostalCode`,
  `RequireMatchingInputs`, `SecurityCode`, `StatusCallback`, `Timeout`, `TokenType`,
  `ValidCardTypes`, and 1 more; risk: creates Twilio payments resources in the connected account;
  approval required.
- `update_payments`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid
  }}/Payments/{{ record.sid }}.json` - kind `update`; body type `form`; path fields `call_sid`,
  `sid`; required record fields `call_sid`, `sid`, `IdempotencyKey`, `StatusCallback`; accepted
  fields `Capture`, `IdempotencyKey`, `Status`, `StatusCallback`, `call_sid`, `sid`; risk: mutates
  Twilio payments resources in the connected account; approval required.
- `create_call_recording`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid
  }}/Recordings.json` - kind `create`; body type `form`; path fields `call_sid`; required record
  fields `call_sid`; accepted fields `RecordingChannels`, `RecordingConfigurationId`,
  `RecordingStatusCallback`, `RecordingStatusCallbackEvent`, `RecordingStatusCallbackMethod`,
  `RecordingTrack`, `Trim`, `call_sid`; risk: creates Twilio call recording resources in the
  connected account; approval required.
- `delete_call_recording`: DELETE `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid
  }}/Recordings/{{ record.sid }}.json` - kind `delete`; body type `none`; path fields `call_sid`,
  `sid`; required record fields `call_sid`, `sid`; accepted fields `call_sid`, `sid`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: deletes Twilio call
  recording resources in the connected account; approval required.
- `update_call_recording`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid
  }}/Recordings/{{ record.sid }}.json` - kind `update`; body type `form`; path fields `call_sid`,
  `sid`; required record fields `call_sid`, `sid`, `Status`; accepted fields `PauseBehavior`,
  `Status`, `call_sid`, `sid`; risk: mutates Twilio call recording resources in the connected
  account; approval required.
- `create_siprec`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid
  }}/Siprec.json` - kind `create`; body type `form`; path fields `call_sid`; required record fields
  `call_sid`; accepted fields `ConnectorName`, `Name`, `Parameter1.Name`, `Parameter1.Value`,
  `Parameter10.Name`, `Parameter10.Value`, `Parameter11.Name`, `Parameter11.Value`,
  `Parameter12.Name`, `Parameter12.Value`, `Parameter13.Name`, `Parameter13.Value`,
  `Parameter14.Name`, `Parameter14.Value`, `Parameter15.Name`, `Parameter15.Value`,
  `Parameter16.Name`, `Parameter16.Value`, and 186 more; risk: creates Twilio siprec resources in
  the connected account; approval required.
- `update_siprec`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid }}/Siprec/{{
  record.sid }}.json` - kind `update`; body type `form`; path fields `call_sid`, `sid`; required
  record fields `call_sid`, `sid`, `Status`; accepted fields `Status`, `call_sid`, `sid`; risk:
  mutates Twilio siprec resources in the connected account; approval required.
- `create_stream`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid
  }}/Streams.json` - kind `create`; body type `form`; path fields `call_sid`; required record fields
  `call_sid`, `Url`; accepted fields `Name`, `Parameter1.Name`, `Parameter1.Value`,
  `Parameter10.Name`, `Parameter10.Value`, `Parameter11.Name`, `Parameter11.Value`,
  `Parameter12.Name`, `Parameter12.Value`, `Parameter13.Name`, `Parameter13.Value`,
  `Parameter14.Name`, `Parameter14.Value`, `Parameter15.Name`, `Parameter15.Value`,
  `Parameter16.Name`, `Parameter16.Value`, `Parameter17.Name`, and 186 more; risk: creates Twilio
  stream resources in the connected account; approval required.
- `update_stream`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid }}/Streams/{{
  record.sid }}.json` - kind `update`; body type `form`; path fields `call_sid`, `sid`; required
  record fields `call_sid`, `sid`, `Status`; accepted fields `Status`, `call_sid`, `sid`; risk:
  mutates Twilio stream resources in the connected account; approval required.
- `create_realtime_transcription`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{
  record.call_sid }}/Transcriptions.json` - kind `create`; body type `form`; path fields `call_sid`;
  required record fields `call_sid`; accepted fields `ConversationConfiguration`, `ConversationId`,
  `EnableAutomaticPunctuation`, `EnableProviderData`, `Hints`, `InboundTrackLabel`,
  `IntelligenceService`, `LanguageCode`, `Name`, `OutboundTrackLabel`, `PartialResults`,
  `ProfanityFilter`, `SpeechModel`, `StatusCallbackMethod`, `StatusCallbackUrl`, `Track`,
  `TranscriptionConfigurationId`, `TranscriptionEngine`, and 1 more; risk: creates Twilio realtime
  transcription resources in the connected account; approval required.
- `update_realtime_transcription`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{
  record.call_sid }}/Transcriptions/{{ record.sid }}.json` - kind `update`; body type `form`; path
  fields `call_sid`, `sid`; required record fields `call_sid`, `sid`, `Status`; accepted fields
  `Status`, `call_sid`, `sid`; risk: mutates Twilio realtime transcription resources in the
  connected account; approval required.
- `create_user_defined_message`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.call_sid
  }}/UserDefinedMessages.json` - kind `create`; body type `form`; path fields `call_sid`; required
  record fields `call_sid`, `Content`; accepted fields `Content`, `IdempotencyKey`, `call_sid`;
  risk: creates Twilio user defined message resources in the connected account; approval required.
- `create_user_defined_message_subscription`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{
  record.call_sid }}/UserDefinedMessageSubscriptions.json` - kind `create`; body type `form`; path
  fields `call_sid`; required record fields `call_sid`, `Callback`; accepted fields `Callback`,
  `IdempotencyKey`, `Method`, `call_sid`; risk: creates Twilio user defined message subscription
  resources in the connected account; approval required.
- `delete_user_defined_message_subscription`: DELETE `/Accounts/{{ secrets.account_sid }}/Calls/{{
  record.call_sid }}/UserDefinedMessageSubscriptions/{{ record.sid }}.json` - kind `delete`; body
  type `none`; path fields `call_sid`, `sid`; required record fields `call_sid`, `sid`; accepted
  fields `call_sid`, `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio user defined message subscription resources in the connected
  account; approval required.
- `delete_call`: DELETE `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.sid }}.json` - kind
  `delete`; body type `none`; path fields `sid`; required record fields `sid`; accepted fields
  `sid`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  deletes Twilio call resources in the connected account; approval required.
- `update_call`: POST `/Accounts/{{ secrets.account_sid }}/Calls/{{ record.sid }}.json` - kind
  `update`; body type `form`; path fields `sid`; required record fields `sid`; accepted fields
  `FallbackMethod`, `FallbackUrl`, `Method`, `Status`, `StatusCallback`, `StatusCallbackMethod`,
  `TimeLimit`, `Twiml`, `Url`, `sid`; risk: mutates Twilio call resources in the connected account;
  approval required.
- `create_participant`: POST `/Accounts/{{ secrets.account_sid }}/Conferences/{{
  record.conference_sid }}/Participants.json` - kind `create`; body type `form`; path fields
  `conference_sid`; required record fields `conference_sid`, `From`, `To`; accepted fields
  `AmdStatusCallback`, `AmdStatusCallbackMethod`, `Beep`, `Byoc`, `CallReason`, `CallSidToCoach`,
  `CallToken`, `CallerDisplayName`, `CallerId`, `ClientNotificationUrl`, `Coaching`,
  `ConferenceRecord`, `ConferenceRecordingStatusCallback`, `ConferenceRecordingStatusCallbackEvent`,
  `ConferenceRecordingStatusCallbackMethod`, `ConferenceStatusCallback`,
  `ConferenceStatusCallbackEvent`, `ConferenceStatusCallbackMethod`, and 34 more; risk: creates
  Twilio participant resources in the connected account; approval required.
- `delete_participant`: DELETE `/Accounts/{{ secrets.account_sid }}/Conferences/{{
  record.conference_sid }}/Participants/{{ record.call_sid }}.json` - kind `delete`; body type
  `none`; path fields `conference_sid`, `call_sid`; required record fields `conference_sid`,
  `call_sid`; accepted fields `call_sid`, `conference_sid`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes Twilio participant resources in the
  connected account; approval required.
- `update_participant`: POST `/Accounts/{{ secrets.account_sid }}/Conferences/{{
  record.conference_sid }}/Participants/{{ record.call_sid }}.json` - kind `update`; body type
  `form`; path fields `conference_sid`, `call_sid`; required record fields `conference_sid`,
  `call_sid`; accepted fields `AnnounceMethod`, `AnnounceUrl`, `BeepOnExit`, `CallSidToCoach`,
  `Coaching`, `EndConferenceOnExit`, `Hold`, `HoldMethod`, `HoldUrl`, `Muted`, `WaitMethod`,
  `WaitUrl`, `call_sid`, `conference_sid`; risk: mutates Twilio participant resources in the
  connected account; approval required.
- `delete_conference_recording`: DELETE `/Accounts/{{ secrets.account_sid }}/Conferences/{{
  record.conference_sid }}/Recordings/{{ record.sid }}.json` - kind `delete`; body type `none`; path
  fields `conference_sid`, `sid`; required record fields `conference_sid`, `sid`; accepted fields
  `conference_sid`, `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio conference recording resources in the connected account;
  approval required.
- `update_conference_recording`: POST `/Accounts/{{ secrets.account_sid }}/Conferences/{{
  record.conference_sid }}/Recordings/{{ record.sid }}.json` - kind `update`; body type `form`; path
  fields `conference_sid`, `sid`; required record fields `conference_sid`, `sid`, `Status`; accepted
  fields `PauseBehavior`, `Status`, `conference_sid`, `sid`; risk: mutates Twilio conference
  recording resources in the connected account; approval required.
- `update_conference`: POST `/Accounts/{{ secrets.account_sid }}/Conferences/{{ record.sid }}.json`
  - kind `update`; body type `form`; path fields `sid`; required record fields `sid`; accepted
  fields `AnnounceMethod`, `AnnounceUrl`, `Status`, `sid`; risk: mutates Twilio conference resources
  in the connected account; approval required.
- `delete_connect_app`: DELETE `/Accounts/{{ secrets.account_sid }}/ConnectApps/{{ record.sid
  }}.json` - kind `delete`; body type `none`; path fields `sid`; required record fields `sid`;
  accepted fields `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio connect app resources in the connected account; approval
  required.
- `update_connect_app`: POST `/Accounts/{{ secrets.account_sid }}/ConnectApps/{{ record.sid }}.json`
  - kind `update`; body type `form`; path fields `sid`; required record fields `sid`; accepted
  fields `AuthorizeRedirectUrl`, `CompanyName`, `DeauthorizeCallbackMethod`,
  `DeauthorizeCallbackUrl`, `Description`, `FriendlyName`, `HomepageUrl`, `Permissions`, `sid`;
  risk: mutates Twilio connect app resources in the connected account; approval required.
- `create_incoming_phone_number`: POST `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers.json` - kind `create`; body type `form`; accepted fields `AddressSid`,
  `ApiVersion`, `AreaCode`, `BundleSid`, `EmergencyAddressSid`, `EmergencyStatus`, `FriendlyName`,
  `IdentitySid`, `PhoneNumber`, `SmsApplicationSid`, `SmsFallbackMethod`, `SmsFallbackUrl`,
  `SmsMethod`, `SmsUrl`, `StatusCallback`, `StatusCallbackMethod`, `TrunkSid`,
  `VoiceApplicationSid`, and 6 more; risk: creates Twilio incoming phone number resources in the
  connected account; approval required.
- `create_incoming_phone_number_assigned_add_on`: POST `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/{{ record.resource_sid }}/AssignedAddOns.json` - kind `create`; body type
  `form`; path fields `resource_sid`; required record fields `resource_sid`, `InstalledAddOnSid`;
  accepted fields `InstalledAddOnSid`, `resource_sid`; risk: creates Twilio incoming phone number
  assigned add on resources in the connected account; approval required.
- `delete_incoming_phone_number_assigned_add_on`: DELETE `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/{{ record.resource_sid }}/AssignedAddOns/{{ record.sid }}.json` - kind
  `delete`; body type `none`; path fields `resource_sid`, `sid`; required record fields
  `resource_sid`, `sid`; accepted fields `resource_sid`, `sid`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes Twilio incoming phone number assigned
  add on resources in the connected account; approval required.
- `delete_incoming_phone_number`: DELETE `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/{{ record.sid }}.json` - kind `delete`; body type `none`; path fields
  `sid`; required record fields `sid`; accepted fields `sid`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes Twilio incoming phone number resources in
  the connected account; approval required.
- `update_incoming_phone_number`: POST `/Accounts/{{ secrets.account_sid }}/IncomingPhoneNumbers/{{
  record.sid }}.json` - kind `update`; body type `form`; path fields `sid`; required record fields
  `sid`; accepted fields `AccountSid`, `AddressSid`, `ApiVersion`, `BundleSid`,
  `EmergencyAddressSid`, `EmergencyStatus`, `FriendlyName`, `IdentitySid`, `SmsApplicationSid`,
  `SmsFallbackMethod`, `SmsFallbackUrl`, `SmsMethod`, `SmsUrl`, `StatusCallback`,
  `StatusCallbackMethod`, `TrunkSid`, `VoiceApplicationSid`, `VoiceCallerIdLookup`, and 6 more;
  risk: mutates Twilio incoming phone number resources in the connected account; approval required.
- `create_incoming_phone_number_local`: POST `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/Local.json` - kind `create`; body type `form`; required record fields
  `PhoneNumber`; accepted fields `AddressSid`, `ApiVersion`, `BundleSid`, `EmergencyAddressSid`,
  `EmergencyStatus`, `FriendlyName`, `IdentitySid`, `PhoneNumber`, `SmsApplicationSid`,
  `SmsFallbackMethod`, `SmsFallbackUrl`, `SmsMethod`, `SmsUrl`, `StatusCallback`,
  `StatusCallbackMethod`, `TrunkSid`, `VoiceApplicationSid`, `VoiceCallerIdLookup`, and 5 more;
  risk: creates Twilio incoming phone number local resources in the connected account; approval
  required.
- `create_incoming_phone_number_mobile`: POST `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/Mobile.json` - kind `create`; body type `form`; required record fields
  `PhoneNumber`; accepted fields `AddressSid`, `ApiVersion`, `BundleSid`, `EmergencyAddressSid`,
  `EmergencyStatus`, `FriendlyName`, `IdentitySid`, `PhoneNumber`, `SmsApplicationSid`,
  `SmsFallbackMethod`, `SmsFallbackUrl`, `SmsMethod`, `SmsUrl`, `StatusCallback`,
  `StatusCallbackMethod`, `TrunkSid`, `VoiceApplicationSid`, `VoiceCallerIdLookup`, and 5 more;
  risk: creates Twilio incoming phone number mobile resources in the connected account; approval
  required.
- `create_incoming_phone_number_toll_free`: POST `/Accounts/{{ secrets.account_sid
  }}/IncomingPhoneNumbers/TollFree.json` - kind `create`; body type `form`; required record fields
  `PhoneNumber`; accepted fields `AddressSid`, `ApiVersion`, `BundleSid`, `EmergencyAddressSid`,
  `EmergencyStatus`, `FriendlyName`, `IdentitySid`, `PhoneNumber`, `SmsApplicationSid`,
  `SmsFallbackMethod`, `SmsFallbackUrl`, `SmsMethod`, `SmsUrl`, `StatusCallback`,
  `StatusCallbackMethod`, `TrunkSid`, `VoiceApplicationSid`, `VoiceCallerIdLookup`, and 5 more;
  risk: creates Twilio incoming phone number toll free resources in the connected account; approval
  required.
- `create_key`: POST `/Accounts/{{ secrets.account_sid }}/Keys.json` - kind `create`; body type
  `form`; accepted fields `FriendlyName`; risk: creates Twilio key resources in the connected
  account; approval required.
- `delete_key`: DELETE `/Accounts/{{ secrets.account_sid }}/Keys/{{ record.sid }}.json` - kind
  `delete`; body type `none`; path fields `sid`; required record fields `sid`; accepted fields
  `sid`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  deletes Twilio key resources in the connected account; approval required.
- `update_key`: POST `/Accounts/{{ secrets.account_sid }}/Keys/{{ record.sid }}.json` - kind
  `update`; body type `form`; path fields `sid`; required record fields `sid`; accepted fields
  `FriendlyName`, `sid`; risk: mutates Twilio key resources in the connected account; approval
  required.
- `create_message`: POST `/Accounts/{{ secrets.account_sid }}/Messages.json` - kind `create`; body
  type `form`; required record fields `To`; accepted fields `AddressRetention`, `ApplicationSid`,
  `Attempt`, `Body`, `ContentRetention`, `ContentSid`, `ContentVariables`, `FallbackFrom`,
  `ForceDelivery`, `From`, `MaxPrice`, `MediaUrl`, `MessagingServiceSid`, `PersistentAction`,
  `ProvideFeedback`, `RiskCheck`, `ScheduleType`, `SendAsMms`, and 7 more; risk: creates Twilio
  message resources in the connected account; approval required.
- `create_message_feedback`: POST `/Accounts/{{ secrets.account_sid }}/Messages/{{
  record.message_sid }}/Feedback.json` - kind `create`; body type `form`; path fields `message_sid`;
  required record fields `message_sid`; accepted fields `Outcome`, `message_sid`; risk: creates
  Twilio message feedback resources in the connected account; approval required.
- `delete_media`: DELETE `/Accounts/{{ secrets.account_sid }}/Messages/{{ record.message_sid
  }}/Media/{{ record.sid }}.json` - kind `delete`; body type `none`; path fields `message_sid`,
  `sid`; required record fields `message_sid`, `sid`; accepted fields `message_sid`, `sid`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: deletes Twilio
  media resources in the connected account; approval required.
- `delete_message`: DELETE `/Accounts/{{ secrets.account_sid }}/Messages/{{ record.sid }}.json` -
  kind `delete`; body type `none`; path fields `sid`; required record fields `sid`; accepted fields
  `sid`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  deletes Twilio message resources in the connected account; approval required.
- `update_message`: POST `/Accounts/{{ secrets.account_sid }}/Messages/{{ record.sid }}.json` - kind
  `update`; body type `form`; path fields `sid`; required record fields `sid`; accepted fields
  `Body`, `Status`, `sid`; risk: mutates Twilio message resources in the connected account; approval
  required.
- `create_outgoing_caller_id_validation_request`: POST `/Accounts/{{ secrets.account_sid
  }}/OutgoingCallerIds.json` - kind `create`; body type `form`; required record fields
  `PhoneNumber`; accepted fields `CallDelay`, `Extension`, `FriendlyName`, `PhoneNumber`,
  `StatusCallback`, `StatusCallbackMethod`; risk: creates Twilio validation request resources in the
  connected account; approval required.
- `delete_outgoing_caller_id`: DELETE `/Accounts/{{ secrets.account_sid }}/OutgoingCallerIds/{{
  record.sid }}.json` - kind `delete`; body type `none`; path fields `sid`; required record fields
  `sid`; accepted fields `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio outgoing caller id resources in the connected account;
  approval required.
- `update_outgoing_caller_id`: POST `/Accounts/{{ secrets.account_sid }}/OutgoingCallerIds/{{
  record.sid }}.json` - kind `update`; body type `form`; path fields `sid`; required record fields
  `sid`; accepted fields `FriendlyName`, `sid`; risk: mutates Twilio outgoing caller id resources in
  the connected account; approval required.
- `create_queue`: POST `/Accounts/{{ secrets.account_sid }}/Queues.json` - kind `create`; body type
  `form`; required record fields `FriendlyName`; accepted fields `FriendlyName`, `MaxSize`; risk:
  creates Twilio queue resources in the connected account; approval required.
- `update_member`: POST `/Accounts/{{ secrets.account_sid }}/Queues/{{ record.queue_sid
  }}/Members/{{ record.call_sid }}.json` - kind `update`; body type `form`; path fields `queue_sid`,
  `call_sid`; required record fields `queue_sid`, `call_sid`, `Url`; accepted fields `Method`,
  `Url`, `call_sid`, `queue_sid`; risk: mutates Twilio member resources in the connected account;
  approval required.
- `delete_queue`: DELETE `/Accounts/{{ secrets.account_sid }}/Queues/{{ record.sid }}.json` - kind
  `delete`; body type `none`; path fields `sid`; required record fields `sid`; accepted fields
  `sid`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  deletes Twilio queue resources in the connected account; approval required.
- `update_queue`: POST `/Accounts/{{ secrets.account_sid }}/Queues/{{ record.sid }}.json` - kind
  `update`; body type `form`; path fields `sid`; required record fields `sid`; accepted fields
  `FriendlyName`, `MaxSize`, `sid`; risk: mutates Twilio queue resources in the connected account;
  approval required.
- `delete_recording_transcription`: DELETE `/Accounts/{{ secrets.account_sid }}/Recordings/{{
  record.recording_sid }}/Transcriptions/{{ record.sid }}.json` - kind `delete`; body type `none`;
  path fields `recording_sid`, `sid`; required record fields `recording_sid`, `sid`; accepted fields
  `recording_sid`, `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio recording transcription resources in the connected account;
  approval required.
- `delete_recording_add_on_result_payload`: DELETE `/Accounts/{{ secrets.account_sid
  }}/Recordings/{{ record.reference_sid }}/AddOnResults/{{ record.add_on_result_sid }}/Payloads/{{
  record.sid }}.json` - kind `delete`; body type `none`; path fields `reference_sid`,
  `add_on_result_sid`, `sid`; required record fields `reference_sid`, `add_on_result_sid`, `sid`;
  accepted fields `add_on_result_sid`, `reference_sid`, `sid`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes Twilio recording add on result payload
  resources in the connected account; approval required.
- `delete_recording_add_on_result`: DELETE `/Accounts/{{ secrets.account_sid }}/Recordings/{{
  record.reference_sid }}/AddOnResults/{{ record.sid }}.json` - kind `delete`; body type `none`;
  path fields `reference_sid`, `sid`; required record fields `reference_sid`, `sid`; accepted fields
  `reference_sid`, `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio recording add on result resources in the connected account;
  approval required.
- `delete_recording`: DELETE `/Accounts/{{ secrets.account_sid }}/Recordings/{{ record.sid }}.json`
  - kind `delete`; body type `none`; path fields `sid`; required record fields `sid`; accepted
  fields `sid`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: deletes Twilio recording resources in the connected account; approval required.
- `create_signing_key`: POST `/Accounts/{{ secrets.account_sid }}/SigningKeys.json` - kind `create`;
  body type `form`; accepted fields `FriendlyName`; risk: creates Twilio signing key resources in
  the connected account; approval required.
- `delete_signing_key`: DELETE `/Accounts/{{ secrets.account_sid }}/SigningKeys/{{ record.sid
  }}.json` - kind `delete`; body type `none`; path fields `sid`; required record fields `sid`;
  accepted fields `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio signing key resources in the connected account; approval
  required.
- `update_signing_key`: POST `/Accounts/{{ secrets.account_sid }}/SigningKeys/{{ record.sid }}.json`
  - kind `update`; body type `form`; path fields `sid`; required record fields `sid`; accepted
  fields `FriendlyName`, `sid`; risk: mutates Twilio signing key resources in the connected account;
  approval required.
- `create_sip_credential_list`: POST `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists.json`
  - kind `create`; body type `form`; required record fields `FriendlyName`; accepted fields
  `FriendlyName`; risk: creates Twilio sip credential list resources in the connected account;
  approval required.
- `create_sip_credential`: POST `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists/{{
  record.credential_list_sid }}/Credentials.json` - kind `create`; body type `form`; path fields
  `credential_list_sid`; required record fields `credential_list_sid`, `Username`, `Password`;
  accepted fields `Password`, `Username`, `credential_list_sid`; risk: creates Twilio sip credential
  resources in the connected account; approval required.
- `delete_sip_credential`: DELETE `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists/{{
  record.credential_list_sid }}/Credentials/{{ record.sid }}.json` - kind `delete`; body type
  `none`; path fields `credential_list_sid`, `sid`; required record fields `credential_list_sid`,
  `sid`; accepted fields `credential_list_sid`, `sid`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: deletes Twilio sip credential resources in the connected
  account; approval required.
- `update_sip_credential`: POST `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists/{{
  record.credential_list_sid }}/Credentials/{{ record.sid }}.json` - kind `update`; body type
  `form`; path fields `credential_list_sid`, `sid`; required record fields `credential_list_sid`,
  `sid`; accepted fields `Password`, `credential_list_sid`, `sid`; risk: mutates Twilio sip
  credential resources in the connected account; approval required.
- `delete_sip_credential_list`: DELETE `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists/{{
  record.sid }}.json` - kind `delete`; body type `none`; path fields `sid`; required record fields
  `sid`; accepted fields `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio sip credential list resources in the connected account;
  approval required.
- `update_sip_credential_list`: POST `/Accounts/{{ secrets.account_sid }}/SIP/CredentialLists/{{
  record.sid }}.json` - kind `update`; body type `form`; path fields `sid`; required record fields
  `sid`, `FriendlyName`; accepted fields `FriendlyName`, `sid`; risk: mutates Twilio sip credential
  list resources in the connected account; approval required.
- `create_sip_domain`: POST `/Accounts/{{ secrets.account_sid }}/SIP/Domains.json` - kind `create`;
  body type `form`; required record fields `DomainName`; accepted fields `ByocTrunkSid`,
  `DomainName`, `EmergencyCallerSid`, `EmergencyCallingEnabled`, `FriendlyName`, `Secure`,
  `SipRegistration`, `VoiceFallbackMethod`, `VoiceFallbackUrl`, `VoiceMethod`,
  `VoiceStatusCallbackMethod`, `VoiceStatusCallbackUrl`, `VoiceUrl`; risk: creates Twilio sip domain
  resources in the connected account; approval required.
- `create_sip_auth_calls_credential_list_mapping`: POST `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ record.domain_sid }}/Auth/Calls/CredentialListMappings.json` - kind `create`;
  body type `form`; path fields `domain_sid`; required record fields `domain_sid`,
  `CredentialListSid`; accepted fields `CredentialListSid`, `domain_sid`; risk: creates Twilio sip
  auth calls credential list mapping resources in the connected account; approval required.
- `delete_sip_auth_calls_credential_list_mapping`: DELETE `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ record.domain_sid }}/Auth/Calls/CredentialListMappings/{{ record.sid }}.json` -
  kind `delete`; body type `none`; path fields `domain_sid`, `sid`; required record fields
  `domain_sid`, `sid`; accepted fields `domain_sid`, `sid`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes Twilio sip auth calls credential list
  mapping resources in the connected account; approval required.
- `create_sip_auth_calls_ip_access_control_list_mapping`: POST `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ record.domain_sid }}/Auth/Calls/IpAccessControlListMappings.json` - kind
  `create`; body type `form`; path fields `domain_sid`; required record fields `domain_sid`,
  `IpAccessControlListSid`; accepted fields `IpAccessControlListSid`, `domain_sid`; risk: creates
  Twilio sip auth calls ip access control list mapping resources in the connected account; approval
  required.
- `delete_sip_auth_calls_ip_access_control_list_mapping`: DELETE `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ record.domain_sid }}/Auth/Calls/IpAccessControlListMappings/{{ record.sid
  }}.json` - kind `delete`; body type `none`; path fields `domain_sid`, `sid`; required record
  fields `domain_sid`, `sid`; accepted fields `domain_sid`, `sid`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes Twilio sip auth calls ip
  access control list mapping resources in the connected account; approval required.
- `create_sip_auth_registrations_credential_list_mapping`: POST `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ record.domain_sid }}/Auth/Registrations/CredentialListMappings.json` - kind
  `create`; body type `form`; path fields `domain_sid`; required record fields `domain_sid`,
  `CredentialListSid`; accepted fields `CredentialListSid`, `domain_sid`; risk: creates Twilio sip
  auth registrations credential list mapping resources in the connected account; approval required.
- `delete_sip_auth_registrations_credential_list_mapping`: DELETE `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ record.domain_sid }}/Auth/Registrations/CredentialListMappings/{{ record.sid
  }}.json` - kind `delete`; body type `none`; path fields `domain_sid`, `sid`; required record
  fields `domain_sid`, `sid`; accepted fields `domain_sid`, `sid`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes Twilio sip auth registrations
  credential list mapping resources in the connected account; approval required.
- `create_sip_credential_list_mapping`: POST `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{
  record.domain_sid }}/CredentialListMappings.json` - kind `create`; body type `form`; path fields
  `domain_sid`; required record fields `domain_sid`, `CredentialListSid`; accepted fields
  `CredentialListSid`, `domain_sid`; risk: creates Twilio sip credential list mapping resources in
  the connected account; approval required.
- `delete_sip_credential_list_mapping`: DELETE `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{
  record.domain_sid }}/CredentialListMappings/{{ record.sid }}.json` - kind `delete`; body type
  `none`; path fields `domain_sid`, `sid`; required record fields `domain_sid`, `sid`; accepted
  fields `domain_sid`, `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio sip credential list mapping resources in the connected
  account; approval required.
- `create_sip_ip_access_control_list_mapping`: POST `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ record.domain_sid }}/IpAccessControlListMappings.json` - kind `create`; body
  type `form`; path fields `domain_sid`; required record fields `domain_sid`,
  `IpAccessControlListSid`; accepted fields `IpAccessControlListSid`, `domain_sid`; risk: creates
  Twilio sip ip access control list mapping resources in the connected account; approval required.
- `delete_sip_ip_access_control_list_mapping`: DELETE `/Accounts/{{ secrets.account_sid
  }}/SIP/Domains/{{ record.domain_sid }}/IpAccessControlListMappings/{{ record.sid }}.json` - kind
  `delete`; body type `none`; path fields `domain_sid`, `sid`; required record fields `domain_sid`,
  `sid`; accepted fields `domain_sid`, `sid`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes Twilio sip ip access control list mapping resources in
  the connected account; approval required.
- `delete_sip_domain`: DELETE `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{ record.sid
  }}.json` - kind `delete`; body type `none`; path fields `sid`; required record fields `sid`;
  accepted fields `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio sip domain resources in the connected account; approval
  required.
- `update_sip_domain`: POST `/Accounts/{{ secrets.account_sid }}/SIP/Domains/{{ record.sid }}.json`
  - kind `update`; body type `form`; path fields `sid`; required record fields `sid`; accepted
  fields `ByocTrunkSid`, `DomainName`, `EmergencyCallerSid`, `EmergencyCallingEnabled`,
  `FriendlyName`, `Secure`, `SipRegistration`, `VoiceFallbackMethod`, `VoiceFallbackUrl`,
  `VoiceMethod`, `VoiceStatusCallbackMethod`, `VoiceStatusCallbackUrl`, `VoiceUrl`, `sid`; risk:
  mutates Twilio sip domain resources in the connected account; approval required.
- `create_sip_ip_access_control_list`: POST `/Accounts/{{ secrets.account_sid
  }}/SIP/IpAccessControlLists.json` - kind `create`; body type `form`; required record fields
  `FriendlyName`; accepted fields `FriendlyName`; risk: creates Twilio sip ip access control list
  resources in the connected account; approval required.
- `create_sip_ip_address`: POST `/Accounts/{{ secrets.account_sid }}/SIP/IpAccessControlLists/{{
  record.ip_access_control_list_sid }}/IpAddresses.json` - kind `create`; body type `form`; path
  fields `ip_access_control_list_sid`; required record fields `ip_access_control_list_sid`,
  `FriendlyName`, `IpAddress`; accepted fields `CidrPrefixLength`, `FriendlyName`, `IpAddress`,
  `ip_access_control_list_sid`; risk: creates Twilio sip ip address resources in the connected
  account; approval required.
- `delete_sip_ip_address`: DELETE `/Accounts/{{ secrets.account_sid }}/SIP/IpAccessControlLists/{{
  record.ip_access_control_list_sid }}/IpAddresses/{{ record.sid }}.json` - kind `delete`; body type
  `none`; path fields `ip_access_control_list_sid`, `sid`; required record fields
  `ip_access_control_list_sid`, `sid`; accepted fields `ip_access_control_list_sid`, `sid`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: deletes Twilio sip
  ip address resources in the connected account; approval required.
- `update_sip_ip_address`: POST `/Accounts/{{ secrets.account_sid }}/SIP/IpAccessControlLists/{{
  record.ip_access_control_list_sid }}/IpAddresses/{{ record.sid }}.json` - kind `update`; body type
  `form`; path fields `ip_access_control_list_sid`, `sid`; required record fields
  `ip_access_control_list_sid`, `sid`; accepted fields `CidrPrefixLength`, `FriendlyName`,
  `IpAddress`, `ip_access_control_list_sid`, `sid`; risk: mutates Twilio sip ip address resources in
  the connected account; approval required.
- `delete_sip_ip_access_control_list`: DELETE `/Accounts/{{ secrets.account_sid
  }}/SIP/IpAccessControlLists/{{ record.sid }}.json` - kind `delete`; body type `none`; path fields
  `sid`; required record fields `sid`; accepted fields `sid`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes Twilio sip ip access control list
  resources in the connected account; approval required.
- `update_sip_ip_access_control_list`: POST `/Accounts/{{ secrets.account_sid
  }}/SIP/IpAccessControlLists/{{ record.sid }}.json` - kind `update`; body type `form`; path fields
  `sid`; required record fields `sid`, `FriendlyName`; accepted fields `FriendlyName`, `sid`; risk:
  mutates Twilio sip ip access control list resources in the connected account; approval required.
- `update_short_code`: POST `/Accounts/{{ secrets.account_sid }}/SMS/ShortCodes/{{ record.sid
  }}.json` - kind `update`; body type `form`; path fields `sid`; required record fields `sid`;
  accepted fields `ApiVersion`, `FriendlyName`, `SmsFallbackMethod`, `SmsFallbackUrl`, `SmsMethod`,
  `SmsUrl`, `sid`; risk: mutates Twilio short code resources in the connected account; approval
  required.
- `create_token`: POST `/Accounts/{{ secrets.account_sid }}/Tokens.json` - kind `create`; body type
  `form`; accepted fields `Ttl`; risk: creates Twilio token resources in the connected account;
  approval required.
- `delete_transcription`: DELETE `/Accounts/{{ secrets.account_sid }}/Transcriptions/{{ record.sid
  }}.json` - kind `delete`; body type `none`; path fields `sid`; required record fields `sid`;
  accepted fields `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio transcription resources in the connected account; approval
  required.
- `create_usage_trigger`: POST `/Accounts/{{ secrets.account_sid }}/Usage/Triggers.json` - kind
  `create`; body type `form`; required record fields `CallbackUrl`, `TriggerValue`, `UsageCategory`;
  accepted fields `CallbackMethod`, `CallbackUrl`, `FriendlyName`, `Recurring`, `TriggerBy`,
  `TriggerValue`, `UsageCategory`; risk: creates Twilio usage trigger resources in the connected
  account; approval required.
- `delete_usage_trigger`: DELETE `/Accounts/{{ secrets.account_sid }}/Usage/Triggers/{{ record.sid
  }}.json` - kind `delete`; body type `none`; path fields `sid`; required record fields `sid`;
  accepted fields `sid`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes Twilio usage trigger resources in the connected account; approval
  required.
- `update_usage_trigger`: POST `/Accounts/{{ secrets.account_sid }}/Usage/Triggers/{{ record.sid
  }}.json` - kind `update`; body type `form`; path fields `sid`; required record fields `sid`;
  accepted fields `CallbackMethod`, `CallbackUrl`, `FriendlyName`, `sid`; risk: mutates Twilio usage
  trigger resources in the connected account; approval required.
- `update_account`: POST `/Accounts/{{ record.sid }}.json` - kind `update`; body type `form`; path
  fields `sid`; required record fields `sid`; accepted fields `FriendlyName`, `Status`, `sid`; risk:
  mutates Twilio account resources in the connected account; approval required.

## Known limits

- Batch defaults: read_page_size=50, write_batch_size=1.
- API coverage includes 103 stream-backed endpoint group(s), 94 write-backed endpoint group(s).
