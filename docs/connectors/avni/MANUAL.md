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
    fields: id(), name(), updated_at()
  encounters:
    primary key: id
    cursor: updated_at
    fields: encounter_type(), id(), subject_id(), updated_at()
  program_enrolments:
    primary key: id
    cursor: updated_at
    fields: enrolment_date_time(), exit_date_time(), id(), program(), subject_id(), updated_at()
  program_encounters:
    primary key: id
    cursor: updated_at
    fields: encounter_date_time(), encounter_type(), enrolment_id(), id(), program(), subject_id(), updated_at()
  group_subjects:
    primary key: id
    cursor: updated_at
    fields: group_subject_id(), id(), member_subject_id(), membership_end_date(), membership_start_date(), updated_at()
  locations:
    primary key: id
    cursor: updated_at
    fields: id(), level(), parent_id(), title(), type(), updated_at()
  approval_statuses:
    primary key: entity_id, entity_type
    cursor: status_date_time
    fields: approval_status(), approval_status_comment(), entity_id(), entity_type(), entity_type_id(), status_date_time()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Avni API read of subjects and encounters
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

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
