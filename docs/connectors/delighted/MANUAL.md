# pm connectors inspect delighted

```text
NAME
  pm connectors inspect delighted - Delighted connector manual

SYNOPSIS
  pm connectors inspect delighted
  pm connectors inspect delighted --json
  pm credentials add <name> --connector delighted [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Delighted survey responses, people, bounces, unsubscribes, and aggregate metrics through the Delighted REST API; can create/update and delete people.

ICON
  id: delighted
  asset: icons/delighted.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://delighted.com/docs/api

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  start_date
  api_key (secret) (required)

ETL STREAMS
  survey_responses:
    primary key: id
    cursor: updated_at
    fields: comment(string), created_at(integer), id(string), notes(array), permalink(string), person(string), person_properties(object), score(integer), survey_type(string), tags(array), updated_at(integer)
  people:
    primary key: id
    fields: created_at(integer), email(string), id(string), last_responded_at(integer), last_sent_at(integer), name(string), next_survey_scheduled_at(integer), phone_number(string)
  bounces:
    primary key: person_id
    fields: bounced_at(integer), email(string), name(string), person_id(string)
  unsubscribes:
    primary key: person_id
    fields: email(string), name(string), person_id(string), unsubscribed_at(integer)
  metrics:
    fields: detractor_count(integer), detractor_percent(number), nps(integer), passive_count(integer), passive_percent(number), promoter_count(integer), promoter_percent(number), response_count(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_person:
    endpoint: POST /people.json
    required fields: email
    risk: creates or updates a Delighted person and may trigger survey workflow depending on account settings
  delete_person:
    endpoint: DELETE /people/{{ record.person_id }}.json
    required fields: person_id
    risk: deletes a Delighted person record

SECURITY
  read risk: external Delighted API read of survey responses, people, and aggregate NPS metrics
  write risk: creates/updates Delighted people and deletes existing people
  approval: reverse-ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect delighted

  # Inspect as structured JSON
  pm connectors inspect delighted --json

AGENT WORKFLOW
  - Run pm connectors inspect delighted before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
