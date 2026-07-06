# Overview

Reads AppFollow account users, app collections, app lists, reviews, review summaries,
ratings/ratings history, ASO keywords, rankings, and version/what's-new metadata through the
AppFollow REST API v2 (config-list-driven fan-out per app/collection); writes review
replies/tags/notes, ASO keyword edits, and account user/app/collection management actions.

Readable streams: `users`, `app_collections`, `app_lists`, `ratings`, `reviews`, `reviews_summary`,
`keywords`, `rankings`, `versions`, `versions_whatsnew`, `ratings_history`.

Write actions: `reply_to_review`, `update_review_tags`, `update_review_notes`, `edit_keywords`,
`add_user`, `update_user`, `remove_user`, `add_collection`, `remove_collection`, `add_app`,
`remove_app`.

Service API documentation: https://docs.api.appfollow.io/reference/overview.

## Auth setup

Connection fields:

- `api_secret` (required, secret, string); AppFollow API token, sent as the X-AppFollow-API-Token
  header. Never logged.
- `app_collection_ids` (optional, string); Comma-separated list of AppFollow app collection ids to
  fan out over for the app_lists stream (one GET /account/apps/app request per id, forwarded as
  apps_id).
- `base_url` (optional, string); default `https://api.appfollow.io/api/v2`; format `uri`; AppFollow
  API base URL override for tests or proxies.
- `ext_ids` (optional, string); Comma-separated list of AppFollow app external ids (ext_id) to fan
  out over for the
  ratings/reviews/reviews_summary/keywords/rankings/versions/versions_whatsnew/ratings_history
  streams (one request per id, forwarded as ext_id). Required for all of those streams.
- `report_country` (optional, string); default `us`; Two-letter store country code for the versions
  stream's required country filter.
- `report_from` (optional, string); format `date`; Start date (YYYY-MM-DD) for the reviews and
  reviews_summary streams' required from/to date-range filter.
- `report_store` (optional, string); default `itunes`; Store code (itunes/google_play) for the
  ratings_history stream's required store filter.
- `report_to` (optional, string); format `date`; End date (YYYY-MM-DD) for the reviews and
  reviews_summary streams' required from/to date-range filter.

Secret fields are redacted in logs and write previews: `api_secret`.

Default configuration values: `base_url=https://api.appfollow.io/api/v2`, `report_country=us`,
`report_store=itunes`.

Authentication behavior:

- API key authentication in `X-AppFollow-API-Token` using `secrets.api_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/account/users`.

## Streams notes

Default pagination: single request; no pagination.

- `users`: GET `/account/users` - records at response root.
- `app_collections`: GET `/account/apps` - records path `apps`.
- `app_lists`: GET `/account/apps/app` - records path `apps_app`; fan-out; ids from config field
  `app_collection_ids`; id sent as query parameter `apps_id`; stamps `app_collection_id`.
- `ratings`: GET `/meta/ratings` - records path `ratings.list`; response-level fields copied to
  records `store`; fan-out; ids from config field `ext_ids`; id sent as query parameter `ext_id`;
  stamps `ext_id`.
- `reviews`: GET `/reviews` - records at response root; query `from`=`{{ config.report_from }}`;
  `to`=`{{ config.report_to }}`; fan-out; ids from config field `ext_ids`; id sent as query
  parameter `ext_id`; stamps `ext_id`.
- `reviews_summary`: GET `/reviews/summary` - records at response root; query `from` from template
  `{{ config.report_from }}`, omitted when absent; `to` from template `{{ config.report_to }}`,
  omitted when absent; fan-out; ids from config field `ext_ids`; id sent as query parameter
  `ext_id`; stamps `ext_id`.
- `keywords`: GET `/aso/keywords` - records at response root; fan-out; ids from config field
  `ext_ids`; id sent as query parameter `ext_id`; stamps `ext_id`.
- `rankings`: GET `/meta/rankings` - records at response root; fan-out; ids from config field
  `ext_ids`; id sent as query parameter `ext_id`; stamps `ext_id`.
- `versions`: GET `/meta/versions` - records at response root; query `country`=`{{
  config.report_country }}`; fan-out; ids from config field `ext_ids`; id sent as query parameter
  `ext_id`; stamps `ext_id`.
- `versions_whatsnew`: GET `/meta/versions/whatsnew` - records at response root; fan-out; ids from
  config field `ext_ids`; id sent as query parameter `ext_id`; stamps `ext_id`.
- `ratings_history`: GET `/meta/ratings/history` - records at response root; query `store`=`{{
  config.report_store }}`; fan-out; ids from config field `ext_ids`; id sent as query parameter
  `ext_id`; stamps `ext_id`.

## Write actions & risks

Overall write risk: external mutations: posts public review replies, edits review
tags/notes/custom-status, replaces tracked ASO keyword sets, and adds/updates/removes account users,
app collections, and tracked apps.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `reply_to_review`: POST `/reviews/reply` - kind `create`; body type `json`; required record fields
  `ext_id`, `review_id`, `answer_text`; accepted fields `answer_text`, `ext_id`, `login`,
  `review_id`; risk: external mutation; posts a public reply to a live app-store review, cannot be
  unsent programmatically; approval required.
- `update_review_tags`: POST `/reviews/tags` - kind `update`; body type `json`; required record
  fields `ext_id`, `review_id`, `tags`; accepted fields `apps_id`, `ext_id`, `review_id`, `tags`;
  risk: external mutation; overwrites a review's tag set; approval required.
- `update_review_notes`: POST `/reviews/notes` - kind `update`; body type `json`; required record
  fields `ext_id`, `review_id`, `content`; accepted fields `content`, `ext_id`, `review_id`; risk:
  external mutation; overwrites a review's internal note; approval required.
- `edit_keywords`: POST `/aso/keywords` - kind `update`; body type `json`; required record fields
  `country`, `device`, `keywords`; accepted fields `apps_id`, `country`, `device`, `keywords`; risk:
  external mutation; replaces the tracked ASO keyword list for a country/device pair; approval
  required.
- `add_user`: POST `/account/users` - kind `create`; body type `json`; required record fields
  `name`, `role`, `email`; accepted fields `collections`, `email`, `name`, `role`; risk: external
  mutation; grants AppFollow account access to a new user; approval required.
- `update_user`: PATCH `/account/users` - kind `update`; body type `json`; required record fields
  `id`, `name`, `role`, `email`; accepted fields `collections`, `email`, `id`, `name`, `role`; risk:
  external mutation; changes an existing account user's role/access; approval required.
- `remove_user`: DELETE `/account/users` - kind `delete`; body type `json`; body fields `id`,
  `email`; required record fields `id`; accepted fields `email`, `id`; risk: irreversible external
  mutation; revokes an AppFollow account user's access; approval required.
- `add_collection`: POST `/account/apps` - kind `create`; body type `json`; required record fields
  `title`, `countries`; accepted fields `appUpdates`, `countries`, `dashboard`, `default_country`,
  `email`, `keywords`, `ranks`, `reviews`, `title`; risk: external mutation; creates a new billable
  app collection; approval required.
- `remove_collection`: DELETE `/account/apps` - kind `delete`; body type `json`; body fields
  `apps_id`; required record fields `apps_id`; accepted fields `apps_id`; risk: irreversible
  external deletion; removes an app collection and every app tracked under it; approval required.
- `add_app`: POST `/account/apps/app` - kind `create`; body type `json`; required record fields
  `store`, `ext_id`, `apps_id`, `locale`; accepted fields `apps_id`, `ext_id`, `locale`, `store`,
  `user_id`; risk: external mutation; adds a tracked app to an existing collection; approval
  required.
- `remove_app`: DELETE `/account/apps/app` - kind `delete`; body type `json`; body fields `store`,
  `ext_id`, `apps_id`, `user_id`; required record fields `store`, `ext_id`, `apps_id`; accepted
  fields `apps_id`, `ext_id`, `store`, `user_id`; risk: irreversible external deletion; stops
  tracking an app under a collection; approval required.

## Known limits

- API coverage includes 11 stream-backed endpoint group(s), 11 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, duplicate_of=6, out_of_scope=15.
