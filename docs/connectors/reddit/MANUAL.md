# pm connectors inspect reddit

```text
NAME
  pm connectors inspect reddit - Reddit connector manual

SYNOPSIS
  pm connectors inspect reddit
  pm connectors inspect reddit --json
  pm credentials add <name> --connector reddit [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads subreddit posts and comments through the Reddit OAuth API listing endpoints.

ICON
  id: simple-icons-reddit
  asset: icons/simple-icons/reddit.svg
  title: Reddit
  simple_icon_slug: reddit
  simple_icon_hex: FF4500
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Reddit
  match: exact-name-or-slug
  matched_by: reddit

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  subreddit (required)
  access_token (secret) (required)

ETL STREAMS
  posts:
    primary key: id
    cursor: created_utc
    fields: author(string), created_utc(number), id(string), name(string), permalink(string), subreddit(string), title(string)
  comments:
    primary key: id
    cursor: created_utc
    fields: author(string), body(string), created_utc(number), id(string), name(string), permalink(string), subreddit(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Reddit OAuth API read of public subreddit posts and comments
  approval: none; read-only, caller-supplied OAuth token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect reddit

  # Inspect as structured JSON
  pm connectors inspect reddit --json

AGENT WORKFLOW
  - Run pm connectors inspect reddit before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
