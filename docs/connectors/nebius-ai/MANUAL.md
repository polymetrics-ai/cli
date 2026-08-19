# pm connectors inspect nebius-ai

```text
NAME
  pm connectors inspect nebius-ai - Nebius AI connector manual

SYNOPSIS
  pm connectors inspect nebius-ai
  pm connectors inspect nebius-ai --json
  pm credentials add <name> --connector nebius-ai [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Nebius Token Factory OpenAI-compatible API resources, including models, files, fine-tuning, datasets, operations, dedicated endpoints, and inference actions.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  checkpoint_id
  dataset_id
  file_id
  job_id
  limit
  operation_id
  upload_id
  api_key (secret) (required)

ETL STREAMS
  models:
    primary key: id
    cursor: created
    fields: created(integer), id(string), object(string), owned_by(string)
  files:
    primary key: id
    cursor: created_at
    fields: bytes(integer), created_at(integer), filename(string), id(string), object(string), purpose(string), status(string)
  batches:
    primary key: id
    cursor: created_at
    fields: completed_at(integer), created_at(integer), endpoint(string), error_file_id(string), id(string), input_file_id(string), object(string), output_file_id(string), status(string)
  files_file_id:
    primary key: id
    fields: bytes(integer), created_at(integer), filename(string), id(string), object(string), purpose(string), status(string), status_details(string)
  files_file_id_content:
  files_file_id_link:
    fields: url(string)
  fine_tuning_jobs:
    primary key: id
    fields: created_at(integer), error(object), estimated_finish(integer), finished_at(integer), from_checkpoint(object), hyperparameters(object), id(string), integrations(array), method(object), model(string), object(string), organization_id(string), result_files(array), seed(integer), status(string), suffix(string), total_steps(integer), trained_steps(integer), trained_tokens(integer), training_file(string), validation_file(string)
  fine_tuning_jobs_job_id:
    primary key: id
    fields: created_at(integer), error(object), estimated_finish(integer), finished_at(integer), from_checkpoint(object), hyperparameters(object), id(string), integrations(array), method(object), model(string), object(string), organization_id(string), result_files(array), seed(integer), status(string), suffix(string), total_steps(integer), trained_steps(integer), trained_tokens(integer), training_file(string), validation_file(string)
  fine_tuning_jobs_job_id_events:
    primary key: id
    fields: created_at(integer), data(object), id(string), level(string), message(string), object(string), type(string)
  fine_tuning_jobs_job_id_checkpoints:
    primary key: id
    fields: created_at(integer), fine_tuned_model_checkpoint(string), fine_tuning_job_id(string), id(string), metrics(object), object(string), result_files(array), step_number(integer)
  fine_tuning_jobs_job_id_checkpoints_checkpoint_id:
    primary key: id
    fields: created_at(integer), fine_tuned_model_checkpoint(string), fine_tuning_job_id(string), id(string), metrics(object), object(string), step_number(integer)
  fine_tuning_models_spec_draft:
    fields: hf_repo_name(string), price(object)
  fine_tuning_models_spec_draft_2:
    fields: hf_repo_name(string), price(object)
  dedicated_endpoints_templates:
    fields: flavors(object), metadata(object), name(string), type(string)
  dedicated_endpoints:
    primary key: id
    fields: created_at(string), custom_weights_id(string), deployment(object), description(string), enabled(boolean), flavor_name(string), gpu_count(integer), gpu_type(string), id(string), model_name(string), name(string), region(string), routing_key(string), scaling(object)
  datasets:
    primary key: id
    fields: ai_project_id(string), created_at(integer), current_version(string), current_version_origin(object), error(string), folder(string), id(string), metadata(object), name(string), schema(array), status(string), type(string)
  datasets_dataset_id:
    primary key: id
    fields: ai_project_id(string), created_at(integer), current_version(string), current_version_origin(object), error(string), folder(string), id(string), metadata(object), name(string), schema(array), status(string), type(string)
  datasets_dataset_id_query_templates:
    fields: yql(string)
  datasets_dataset_id_content:
  datasets_dataset_id_export:
  datasets_uploads_upload_id_parts:
    primary key: id
    fields: created_at(integer), id(string), status(string), upload_id(string)
  datasets_uploads_upload_id:
    primary key: id
    fields: id(string), upload_info(object)
  operations:
    primary key: id
    fields: ai_project_id(string), created_at(integer), dst(array), finished_at(integer), id(string), in_progress_at(integer), params(object), src(array), status(string), type(string)
  operations_operation_id:
    primary key: id
    fields: ai_project_id(string), created_at(integer), dst(array), finished_at(integer), id(string), in_progress_at(integer), params(object), src(array), status(string), type(string)
  operations_operation_id_results:
    primary key: id
    fields: created_at(integer), fine_tuned_model_checkpoint(string), fine_tuning_job_id(string), id(string), metrics(object), object(string), result_files(array), step_number(integer)
  operations_operation_id_errors:
    fields: data(array), object(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_completions:
    endpoint: POST /v1/completions
    required fields: model, prompt
    risk: high: external Nebius API side effect or mutation; approval required
  create_chat_completions:
    endpoint: POST /v1/chat/completions
    required fields: model, messages
    risk: high: external Nebius API side effect or mutation; approval required
  create_embeddings:
    endpoint: POST /v1/embeddings
    required fields: model, input
    risk: high: external Nebius API side effect or mutation; approval required
  create_rerank:
    endpoint: POST /v1/rerank
    required fields: model, query, documents
    risk: high: external Nebius API side effect or mutation; approval required
  create_responses:
    endpoint: POST /v1/responses
    required fields: input, model
    risk: high: external Nebius API side effect or mutation; approval required
  delete_files_file_id:
    endpoint: DELETE /v1/files/{{ record.file_id }}
    required fields: file_id
    risk: medium: external Nebius API side effect or mutation; approval required
  create_images_generations:
    endpoint: POST /v1/images/generations
    required fields: model, prompt
    risk: high: external Nebius API side effect or mutation; approval required
  create_fine_tuning_jobs:
    endpoint: POST /v1/fine_tuning/jobs
    required fields: model, training_file
    risk: high: external Nebius API side effect or mutation; approval required
  execute_fine_tuning_jobs_job_id_cancel:
    endpoint: POST /v1/fine_tuning/jobs/{{ record.job_id }}/cancel
    required fields: job_id
    risk: high: external Nebius API side effect or mutation; approval required
  create_dedicated_endpoints:
    endpoint: POST /v0/dedicated_endpoints
    required fields: name, model_name, flavor_name, gpu_type, region, gpu_count, scaling
    risk: high: external Nebius API side effect or mutation; approval required
  update_dedicated_endpoints_endpoint_id:
    endpoint: PATCH /v0/dedicated_endpoints/{{ record.endpoint_id }}
    required fields: endpoint_id
    risk: high: external Nebius API side effect or mutation; approval required
  delete_dedicated_endpoints_endpoint_id:
    endpoint: DELETE /v0/dedicated_endpoints/{{ record.endpoint_id }}
    required fields: endpoint_id
    risk: high: external Nebius API side effect or mutation; approval required
  create_datasets:
    endpoint: POST /v1/datasets
    required fields: name, schema, folder, rows
    risk: high: external Nebius API side effect or mutation; approval required
  update_datasets_dataset_id:
    endpoint: PATCH /v1/datasets/{{ record.dataset_id }}
    required fields: dataset_id
    risk: high: external Nebius API side effect or mutation; approval required
  delete_datasets_dataset_id:
    endpoint: DELETE /v1/datasets/{{ record.dataset_id }}
    required fields: dataset_id
    risk: high: external Nebius API side effect or mutation; approval required
  create_datasets_uploads:
    endpoint: POST /v1/datasets/uploads
    required fields: name, schema, folder
    risk: high: external Nebius API side effect or mutation; approval required
  create_datasets_uploads_upload_id_complete:
    endpoint: POST /v1/datasets/uploads/{{ record.upload_id }}/complete
    required fields: upload_id, part_ids
    risk: high: external Nebius API side effect or mutation; approval required
  execute_datasets_uploads_upload_id_cancel:
    endpoint: POST /v1/datasets/uploads/{{ record.upload_id }}/cancel
    required fields: upload_id
    risk: high: external Nebius API side effect or mutation; approval required
  create_operations:
    endpoint: POST /v1/operations
    required fields: params, src
    risk: high: external Nebius API side effect or mutation; approval required
  execute_operations_operation_id_cancel:
    endpoint: POST /v1/operations/{{ record.operation_id }}/cancel
    required fields: operation_id
    risk: high: external Nebius API side effect or mutation; approval required

SECURITY
  read risk: external Nebius API reads of model, file, fine-tuning, dataset, operation, and dedicated-endpoint metadata
  write risk: external Nebius API writes and paid inference/operation side effects
  approval: required for every write action; destructive deletes require destructive confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect nebius-ai

  # Inspect as structured JSON
  pm connectors inspect nebius-ai --json

AGENT WORKFLOW
  - Run pm connectors inspect nebius-ai before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
