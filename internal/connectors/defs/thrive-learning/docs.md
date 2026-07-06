# Overview

Reads users, content, completions, assignments, audiences, tags, CPD records, and activity data
through the Thrive Learning Public API.

Readable streams: `users`, `content`, `completions`, `activities`, `contents_v1`,
`learning_completions`, `assignments`, `assignment_enrolments`, `audiences`, `audience_members`,
`audience_managers`, `tags`, `cpd_categories`, `cpd_entries`, `cpd_requirements`, `skill_levels`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.thrivelearning.com/apidocs/thrive-api-v1.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.thrivelearning.com`; format `uri`; Thrive
  Learning API base URL override for tests or proxies.
- `password` (required, secret, string); Thrive Learning API password, sent via HTTP Basic auth.
  Never logged.
- `start_date` (optional, string); Optional lower-bound timestamp sent as the updated_since query
  parameter on every stream; when unset, no updated_since filter is applied.
- `username` (required, string); Thrive Learning tenant username, sent via HTTP Basic auth.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://api.thrivelearning.com`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Pagination by stream: none: `audience_managers`, `tags`, `skill_levels`; page_number: `users`,
`content`, `completions`, `activities`, `contents_v1`, `learning_completions`, `assignments`,
`assignment_enrolments`, `audiences`, `audience_members`, `cpd_categories`, `cpd_entries`,
`cpd_requirements`.

- `users`: GET `/users` - records path `items`; query `updated_since` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `content`: GET `/content` - records path `items`; query `updated_since` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `completions`: GET `/completions` - records path `items`; query `updated_since` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `activities`: GET `https://public.api.learn.link/rest/v1/activities` - records path `results`;
  page-number pagination; page parameter `page`; size parameter `perPage`; starts at 1; page size
  1000.
- `contents_v1`: GET `https://public.api.learn.link/rest/v1/contents` - records path `results`;
  query `updatedSince` from template `{{ config.start_date }}`, omitted when absent; page-number
  pagination; page parameter `page`; size parameter `perPage`; starts at 1; page size 1000.
- `learning_completions`: GET `https://public.api.learn.link/rest/v1/learning/completions` - records
  path `.`; page-number pagination; page parameter `page`; size parameter `perPage`; starts at 1;
  page size 1000.
- `assignments`: GET `https://public.api.learn.link/rest/v1/assignments` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `perPage`; starts at 1; page size
  1000.
- `assignment_enrolments`: GET `https://public.api.learn.link/rest/v1/assignments/{{ fanout.id
  }}/enrolments` - records path `.`; page-number pagination; page parameter `page`; size parameter
  `perPage`; starts at 1; page size 1000; fan-out; ids from request
  `https://public.api.learn.link/rest/v1/assignments`; id-list records path `.`; id field `id`; id
  inserted into the request path; stamps `assignment_id`.
- `audiences`: GET `https://public.api.learn.link/rest/v1/audiences` - records path `results`;
  page-number pagination; page parameter `page`; size parameter `perPage`; starts at 1; page size
  1000.
- `audience_members`: GET `https://public.api.learn.link/rest/v1/audiences/{{ fanout.id }}/members`
  - records path `results`; page-number pagination; page parameter `page`; size parameter `perPage`;
  starts at 1; page size 1000; fan-out; ids from request
  `https://public.api.learn.link/rest/v1/audiences`; id-list records path `results`; id field `id`;
  id inserted into the request path; stamps `audience_id`.
- `audience_managers`: GET `https://public.api.learn.link/rest/v1/audiences/{{ fanout.id
  }}/managers` - records path `.`; fan-out; ids from request
  `https://public.api.learn.link/rest/v1/audiences`; id-list records path `results`; id field `id`;
  id inserted into the request path; stamps `audience_id`.
- `tags`: GET `https://public.api.learn.link/rest/v1/tags` - records path `results`.
- `cpd_categories`: GET `https://public.api.learn.link/rest/v1/cpdCategories` - records path
  `results`; page-number pagination; page parameter `page`; size parameter `perPage`; starts at 1;
  page size 1000.
- `cpd_entries`: GET `https://public.api.learn.link/rest/v1/cpdEntries` - records path `results`;
  page-number pagination; page parameter `page`; size parameter `perPage`; starts at 1; page size
  1000.
- `cpd_requirements`: GET `https://public.api.learn.link/rest/v1/cpdRequirementSummaries` - records
  path `results`; page-number pagination; page parameter `page`; size parameter `perPage`; starts at
  1; page size 1000.
- `skill_levels`: GET `https://public.api.learn.link/rest/v1/skills/levels` - records path `.`.

## Write actions & risks

This connector is read-only. Read behavior: external Thrive Learning API read of user, content,
completion, assignment, audience, tag, and CPD data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 16 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=1, duplicate_of=13, non_data_endpoint=2, out_of_scope=23.
