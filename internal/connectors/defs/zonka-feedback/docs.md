# Overview

Reads Zonka Feedback responses, surveys, contacts, devices, tasks, locations, users, workspaces,
stats, and distribution logs; writes responses, contacts, survey sends, and tasks through the Zonka
Feedback REST API.

Readable streams: `responses`, `surveys`, `contacts`, `response_details`, `workspaces`,
`survey_stats`, `survey_details`, `contact_details`, `contact_segments`, `devices`,
`device_details`, `device_uptime`, `device_responses`, `tasks`, `locations`, `location_details`,
`users`, `user_details`, `distribution_logs`.

Write actions: `add_response`, `update_response`, `upsert_contacts`, `send_email_survey`,
`send_sms_survey`, `send_two_way_sms_survey`, `send_whatsapp_survey`, `add_task`, `update_task`,
`delete_tasks`.

Service API documentation: https://developers.zonkafeedback.com/docs/api-and-webhooks.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Fallback Zonka Feedback API token, used only when
  auth_token is unset. Sent as the Z-API-TOKEN header. Never logged.
- `auth_token` (optional, secret, string); Zonka Feedback API token, sent as the Z-API-TOKEN header.
  Takes precedence over access_token when both are set. Never logged.
- `base_url` (optional, string); default `https://us1.apis.zonkafeedback.com`; format `uri`; Zonka
  Feedback API base URL for the account datacenter, such as https://us1.apis.zonkafeedback.com,
  https://e.apis.zonkafeedback.com, https://in.apis.zonkafeedback.com, or
  https://au.apis.zonkafeedback.com.
- `contact_id` (optional, string); Contact ID or external ID used by the contact_details stream.
- `device_id` (optional, string); Device ID used by the device_details, device_uptime, and
  device_responses streams.
- `device_stats_end_date` (optional, string); End date/time for device uptime and response stats, in
  the format expected by Zonka Feedback.
- `device_stats_start_date` (optional, string); Start date/time for device uptime and response
  stats, in the format expected by Zonka Feedback.
- `distribution_logs_channel` (optional, string); Distribution log channel filter, such as email,
  sms, or whatsapp.
- `distribution_logs_end_date` (optional, string); End date for distribution logs in YYYY-MM-DD
  format.
- `distribution_logs_start_date` (optional, string); Start date for distribution logs in YYYY-MM-DD
  format.
- `location_id` (optional, string); Location ID or external ID used by the location_details stream
  and optional log filtering.
- `response_id` (optional, string); Response ID used by the response_details stream.
- `survey_id` (optional, string); Survey ID used by the survey_details stream and distribution_logs
  stream.
- `user_id` (optional, string); User ID used by the user_details stream.

Secret fields are redacted in logs and write previews: `access_token`, `auth_token`.

Default configuration values: `base_url=https://us1.apis.zonkafeedback.com`.

Authentication behavior:

- API key authentication in `Z-API-TOKEN` using `secrets.auth_token` when `{{ secrets.auth_token
  }}`.
- API key authentication in `Z-API-TOKEN` using `secrets.access_token` when `{{ secrets.access_token
  }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/responses` with query `page`=`1`; `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `response_details`, `workspaces`, `survey_stats`, `survey_details`,
`contact_details`, `contact_segments`, `device_details`, `device_uptime`, `device_responses`,
`location_details`, `user_details`; page_number: `responses`, `surveys`, `contacts`, `devices`,
`tasks`, `locations`, `users`, `distribution_logs`.

- `responses`: GET `/responses` - records path `responses`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `surveys`: GET `/surveys` - records path `surveys`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100.
- `contacts`: GET `/contacts` - records path `contacts`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100.
- `response_details`: GET `/responses/{{ config.response_id }}` - records path `.`; emits
  passthrough records.
- `workspaces`: GET `/workspaces` - records path `.`; emits passthrough records.
- `survey_stats`: GET `/snapshot` - records path `.`; emits passthrough records.
- `survey_details`: GET `/surveys/{{ config.survey_id }}` - records path `.`; emits passthrough
  records.
- `contact_details`: GET `/contacts/{{ config.contact_id }}` - records path `.`; emits passthrough
  records.
- `contact_segments`: GET `/contactlist` - records path `.`; emits passthrough records.
- `devices`: GET `/devices` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `device_details`: GET `/devices/{{ config.device_id }}` - records path `.`; emits passthrough
  records.
- `device_uptime`: GET `/devices/uptimestat` - records path `.`; query `deviceId`=`{{
  config.device_id }}`; `endDate`=`{{ config.device_stats_end_date }}`; `startDate`=`{{
  config.device_stats_start_date }}`; emits passthrough records.
- `device_responses`: GET `/devices/responses` - records path `.`; query `deviceId`=`{{
  config.device_id }}`; `endDate`=`{{ config.device_stats_end_date }}`; `startDate`=`{{
  config.device_stats_start_date }}`; emits passthrough records.
- `tasks`: GET `/tasks` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `locations`: GET `/locations` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `location_details`: GET `/locations/{{ config.location_id }}` - records path `.`; emits
  passthrough records.
- `users`: GET `/users` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `user_details`: GET `/users/{{ config.user_id }}` - records path `.`; emits passthrough records.
- `distribution_logs`: GET `/distribution-logs` - records path `.`; query `channel`=`{{
  config.distribution_logs_channel }}`; `endDate`=`{{ config.distribution_logs_end_date }}`;
  `locationId` from template `{{ config.location_id }}`, omitted when absent; `startDate`=`{{
  config.distribution_logs_start_date }}`; `surveyId`=`{{ config.survey_id }}`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits
  passthrough records.

## Write actions & risks

Overall write risk: external Zonka Feedback API mutations that add or update responses, upsert
contacts, send surveys over email/SMS/WhatsApp, and create, update, or delete tasks.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `add_response`: POST `/responses` - kind `create`; body type `json`; required record fields
  `surveyId`, `response`; accepted fields `agent`, `attributes`, `channel`, `contact`, `deviceid`,
  `language`, `locationId`, `response`, `submitDateTime`, `surveyEndTime`, `surveyId`,
  `surveyStartTime`; risk: creates a Zonka Feedback survey response; approval required.
- `update_response`: PATCH `/responses/{{ record.responseId }}` - kind `update`; body type `json`;
  path fields `responseId`; required record fields `responseId`; accepted fields `agent`,
  `attributes`, `channel`, `contact`, `deviceid`, `language`, `locationId`, `response`,
  `responseId`, `submitDateTime`, `surveyEndTime`, `surveyStartTime`; risk: updates a Zonka Feedback
  survey response; approval required.
- `upsert_contacts`: POST `/contacts/upsert` - kind `upsert`; body type `json`; accepted fields
  `attributes`, `email`, `lists`, `mobile`, `name`, `uniqueId`; risk: creates or updates Zonka
  Feedback contact records; approval required.
- `send_email_survey`: POST `/sendemail` - kind `create`; body type `json`; required record fields
  `surveyId`, `email`; accepted fields `attributes`, `email`, `embedField`, `ignoreThrottling`,
  `list`, `locationId`, `message`, `name`, `reminder`, `reminderMaxCount`, `reminderMessage`,
  `reminderSendAfter`, `scheduleDateTime`, `signature`, `subject`, `surveyId`; risk: sends or
  schedules email survey invitations; approval required.
- `send_sms_survey`: POST `/sendsms` - kind `create`; body type `json`; required record fields
  `surveyId`, `mobile`; accepted fields `attributes`, `ignoreThrottling`, `list`, `locationId`,
  `message`, `mobile`, `name`, `scheduleDateTime`, `surveyId`; risk: sends or schedules SMS survey
  invitations; approval required.
- `send_two_way_sms_survey`: POST `/send2waysms` - kind `create`; body type `json`; required record
  fields `surveyId`, `mobile`; accepted fields `attributes`, `ignoreThrottling`, `list`,
  `locationId`, `message`, `mobile`, `name`, `scheduleDateTime`, `surveyId`; risk: sends or
  schedules two-way SMS survey invitations; approval required.
- `send_whatsapp_survey`: POST `/send-wa-message` - kind `create`; body type `json`; required record
  fields `surveyId`, `mobile`; accepted fields `attributes`, `from`, `ignoreThrottling`, `language`,
  `list`, `locationId`, `mobile`, `name`, `scheduleDateTime`, `surveyId`, `templateLanguage`,
  `templateName`; risk: sends or schedules WhatsApp survey invitations; approval required.
- `add_task`: POST `/tasks/add` - kind `create`; body type `json`; required record fields
  `taskName`; accepted fields `assignTo`, `contactId`, `description`, `dueDateTime`,
  `reminderSetting`, `responseId`, `taskName`, `typeName`; risk: creates a Zonka Feedback task;
  approval required.
- `update_task`: POST `/tasks/{{ record.taskId }}` - kind `update`; body type `json`; path fields
  `taskId`; required record fields `taskId`; accepted fields `assignTo`, `description`,
  `dueDateTime`, `reminderSetting`, `taskId`, `taskName`, `typeName`; risk: updates a Zonka Feedback
  task; approval required.
- `delete_tasks`: DELETE `/tasks/delete` - kind `delete`; body type `json`; required record fields
  `taskId`; accepted fields `taskId`; confirmation `destructive`; risk: deletes one or more Zonka
  Feedback tasks; approval required.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 19 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
