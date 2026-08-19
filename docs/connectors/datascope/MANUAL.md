# pm connectors inspect datascope

```text
NAME
  pm connectors inspect datascope - DataScope connector manual

SYNOPSIS
  pm connectors inspect datascope
  pm connectors inspect datascope --json
  pm credentials add <name> --connector datascope [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads DataScope locations, form answers, lists, notifications, task assignments, tickets (findings), and generated files, and writes location/list/task-assignment/form-answer mutations, through the DataScope external REST API (full-refresh).

ICON
  id: datascope
  asset: icons/datascope.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://app.mydatascope.com/api/external/docs/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  locations:
    primary key: id
    fields: address(string), city(string), code(string), company_code(string), company_name(string), country(string), description(string), id(integer), latitude(number), longitude(number), name(string), phone(string), region(string)
  answers:
    primary key: form_answer_id
    cursor: created_at
    fields: code(string), created_at(string), form_answer_id(integer), form_id(integer), form_name(string), form_state(string), latitude(number), longitude(number), user_identifier(string), user_name(string)
  lists:
    primary key: id
    cursor: updated_at
    fields: account_id(integer), attribute1(string), attribute2(string), code(string), created_at(string), description(string), id(integer), list_id(integer), name(string), updated_at(string)
  notifications:
    primary key: id
    cursor: created_at
    fields: created_at(string), form_code(string), form_name(string), id(integer), type(string), url(string), user(string)
  task_assigns:
    primary key: id
    cursor: created_at
    fields: assign_id(string), completed(string), completed_datetime(string), confirmation_status(string), created_at(string), created_by(string), description(string), form_name(string), gap(integer), id(integer), location_address(string), location_code(string), location_email(string), location_latitude(number), location_longitude(number), location_name(string), location_phone(string), mandatory(string), on_time(string), priority(integer), response_code(string), response_end(string), response_start(string), start_time(string), status(string), time_to_perform_minutes(number), user_email(string)
  findings:
    primary key: id
    fields: closure_date(string), closure_message(string), code(integer), creation_date(string), creator_email(string), creator_id(integer), creator_name(string), description(string), expiration_date(string), form_answer_code(string), form_answer_id(integer), id(string), last_updated_by(string), location_code(string), location_id(integer), location_name(string), name(string), priority(string), status(string), task_form_question(string), task_form_title(string), type(string)
  files:
    primary key: id
    fields: form_code(string), form_name(string), id(integer), type(string), url(string), user(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_location:
    endpoint: POST /locations
    required fields: name
    risk: creates a new field-data-collection location record; low-risk external mutation, no approval required
  update_location:
    endpoint: POST /locations/{{ record.id }}
    required fields: id
    risk: mutates an existing location's address/contact metadata; external mutation, approval required
  assign_task:
    endpoint: POST /assign_task
    required fields: form_id, user_id, date
    risk: assigns a new field task/inspection to a user for a scheduled date; low-risk external mutation, no approval required
  create_metadata_object:
    endpoint: POST /metadata_object
    required fields: metadata_type, name
    risk: creates a new list (metadata object) element; low-risk external mutation, no approval required
  update_metadata_object:
    endpoint: POST /metadata_object/{{ record.id }}
    required fields: id
    risk: mutates an existing list element's fields; external mutation, approval required
  bulk_update_metadata_objects:
    endpoint: POST /metadata_objects/bulk_update
    required fields: metadata_type, list_objects
    risk: replaces/updates many list elements of one metadata_type in a single call; higher blast radius than a single-object update, approval required
  create_metadata_type:
    endpoint: POST /metadata_types
    required fields: name
    risk: creates a new empty list (metadata type/category); low-risk external mutation, no approval required
  update_metadata_type:
    endpoint: POST /metadata_types/{{ record.id }}
    required fields: id
    risk: renames/reconfigures an existing list definition; every list element under it is affected, external mutation, approval required
  change_form_answer:
    endpoint: POST /change_form_answer
    required fields: form_name, form_code, question_name, question_value
    risk: overwrites a previously-submitted form answer's value in place, rewriting collected field data after the fact; external mutation, approval required

SECURITY
  read risk: external DataScope API read of field-data-collection form submissions, location data, task assignments, tickets, and generated files
  write risk: external mutation of DataScope locations, lists (metadata objects/types), task assignments, and previously-submitted form answers; change_form_answer rewrites collected field data after the fact and bulk_update_metadata_objects affects many list elements in one call, so every write ships an explicit per-action risk string
  approval: required for update_location/update_metadata_object/update_metadata_type/bulk_update_metadata_objects/change_form_answer; create_location/assign_task/create_metadata_object/create_metadata_type are low-risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect datascope

  # Inspect as structured JSON
  pm connectors inspect datascope --json

AGENT WORKFLOW
  - Run pm connectors inspect datascope before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
