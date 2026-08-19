# pm connectors inspect blogger

```text
NAME
  pm connectors inspect blogger - Blogger connector manual

SYNOPSIS
  pm connectors inspect blogger
  pm connectors inspect blogger --json
  pm credentials add <name> --connector blogger [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Blogger (Google Blogger API v3) blogs, posts, pages, comments, and page-view counts using an OAuth 2.0 refresh-token grant. Read-only.

ICON
  id: simple-icons-blogger
  asset: icons/simple-icons/blogger.svg
  title: Blogger
  simple_icon_slug: blogger
  simple_icon_hex: FF5722
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Blogger
  match: exact-name-or-slug
  matched_by: blogger

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  blog_id (required)
  page_size
  token_url
  client_id (secret) (required)
  client_refresh_token (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  blogs:
    primary key: id
    cursor: updated
    fields: description(string), id(string), kind(string), name(string), pages_total(integer), posts_total(integer), published(string), updated(string), url(string)
  posts:
    primary key: id
    cursor: updated
    fields: author_display_name(string), author_id(string), blog_id(string), content(string), id(string), kind(string), published(string), replies_total(integer), status(string), title(string), updated(string), url(string)
  pages:
    primary key: id
    cursor: updated
    fields: author_display_name(string), author_id(string), blog_id(string), content(string), id(string), kind(string), published(string), status(string), title(string), updated(string), url(string)
  comments:
    primary key: id
    cursor: updated
    fields: author_display_name(string), author_id(string), blog_id(string), content(string), id(string), kind(string), post_id(string), published(string), status(string), updated(string)
  pageviews:
    primary key: blog_id, time_range
    fields: blog_id(string), count(string), time_range(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Blogger API read of blog/post/page/comment metadata and page-view counts
  approval: none; read-only, no reverse-ETL write surface
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect blogger

  # Inspect as structured JSON
  pm connectors inspect blogger --json

AGENT WORKFLOW
  - Run pm connectors inspect blogger before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
