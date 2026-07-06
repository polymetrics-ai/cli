# Overview

Reads Teamtailor jobs, candidates, job applications, departments, locations, roles, stages, teams,
users, and regions, and writes approved recruiting mutations (jobs, candidates, job applications,
departments, locations, teams, todos, notes) through the Teamtailor JSON:API.

Readable streams: `jobs`, `candidates`, `job_applications`, `departments`, `locations`, `roles`,
`stages`, `teams`, `users`, `regions`.

Write actions: `create_job`, `create_candidate`, `update_candidate`, `create_job_application`,
`update_job_application`, `create_department`, `create_location`, `create_team`, `create_todo`,
`create_note`.

Service API documentation: https://docs.teamtailor.com/.

## Auth setup

Connection fields:

- `api` (optional, secret, string); Teamtailor API key, sent as the Authorization header (Token
  token=<api>). Takes precedence over api_key when both are set. Never logged.
- `api_key` (optional, secret, string); Teamtailor API key (secret fallback), sent as the
  Authorization header (Token token=<api_key>) when 'api' is unset. Never logged.
- `base_url` (optional, string); default `https://api.teamtailor.com/v1`; format `uri`; Teamtailor
  API base URL override for tests or proxies.
- `x_api_version` (optional, string); Optional Teamtailor API version, sent as the X-Api-Version
  header when set.

Secret fields are redacted in logs and write previews: `api`, `api_key`.

Default configuration values: `base_url=https://api.teamtailor.com/v1`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token token=` using `secrets.api` when `{{
  secrets.api }}`.
- API key authentication in `Authorization` with prefix `Token token=` using `secrets.api_key` when
  `{{ secrets.api_key }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/jobs`.

## Streams notes

Default pagination: page-number pagination; page parameter `page[number]`; size parameter
`page[size]`; starts at 1; page size 100.

- `jobs`: GET `/jobs` - records path `data`; page-number pagination; page parameter `page[number]`;
  size parameter `page[size]`; starts at 1; page size 100; computed output fields `created_at`,
  `title`.
- `candidates`: GET `/candidates` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 1; page size 100; computed output fields
  `created_at`, `email`, `first_name`, `last_name`.
- `job_applications`: GET `/job-applications` - records path `data`; page-number pagination; page
  parameter `page[number]`; size parameter `page[size]`; starts at 1; page size 100; computed output
  fields `candidate_id`, `created_at`, `job_id`.
- `departments`: GET `/departments` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 1; page size 100; computed output fields
  `name`.
- `locations`: GET `/locations` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 1; page size 100; computed output fields
  `city`, `country`, `name`.
- `roles`: GET `/roles` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 1; page size 100; computed output fields
  `name`.
- `stages`: GET `/stages` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 1; page size 100; computed output fields
  `name`.
- `teams`: GET `/teams` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 1; page size 100; computed output fields
  `name`.
- `users`: GET `/users` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 1; page size 100; computed output fields
  `created_at`, `email`, `name`.
- `regions`: GET `/regions` - records path `data`; page-number pagination; page parameter
  `page[number]`; size parameter `page[size]`; starts at 1; page size 100; computed output fields
  `name`.

## Write actions & risks

Overall write risk: external Teamtailor API mutation (create/update jobs, candidates, job
applications, departments, locations, teams, todos, notes); candidate/note writes touch personal
data about real individuals.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_job`: POST `/jobs` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: creates a new job posting; low-risk external mutation, no approval
  required.
- `create_candidate`: POST `/candidates` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: creates a new candidate record; stores personal data
  (name/email) about a real individual, subject to data-protection obligations.
- `update_candidate`: PATCH `/candidates/{{ record.data.id }}` - kind `update`; body type `json`;
  required record fields `data`; accepted fields `data`; risk: mutates an existing candidate's
  personal data (name/email/pitch).
- `create_job_application`: POST `/job-applications` - kind `create`; body type `json`; required
  record fields `data`; accepted fields `data`; risk: links a candidate to a job as a new
  application; moves the candidate into that job's active pipeline and may trigger applicant
  notifications.
- `update_job_application`: PATCH `/job-applications/{{ record.data.id }}` - kind `update`; body
  type `json`; required record fields `data`; accepted fields `data`; risk: mutates an existing job
  application (e.g. moves it to a different stage); may trigger applicant-facing notifications.
- `create_department`: POST `/departments` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: creates a new department record; low-risk external mutation,
  no approval required.
- `create_location`: POST `/locations` - kind `create`; body type `json`; required record fields
  `data`; accepted fields `data`; risk: creates a new office/location record; low-risk external
  mutation, no approval required.
- `create_team`: POST `/teams` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: creates a new hiring team; low-risk external mutation, no approval
  required.
- `create_todo`: POST `/todos` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: creates a new to-do reminder, optionally assigned to a user against
  a candidate; low-risk external mutation, no approval required.
- `create_note`: POST `/notes` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: creates a new internal note on a candidate; stores potentially
  sensitive recruiter commentary about a real individual.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, destructive_admin=8, duplicate_of=34, non_data_endpoint=3, out_of_scope=45,
  requires_elevated_scope=12.
