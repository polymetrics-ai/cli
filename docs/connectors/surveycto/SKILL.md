---
name: pm-surveycto
description: SurveyCTO connector knowledge and safe action guide.
---

# pm-surveycto

## Purpose

Reads SurveyCTO form IDs, submissions, datasets (including case-management datasets), dataset records, groups, roles, teams, and users, and writes dataset lifecycle mutations, dataset record creation, and user lifecycle mutations, through the SurveyCTO Server API v2.

## Icon

- asset: icons/surveycto.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.surveycto.com/05-exporting-and-publishing-data/02-api-access/01.api-access.html

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- form_id
- mode
- server_name
- password (secret)
- username (secret)

## ETL Streams

- datasets:
  - primary key: id
  - fields: id(), title(), version()
- dataset_records:
  - primary key: dataset_id, recordId
  - cursor: modifiedAt
  - fields: dataset_id(), modifiedAt(), recordId(), values()
- submissions:
  - primary key: id
  - cursor: submissionDate
  - fields: form_id(), id(), submissionDate()
- groups:
  - primary key: id
  - cursor: createdOn
  - fields: createdOn(), id(), parentGroupId(), title()
- roles:
  - primary key: id
  - fields: id(), name()
- users:
  - primary key: username
  - fields: roleId(), username()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_dataset:
  - endpoint: POST /datasets
  - optional fields: id, title, discriminator, uniqueRecordField, allowOfflineUpdates
  - risk: creates a new server dataset (a general-purpose, enumerator, or case-management dataset); low-risk external mutation, no approval required
- update_dataset:
  - endpoint: PUT /datasets/{{ record.id }}
  - required fields: id
  - optional fields: title, discriminator, uniqueRecordField, allowOfflineUpdates
  - risk: updates an existing dataset's metadata/configuration (the dataset type/discriminator itself cannot be changed after creation, per SurveyCTO's own API); external mutation, no approval required
- delete_dataset:
  - endpoint: DELETE /datasets/{{ record.id }}
  - required fields: id
  - risk: irreversibly deletes a dataset and its records; approval required
- create_dataset_record:
  - endpoint: POST /datasets/{{ record.dataset_id }}/records
  - required fields: dataset_id
  - risk: adds a new record to a dataset; the field name set is dataset-defined (SurveyCTO's own DatasetRecordFieldMap has no fixed schema), so record_schema only requires the routing field dataset_id -- every other record property is sent verbatim as the record's field-name/value map; low-risk external mutation, no approval required
- create_user:
  - endpoint: POST /users
  - risk: creates a new SurveyCTO server user AND sets their initial password in the same call; a credential-provisioning action, not an ordinary data mutation -- approval required
- update_user:
  - endpoint: PUT /users/{{ record.username }}
  - required fields: username
  - risk: updates an existing user's password and/or role; a credential-provisioning action when password is set -- approval required
- delete_user:
  - endpoint: DELETE /users/{{ record.username }}
  - required fields: username
  - risk: irreversibly deletes a server user and revokes their access; approval required

## Security

- read risk: external SurveyCTO API read of form, submission, dataset, group, role, team, and user data
- write risk: external SurveyCTO API mutation (dataset lifecycle, dataset record creation, user lifecycle including password-setting)
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect surveycto
```

### Inspect as structured JSON

```bash
pm connectors inspect surveycto --json
```

## Agent Rules

- Run pm connectors inspect surveycto before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
