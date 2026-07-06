# pm connectors inspect watchmode

```text
NAME
  pm connectors inspect watchmode - Watchmode connector manual

SYNOPSIS
  pm connectors inspect watchmode
  pm connectors inspect watchmode --json
  pm credentials add <name> --connector watchmode [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Watchmode title search results, streaming sources, regions, networks, genres, list-titles, releases, per-title details/sources/seasons/episodes/cast-crew, and person details. Read-only.

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
  mode
  page_size
  person_ids
  regions
  search_val
  start_date
  title_ids
  types
  api_key (secret)

ETL STREAMS
  search:
    primary key: id
    fields: id(), name(), type(), year()
  sources:
    primary key: id
    fields: id(), name(), region(), type()
  regions:
    primary key: country
    fields: country(), data_tier(), flag(), name(), plan_enabled()
  networks:
    primary key: id
    fields: id(), name(), origin_country(), tmdb_id()
  genres:
    primary key: id
    fields: id(), name(), tmdb_id()
  titles:
    primary key: id
    fields: id(), imdb_id(), title(), tmdb_id(), tmdb_type(), type(), year()
  releases:
    primary key: id, source_id, source_release_date
    fields: id(), imdb_id(), is_original(), poster_url(), season_number(), source_id(), source_name(), source_release_date(), title(), tmdb_id(), tmdb_type(), type()
  title_details:
    primary key: id
    fields: backdrop(), critic_score(), end_year(), genre_names(), id(), imdb_id(), original_title(), plot_overview(), poster(), release_date(), runtime_minutes(), title(), tmdb_id(), tmdb_type(), type(), us_rating(), user_rating(), watchmode_title_id(), year()
  title_sources:
    primary key: watchmode_title_id, source_id, region, type
    fields: episodes(), format(), name(), price(), region(), seasons(), source_id(), type(), watchmode_title_id(), web_url()
  title_seasons:
    primary key: id
    fields: air_date(), episode_count(), id(), name(), number(), overview(), poster_url(), watchmode_title_id()
  title_episodes:
    primary key: id
    fields: episode_number(), id(), imdb_id(), name(), overview(), release_date(), runtime_minutes(), season_id(), season_number(), sources(), tmdb_id(), watchmode_title_id()
  title_cast_crew:
    primary key: watchmode_title_id, person_id, type, role
    fields: episode_count(), full_name(), order(), person_id(), role(), type(), watchmode_title_id()
  person_details:
    primary key: id
    fields: date_of_birth(), date_of_death(), first_name(), full_name(), gender(), id(), imdb_id(), known_for(), last_name(), main_profession(), place_of_birth(), relevance_percentile(), tmdb_id(), watchmode_person_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Watchmode API read of public title/streaming-source/person media metadata
  approval: none; read-only public media metadata connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect watchmode

  # Inspect as structured JSON
  pm connectors inspect watchmode --json

AGENT WORKFLOW
  - Run pm connectors inspect watchmode before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
