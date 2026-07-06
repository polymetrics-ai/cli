# Overview

Reads Formbricks surveys, responses, contacts, contact attributes, action classes, webhooks, and
account metadata; writes approved management API mutations.

Readable streams: `surveys`, `survey_details`, `responses`, `response_details`, `action_classes`,
`action_class_details`, `attribute_classes`, `contact_attribute_keys`,
`contact_attribute_key_details`, `contact_attributes`, `contacts`, `contact_details`, `me`,
`webhooks`, `webhook_details`.

Write actions: `create_action_class`, `delete_action_class`, `create_response`, `update_response`,
`delete_response`, `create_public_file_upload`, `create_survey`, `update_survey`, `delete_survey`,
`create_webhook`, `delete_webhook`.

Service API documentation: https://formbricks.com/docs/api-reference/rest-api.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Formbricks management API key, sent as the X-API-Key request
  header.
- `base_url` (optional, string); default `https://app.formbricks.com/api/v1`; format `uri`;
  Formbricks management API base URL. Defaults to the hosted app.formbricks.com instance; override
  for a self-hosted Formbricks deployment.
- `response_ids` (optional, string); Optional comma-separated response ids used by the
  response_details stream.
- `survey_id` (optional, string); Optional survey id used to filter the responses stream with
  Formbricks' surveyId query parameter.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.formbricks.com/api/v1`.

Authentication behavior:

- API key authentication in `X-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `management/surveys`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `surveys`, `survey_details`, `response_details`, `action_classes`,
`action_class_details`, `attribute_classes`, `contact_attribute_keys`,
`contact_attribute_key_details`, `contact_attributes`, `contacts`, `contact_details`, `me`,
`webhooks`, `webhook_details`; offset_limit: `responses`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `surveys`: GET `management/surveys` - records path `data`; incremental cursor `updated_at`;
  formatted as `rfc3339`; computed output fields `created_at`, `environment_id`, `updated_at`.
- `survey_details`: GET `management/surveys/{{ fanout.id }}` - single-object response; records path
  `data`; computed output fields `created_at`, `created_by`, `display_option`, `updated_at`,
  `workspace_id`; fan-out; ids from request `management/surveys`; id-list records path `data`; id
  field `id`; id inserted into the request path.
- `responses`: GET `management/responses` - records path `data`; query `surveyId` from template `{{
  config.survey_id }}`, omitted when absent; offset/limit pagination; offset parameter `skip`; limit
  parameter `limit`; page size 50; incremental cursor `updated_at`; formatted as `rfc3339`; computed
  output fields `contact_id`, `created_at`, `survey_id`, `updated_at`.
- `response_details`: GET `management/responses/{{ fanout.id }}` - single-object response; records
  path `data`; computed output fields `contact_id`, `created_at`, `survey_id`, `updated_at`;
  fan-out; ids from config field `response_ids`; id inserted into the request path.
- `action_classes`: GET `management/action-classes` - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `created_at`, `environment_id`,
  `updated_at`.
- `action_class_details`: GET `management/action-classes/{{ fanout.id }}` - single-object response;
  records path `data`; computed output fields `created_at`, `updated_at`, `workspace_id`; fan-out;
  ids from request `management/action-classes`; id-list records path `data`; id field `id`; id
  inserted into the request path.
- `attribute_classes`: GET `management/attribute-classes` - records path `data`; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `created_at`, `environment_id`,
  `updated_at`.
- `contact_attribute_keys`: GET `management/contact-attribute-keys` - records path `data`; computed
  output fields `created_at`, `is_unique`, `updated_at`, `workspace_id`.
- `contact_attribute_key_details`: GET `management/contact-attribute-keys/{{ fanout.id }}` -
  single-object response; records path `data`; computed output fields `created_at`, `is_unique`,
  `updated_at`, `workspace_id`; fan-out; ids from request `management/contact-attribute-keys`;
  id-list records path `data`; id field `id`; id inserted into the request path.
- `contact_attributes`: GET `management/contact-attributes` - records path `data`; computed output
  fields `attribute_key_id`, `contact_id`, `created_at`, `updated_at`.
- `contacts`: GET `management/contacts` - records path `data`; computed output fields `created_at`,
  `updated_at`, `user_id`, `workspace_id`.
- `contact_details`: GET `management/contacts/{{ fanout.id }}` - single-object response; records
  path `data`; computed output fields `created_at`, `updated_at`, `user_id`, `workspace_id`;
  fan-out; ids from request `management/contacts`; id-list records path `data`; id field `id`; id
  inserted into the request path.
- `me`: GET `management/me` - single-object response; records path `.`; computed output fields
  `app_setup_completed`, `created_at`, `environment_permissions`, `id`, `organization_access`,
  `organization_id`, `updated_at`, `website_setup_completed`.
- `webhooks`: GET `webhooks` - records path `data`; computed output fields `created_at`,
  `environment_id`, `updated_at`.
- `webhook_details`: GET `webhooks/{{ fanout.id }}` - single-object response; records path `data`;
  computed output fields `created_at`, `updated_at`, `workspace_id`; fan-out; ids from request
  `webhooks`; id-list records path `data`; id field `id`; id inserted into the request path.

## Write actions & risks

Overall write risk: external Formbricks management API mutations for action classes, responses,
public upload URLs, surveys, and webhooks.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_action_class`: POST `management/action-classes` - kind `create`; body type `json`;
  required record fields `workspaceId`, `name`, `type`; accepted fields `description`, `key`,
  `name`, `noCodeConfig`, `type`, `workspaceId`; risk: creates an action class in the configured
  Formbricks workspace.
- `delete_action_class`: DELETE `management/action-classes/{{ record.actionClassId }}` - kind
  `delete`; body type `none`; path fields `actionClassId`; required record fields `actionClassId`;
  accepted fields `actionClassId`; confirmation `destructive`; risk: deletes an action class;
  automatic action classes may be rejected by Formbricks.
- `create_response`: POST `management/responses` - kind `create`; body type `json`; required record
  fields `surveyId`; accepted fields `createdAt`, `data`, `finished`, `language`, `surveyId`,
  `updatedAt`; risk: creates a survey response and may trigger configured response pipelines.
- `update_response`: PUT `management/responses/{{ record.responseId }}` - kind `update`; body type
  `json`; path fields `responseId`; required record fields `responseId`; accepted fields `data`,
  `finished`, `language`, `responseId`; risk: updates a survey response and may trigger configured
  response pipelines.
- `delete_response`: DELETE `management/responses/{{ record.responseId }}` - kind `delete`; body
  type `none`; path fields `responseId`; required record fields `responseId`; accepted fields
  `responseId`; confirmation `destructive`; risk: deletes a survey response.
- `create_public_file_upload`: POST `management/storage` - kind `create`; body type `json`; required
  record fields `fileName`, `fileType`, `workspaceId`; accepted fields `allowedFileExtensions`,
  `fileName`, `fileType`, `workspaceId`; risk: creates a public file upload target and returns
  upload metadata.
- `create_survey`: POST `management/surveys` - kind `create`; body type `json`; required record
  fields `workspaceId`, `name`, `type`, `status`; accepted fields `displayOption`, `languages`,
  `name`, `questions`, `status`, `type`, `workspaceId`; risk: creates a survey in the configured
  Formbricks workspace.
- `update_survey`: PUT `management/surveys/{{ record.surveyId }}` - kind `update`; body type `json`;
  path fields `surveyId`; required record fields `surveyId`; accepted fields `displayOption`,
  `languages`, `name`, `questions`, `status`, `surveyId`, `type`; risk: updates an existing survey.
- `delete_survey`: DELETE `management/surveys/{{ record.surveyId }}` - kind `delete`; body type
  `none`; path fields `surveyId`; required record fields `surveyId`; accepted fields `surveyId`;
  confirmation `destructive`; risk: deletes a survey and its configured collection surface.
- `create_webhook`: POST `webhooks` - kind `create`; body type `json`; required record fields `url`,
  `triggers`; accepted fields `name`, `surveyIds`, `triggers`, `url`; risk: creates a webhook that
  sends Formbricks events to the configured URL.
- `delete_webhook`: DELETE `webhooks/{{ record.webhookId }}` - kind `delete`; body type `none`; path
  fields `webhookId`; required record fields `webhookId`; accepted fields `webhookId`; confirmation
  `destructive`; risk: deletes a webhook and stops future deliveries.

## Known limits

- Batch defaults: read_page_size=50, write_batch_size=1.
- API coverage includes 15 stream-backed endpoint group(s), 11 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, non_data_endpoint=2, out_of_scope=5.
