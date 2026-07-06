# Overview

Reads Clarifai applications, datasets, models, model versions, and workflows, and writes
application/dataset lifecycle mutations, through the Clarifai v2 REST API.

Readable streams: `applications`, `datasets`, `models`, `model_versions`, `workflows`.

Write actions: `create_application`, `update_application`, `create_dataset`, `delete_dataset`.

Service API documentation: https://docs.clarifai.com/api-guide/api-overview.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Clarifai personal access token (PAT). Sent as
  'Authorization: Key <api_key>'; never logged.
- `app_id` (optional, string); Clarifai app id. Required for the create_dataset/delete_dataset write
  actions, which are scoped under users/{user_id}/apps/{app_id}/datasets.
- `base_url` (optional, string); default `https://api.clarifai.com/v2`; format `uri`; Clarifai API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `user_id` (required, string); Clarifai user id all streams are scoped under (users/{user_id}/...).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.clarifai.com/v2`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users/{{ config.user_id }}/apps`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

- `applications`: GET `/users/{{ config.user_id }}/apps` - records path `apps`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `datasets`: GET `/users/{{ config.user_id }}/datasets` - records path `datasets`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `models`: GET `/users/{{ config.user_id }}/models` - records path `models`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.
- `model_versions`: GET `/users/{{ config.user_id }}/models/versions` - records path
  `model_versions`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100.
- `workflows`: GET `/users/{{ config.user_id }}/workflows` - records path `workflows`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100.

## Write actions & risks

Overall write risk: external mutation of Clarifai applications and datasets; delete_dataset is
irreversible (deletes the dataset and all its inputs/annotations) and update_application's
action=overwrite can replace application settings wholesale; every write ships with an explicit
per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_application`: POST `/users/{{ config.user_id }}/apps` - kind `create`; body type `json`;
  body fields `apps`; required record fields `apps`; accepted fields `apps`; risk: creates a new
  Clarifai application (workspace for datasets/models/workflows); low-risk (additive, no data loss).
- `update_application`: PATCH `/users/{{ config.user_id }}/apps` - kind `update`; body type `json`;
  body fields `action`, `apps`; required record fields `action`, `apps`; accepted fields `action`,
  `apps`; risk: updates an existing Clarifai application's settings (description, default workflow,
  notes); action=overwrite fully replaces the named fields rather than merging, so review the action
  value before use; approval required.
- `create_dataset`: POST `/users/{{ config.user_id }}/apps/{{ config.app_id }}/datasets` - kind
  `create`; body type `json`; body fields `datasets`; required record fields `datasets`; accepted
  fields `datasets`; risk: creates a new Clarifai dataset within the configured app; low-risk
  (additive, no data loss).
- `delete_dataset`: DELETE `/users/{{ config.user_id }}/apps/{{ config.app_id }}/datasets` - kind
  `delete`; body type `none`; body fields `dataset_ids`; required record fields `dataset_ids`;
  accepted fields `dataset_ids`; missing records treated as success for status `404`; risk:
  permanently deletes one or more Clarifai datasets and their inputs/annotations within the
  configured app; irreversible; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s), 4 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, duplicate_of=4, non_data_endpoint=4, out_of_scope=9,
  requires_elevated_scope=2.
