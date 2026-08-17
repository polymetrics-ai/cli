# pm connectors inspect formbricks

```text
NAME
  pm connectors inspect formbricks - Formbricks connector manual

SYNOPSIS
  pm connectors inspect formbricks
  pm connectors inspect formbricks --json
  pm credentials add <name> --connector formbricks [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Formbricks surveys, responses, contacts, contact attributes, action classes, webhooks, and account metadata; writes approved management API mutations.

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
  response_ids
  survey_id
  api_key (secret)

ETL STREAMS
  surveys:
    primary key: id
    cursor: updated_at
    fields: created_at(), environment_id(), id(), name(), status(), type(), updated_at()
  survey_details:
    primary key: id
    cursor: updated_at
    fields: blocks(), created_at(), created_by(), display_option(), endings(), hiddenFields(), id(), languages(), name(), questions(), segment(), singleUse(), status(), triggers(), type(), updated_at(), welcomeCard(), workspace_id()
  responses:
    primary key: id
    cursor: updated_at
    fields: contact_id(), created_at(), data(), finished(), id(), meta(), survey_id(), updated_at()
  response_details:
    primary key: id
    cursor: updated_at
    fields: contact_id(), created_at(), data(), finished(), id(), language(), meta(), person(), personAttributes(), singleUseId(), survey_id(), tags(), ttc(), updated_at()
  action_classes:
    primary key: id
    cursor: updated_at
    fields: created_at(), description(), environment_id(), id(), name(), type(), updated_at()
  action_class_details:
    primary key: id
    cursor: updated_at
    fields: created_at(), description(), id(), name(), noCodeConfig(), type(), updated_at(), workspace_id()
  attribute_classes:
    primary key: id
    cursor: updated_at
    fields: archived(), created_at(), description(), environment_id(), id(), name(), type(), updated_at()
  contact_attribute_keys:
    primary key: id
    cursor: updated_at
    fields: created_at(), description(), id(), is_unique(), key(), name(), type(), updated_at(), workspace_id()
  contact_attribute_key_details:
    primary key: id
    cursor: updated_at
    fields: created_at(), description(), id(), is_unique(), key(), name(), type(), updated_at(), workspace_id()
  contact_attributes:
    primary key: id
    cursor: updated_at
    fields: attribute_key_id(), contact_id(), created_at(), id(), updated_at(), value()
  contacts:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), updated_at(), user_id(), workspace_id()
  contact_details:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), updated_at(), user_id(), workspace_id()
  me:
    primary key: id
    fields: app_setup_completed(), created_at(), environment_permissions(), id(), organization_access(), organization_id(), project(), type(), updated_at(), website_setup_completed()
  webhooks:
    primary key: id
    fields: created_at(), environment_id(), id(), source(), surveyIds(), triggers(), updated_at(), url()
  webhook_details:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), name(), source(), surveyIds(), triggers(), updated_at(), url(), workspace_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_action_class:
    endpoint: POST management/action-classes
    risk: creates an action class in the configured Formbricks workspace
  delete_action_class:
    endpoint: DELETE management/action-classes/{{ record.actionClassId }}
    required fields: actionClassId
    risk: deletes an action class; automatic action classes may be rejected by Formbricks
  create_response:
    endpoint: POST management/responses
    risk: creates a survey response and may trigger configured response pipelines
  update_response:
    endpoint: PUT management/responses/{{ record.responseId }}
    required fields: responseId
    risk: updates a survey response and may trigger configured response pipelines
  delete_response:
    endpoint: DELETE management/responses/{{ record.responseId }}
    required fields: responseId
    risk: deletes a survey response
  create_public_file_upload:
    endpoint: POST management/storage
    risk: creates a public file upload target and returns upload metadata
  create_survey:
    endpoint: POST management/surveys
    risk: creates a survey in the configured Formbricks workspace
  update_survey:
    endpoint: PUT management/surveys/{{ record.surveyId }}
    required fields: surveyId
    risk: updates an existing survey
  delete_survey:
    endpoint: DELETE management/surveys/{{ record.surveyId }}
    required fields: surveyId
    risk: deletes a survey and its configured collection surface
  create_webhook:
    endpoint: POST webhooks
    risk: creates a webhook that sends Formbricks events to the configured URL
  delete_webhook:
    endpoint: DELETE webhooks/{{ record.webhookId }}
    required fields: webhookId
    risk: deletes a webhook and stops future deliveries

SECURITY
  read risk: external Formbricks management API reads of surveys, responses, contacts, contact attributes, action classes, webhooks, and API-key metadata
  write risk: external Formbricks management API mutations for action classes, responses, public upload URLs, surveys, and webhooks
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect formbricks

  # Inspect as structured JSON
  pm connectors inspect formbricks --json

AGENT WORKFLOW
  - Run pm connectors inspect formbricks before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
