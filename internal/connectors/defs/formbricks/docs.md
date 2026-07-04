# Overview

Formbricks is an open-source survey and experience-management platform. This
bundle targets the Formbricks API v1 Management API documented at
`https://formbricks.com/docs/api-reference/rest-api`, using
`https://app.formbricks.com/api/v1` by default.

Pass B expands the legacy parity bundle beyond surveys, responses, action
classes, attribute classes, and webhooks. It now also reads management account
metadata, contact attribute keys, contact attributes, contacts, and detail
resources, and exposes approved JSON write actions for action classes,
responses, public file upload metadata, surveys, and webhooks.

## Auth setup

Formbricks Management API requests authenticate with an `X-API-Key` header from
the `api_key` secret. `base_url` defaults to the hosted Formbricks API v1 base
URL and can be overridden for self-hosted deployments. The optional `survey_id`
config narrows the `responses` stream through the documented `surveyId` query
parameter. The optional comma-separated `response_ids` config drives the
`response_details` stream because the response detail endpoint needs explicit
response ids.

## Streams notes

Legacy parity streams keep their original projected record shape:
`surveys`, `responses`, `action_classes`, `attribute_classes`, and `webhooks`.
Those streams intentionally retain the same field set as
`internal/connectors/formbricks` instead of passing through every raw API field.
CamelCase raw fields such as `environmentId`, `surveyId`, `contactId`,
`createdAt`, and `updatedAt` are mapped with `computed_fields`.

The current Formbricks v1 docs no longer list the legacy
`management/attribute-classes` endpoint; they expose contact attribute keys and
contact attributes instead. The `attribute_classes` stream remains for legacy
parity, while `contact_attribute_keys`, `contact_attribute_key_details`, and
`contact_attributes` cover the current documented contact attribute surface.

Detail streams use `fan_out` where a safe unpaginated list endpoint provides
ids: surveys, action classes, contact attribute keys, contacts, and webhooks.
`response_details` uses configured `response_ids` instead, because the response
list endpoint is offset-paginated and the current fan-out dialect cannot assign
separate pagination behavior to the id-list request and the child detail
request.

Only `responses` paginates, using documented `limit` and `skip` offset
pagination with the legacy default page size of 50. No stream declares an
incremental request parameter because legacy published cursor metadata only and
did not send server-side incremental filters.

## Write actions & risks

Write actions are intentionally resource-specific:
`create_action_class`, `delete_action_class`, `create_response`,
`update_response`, `delete_response`, `create_public_file_upload`,
`create_survey`, `update_survey`, `delete_survey`, `create_webhook`, and
`delete_webhook`.

All writes use Formbricks JSON request bodies except deletes, which send no
body. Delete actions are marked destructive and require the normal reverse ETL
plan, preview, approval token, and execute flow. Response writes can trigger
Formbricks response pipelines. Webhook creation starts future event deliveries
to the configured URL. `create_public_file_upload` requests public upload
metadata; it does not upload binary file bytes.

## Known limits

- The API v1 Public Client API endpoints are documented on the same REST API
  page but are excluded from this management API connector. They use a client
  workspace path model and are intended for SDK event/response submission, not
  backend management API-key sync.
- `GET /health` is excluded as an operational health check.
- `GET /management/surveys/{surveyId}/singleUseIds` is excluded because it
  generates new single-use survey links from a GET request. Treating it as a
  read stream would introduce side effects.
- The old API-key test helper `GET /api/v1/me` is excluded as a duplicate of
  the current documented `GET /api/v1/management/me` stream.
- `page_size` and `max_pages` config overrides from legacy are not modeled.
  The declarative pagination spec uses the legacy default page size of 50 and
  no max-page cap, matching legacy defaults.
