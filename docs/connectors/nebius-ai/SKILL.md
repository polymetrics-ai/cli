---
name: pm-nebius-ai
description: Nebius AI connector knowledge and safe action guide.
---

# pm-nebius-ai

## Purpose

Reads and writes Nebius Token Factory OpenAI-compatible API resources, including models, files, fine-tuning, datasets, operations, dedicated endpoints, and inference actions.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- checkpoint_id
- dataset_id
- file_id
- job_id
- limit
- operation_id
- upload_id
- api_key (secret)

## ETL Streams

- models:
  - primary key: id
  - cursor: created
  - fields: created(integer), id(string), object(string), owned_by(string)
- files:
  - primary key: id
  - cursor: created_at
  - fields: bytes(integer), created_at(integer), filename(string), id(string), object(string), purpose(string), status(string)
- batches:
  - primary key: id
  - cursor: created_at
  - fields: completed_at(integer), created_at(integer), endpoint(string), error_file_id(string), id(string), input_file_id(string), object(string), output_file_id(string), status(string)
- files_file_id:
  - primary key: id
  - fields: bytes(integer), created_at(integer), filename(string), id(string), object(string), purpose(string), status(string), status_details(string)
- files_file_id_content:
- files_file_id_link:
  - fields: url(string)
- fine_tuning_jobs:
  - primary key: id
  - fields: created_at(integer), error(object), estimated_finish(integer), finished_at(integer), from_checkpoint(object), hyperparameters(object), id(string), integrations(array), method(object), model(string), object(string), organization_id(string), result_files(array), seed(integer), status(string), suffix(string), total_steps(integer), trained_steps(integer), trained_tokens(integer), training_file(string), validation_file(string)
- fine_tuning_jobs_job_id:
  - primary key: id
  - fields: created_at(integer), error(object), estimated_finish(integer), finished_at(integer), from_checkpoint(object), hyperparameters(object), id(string), integrations(array), method(object), model(string), object(string), organization_id(string), result_files(array), seed(integer), status(string), suffix(string), total_steps(integer), trained_steps(integer), trained_tokens(integer), training_file(string), validation_file(string)
- fine_tuning_jobs_job_id_events:
  - primary key: id
  - fields: created_at(integer), data(object), id(string), level(string), message(string), object(string), type(string)
- fine_tuning_jobs_job_id_checkpoints:
  - primary key: id
  - fields: created_at(integer), fine_tuned_model_checkpoint(string), fine_tuning_job_id(string), id(string), metrics(object), object(string), result_files(array), step_number(integer)
- fine_tuning_jobs_job_id_checkpoints_checkpoint_id:
  - primary key: id
  - fields: created_at(integer), fine_tuned_model_checkpoint(string), fine_tuning_job_id(string), id(string), metrics(object), object(string), step_number(integer)
- fine_tuning_models_spec_draft:
  - fields: hf_repo_name(string), price(object)
- fine_tuning_models_spec_draft_2:
  - fields: hf_repo_name(string), price(object)
- dedicated_endpoints_templates:
  - fields: flavors(object), metadata(object), name(string), type(string)
- dedicated_endpoints:
  - primary key: id
  - fields: created_at(string), custom_weights_id(string), deployment(object), description(string), enabled(boolean), flavor_name(string), gpu_count(integer), gpu_type(string), id(string), model_name(string), name(string), region(string), routing_key(string), scaling(object)
- datasets:
  - primary key: id
  - fields: ai_project_id(string), created_at(integer), current_version(string), current_version_origin(object), error(string), folder(string), id(string), metadata(object), name(string), schema(array), status(string), type(string)
- datasets_dataset_id:
  - primary key: id
  - fields: ai_project_id(string), created_at(integer), current_version(string), current_version_origin(object), error(string), folder(string), id(string), metadata(object), name(string), schema(array), status(string), type(string)
- datasets_dataset_id_query_templates:
  - fields: yql(string)
- datasets_dataset_id_content:
- datasets_dataset_id_export:
- datasets_uploads_upload_id_parts:
  - primary key: id
  - fields: created_at(integer), id(string), status(string), upload_id(string)
- datasets_uploads_upload_id:
  - primary key: id
  - fields: id(string), upload_info(object)
- operations:
  - primary key: id
  - fields: ai_project_id(string), created_at(integer), dst(array), finished_at(integer), id(string), in_progress_at(integer), params(object), src(array), status(string), type(string)
- operations_operation_id:
  - primary key: id
  - fields: ai_project_id(string), created_at(integer), dst(array), finished_at(integer), id(string), in_progress_at(integer), params(object), src(array), status(string), type(string)
- operations_operation_id_results:
  - primary key: id
  - fields: created_at(integer), fine_tuned_model_checkpoint(string), fine_tuning_job_id(string), id(string), metrics(object), object(string), result_files(array), step_number(integer)
- operations_operation_id_errors:
  - fields: data(array), object(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_completions:
  - endpoint: POST /v1/completions
  - required fields: model, prompt
  - risk: high: external Nebius API side effect or mutation; approval required
- create_chat_completions:
  - endpoint: POST /v1/chat/completions
  - required fields: model, messages
  - risk: high: external Nebius API side effect or mutation; approval required
- create_embeddings:
  - endpoint: POST /v1/embeddings
  - required fields: model, input
  - risk: high: external Nebius API side effect or mutation; approval required
- create_rerank:
  - endpoint: POST /v1/rerank
  - required fields: model, query, documents
  - risk: high: external Nebius API side effect or mutation; approval required
- create_responses:
  - endpoint: POST /v1/responses
  - required fields: input, model
  - risk: high: external Nebius API side effect or mutation; approval required
- delete_files_file_id:
  - endpoint: DELETE /v1/files/{{ record.file_id }}
  - required fields: file_id
  - risk: medium: external Nebius API side effect or mutation; approval required
- create_images_generations:
  - endpoint: POST /v1/images/generations
  - required fields: model, prompt
  - risk: high: external Nebius API side effect or mutation; approval required
- create_fine_tuning_jobs:
  - endpoint: POST /v1/fine_tuning/jobs
  - required fields: model, training_file
  - risk: high: external Nebius API side effect or mutation; approval required
- execute_fine_tuning_jobs_job_id_cancel:
  - endpoint: POST /v1/fine_tuning/jobs/{{ record.job_id }}/cancel
  - required fields: job_id
  - risk: high: external Nebius API side effect or mutation; approval required
- create_dedicated_endpoints:
  - endpoint: POST /v0/dedicated_endpoints
  - required fields: name, model_name, flavor_name, gpu_type, region, gpu_count, scaling
  - risk: high: external Nebius API side effect or mutation; approval required
- update_dedicated_endpoints_endpoint_id:
  - endpoint: PATCH /v0/dedicated_endpoints/{{ record.endpoint_id }}
  - required fields: endpoint_id
  - risk: high: external Nebius API side effect or mutation; approval required
- delete_dedicated_endpoints_endpoint_id:
  - endpoint: DELETE /v0/dedicated_endpoints/{{ record.endpoint_id }}
  - required fields: endpoint_id
  - risk: high: external Nebius API side effect or mutation; approval required
- create_datasets:
  - endpoint: POST /v1/datasets
  - required fields: name, schema, folder, rows
  - risk: high: external Nebius API side effect or mutation; approval required
- update_datasets_dataset_id:
  - endpoint: PATCH /v1/datasets/{{ record.dataset_id }}
  - required fields: dataset_id
  - risk: high: external Nebius API side effect or mutation; approval required
- delete_datasets_dataset_id:
  - endpoint: DELETE /v1/datasets/{{ record.dataset_id }}
  - required fields: dataset_id
  - risk: high: external Nebius API side effect or mutation; approval required
- create_datasets_uploads:
  - endpoint: POST /v1/datasets/uploads
  - required fields: name, schema, folder
  - risk: high: external Nebius API side effect or mutation; approval required
- create_datasets_uploads_upload_id_complete:
  - endpoint: POST /v1/datasets/uploads/{{ record.upload_id }}/complete
  - required fields: upload_id, part_ids
  - risk: high: external Nebius API side effect or mutation; approval required
- execute_datasets_uploads_upload_id_cancel:
  - endpoint: POST /v1/datasets/uploads/{{ record.upload_id }}/cancel
  - required fields: upload_id
  - risk: high: external Nebius API side effect or mutation; approval required
- create_operations:
  - endpoint: POST /v1/operations
  - required fields: params, src
  - risk: high: external Nebius API side effect or mutation; approval required
- execute_operations_operation_id_cancel:
  - endpoint: POST /v1/operations/{{ record.operation_id }}/cancel
  - required fields: operation_id
  - risk: high: external Nebius API side effect or mutation; approval required

## Security

- read risk: external Nebius API reads of model, file, fine-tuning, dataset, operation, and dedicated-endpoint metadata
- write risk: external Nebius API writes and paid inference/operation side effects
- approval: required for every write action; destructive deletes require destructive confirmation
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Nebius AI's declared streams and reverse-ETL actions.
- Usage: pm nebius-ai <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api post v0 models upload - Documented POST /v0/models/upload (not implemented) [intent=direct_write availability=not_implemented operation=nebius-ai.post.v0-models-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 datasets uploads upload-id parts - Documented POST /v1/datasets/uploads/{upload_id}/parts (not implemented) [intent=direct_write availability=not_implemented operation=nebius-ai.post.v1-datasets-uploads-upload-id-parts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 files - Documented POST /v1/files (not implemented) [intent=direct_write availability=not_implemented operation=nebius-ai.post.v1-files]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - batches list - Run the batches ETL stream [intent=etl availability=implemented stream=batches]; notes: discrepancy=present-in-surface-absent-from-artifact
  - create chat completions apply - Plan and execute the create chat completions reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_chat_completions]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create completions apply - Plan and execute the create completions reverse-ETL action [intent=reverse_etl availability=implemented write=create_completions]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --model (required), --prompt (required)
  - create datasets apply - Plan and execute the create datasets reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_datasets]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create datasets uploads apply - Plan and execute the create datasets uploads reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_datasets_uploads]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create datasets uploads upload id complete apply - Plan and execute the create datasets uploads upload id complete reverse-ETL action [intent=reverse_etl availability=implemented write=create_datasets_uploads_upload_id_complete]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --part_ids (required), --upload_id (required)
  - create dedicated endpoints apply - Plan and execute the create dedicated endpoints reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_dedicated_endpoints]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create embeddings apply - Plan and execute the create embeddings reverse-ETL action [intent=reverse_etl availability=implemented write=create_embeddings]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --input (required), --model (required)
  - create fine tuning jobs apply - Plan and execute the create fine tuning jobs reverse-ETL action [intent=reverse_etl availability=implemented write=create_fine_tuning_jobs]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --model (required), --training_file (required)
  - create images generations apply - Plan and execute the create images generations reverse-ETL action [intent=reverse_etl availability=implemented write=create_images_generations]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --model (required), --prompt (required)
  - create operations apply - Plan and execute the create operations reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_operations]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create rerank apply - Plan and execute the create rerank reverse-ETL action [intent=reverse_etl availability=implemented write=create_rerank]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --documents (required), --model (required), --query (required)
  - create responses apply - Plan and execute the create responses reverse-ETL action [intent=reverse_etl availability=implemented write=create_responses]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --input (required), --model (required)
  - datasets dataset id content list - Run the datasets dataset id content ETL stream [intent=etl availability=implemented stream=datasets_dataset_id_content]
  - datasets dataset id export list - Run the datasets dataset id export ETL stream [intent=etl availability=implemented stream=datasets_dataset_id_export]
  - datasets dataset id list - Run the datasets dataset id ETL stream [intent=etl availability=implemented stream=datasets_dataset_id]
  - datasets dataset id query templates list - Run the datasets dataset id query templates ETL stream [intent=etl availability=implemented stream=datasets_dataset_id_query_templates]
  - datasets list - Run the datasets ETL stream [intent=etl availability=implemented stream=datasets]
  - datasets uploads upload id list - Run the datasets uploads upload id ETL stream [intent=etl availability=implemented stream=datasets_uploads_upload_id]
  - datasets uploads upload id parts list - Run the datasets uploads upload id parts ETL stream [intent=etl availability=implemented stream=datasets_uploads_upload_id_parts]
  - dedicated endpoints list - Run the dedicated endpoints ETL stream [intent=etl availability=implemented stream=dedicated_endpoints]
  - dedicated endpoints templates list - Run the dedicated endpoints templates ETL stream [intent=etl availability=implemented stream=dedicated_endpoints_templates]
  - delete datasets dataset id apply - Plan and execute the delete datasets dataset id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_datasets_dataset_id]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --dataset_id (required)
  - delete dedicated endpoints endpoint id apply - Plan and execute the delete dedicated endpoints endpoint id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_dedicated_endpoints_endpoint_id]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --endpoint_id (required)
  - delete files file id apply - Plan and execute the delete files file id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_files_file_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Nebius API side effect or mutation; approval required; flags: --file_id (required)
  - execute datasets uploads upload id cancel apply - Plan and execute the execute datasets uploads upload id cancel reverse-ETL action [intent=reverse_etl availability=implemented write=execute_datasets_uploads_upload_id_cancel]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --upload_id (required)
  - execute fine tuning jobs job id cancel apply - Plan and execute the execute fine tuning jobs job id cancel reverse-ETL action [intent=reverse_etl availability=implemented write=execute_fine_tuning_jobs_job_id_cancel]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --job_id (required)
  - execute operations operation id cancel apply - Plan and execute the execute operations operation id cancel reverse-ETL action [intent=reverse_etl availability=implemented write=execute_operations_operation_id_cancel]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --operation_id (required)
  - files file id content list - Run the files file id content ETL stream [intent=etl availability=implemented stream=files_file_id_content]
  - files file id link list - Run the files file id link ETL stream [intent=etl availability=implemented stream=files_file_id_link]
  - files file id list - Run the files file id ETL stream [intent=etl availability=implemented stream=files_file_id]
  - files list - Run the files ETL stream [intent=etl availability=implemented stream=files]
  - fine tuning jobs job id checkpoints checkpoint id list - Run the fine tuning jobs job id checkpoints checkpoint id ETL stream [intent=etl availability=implemented stream=fine_tuning_jobs_job_id_checkpoints_checkpoint_id]
  - fine tuning jobs job id checkpoints list - Run the fine tuning jobs job id checkpoints ETL stream [intent=etl availability=implemented stream=fine_tuning_jobs_job_id_checkpoints]
  - fine tuning jobs job id events list - Run the fine tuning jobs job id events ETL stream [intent=etl availability=implemented stream=fine_tuning_jobs_job_id_events]
  - fine tuning jobs job id list - Run the fine tuning jobs job id ETL stream [intent=etl availability=implemented stream=fine_tuning_jobs_job_id]
  - fine tuning jobs list - Run the fine tuning jobs ETL stream [intent=etl availability=implemented stream=fine_tuning_jobs]
  - fine tuning models spec draft 2 list - Run the fine tuning models spec draft 2 ETL stream [intent=etl availability=implemented stream=fine_tuning_models_spec_draft_2]
  - fine tuning models spec draft list - Run the fine tuning models spec draft ETL stream [intent=etl availability=implemented stream=fine_tuning_models_spec_draft]
  - models list - Run the models ETL stream [intent=etl availability=implemented stream=models]
  - operations list - Run the operations ETL stream [intent=etl availability=implemented stream=operations]
  - operations operation id errors list - Run the operations operation id errors ETL stream [intent=etl availability=implemented stream=operations_operation_id_errors]
  - operations operation id list - Run the operations operation id ETL stream [intent=etl availability=implemented stream=operations_operation_id]
  - operations operation id results list - Run the operations operation id results ETL stream [intent=etl availability=implemented stream=operations_operation_id_results]
  - update datasets dataset id apply - Plan and execute the update datasets dataset id reverse-ETL action [intent=reverse_etl availability=implemented write=update_datasets_dataset_id]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --dataset_id (required)
  - update dedicated endpoints endpoint id apply - Plan and execute the update dedicated endpoints endpoint id reverse-ETL action [intent=reverse_etl availability=implemented write=update_dedicated_endpoints_endpoint_id]; approval: requires plan, preview, approval, and execute; risk: high: external Nebius API side effect or mutation; approval required; flags: --endpoint_id (required)

## Commands

### Inspect as a manual

```bash
pm connectors inspect nebius-ai
```

### Inspect as structured JSON

```bash
pm connectors inspect nebius-ai --json
```

## Agent Rules

- Run pm connectors inspect nebius-ai before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
