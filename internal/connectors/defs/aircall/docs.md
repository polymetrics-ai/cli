# Overview

Reads Aircall calls, users, contacts, numbers, teams, tags, and webhooks, and writes
user/team/contact/tag/webhook mutations plus call archive/comment/tag actions, through the Aircall
REST API.

Readable streams: `calls`, `users`, `contacts`, `numbers`, `teams`, `tags`, `webhooks`.

Write actions: `create_user`, `update_user`, `delete_user`, `create_team`, `delete_team`,
`add_user_to_team`, `remove_user_from_team`, `create_contact`, `update_contact`, `delete_contact`,
`create_tag`, `update_tag`, `delete_tag`, `create_webhook`, `update_webhook`, `delete_webhook`,
`archive_call`, `unarchive_call`, `comment_call`, `tag_call`.

Service API documentation: https://developer.aircall.io/api-references/.

## Auth setup

Connection fields:

- `api_id` (required, secret, string); Aircall API ID, used as the HTTP Basic auth username. Never
  logged.
- `api_token` (required, secret, string); Aircall API token, used as the HTTP Basic auth password.
  Never logged.
- `base_url` (optional, string); default `https://api.aircall.io/v1`; format `uri`; Aircall API base
  URL override for tests or proxies.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; only calls/contacts
  created at or after this time are read (sent as the unix-seconds `from` filter).

Secret fields are redacted in logs and write previews: `api_id`, `api_token`.

Default configuration values: `base_url=https://api.aircall.io/v1`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_id`, `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/calls` with query `per_page`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `meta.next_page_link`;
next URLs stay on the configured API host.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `calls`: GET `/calls` - records path `calls`; query `per_page`=`50`; follows a next-page URL from
  the response body; URL path `meta.next_page_link`; next URLs stay on the configured API host;
  incremental cursor `started_at`; sent as `from`; formatted as Unix-seconds timestamp; initial
  lower bound from `start_date`.
- `users`: GET `/users` - records path `users`; query `per_page`=`50`; follows a next-page URL from
  the response body; URL path `meta.next_page_link`; next URLs stay on the configured API host;
  incremental cursor `created_at`; formatted as `rfc3339`.
- `contacts`: GET `/contacts` - records path `contacts`; query `per_page`=`50`; follows a next-page
  URL from the response body; URL path `meta.next_page_link`; next URLs stay on the configured API
  host; incremental cursor `created_at`; sent as `from`; formatted as Unix-seconds timestamp;
  initial lower bound from `start_date`.
- `numbers`: GET `/numbers` - records path `numbers`; query `per_page`=`50`; follows a next-page URL
  from the response body; URL path `meta.next_page_link`; next URLs stay on the configured API host.
- `teams`: GET `/teams` - records path `teams`; query `per_page`=`50`; follows a next-page URL from
  the response body; URL path `meta.next_page_link`; next URLs stay on the configured API host.
- `tags`: GET `/tags` - records path `tags`; query `per_page`=`50`; follows a next-page URL from the
  response body; URL path `meta.next_page_link`; next URLs stay on the configured API host.
- `webhooks`: GET `/webhooks` - records path `webhooks`; query `per_page`=`50`; follows a next-page
  URL from the response body; URL path `meta.next_page_link`; next URLs stay on the configured API
  host.

## Write actions & risks

Overall write risk: external Aircall API mutation of agents, teams, contacts, tags, webhooks, and
call archive/comment/tag state; approval required for user/team/contact/webhook
create-update-delete, low-risk for additive call tagging/commenting.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user`: POST `/users` - kind `create`; body type `json`; required record fields `name`,
  `email`; accepted fields `available`, `email`, `language`, `name`, `time_zone`; risk: creates a
  new Aircall agent seat, which may consume a billable license; external mutation, approval
  required.
- `update_user`: PUT `/users/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `available`, `id`, `language`, `name`, `time_zone`,
  `wrap_up_time`; risk: mutates an existing agent's profile/availability; a visible change for that
  agent's call routing.
- `delete_user`: DELETE `/users/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently removes an Aircall agent seat; irreversible, frees the associated
  license; approval required.
- `create_team`: POST `/teams` - kind `create`; body type `json`; required record fields `name`;
  accepted fields `name`, `user_ids`; risk: creates a new team container; low-risk external
  mutation, no approval required.
- `delete_team`: DELETE `/teams/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently removes a team; agents assigned to it lose that team's
  routing/membership immediately.
- `add_user_to_team`: POST `/teams/{{ record.team_id }}/users/{{ record.user_id }}` - kind `update`;
  body type `none`; path fields `team_id`, `user_id`; required record fields `team_id`, `user_id`;
  accepted fields `team_id`, `user_id`; risk: adds an agent to a team, changing that agent's call
  routing/membership; low-risk, reversible via remove_user_from_team.
- `remove_user_from_team`: DELETE `/teams/{{ record.team_id }}/users/{{ record.user_id }}` - kind
  `delete`; body type `none`; path fields `team_id`, `user_id`; required record fields `team_id`,
  `user_id`; accepted fields `team_id`, `user_id`; missing records treated as success for status
  `404`; risk: removes an agent from a team, changing that agent's call routing/membership
  immediately.
- `create_contact`: POST `/contacts` - kind `create`; body type `json`; accepted fields
  `company_name`, `emails`, `first_name`, `information`, `is_shared`, `last_name`, `phone_numbers`;
  risk: creates a new shared/personal directory contact; low-risk external mutation, no approval
  required.
- `update_contact`: PUT `/contacts/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `company_name`, `emails`, `first_name`, `id`,
  `information`, `is_shared`, `last_name`, `phone_numbers`; risk: mutates an existing contact's
  directory record, including its full phone_numbers/emails arrays (a partial array here replaces
  the previous set).
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a directory contact; irreversible.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `name`,
  `color`; accepted fields `color`, `description`, `name`; risk: creates a new call-tagging label;
  low-risk external mutation, no approval required.
- `update_tag`: PUT `/tags/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `color`, `description`, `id`, `name`; risk: renames
  or recolors an existing tag; a visible change everywhere the tag is already applied.
- `delete_tag`: DELETE `/tags/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; risk: permanently removes a tag; it is un-applied from every call that previously carried
  it.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `url`, `events`; accepted fields `active`, `events`, `url`; risk: registers a new outbound webhook
  that will POST live call/event data to an external URL of the caller's choosing; verify the target
  endpoint before enabling.
- `update_webhook`: PUT `/webhooks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `active`, `events`, `id`, `url`; risk: mutates
  an existing webhook's target URL, subscribed events, or active state; a changed url redirects
  future event deliveries to a different endpoint.
- `delete_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: permanently removes a webhook subscription; event delivery to its target
  URL stops immediately.
- `archive_call`: PUT `/calls/{{ record.id }}/archive` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: marks a call as archived,
  hiding it from default call-list views; reversible via unarchive_call.
- `unarchive_call`: PUT `/calls/{{ record.id }}/unarchive` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: restores a previously
  archived call to default call-list views.
- `comment_call`: POST `/calls/{{ record.id }}/comments` - kind `create`; body type `json`; path
  fields `id`; body fields `content`; required record fields `id`, `content`; accepted fields
  `content`, `id`; risk: adds an internal comment note to a call record; visible to other agents
  with call access, no external side effect.
- `tag_call`: POST `/calls/{{ record.id }}/tags` - kind `update`; body type `json`; path fields
  `id`; body fields `tag_ids`; required record fields `id`, `tag_ids`; accepted fields `id`,
  `tag_ids`; risk: applies the given tags to a call; additive, does not remove tags already present.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 7 stream-backed endpoint group(s), 20 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=14, duplicate_of=13, out_of_scope=24, requires_elevated_scope=15.
