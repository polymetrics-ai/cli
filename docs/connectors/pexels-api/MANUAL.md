# pm connectors inspect pexels-api

```text
NAME
  pm connectors inspect pexels-api - Pexels API connector manual

SYNOPSIS
  pm connectors inspect pexels-api
  pm connectors inspect pexels-api --json
  pm credentials add <name> --connector pexels-api [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Pexels photo/video search and curated/popular results plus featured and personal collections and their media through the Pexels REST API.

ICON
  asset: icons/pexels.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.pexels.com/api/documentation/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  collection_media_sort
  collection_media_type
  color
  locale
  orientation
  query
  size
  api_key (secret)

ETL STREAMS
  photos:
    primary key: id
    fields: alt(), id(), photographer(), photographer_url(), src(), url()
  curated_photos:
    primary key: id
    fields: alt(), id(), photographer(), photographer_url(), src(), url()
  videos:
    primary key: id
    fields: duration(), id(), image(), url(), user()
  popular_videos:
    primary key: id
    fields: duration(), id(), image(), url(), user()
  featured_collections:
    primary key: id
    fields: description(), id(), media_count(), photos_count(), private(), title(), videos_count()
  my_collections:
    primary key: id
    fields: description(), id(), media_count(), photos_count(), private(), title(), videos_count()
  collection_media:
    primary key: id
    fields: alt(), collection_id(), duration(), height(), id(), image(), photographer(), photographer_url(), src(), type(), url(), user(), video_files(), video_pictures(), width()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Pexels API read of photo/video search, curated/popular results, and collection metadata/media; all publicly-licensed stock media, no PII
  approval: none; read-only, no writes (the Pexels API has no create/update/delete endpoint anywhere in its documented surface, per its own docs: "Collections cannot be created or modified using the Pexels API")
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pexels-api

  # Inspect as structured JSON
  pm connectors inspect pexels-api --json

AGENT WORKFLOW
  - Run pm connectors inspect pexels-api before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
