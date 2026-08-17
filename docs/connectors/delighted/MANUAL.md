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
  api_key (secret)

ETL STREAMS
  survey_responses:
    primary key: id
    cursor: updated_at
    fields: comment(), created_at(), id(), notes(), permalink(), person(), person_properties(), score(), survey_type(), tags(), updated_at()
  people:
    primary key: id
    fields: created_at(), email(), id(), last_responded_at(), last_sent_at(), name(), next_survey_scheduled_at(), phone_number()
  bounces:
    primary key: person_id
    fields: bounced_at(), email(), name(), person_id()
  unsubscribes:
    primary key: person_id
    fields: email(), name(), person_id(), unsubscribed_at()
  metrics:
    fields: detractor_count(), detractor_percent(), nps(), passive_count(), passive_percent(), promoter_count(), promoter_percent(), response_count()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_person:
    endpoint: POST /people.json
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
