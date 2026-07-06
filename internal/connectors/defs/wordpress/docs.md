# Overview

Reads and writes WordPress REST API content: posts, pages, comments, media, users, categories, tags,
taxonomies, post types, and post statuses.

Readable streams: `posts`, `pages`, `comments`, `media`, `users`, `categories`, `tags`,
`taxonomies`, `types`, `statuses`.

Write actions: `create_post`, `update_post`, `delete_post`, `create_page`, `update_page`,
`delete_page`, `create_comment`, `update_comment`, `delete_comment`, `update_media`, `delete_media`,
`create_user`, `update_user`, `delete_user`, `create_category`, `update_category`,
`delete_category`, `create_tag`, `update_tag`, `delete_tag`.

Service API documentation: https://developer.wordpress.org/rest-api/.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; Your WordPress site's base URL (e.g.
  https://example.com).
- `password` (optional, secret, string); Optional HTTP Basic auth password (or application password)
  for private/authenticated WordPress sites. Only applied when both username and password are set.
  Never logged.
- `start_date` (optional, string); format `date-time`.
- `username` (optional, secret, string); Optional HTTP Basic auth username for private/authenticated
  WordPress sites. Only applied when both username and password are set.

Secret fields are redacted in logs and write previews: `password`, `username`.

Authentication behavior:

- HTTP Basic authentication using `secrets.username`, `secrets.password` when `{{ secrets.password
  }}`.
- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/wp-json/wp/v2/posts` with query `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100; maximum 1 page(s).

Pagination by stream: none: `taxonomies`, `types`, `statuses`; page_number: `posts`, `pages`,
`comments`, `media`, `users`, `categories`, `tags`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `posts`: GET `/wp-json/wp/v2/posts` - records path `.`; query `after` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); incremental cursor `date`;
  formatted as `rfc3339`; emits passthrough records.
- `pages`: GET `/wp-json/wp/v2/pages` - records path `.`; query `after` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); incremental cursor `date`;
  formatted as `rfc3339`; emits passthrough records.
- `comments`: GET `/wp-json/wp/v2/comments` - records path `.`; query `after` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); incremental cursor `date`;
  formatted as `rfc3339`; emits passthrough records.
- `media`: GET `/wp-json/wp/v2/media` - records path `.`; query `after` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); incremental cursor `date`;
  formatted as `rfc3339`; emits passthrough records.
- `users`: GET `/wp-json/wp/v2/users` - records path `.`; query `after` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits passthrough records.
- `categories`: GET `/wp-json/wp/v2/categories` - records path `.`; query `after` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits passthrough records.
- `tags`: GET `/wp-json/wp/v2/tags` - records path `.`; query `after` from template `{{
  config.start_date }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; maximum 1 page(s); emits passthrough records.
- `taxonomies`: GET `/wp-json/wp/v2/taxonomies` - records path `.`; flattens keyed objects; key
  field `slug`; emits passthrough records.
- `types`: GET `/wp-json/wp/v2/types` - records path `.`; flattens keyed objects; key field `slug`;
  emits passthrough records.
- `statuses`: GET `/wp-json/wp/v2/statuses` - records path `.`; flattens keyed objects; key field
  `slug`; emits passthrough records.

## Write actions & risks

Overall write risk: external mutation of public site content and accounts (posts, pages, comments,
media metadata, users, categories, tags); requires authenticated (Basic auth) credentials with
sufficient WordPress capabilities; deletes are irreversible for users/categories/tags/media
(WordPress core requires force=true, no trash) and approval-gated for all actions.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_post`: POST `/wp-json/wp/v2/posts` - kind `create`; body type `json`; accepted fields
  `author`, `categories`, `content`, `excerpt`, `slug`, `status`, `tags`, `title`; risk: external
  mutation; publishes/creates public site content; approval required.
- `update_post`: POST `/wp-json/wp/v2/posts/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `content`, `id`, `status`, `title`;
  risk: external mutation; edits public site content; approval required.
- `delete_post`: DELETE `/wp-json/wp/v2/posts/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: external deletion of public site
  content (moves to trash unless force=true is embedded in the path); approval required.
- `create_page`: POST `/wp-json/wp/v2/pages` - kind `create`; body type `json`; accepted fields
  `content`, `parent`, `slug`, `status`, `title`; risk: external mutation; publishes/creates public
  site content; approval required.
- `update_page`: POST `/wp-json/wp/v2/pages/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `content`, `id`, `status`, `title`;
  risk: external mutation; edits public site content; approval required.
- `delete_page`: DELETE `/wp-json/wp/v2/pages/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: external deletion of public site
  content (moves to trash unless force=true is embedded in the path); approval required.
- `create_comment`: POST `/wp-json/wp/v2/comments` - kind `create`; body type `json`; required
  record fields `post`, `content`; accepted fields `author_email`, `author_name`, `content`,
  `parent`, `post`, `status`; risk: external mutation; publishes a public-facing comment; approval
  required.
- `update_comment`: POST `/wp-json/wp/v2/comments/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `content`, `id`, `status`;
  risk: external mutation; edits/moderates a public-facing comment; approval required.
- `delete_comment`: DELETE `/wp-json/wp/v2/comments/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: external deletion of a
  comment (moves to trash unless force=true is embedded in the path); approval required.
- `update_media`: POST `/wp-json/wp/v2/media/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `alt_text`, `caption`,
  `description`, `id`, `title`; risk: external mutation; edits media-item metadata (title/alt
  text/caption/description); approval required.
- `delete_media`: DELETE `/wp-json/wp/v2/media/{{ record.id }}?force=true` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: irreversible external
  deletion of a media/attachment item (WordPress core requires force=true; attachments do not
  support trashing); approval required.
- `create_user`: POST `/wp-json/wp/v2/users` - kind `create`; body type `json`; required record
  fields `username`, `email`, `password`; accepted fields `email`, `first_name`, `last_name`,
  `name`, `password`, `roles`, `username`; risk: external mutation; creates a new site user account
  with a password; approval required.
- `update_user`: POST `/wp-json/wp/v2/users/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `email`, `id`, `name`, `roles`; risk:
  external mutation; edits a site user account, including role/permission assignment; approval
  required.
- `delete_user`: DELETE `/wp-json/wp/v2/users/{{ record.id }}?force=true&reassign={{ record.reassign
  }}` - kind `delete`; body type `none`; path fields `id`, `reassign`; required record fields `id`,
  `reassign`; accepted fields `id`, `reassign`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: irreversible external deletion of a site user account (WordPress
  core requires force=true and a reassign target; users do not support trashing); approval required.
- `create_category`: POST `/wp-json/wp/v2/categories` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `description`, `name`, `parent`, `slug`; risk: external
  mutation; approval required.
- `update_category`: POST `/wp-json/wp/v2/categories/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `description`, `id`,
  `name`; risk: external mutation; approval required.
- `delete_category`: DELETE `/wp-json/wp/v2/categories/{{ record.id }}?force=true` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: irreversible
  external deletion of a category (WordPress core requires force=true; terms do not support
  trashing); approval required.
- `create_tag`: POST `/wp-json/wp/v2/tags` - kind `create`; body type `json`; required record fields
  `name`; accepted fields `description`, `name`, `slug`; risk: external mutation; approval required.
- `update_tag`: POST `/wp-json/wp/v2/tags/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `description`, `id`, `name`; risk:
  external mutation; approval required.
- `delete_tag`: DELETE `/wp-json/wp/v2/tags/{{ record.id }}?force=true` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: irreversible external
  deletion of a tag (WordPress core requires force=true; terms do not support trashing); approval
  required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 10 stream-backed endpoint group(s), 20 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=9, duplicate_of=16, non_data_endpoint=7, out_of_scope=19,
  requires_elevated_scope=8.
