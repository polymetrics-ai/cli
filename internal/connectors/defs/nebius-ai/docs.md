# Overview

Reads and writes Nebius Token Factory OpenAI-compatible API resources, including models, files,
fine-tuning, datasets, operations, dedicated endpoints, and inference actions.

Readable streams: `models`, `files`, `batches`, `files_file_id`, `files_file_id_content`,
`files_file_id_link`, `fine_tuning_jobs`, `fine_tuning_jobs_job_id`,
`fine_tuning_jobs_job_id_events`, `fine_tuning_jobs_job_id_checkpoints`,
`fine_tuning_jobs_job_id_checkpoints_checkpoint_id`, `fine_tuning_models_spec_draft`,
`fine_tuning_models_spec_draft_2`, `dedicated_endpoints_templates`, `dedicated_endpoints`,
`datasets`, `datasets_dataset_id`, `datasets_dataset_id_query_templates`,
`datasets_dataset_id_content`, `datasets_dataset_id_export`, `datasets_uploads_upload_id_parts`,
`datasets_uploads_upload_id`, `operations`, `operations_operation_id`,
`operations_operation_id_results`, `operations_operation_id_errors`.

Write actions: `create_completions`, `create_chat_completions`, `create_embeddings`,
`create_rerank`, `create_responses`, `delete_files_file_id`, `create_images_generations`,
`create_fine_tuning_jobs`, `execute_fine_tuning_jobs_job_id_cancel`, `create_dedicated_endpoints`,
`update_dedicated_endpoints_endpoint_id`, `delete_dedicated_endpoints_endpoint_id`,
`create_datasets`, `update_datasets_dataset_id`, `delete_datasets_dataset_id`,
`create_datasets_uploads`, `create_datasets_uploads_upload_id_complete`,
`execute_datasets_uploads_upload_id_cancel`, `create_operations`,
`execute_operations_operation_id_cancel`.

Service API documentation: https://api.tokenfactory.nebius.com/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Nebius Token Factory API key. Sent as a Bearer token and
  never logged.
- `base_url` (optional, string); default `https://api.tokenfactory.nebius.com`; format `uri`; Nebius
  OpenAI-compatible API base URL override.
- `checkpoint_id` (optional, string); Path parameter checkpoint_id for the
  fine_tuning_jobs_job_id_checkpoints_checkpoint_id stream.
- `dataset_id` (optional, string); Path parameter dataset_id for the datasets_dataset_id stream.
- `file_id` (optional, string); Path parameter file_id for the files_file_id stream.
- `job_id` (optional, string); Path parameter job_id for the fine_tuning_jobs_job_id stream.
- `limit` (optional, integer); default `20`; Page size sent as the limit query parameter for list
  endpoints.
- `operation_id` (optional, string); Path parameter operation_id for the operations_operation_id
  stream.
- `upload_id` (optional, string); Path parameter upload_id for the datasets_uploads_upload_id_parts
  stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.tokenfactory.nebius.com`, `limit=20`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1/models`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `models`, `files`, `batches`, `fine_tuning_jobs`,
`fine_tuning_jobs_job_id_events`, `fine_tuning_jobs_job_id_checkpoints`, `datasets`,
`operations_operation_id_results`; none: `files_file_id`, `files_file_id_content`,
`files_file_id_link`, `fine_tuning_jobs_job_id`,
`fine_tuning_jobs_job_id_checkpoints_checkpoint_id`, `fine_tuning_models_spec_draft`,
`fine_tuning_models_spec_draft_2`, `dedicated_endpoints_templates`, `dedicated_endpoints`,
`datasets_dataset_id`, `datasets_dataset_id_query_templates`, `datasets_dataset_id_export`,
`datasets_uploads_upload_id_parts`, `datasets_uploads_upload_id`, `operations`,
`operations_operation_id`, `operations_operation_id_errors`; offset_limit:
`datasets_dataset_id_content`.

- `models`: GET `/v1/models` - records path `data`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `after`; next cursor from last record field `id`; stop flag
  `has_more`.
- `files`: GET `/v1/files` - records path `data`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `after`; next cursor from last record field `id`; stop flag
  `has_more`.
- `batches`: GET `/v1/batches` - records path `data`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `after`; next cursor from last record field `id`; stop flag
  `has_more`.
- `files_file_id`: GET `/v1/files/{{ config.file_id }}` - records at response root; emits
  passthrough records.
- `files_file_id_content`: GET `/v1/files/{{ config.file_id }}/content` - records at response root;
  emits passthrough records.
- `files_file_id_link`: GET `/v1/files/{{ config.file_id }}/link` - records at response root; emits
  passthrough records.
- `fine_tuning_jobs`: GET `/v1/fine_tuning/jobs` - records path `data`; query `limit`=`{{
  config.limit }}`; cursor pagination; cursor parameter `after`; next cursor from last record field
  `id`; stop flag `has_more`; emits passthrough records.
- `fine_tuning_jobs_job_id`: GET `/v1/fine_tuning/jobs/{{ config.job_id }}` - records at response
  root; emits passthrough records.
- `fine_tuning_jobs_job_id_events`: GET `/v1/fine_tuning/jobs/{{ config.job_id }}/events` - records
  path `data`; query `limit`=`{{ config.limit }}`; cursor pagination; cursor parameter `after`; next
  cursor from last record field `id`; stop flag `has_more`; emits passthrough records.
- `fine_tuning_jobs_job_id_checkpoints`: GET `/v1/fine_tuning/jobs/{{ config.job_id }}/checkpoints`
  - records path `data`; query `limit`=`{{ config.limit }}`; cursor pagination; cursor parameter
  `after`; next cursor from last record field `id`; stop flag `has_more`; emits passthrough records.
- `fine_tuning_jobs_job_id_checkpoints_checkpoint_id`: GET `/v1/fine_tuning/jobs/{{ config.job_id
  }}/checkpoints/{{ config.checkpoint_id }}` - records at response root; emits passthrough records.
- `fine_tuning_models_spec_draft`: GET `/v1/fine_tuning/models/spec-draft` - records path `models`;
  emits passthrough records.
- `fine_tuning_models_spec_draft_2`: GET `/fine_tuning/models/spec-draft` - records path `models`;
  emits passthrough records.
- `dedicated_endpoints_templates`: GET `/v0/dedicated_endpoints/templates` - records path `data`;
  emits passthrough records.
- `dedicated_endpoints`: GET `/v0/dedicated_endpoints` - records path `data`; emits passthrough
  records.
- `datasets`: GET `/v1/datasets` - records path `data`; query `limit`=`{{ config.limit }}`; cursor
  pagination; cursor parameter `after`; next cursor from last record field `id`; stop flag
  `has_more`; emits passthrough records.
- `datasets_dataset_id`: GET `/v1/datasets/{{ config.dataset_id }}` - records at response root;
  emits passthrough records.
- `datasets_dataset_id_query_templates`: GET `/v1/datasets/{{ config.dataset_id }}/query_templates`
  - records at response root; emits passthrough records.
- `datasets_dataset_id_content`: GET `/v1/datasets/{{ config.dataset_id }}/content` - records at
  response root; query `limit`=`{{ config.limit }}`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 20; emits passthrough records.
- `datasets_dataset_id_export`: GET `/v1/datasets/{{ config.dataset_id }}/export` - records at
  response root; query `format`=`jsonl`; `limit`=`{{ config.limit }}`; emits passthrough records.
- `datasets_uploads_upload_id_parts`: GET `/v1/datasets/uploads/{{ config.upload_id }}/parts` -
  records path `data`; emits passthrough records.
- `datasets_uploads_upload_id`: GET `/v1/datasets/uploads/{{ config.upload_id }}` - records at
  response root; emits passthrough records.
- `operations`: GET `/v1/operations` - records path `data`; query `limit`=`{{ config.limit }}`;
  emits passthrough records.
- `operations_operation_id`: GET `/v1/operations/{{ config.operation_id }}` - records at response
  root; emits passthrough records.
- `operations_operation_id_results`: GET `/v1/operations/{{ config.operation_id }}/results` -
  records path `data`; query `limit`=`{{ config.limit }}`; cursor pagination; cursor parameter
  `after`; next cursor from last record field `id`; stop flag `has_more`; emits passthrough records.
- `operations_operation_id_errors`: GET `/v1/operations/{{ config.operation_id }}/errors` - records
  at response root; emits passthrough records.

## Write actions & risks

Overall write risk: external Nebius API writes and paid inference/operation side effects.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_completions`: POST `/v1/completions` - kind `create`; body type `json`; required record
  fields `model`, `prompt`; accepted fields `echo`, `extra_body`, `frequency_penalty`, `logit_bias`,
  `logprobs`, `max_tokens`, `model`, `n`, `presence_penalty`, `prompt`, `service_tier`, `stop`,
  `stream`, `stream_options`, `temperature`, `top_p`, `user`; risk: high: external Nebius API side
  effect or mutation; approval required.
- `create_chat_completions`: POST `/v1/chat/completions` - kind `create`; body type `json`; required
  record fields `model`, `messages`; accepted fields `extra_body`, `frequency_penalty`,
  `logit_bias`, `logprobs`, `max_completion_tokens`, `max_tokens`, `messages`, `model`, `n`,
  `presence_penalty`, `reasoning_effort`, `response_format`, `service_tier`, `stop`, `store`,
  `stream`, `stream_options`, `temperature`, and 5 more; risk: high: external Nebius API side effect
  or mutation; approval required.
- `create_embeddings`: POST `/v1/embeddings` - kind `create`; body type `json`; required record
  fields `model`, `input`; accepted fields `dimensions`, `encoding_format`, `input`, `model`,
  `service_tier`, `user`; risk: high: external Nebius API side effect or mutation; approval
  required.
- `create_rerank`: POST `/v1/rerank` - kind `create`; body type `json`; required record fields
  `model`, `query`, `documents`; accepted fields `documents`, `model`, `query`, `service_tier`,
  `user`; risk: high: external Nebius API side effect or mutation; approval required.
- `create_responses`: POST `/v1/responses` - kind `create`; body type `json`; required record fields
  `input`, `model`; accepted fields `background`, `include`, `input`, `instructions`,
  `max_output_tokens`, `max_tool_calls`, `metadata`, `model`, `parallel_tool_calls`,
  `previous_response_id`, `prompt`, `prompt_cache_key`, `reasoning`, `service_tier`, `store`,
  `stream`, `temperature`, `text`, and 6 more; risk: high: external Nebius API side effect or
  mutation; approval required.
- `delete_files_file_id`: DELETE `/v1/files/{{ record.file_id }}` - kind `delete`; body type `none`;
  path fields `file_id`; required record fields `file_id`; accepted fields `file_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: medium: external
  Nebius API side effect or mutation; approval required.
- `create_images_generations`: POST `/v1/images/generations` - kind `create`; body type `json`;
  required record fields `model`, `prompt`; accepted fields `guidance_scale`, `height`, `loras`,
  `model`, `negative_prompt`, `num_inference_steps`, `prompt`, `response_extension`,
  `response_format`, `seed`, `width`; risk: high: external Nebius API side effect or mutation;
  approval required.
- `create_fine_tuning_jobs`: POST `/v1/fine_tuning/jobs` - kind `create`; body type `json`; required
  record fields `model`, `training_file`; accepted fields `extra_body`, `from_checkpoint`,
  `hyperparameters`, `integrations`, `method`, `model`, `seed`, `suffix`, `tags`, `training_file`,
  `validation_file`; risk: high: external Nebius API side effect or mutation; approval required.
- `execute_fine_tuning_jobs_job_id_cancel`: POST `/v1/fine_tuning/jobs/{{ record.job_id }}/cancel` -
  kind `custom`; body type `none`; path fields `job_id`; required record fields `job_id`; accepted
  fields `job_id`; risk: high: external Nebius API side effect or mutation; approval required.
- `create_dedicated_endpoints`: POST `/v0/dedicated_endpoints` - kind `create`; body type `json`;
  required record fields `name`, `model_name`, `flavor_name`, `gpu_type`, `region`, `gpu_count`,
  `scaling`; accepted fields `custom_weights_id`, `description`, `flavor_name`, `gpu_count`,
  `gpu_type`, `model_name`, `name`, `region`, `scaling`; risk: high: external Nebius API side effect
  or mutation; approval required.
- `update_dedicated_endpoints_endpoint_id`: PATCH `/v0/dedicated_endpoints/{{ record.endpoint_id }}`
  - kind `update`; body type `json`; path fields `endpoint_id`; required record fields
  `endpoint_id`; accepted fields `custom_weights_id`, `description`, `enabled`, `endpoint_id`,
  `gpu_count`, `gpu_type`, `name`, `scaling`; risk: high: external Nebius API side effect or
  mutation; approval required.
- `delete_dedicated_endpoints_endpoint_id`: DELETE `/v0/dedicated_endpoints/{{ record.endpoint_id
  }}` - kind `delete`; body type `none`; path fields `endpoint_id`; required record fields
  `endpoint_id`; accepted fields `endpoint_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Nebius API side effect or mutation; approval
  required.
- `create_datasets`: POST `/v1/datasets` - kind `create`; body type `json`; required record fields
  `name`, `schema`, `folder`, `rows`; accepted fields `ai_project_id`, `folder`, `name`, `rows`,
  `schema`; risk: high: external Nebius API side effect or mutation; approval required.
- `update_datasets_dataset_id`: PATCH `/v1/datasets/{{ record.dataset_id }}` - kind `update`; body
  type `json`; path fields `dataset_id`; required record fields `dataset_id`; accepted fields
  `dataset_id`, `folder`, `name`; risk: high: external Nebius API side effect or mutation; approval
  required.
- `delete_datasets_dataset_id`: DELETE `/v1/datasets/{{ record.dataset_id }}` - kind `delete`; body
  type `none`; path fields `dataset_id`; required record fields `dataset_id`; accepted fields
  `dataset_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Nebius API side effect or mutation; approval required.
- `create_datasets_uploads`: POST `/v1/datasets/uploads` - kind `create`; body type `json`; required
  record fields `name`, `schema`, `folder`; accepted fields `folder`, `name`, `schema`; risk: high:
  external Nebius API side effect or mutation; approval required.
- `create_datasets_uploads_upload_id_complete`: POST `/v1/datasets/uploads/{{ record.upload_id
  }}/complete` - kind `create`; body type `json`; path fields `upload_id`; required record fields
  `upload_id`, `part_ids`; accepted fields `part_ids`, `upload_id`; risk: high: external Nebius API
  side effect or mutation; approval required.
- `execute_datasets_uploads_upload_id_cancel`: POST `/v1/datasets/uploads/{{ record.upload_id
  }}/cancel` - kind `custom`; body type `none`; path fields `upload_id`; required record fields
  `upload_id`; accepted fields `upload_id`; risk: high: external Nebius API side effect or mutation;
  approval required.
- `create_operations`: POST `/v1/operations` - kind `create`; body type `json`; required record
  fields `params`, `src`; accepted fields `ai_project_id`, `dst`, `params`, `src`, `type`; risk:
  high: external Nebius API side effect or mutation; approval required.
- `execute_operations_operation_id_cancel`: POST `/v1/operations/{{ record.operation_id }}/cancel` -
  kind `custom`; body type `none`; path fields `operation_id`; required record fields
  `operation_id`; accepted fields `operation_id`; risk: high: external Nebius API side effect or
  mutation; approval required.

## Known limits

- Batch defaults: read_page_size=20, write_batch_size=1.
- API coverage includes 26 stream-backed endpoint group(s), 20 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=3.
