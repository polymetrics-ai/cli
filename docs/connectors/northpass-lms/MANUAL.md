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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  people:
    primary key: id
    fields: created_at(string), email(string), first_name(string), id(string), last_name(string), status(string), type(string), updated_at(string)
  courses:
    primary key: id
    fields: created_at(string), id(string), name(string), slug(string), status(string), type(string), updated_at(string)
  course_enrollments:
    primary key: id
    fields: completed_at(string), course_id(string), created_at(string), id(string), learner_id(string), percentage(integer), status(string), type(string), updated_at(string)
  groups:
    primary key: id
    fields: created_at(string), id(string), name(string), slug(string), type(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

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
