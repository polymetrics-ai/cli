# Overview

Reads Blogger (Google Blogger API v3) blogs, posts, pages, comments, and page-view counts using an
OAuth 2.0 refresh-token grant. Read-only.

Readable streams: `blogs`, `posts`, `pages`, `comments`, `pageviews`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/blogger/docs/3.0/reference.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://www.googleapis.com/blogger/v3`; format `uri`;
  Blogger API base URL override for tests or proxies.
- `blog_id` (required, string); The Blogger blog id to read (blogs/{blog_id}/... resources).
- `client_id` (required, secret, string); Google OAuth 2.0 client ID for the refresh-token grant.
  Used only in the token-request form; never logged.
- `client_refresh_token` (required, secret, string); Long-lived Google OAuth 2.0 refresh token.
  Exchanged for a short-lived access token at token_url; never logged. The 3-legged
  consent/acquisition dance is out of scope for this connector (credentials layer already owns it).
- `client_secret` (required, secret, string); Google OAuth 2.0 client secret. Used only in the
  token-request form; never logged.
- `page_size` (optional, string); default `100`; Records per page (1-500, maxResults) for paginated
  streams (posts, pages, comments).
- `token_url` (optional, string); default `https://oauth2.googleapis.com/token`; format `uri`;
  Google OAuth 2.0 token endpoint override. MUST be https in production; the hook fails closed on a
  non-https or unparseable value to prevent exfiltrating the refresh token to an attacker-chosen
  endpoint.

Secret fields are redacted in logs and write previews: `client_id`, `client_refresh_token`,
`client_secret`.

Default configuration values: `base_url=https://www.googleapis.com/blogger/v3`, `page_size=100`,
`token_url=https://oauth2.googleapis.com/token`.

Authentication behavior:

- Connector-specific authentication using `secrets.client_refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/blogs/{{ config.blog_id }}`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `pageToken`; next token from
`nextPageToken`; page size 100.

Pagination by stream: cursor: `posts`, `pages`, `comments`; none: `blogs`, `pageviews`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `blogs`: GET `/blogs/{{ config.blog_id }}` - records at response root; incremental cursor
  `updated`; formatted as `rfc3339`; computed output fields `pages_total`, `posts_total`.
- `posts`: GET `/blogs/{{ config.blog_id }}/posts` - records path `items`; query `maxResults`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; page size 100; incremental cursor `updated`; formatted as `rfc3339`; computed
  output fields `author_display_name`, `author_id`, `blog_id`, `replies_total`.
- `pages`: GET `/blogs/{{ config.blog_id }}/pages` - records path `items`; query `maxResults`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; page size 100; incremental cursor `updated`; formatted as `rfc3339`; computed
  output fields `author_display_name`, `author_id`, `blog_id`.
- `comments`: GET `/blogs/{{ config.blog_id }}/comments` - records path `items`; query
  `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next token
  from `nextPageToken`; page size 100; incremental cursor `updated`; formatted as `rfc3339`;
  computed output fields `author_display_name`, `author_id`, `blog_id`, `post_id`.
- `pageviews`: GET `/blogs/{{ config.blog_id }}/pageviews` - records path `counts`; computed output
  fields `blog_id`, `time_range`.

## Write actions & risks

This connector is read-only. Read behavior: external Blogger API read of blog/post/page/comment
metadata and page-view counts.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=2, out_of_scope=1, requires_elevated_scope=3.
