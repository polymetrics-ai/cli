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
    fields: date_created(string), date_modified(string), href(string), id(string), name(string), total_size(integer), user_owned_by(object)
  runs:
    primary key: id
    cursor: date_created
    fields: date_created(string), date_modified(string), experiment_name(string), href(string), id(string), instrument_name(string), name(string), status(string), total_size(integer)
  samples:
    primary key: id
    cursor: date_created
    fields: date_created(string), href(string), id(string), name(string), num_reads_pf(integer), num_reads_raw(integer), sample_id(string), status(string), total_size(integer)
  appsessions:
    primary key: id
    cursor: date_created
    fields: application(object), date_completed(string), date_created(string), href(string), id(string), name(string), status(string), status_summary(string), total_size(integer)
  datasets:
    primary key: id
    cursor: date_created
    fields: dataset_type(object), date_created(string), href(string), id(string), name(string), project(object), total_size(integer)
  applications:
    fields: Id(string), id(string)
  application:
    fields: Id(string), id(string)
  application_qcthresholds:
    fields: Id(string), id(string)
  application_settings:
    fields: Id(string), id(string)
  application_workflowdependencies:
    fields: Id(string), id(string)
  appsessions_all:
    fields: Id(string), id(string)
  appsession:
    fields: Id(string), id(string)
  appsession_comments:
    fields: Id(string), id(string)
  appsessions_logfiles:
    fields: Id(string), id(string)
  appsession_properties:
    fields: Id(string), id(string)
  appsession_property:
    fields: Id(string), id(string)
  appsession_property_items:
    fields: Id(string), id(string)
  biosamples:
    fields: Id(string), id(string)
  biosample_labrequeues:
    fields: Id(string), id(string)
  biosample:
    fields: Id(string), id(string)
  biosample_libraries:
    fields: Id(string), id(string)
  biosample_runlane_summaries:
    fields: Id(string), id(string)
  datasets_all:
    fields: Id(string), id(string)
  dataset:
    fields: Id(string), id(string)
  dataset_comments:
    fields: Id(string), id(string)
  dataset_files:
    fields: Id(string), id(string)
  datasettype:
    fields: Id(string), id(string)
  instrumentstatistics:
    fields: Id(string), id(string)
  labrequeues:
    fields: Id(string), id(string)
  labrequeue:
    fields: Id(string), id(string)
  laneqcthresholds:
    fields: Id(string), id(string)
  lane:
    fields: Id(string), id(string)
  lane_comments:
    fields: Id(string), id(string)
  libraries:
    fields: Id(string), id(string)
  librarypool_libraries:
    fields: Id(string), id(string)
  project:
    fields: Id(string), id(string)
  project_datasets:
    fields: Id(string), id(string)
  run_files:
    fields: Id(string), id(string)
  trash:
    fields: Id(string), id(string)
  trash_2:
    fields: Id(string), id(string)
  current_user:
    fields: Id(string), id(string)
  current_user_subscription:
    fields: Id(string), id(string)
  current_user_usage:
    fields: Id(string), id(string)
  current_user_workgroups:
    fields: Id(string), id(string)
  user:
    fields: Id(string), id(string)
  user_settings:
    fields: Id(string), id(string)
  workgroup:
    fields: Id(string), id(string)
  configured_user:
    fields: Id(string), id(string)

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

COMMAND SURFACE
  Run Illumina BaseSpace's declared streams and reverse-ETL actions.
  Usage: pm illumina-basespace <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete appsessions id - Documented DELETE /appsessions/{id} (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.delete.appsessions-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete appsessions id properties name - Documented DELETE /appsessions/{id}/properties/{name} (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.delete.appsessions-id-properties-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete trash - Documented DELETE /trash (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.delete.trash]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get applications - Documented GET /applications (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.applications]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get applications id - Documented GET /applications/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.applications-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get applications id qcthresholds - Documented GET /applications/{id}/qcthresholds (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.applications-id-qcthresholds]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get applications id settings - Documented GET /applications/{id}/settings (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.applications-id-settings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get applications id workflowdependencies - Documented GET /applications/{id}/workflowdependencies (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.applications-id-workflowdependencies]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get appsessions - Documented GET /appsessions (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.appsessions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get appsessions id - Documented GET /appsessions/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.appsessions-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get appsessions id comments - Documented GET /appsessions/{id}/comments (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.appsessions-id-comments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get appsessions id logfiles - Documented GET /appsessions/{id}/logfiles (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.appsessions-id-logfiles]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get appsessions id properties - Documented GET /appsessions/{id}/properties (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.appsessions-id-properties]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get appsessions id properties name - Documented GET /appsessions/{id}/properties/{name} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.appsessions-id-properties-name]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get appsessions id properties name items - Documented GET /appsessions/{id}/properties/{name}/items (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.appsessions-id-properties-name-items]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get biosamples - Documented GET /biosamples (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.biosamples]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get biosamples biosampleid labrequeues - Documented GET /biosamples/{biosampleid}/labrequeues (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.biosamples-biosampleid-labrequeues]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get biosamples id - Documented GET /biosamples/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.biosamples-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get biosamples id libraries - Documented GET /biosamples/{id}/libraries (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.biosamples-id-libraries]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get biosamples id runlanesummaries - Documented GET /biosamples/{id}/runlanesummaries (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.biosamples-id-runlanesummaries]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get datasets - Documented GET /datasets (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.datasets]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get datasets id - Documented GET /datasets/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.datasets-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get datasets id comments - Documented GET /datasets/{id}/comments (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.datasets-id-comments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get datasets id files - Documented GET /datasets/{id}/files (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.datasets-id-files]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get datasettypes id - Documented GET /datasettypes/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.datasettypes-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get instrumentstatistics - Documented GET /instrumentstatistics (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.instrumentstatistics]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get labrequeues - Documented GET /labrequeues (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.labrequeues]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get labrequeues id - Documented GET /labrequeues/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.labrequeues-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get laneqcthresholds - Documented GET /laneqcthresholds (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.laneqcthresholds]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get lanes id - Documented GET /lanes/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.lanes-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get lanes id comments - Documented GET /lanes/{id}/comments (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.lanes-id-comments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get libraries - Documented GET /libraries (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.libraries]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get librarypools id libraries - Documented GET /librarypools/{id}/libraries (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.librarypools-id-libraries]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get projects id - Documented GET /projects/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.projects-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get projects id datasets - Documented GET /projects/{id}/datasets (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.projects-id-datasets]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get runs id files - Documented GET /runs/{id}/files (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.runs-id-files]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get trash - Documented GET /trash (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.trash]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get trash id - Documented GET /trash/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.trash-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get users current - Documented GET /users/current (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.users-current]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get users current subscription - Documented GET /users/current/subscription (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.users-current-subscription]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get users current usage - Documented GET /users/current/usage (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.users-current-usage]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get users current workgroups - Documented GET /users/current/workgroups (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.users-current-workgroups]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get users id - Documented GET /users/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.users-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get users id settings - Documented GET /users/{id}/settings (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.users-id-settings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 trash - Documented GET /v2/trash (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.v2-trash]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get workgroups id - Documented GET /workgroups/{id} (not implemented) [intent=direct_read availability=not_implemented operation=illumina-basespace.get.workgroups-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post applications id workflows - Documented POST /applications/{id}/workflows (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.applications-id-workflows]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post appsessions id - Documented POST /appsessions/{id} (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.appsessions-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post appsessions id properties - Documented POST /appsessions/{id}/properties (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.appsessions-id-properties]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post appsessions id stop - Documented POST /appsessions/{id}/stop (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.appsessions-id-stop]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post biosamples bulkupdate - Documented POST /biosamples/bulkupdate (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.biosamples-bulkupdate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post biosamples id - Documented POST /biosamples/{id} (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.biosamples-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post datasets id - Documented POST /datasets/{id} (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.datasets-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post lanes id - Documented POST /lanes/{id} (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.lanes-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post libraries libraryid labrequeues - Documented POST /libraries/{libraryid}/labrequeues (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.libraries-libraryid-labrequeues]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post librarypools id - Documented POST /librarypools/{id} (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.librarypools-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post librarypools poolid labrequeues - Documented POST /librarypools/{poolid}/labrequeues (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.librarypools-poolid-labrequeues]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post preprequests preprequestid labrequeues - Documented POST /preprequests/{preprequestid}/labrequeues (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.preprequests-preprequestid-labrequeues]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post trash id restorefromtrash - Documented POST /trash/{id}/restorefromtrash (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.trash-id-restorefromtrash]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post users id settings - Documented POST /users/{id}/settings (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.post.users-id-settings]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put applications id qcthresholds - Documented PUT /applications/{id}/qcthresholds (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.put.applications-id-qcthresholds]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put applications id workflowdependencies - Documented PUT /applications/{id}/workflowdependencies (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.put.applications-id-workflowdependencies]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put laneqcthresholds - Documented PUT /laneqcthresholds (not implemented) [intent=direct_write availability=not_implemented operation=illumina-basespace.put.laneqcthresholds]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    application list - Run the application ETL stream [intent=etl availability=implemented stream=application]; notes: discrepancy=present-in-surface-absent-from-artifact
    application qcthresholds list - Run the application qcthresholds ETL stream [intent=etl availability=implemented stream=application_qcthresholds]; notes: discrepancy=present-in-surface-absent-from-artifact
    application settings list - Run the application settings ETL stream [intent=etl availability=implemented stream=application_settings]; notes: discrepancy=present-in-surface-absent-from-artifact
    application workflowdependencies list - Run the application workflowdependencies ETL stream [intent=etl availability=implemented stream=application_workflowdependencies]; notes: discrepancy=present-in-surface-absent-from-artifact
    applications id workflows apply - Plan and execute the applications id workflows reverse-ETL action [intent=reverse_etl availability=not_implemented write=applications_id_workflows]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/applications/{application_id}/workflows.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    applications list - Run the applications ETL stream [intent=etl availability=implemented stream=applications]; notes: discrepancy=present-in-surface-absent-from-artifact
    appsession comments list - Run the appsession comments ETL stream [intent=etl availability=implemented stream=appsession_comments]; notes: discrepancy=present-in-surface-absent-from-artifact
    appsession list - Run the appsession ETL stream [intent=etl availability=implemented stream=appsession]; notes: discrepancy=present-in-surface-absent-from-artifact
    appsession properties list - Run the appsession properties ETL stream [intent=etl availability=implemented stream=appsession_properties]; notes: discrepancy=present-in-surface-absent-from-artifact
    appsession property items list - Run the appsession property items ETL stream [intent=etl availability=implemented stream=appsession_property_items]; notes: discrepancy=present-in-surface-absent-from-artifact
    appsession property list - Run the appsession property ETL stream [intent=etl availability=implemented stream=appsession_property]; notes: discrepancy=present-in-surface-absent-from-artifact
    appsessions all list - Run the appsessions all ETL stream [intent=etl availability=implemented stream=appsessions_all]; notes: discrepancy=present-in-surface-absent-from-artifact
    appsessions id apply - Plan and execute the appsessions id reverse-ETL action [intent=reverse_etl availability=not_implemented write=appsessions_id]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    appsessions id properties apply - Plan and execute the appsessions id properties reverse-ETL action [intent=reverse_etl availability=not_implemented write=appsessions_id_properties]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}/properties.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    appsessions id stop apply - Plan and execute the appsessions id stop reverse-ETL action [intent=reverse_etl availability=not_implemented write=appsessions_id_stop]; approval: requires plan, preview, approval, and execute; risk: Destructive BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}/stop.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    appsessions list - Run the appsessions ETL stream [intent=etl availability=implemented stream=appsessions]; notes: discrepancy=present-in-surface-absent-from-artifact
    appsessions logfiles list - Run the appsessions logfiles ETL stream [intent=etl availability=implemented stream=appsessions_logfiles]; notes: discrepancy=present-in-surface-absent-from-artifact
    biosample labrequeues list - Run the biosample labrequeues ETL stream [intent=etl availability=implemented stream=biosample_labrequeues]; notes: discrepancy=present-in-surface-absent-from-artifact
    biosample libraries list - Run the biosample libraries ETL stream [intent=etl availability=implemented stream=biosample_libraries]; notes: discrepancy=present-in-surface-absent-from-artifact
    biosample list - Run the biosample ETL stream [intent=etl availability=implemented stream=biosample]; notes: discrepancy=present-in-surface-absent-from-artifact
    biosample runlane summaries list - Run the biosample runlane summaries ETL stream [intent=etl availability=implemented stream=biosample_runlane_summaries]; notes: discrepancy=present-in-surface-absent-from-artifact
    biosamples bulkupdate apply - Plan and execute the biosamples bulkupdate reverse-ETL action [intent=reverse_etl availability=implemented write=biosamples_bulkupdate]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/biosamples/bulkupdate.
    biosamples id apply - Plan and execute the biosamples id reverse-ETL action [intent=reverse_etl availability=not_implemented write=biosamples_id]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/biosamples/{biosample_id}.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    biosamples list - Run the biosamples ETL stream [intent=etl availability=implemented stream=biosamples]; notes: discrepancy=present-in-surface-absent-from-artifact
    configured user list - Run the configured user ETL stream [intent=etl availability=implemented stream=configured_user]; notes: discrepancy=present-in-surface-absent-from-artifact
    current user list - Run the current user ETL stream [intent=etl availability=implemented stream=current_user]; notes: discrepancy=present-in-surface-absent-from-artifact
    current user subscription list - Run the current user subscription ETL stream [intent=etl availability=implemented stream=current_user_subscription]; notes: discrepancy=present-in-surface-absent-from-artifact
    current user usage list - Run the current user usage ETL stream [intent=etl availability=implemented stream=current_user_usage]; notes: discrepancy=present-in-surface-absent-from-artifact
    current user workgroups list - Run the current user workgroups ETL stream [intent=etl availability=implemented stream=current_user_workgroups]; notes: discrepancy=present-in-surface-absent-from-artifact
    dataset comments list - Run the dataset comments ETL stream [intent=etl availability=implemented stream=dataset_comments]; notes: discrepancy=present-in-surface-absent-from-artifact
    dataset files list - Run the dataset files ETL stream [intent=etl availability=implemented stream=dataset_files]; notes: discrepancy=present-in-surface-absent-from-artifact
    dataset list - Run the dataset ETL stream [intent=etl availability=implemented stream=dataset]; notes: discrepancy=present-in-surface-absent-from-artifact
    datasets all list - Run the datasets all ETL stream [intent=etl availability=implemented stream=datasets_all]; notes: discrepancy=present-in-surface-absent-from-artifact
    datasets id apply - Plan and execute the datasets id reverse-ETL action [intent=reverse_etl availability=not_implemented write=datasets_id]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/datasets/{dataset_id}.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    datasets list - Run the datasets ETL stream [intent=etl availability=implemented stream=datasets]; notes: discrepancy=present-in-surface-absent-from-artifact
    datasettype list - Run the datasettype ETL stream [intent=etl availability=implemented stream=datasettype]; notes: discrepancy=present-in-surface-absent-from-artifact
    delete appsessions id apply - Plan and execute the delete appsessions id reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_appsessions_id]; approval: requires plan, preview, approval, and execute; risk: Destructive BaseSpace mutation: DELETE /v1pre3/appsessions/{appsession_id}.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete appsessions id properties name apply - Plan and execute the delete appsessions id properties name reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_appsessions_id_properties_name]; approval: requires plan, preview, approval, and execute; risk: Destructive BaseSpace mutation: DELETE /v1pre3/appsessions/{appsession_id}/properties/{name}.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete trash apply - Plan and execute the delete trash reverse-ETL action [intent=reverse_etl availability=implemented write=delete_trash]; approval: requires plan, preview, approval, and execute; risk: Destructive BaseSpace mutation: DELETE /v1pre3/trash.
    instrumentstatistics list - Run the instrumentstatistics ETL stream [intent=etl availability=implemented stream=instrumentstatistics]; notes: discrepancy=present-in-surface-absent-from-artifact
    labrequeue list - Run the labrequeue ETL stream [intent=etl availability=implemented stream=labrequeue]; notes: discrepancy=present-in-surface-absent-from-artifact
    labrequeues list - Run the labrequeues ETL stream [intent=etl availability=implemented stream=labrequeues]; notes: discrepancy=present-in-surface-absent-from-artifact
    lane comments list - Run the lane comments ETL stream [intent=etl availability=implemented stream=lane_comments]; notes: discrepancy=present-in-surface-absent-from-artifact
    lane list - Run the lane ETL stream [intent=etl availability=implemented stream=lane]; notes: discrepancy=present-in-surface-absent-from-artifact
    laneqcthresholds list - Run the laneqcthresholds ETL stream [intent=etl availability=implemented stream=laneqcthresholds]; notes: discrepancy=present-in-surface-absent-from-artifact
    lanes id apply - Plan and execute the lanes id reverse-ETL action [intent=reverse_etl availability=not_implemented write=lanes_id]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/lanes/{lane_id}.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    libraries libraryid labrequeues apply - Plan and execute the libraries libraryid labrequeues reverse-ETL action [intent=reverse_etl availability=not_implemented write=libraries_libraryid_labrequeues]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/libraries/{library_id}/labrequeues.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    libraries list - Run the libraries ETL stream [intent=etl availability=implemented stream=libraries]; notes: discrepancy=present-in-surface-absent-from-artifact
    librarypool libraries list - Run the librarypool libraries ETL stream [intent=etl availability=implemented stream=librarypool_libraries]; notes: discrepancy=present-in-surface-absent-from-artifact
    librarypools id apply - Plan and execute the librarypools id reverse-ETL action [intent=reverse_etl availability=not_implemented write=librarypools_id]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/librarypools/{librarypool_id}.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    librarypools poolid labrequeues apply - Plan and execute the librarypools poolid labrequeues reverse-ETL action [intent=reverse_etl availability=not_implemented write=librarypools_poolid_labrequeues]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/librarypools/{pool_id}/labrequeues.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    preprequests preprequestid labrequeues apply - Plan and execute the preprequests preprequestid labrequeues reverse-ETL action [intent=reverse_etl availability=not_implemented write=preprequests_preprequestid_labrequeues]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/preprequests/{preprequest_id}/labrequeues.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    project datasets list - Run the project datasets ETL stream [intent=etl availability=implemented stream=project_datasets]; notes: discrepancy=present-in-surface-absent-from-artifact
    project list - Run the project ETL stream [intent=etl availability=implemented stream=project]; notes: discrepancy=present-in-surface-absent-from-artifact
    projects list - Run the projects ETL stream [intent=etl availability=implemented stream=projects]; notes: discrepancy=present-in-surface-absent-from-artifact
    run files list - Run the run files ETL stream [intent=etl availability=implemented stream=run_files]; notes: discrepancy=present-in-surface-absent-from-artifact
    runs list - Run the runs ETL stream [intent=etl availability=implemented stream=runs]; notes: discrepancy=present-in-surface-absent-from-artifact
    samples list - Run the samples ETL stream [intent=etl availability=implemented stream=samples]; notes: discrepancy=present-in-surface-absent-from-artifact
    trash 2 list - Run the trash 2 ETL stream [intent=etl availability=implemented stream=trash_2]; notes: discrepancy=present-in-surface-absent-from-artifact
    trash id restorefromtrash apply - Plan and execute the trash id restorefromtrash reverse-ETL action [intent=reverse_etl availability=not_implemented write=trash_id_restorefromtrash]; approval: requires plan, preview, approval, and execute; risk: Destructive BaseSpace mutation: POST /v1pre3/trash/{trash_id}/restorefromtrash.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    trash list - Run the trash ETL stream [intent=etl availability=implemented stream=trash]; notes: discrepancy=present-in-surface-absent-from-artifact
    update applications id qcthresholds apply - Plan and execute the update applications id qcthresholds reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_applications_id_qcthresholds]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: PUT /v1pre3/applications/{application_id}/qcthresholds.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update applications id workflowdependencies apply - Plan and execute the update applications id workflowdependencies reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_applications_id_workflowdependencies]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: PUT /v1pre3/applications/{application_id}/workflowdependencies.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update laneqcthresholds apply - Plan and execute the update laneqcthresholds reverse-ETL action [intent=reverse_etl availability=implemented write=update_laneqcthresholds]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: PUT /v1pre3/laneqcthresholds.
    user list - Run the user ETL stream [intent=etl availability=implemented stream=user]; notes: discrepancy=present-in-surface-absent-from-artifact
    user settings list - Run the user settings ETL stream [intent=etl availability=implemented stream=user_settings]; notes: discrepancy=present-in-surface-absent-from-artifact
    users id settings apply - Plan and execute the users id settings reverse-ETL action [intent=reverse_etl availability=not_implemented write=users_id_settings]; approval: requires plan, preview, approval, and execute; risk: BaseSpace mutation: POST /v1pre3/users/{user_id}/settings.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    workgroup list - Run the workgroup ETL stream [intent=etl availability=implemented stream=workgroup]; notes: discrepancy=present-in-surface-absent-from-artifact

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
