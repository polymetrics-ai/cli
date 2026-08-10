# pm connectors inspect instagram

```text
NAME
  pm connectors inspect instagram - Instagram connector manual

SYNOPSIS
  pm connectors inspect instagram
  pm connectors inspect instagram --json
  pm credentials add <name> --connector instagram [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Instagram Business/Creator account profile, media, and stories through the Facebook Graph API.

ICON
  id: instagram
  asset: icons/instagram.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.facebook.com/docs/instagram-platform/changelog

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  ig_user_id (required)
  page_size
  access_token (secret) (required)

ETL STREAMS
  users:
    primary key: id
    fields: biography(string), followers_count(integer), follows_count(integer), id(string), media_count(integer), name(string), profile_picture_url(string), username(string), website(string)
  media:
    primary key: id
    cursor: timestamp
    fields: caption(string), comments_count(integer), id(string), like_count(integer), media_product_type(string), media_type(string), media_url(string), permalink(string), thumbnail_url(string), timestamp(string), username(string)
  stories:
    primary key: id
    fields: caption(string), id(string), media_product_type(string), media_type(string), media_url(string), permalink(string), thumbnail_url(string), timestamp(string), username(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Facebook Graph API read of Instagram Business/Creator account data
  approval: none; read-only source
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect instagram

  # Inspect as structured JSON
  pm connectors inspect instagram --json

AGENT WORKFLOW
  - Run pm connectors inspect instagram before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
