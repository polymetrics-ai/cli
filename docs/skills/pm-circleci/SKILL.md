---
name: pm-circleci
description: CircleCI connector knowledge and safe action guide.
---

# pm-circleci

## Purpose

Reads and writes CircleCI projects, pipelines, workflows, jobs, contexts, schedules, environment variables, checkout keys, and workflow insights through the CircleCI v2 REST API.

## Icon

- id: simple-icons-circleci
- asset: icons/simple-icons/circleci.svg
- title: CircleCI
- simple_icon_slug: circleci
- simple_icon_hex: 343434
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=CircleCI
- match: exact-name-or-slug
- matched_by: circleci

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- org
- pipeline_id
- repo
- vcs_type
- workflow_id
- api_key (secret) (required)

## ETL Streams

- projects:
  - primary key: id
  - fields: default_branch(string), id(string), name(string), organization_id(string), organization_name(string), organization_slug(string), slug(string), vcs_url(string)
- pipelines:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), number(integer), project_slug(string), state(string), updated_at(string)
- workflows:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), name(string), pipeline_id(string), pipeline_number(integer), project_slug(string), status(string), stopped_at(string)
- jobs:
  - primary key: id
  - cursor: started_at
  - fields: id(string), job_number(integer), name(string), project_slug(string), started_at(string), status(string), stopped_at(string), type(string)
- contexts:
  - primary key: id
  - fields: created_at(string), id(string), name(string)
- schedules:
  - primary key: id
  - cursor: updated-at
  - fields: actor(object), created-at(string), description(string), id(string), name(string), parameters(object), project-slug(string), timetable(object), updated-at(string)
- checkout_keys:
  - primary key: fingerprint
  - fields: created-at(string), fingerprint(string), preferred(boolean), public-key(string), type(string)
- environment_variables:
  - primary key: name
  - fields: created-at(string), name(string), value(string)
- insights_workflow_summary:
  - primary key: name
  - fields: metrics(object), name(string), project_id(string), window_end(string), window_start(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_schedule:
  - endpoint: POST /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/schedule
  - required fields: name, timetable, attribution-actor, parameters
  - risk: external mutation; creates a new scheduled-pipeline trigger for this project
- update_schedule:
  - endpoint: PATCH /schedule/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an existing scheduled-pipeline trigger's timetable or parameters
- delete_schedule:
  - endpoint: DELETE /schedule/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a scheduled-pipeline trigger; approval required
- create_environment_variable:
  - endpoint: POST /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/envvar
  - required fields: name, value
  - risk: external mutation; creates or overwrites a project environment variable used by every future CI run
- delete_environment_variable:
  - endpoint: DELETE /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/envvar/{{ record.name }}
  - required fields: name
  - risk: irreversible external deletion of a project environment variable; may break future CI runs that depend on it; approval required
- create_checkout_key:
  - endpoint: POST /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/checkout-key
  - required fields: type
  - risk: external mutation; creates a new deploy/checkout SSH key with repository access
- delete_checkout_key:
  - endpoint: DELETE /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/checkout-key/{{ record.fingerprint }}
  - required fields: fingerprint
  - risk: irreversible external revocation of a deploy/checkout SSH key; may break future CI checkouts that depend on it; approval required
- delete_context:
  - endpoint: DELETE /context/{{ record.context_id }}
  - required fields: context_id
  - risk: external CircleCI API mutation: deleteContext
- delete_environment_variable_from_context:
  - endpoint: DELETE /context/{{ record.context_id }}/environment-variable/{{ record.env_var_name }}
  - required fields: context_id, env_var_name
  - risk: external CircleCI API mutation: deleteEnvironmentVariableFromContext
- delete_context_restriction:
  - endpoint: DELETE /context/{{ record.context_id }}/restrictions/{{ record.restriction_id }}
  - required fields: context_id, restriction_id
  - risk: external CircleCI API mutation: deleteContextRestriction
- delete_org_claims:
  - endpoint: DELETE /org/{{ record.orgID }}/oidc-custom-claims
  - required fields: orgID, claims
  - risk: external CircleCI API mutation: DeleteOrgClaims
- delete_project_claims:
  - endpoint: DELETE /org/{{ record.orgID }}/project/{{ record.projectID }}/oidc-custom-claims
  - required fields: orgID, projectID, claims
  - risk: external CircleCI API mutation: DeleteProjectClaims
- delete_organization:
  - endpoint: DELETE /organization/{{ record.org-slug-or-id }}
  - required fields: org-slug-or-id
  - risk: external CircleCI API mutation: deleteOrganization
- remove_u_r_l_orb_allow_list_entry:
  - endpoint: DELETE /organization/{{ record.org-slug-or-id }}/url-orb-allow-list/{{ record.allow-list-entry-id }}
  - required fields: allow-list-entry-id, org-slug-or-id
  - risk: external CircleCI API mutation: removeURLOrbAllowListEntry
- delete_group:
  - endpoint: DELETE /organizations/{{ record.org_id }}/groups/{{ record.group_id }}
  - required fields: group_id, org_id
  - risk: external CircleCI API mutation: deleteGroup
- delete_otel_exporter:
  - endpoint: DELETE /otel/exporters/{{ record.otel_exporter_id }}
  - required fields: otel_exporter_id
  - risk: external CircleCI API mutation: deleteOtelExporter
- delete_project_by_slug:
  - endpoint: DELETE /project/{{ record.project-slug }}
  - required fields: project-slug
  - risk: external CircleCI API mutation: deleteProjectBySlug
- delete_checkout_key_by_slug:
  - endpoint: DELETE /project/{{ record.project-slug }}/checkout-key/{{ record.fingerprint }}
  - required fields: fingerprint, project-slug
  - risk: external CircleCI API mutation: deleteCheckoutKey
- delete_env_var:
  - endpoint: DELETE /project/{{ record.project-slug }}/envvar/{{ record.name }}
  - required fields: name, project-slug
  - risk: external CircleCI API mutation: deleteEnvVar
- delete_pipeline_definition:
  - endpoint: DELETE /projects/{{ record.project_id }}/pipeline-definitions/{{ record.pipeline_definition_id }}
  - required fields: pipeline_definition_id, project_id
  - risk: external CircleCI API mutation: deletePipelineDefinition
- delete_trigger:
  - endpoint: DELETE /projects/{{ record.project_id }}/triggers/{{ record.trigger_id }}
  - required fields: project_id, trigger_id
  - risk: external CircleCI API mutation: deleteTrigger
- delete_schedule_by_id:
  - endpoint: DELETE /schedule/{{ record.schedule-id }}
  - required fields: schedule-id
  - risk: external CircleCI API mutation: deleteScheduleById
- delete_webhook:
  - endpoint: DELETE /webhook/{{ record.webhook_id }}
  - required fields: webhook_id
  - risk: external CircleCI API mutation: deleteWebhook
- patch_project_settings:
  - endpoint: PATCH /project/{{ record.provider }}/{{ record.organization }}/{{ record.project }}/settings
  - required fields: organization, project, provider
  - optional fields: advanced
  - risk: external CircleCI API mutation: patchProjectSettings
- update_pipeline_definition:
  - endpoint: PATCH /projects/{{ record.project_id }}/pipeline-definitions/{{ record.pipeline_definition_id }}
  - required fields: pipeline_definition_id, project_id
  - optional fields: checkout_source, config_source, description, name
  - risk: external CircleCI API mutation: updatePipelineDefinition
- create_context_restriction:
  - endpoint: POST /context/{{ record.context_id }}/restrictions
  - required fields: context_id
  - optional fields: restriction_type, restriction_value
  - risk: external CircleCI API mutation: createContextRestriction
- cancel_job_by_job_i_d:
  - endpoint: POST /jobs/{{ record.job-id }}/cancel
  - required fields: job-id
  - risk: external CircleCI API mutation: cancelJobByJobID
- create_organization_group:
  - endpoint: POST /organizations/{{ record.org_id }}/groups
  - required fields: org_id, name
  - optional fields: description
  - risk: external CircleCI API mutation: createOrganizationGroup
- create_usage_export:
  - endpoint: POST /organizations/{{ record.org_id }}/usage_export_job
  - required fields: org_id, end, start
  - optional fields: shared_org_ids
  - risk: external CircleCI API mutation: createUsageExport
- create_webhook:
  - endpoint: POST /webhook
  - required fields: events, name, scope, signing-secret, url, verify-tls
  - risk: external CircleCI API mutation: createWebhook
- approve_pending_approval_job_by_id:
  - endpoint: POST /workflow/{{ record.id }}/approve/{{ record.approval_request_id }}
  - required fields: approval_request_id, id
  - risk: external CircleCI API mutation: approvePendingApprovalJobById
- cancel_workflow:
  - endpoint: POST /workflow/{{ record.id }}/cancel
  - required fields: id
  - risk: external CircleCI API mutation: cancelWorkflow
- add_environment_variable_to_context:
  - endpoint: PUT /context/{{ record.context_id }}/environment-variable/{{ record.env_var_name }}
  - required fields: context_id, env_var_name, value
  - risk: external CircleCI API mutation: addEnvironmentVariableToContext
- update_webhook:
  - endpoint: PUT /webhook/{{ record.webhook_id }}
  - required fields: webhook_id
  - optional fields: events, name, signing-secret, url, verify-tls
  - risk: external CircleCI API mutation: updateWebhook

## Security

- read risk: external CircleCI API read of CI project, pipeline, workflow, job, context, schedule, environment-variable, checkout-key, and workflow-insight metadata
- write risk: external mutation of CircleCI project configuration: schedule/environment-variable/checkout-key create and delete; never triggers, cancels, or approves a live CI run
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run CircleCI's declared streams and reverse-ETL actions.
- Usage: pm circleci <command> [flags]
- PM execution policy pm-request-contract-bounds-v1: each max N bytes qualifier is the effective PM request limit, not a provider schema assertion; path/query values are measured after exact wire encoding and rejected rather than truncated.
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Read streams
- Reverse ETL writes
- Other Commands
  - add environment variable to context apply - Plan and execute the add environment variable to context reverse-ETL action [intent=reverse_etl availability=implemented write=add_environment_variable_to_context]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: addEnvironmentVariableToContext; flags: --context-id (required, max 32768 bytes), --env-var-name (required, max 32768 bytes), --value (required, max 32768 bytes)
  - approve pending approval job by id apply - Plan and execute the approve pending approval job by id reverse-ETL action [intent=reverse_etl availability=implemented write=approve_pending_approval_job_by_id]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: approvePendingApprovalJobById; flags: --approval-request-id (required, max 32768 bytes), --id (required, max 32768 bytes)
  - cancel job by job i d apply - Plan and execute the cancel job by job i d reverse-ETL action [intent=reverse_etl availability=implemented write=cancel_job_by_job_i_d]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: cancelJobByJobID; flags: --job-id (required, max 32768 bytes)
  - cancel workflow apply - Plan and execute the cancel workflow reverse-ETL action [intent=reverse_etl availability=implemented write=cancel_workflow]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: cancelWorkflow; flags: --id (required, max 32768 bytes)
  - checkout keys list - Run the checkout keys ETL stream [intent=etl availability=implemented stream=checkout_keys]
  - contexts list - Run the contexts ETL stream [intent=etl availability=implemented stream=contexts]
  - create checkout key apply - Plan and execute the create checkout key reverse-ETL action [intent=reverse_etl availability=implemented write=create_checkout_key]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a new deploy/checkout SSH key with repository access; flags: --type (required)
  - create context restriction apply - Plan and execute the create context restriction reverse-ETL action [intent=reverse_etl availability=implemented write=create_context_restriction]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: createContextRestriction; flags: --context-id (required, max 32768 bytes), --restriction-type, --restriction-value (max 32768 bytes)
  - create environment variable apply - Plan and execute the create environment variable reverse-ETL action [intent=reverse_etl availability=implemented write=create_environment_variable]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates or overwrites a project environment variable used by every future CI run; flags: --name (required), --value (required)
  - create organization group apply - Plan and execute the create organization group reverse-ETL action [intent=reverse_etl availability=implemented write=create_organization_group]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: createOrganizationGroup; flags: --description (max 32768 bytes), --name (required, max 32768 bytes), --org-id (required, max 32768 bytes)
  - create schedule apply - Plan and execute the create schedule reverse-ETL action [intent=reverse_etl availability=implemented write=create_schedule]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a new scheduled-pipeline trigger for this project; flags: --attribution-actor (required), --name (required), --parameters (required), --timetable (required)
  - create usage export apply - Plan and execute the create usage export reverse-ETL action [intent=reverse_etl availability=implemented write=create_usage_export]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: createUsageExport; flags: --end (required, max 32768 bytes), --org-id (required, max 32768 bytes), --shared-org-ids (max 1048576 bytes), --start (required, max 32768 bytes)
  - create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: createWebhook; flags: --events (required, max 1048576 bytes), --name (required, max 32768 bytes), --scope (required, max 1048576 bytes), --signing-secret (required, max 32768 bytes), --url (required, max 32768 bytes), --verify-tls (required)
  - delete checkout key apply - Plan and execute the delete checkout key reverse-ETL action [intent=reverse_etl availability=implemented write=delete_checkout_key]; approval: requires plan, preview, approval, and execute; risk: irreversible external revocation of a deploy/checkout SSH key; may break future CI checkouts that depend on it; approval required; flags: --fingerprint (required)
  - delete checkout key by slug apply - Plan and execute the delete checkout key by slug reverse-ETL action [intent=reverse_etl availability=implemented write=delete_checkout_key_by_slug]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteCheckoutKey; flags: --fingerprint (required, max 32768 bytes), --project-slug (required, max 32768 bytes)
  - delete context apply - Plan and execute the delete context reverse-ETL action [intent=reverse_etl availability=implemented write=delete_context]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteContext; flags: --context-id (required, max 32768 bytes)
  - delete context restriction apply - Plan and execute the delete context restriction reverse-ETL action [intent=reverse_etl availability=implemented write=delete_context_restriction]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteContextRestriction; flags: --context-id (required, max 32768 bytes), --restriction-id (required, max 32768 bytes)
  - delete env var apply - Plan and execute the delete env var reverse-ETL action [intent=reverse_etl availability=implemented write=delete_env_var]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteEnvVar; flags: --name (required, max 32768 bytes), --project-slug (required, max 32768 bytes)
  - delete environment variable apply - Plan and execute the delete environment variable reverse-ETL action [intent=reverse_etl availability=implemented write=delete_environment_variable]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion of a project environment variable; may break future CI runs that depend on it; approval required; flags: --name (required)
  - delete environment variable from context apply - Plan and execute the delete environment variable from context reverse-ETL action [intent=reverse_etl availability=implemented write=delete_environment_variable_from_context]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteEnvironmentVariableFromContext; flags: --context-id (required, max 32768 bytes), --env-var-name (required, max 32768 bytes)
  - delete group apply - Plan and execute the delete group reverse-ETL action [intent=reverse_etl availability=implemented write=delete_group]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteGroup; flags: --group-id (required, max 32768 bytes), --org-id (required, max 32768 bytes)
  - delete org claims apply - Plan and execute the delete org claims reverse-ETL action [intent=reverse_etl availability=implemented write=delete_org_claims]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: DeleteOrgClaims; flags: --claims (required, max 32768 bytes), --orgID (required, max 32768 bytes)
  - delete organization apply - Plan and execute the delete organization reverse-ETL action [intent=reverse_etl availability=implemented write=delete_organization]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteOrganization; flags: --org-slug-or-id (required, max 32768 bytes)
  - delete otel exporter apply - Plan and execute the delete otel exporter reverse-ETL action [intent=reverse_etl availability=implemented write=delete_otel_exporter]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteOtelExporter; flags: --otel-exporter-id (required, max 32768 bytes)
  - delete pipeline definition apply - Plan and execute the delete pipeline definition reverse-ETL action [intent=reverse_etl availability=implemented write=delete_pipeline_definition]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deletePipelineDefinition; flags: --pipeline-definition-id (required, max 32768 bytes), --project-id (required, max 32768 bytes)
  - delete project by slug apply - Plan and execute the delete project by slug reverse-ETL action [intent=reverse_etl availability=implemented write=delete_project_by_slug]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteProjectBySlug; flags: --project-slug (required, max 32768 bytes)
  - delete project claims apply - Plan and execute the delete project claims reverse-ETL action [intent=reverse_etl availability=implemented write=delete_project_claims]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: DeleteProjectClaims; flags: --claims (required, max 32768 bytes), --orgID (required, max 32768 bytes), --projectID (required, max 32768 bytes)
  - delete schedule apply - Plan and execute the delete schedule reverse-ETL action [intent=reverse_etl availability=implemented write=delete_schedule]; approval: requires plan, preview, approval, and execute; risk: irreversible external deletion of a scheduled-pipeline trigger; approval required; flags: --id (required)
  - delete schedule by id apply - Plan and execute the delete schedule by id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_schedule_by_id]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteScheduleById; flags: --schedule-id (required, max 32768 bytes)
  - delete trigger apply - Plan and execute the delete trigger reverse-ETL action [intent=reverse_etl availability=implemented write=delete_trigger]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteTrigger; flags: --project-id (required, max 32768 bytes), --trigger-id (required, max 32768 bytes)
  - delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: deleteWebhook; flags: --webhook-id (required, max 32768 bytes)
  - environment variables list - Run the environment variables ETL stream [intent=etl availability=implemented stream=environment_variables]
  - insights workflow summary list - Run the insights workflow summary ETL stream [intent=etl availability=implemented stream=insights_workflow_summary]
  - jobs list - Run the jobs ETL stream [intent=etl availability=implemented stream=jobs]
  - patch project settings apply - Plan and execute the patch project settings reverse-ETL action [intent=reverse_etl availability=implemented write=patch_project_settings]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: patchProjectSettings; flags: --advanced (max 1048576 bytes), --organization (required, max 32768 bytes), --project (required, max 32768 bytes), --provider (required)
  - pipelines list - Run the pipelines ETL stream [intent=etl availability=implemented stream=pipelines]
  - projects list - Run the projects ETL stream [intent=etl availability=implemented stream=projects]
  - remove u r l orb allow list entry apply - Plan and execute the remove u r l orb allow list entry reverse-ETL action [intent=reverse_etl availability=implemented write=remove_u_r_l_orb_allow_list_entry]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: removeURLOrbAllowListEntry; flags: --allow-list-entry-id (required, max 32768 bytes), --org-slug-or-id (required, max 32768 bytes)
  - schedules list - Run the schedules ETL stream [intent=etl availability=implemented stream=schedules]
  - update pipeline definition apply - Plan and execute the update pipeline definition reverse-ETL action [intent=reverse_etl availability=implemented write=update_pipeline_definition]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: updatePipelineDefinition; flags: --checkout-source (max 1048576 bytes), --config-source (max 1048576 bytes), --description (max 32768 bytes), --name (max 32768 bytes), --pipeline-definition-id (required, max 32768 bytes), --project-id (required, max 32768 bytes)
  - update schedule apply - Plan and execute the update schedule reverse-ETL action [intent=reverse_etl availability=implemented write=update_schedule]; approval: requires plan, preview, approval, and execute; risk: external mutation; updates an existing scheduled-pipeline trigger's timetable or parameters; flags: --id (required)
  - update webhook apply - Plan and execute the update webhook reverse-ETL action [intent=reverse_etl availability=implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: external CircleCI API mutation: updateWebhook; flags: --events (max 1048576 bytes), --name (max 32768 bytes), --signing-secret (max 32768 bytes), --url (max 32768 bytes), --verify-tls, --webhook-id (required, max 32768 bytes)
  - workflows list - Run the workflows ETL stream [intent=etl availability=implemented stream=workflows]

## Sync Transport

- Source transport: declared
- Destination transport: unsupported
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: declarative_api/declarative_stream_source

## Commands

### Inspect as a manual

```bash
pm connectors inspect circleci
```

### Inspect as structured JSON

```bash
pm connectors inspect circleci --json
```

## Agent Rules

- Run pm connectors inspect circleci before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
