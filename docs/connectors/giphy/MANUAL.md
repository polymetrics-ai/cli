# pm connectors inspect giphy

```text
NAME
  pm connectors inspect giphy - Giphy connector manual

SYNOPSIS
  pm connectors inspect giphy
  pm connectors inspect giphy --json
  pm credentials add <name> --connector giphy [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads GIFs, stickers, and clips from the Giphy search and trending REST endpoints. Read-only.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  query_for_clips
  query_for_gif
  query_for_stickers
  rating
  api_key (secret)

ETL STREAMS
  gif_search:
    primary key: id
    fields: bitly_url(), content_url(), embed_url(), id(), import_datetime(), rating(), slug(), source(), source_tld(), title(), trending_datetime(), type(), url(), username()
  sticker_search:
    primary key: id
    fields: bitly_url(), content_url(), embed_url(), id(), import_datetime(), rating(), slug(), source(), source_tld(), title(), trending_datetime(), type(), url(), username()
  clip_search:
    primary key: id
    fields: bitly_url(), content_url(), embed_url(), id(), import_datetime(), rating(), slug(), source(), source_tld(), title(), trending_datetime(), type(), url(), username()
  trending_gifs:
    primary key: id
    fields: bitly_url(), content_url(), embed_url(), id(), import_datetime(), rating(), slug(), source(), source_tld(), title(), trending_datetime(), type(), url(), username()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Giphy API read of public media search/trending results
  approval: none; read-only public media source connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect giphy

  # Inspect as structured JSON
  pm connectors inspect giphy --json

AGENT WORKFLOW
  - Run pm connectors inspect giphy before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
