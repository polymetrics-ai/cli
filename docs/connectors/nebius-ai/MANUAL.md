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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

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
  api_key (secret)

ETL STREAMS
  models:
    primary key: id
    cursor: created
    fields: created(), id(), object(), owned_by()
  files:
    primary key: id
    cursor: created_at
    fields: bytes(), created_at(), filename(), id(), object(), purpose(), status()
  batches:
    primary key: id
    cursor: created_at
    fields: completed_at(), created_at(), endpoint(), error_file_id(), id(), input_file_id(), object(), output_file_id(), status()
  files_file_id:
    primary key: id
    fields: bytes(), created_at(), filename(), id(), object(), purpose(), status(), status_details()
  files_file_id_content:
  files_file_id_link:
    fields: url()
  fine_tuning_jobs:
    primary key: id
    fields: created_at(), error(), estimated_finish(), finished_at(), from_checkpoint(), hyperparameters(), id(), integrations(), method(), model(), object(), organization_id(), result_files(), seed(), status(), suffix(), total_steps(), trained_steps(), trained_tokens(), training_file(), validation_file()
  fine_tuning_jobs_job_id:
    primary key: id
    fields: created_at(), error(), estimated_finish(), finished_at(), from_checkpoint(), hyperparameters(), id(), integrations(), method(), model(), object(), organization_id(), result_files(), seed(), status(), suffix(), total_steps(), trained_steps(), trained_tokens(), training_file(), validation_file()
  fine_tuning_jobs_job_id_events:
    primary key: id
    fields: created_at(), data(), id(), level(), message(), object(), type()
  fine_tuning_jobs_job_id_checkpoints:
    primary key: id
    fields: created_at(), fine_tuned_model_checkpoint(), fine_tuning_job_id(), id(), metrics(), object(), result_files(), step_number()
  fine_tuning_jobs_job_id_checkpoints_checkpoint_id:
    primary key: id
    fields: created_at(), fine_tuned_model_checkpoint(), fine_tuning_job_id(), id(), metrics(), object(), step_number()
  fine_tuning_models_spec_draft:
    fields: hf_repo_name(), price()
  fine_tuning_models_spec_draft_2:
    fields: hf_repo_name(), price()
  dedicated_endpoints_templates:
    fields: flavors(), metadata(), name(), type()
  dedicated_endpoints:
    primary key: id
    fields: created_at(), custom_weights_id(), deployment(), description(), enabled(), flavor_name(), gpu_count(), gpu_type(), id(), model_name(), name(), region(), routing_key(), scaling()
  datasets:
    primary key: id
    fields: ai_project_id(), created_at(), current_version(), current_version_origin(), error(), folder(), id(), metadata(), name(), schema(), status(), type()
  datasets_dataset_id:
    primary key: id
    fields: ai_project_id(), created_at(), current_version(), current_version_origin(), error(), folder(), id(), metadata(), name(), schema(), status(), type()
  datasets_dataset_id_query_templates:
    fields: yql()
  datasets_dataset_id_content:
  datasets_dataset_id_export:
  datasets_uploads_upload_id_parts:
    primary key: id
    fields: created_at(), id(), status(), upload_id()
  datasets_uploads_upload_id:
    primary key: id
    fields: id(), upload_info()
  operations:
    primary key: id
    fields: ai_project_id(), created_at(), dst(), finished_at(), id(), in_progress_at(), params(), src(), status(), type()
  operations_operation_id:
    primary key: id
    fields: ai_project_id(), created_at(), dst(), finished_at(), id(), in_progress_at(), params(), src(), status(), type()
  operations_operation_id_results:
    primary key: id
    fields: created_at(), fine_tuned_model_checkpoint(), fine_tuning_job_id(), id(), metrics(), object(), result_files(), step_number()
  operations_operation_id_errors:
    fields: data(), object()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_completions:
    endpoint: POST /v1/completions
    risk: high: external Nebius API side effect or mutation; approval required
  create_chat_completions:
    endpoint: POST /v1/chat/completions
    risk: high: external Nebius API side effect or mutation; approval required
  create_embeddings:
    endpoint: POST /v1/embeddings
    risk: high: external Nebius API side effect or mutation; approval required
  create_rerank:
    endpoint: POST /v1/rerank
    risk: high: external Nebius API side effect or mutation; approval required
  create_responses:
    endpoint: POST /v1/responses
    risk: high: external Nebius API side effect or mutation; approval required
  delete_files_file_id:
    endpoint: DELETE /v1/files/{{ record.file_id }}
    required fields: file_id
    risk: medium: external Nebius API side effect or mutation; approval required
  create_images_generations:
    endpoint: POST /v1/images/generations
    risk: high: external Nebius API side effect or mutation; approval required
  create_fine_tuning_jobs:
    endpoint: POST /v1/fine_tuning/jobs
    risk: high: external Nebius API side effect or mutation; approval required
  execute_fine_tuning_jobs_job_id_cancel:
    endpoint: POST /v1/fine_tuning/jobs/{{ record.job_id }}/cancel
    required fields: job_id
    risk: high: external Nebius API side effect or mutation; approval required
  create_dedicated_endpoints:
    endpoint: POST /v0/dedicated_endpoints
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
    risk: high: external Nebius API side effect or mutation; approval required
  create_datasets_uploads_upload_id_complete:
    endpoint: POST /v1/datasets/uploads/{{ record.upload_id }}/complete
    required fields: upload_id
    risk: high: external Nebius API side effect or mutation; approval required
  execute_datasets_uploads_upload_id_cancel:
    endpoint: POST /v1/datasets/uploads/{{ record.upload_id }}/cancel
    required fields: upload_id
    risk: high: external Nebius API side effect or mutation; approval required
  create_operations:
    endpoint: POST /v1/operations
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
