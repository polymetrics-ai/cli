# pm connectors inspect tmdb

```text
NAME
  pm connectors inspect tmdb - TMDb connector manual

SYNOPSIS
  pm connectors inspect tmdb
  pm connectors inspect tmdb --json
  pm credentials add <name> --connector tmdb [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads TMDb movie, TV, person, collection, company, keyword, review, account, search, trending, and reference metadata from The Movie Database API.

ICON
  asset: icons/tmdb.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.themoviedb.org/reference/intro/getting-started

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_ids
  base_url
  collection_ids
  company_ids
  credit_ids
  episode_ids
  episode_number
  external_ids
  external_source
  first_air_date_year
  guest_session_ids
  include_adult
  keyword_ids
  language
  list_ids
  movie_id
  network_ids
  person_ids
  primary_release_year
  query
  region
  review_ids
  season_ids
  season_number
  series_id
  series_ids
  sort_by
  tv_episode_group_ids
  watch_region
  with_watch_providers
  year
  api_key (secret)
  guest_session_id (secret)
  session_id (secret)

ETL STREAMS
  popular_movies:
    primary key: id
    fields: id(), overview(), release_date(), title(), vote_average()
  now_playing_movies:
    primary key: id
    fields: id(), overview(), release_date(), title(), vote_average()
  search_movies:
    primary key: id
    fields: id(), overview(), release_date(), title(), vote_average()
  movie_details:
    primary key: id
    fields: id(), overview(), release_date(), runtime(), title()
  account_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  account_get_favorites:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  account_favorite_tv:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  account_lists:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  account_rated_movies:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  account_rated_tv:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  account_rated_tv_episodes:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  account_watchlist_movies:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  account_watchlist_tv:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  certification_movie_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  certifications_tv_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  changes_movie_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  changes_people_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  changes_tv_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  collection_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  collection_images:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  collection_translations:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  company_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  company_alternative_names:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  company_images:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  configuration_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  configuration_countries:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  configuration_jobs:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  configuration_languages:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  configuration_timezones:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  credit_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  discover_movie:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  discover_tv:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  find_by_id:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  genre_movie_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  genre_tv_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  guest_session_rated_movies:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  guest_session_rated_tv:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  guest_session_rated_tv_episodes:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  keyword_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  keyword_movies:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  list_check_item_status:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  list_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_top_rated_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_upcoming_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_account_states:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_alternative_titles:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_changes:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_credits:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_external_ids:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_images:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_keywords:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_latest_id:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_lists:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_recommendations:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_release_dates:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_reviews:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_similar:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_translations:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_videos:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  movie_watch_providers:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  network_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  details_copy:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  alternative_names_copy:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_popular_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_changes:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_combined_credits:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_external_ids:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_images:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_latest_id:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_movie_credits:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_tv_credits:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  person_tagged_images:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  translations:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  review_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  search_collection:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  search_company:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  search_keyword:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  search_multi:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  search_person:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  search_tv:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  trending_all:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  trending_movies:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  trending_people:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  trending_tv:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_airing_today_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_on_the_air_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_popular_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_top_rated_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_account_states:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_aggregate_credits:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_alternative_titles:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_changes:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_content_ratings:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_credits:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_episode_groups:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_external_ids:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_images:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_keywords:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_latest_id:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  lists_copy:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_recommendations:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_reviews:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_screened_theatrically:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_similar:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_translations:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_videos:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_series_watch_providers:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_account_states:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_aggregate_credits:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_changes_by_id:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_credits:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_external_ids:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_images:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_translations:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_videos:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_season_watch_providers:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_episode_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_episode_account_states:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_episode_changes_by_id:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_episode_credits:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_episode_external_ids:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_episode_images:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_episode_translations:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_episode_videos:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  tv_episode_group_details:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  watch_providers_available_regions:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  watch_providers_movie_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()
  watch_provider_tv_list:
    primary key: id
    fields: cast(), crew(), first_air_date(), id(), media_type(), name(), overview(), popularity(), release_date(), results(), title(), vote_average()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external TMDb API read of public catalog, search, account-state, and reference metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect tmdb

  # Inspect as structured JSON
  pm connectors inspect tmdb --json

AGENT WORKFLOW
  - Run pm connectors inspect tmdb before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
