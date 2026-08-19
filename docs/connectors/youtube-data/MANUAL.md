# pm connectors inspect youtube-data

```text
NAME
  pm connectors inspect youtube-data - YouTube Data connector manual

SYNOPSIS
  pm connectors inspect youtube-data
  pm connectors inspect youtube-data --json
  pm credentials add <name> --connector youtube-data [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads channels, videos, playlists, playlist items, comment threads, search results, video categories, and i18n region/language reference data through the YouTube Data API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  channel_ids
  ids
  mode
  playlist_ids
  region_code
  search_query
  video_ids
  api_key (secret) (required)

ETL STREAMS
  channels:
    primary key: id
    fields: id(string), title(string), view_count(string)
  videos:
    primary key: id
    fields: id(string), published_at(string), title(string)
  playlists:
    primary key: id
    fields: id(string), published_at(string), title(string)
  playlist_items:
    primary key: id
    cursor: published_at
    fields: id(string), playlist_id(string), published_at(string), title(string), video_id(string)
  comment_threads:
    primary key: id
    cursor: published_at
    fields: id(string), published_at(string), text(string), video_id(string)
  search:
    primary key: id
    cursor: published_at
    fields: id(string), published_at(string), title(string)
  video_categories:
    primary key: id
    fields: id(string), title(string)
  i18n_regions:
    primary key: id
    fields: id(string), name(string)
  i18n_languages:
    primary key: id
    fields: id(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external YouTube Data API read of public channel, video, playlist, playlist item, comment, search result, and reference data
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect youtube-data

  # Inspect as structured JSON
  pm connectors inspect youtube-data --json

AGENT WORKFLOW
  - Run pm connectors inspect youtube-data before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
