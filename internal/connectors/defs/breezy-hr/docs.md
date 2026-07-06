# Overview

Reads Breezy HR positions, hiring pipelines, per-position candidates, departments, categories,
custom attribute definitions, questionnaires, and message templates; writes position
create/update/state-change and candidate create/update/pipeline-stage-move mutations, through the
Breezy v3 REST API.

Readable streams: `positions`, `pipelines`, `candidates`, `departments`, `categories`,
`custom_attributes_candidate`, `custom_attributes_position`, `questionnaires`, `templates`.

Write actions: `create_position`, `update_position`, `update_position_state`, `create_candidate`,
`update_candidate`, `move_candidate_stage`.

Service API documentation: https://developer.breezy.hr/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Breezy HR raw API key, sent verbatim as the Authorization
  header value; never logged.
- `base_url` (optional, string); default `https://api.breezy.hr/v3`; format `uri`; Breezy API root
  override for tests or proxies. The per-company base URL is base_url + /company/<company_id>.
- `company_id` (required, secret, string); Breezy company id.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`, `company_id`.

Default configuration values: `base_url=https://api.breezy.hr/v3`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use base URL `{{ config.base_url }}/company/{{ secrets.company_id }}` after applying
configuration defaults.

Connection checks call GET `/positions`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `pipelines`, `departments`, `categories`, `custom_attributes_candidate`,
`custom_attributes_position`, `questionnaires`, `templates`; page_number: `positions`, `candidates`.

- `positions`: GET `/positions` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; computed output fields `country_id`,
  `country_name`, `position_id`, `type`.
- `pipelines`: GET `/pipelines` - records path `.`; computed output fields `id`.
- `candidates`: GET `/position/{{ fanout.id }}/candidates` - records path `.`; query
  `sort`=`updated_date`; page-number pagination; page parameter `page`; size parameter `limit`;
  starts at 1; page size 100; computed output fields `id`, `stage`; fan-out; ids from request
  `/positions`; id field `_id`; id inserted into the request path; stamps `position_id`.
- `departments`: GET `/departments` - records path `.`; computed output fields `id`.
- `categories`: GET `/categories` - records path `.`; computed output fields `id`.
- `custom_attributes_candidate`: GET `/custom-attributes/candidate` - records path `.`.
- `custom_attributes_position`: GET `/custom-attributes/position` - records path `.`.
- `questionnaires`: GET `/questionnaires` - records path `.`; computed output fields `id`.
- `templates`: GET `/templates` - records path `.`; computed output fields `id`.

## Write actions & risks

Overall write risk: external mutation of Breezy HR positions and candidates; update_position_state
can publish a position to the company's public careers page and job boards, and move_candidate_stage
to a terminal stage (hired/disqualified) may trigger configured stage-action auto-emails/webhooks -
every write ships with an explicit per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_position`: POST `/positions` - kind `create`; body type `json`; required record fields
  `name`, `description`, `type`, `location`; accepted fields `category`, `department`,
  `description`, `education`, `experience`, `location`, `name`, `pipeline_id`, `requisition_id`,
  `tags`, `type`; risk: creates a new job opening; if not left in draft state, may become publicly
  visible on the company's careers page and job boards depending on the configured state.
- `update_position`: PUT `/position/{{ record.position_id }}` - kind `update`; body type `json`;
  path fields `position_id`; required record fields `position_id`; accepted fields `category`,
  `department`, `description`, `location`, `name`, `position_id`, `tags`, `type`; risk: mutates an
  existing job opening's title/description/location/department; a live (published) posting's public
  listing reflects the change immediately.
- `update_position_state`: PUT `/position/{{ record.position_id }}/state` - kind `update`; body type
  `json`; path fields `position_id`; required record fields `position_id`, `state`; accepted fields
  `position_id`, `state`; risk: changes a position's lifecycle state
  (published/draft/closed/archived); setting state to published makes the job publicly visible on
  the company's careers page and job boards, and closed/archived stops accepting new applicants.
- `create_candidate`: POST `/position/{{ record.position_id }}/candidates` - kind `create`; body
  type `json`; path fields `position_id`; required record fields `position_id`, `name`; accepted
  fields `address`, `email_address`, `name`, `origin`, `phone_number`, `position_id`, `source`,
  `summary`, `tags`; risk: adds a new candidate to a position's hiring pipeline; low-risk additive
  mutation, no approval required.
- `update_candidate`: PUT `/position/{{ record.position_id }}/candidate/{{ record.candidate_id }}` -
  kind `update`; body type `json`; path fields `position_id`, `candidate_id`; required record fields
  `position_id`, `candidate_id`; accepted fields `address`, `candidate_id`, `email_address`,
  `headline`, `name`, `phone_number`, `position_id`, `summary`, `tags`; risk: mutates an existing
  candidate's contact/profile information.
- `move_candidate_stage`: PUT `/position/{{ record.position_id }}/candidate/{{ record.candidate_id
  }}/stage` - kind `update`; body type `json`; path fields `position_id`, `candidate_id`; body
  fields `stage_id`; required record fields `position_id`, `candidate_id`, `stage_id`; accepted
  fields `candidate_id`, `position_id`, `stage_id`; risk: moves a candidate to a different pipeline
  stage within the SAME position (e.g. Applied to Interviewing to Hired/Disqualified); moving to a
  terminal stage (hired/disqualified) may trigger configured stage actions (auto-emails, webhook
  notifications) depending on the position's stage_actions_enabled setting.

## Known limits

- API coverage includes 9 stream-backed endpoint group(s), 6 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, duplicate_of=6, non_data_endpoint=2, out_of_scope=2, requires_elevated_scope=25.
