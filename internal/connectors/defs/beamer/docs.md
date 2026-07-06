# Overview

Reads and writes Beamer NPS survey responses, announcement posts, feature requests, comments,
reactions, votes, and end users through the Beamer REST API.

Readable streams: `nps`, `posts`, `feature_requests`, `comments`, `post_reactions`,
`feature_request_votes`, `users`.

Write actions: `create_post`, `update_post`, `delete_post`, `create_post_comment`,
`delete_post_comment`, `create_feature_request`, `update_feature_request`, `delete_feature_request`,
`create_feature_request_comment`, `delete_feature_request_comment`, `create_post_reaction`,
`delete_post_reaction`, `create_feature_request_vote`, `delete_feature_request_vote`.

Service API documentation: https://www.getbeamer.com/api/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Beamer API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.getbeamer.com`; format `uri`; Beamer API base
  URL override for tests or proxies.
- `mode` (optional, string).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound; the nps stream's
  dateFrom filter uses this on a fresh sync (no persisted cursor yet).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.getbeamer.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use base URL `{{ config.base_url }}/v0` after applying configuration defaults.

Connection checks call GET `/nps` with query `maxResults`=`1`; `page`=`0`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `maxResults`;
starts at 0; page size 100.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `nps`: GET `/nps` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `maxResults`; starts at 0; page size 100; incremental cursor `date`; sent as `dateFrom`;
  formatted as `rfc3339`; initial lower bound from `start_date`.
- `posts`: GET `/posts` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `maxResults`; starts at 0; page size 100.
- `feature_requests`: GET `/feature-requests` - records path `.`; page-number pagination; page
  parameter `page`; size parameter `maxResults`; starts at 0; page size 100.
- `comments`: GET `/comments` - records path `.`; page-number pagination; page parameter `page`;
  size parameter `maxResults`; starts at 0; page size 100.
- `post_reactions`: GET `/posts/{{ fanout.id }}/reactions` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `maxResults`; starts at 0; page size 100;
  fan-out; ids from request `/posts`; id-list records path `.`; id field `id`; id inserted into the
  request path; stamps `post_id`.
- `feature_request_votes`: GET `/feature-requests/{{ fanout.id }}/votes` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `maxResults`; starts at 0; page size
  100; fan-out; ids from request `/feature-requests`; id-list records path `.`; id field `id`; id
  inserted into the request path; stamps `feature_request_id`.
- `users`: GET `/users` - records path `.`; page-number pagination; page parameter `page`; size
  parameter `maxResults`; starts at 0; page size 100.

## Write actions & risks

Overall write risk: external mutation of Beamer posts, feature requests, comments, reactions, and
votes; a published post/feature-request write is immediately end-user-visible in the customer-facing
widget.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_post`: POST `/posts` - kind `create`; body type `json`; required record fields `title`,
  `content`; accepted fields `category`, `content`, `date`, `dueDate`, `enableFeedback`,
  `enableReactions`, `filter`, `language`, `linkText`, `linkUrl`, `pinned`, `publish`,
  `showInStandalone`, `showInWidget`, `title`; risk: external mutation; creates a new Beamer
  announcement post, optionally published immediately (visible to end users); approval required.
- `update_post`: PUT `/posts/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `category`, `content`, `date`, `dueDate`,
  `enableFeedback`, `enableReactions`, `filter`, `id`, `pinned`, `publish`, `showInStandalone`,
  `showInWidget`, `title`; risk: external mutation; updates a live announcement post visible to end
  users; approval required.
- `delete_post`: DELETE `/posts/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; risk: permanently removes an announcement post; irreversible; approval required.
- `create_post_comment`: POST `/posts/{{ record.post_id }}/comments` - kind `create`; body type
  `json`; path fields `post_id`; required record fields `post_id`, `text`; accepted fields
  `post_id`, `text`, `userEmail`, `userFirstname`, `userId`, `userLastname`; risk: external
  mutation; adds a comment to a live announcement post on behalf of a user; approval required.
- `delete_post_comment`: DELETE `/posts/{{ record.post_id }}/comments/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `post_id`, `id`; required record fields `post_id`, `id`;
  accepted fields `id`, `post_id`; missing records treated as success for status `404`; risk:
  permanently removes a comment from a post; irreversible; approval required.
- `create_feature_request`: POST `/feature-requests` - kind `create`; body type `json`; required
  record fields `title`, `content`; accepted fields `category`, `content`, `date`, `notes`,
  `status`, `title`, `userEmail`, `userFirstname`, `userId`, `userLastname`, `visible`; risk:
  external mutation; creates a new feature request, optionally visible immediately to end users;
  approval required.
- `update_feature_request`: PUT `/feature-requests/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `category`, `content`,
  `id`, `notes`, `status`, `title`, `visible`; risk: external mutation; updates a feature request
  visible to end users (status changes are commonly user-facing); approval required.
- `delete_feature_request`: DELETE `/feature-requests/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently removes a feature request; irreversible;
  approval required.
- `create_feature_request_comment`: POST `/feature-requests/{{ record.feature_request_id
  }}/comments` - kind `create`; body type `json`; path fields `feature_request_id`; required record
  fields `feature_request_id`, `text`; accepted fields `feature_request_id`, `text`, `userEmail`,
  `userFirstname`, `userId`, `userLastname`, `visible`; risk: external mutation; adds a comment to a
  feature request on behalf of a user; approval required.
- `delete_feature_request_comment`: DELETE `/feature-requests/{{ record.feature_request_id
  }}/comments/{{ record.id }}` - kind `delete`; body type `none`; path fields `feature_request_id`,
  `id`; required record fields `feature_request_id`, `id`; accepted fields `feature_request_id`,
  `id`; missing records treated as success for status `404`; risk: permanently removes a comment
  from a feature request; irreversible; approval required.
- `create_post_reaction`: POST `/posts/{{ record.post_id }}/reactions` - kind `create`; body type
  `json`; path fields `post_id`; required record fields `post_id`, `reaction`; accepted fields
  `post_id`, `reaction`, `userEmail`, `userFirstname`, `userId`, `userLastname`; risk: external
  mutation; records a reaction to a post on behalf of a user; approval required.
- `delete_post_reaction`: DELETE `/posts/{{ record.post_id }}/reactions/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `post_id`, `id`; required record fields `post_id`, `id`;
  accepted fields `id`, `post_id`; missing records treated as success for status `404`; risk:
  permanently removes a reaction from a post; irreversible; approval required.
- `create_feature_request_vote`: POST `/feature-requests/{{ record.feature_request_id }}/votes` -
  kind `create`; body type `json`; path fields `feature_request_id`; required record fields
  `feature_request_id`; accepted fields `feature_request_id`, `userEmail`, `userFirstname`,
  `userId`, `userLastname`; risk: external mutation; records a vote for a feature request on behalf
  of a user; approval required.
- `delete_feature_request_vote`: DELETE `/feature-requests/{{ record.feature_request_id }}/votes/{{
  record.id }}` - kind `delete`; body type `none`; path fields `feature_request_id`, `id`; required
  record fields `feature_request_id`, `id`; accepted fields `feature_request_id`, `id`; missing
  records treated as success for status `404`; risk: permanently removes a vote from a feature
  request; irreversible; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 7 stream-backed endpoint group(s), 14 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=4, duplicate_of=16, non_data_endpoint=2, requires_elevated_scope=4.
