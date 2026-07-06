# pm connectors inspect northpass-lms

```text
NAME
  pm connectors inspect northpass-lms - Northpass LMS connector manual

SYNOPSIS
  pm connectors inspect northpass-lms
  pm connectors inspect northpass-lms --json
  pm credentials add <name> --connector northpass-lms [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Northpass LMS people, courses, course enrollments, and groups through the Northpass REST API. Read-only.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  people:
    primary key: id
    fields: created_at(), email(), first_name(), id(), last_name(), status(), type(), updated_at()
  courses:
    primary key: id
    fields: created_at(), id(), name(), slug(), status(), type(), updated_at()
  course_enrollments:
    primary key: id
    fields: completed_at(), course_id(), created_at(), id(), learner_id(), percentage(), status(), type(), updated_at()
  groups:
    primary key: id
    fields: created_at(), id(), name(), slug(), type(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Northpass LMS API read of learner and course data
  approval: none; read-only source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect northpass-lms

  # Inspect as structured JSON
  pm connectors inspect northpass-lms --json

AGENT WORKFLOW
  - Run pm connectors inspect northpass-lms before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
