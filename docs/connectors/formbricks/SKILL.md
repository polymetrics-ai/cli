---
name: pm-formbricks
description: Formbricks connector knowledge and safe action guide.
---

# pm-formbricks

## Purpose

Reads Formbricks surveys, responses, contacts, contact attributes, action classes, webhooks, and account metadata; writes approved management API mutations.

## Icon

- id: simple-icons-formbricks
- asset: icons/simple-icons/formbricks.svg
- title: Formbricks
- simple_icon_slug: formbricks
- simple_icon_hex: 00C4B8
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Formbricks
- match: exact-name-or-slug
- matched_by: formbricks

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- response_ids
- survey_id
- api_key (secret)

## ETL Streams

- surveys:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), environment_id(string), id(string), name(string), status(string), type(string), updated_at(string)
- survey_details:
  - primary key: id
  - cursor: updated_at
  - fields: blocks(array), created_at(string), created_by(string), display_option(string), endings(array), hiddenFields(object), id(string), languages(array), name(string), questions(array), segment(object), singleUse(object), status(string), triggers(array), type(string), updated_at(string), welcomeCard(object), workspace_id(string)
- responses:
  - primary key: id
  - cursor: updated_at
  - fields: contact_id(string), created_at(string), data(object), finished(boolean), id(string), meta(object), survey_id(string), updated_at(string)
- response_details:
  - primary key: id
  - cursor: updated_at
  - fields: contact_id(string), created_at(string), data(object), finished(boolean), id(string), language(string), meta(object), person(object), personAttributes(object), singleUseId(string), survey_id(string), tags(array), ttc(object), updated_at(string)
- action_classes:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), description(string), environment_id(string), id(string), name(string), type(string), updated_at(string)
- action_class_details:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), description(string), id(string), name(string), noCodeConfig(object), type(string), updated_at(string), workspace_id(string)
- attribute_classes:
  - primary key: id
  - cursor: updated_at
  - fields: archived(boolean), created_at(string), description(string), environment_id(string), id(string), name(string), type(string), updated_at(string)
- contact_attribute_keys:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), description(string), id(string), is_unique(boolean), key(string), name(string), type(string), updated_at(string), workspace_id(string)
- contact_attribute_key_details:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), description(string), id(string), is_unique(boolean), key(string), name(string), type(string), updated_at(string), workspace_id(string)
- contact_attributes:
  - primary key: id
  - cursor: updated_at
  - fields: attribute_key_id(string), contact_id(string), created_at(string), id(string), updated_at(string), value(string)
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), updated_at(string), user_id(string), workspace_id(string)
- contact_details:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), updated_at(string), user_id(string), workspace_id(string)
- me:
  - primary key: id
  - fields: app_setup_completed(boolean), created_at(string), environment_permissions(array), id(string), organization_access(object), organization_id(string), project(object), type(string), updated_at(string), website_setup_completed(boolean)
- webhooks:
  - primary key: id
  - fields: created_at(string), environment_id(string), id(string), source(string), surveyIds(array), triggers(array), updated_at(string), url(string)
- webhook_details:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), name(string), source(string), surveyIds(array), triggers(array), updated_at(string), url(string), workspace_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_action_class:
  - endpoint: POST management/action-classes
  - required fields: workspaceId, name, type
  - risk: creates an action class in the configured Formbricks workspace
- delete_action_class:
  - endpoint: DELETE management/action-classes/{{ record.actionClassId }}
  - required fields: actionClassId
  - risk: deletes an action class; automatic action classes may be rejected by Formbricks
- create_response:
  - endpoint: POST management/responses
  - required fields: surveyId
  - risk: creates a survey response and may trigger configured response pipelines
- update_response:
  - endpoint: PUT management/responses/{{ record.responseId }}
  - required fields: responseId
  - risk: updates a survey response and may trigger configured response pipelines
- delete_response:
  - endpoint: DELETE management/responses/{{ record.responseId }}
  - required fields: responseId
  - risk: deletes a survey response
- create_public_file_upload:
  - endpoint: POST management/storage
  - required fields: fileName, fileType, workspaceId
  - risk: creates a public file upload target and returns upload metadata
- create_survey:
  - endpoint: POST management/surveys
  - required fields: workspaceId, name, type, status
  - risk: creates a survey in the configured Formbricks workspace
- update_survey:
  - endpoint: PUT management/surveys/{{ record.surveyId }}
  - required fields: surveyId
  - risk: updates an existing survey
- delete_survey:
  - endpoint: DELETE management/surveys/{{ record.surveyId }}
  - required fields: surveyId
  - risk: deletes a survey and its configured collection surface
- create_webhook:
  - endpoint: POST webhooks
  - required fields: url, triggers
  - risk: creates a webhook that sends Formbricks events to the configured URL
- delete_webhook:
  - endpoint: DELETE webhooks/{{ record.webhookId }}
  - required fields: webhookId
  - risk: deletes a webhook and stops future deliveries

## Security

- read risk: external Formbricks management API reads of surveys, responses, contacts, contact attributes, action classes, webhooks, and API-key metadata
- write risk: external Formbricks management API mutations for action classes, responses, public upload URLs, surveys, and webhooks
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect formbricks
```

### Inspect as structured JSON

```bash
pm connectors inspect formbricks --json
```

## Agent Rules

- Run pm connectors inspect formbricks before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
