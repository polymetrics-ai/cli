# pm connectors inspect circleci

```text
NAME
  pm connectors inspect circleci - CircleCI connector manual

SYNOPSIS
  pm connectors inspect circleci
  pm connectors inspect circleci --json
  pm credentials add <name> --connector circleci [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes CircleCI projects, pipelines, workflows, jobs, contexts, schedules, environment variables, checkout keys, and workflow insights through the CircleCI v2 REST API.

ICON
  id: simple-icons-circleci
  asset: icons/simple-icons/circleci.svg
  title: CircleCI
  simple_icon_slug: circleci
  simple_icon_hex: 343434
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=CircleCI
  match: exact-name-or-slug
  matched_by: circleci

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  org
  pipeline_id
  repo
  vcs_type
  workflow_id
  api_key (secret) (required)

ETL STREAMS
  projects:
    primary key: id
    fields: default_branch(string), id(string), name(string), organization_id(string), organization_name(string), organization_slug(string), slug(string), vcs_url(string)
  pipelines:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(string), number(integer), project_slug(string), state(string), updated_at(string)
  workflows:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(string), name(string), pipeline_id(string), pipeline_number(integer), project_slug(string), status(string), stopped_at(string)
  jobs:
    primary key: id
    cursor: started_at
    fields: id(string), job_number(integer), name(string), project_slug(string), started_at(string), status(string), stopped_at(string), type(string)
  contexts:
    primary key: id
    fields: created_at(string), id(string), name(string)
  schedules:
    primary key: id
    cursor: updated-at
    fields: actor(object), created-at(string), description(string), id(string), name(string), parameters(object), project-slug(string), timetable(object), updated-at(string)
  checkout_keys:
    primary key: fingerprint
    fields: created-at(string), fingerprint(string), preferred(boolean), public-key(string), type(string)
  environment_variables:
    primary key: name
    fields: created-at(string), name(string), value(string)
  insights_workflow_summary:
    primary key: name
    fields: metrics(object), name(string), project_id(string), window_end(string), window_start(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_schedule:
    endpoint: POST /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/schedule
    required fields: name, timetable, attribution-actor, parameters
    risk: external mutation; creates a new scheduled-pipeline trigger for this project
  update_schedule:
    endpoint: PATCH /schedule/{{ record.id }}
    required fields: id
    risk: external mutation; updates an existing scheduled-pipeline trigger's timetable or parameters
  delete_schedule:
    endpoint: DELETE /schedule/{{ record.id }}
    required fields: id
    risk: irreversible external deletion of a scheduled-pipeline trigger; approval required
  create_environment_variable:
    endpoint: POST /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/envvar
    required fields: name, value
    risk: external mutation; creates or overwrites a project environment variable used by every future CI run
  delete_environment_variable:
    endpoint: DELETE /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/envvar/{{ record.name }}
    required fields: name
    risk: irreversible external deletion of a project environment variable; may break future CI runs that depend on it; approval required
  create_checkout_key:
    endpoint: POST /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/checkout-key
    required fields: type
    risk: external mutation; creates a new deploy/checkout SSH key with repository access
  delete_checkout_key:
    endpoint: DELETE /project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}/checkout-key/{{ record.fingerprint }}
    required fields: fingerprint
    risk: irreversible external revocation of a deploy/checkout SSH key; may break future CI checkouts that depend on it; approval required

SECURITY
  read risk: external CircleCI API read of CI project, pipeline, workflow, job, context, schedule, environment-variable, checkout-key, and workflow-insight metadata
  write risk: external mutation of CircleCI project configuration: schedule/environment-variable/checkout-key create and delete; never triggers, cancels, or approves a live CI run
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

EXAMPLES
  # Inspect as a manual
  pm connectors inspect circleci

  # Inspect as structured JSON
  pm connectors inspect circleci --json

AGENT WORKFLOW
  - Run pm connectors inspect circleci before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
