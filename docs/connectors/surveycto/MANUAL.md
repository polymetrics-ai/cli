# pm connectors inspect surveycto

```text
NAME
  pm connectors inspect surveycto - SurveyCTO connector manual

SYNOPSIS
  pm connectors inspect surveycto
  pm connectors inspect surveycto --json
  pm credentials add <name> --connector surveycto [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SurveyCTO form IDs, submissions, datasets (including case-management datasets), dataset records, groups, roles, teams, and users, and writes dataset lifecycle mutations, dataset record creation, and user lifecycle mutations, through the SurveyCTO Server API v2.

ICON
  id: surveycto
  asset: icons/surveycto.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.surveycto.com/05-exporting-and-publishing-data/02-api-access/01.api-access.html

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  form_id
  mode
  server_name
  password (secret)
  username (secret)

ETL STREAMS
  datasets:
    primary key: id
    fields: id(string), title(string), version(string)
  dataset_records:
    primary key: dataset_id, recordId
    cursor: modifiedAt
    fields: dataset_id(string), modifiedAt(string), recordId(string), values(object)
  submissions:
    primary key: id
    cursor: submissionDate
    fields: form_id(string), id(string), submissionDate(string)
  groups:
    primary key: id
    cursor: createdOn
    fields: createdOn(string), id(integer), parentGroupId(integer), title(string)
  roles:
    primary key: id
    fields: id(string), name(string)
  users:
    primary key: username
    fields: roleId(string), username(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_dataset:
    endpoint: POST /datasets
    required fields: discriminator
    optional fields: id, title, uniqueRecordField, allowOfflineUpdates
    risk: creates a new server dataset (a general-purpose, enumerator, or case-management dataset); low-risk external mutation, no approval required
  update_dataset:
    endpoint: PUT /datasets/{{ record.id }}
    required fields: id, discriminator
    optional fields: title, uniqueRecordField, allowOfflineUpdates
    risk: updates an existing dataset's metadata/configuration (the dataset type/discriminator itself cannot be changed after creation, per SurveyCTO's own API); external mutation, no approval required
  delete_dataset:
    endpoint: DELETE /datasets/{{ record.id }}
    required fields: id
    risk: irreversibly deletes a dataset and its records; approval required
  create_dataset_record:
    endpoint: POST /datasets/{{ record.dataset_id }}/records
    required fields: dataset_id
    risk: adds a new record to a dataset; the field name set is dataset-defined (SurveyCTO's own DatasetRecordFieldMap has no fixed schema), so record_schema only requires the routing field dataset_id -- every other record property is sent verbatim as the record's field-name/value map; low-risk external mutation, no approval required
  create_user:
    endpoint: POST /users
    required fields: username, roleId, password
    risk: creates a new SurveyCTO server user AND sets their initial password in the same call; a credential-provisioning action, not an ordinary data mutation -- approval required
  update_user:
    endpoint: PUT /users/{{ record.username }}
    required fields: username
    risk: updates an existing user's password and/or role; a credential-provisioning action when password is set -- approval required
  delete_user:
    endpoint: DELETE /users/{{ record.username }}
    required fields: username
    risk: irreversibly deletes a server user and revokes their access; approval required

SECURITY
  read risk: external SurveyCTO API read of form, submission, dataset, group, role, team, and user data
  write risk: external SurveyCTO API mutation (dataset lifecycle, dataset record creation, user lifecycle including password-setting)
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run SurveyCTO's declared streams and reverse-ETL actions.
  Usage: pm surveycto <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete api v2 datasets datasetid record - Documented DELETE /api/v2/datasets/{datasetId}/record (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.delete.api-v2-datasets-datasetid-record]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete api v2 users bulk - Documented DELETE /api/v2/users/bulk (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.delete.api-v2-users-bulk]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get api v2 datasets data csv datasetid - Documented GET /api/v2/datasets/data/csv/{datasetID:.+} (not implemented) [intent=direct_read availability=not_implemented operation=surveycto.get.api-v2-datasets-data-csv-datasetid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 datasets datasetid - Documented GET /api/v2/datasets/{datasetId} (not implemented) [intent=direct_read availability=not_implemented operation=surveycto.get.api-v2-datasets-datasetid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 datasets datasetid record - Documented GET /api/v2/datasets/{datasetId}/record (not implemented) [intent=direct_read availability=not_implemented operation=surveycto.get.api-v2-datasets-datasetid-record]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 forms ids - Documented GET /api/v2/forms/ids (not implemented) [intent=direct_read availability=not_implemented operation=surveycto.get.api-v2-forms-ids]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 roles roleid - Documented GET /api/v2/roles/{roleId} (not implemented) [intent=direct_read availability=not_implemented operation=surveycto.get.api-v2-roles-roleid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 teams ids - Documented GET /api/v2/teams/ids (not implemented) [intent=direct_read availability=not_implemented operation=surveycto.get.api-v2-teams-ids]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v2 users username - Documented GET /api/v2/users/{username} (not implemented) [intent=direct_read availability=not_implemented operation=surveycto.get.api-v2-users-username]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch api v2 datasets datasetid record - Documented PATCH /api/v2/datasets/{datasetId}/record (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.patch.api-v2-datasets-datasetid-record]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 datasets datasetid purge - Documented POST /api/v2/datasets/{datasetId}/purge (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.post.api-v2-datasets-datasetid-purge]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 datasets datasetid records search - Documented POST /api/v2/datasets/{datasetId}/records/search (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.post.api-v2-datasets-datasetid-records-search]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 datasets datasetid records upload - Documented POST /api/v2/datasets/{datasetId}/records/upload (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.post.api-v2-datasets-datasetid-records-upload]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 forms data wide json formid - Documented POST /api/v2/forms/data/wide/json/{formID:.+} (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.post.api-v2-forms-data-wide-json-formid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 forms formid submissions - Documented POST /api/v2/forms/{formID:.+}/submissions (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.post.api-v2-forms-formid-submissions]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 forms formid submissions instanceid attachments filename - Documented POST /api/v2/forms/{formID:.+}/submissions/{instanceID:.+}/attachments/{filename:.+} (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.post.api-v2-forms-formid-submissions-instanceid-attachments-filename]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 forms formid submissions search - Documented POST /api/v2/forms/{formID:.+}/submissions/search (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.post.api-v2-forms-formid-submissions-search]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 users bulk file - Documented POST /api/v2/users/bulk/file (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.post.api-v2-users-bulk-file]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v2 users bulk json - Documented POST /api/v2/users/bulk/json (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.post.api-v2-users-bulk-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v2 datasets datasetid record - Documented PUT /api/v2/datasets/{datasetId}/record (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.put.api-v2-datasets-datasetid-record]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v2 users bulk file - Documented PUT /api/v2/users/bulk/file (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.put.api-v2-users-bulk-file]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api v2 users bulk json - Documented PUT /api/v2/users/bulk/json (not implemented) [intent=direct_write availability=not_implemented operation=surveycto.put.api-v2-users-bulk-json]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    create dataset apply - Plan and execute the create dataset reverse-ETL action [intent=reverse_etl availability=implemented write=create_dataset]; approval: requires plan, preview, approval, and execute; risk: creates a new server dataset (a general-purpose, enumerator, or case-management dataset); low-risk external mutation, no approval required; flags: --discriminator (required)
    create dataset record apply - Plan and execute the create dataset record reverse-ETL action [intent=reverse_etl availability=implemented write=create_dataset_record]; approval: requires plan, preview, approval, and execute; risk: adds a new record to a dataset; the field name set is dataset-defined (SurveyCTO's own DatasetRecordFieldMap has no fixed schema), so record_schema only requires the routing field dataset_id -- every other record property is sent verbatim as the record's field-name/value map; low-risk external mutation, no approval required; flags: --dataset_id (required)
    create user apply - Plan and execute the create user reverse-ETL action [intent=reverse_etl availability=implemented write=create_user]; approval: requires plan, preview, approval, and execute; risk: creates a new SurveyCTO server user AND sets their initial password in the same call; a credential-provisioning action, not an ordinary data mutation -- approval required; flags: --password (required), --roleId (required), --username (required)
    dataset records list - Run the dataset records ETL stream [intent=etl availability=implemented stream=dataset_records]
    datasets list - Run the datasets ETL stream [intent=etl availability=implemented stream=datasets]
    delete dataset apply - Plan and execute the delete dataset reverse-ETL action [intent=reverse_etl availability=implemented write=delete_dataset]; approval: requires plan, preview, approval, and execute; risk: irreversibly deletes a dataset and its records; approval required; flags: --id (required)
    delete user apply - Plan and execute the delete user reverse-ETL action [intent=reverse_etl availability=implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: irreversibly deletes a server user and revokes their access; approval required; flags: --username (required)
    groups list - Run the groups ETL stream [intent=etl availability=implemented stream=groups]
    roles list - Run the roles ETL stream [intent=etl availability=implemented stream=roles]
    submissions list - Run the submissions ETL stream [intent=etl availability=implemented stream=submissions]
    update dataset apply - Plan and execute the update dataset reverse-ETL action [intent=reverse_etl availability=implemented write=update_dataset]; approval: requires plan, preview, approval, and execute; risk: updates an existing dataset's metadata/configuration (the dataset type/discriminator itself cannot be changed after creation, per SurveyCTO's own API); external mutation, no approval required; flags: --discriminator (required), --id (required)
    update user apply - Plan and execute the update user reverse-ETL action [intent=reverse_etl availability=implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: updates an existing user's password and/or role; a credential-provisioning action when password is set -- approval required; flags: --username (required)
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect surveycto

  # Inspect as structured JSON
  pm connectors inspect surveycto --json

AGENT WORKFLOW
  - Run pm connectors inspect surveycto before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
