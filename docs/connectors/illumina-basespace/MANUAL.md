# pm connectors inspect illumina-basespace

```text
NAME
  pm connectors inspect illumina-basespace - Illumina BaseSpace connector manual

SYNOPSIS
  pm connectors inspect illumina-basespace
  pm connectors inspect illumina-basespace --json
  pm credentials add <name> --connector illumina-basespace [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes documented Illumina BaseSpace v1pre3 REST API resources through the declarative connector engine.

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
  application_id
  appsession_id
  base_url
  biosample_id
  dataset_id
  datasettype_id
  labrequeue_id
  lane_id
  librarypool_id
  name
  page_size
  project_id
  run_id
  trash_id
  user
  user_id
  workgroup_id
  access_token (secret)

ETL STREAMS
  projects:
    primary key: id
    cursor: date_created
    fields: date_created(), date_modified(), href(), id(), name(), total_size(), user_owned_by()
  runs:
    primary key: id
    cursor: date_created
    fields: date_created(), date_modified(), experiment_name(), href(), id(), instrument_name(), name(), status(), total_size()
  samples:
    primary key: id
    cursor: date_created
    fields: date_created(), href(), id(), name(), num_reads_pf(), num_reads_raw(), sample_id(), status(), total_size()
  appsessions:
    primary key: id
    cursor: date_created
    fields: application(), date_completed(), date_created(), href(), id(), name(), status(), status_summary(), total_size()
  datasets:
    primary key: id
    cursor: date_created
    fields: dataset_type(), date_created(), href(), id(), name(), project(), total_size()
  applications:
    fields: Id(), id()
  application:
    fields: Id(), id()
  application_qcthresholds:
    fields: Id(), id()
  application_settings:
    fields: Id(), id()
  application_workflowdependencies:
    fields: Id(), id()
  appsessions_all:
    fields: Id(), id()
  appsession:
    fields: Id(), id()
  appsession_comments:
    fields: Id(), id()
  appsessions_logfiles:
    fields: Id(), id()
  appsession_properties:
    fields: Id(), id()
  appsession_property:
    fields: Id(), id()
  appsession_property_items:
    fields: Id(), id()
  biosamples:
    fields: Id(), id()
  biosample_labrequeues:
    fields: Id(), id()
  biosample:
    fields: Id(), id()
  biosample_libraries:
    fields: Id(), id()
  biosample_runlane_summaries:
    fields: Id(), id()
  datasets_all:
    fields: Id(), id()
  dataset:
    fields: Id(), id()
  dataset_comments:
    fields: Id(), id()
  dataset_files:
    fields: Id(), id()
  datasettype:
    fields: Id(), id()
  instrumentstatistics:
    fields: Id(), id()
  labrequeues:
    fields: Id(), id()
  labrequeue:
    fields: Id(), id()
  laneqcthresholds:
    fields: Id(), id()
  lane:
    fields: Id(), id()
  lane_comments:
    fields: Id(), id()
  libraries:
    fields: Id(), id()
  librarypool_libraries:
    fields: Id(), id()
  project:
    fields: Id(), id()
  project_datasets:
    fields: Id(), id()
  run_files:
    fields: Id(), id()
  trash:
    fields: Id(), id()
  trash_2:
    fields: Id(), id()
  current_user:
    fields: Id(), id()
  current_user_subscription:
    fields: Id(), id()
  current_user_usage:
    fields: Id(), id()
  current_user_workgroups:
    fields: Id(), id()
  user:
    fields: Id(), id()
  user_settings:
    fields: Id(), id()
  workgroup:
    fields: Id(), id()
  configured_user:
    fields: Id(), id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  update_applications_id_qcthresholds:
    endpoint: PUT /v1pre3/applications/{{ record.application_id }}/qcthresholds
    required fields: application_id
    risk: BaseSpace mutation: PUT /v1pre3/applications/{application_id}/qcthresholds.
  update_applications_id_workflowdependencies:
    endpoint: PUT /v1pre3/applications/{{ record.application_id }}/workflowdependencies
    required fields: application_id
    risk: BaseSpace mutation: PUT /v1pre3/applications/{application_id}/workflowdependencies.
  applications_id_workflows:
    endpoint: POST /v1pre3/applications/{{ record.application_id }}/workflows
    required fields: application_id
    risk: BaseSpace mutation: POST /v1pre3/applications/{application_id}/workflows.
  delete_appsessions_id:
    endpoint: DELETE /v1pre3/appsessions/{{ record.appsession_id }}
    required fields: appsession_id
    risk: Destructive BaseSpace mutation: DELETE /v1pre3/appsessions/{appsession_id}.
  appsessions_id:
    endpoint: POST /v1pre3/appsessions/{{ record.appsession_id }}
    required fields: appsession_id
    risk: BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}.
  appsessions_id_properties:
    endpoint: POST /v1pre3/appsessions/{{ record.appsession_id }}/properties
    required fields: appsession_id
    risk: BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}/properties.
  delete_appsessions_id_properties_name:
    endpoint: DELETE /v1pre3/appsessions/{{ record.appsession_id }}/properties/{{ record.name }}
    required fields: appsession_id, name
    risk: Destructive BaseSpace mutation: DELETE /v1pre3/appsessions/{appsession_id}/properties/{name}.
  appsessions_id_stop:
    endpoint: POST /v1pre3/appsessions/{{ record.appsession_id }}/stop
    required fields: appsession_id
    risk: Destructive BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}/stop.
  biosamples_bulkupdate:
    endpoint: POST /v1pre3/biosamples/bulkupdate
    risk: BaseSpace mutation: POST /v1pre3/biosamples/bulkupdate.
  biosamples_id:
    endpoint: POST /v1pre3/biosamples/{{ record.biosample_id }}
    required fields: biosample_id
    risk: BaseSpace mutation: POST /v1pre3/biosamples/{biosample_id}.
  datasets_id:
    endpoint: POST /v1pre3/datasets/{{ record.dataset_id }}
    required fields: dataset_id
    risk: BaseSpace mutation: POST /v1pre3/datasets/{dataset_id}.
  update_laneqcthresholds:
    endpoint: PUT /v1pre3/laneqcthresholds
    risk: BaseSpace mutation: PUT /v1pre3/laneqcthresholds.
  lanes_id:
    endpoint: POST /v1pre3/lanes/{{ record.lane_id }}
    required fields: lane_id
    risk: BaseSpace mutation: POST /v1pre3/lanes/{lane_id}.
  libraries_libraryid_labrequeues:
    endpoint: POST /v1pre3/libraries/{{ record.library_id }}/labrequeues
    required fields: library_id
    risk: BaseSpace mutation: POST /v1pre3/libraries/{library_id}/labrequeues.
  librarypools_id:
    endpoint: POST /v1pre3/librarypools/{{ record.librarypool_id }}
    required fields: librarypool_id
    risk: BaseSpace mutation: POST /v1pre3/librarypools/{librarypool_id}.
  librarypools_poolid_labrequeues:
    endpoint: POST /v1pre3/librarypools/{{ record.pool_id }}/labrequeues
    required fields: pool_id
    risk: BaseSpace mutation: POST /v1pre3/librarypools/{pool_id}/labrequeues.
  preprequests_preprequestid_labrequeues:
    endpoint: POST /v1pre3/preprequests/{{ record.preprequest_id }}/labrequeues
    required fields: preprequest_id
    risk: BaseSpace mutation: POST /v1pre3/preprequests/{preprequest_id}/labrequeues.
  delete_trash:
    endpoint: DELETE /v1pre3/trash
    risk: Destructive BaseSpace mutation: DELETE /v1pre3/trash.
  trash_id_restorefromtrash:
    endpoint: POST /v1pre3/trash/{{ record.trash_id }}/restorefromtrash
    required fields: trash_id
    risk: Destructive BaseSpace mutation: POST /v1pre3/trash/{trash_id}/restorefromtrash.
  users_id_settings:
    endpoint: POST /v1pre3/users/{{ record.user_id }}/settings
    required fields: user_id
    risk: BaseSpace mutation: POST /v1pre3/users/{user_id}/settings.

SECURITY
  read risk: external Illumina BaseSpace API read of documented v1pre3 resources
  write risk: external Illumina BaseSpace API mutations including settings, workflow, threshold, trash, lab requeue, and app session actions
  approval: required for every write action; destructive actions are marked confirm: destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect illumina-basespace

  # Inspect as structured JSON
  pm connectors inspect illumina-basespace --json

AGENT WORKFLOW
  - Run pm connectors inspect illumina-basespace before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
