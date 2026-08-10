# pm connectors inspect avni

```text
NAME
  pm connectors inspect avni - Avni connector manual

SYNOPSIS
  pm connectors inspect avni
  pm connectors inspect avni --json
  pm credentials add <name> --connector avni [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Avni subjects and encounters through a read-only HTTP API using HTTP Basic authentication.

ICON
  id: avni
  asset: icons/avni.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page_size
  start_date
  username
  password (secret)

ETL STREAMS
  subjects:
    primary key: id
    cursor: updated_at
    fields: id(string), name(string), updated_at(string)
  encounters:
    primary key: id
    cursor: updated_at
    fields: encounter_type(string), id(string), subject_id(string), updated_at(string)
  program_enrolments:
    primary key: id
    cursor: updated_at
    fields: enrolment_date_time(string), exit_date_time(string), id(string), program(string), subject_id(string), updated_at(string)
  program_encounters:
    primary key: id
    cursor: updated_at
    fields: encounter_date_time(string), encounter_type(string), enrolment_id(string), id(string), program(string), subject_id(string), updated_at(string)
  group_subjects:
    primary key: id
    cursor: updated_at
    fields: group_subject_id(string), id(string), member_subject_id(string), membership_end_date(string), membership_start_date(string), updated_at(string)
  locations:
    primary key: id
    cursor: updated_at
    fields: id(string), level(number), parent_id(string), title(string), type(string), updated_at(string)
  approval_statuses:
    primary key: entity_id, entity_type
    cursor: status_date_time
    fields: approval_status(string), approval_status_comment(string), entity_id(string), entity_type(string), entity_type_id(string), status_date_time(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Avni API read of subjects and encounters
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Avni's declared streams and reverse-ETL actions.
  Usage: pm avni <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete api encounter id - Documented DELETE /api/encounter/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.delete.api-encounter-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api programencounter id - Documented DELETE /api/programEncounter/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.delete.api-programencounter-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api programenrolment id - Documented DELETE /api/programEnrolment/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.delete.api-programenrolment-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api subject id - Documented DELETE /api/subject/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.delete.api-subject-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api subjecttree - Documented DELETE /api/subjectTree (not implemented) [intent=direct_write availability=not_implemented operation=avni.delete.api-subjecttree]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get api encounter id - Documented GET /api/encounter/{ID} (not implemented) [intent=direct_read availability=not_implemented operation=avni.get.api-encounter-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api location id - Documented GET /api/location/{ID} (not implemented) [intent=direct_read availability=not_implemented operation=avni.get.api-location-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api programencounter id - Documented GET /api/programEncounter/{ID} (not implemented) [intent=direct_read availability=not_implemented operation=avni.get.api-programencounter-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api programenrolment id - Documented GET /api/programEnrolment/{ID} (not implemented) [intent=direct_read availability=not_implemented operation=avni.get.api-programenrolment-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api subject id - Documented GET /api/subject/{ID} (not implemented) [intent=direct_read availability=not_implemented operation=avni.get.api-subject-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api subjectmigration bulk status jobid - Documented GET /api/subjectMigration/bulk/status/{jobId} (not implemented) [intent=direct_read availability=not_implemented operation=avni.get.api-subjectmigration-bulk-status-jobid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch api encounter id - Documented PATCH /api/encounter/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.patch.api-encounter-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch api subject id - Documented PATCH /api/subject/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.patch.api-subject-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api encounter - Documented POST /api/encounter (not implemented) [intent=direct_write availability=not_implemented operation=avni.post.api-encounter]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api programencounter - Documented POST /api/programEncounter (not implemented) [intent=direct_write availability=not_implemented operation=avni.post.api-programencounter]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api programenrolment - Documented POST /api/programEnrolment (not implemented) [intent=direct_write availability=not_implemented operation=avni.post.api-programenrolment]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api subject - Documented POST /api/subject (not implemented) [intent=direct_write availability=not_implemented operation=avni.post.api-subject]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api subjectmigration bulk - Documented POST /api/subjectMigration/bulk (not implemented) [intent=direct_write availability=not_implemented operation=avni.post.api-subjectmigration-bulk]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api task - Documented POST /api/task (not implemented) [intent=direct_write availability=not_implemented operation=avni.post.api-task]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api user enable - Documented POST /api/user/enable (not implemented) [intent=direct_write availability=not_implemented operation=avni.post.api-user-enable]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api encounter id - Documented PUT /api/encounter/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.put.api-encounter-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api programencounter id - Documented PUT /api/programEncounter/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.put.api-programencounter-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api programenrolment id - Documented PUT /api/programEnrolment/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.put.api-programenrolment-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api subject id - Documented PUT /api/subject/{ID} (not implemented) [intent=direct_write availability=not_implemented operation=avni.put.api-subject-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    approval statuses list - Run the approval statuses ETL stream [intent=etl availability=implemented stream=approval_statuses]
    encounters list - Run the encounters ETL stream [intent=etl availability=implemented stream=encounters]
    group subjects list - Run the group subjects ETL stream [intent=etl availability=implemented stream=group_subjects]
    locations list - Run the locations ETL stream [intent=etl availability=implemented stream=locations]
    program encounters list - Run the program encounters ETL stream [intent=etl availability=implemented stream=program_encounters]
    program enrolments list - Run the program enrolments ETL stream [intent=etl availability=implemented stream=program_enrolments]
    subjects list - Run the subjects ETL stream [intent=etl availability=implemented stream=subjects]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect avni

  # Inspect as structured JSON
  pm connectors inspect avni --json

AGENT WORKFLOW
  - Run pm connectors inspect avni before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
