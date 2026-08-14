# pm connectors inspect breezy-hr

```text
NAME
  pm connectors inspect breezy-hr - Breezy HR connector manual

SYNOPSIS
  pm connectors inspect breezy-hr
  pm connectors inspect breezy-hr --json
  pm credentials add <name> --connector breezy-hr [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Breezy HR positions, hiring pipelines, per-position candidates, departments, categories, custom attribute definitions, questionnaires, and message templates; writes position create/update/state-change and candidate create/update/pipeline-stage-move mutations, through the Breezy v3 REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret) (required)
  company_id (secret) (required)

ETL STREAMS
  positions:
    primary key: position_id
    fields: country_id(string), country_name(string), creation_date(string), department(string), name(string), org_type(string), pipeline_id(string), position_id(string), state(string), type(string), updated_date(string)
  pipelines:
    primary key: id
    fields: id(string), name(string)
  candidates:
    primary key: id
    fields: creation_date(string), email_address(string), headline(string), id(string), name(string), origin(string), phone_number(string), position_id(string), stage(string), updated_date(string)
  departments:
    primary key: id
    fields: id(string), name(string)
  categories:
    primary key: id
    fields: id(string), name(string)
  custom_attributes_candidate:
    primary key: id
    fields: id(string), name(string), secure(boolean)
  custom_attributes_position:
    primary key: id
    fields: id(string), name(string), secure(boolean)
  questionnaires:
    primary key: id
    fields: id(string), name(string)
  templates:
    primary key: id
    fields: id(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_position:
    endpoint: POST /positions
    required fields: name, description, type, location
    risk: creates a new job opening; if not left in draft state, may become publicly visible on the company's careers page and job boards depending on the configured state
  update_position:
    endpoint: PUT /position/{{ record.position_id }}
    required fields: position_id
    risk: mutates an existing job opening's title/description/location/department; a live (published) posting's public listing reflects the change immediately
  update_position_state:
    endpoint: PUT /position/{{ record.position_id }}/state
    required fields: position_id, state
    risk: changes a position's lifecycle state (published/draft/closed/archived); setting state to published makes the job publicly visible on the company's careers page and job boards, and closed/archived stops accepting new applicants
  create_candidate:
    endpoint: POST /position/{{ record.position_id }}/candidates
    required fields: position_id, name
    risk: adds a new candidate to a position's hiring pipeline; low-risk additive mutation, no approval required
  update_candidate:
    endpoint: PUT /position/{{ record.position_id }}/candidate/{{ record.candidate_id }}
    required fields: position_id, candidate_id
    risk: mutates an existing candidate's contact/profile information
  move_candidate_stage:
    endpoint: PUT /position/{{ record.position_id }}/candidate/{{ record.candidate_id }}/stage
    required fields: position_id, candidate_id, stage_id
    risk: moves a candidate to a different pipeline stage within the SAME position (e.g. Applied to Interviewing to Hired/Disqualified); moving to a terminal stage (hired/disqualified) may trigger configured stage actions (auto-emails, webhook notifications) depending on the position's stage_actions_enabled setting

SECURITY
  read risk: external Breezy HR API read of company position, hiring pipeline, candidate, department, category, custom-attribute, questionnaire, and template metadata
  write risk: external mutation of Breezy HR positions and candidates; update_position_state can publish a position to the company's public careers page and job boards, and move_candidate_stage to a terminal stage (hired/disqualified) may trigger configured stage-action auto-emails/webhooks — every write ships with an explicit per-action risk string
  approval: none required by default; review update_position_state (publishes job postings) and move_candidate_stage (may trigger candidate-facing communications) before use in an automated pipeline
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect breezy-hr

  # Inspect as structured JSON
  pm connectors inspect breezy-hr --json

AGENT WORKFLOW
  - Run pm connectors inspect breezy-hr before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
