# pm connectors inspect survicate

```text
NAME
  pm connectors inspect survicate - Survicate connector manual

SYNOPSIS
  pm connectors inspect survicate
  pm connectors inspect survicate --json
  pm credentials add <name> --connector survicate [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Survicate surveys, survey questions, responses, and respondent attributes, and manages GDPR personal-data requests, through the Survicate Data Export API v2. Read-only.

ICON
  id: survicate
  asset: icons/survicate.svg
  source: official
  review_status: official_verified
  review_url: https://developers.survicate.com/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  surveys:
    primary key: id
    cursor: updated_at
    fields: created_at(string), id(string), name(string), updated_at(string)
  survey_questions:
    primary key: survey_id, id
    fields: answer_choices(array), columns(array), fields(array), id(integer), introduction(string), question(string), survey_id(string), type(string)
  responses:
    primary key: uuid
    fields: attributes(array), collected_at(string), device_type(string), language(string), operating_system(string), questions(array), respondent_uuid(string), survey_id(string), url(string), uuid(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Survicate API read of survey, response, and respondent data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect survicate

  # Inspect as structured JSON
  pm connectors inspect survicate --json

AGENT WORKFLOW
  - Run pm connectors inspect survicate before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
