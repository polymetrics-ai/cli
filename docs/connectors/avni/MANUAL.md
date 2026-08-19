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
  username (required)
  password (secret) (required)

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
  Other Commands
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
