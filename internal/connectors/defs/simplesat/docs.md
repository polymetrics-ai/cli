# Overview

Reads and writes Simplesat surveys, answers, questions, customers, and responses (including nested
ticket data) through the Simplesat v1 API.

Readable streams: `answers`, `surveys`, `questions`, `customers`, `responses`.

Write actions: `create_or_update_customer`, `update_customer`, `create_or_update_team_member`,
`update_answer`, `create_or_update_response`, `update_response`, `send_survey_email`.

Service API documentation: https://developer.simplesat.io/api.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Simplesat API token, sent as the X-Simplesat-Token header.
  Never logged.
- `base_url` (optional, string); default `https://api.simplesat.io/api/v1`; format `uri`; Simplesat
  v1 API base URL override for tests or proxies.
- `created_after` (optional, string); Optional customers-stream created_after (RFC3339) query
  filter, passed through verbatim.
- `page_size` (optional, integer); default `100`; Number of records requested per page (page_size
  query param), 1-250 (customers) or provider default elsewhere.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.simplesat.io/api/v1`, `page_size=100`.

Authentication behavior:

- API key authentication in `X-Simplesat-Token` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/surveys`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next`; next URLs stay
on the configured API host.

- `answers`: POST `/answers/search` - records path `answers`; query `page_size` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host; emits passthrough records.
- `surveys`: GET `/surveys` - records path `surveys`; query `page_size` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host; emits passthrough records.
- `questions`: GET `/questions` - records path `questions`; query `page_size` from template `{{
  config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host; emits passthrough records.
- `customers`: GET `/customers` - records path `customers`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `page_size` from template `{{ config.page_size }}`,
  default `100`; follows a next-page URL from the response body; URL path `next`; next URLs stay on
  the configured API host; emits passthrough records.
- `responses`: POST `/responses/search` - records path `responses`; query `page_size` from template
  `{{ config.page_size }}`, default `100`; follows a next-page URL from the response body; URL path
  `next`; next URLs stay on the configured API host; emits passthrough records.

## Write actions & risks

Overall write risk: creates/updates customers and team members, updates individual answers and
survey responses, and can trigger a live survey-invitation email to a real customer inbox
(send_survey_email).

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_or_update_customer`: POST `/customers` - kind `upsert`; body type `json`; accepted fields
  `company`, `custom_attributes`, `email`, `external_id`, `language`, `name`, `tags`; risk: creates
  a new customer or updates the existing one matched by external_id/email; low-risk external
  mutation, no approval required.
- `update_customer`: PUT `/customers/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `company`, `custom_attributes`, `external_id`,
  `id`, `language`, `name`, `tags`; risk: mutates an existing customer's profile fields by id;
  overwrites tags/custom_attributes wholesale with the submitted value.
- `create_or_update_team_member`: POST `/team-members` - kind `upsert`; body type `json`; accepted
  fields `custom_attributes`, `email`, `external_id`, `name`; risk: creates a new team member or
  updates the existing one matched by external_id/email; low-risk external mutation, no approval
  required.
- `update_answer`: PUT `/answers/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `choice`, `choices`, `comment`,
  `follow_up_answer`, `follow_up_answer_choice`, `follow_up_answer_choices`, `id`; risk: mutates an
  existing survey answer's recorded choice/comment/follow-up fields; changes the customer-submitted
  response data an already-collected survey answer represents.
- `create_or_update_response`: POST `/responses/create-or-update` - kind `upsert`; body type `json`;
  required record fields `survey_id`; accepted fields `answers`, `created`, `customer`, `language`,
  `survey_id`, `tags`, `team_members`, `ticket`; risk: creates a new survey response (or updates one
  matched by the API's own dedup rule) including its nested answers/customer/ticket/team_members
  sub-objects; commonly used to import or backfill historical survey data with an explicit created
  timestamp.
- `update_response`: PUT `/responses/{{ record.id }}/update` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `survey_id`; accepted fields `answers`, `created`, `id`,
  `language`, `survey_id`, `tags`, `team_members`; risk: mutates an existing survey response's
  tags/answers/team_members by id; overwrites the identified response's recorded data.
- `send_survey_email`: POST `/surveys/{{ record.survey_token }}/email` - kind `custom`; body type
  `json`; path fields `survey_token`; required record fields `survey_token`, `customer`; accepted
  fields `customer`, `survey_token`, `team_member`, `ticket`; risk: sends a live survey invitation
  email to the named customer's real inbox; each call generates one outbound email delivery, not a
  reversible data mutation.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s), 7 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=4, out_of_scope=1.
