# Overview

Reads Confluence Cloud spaces, pages, blog posts, labels, attachments, comments, tasks, and custom
content, and writes pages, blog posts, and comments through the Confluence Cloud REST API v2.

Readable streams: `spaces`, `pages`, `blogposts`, `labels`, `attachments`, `footer_comments`,
`inline_comments`, `tasks`, `custom_content`.

Write actions: `create_page`, `update_page`, `create_blogpost`, `create_footer_comment`,
`create_inline_comment`.

Service API documentation: https://developer.atlassian.com/cloud/confluence/rest/v2/intro/.

## Auth setup

Connection fields:

- `api_token` (required, secret, string); Atlassian API token, used as the HTTP Basic auth password.
  Never logged.
- `base_url` (required, string); format `uri`; Your Confluence Cloud site origin, without the
  /wiki/api/v2 path (e.g. https://mysite.atlassian.net). The bundle paths include /wiki/api/v2 so
  relative _links.next pagination follows the same host correctly.
- `custom_content_type` (optional, string); default `com.atlassian.confluence.macro.core`;
  Confluence custom-content type key to sync via the custom_content stream. GET /custom-content
  requires a type filter (Confluence has no type-agnostic custom-content list); set this to the
  app-defined type key you want to read (e.g. a Forge/Connect app's registered custom content type).
- `email` (required, string); Atlassian account email used as the HTTP Basic auth username.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_token`.

Default configuration values: `custom_content_type=com.atlassian.confluence.macro.core`.

Authentication behavior:

- HTTP Basic authentication using `config.email`, `secrets.api_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/wiki/api/v2/spaces`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `_links.next`;
cross-host next URLs are allowed.

- `spaces`: GET `/wiki/api/v2/spaces` - records path `results`; query `limit`=`25`; follows a
  next-page URL from the response body; URL path `_links.next`; cross-host next URLs are allowed.
- `pages`: GET `/wiki/api/v2/pages` - records path `results`; query `limit`=`25`; follows a
  next-page URL from the response body; URL path `_links.next`; cross-host next URLs are allowed;
  computed output fields `version`.
- `blogposts`: GET `/wiki/api/v2/blogposts` - records path `results`; query `limit`=`25`; follows a
  next-page URL from the response body; URL path `_links.next`; cross-host next URLs are allowed;
  computed output fields `version`.
- `labels`: GET `/wiki/api/v2/labels` - records path `results`; query `limit`=`25`; follows a
  next-page URL from the response body; URL path `_links.next`; cross-host next URLs are allowed.
- `attachments`: GET `/wiki/api/v2/attachments` - records path `results`; query `limit`=`25`;
  follows a next-page URL from the response body; URL path `_links.next`; cross-host next URLs are
  allowed.
- `footer_comments`: GET `/wiki/api/v2/footer-comments` - records path `results`; query
  `body-format`=`storage`; `limit`=`25`; follows a next-page URL from the response body; URL path
  `_links.next`; cross-host next URLs are allowed; computed output fields `version`.
- `inline_comments`: GET `/wiki/api/v2/inline-comments` - records path `results`; query
  `body-format`=`storage`; `limit`=`25`; follows a next-page URL from the response body; URL path
  `_links.next`; cross-host next URLs are allowed; computed output fields `version`.
- `tasks`: GET `/wiki/api/v2/tasks` - records path `results`; query `body-format`=`storage`;
  `limit`=`25`; follows a next-page URL from the response body; URL path `_links.next`; cross-host
  next URLs are allowed.
- `custom_content`: GET `/wiki/api/v2/custom-content` - records path `results`; query `limit`=`25`;
  `type`=`{{ config.custom_content_type }}`; follows a next-page URL from the response body; URL
  path `_links.next`; cross-host next URLs are allowed; computed output fields `version`.

## Write actions & risks

Overall write risk: external mutation: creates/updates Confluence pages, blog posts, and comments;
no destructive (delete) actions are exposed.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_page`: POST `/wiki/api/v2/pages` - kind `create`; body type `json`; required record fields
  `spaceId`, `title`, `body`; accepted fields `body`, `parentId`, `spaceId`, `status`, `title`;
  risk: creates a new published or draft page in the target space; external mutation, no approval
  required.
- `update_page`: PUT `/wiki/api/v2/pages/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `status`, `title`, `spaceId`, `version`; accepted fields
  `body`, `id`, `spaceId`, `status`, `title`, `version`; risk: mutates an existing page's
  content/status; requires the caller to supply the next version.number (Confluence rejects a stale
  version number), external mutation, no approval required.
- `create_blogpost`: POST `/wiki/api/v2/blogposts` - kind `create`; body type `json`; required
  record fields `spaceId`, `title`, `body`; accepted fields `body`, `spaceId`, `status`, `title`;
  risk: creates a new published or draft blog post in the target space; external mutation, no
  approval required.
- `create_footer_comment`: POST `/wiki/api/v2/footer-comments` - kind `create`; body type `json`;
  required record fields `pageId`, `body`; accepted fields `blogPostId`, `body`, `pageId`,
  `parentCommentId`; risk: creates a new footer comment (or reply) on a page/blogpost; external
  mutation, no approval required.
- `create_inline_comment`: POST `/wiki/api/v2/inline-comments` - kind `create`; body type `json`;
  required record fields `pageId`, `body`, `inlineCommentProperties`; accepted fields `blogPostId`,
  `body`, `inlineCommentProperties`, `pageId`, `parentCommentId`; risk: creates a new inline comment
  (or reply) anchored to a text selection on a page/blogpost; external mutation, no approval
  required.

## Known limits

- Batch defaults: read_page_size=25.
- API coverage includes 9 stream-backed endpoint group(s), 5 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=10, duplicate_of=30, non_data_endpoint=10, out_of_scope=31,
  requires_elevated_scope=10.
