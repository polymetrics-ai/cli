# pm connectors inspect commcare

```text
NAME
  pm connectors inspect commcare - CommCare connector manual

SYNOPSIS
  pm connectors inspect commcare
  pm connectors inspect commcare --json
  pm credentials add <name> --connector commcare [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads CommCare HQ project, application, form, case, user, group, report, location, lookup table, export, and messaging API data; writes supported JSON mutations for cases, users, groups, locations, and lookup tables.

ICON
  asset: icons/commcare.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://confluence.dimagi.com/display/commcarepublic/CommCare+HQ+APIs

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  app_id
  base_url
  case_id
  case_ids
  external_id
  fixture_item_id
  fixture_type
  form_id
  group_id
  location_id
  location_type_id
  lookup_table_id
  lookup_table_item_id
  mobile_worker_id
  processing_id
  project_space
  report_id
  web_user_id
  api_key (secret)

ETL STREAMS
  forms:
    primary key: id
    cursor: received_on
    fields: id(), received_on(), server_modified_on()
  cases:
    primary key: id
    cursor: server_modified_on
    fields: id(), received_on(), server_modified_on()
  applications:
    primary key: id
    fields: id(), is_released(), modules(), name(), version(), versions()
  application:
    primary key: id
    fields: id(), is_released(), modules(), name(), version(), versions()
  multimedia_upload_status:
    primary key: processing_id
    fields: complete(), errors(), in_celery(), processing_id(), progress(), success()
  forms_v1:
    primary key: id
    cursor: received_on
    fields: id(), received_on(), server_modified_on()
  form:
    primary key: id
    cursor: received_on
    fields: id(), received_on(), server_modified_on()
  cases_v1:
    primary key: case_id
    cursor: server_date_modified
    fields: case_id(), closed(), date_modified(), domain(), indices(), properties(), server_date_modified(), user_id(), xform_ids()
  case_v1:
    primary key: case_id
    cursor: server_date_modified
    fields: case_id(), closed(), date_modified(), domain(), indices(), properties(), server_date_modified(), user_id(), xform_ids()
  cases_v2:
    primary key: case_id
    cursor: indexed_on
    fields: case_id(), case_name(), case_type(), closed(), date_closed(), date_opened(), domain(), error(), external_id(), indexed_on(), indices(), last_modified(), owner_id(), properties(), server_last_modified()
  case_v2:
    primary key: case_id
    cursor: indexed_on
    fields: case_id(), case_name(), case_type(), closed(), date_closed(), date_opened(), domain(), error(), external_id(), indexed_on(), indices(), last_modified(), owner_id(), properties(), server_last_modified()
  case_v2_by_external_id:
    primary key: case_id
    cursor: indexed_on
    fields: case_id(), case_name(), case_type(), closed(), date_closed(), date_opened(), domain(), error(), external_id(), indexed_on(), indices(), last_modified(), owner_id(), properties(), server_last_modified()
  case_v2_bulk_by_ids:
    primary key: case_id
    cursor: indexed_on
    fields: case_id(), case_name(), case_type(), closed(), date_closed(), date_opened(), domain(), error(), external_id(), indexed_on(), indices(), last_modified(), owner_id(), properties(), server_last_modified()
  case_v2_index_children:
    primary key: case_id
    cursor: indexed_on
    fields: case_id(), case_name(), case_type(), closed(), date_closed(), date_opened(), domain(), error(), external_id(), indexed_on(), indices(), last_modified(), owner_id(), properties(), server_last_modified()
  mobile_workers:
    primary key: id
    fields: default_phone_number(), email(), first_name(), groups(), id(), last_name(), locations(), phone_numbers(), primary_location(), type(), user_data(), username()
  mobile_worker:
    primary key: id
    fields: default_phone_number(), email(), first_name(), groups(), id(), last_name(), locations(), phone_numbers(), primary_location(), type(), user_data(), username()
  bulk_users:
    primary key: id
    fields: email(), first_name(), id(), last_name(), phone_numbers(), resource_uri(), username()
  web_users:
    primary key: id
    fields: default_phone_number(), email(), first_name(), id(), is_admin(), last_name(), permissions(), phone_numbers(), resource_uri(), role(), username()
  web_user:
    primary key: id
    fields: default_phone_number(), email(), first_name(), id(), is_admin(), last_name(), permissions(), phone_numbers(), resource_uri(), role(), username()
  user_domains:
    primary key: domain_name
    fields: domain_name(), project_name()
  user_identity:
    primary key: id
    fields: email(), first_name(), id(), last_name(), username()
  groups:
    primary key: id
    fields: case_sharing(), domain(), id(), metadata(), name(), path(), reporting(), users()
  group:
    primary key: id
    fields: case_sharing(), domain(), id(), metadata(), name(), path(), reporting(), users()
  reports:
    primary key: id
    fields: columns(), filters(), id(), resource_uri(), title()
  report_data:
  locations_v1:
    primary key: location_id
    cursor: last_modified
    fields: created_at(), domain(), external_id(), id(), last_modified(), latitude(), location_data(), location_id(), location_type(), longitude(), name(), parent(), resource_uri(), site_code()
  location_v1:
    primary key: location_id
    cursor: last_modified
    fields: created_at(), domain(), external_id(), id(), last_modified(), latitude(), location_data(), location_id(), location_type(), longitude(), name(), parent(), resource_uri(), site_code()
  locations_v2:
    primary key: location_id
    cursor: last_modified
    fields: domain(), last_modified(), latitude(), location_data(), location_id(), location_type_code(), location_type_name(), longitude(), name(), parent_location_id(), site_code()
  location_v2:
    primary key: location_id
    cursor: last_modified
    fields: domain(), last_modified(), latitude(), location_data(), location_id(), location_type_code(), location_type_name(), longitude(), name(), parent_location_id(), site_code()
  location_types:
    primary key: id
    fields: administrative(), code(), domain(), id(), name(), parent(), resource_uri(), shares_cases(), view_descendants()
  location_type:
    primary key: id
    fields: administrative(), code(), domain(), id(), name(), parent(), resource_uri(), shares_cases(), view_descendants()
  fixtures:
    primary key: id
    fields: fields(), fixture_type(), id(), resource_uri()
  fixture_table_items:
    primary key: id
    fields: fields(), fixture_type(), id(), resource_uri()
  fixture_item:
    primary key: id
    fields: fields(), fixture_type(), id(), resource_uri()
  lookup_tables:
    primary key: id
    fields: fields(), id(), is_global(), resource_uri(), tag()
  lookup_table_rows:
    primary key: id
    fields: data_type_id(), fields(), id(), sort_key()
  det_exports:
    primary key: id
    fields: case_type(), det_config_url(), export_format(), id(), is_deidentified(), name(), type(), xmlns()
  messaging_events:
    primary key: id
    cursor: date
    fields: case_id(), content_type(), date(), date_last_activity(), domain(), error(), form(), id(), messages(), recipient(), source(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_case_v2:
    endpoint: POST /a/{{ config.project_space }}/api/case/v2/
    risk: creates a CommCare case by submitting a server-generated XForm
  update_case_v2:
    endpoint: PUT /a/{{ config.project_space }}/api/case/v2/{{ record.case_id }}
    required fields: case_id
    risk: updates an existing CommCare case by id
  upsert_case_v2_by_external_id:
    endpoint: PUT /a/{{ config.project_space }}/api/case/v2/ext/{{ record.external_id }}/
    required fields: external_id
    risk: updates or creates a CommCare case matched by external id
  upsert_case_v2:
    endpoint: PUT /a/{{ config.project_space }}/api/case/v2/
    risk: updates or creates a CommCare case matched by the request body's external_id
  create_mobile_worker:
    endpoint: POST /a/{{ config.project_space }}/api/user/v1/
    risk: creates a mobile worker account; password-bearing creation is intentionally not represented in fixtures or docs
  update_mobile_worker:
    endpoint: PUT /a/{{ config.project_space }}/api/user/v1/{{ record.mobile_worker_id }}/
    required fields: mobile_worker_id
    risk: updates a mobile worker profile and assignments
  delete_mobile_worker:
    endpoint: DELETE /a/{{ config.project_space }}/api/user/v1/{{ record.mobile_worker_id }}/
    required fields: mobile_worker_id
    risk: deletes a mobile worker
  send_mobile_worker_password_reset:
    endpoint: POST /a/{{ config.project_space }}/api/user/v1/{{ record.mobile_worker_id }}/email_password_reset/
    required fields: mobile_worker_id
    risk: sends a password reset email to a mobile worker
  create_web_user_invitation:
    endpoint: POST /a/{{ config.project_space }}/api/invitation/v1/
    risk: invites a web user to the project
  update_web_user:
    endpoint: PATCH /a/{{ config.project_space }}/api/web-user/v1/{{ record.web_user_id }}/
    required fields: web_user_id
    risk: updates a web user's role, locations, profile, and custom data
  enable_web_user:
    endpoint: POST /a/{{ config.project_space }}/api/web-user/v1/{{ record.web_user_id }}/enable
    required fields: web_user_id
    risk: enables a web user account
  disable_web_user:
    endpoint: POST /a/{{ config.project_space }}/api/web-user/v1/{{ record.web_user_id }}/disable
    required fields: web_user_id
    risk: disables a web user account
  create_group:
    endpoint: POST /a/{{ config.project_space }}/api/group/v1/
    risk: creates a user group
  create_groups_bulk:
    endpoint: PATCH /a/{{ config.project_space }}/api/group/v1/
    risk: creates multiple user groups from one request body
  update_group:
    endpoint: PUT /a/{{ config.project_space }}/api/group/v1/{{ record.group_id }}/
    required fields: group_id
    risk: updates a user group and replaces provided assignments/custom metadata
  delete_group:
    endpoint: DELETE /a/{{ config.project_space }}/api/group/v1/{{ record.group_id }}/
    required fields: group_id
    risk: deletes a user group
  create_location_v2:
    endpoint: POST /a/{{ config.project_space }}/api/location/v2/
    risk: creates a location in the project hierarchy
  update_location_v2:
    endpoint: PUT /a/{{ config.project_space }}/api/location/v2/{{ record.location_id }}
    required fields: location_id
    risk: updates a location in the project hierarchy
  bulk_upsert_locations_v2:
    endpoint: PATCH /a/{{ config.project_space }}/api/location/v2/
    risk: atomically creates and updates multiple locations
  create_lookup_table:
    endpoint: POST /a/{{ config.project_space }}/api/lookup_table/v1/
    risk: creates a lookup table definition
  update_lookup_table:
    endpoint: PUT /a/{{ config.project_space }}/api/lookup_table/v1/{{ record.lookup_table_id }}
    required fields: lookup_table_id
    risk: updates a lookup table definition
  delete_lookup_table:
    endpoint: DELETE /a/{{ config.project_space }}/api/lookup_table/v1/{{ record.lookup_table_id }}
    required fields: lookup_table_id
    risk: deletes a lookup table definition
  create_lookup_table_row:
    endpoint: POST /a/{{ config.project_space }}/api/lookup_table_item/v1/
    risk: creates a lookup table row
  update_lookup_table_row:
    endpoint: PUT /a/{{ config.project_space }}/api/lookup_table_item/v1/{{ record.lookup_table_item_id }}
    required fields: lookup_table_item_id
    risk: updates a lookup table row
  delete_lookup_table_row:
    endpoint: DELETE /a/{{ config.project_space }}/api/lookup_table_item/v1/{{ record.lookup_table_item_id }}
    required fields: lookup_table_item_id
    risk: deletes a lookup table row

SECURITY
  read risk: external CommCare HQ API reads across configured project resources and account-level user identity/domain endpoints
  write risk: external CommCare HQ mutations for cases, mobile workers, web-user invitations and access, groups, locations, lookup tables, and lookup table rows
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect commcare

  # Inspect as structured JSON
  pm connectors inspect commcare --json

AGENT WORKFLOW
  - Run pm connectors inspect commcare before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
