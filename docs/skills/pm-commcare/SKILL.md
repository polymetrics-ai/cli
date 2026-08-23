---
name: pm-commcare
description: CommCare connector knowledge and safe action guide.
---

# pm-commcare

## Purpose

Reads CommCare HQ project, application, form, case, user, group, report, location, lookup table, export, and messaging API data; writes supported JSON mutations for cases, users, groups, locations, and lookup tables.

## Icon

- id: commcare
- asset: icons/commcare.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://confluence.dimagi.com/display/commcarepublic/CommCare+HQ+APIs

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- app_id
- base_url
- case_id
- case_ids
- external_id
- fixture_item_id
- fixture_type
- form_id
- group_id
- location_id
- location_type_id
- lookup_table_id
- lookup_table_item_id
- mobile_worker_id
- processing_id
- project_space (required)
- report_id
- web_user_id
- api_key (secret) (required)

## ETL Streams

- forms:
  - primary key: id
  - cursor: received_on
  - fields: id(string), received_on(string), server_modified_on(string)
- cases:
  - primary key: id
  - cursor: server_modified_on
  - fields: id(string), received_on(string), server_modified_on(string)
- applications:
  - primary key: id
  - fields: id(string), is_released(boolean), modules(array), name(string), version(integer), versions(array)
- application:
  - primary key: id
  - fields: id(string), is_released(boolean), modules(array), name(string), version(integer), versions(array)
- multimedia_upload_status:
  - primary key: processing_id
  - fields: complete(boolean), errors(array), in_celery(boolean), processing_id(string), progress(object), success(boolean)
- forms_v1:
  - primary key: id
  - cursor: received_on
  - fields: id(string), received_on(string), server_modified_on(string)
- form:
  - primary key: id
  - cursor: received_on
  - fields: id(string), received_on(string), server_modified_on(string)
- cases_v1:
  - primary key: case_id
  - cursor: server_date_modified
  - fields: case_id(string), closed(boolean), date_modified(string), domain(string), indices(object), properties(object), server_date_modified(string), user_id(string), xform_ids(array)
- case_v1:
  - primary key: case_id
  - cursor: server_date_modified
  - fields: case_id(string), closed(boolean), date_modified(string), domain(string), indices(object), properties(object), server_date_modified(string), user_id(string), xform_ids(array)
- cases_v2:
  - primary key: case_id
  - cursor: indexed_on
  - fields: case_id(string), case_name(string), case_type(string), closed(boolean), date_closed(string), date_opened(string), domain(string), error(string), external_id(string), indexed_on(string), indices(object), last_modified(string), owner_id(string), properties(object), server_last_modified(string)
- case_v2:
  - primary key: case_id
  - cursor: indexed_on
  - fields: case_id(string), case_name(string), case_type(string), closed(boolean), date_closed(string), date_opened(string), domain(string), error(string), external_id(string), indexed_on(string), indices(object), last_modified(string), owner_id(string), properties(object), server_last_modified(string)
- case_v2_by_external_id:
  - primary key: case_id
  - cursor: indexed_on
  - fields: case_id(string), case_name(string), case_type(string), closed(boolean), date_closed(string), date_opened(string), domain(string), error(string), external_id(string), indexed_on(string), indices(object), last_modified(string), owner_id(string), properties(object), server_last_modified(string)
- case_v2_bulk_by_ids:
  - primary key: case_id
  - cursor: indexed_on
  - fields: case_id(string), case_name(string), case_type(string), closed(boolean), date_closed(string), date_opened(string), domain(string), error(string), external_id(string), indexed_on(string), indices(object), last_modified(string), owner_id(string), properties(object), server_last_modified(string)
- case_v2_index_children:
  - primary key: case_id
  - cursor: indexed_on
  - fields: case_id(string), case_name(string), case_type(string), closed(boolean), date_closed(string), date_opened(string), domain(string), error(string), external_id(string), indexed_on(string), indices(object), last_modified(string), owner_id(string), properties(object), server_last_modified(string)
- mobile_workers:
  - primary key: id
  - fields: default_phone_number(string), email(string), first_name(string), groups(array), id(string), last_name(string), locations(array), phone_numbers(array), primary_location(string), type(string), user_data(object), username(string)
- mobile_worker:
  - primary key: id
  - fields: default_phone_number(string), email(string), first_name(string), groups(array), id(string), last_name(string), locations(array), phone_numbers(array), primary_location(string), type(string), user_data(object), username(string)
- bulk_users:
  - primary key: id
  - fields: email(string), first_name(string), id(string), last_name(string), phone_numbers(array), resource_uri(string), username(string)
- web_users:
  - primary key: id
  - fields: default_phone_number(string), email(string), first_name(string), id(string), is_admin(boolean), last_name(string), permissions(object), phone_numbers(array), resource_uri(string), role(string), username(string)
- web_user:
  - primary key: id
  - fields: default_phone_number(string), email(string), first_name(string), id(string), is_admin(boolean), last_name(string), permissions(object), phone_numbers(array), resource_uri(string), role(string), username(string)
- user_domains:
  - primary key: domain_name
  - fields: domain_name(string), project_name(string)
- user_identity:
  - primary key: id
  - fields: email(string), first_name(string), id(string), last_name(string), username(string)
- groups:
  - primary key: id
  - fields: case_sharing(boolean), domain(string), id(string), metadata(object), name(string), path(array), reporting(boolean), users(array)
- group:
  - primary key: id
  - fields: case_sharing(boolean), domain(string), id(string), metadata(object), name(string), path(array), reporting(boolean), users(array)
- reports:
  - primary key: id
  - fields: columns(array), filters(array), id(string), resource_uri(string), title(string)
- report_data:
- locations_v1:
  - primary key: location_id
  - cursor: last_modified
  - fields: created_at(string), domain(string), external_id(string), id(integer), last_modified(string), latitude(string), location_data(object), location_id(string), location_type(string), longitude(string), name(string), parent(string), resource_uri(string), site_code(string)
- location_v1:
  - primary key: location_id
  - cursor: last_modified
  - fields: created_at(string), domain(string), external_id(string), id(integer), last_modified(string), latitude(string), location_data(object), location_id(string), location_type(string), longitude(string), name(string), parent(string), resource_uri(string), site_code(string)
- locations_v2:
  - primary key: location_id
  - cursor: last_modified
  - fields: domain(string), last_modified(string), latitude(string), location_data(object), location_id(string), location_type_code(string), location_type_name(string), longitude(string), name(string), parent_location_id(string), site_code(string)
- location_v2:
  - primary key: location_id
  - cursor: last_modified
  - fields: domain(string), last_modified(string), latitude(string), location_data(object), location_id(string), location_type_code(string), location_type_name(string), longitude(string), name(string), parent_location_id(string), site_code(string)
- location_types:
  - primary key: id
  - fields: administrative(boolean), code(string), domain(string), id(integer), name(string), parent(integer), resource_uri(string), shares_cases(boolean), view_descendants(boolean)
- location_type:
  - primary key: id
  - fields: administrative(boolean), code(string), domain(string), id(integer), name(string), parent(integer), resource_uri(string), shares_cases(boolean), view_descendants(boolean)
- fixtures:
  - primary key: id
  - fields: fields(object), fixture_type(string), id(string), resource_uri(string)
- fixture_table_items:
  - primary key: id
  - fields: fields(object), fixture_type(string), id(string), resource_uri(string)
- fixture_item:
  - primary key: id
  - fields: fields(object), fixture_type(string), id(string), resource_uri(string)
- lookup_tables:
  - primary key: id
  - fields: fields(array), id(string), is_global(boolean), resource_uri(string), tag(string)
- lookup_table_rows:
  - primary key: id
  - fields: data_type_id(string), fields(object), id(string), sort_key(integer)
- det_exports:
  - primary key: id
  - fields: case_type(string), det_config_url(string), export_format(string), id(string), is_deidentified(boolean), name(string), type(string), xmlns(string)
- messaging_events:
  - primary key: id
  - cursor: date
  - fields: case_id(string), content_type(string), date(string), date_last_activity(string), domain(string), error(string), form(string), id(integer), messages(array), recipient(object), source(object), status(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_case_v2:
  - endpoint: POST /a/{{ config.project_space }}/api/case/v2/
  - required fields: case_type, case_name, owner_id
  - risk: creates a CommCare case by submitting a server-generated XForm
- update_case_v2:
  - endpoint: PUT /a/{{ config.project_space }}/api/case/v2/{{ record.case_id }}
  - required fields: case_id
  - risk: updates an existing CommCare case by id
- upsert_case_v2_by_external_id:
  - endpoint: PUT /a/{{ config.project_space }}/api/case/v2/ext/{{ record.external_id }}/
  - required fields: external_id
  - risk: updates or creates a CommCare case matched by external id
- upsert_case_v2:
  - endpoint: PUT /a/{{ config.project_space }}/api/case/v2/
  - required fields: external_id
  - risk: updates or creates a CommCare case matched by the request body's external_id
- create_mobile_worker:
  - endpoint: POST /a/{{ config.project_space }}/api/user/v1/
  - required fields: username, email
  - risk: creates a mobile worker account; password-bearing creation is intentionally not represented in fixtures or docs
- update_mobile_worker:
  - endpoint: PUT /a/{{ config.project_space }}/api/user/v1/{{ record.mobile_worker_id }}/
  - required fields: mobile_worker_id
  - risk: updates a mobile worker profile and assignments
- delete_mobile_worker:
  - endpoint: DELETE /a/{{ config.project_space }}/api/user/v1/{{ record.mobile_worker_id }}/
  - required fields: mobile_worker_id
  - risk: deletes a mobile worker
- send_mobile_worker_password_reset:
  - endpoint: POST /a/{{ config.project_space }}/api/user/v1/{{ record.mobile_worker_id }}/email_password_reset/
  - required fields: mobile_worker_id
  - risk: sends a password reset email to a mobile worker
- create_web_user_invitation:
  - endpoint: POST /a/{{ config.project_space }}/api/invitation/v1/
  - required fields: email, role
  - risk: invites a web user to the project
- update_web_user:
  - endpoint: PATCH /a/{{ config.project_space }}/api/web-user/v1/{{ record.web_user_id }}/
  - required fields: web_user_id
  - risk: updates a web user's role, locations, profile, and custom data
- enable_web_user:
  - endpoint: POST /a/{{ config.project_space }}/api/web-user/v1/{{ record.web_user_id }}/enable
  - required fields: web_user_id
  - risk: enables a web user account
- disable_web_user:
  - endpoint: POST /a/{{ config.project_space }}/api/web-user/v1/{{ record.web_user_id }}/disable
  - required fields: web_user_id
  - risk: disables a web user account
- create_group:
  - endpoint: POST /a/{{ config.project_space }}/api/group/v1/
  - required fields: name
  - risk: creates a user group
- create_groups_bulk:
  - endpoint: PATCH /a/{{ config.project_space }}/api/group/v1/
  - required fields: objects
  - risk: creates multiple user groups from one request body
- update_group:
  - endpoint: PUT /a/{{ config.project_space }}/api/group/v1/{{ record.group_id }}/
  - required fields: group_id
  - risk: updates a user group and replaces provided assignments/custom metadata
- delete_group:
  - endpoint: DELETE /a/{{ config.project_space }}/api/group/v1/{{ record.group_id }}/
  - required fields: group_id
  - risk: deletes a user group
- create_location_v2:
  - endpoint: POST /a/{{ config.project_space }}/api/location/v2/
  - required fields: name, location_type_code
  - risk: creates a location in the project hierarchy
- update_location_v2:
  - endpoint: PUT /a/{{ config.project_space }}/api/location/v2/{{ record.location_id }}
  - required fields: location_id
  - risk: updates a location in the project hierarchy
- bulk_upsert_locations_v2:
  - endpoint: PATCH /a/{{ config.project_space }}/api/location/v2/
  - required fields: objects
  - risk: atomically creates and updates multiple locations
- create_lookup_table:
  - endpoint: POST /a/{{ config.project_space }}/api/lookup_table/v1/
  - required fields: tag, fields
  - risk: creates a lookup table definition
- update_lookup_table:
  - endpoint: PUT /a/{{ config.project_space }}/api/lookup_table/v1/{{ record.lookup_table_id }}
  - required fields: lookup_table_id, tag, fields
  - risk: updates a lookup table definition
- delete_lookup_table:
  - endpoint: DELETE /a/{{ config.project_space }}/api/lookup_table/v1/{{ record.lookup_table_id }}
  - required fields: lookup_table_id
  - risk: deletes a lookup table definition
- create_lookup_table_row:
  - endpoint: POST /a/{{ config.project_space }}/api/lookup_table_item/v1/
  - required fields: data_type_id, fields
  - risk: creates a lookup table row
- update_lookup_table_row:
  - endpoint: PUT /a/{{ config.project_space }}/api/lookup_table_item/v1/{{ record.lookup_table_item_id }}
  - required fields: lookup_table_item_id, data_type_id, fields
  - risk: updates a lookup table row
- delete_lookup_table_row:
  - endpoint: DELETE /a/{{ config.project_space }}/api/lookup_table_item/v1/{{ record.lookup_table_item_id }}
  - required fields: lookup_table_item_id
  - risk: deletes a lookup table row

## Security

- read risk: external CommCare HQ API reads across configured project resources and account-level user identity/domain endpoints
- write risk: external CommCare HQ mutations for cases, mobile workers, web-user invitations and access, groups, locations, lookup tables, and lookup table rows
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect commcare
```

### Inspect as structured JSON

```bash
pm connectors inspect commcare --json
```

## Agent Rules

- Run pm connectors inspect commcare before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
