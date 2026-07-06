# pm connectors inspect buzzsprout

```text
NAME
  pm connectors inspect buzzsprout - Buzzsprout connector manual

SYNOPSIS
  pm connectors inspect buzzsprout
  pm connectors inspect buzzsprout --json
  pm credentials add <name> --connector buzzsprout [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Buzzsprout podcasts and episodes (titles, publish dates, durations, play counts) and creates/updates episodes through the Buzzsprout REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  podcast_id
  api_key (secret)

ETL STREAMS
  episodes:
    primary key: id
    cursor: published_at
    fields: artist(), artwork_url(), audio_url(), description(), duration(), episode_number(), explicit(), guid(), hq(), id(), inactive_at(), magic_mastering(), private(), published_at(), season_number(), summary(), tags(), title(), total_plays()
  podcasts:
    primary key: id
    fields: artwork_url(), author(), contact_email(), description(), explicit(), id(), keywords(), language(), main_category(), sub_category(), timezone(), title(), website_address()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_episode:
    endpoint: POST /api/{{ config.podcast_id }}/episodes.json
    risk: external mutation; creates a new episode (and can trigger audio processing/publication) on the configured podcast; approval required
  update_episode:
    endpoint: PUT /api/{{ config.podcast_id }}/episodes/{{ record.id }}.json
    required fields: id
    risk: external mutation; overwrites episode metadata on the configured podcast; approval required

SECURITY
  read risk: external Buzzsprout API read of podcast and episode data
  write risk: external mutation of episode metadata/audio (create_episode, update_episode) on the configured podcast; can trigger audio processing and publish/unpublish an episode
  approval: required for create_episode/update_episode
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect buzzsprout

  # Inspect as structured JSON
  pm connectors inspect buzzsprout --json

AGENT WORKFLOW
  - Run pm connectors inspect buzzsprout before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
