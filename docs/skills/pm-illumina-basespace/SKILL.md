---
name: pm-illumina-basespace
description: Illumina BaseSpace connector knowledge and safe action guide.
---

# pm-illumina-basespace

## Purpose

Reads and writes documented Illumina BaseSpace v1pre3 REST API resources through the declarative connector engine.

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

- application_id
- appsession_id
- base_url (required)
- biosample_id
- dataset_id
- datasettype_id
- labrequeue_id
- lane_id
- librarypool_id
- name
- page_size
- project_id
- run_id
- trash_id
- user
- user_id
- workgroup_id
- access_token (secret) (required)

## ETL Streams

- projects:
  - primary key: id
  - cursor: date_created
  - fields: date_created(string), date_modified(string), href(string), id(string), name(string), total_size(integer), user_owned_by(object)
- runs:
  - primary key: id
  - cursor: date_created
  - fields: date_created(string), date_modified(string), experiment_name(string), href(string), id(string), instrument_name(string), name(string), status(string), total_size(integer)
- samples:
  - primary key: id
  - cursor: date_created
  - fields: date_created(string), href(string), id(string), name(string), num_reads_pf(integer), num_reads_raw(integer), sample_id(string), status(string), total_size(integer)
- appsessions:
  - primary key: id
  - cursor: date_created
  - fields: application(object), date_completed(string), date_created(string), href(string), id(string), name(string), status(string), status_summary(string), total_size(integer)
- datasets:
  - primary key: id
  - cursor: date_created
  - fields: dataset_type(object), date_created(string), href(string), id(string), name(string), project(object), total_size(integer)
- applications:
  - fields: Id(string), id(string)
- application:
  - fields: Id(string), id(string)
- application_qcthresholds:
  - fields: Id(string), id(string)
- application_settings:
  - fields: Id(string), id(string)
- application_workflowdependencies:
  - fields: Id(string), id(string)
- appsessions_all:
  - fields: Id(string), id(string)
- appsession:
  - fields: Id(string), id(string)
- appsession_comments:
  - fields: Id(string), id(string)
- appsessions_logfiles:
  - fields: Id(string), id(string)
- appsession_properties:
  - fields: Id(string), id(string)
- appsession_property:
  - fields: Id(string), id(string)
- appsession_property_items:
  - fields: Id(string), id(string)
- biosamples:
  - fields: Id(string), id(string)
- biosample_labrequeues:
  - fields: Id(string), id(string)
- biosample:
  - fields: Id(string), id(string)
- biosample_libraries:
  - fields: Id(string), id(string)
- biosample_runlane_summaries:
  - fields: Id(string), id(string)
- datasets_all:
  - fields: Id(string), id(string)
- dataset:
  - fields: Id(string), id(string)
- dataset_comments:
  - fields: Id(string), id(string)
- dataset_files:
  - fields: Id(string), id(string)
- datasettype:
  - fields: Id(string), id(string)
- instrumentstatistics:
  - fields: Id(string), id(string)
- labrequeues:
  - fields: Id(string), id(string)
- labrequeue:
  - fields: Id(string), id(string)
- laneqcthresholds:
  - fields: Id(string), id(string)
- lane:
  - fields: Id(string), id(string)
- lane_comments:
  - fields: Id(string), id(string)
- libraries:
  - fields: Id(string), id(string)
- librarypool_libraries:
  - fields: Id(string), id(string)
- project:
  - fields: Id(string), id(string)
- project_datasets:
  - fields: Id(string), id(string)
- run_files:
  - fields: Id(string), id(string)
- trash:
  - fields: Id(string), id(string)
- trash_2:
  - fields: Id(string), id(string)
- current_user:
  - fields: Id(string), id(string)
- current_user_subscription:
  - fields: Id(string), id(string)
- current_user_usage:
  - fields: Id(string), id(string)
- current_user_workgroups:
  - fields: Id(string), id(string)
- user:
  - fields: Id(string), id(string)
- user_settings:
  - fields: Id(string), id(string)
- workgroup:
  - fields: Id(string), id(string)
- configured_user:
  - fields: Id(string), id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- update_applications_id_qcthresholds:
  - endpoint: PUT /v1pre3/applications/{{ record.application_id }}/qcthresholds
  - required fields: application_id
  - risk: BaseSpace mutation: PUT /v1pre3/applications/{application_id}/qcthresholds.
- update_applications_id_workflowdependencies:
  - endpoint: PUT /v1pre3/applications/{{ record.application_id }}/workflowdependencies
  - required fields: application_id
  - risk: BaseSpace mutation: PUT /v1pre3/applications/{application_id}/workflowdependencies.
- applications_id_workflows:
  - endpoint: POST /v1pre3/applications/{{ record.application_id }}/workflows
  - required fields: application_id
  - risk: BaseSpace mutation: POST /v1pre3/applications/{application_id}/workflows.
- delete_appsessions_id:
  - endpoint: DELETE /v1pre3/appsessions/{{ record.appsession_id }}
  - required fields: appsession_id
  - risk: Destructive BaseSpace mutation: DELETE /v1pre3/appsessions/{appsession_id}.
- appsessions_id:
  - endpoint: POST /v1pre3/appsessions/{{ record.appsession_id }}
  - required fields: appsession_id
  - risk: BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}.
- appsessions_id_properties:
  - endpoint: POST /v1pre3/appsessions/{{ record.appsession_id }}/properties
  - required fields: appsession_id
  - risk: BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}/properties.
- delete_appsessions_id_properties_name:
  - endpoint: DELETE /v1pre3/appsessions/{{ record.appsession_id }}/properties/{{ record.name }}
  - required fields: appsession_id, name
  - risk: Destructive BaseSpace mutation: DELETE /v1pre3/appsessions/{appsession_id}/properties/{name}.
- appsessions_id_stop:
  - endpoint: POST /v1pre3/appsessions/{{ record.appsession_id }}/stop
  - required fields: appsession_id
  - risk: Destructive BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}/stop.
- biosamples_bulkupdate:
  - endpoint: POST /v1pre3/biosamples/bulkupdate
  - risk: BaseSpace mutation: POST /v1pre3/biosamples/bulkupdate.
- biosamples_id:
  - endpoint: POST /v1pre3/biosamples/{{ record.biosample_id }}
  - required fields: biosample_id
  - risk: BaseSpace mutation: POST /v1pre3/biosamples/{biosample_id}.
- datasets_id:
  - endpoint: POST /v1pre3/datasets/{{ record.dataset_id }}
  - required fields: dataset_id
  - risk: BaseSpace mutation: POST /v1pre3/datasets/{dataset_id}.
- update_laneqcthresholds:
  - endpoint: PUT /v1pre3/laneqcthresholds
  - risk: BaseSpace mutation: PUT /v1pre3/laneqcthresholds.
- lanes_id:
  - endpoint: POST /v1pre3/lanes/{{ record.lane_id }}
  - required fields: lane_id
  - risk: BaseSpace mutation: POST /v1pre3/lanes/{lane_id}.
- libraries_libraryid_labrequeues:
  - endpoint: POST /v1pre3/libraries/{{ record.library_id }}/labrequeues
  - required fields: library_id
  - risk: BaseSpace mutation: POST /v1pre3/libraries/{library_id}/labrequeues.
- librarypools_id:
  - endpoint: POST /v1pre3/librarypools/{{ record.librarypool_id }}
  - required fields: librarypool_id
  - risk: BaseSpace mutation: POST /v1pre3/librarypools/{librarypool_id}.
- librarypools_poolid_labrequeues:
  - endpoint: POST /v1pre3/librarypools/{{ record.pool_id }}/labrequeues
  - required fields: pool_id
  - risk: BaseSpace mutation: POST /v1pre3/librarypools/{pool_id}/labrequeues.
- preprequests_preprequestid_labrequeues:
  - endpoint: POST /v1pre3/preprequests/{{ record.preprequest_id }}/labrequeues
  - required fields: preprequest_id
  - risk: BaseSpace mutation: POST /v1pre3/preprequests/{preprequest_id}/labrequeues.
- delete_trash:
  - endpoint: DELETE /v1pre3/trash
  - risk: Destructive BaseSpace mutation: DELETE /v1pre3/trash.
- trash_id_restorefromtrash:
  - endpoint: POST /v1pre3/trash/{{ record.trash_id }}/restorefromtrash
  - required fields: trash_id
  - risk: Destructive BaseSpace mutation: POST /v1pre3/trash/{trash_id}/restorefromtrash.
- users_id_settings:
  - endpoint: POST /v1pre3/users/{{ record.user_id }}/settings
  - required fields: user_id
  - risk: BaseSpace mutation: POST /v1pre3/users/{user_id}/settings.

## Security

- read risk: external Illumina BaseSpace API read of documented v1pre3 resources
- write risk: external Illumina BaseSpace API mutations including settings, workflow, threshold, trash, lab requeue, and app session actions
- approval: required for every write action; destructive actions are marked confirm: destructive
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect illumina-basespace
```

### Inspect as structured JSON

```bash
pm connectors inspect illumina-basespace --json
```

## Agent Rules

- Run pm connectors inspect illumina-basespace before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
