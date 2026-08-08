# pm connectors inspect teamtailor

```text
NAME
  pm connectors inspect teamtailor - Teamtailor connector manual

SYNOPSIS
  pm connectors inspect teamtailor
  pm connectors inspect teamtailor --json
  pm credentials add <name> --connector teamtailor [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Teamtailor jobs, candidates, job applications, departments, locations, roles, stages, teams, users, and regions, and writes approved recruiting mutations (jobs, candidates, job applications, departments, locations, teams, todos, notes) through the Teamtailor JSON:API.

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
  x_api_version
  api (secret)
  api_key (secret)

ETL STREAMS
  jobs:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(string), title(string)
  candidates:
    primary key: id
    fields: created_at(string), email(string), first_name(string), id(string), last_name(string)
  job_applications:
    primary key: id
    fields: candidate_id(string), created_at(string), id(string), job_id(string)
  departments:
    primary key: id
    fields: id(string), name(string)
  locations:
    primary key: id
    fields: city(string), country(string), id(string), name(string)
  roles:
    primary key: id
    fields: id(string), name(string)
  stages:
    primary key: id
    fields: id(string), name(string)
  teams:
    primary key: id
    fields: id(string), name(string)
  users:
    primary key: id
    fields: created_at(string), email(string), id(string), name(string)
  regions:
    primary key: id
    fields: id(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_job:
    endpoint: POST /jobs
    required fields: data
    risk: creates a new job posting; low-risk external mutation, no approval required
  create_candidate:
    endpoint: POST /candidates
    required fields: data
    risk: creates a new candidate record; stores personal data (name/email) about a real individual, subject to data-protection obligations
  update_candidate:
    endpoint: PATCH /candidates/{{ record.data.id }}
    required fields: data
    risk: mutates an existing candidate's personal data (name/email/pitch)
  create_job_application:
    endpoint: POST /job-applications
    required fields: data
    risk: links a candidate to a job as a new application; moves the candidate into that job's active pipeline and may trigger applicant notifications
  update_job_application:
    endpoint: PATCH /job-applications/{{ record.data.id }}
    required fields: data
    risk: mutates an existing job application (e.g. moves it to a different stage); may trigger applicant-facing notifications
  create_department:
    endpoint: POST /departments
    required fields: data
    risk: creates a new department record; low-risk external mutation, no approval required
  create_location:
    endpoint: POST /locations
    required fields: data
    risk: creates a new office/location record; low-risk external mutation, no approval required
  create_team:
    endpoint: POST /teams
    required fields: data
    risk: creates a new hiring team; low-risk external mutation, no approval required
  create_todo:
    endpoint: POST /todos
    required fields: data
    risk: creates a new to-do reminder, optionally assigned to a user against a candidate; low-risk external mutation, no approval required
  create_note:
    endpoint: POST /notes
    required fields: data
    risk: creates a new internal note on a candidate; stores potentially sensitive recruiter commentary about a real individual

SECURITY
  read risk: external Teamtailor API read of job, candidate, job-application, and organizational data, including candidate personal data
  write risk: external Teamtailor API mutation (create/update jobs, candidates, job applications, departments, locations, teams, todos, notes); candidate/note writes touch personal data about real individuals
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect teamtailor

  # Inspect as structured JSON
  pm connectors inspect teamtailor --json

AGENT WORKFLOW
  - Run pm connectors inspect teamtailor before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
