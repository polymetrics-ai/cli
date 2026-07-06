# Overview

Reads and writes documented Illumina BaseSpace v1pre3 REST API resources through the
connector engine.

Readable streams: `projects`, `runs`, `samples`, `appsessions`, `datasets`, `applications`,
`application`, `application_qcthresholds`, `application_settings`,
`application_workflowdependencies`, `appsessions_all`, `appsession`, `appsession_comments`,
`appsessions_logfiles`, `appsession_properties`, `appsession_property`, `appsession_property_items`,
`biosamples`, `biosample_labrequeues`, `biosample`, `biosample_libraries`,
`biosample_runlane_summaries`, `datasets_all`, `dataset`, `dataset_comments`, `dataset_files`,
`datasettype`, `instrumentstatistics`, `labrequeues`, `labrequeue`, `laneqcthresholds`, `lane`,
`lane_comments`, `libraries`, `librarypool_libraries`, `project`, `project_datasets`, `run_files`,
`trash`, `trash_2`, `current_user`, `current_user_subscription`, `current_user_usage`,
`current_user_workgroups`, `user`, `user_settings`, `workgroup`, `configured_user`.

Write actions: `update_applications_id_qcthresholds`, `update_applications_id_workflowdependencies`,
`applications_id_workflows`, `delete_appsessions_id`, `appsessions_id`, `appsessions_id_properties`,
`delete_appsessions_id_properties_name`, `appsessions_id_stop`, `biosamples_bulkupdate`,
`biosamples_id`, `datasets_id`, `update_laneqcthresholds`, `lanes_id`,
`libraries_libraryid_labrequeues`, `librarypools_id`, `librarypools_poolid_labrequeues`,
`preprequests_preprequestid_labrequeues`, `delete_trash`, `trash_id_restorefromtrash`,
`users_id_settings`.

Service API documentation:
https://developer.basespace.illumina.com/docs/content/documentation/rest-api/api-reference.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Illumina BaseSpace access token, sent as the
  x-access-token header. Never logged.
- `application_id` (optional, string); Application Id used by BaseSpace stream paths.
- `appsession_id` (optional, string); Appsession Id used by BaseSpace stream paths.
- `base_url` (required, string); format `uri`; BaseSpace API domain, e.g.
  https://api.basespace.illumina.com or a regional domain such as
  https://euw2.sh.basespace.illumina.com. BaseSpace is domain-scoped with no single fixed default,
  so this is required.
- `biosample_id` (optional, string); Biosample Id used by BaseSpace stream paths.
- `dataset_id` (optional, string); Dataset Id used by BaseSpace stream paths.
- `datasettype_id` (optional, string); Datasettype Id used by BaseSpace stream paths.
- `labrequeue_id` (optional, string); Labrequeue Id used by BaseSpace stream paths.
- `lane_id` (optional, string); Lane Id used by BaseSpace stream paths.
- `librarypool_id` (optional, string); Librarypool Id used by BaseSpace stream paths.
- `name` (optional, string); Name used by BaseSpace stream paths.
- `page_size` (optional, integer); default `100`; Pagination uses the fixed
  bundle page size of 100.
- `project_id` (optional, string); Project Id used by BaseSpace stream paths.
- `run_id` (optional, string); Run Id used by BaseSpace stream paths.
- `trash_id` (optional, string); Trash Id used by BaseSpace stream paths.
- `user` (optional, string); default `current`.
- `user_id` (optional, string); User Id used by BaseSpace stream paths.
- `workgroup_id` (optional, string); Workgroup Id used by BaseSpace stream paths.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `page_size=100`, `user=current`.

Authentication behavior:

- API key authentication in `x-access-token` using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v1pre3/users/{{ config.user }}`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`;
page size 100.

Pagination by stream: none: `application`, `application_qcthresholds`, `application_settings`,
`application_workflowdependencies`, `appsession`, `appsession_property`, `biosample`, `dataset`,
`datasettype`, `labrequeue`, `laneqcthresholds`, `lane`, `project`, `trash_2`, `current_user`,
`current_user_subscription`, `current_user_usage`, `user`, `workgroup`, `configured_user`;
offset_limit: `projects`, `runs`, `samples`, `appsessions`, `datasets`, `applications`,
`appsessions_all`, `appsession_comments`, `appsessions_logfiles`, `appsession_properties`,
`appsession_property_items`, `biosamples`, `biosample_labrequeues`, `biosample_libraries`,
`biosample_runlane_summaries`, `datasets_all`, `dataset_comments`, `dataset_files`,
`instrumentstatistics`, `labrequeues`, `lane_comments`, `libraries`, `librarypool_libraries`,
`project_datasets`, `run_files`, `trash`, `current_user_workgroups`, `user_settings`.

- `projects`: GET `/v1pre3/users/{{ config.user }}/projects` - records path `Response.Items`;
  offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100;
  computed output fields `date_created`, `date_modified`, `href`, `id`, `name`, `total_size`,
  `user_owned_by`.
- `runs`: GET `/v1pre3/users/{{ config.user }}/runs` - records path `Response.Items`; offset/limit
  pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100; computed output
  fields `date_created`, `date_modified`, `experiment_name`, `href`, `id`, `instrument_name`,
  `name`, `status`, `total_size`.
- `samples`: GET `/v1pre3/users/{{ config.user }}/samples` - records path `Response.Items`;
  offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100;
  computed output fields `date_created`, `href`, `id`, `name`, `num_reads_pf`, `num_reads_raw`,
  `sample_id`, `status`, `total_size`.
- `appsessions`: GET `/v1pre3/users/{{ config.user }}/appsessions` - records path `Response.Items`;
  offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100;
  computed output fields `application`, `date_completed`, `date_created`, `href`, `id`, `name`,
  `status`, `status_summary`, `total_size`.
- `datasets`: GET `/v1pre3/users/{{ config.user }}/datasets` - records path `Response.Items`;
  offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100;
  computed output fields `dataset_type`, `date_created`, `href`, `id`, `name`, `project`,
  `total_size`.
- `applications`: GET `/v1pre3/applications` - records path `Response.Items`; offset/limit
  pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100; emits passthrough
  records.
- `application`: GET `/v1pre3/applications/{{ config.application_id }}` - records path `Response`;
  emits passthrough records.
- `application_qcthresholds`: GET `/v1pre3/applications/{{ config.application_id }}/qcthresholds` -
  records path `Response`; emits passthrough records.
- `application_settings`: GET `/v1pre3/applications/{{ config.application_id }}/settings` - records
  path `Response`; emits passthrough records.
- `application_workflowdependencies`: GET `/v1pre3/applications/{{ config.application_id
  }}/workflowdependencies` - records path `Response`; emits passthrough records.
- `appsessions_all`: GET `/v1pre3/appsessions` - records path `Response.Items`; offset/limit
  pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100; emits passthrough
  records.
- `appsession`: GET `/v1pre3/appsessions/{{ config.appsession_id }}` - records path `Response`;
  emits passthrough records.
- `appsession_comments`: GET `/v1pre3/appsessions/{{ config.appsession_id }}/comments` - records
  path `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter
  `Limit`; page size 100; emits passthrough records.
- `appsessions_logfiles`: GET `/v1pre3/appsessions/{{ config.appsession_id }}/logfiles` - records
  path `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter
  `Limit`; page size 100; emits passthrough records.
- `appsession_properties`: GET `/v1pre3/appsessions/{{ config.appsession_id }}/properties` - records
  path `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter
  `Limit`; page size 100; emits passthrough records.
- `appsession_property`: GET `/v1pre3/appsessions/{{ config.appsession_id }}/properties/{{
  config.name }}` - records path `Response`; emits passthrough records.
- `appsession_property_items`: GET `/v1pre3/appsessions/{{ config.appsession_id }}/properties/{{
  config.name }}/items` - records path `Response.Items`; offset/limit pagination; offset parameter
  `Offset`; limit parameter `Limit`; page size 100; emits passthrough records.
- `biosamples`: GET `/v1pre3/biosamples` - records path `Response.Items`; offset/limit pagination;
  offset parameter `Offset`; limit parameter `Limit`; page size 100; emits passthrough records.
- `biosample_labrequeues`: GET `/v1pre3/biosamples/{{ config.biosample_id }}/labrequeues` - records
  path `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter
  `Limit`; page size 100; emits passthrough records.
- `biosample`: GET `/v1pre3/biosamples/{{ config.biosample_id }}` - records path `Response`; emits
  passthrough records.
- `biosample_libraries`: GET `/v1pre3/biosamples/{{ config.biosample_id }}/libraries` - records path
  `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`;
  page size 100; emits passthrough records.
- `biosample_runlane_summaries`: GET `/v1pre3/biosamples/{{ config.biosample_id }}/runlanesummaries`
  - records path `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit
  parameter `Limit`; page size 100; emits passthrough records.
- `datasets_all`: GET `/v1pre3/datasets` - records path `Response.Items`; offset/limit pagination;
  offset parameter `Offset`; limit parameter `Limit`; page size 100; emits passthrough records.
- `dataset`: GET `/v1pre3/datasets/{{ config.dataset_id }}` - records path `Response`; emits
  passthrough records.
- `dataset_comments`: GET `/v1pre3/datasets/{{ config.dataset_id }}/comments` - records path
  `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`;
  page size 100; emits passthrough records.
- `dataset_files`: GET `/v1pre3/datasets/{{ config.dataset_id }}/files` - records path
  `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`;
  page size 100; emits passthrough records.
- `datasettype`: GET `/v1pre3/datasettypes/{{ config.datasettype_id }}` - records path `Response`;
  emits passthrough records.
- `instrumentstatistics`: GET `/v1pre3/instrumentstatistics` - records path `Response.Items`;
  offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100; emits
  passthrough records.
- `labrequeues`: GET `/v1pre3/labrequeues` - records path `Response.Items`; offset/limit pagination;
  offset parameter `Offset`; limit parameter `Limit`; page size 100; emits passthrough records.
- `labrequeue`: GET `/v1pre3/labrequeues/{{ config.labrequeue_id }}` - records path `Response`;
  emits passthrough records.
- `laneqcthresholds`: GET `/v1pre3/laneqcthresholds` - records path `Response`; emits passthrough
  records.
- `lane`: GET `/v1pre3/lanes/{{ config.lane_id }}` - records path `Response`; emits passthrough
  records.
- `lane_comments`: GET `/v1pre3/lanes/{{ config.lane_id }}/comments` - records path
  `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`;
  page size 100; emits passthrough records.
- `libraries`: GET `/v1pre3/libraries` - records path `Response.Items`; offset/limit pagination;
  offset parameter `Offset`; limit parameter `Limit`; page size 100; emits passthrough records.
- `librarypool_libraries`: GET `/v1pre3/librarypools/{{ config.librarypool_id }}/libraries` -
  records path `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter
  `Limit`; page size 100; emits passthrough records.
- `project`: GET `/v1pre3/projects/{{ config.project_id }}` - records path `Response`; emits
  passthrough records.
- `project_datasets`: GET `/v1pre3/projects/{{ config.project_id }}/datasets` - records path
  `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`;
  page size 100; emits passthrough records.
- `run_files`: GET `/v1pre3/runs/{{ config.run_id }}/files` - records path `Response.Items`;
  offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100; emits
  passthrough records.
- `trash`: GET `/v1pre3/trash` - records path `Response.Items`; offset/limit pagination; offset
  parameter `Offset`; limit parameter `Limit`; page size 100; emits passthrough records.
- `trash_2`: GET `/v1pre3/trash/{{ config.trash_id }}` - records path `Response`; emits passthrough
  records.
- `current_user`: GET `/v1pre3/users/current` - records path `Response`; emits passthrough records.
- `current_user_subscription`: GET `/v1pre3/users/current/subscription` - records path `Response`;
  emits passthrough records.
- `current_user_usage`: GET `/v1pre3/users/current/usage` - records path `Response`; emits
  passthrough records.
- `current_user_workgroups`: GET `/v1pre3/users/current/workgroups` - records path `Response.Items`;
  offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`; page size 100; emits
  passthrough records.
- `user`: GET `/v1pre3/users/{{ config.user_id }}` - records path `Response`; emits passthrough
  records.
- `user_settings`: GET `/v1pre3/users/{{ config.user_id }}/settings` - records path
  `Response.Items`; offset/limit pagination; offset parameter `Offset`; limit parameter `Limit`;
  page size 100; emits passthrough records.
- `workgroup`: GET `/v1pre3/workgroups/{{ config.workgroup_id }}` - records path `Response`; emits
  passthrough records.
- `configured_user`: GET `/v1pre3/users/{{ config.user }}` - records path `Response`; emits
  passthrough records.

## Write actions & risks

Overall write risk: external Illumina BaseSpace API mutations including settings, workflow,
threshold, trash, lab requeue, and app session actions.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_applications_id_qcthresholds`: PUT `/v1pre3/applications/{{ record.application_id
  }}/qcthresholds` - kind `update`; body type `json`; path fields `application_id`; required record
  fields `application_id`; accepted fields `application_id`; risk: BaseSpace mutation: PUT
  /v1pre3/applications/{application_id}/qcthresholds.
- `update_applications_id_workflowdependencies`: PUT `/v1pre3/applications/{{ record.application_id
  }}/workflowdependencies` - kind `update`; body type `json`; path fields `application_id`; required
  record fields `application_id`; accepted fields `application_id`; risk: BaseSpace mutation: PUT
  /v1pre3/applications/{application_id}/workflowdependencies.
- `applications_id_workflows`: POST `/v1pre3/applications/{{ record.application_id }}/workflows` -
  kind `create`; body type `json`; path fields `application_id`; required record fields
  `application_id`; accepted fields `application_id`; risk: BaseSpace mutation: POST
  /v1pre3/applications/{application_id}/workflows.
- `delete_appsessions_id`: DELETE `/v1pre3/appsessions/{{ record.appsession_id }}` - kind `delete`;
  body type `none`; path fields `appsession_id`; required record fields `appsession_id`; accepted
  fields `appsession_id`; confirmation `destructive`; risk: Destructive BaseSpace mutation: DELETE
  /v1pre3/appsessions/{appsession_id}.
- `appsessions_id`: POST `/v1pre3/appsessions/{{ record.appsession_id }}` - kind `create`; body type
  `json`; path fields `appsession_id`; required record fields `appsession_id`; accepted fields
  `appsession_id`; risk: BaseSpace mutation: POST /v1pre3/appsessions/{appsession_id}.
- `appsessions_id_properties`: POST `/v1pre3/appsessions/{{ record.appsession_id }}/properties` -
  kind `create`; body type `json`; path fields `appsession_id`; required record fields
  `appsession_id`; accepted fields `appsession_id`; risk: BaseSpace mutation: POST
  /v1pre3/appsessions/{appsession_id}/properties.
- `delete_appsessions_id_properties_name`: DELETE `/v1pre3/appsessions/{{ record.appsession_id
  }}/properties/{{ record.name }}` - kind `delete`; body type `none`; path fields `appsession_id`,
  `name`; required record fields `appsession_id`, `name`; accepted fields `appsession_id`, `name`;
  confirmation `destructive`; risk: Destructive BaseSpace mutation: DELETE
  /v1pre3/appsessions/{appsession_id}/properties/{name}.
- `appsessions_id_stop`: POST `/v1pre3/appsessions/{{ record.appsession_id }}/stop` - kind `create`;
  body type `json`; path fields `appsession_id`; required record fields `appsession_id`; accepted
  fields `appsession_id`; confirmation `destructive`; risk: Destructive BaseSpace mutation: POST
  /v1pre3/appsessions/{appsession_id}/stop.
- `biosamples_bulkupdate`: POST `/v1pre3/biosamples/bulkupdate` - kind `create`; body type `json`;
  risk: BaseSpace mutation: POST /v1pre3/biosamples/bulkupdate.
- `biosamples_id`: POST `/v1pre3/biosamples/{{ record.biosample_id }}` - kind `create`; body type
  `json`; path fields `biosample_id`; required record fields `biosample_id`; accepted fields
  `biosample_id`; risk: BaseSpace mutation: POST /v1pre3/biosamples/{biosample_id}.
- `datasets_id`: POST `/v1pre3/datasets/{{ record.dataset_id }}` - kind `create`; body type `json`;
  path fields `dataset_id`; required record fields `dataset_id`; accepted fields `dataset_id`; risk:
  BaseSpace mutation: POST /v1pre3/datasets/{dataset_id}.
- `update_laneqcthresholds`: PUT `/v1pre3/laneqcthresholds` - kind `update`; body type `json`; risk:
  BaseSpace mutation: PUT /v1pre3/laneqcthresholds.
- `lanes_id`: POST `/v1pre3/lanes/{{ record.lane_id }}` - kind `create`; body type `json`; path
  fields `lane_id`; required record fields `lane_id`; accepted fields `lane_id`; risk: BaseSpace
  mutation: POST /v1pre3/lanes/{lane_id}.
- `libraries_libraryid_labrequeues`: POST `/v1pre3/libraries/{{ record.library_id }}/labrequeues` -
  kind `create`; body type `json`; path fields `library_id`; required record fields `library_id`;
  accepted fields `library_id`; risk: BaseSpace mutation: POST
  /v1pre3/libraries/{library_id}/labrequeues.
- `librarypools_id`: POST `/v1pre3/librarypools/{{ record.librarypool_id }}` - kind `create`; body
  type `json`; path fields `librarypool_id`; required record fields `librarypool_id`; accepted
  fields `librarypool_id`; risk: BaseSpace mutation: POST /v1pre3/librarypools/{librarypool_id}.
- `librarypools_poolid_labrequeues`: POST `/v1pre3/librarypools/{{ record.pool_id }}/labrequeues` -
  kind `create`; body type `json`; path fields `pool_id`; required record fields `pool_id`; accepted
  fields `pool_id`; risk: BaseSpace mutation: POST /v1pre3/librarypools/{pool_id}/labrequeues.
- `preprequests_preprequestid_labrequeues`: POST `/v1pre3/preprequests/{{ record.preprequest_id
  }}/labrequeues` - kind `create`; body type `json`; path fields `preprequest_id`; required record
  fields `preprequest_id`; accepted fields `preprequest_id`; risk: BaseSpace mutation: POST
  /v1pre3/preprequests/{preprequest_id}/labrequeues.
- `delete_trash`: DELETE `/v1pre3/trash` - kind `delete`; body type `none`; confirmation
  `destructive`; risk: Destructive BaseSpace mutation: DELETE /v1pre3/trash.
- `trash_id_restorefromtrash`: POST `/v1pre3/trash/{{ record.trash_id }}/restorefromtrash` - kind
  `create`; body type `json`; path fields `trash_id`; required record fields `trash_id`; accepted
  fields `trash_id`; confirmation `destructive`; risk: Destructive BaseSpace mutation: POST
  /v1pre3/trash/{trash_id}/restorefromtrash.
- `users_id_settings`: POST `/v1pre3/users/{{ record.user_id }}/settings` - kind `create`; body type
  `json`; path fields `user_id`; required record fields `user_id`; accepted fields `user_id`; risk:
  BaseSpace mutation: POST /v1pre3/users/{user_id}/settings.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 48 stream-backed endpoint group(s), 20 write-backed endpoint group(s).
