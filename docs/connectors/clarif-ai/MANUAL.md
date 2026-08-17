# pm connectors inspect clarif-ai

```text
NAME
  pm connectors inspect clarif-ai - Clarif-ai connector manual

SYNOPSIS
  pm connectors inspect clarif-ai
  pm connectors inspect clarif-ai --json
  pm credentials add <name> --connector clarif-ai [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Clarifai applications, datasets, models, model versions, and workflows, and writes application/dataset lifecycle mutations, through the Clarifai v2 REST API.

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
  app_id
  base_url
  mode
  user_id
  api_key (secret)

ETL STREAMS
  applications:
    primary key: id
    fields: created_at(), default_language(), description(), id(), modified_at(), name(), user_id()
  datasets:
    primary key: id
    fields: app_id(), created_at(), default_processing_info(), description(), id(), modified_at(), user_id()
  models:
    primary key: id
    fields: app_id(), created_at(), id(), model_type_id(), modified_at(), name(), user_id(), visibility()
  model_versions:
    primary key: id
    fields: app_id(), created_at(), description(), id(), modified_at(), status(), user_id()
  workflows:
    primary key: id
    fields: app_id(), created_at(), id(), modified_at(), user_id(), version()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_application:
    endpoint: POST /users/{{ config.user_id }}/apps
    optional fields: apps
    risk: creates a new Clarifai application (workspace for datasets/models/workflows); low-risk (additive, no data loss)
  update_application:
    endpoint: PATCH /users/{{ config.user_id }}/apps
    optional fields: action, apps
    risk: updates an existing Clarifai application's settings (description, default workflow, notes); action=overwrite fully replaces the named fields rather than merging, so review the action value before use; approval required
  create_dataset:
    endpoint: POST /users/{{ config.user_id }}/apps/{{ config.app_id }}/datasets
    optional fields: datasets
    risk: creates a new Clarifai dataset within the configured app; low-risk (additive, no data loss)
  delete_dataset:
    endpoint: DELETE /users/{{ config.user_id }}/apps/{{ config.app_id }}/datasets
    optional fields: dataset_ids
    risk: permanently deletes one or more Clarifai datasets and their inputs/annotations within the configured app; irreversible; approval required

SECURITY
  read risk: external Clarifai API read of application, dataset, model, and workflow metadata
  write risk: external mutation of Clarifai applications and datasets; delete_dataset is irreversible (deletes the dataset and all its inputs/annotations) and update_application's action=overwrite can replace application settings wholesale; every write ships with an explicit per-action risk string
  approval: required for update_application and delete_dataset (destructive or full-replace semantics); create_application and create_dataset are low-risk (additive only)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect clarif-ai

  # Inspect as structured JSON
  pm connectors inspect clarif-ai --json

AGENT WORKFLOW
  - Run pm connectors inspect clarif-ai before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
