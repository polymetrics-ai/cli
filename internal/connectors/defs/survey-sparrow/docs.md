# Overview

Reads and manages SurveySparrow surveys, contacts, responses, questions, channels, contact
lists/properties, reminders, reputation platforms/reviews, survey folders, ticket fields, tickets,
teams, roles, variables, webhooks, users, templates, email themes, and expressions through the
SurveySparrow API.

Readable streams: `surveys`, `contacts`, `responses`, `questions`, `channels`, `contact_lists`,
`contact_properties`, `reminders`, `reputation_platforms`, `reputation_app_platforms`,
`reputation_reviews`, `survey_folders`, `ticket_fields`, `tickets`, `teams`, `roles`, `variables`,
`webhooks`, `users`, `templates`, `email_themes`, `expressions`.

Write actions: `create_survey`, `update_survey`, `create_contact`, `update_contact`,
`delete_contact`, `create_question`, `update_question`, `delete_question`, `create_contact_list`,
`update_contact_list`, `delete_contact_list`, `create_contact_property`, `update_contact_property`,
`delete_contact_property`, `create_survey_folder`, `update_survey_folder`, `delete_survey_folder`,
`create_team`, `create_ticket`, `update_ticket`, `delete_ticket`, `create_webhook`,
`update_webhook`, `delete_webhook`, `create_user`, `update_user`, `delete_user`, `create_reminder`,
`delete_reminder`, `create_variable`, `delete_variable`, `create_channel`, `delete_channel`.

Service API documentation: https://developers.surveysparrow.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); SurveySparrow access token, sent as a Bearer token
  (Authorization: Bearer <access_token>). Never logged.
- `base_url` (optional, string); default `https://api.surveysparrow.com/v3`; format `uri`;
  SurveySparrow API base URL override for tests or proxies.
- `survey_id` (optional, string); SurveySparrow survey ID that the 'questions' stream is scoped to
  (required for that stream; sent as the required survey_id query parameter on GET /v3/questions per
  the real documented API -- NOT a path segment). Not required for any other stream.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.surveysparrow.com/v3`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/surveys` with query `limit`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `surveys`: GET `/surveys` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100.
- `contacts`: GET `/contacts` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 50.
- `responses`: GET `/responses` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; incremental cursor `completed_time`;
  formatted as `rfc3339`.
- `questions`: GET `/questions` - records path `data`; query `survey_id`=`{{ config.survey_id }}`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `channels`: GET `/channels` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100.
- `contact_lists`: GET `/contact_lists` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `contact_properties`: GET `/contact_properties` - records path `data`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `reminders`: GET `/reminders` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100.
- `reputation_platforms`: GET `/reputation/platforms` - records path `data`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `reputation_app_platforms`: GET `/reputation/app_platforms` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 50.
- `reputation_reviews`: GET `/reputation/reviews` - records path `data`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 50.
- `survey_folders`: GET `/survey_folders` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `ticket_fields`: GET `/ticket_fields` - records path `data`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `tickets`: GET `/tickets` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100.
- `teams`: GET `/teams` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100.
- `roles`: GET `/roles` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100.
- `variables`: GET `/variables` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100.
- `webhooks`: GET `/webhooks` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100.
- `users`: GET `/users` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 50.
- `templates`: GET `/templates` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100.
- `email_themes`: GET `/email_themes` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100.
- `expressions`: GET `/expressions` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100.

## Write actions & risks

Overall write risk: external mutation of SurveySparrow surveys, contacts, questions, contact
lists/properties, survey folders, teams, tickets, webhooks, users, reminders, variables, and
channels, including irreversible deletes and live-user-account creation/deletion.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_survey`: POST `/surveys` - kind `create`; body type `json`; required record fields `name`,
  `survey_type`; accepted fields `description`, `name`, `survey_folder_id`, `survey_type`,
  `theme_id`, `visibility`; risk: external mutation; approval required.
- `update_survey`: PATCH `/surveys/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `description`, `id`, `name`, `theme_id`; risk:
  external mutation; approval required.
- `create_contact`: POST `/contacts` - kind `create`; body type `json`; accepted fields
  `contact_type`, `email`, `full_name`, `job_title`, `mobile`, `phone`, `referenceId`, `unique_id`;
  risk: external mutation; approval required.
- `update_contact`: PUT `/contacts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `email`, `full_name`, `id`, `job_title`,
  `mobile`, `phone`; risk: external mutation; approval required.
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion; approval required.
- `create_question`: POST `/questions` - kind `create`; body type `json`; required record fields
  `survey_id`, `text`, `type`; accepted fields `description`, `multiple_answers`, `required`,
  `section_id`, `survey_id`, `text`, `type`; risk: external mutation; approval required.
- `update_question`: PUT `/questions/{{ record.question_id }}` - kind `update`; body type `json`;
  path fields `question_id`; required record fields `question_id`, `survey_id`; accepted fields
  `description`, `question_id`, `required`, `survey_id`, `text`; risk: external mutation; approval
  required.
- `delete_question`: DELETE `/questions/{{ record.question_id }}` - kind `delete`; body type `none`;
  path fields `question_id`; required record fields `question_id`; accepted fields `question_id`;
  missing records treated as success for status `404`; risk: irreversible external deletion;
  approval required.
- `create_contact_list`: POST `/contact_lists` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `description`, `name`; risk: external mutation; approval required.
- `update_contact_list`: PATCH `/contact_lists/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `description`, `id`, `name`; risk:
  external mutation; approval required.
- `delete_contact_list`: DELETE `/contact_lists/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: irreversible external deletion; approval required.
- `create_contact_property`: POST `/contact_properties` - kind `create`; body type `json`; required
  record fields `type`, `label`; accepted fields `contact_property_group_id`, `description`,
  `label`, `type`; risk: external mutation; approval required.
- `update_contact_property`: PATCH `/contact_properties/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `description`, `id`,
  `label`; risk: external mutation; approval required.
- `delete_contact_property`: DELETE `/contact_properties/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: irreversible external deletion; approval required.
- `create_survey_folder`: POST `/survey_folders` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `name`, `parent_survey_folder_id`, `visibility`; risk: external
  mutation; approval required.
- `update_survey_folder`: PATCH `/survey_folders/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`, `name`, `visibility`; risk:
  external mutation; approval required.
- `delete_survey_folder`: DELETE `/survey_folders/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: irreversible external deletion; approval required.
- `create_team`: POST `/teams` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `enable_round_robin`, `name`, `type`; risk: external mutation; approval required.
- `create_ticket`: POST `/tickets` - kind `create`; body type `json`; required record fields
  `subject`, `priority`, `status`; accepted fields `assignee_id`, `description`, `email`,
  `priority`, `requester_id`, `status`, `subject`, `team_id`; risk: external mutation; approval
  required.
- `update_ticket`: PUT `/tickets/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `assignee_id`, `id`, `priority`, `status`,
  `team_id`; risk: external mutation; approval required.
- `delete_ticket`: DELETE `/tickets/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: irreversible external deletion; approval required.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `url`, `survey_id`, `http_method`; accepted fields `description`, `event_type`, `http_method`,
  `name`, `object_type`, `survey_id`, `url`; risk: external mutation; approval required.
- `update_webhook`: PUT `/webhooks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `http_method`, `id`, `name`, `url`; risk:
  external mutation; approval required.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion; approval required.
- `create_user`: POST `/users` - kind `create`; body type `json`; required record fields `name`,
  `email`, `role_id`; accepted fields `email`, `name`, `role_id`; risk: external mutation creating a
  live user account with console access; approval required.
- `update_user`: PATCH `/users/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `id`, `name`, `role_id`; risk: external mutation;
  approval required.
- `delete_user`: DELETE `/users/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: irreversible external deletion of a user account; approval required.
- `create_reminder`: POST `/reminders` - kind `create`; body type `json`; required record fields
  `channel_id`, `survey_id`, `frequency`, `type`, `interval`, `embed_first_question`,
  `custom_footer`; accepted fields `body`, `channel_id`, `custom_footer`, `embed_first_question`,
  `frequency`, `interval`, `subject`, `survey_id`, `type`; risk: external mutation; approval
  required.
- `delete_reminder`: DELETE `/reminders/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion; approval required.
- `create_variable`: POST `/variables` - kind `create`; body type `json`; required record fields
  `survey_id`, `label`, `name`, `type`; accepted fields `description`, `label`, `name`, `survey_id`,
  `type`; risk: external mutation; approval required.
- `delete_variable`: DELETE `/variables/{{ record.variable_id }}` - kind `delete`; body type `none`;
  path fields `variable_id`; required record fields `variable_id`; accepted fields `variable_id`;
  missing records treated as success for status `404`; risk: irreversible external deletion;
  approval required.
- `create_channel`: POST `/channels` - kind `create`; body type `json`; required record fields
  `type`; accepted fields `email`, `mobile`, `mode`, `type`; risk: external mutation; approval
  required.
- `delete_channel`: DELETE `/channels/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 22 stream-backed endpoint group(s), 33 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=3, duplicate_of=13, out_of_scope=42.
