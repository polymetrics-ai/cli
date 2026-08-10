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
  id: simple-icons-formbricks
  asset: icons/simple-icons/formbricks.svg
  title: Formbricks
  simple_icon_slug: formbricks
  simple_icon_hex: 00C4B8
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Formbricks
  match: exact-name-or-slug
  matched_by: formbricks

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
    fields: created_at(string), environment_id(string), id(string), name(string), status(string), type(string), updated_at(string)
  survey_details:
    primary key: id
    cursor: updated_at
    fields: blocks(array), created_at(string), created_by(string), display_option(string), endings(array), hiddenFields(object), id(string), languages(array), name(string), questions(array), segment(object), singleUse(object), status(string), triggers(array), type(string), updated_at(string), welcomeCard(object), workspace_id(string)
  responses:
    primary key: id
    cursor: updated_at
    fields: contact_id(string), created_at(string), data(object), finished(boolean), id(string), meta(object), survey_id(string), updated_at(string)
  response_details:
    primary key: id
    cursor: updated_at
    fields: contact_id(string), created_at(string), data(object), finished(boolean), id(string), language(string), meta(object), person(object), personAttributes(object), singleUseId(string), survey_id(string), tags(array), ttc(object), updated_at(string)
  action_classes:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), environment_id(string), id(string), name(string), type(string), updated_at(string)
  action_class_details:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), name(string), noCodeConfig(object), type(string), updated_at(string), workspace_id(string)
  attribute_classes:
    primary key: id
    cursor: updated_at
    fields: archived(boolean), created_at(string), description(string), environment_id(string), id(string), name(string), type(string), updated_at(string)
  contact_attribute_keys:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), is_unique(boolean), key(string), name(string), type(string), updated_at(string), workspace_id(string)
  contact_attribute_key_details:
    primary key: id
    cursor: updated_at
    fields: created_at(string), description(string), id(string), is_unique(boolean), key(string), name(string), type(string), updated_at(string), workspace_id(string)
  contact_attributes:
    primary key: id
    cursor: updated_at
    fields: attribute_key_id(string), contact_id(string), created_at(string), id(string), updated_at(string), value(string)
  contacts:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), updated_at(string), user_id(string), workspace_id(string)
  contact_details:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), updated_at(string), user_id(string), workspace_id(string)
  me:
    primary key: id
    fields: app_setup_completed(boolean), created_at(string), environment_permissions(array), id(string), organization_access(object), organization_id(string), project(object), type(string), updated_at(string), website_setup_completed(boolean)
  webhooks:
    primary key: id
    fields: created_at(string), environment_id(string), id(string), source(string), surveyIds(array), triggers(array), updated_at(string), url(string)
  webhook_details:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), name(string), source(string), surveyIds(array), triggers(array), updated_at(string), url(string), workspace_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_action_class:
    endpoint: POST management/action-classes
    required fields: workspaceId, name, type
    risk: creates an action class in the configured Formbricks workspace
  delete_action_class:
    endpoint: DELETE management/action-classes/{{ record.actionClassId }}
    required fields: actionClassId
    risk: deletes an action class; automatic action classes may be rejected by Formbricks
  create_response:
    endpoint: POST management/responses
    required fields: surveyId
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
    required fields: fileName, fileType, workspaceId
    risk: creates a public file upload target and returns upload metadata
  create_survey:
    endpoint: POST management/surveys
    required fields: workspaceId, name, type, status
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
    required fields: url, triggers
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

COMMAND SURFACE
  Run Formbricks's declared streams and reverse-ETL actions.
  Usage: pm formbricks <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    action class details list - Run the action class details ETL stream [intent=etl availability=implemented stream=action_class_details]; notes: discrepancy=present-in-surface-absent-from-artifact
    action classes list - Run the action classes ETL stream [intent=etl availability=implemented stream=action_classes]; notes: discrepancy=present-in-surface-absent-from-artifact
    api delete api v1 management action-classes actionclassid - Documented DELETE /api/v1/management/action-classes/{actionClassId} (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.delete.api-v1-management-action-classes-actionclassid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 management responses responseid - Documented DELETE /api/v1/management/responses/{responseId} (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.delete.api-v1-management-responses-responseid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 management surveys surveyid - Documented DELETE /api/v1/management/surveys/{surveyId} (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.delete.api-v1-management-surveys-surveyid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v1 webhooks webhookid - Documented DELETE /api/v1/webhooks/{webhookId} (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.delete.api-v1-webhooks-webhookid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get api v1 client workspaceid environment - Documented GET /api/v1/client/{workspaceId}/environment (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-client-workspaceid-environment]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management action-classes - Documented GET /api/v1/management/action-classes (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-action-classes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management action-classes actionclassid - Documented GET /api/v1/management/action-classes/{actionClassId} (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-action-classes-actionclassid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management contact-attribute-keys - Documented GET /api/v1/management/contact-attribute-keys (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-contact-attribute-keys]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management contact-attribute-keys contactattributekeyid - Documented GET /api/v1/management/contact-attribute-keys/{contactAttributeKeyId} (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-contact-attribute-keys-contactattributekeyid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management contact-attributes - Documented GET /api/v1/management/contact-attributes (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-contact-attributes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management contacts - Documented GET /api/v1/management/contacts (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-contacts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management contacts contactid - Documented GET /api/v1/management/contacts/{contactId} (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-contacts-contactid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management me - Documented GET /api/v1/management/me (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-me]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management responses - Documented GET /api/v1/management/responses (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-responses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management responses responseid - Documented GET /api/v1/management/responses/{responseId} (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-responses-responseid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management surveys - Documented GET /api/v1/management/surveys (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-surveys]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management surveys surveyid - Documented GET /api/v1/management/surveys/{surveyId} (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-surveys-surveyid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 management surveys surveyid singleuseids - Documented GET /api/v1/management/surveys/{surveyId}/singleUseIds (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-management-surveys-surveyid-singleuseids]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 webhooks - Documented GET /api/v1/webhooks (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-webhooks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 webhooks webhookid - Documented GET /api/v1/webhooks/{webhookId} (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.api-v1-webhooks-webhookid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get client workspaceid environment - Documented GET client/{workspaceId}/environment (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.client-workspaceid-environment]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get health - Documented GET /health (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.health]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get health operation - Documented GET health (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.health-2]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get management surveys surveyid singleuseids - Documented GET management/surveys/{surveyId}/singleUseIds (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.management-surveys-surveyid-singleuseids]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get me - Documented GET me (not implemented) [intent=direct_read availability=not_implemented operation=formbricks.get.me]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post api v1 client workspaceid displays - Documented POST /api/v1/client/{workspaceId}/displays (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.api-v1-client-workspaceid-displays]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 client workspaceid responses - Documented POST /api/v1/client/{workspaceId}/responses (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.api-v1-client-workspaceid-responses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 client workspaceid user - Documented POST /api/v1/client/{workspaceId}/user (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.api-v1-client-workspaceid-user]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 management action-classes - Documented POST /api/v1/management/action-classes (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.api-v1-management-action-classes]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 management responses - Documented POST /api/v1/management/responses (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.api-v1-management-responses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 management storage - Documented POST /api/v1/management/storage (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.api-v1-management-storage]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 management surveys - Documented POST /api/v1/management/surveys (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.api-v1-management-surveys]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 webhooks - Documented POST /api/v1/webhooks (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.api-v1-webhooks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post client workspaceid displays - Documented POST client/{workspaceId}/displays (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.client-workspaceid-displays]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post client workspaceid responses - Documented POST client/{workspaceId}/responses (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.client-workspaceid-responses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post client workspaceid user - Documented POST client/{workspaceId}/user (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.post.client-workspaceid-user]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v1 client workspaceid responses responseid - Documented PUT /api/v1/client/{workspaceId}/responses/{responseId} (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.put.api-v1-client-workspaceid-responses-responseid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v1 management responses responseid - Documented PUT /api/v1/management/responses/{responseId} (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.put.api-v1-management-responses-responseid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v1 management surveys surveyid - Documented PUT /api/v1/management/surveys/{surveyId} (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.put.api-v1-management-surveys-surveyid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put client workspaceid responses responseid - Documented PUT client/{workspaceId}/responses/{responseId} (not implemented) [intent=direct_write availability=not_implemented operation=formbricks.put.client-workspaceid-responses-responseid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    attribute classes list - Run the attribute classes ETL stream [intent=etl availability=implemented stream=attribute_classes]; notes: discrepancy=present-in-surface-absent-from-artifact
    contact attribute key details list - Run the contact attribute key details ETL stream [intent=etl availability=implemented stream=contact_attribute_key_details]; notes: discrepancy=present-in-surface-absent-from-artifact
    contact attribute keys list - Run the contact attribute keys ETL stream [intent=etl availability=implemented stream=contact_attribute_keys]; notes: discrepancy=present-in-surface-absent-from-artifact
    contact attributes list - Run the contact attributes ETL stream [intent=etl availability=implemented stream=contact_attributes]; notes: discrepancy=present-in-surface-absent-from-artifact
    contact details list - Run the contact details ETL stream [intent=etl availability=implemented stream=contact_details]; notes: discrepancy=present-in-surface-absent-from-artifact
    contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]; notes: discrepancy=present-in-surface-absent-from-artifact
    create action class apply - Plan and execute the create action class reverse-ETL action [intent=reverse_etl availability=implemented write=create_action_class]; approval: requires plan, preview, approval, and execute; risk: creates an action class in the configured Formbricks workspace; flags: --name (required), --type (required), --workspaceId (required)
    create public file upload apply - Plan and execute the create public file upload reverse-ETL action [intent=reverse_etl availability=implemented write=create_public_file_upload]; approval: requires plan, preview, approval, and execute; risk: creates a public file upload target and returns upload metadata; flags: --fileName (required), --fileType (required), --workspaceId (required)
    create response apply - Plan and execute the create response reverse-ETL action [intent=reverse_etl availability=implemented write=create_response]; approval: requires plan, preview, approval, and execute; risk: creates a survey response and may trigger configured response pipelines; flags: --surveyId (required)
    create survey apply - Plan and execute the create survey reverse-ETL action [intent=reverse_etl availability=implemented write=create_survey]; approval: requires plan, preview, approval, and execute; risk: creates a survey in the configured Formbricks workspace; flags: --name (required), --status (required), --type (required), --workspaceId (required)
    create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: creates a webhook that sends Formbricks events to the configured URL; flags: --triggers (required), --url (required)
    delete action class apply - Plan and execute the delete action class reverse-ETL action [intent=reverse_etl availability=implemented write=delete_action_class]; approval: requires plan, preview, approval, and execute; risk: deletes an action class; automatic action classes may be rejected by Formbricks; flags: --actionClassId (required)
    delete response apply - Plan and execute the delete response reverse-ETL action [intent=reverse_etl availability=implemented write=delete_response]; approval: requires plan, preview, approval, and execute; risk: deletes a survey response; flags: --responseId (required)
    delete survey apply - Plan and execute the delete survey reverse-ETL action [intent=reverse_etl availability=implemented write=delete_survey]; approval: requires plan, preview, approval, and execute; risk: deletes a survey and its configured collection surface; flags: --surveyId (required)
    delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: deletes a webhook and stops future deliveries; flags: --webhookId (required)
    me list - Run the me ETL stream [intent=etl availability=implemented stream=me]; notes: discrepancy=present-in-surface-absent-from-artifact
    response details list - Run the response details ETL stream [intent=etl availability=implemented stream=response_details]; notes: discrepancy=present-in-surface-absent-from-artifact
    responses list - Run the responses ETL stream [intent=etl availability=implemented stream=responses]; notes: discrepancy=present-in-surface-absent-from-artifact
    survey details list - Run the survey details ETL stream [intent=etl availability=implemented stream=survey_details]; notes: discrepancy=present-in-surface-absent-from-artifact
    surveys list - Run the surveys ETL stream [intent=etl availability=implemented stream=surveys]; notes: discrepancy=present-in-surface-absent-from-artifact
    update response apply - Plan and execute the update response reverse-ETL action [intent=reverse_etl availability=implemented write=update_response]; approval: requires plan, preview, approval, and execute; risk: updates a survey response and may trigger configured response pipelines; flags: --responseId (required)
    update survey apply - Plan and execute the update survey reverse-ETL action [intent=reverse_etl availability=implemented write=update_survey]; approval: requires plan, preview, approval, and execute; risk: updates an existing survey; flags: --surveyId (required)
    webhook details list - Run the webhook details ETL stream [intent=etl availability=implemented stream=webhook_details]; notes: discrepancy=present-in-surface-absent-from-artifact
    webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]; notes: discrepancy=present-in-surface-absent-from-artifact

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
