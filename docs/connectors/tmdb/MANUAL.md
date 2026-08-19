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
  id: tmdb
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
  api_key (secret) (required)
  guest_session_id (secret)
  session_id (secret)

ETL STREAMS
  popular_movies:
    primary key: id
    fields: id(integer), overview(string), release_date(string), title(string), vote_average(number)
  now_playing_movies:
    primary key: id
    fields: id(integer), overview(string), release_date(string), title(string), vote_average(number)
  search_movies:
    primary key: id
    fields: id(integer), overview(string), release_date(string), title(string), vote_average(number)
  movie_details:
    primary key: id
    fields: id(integer), overview(string), release_date(string), runtime(integer), title(string)
  account_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  account_get_favorites:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  account_favorite_tv:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  account_lists:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  account_rated_movies:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  account_rated_tv:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  account_rated_tv_episodes:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  account_watchlist_movies:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  account_watchlist_tv:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  certification_movie_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  certifications_tv_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  changes_movie_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  changes_people_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  changes_tv_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  collection_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  collection_images:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  collection_translations:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  company_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  company_alternative_names:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  company_images:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  configuration_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  configuration_countries:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  configuration_jobs:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  configuration_languages:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  configuration_timezones:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  credit_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  discover_movie:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  discover_tv:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  find_by_id:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  genre_movie_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  genre_tv_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  guest_session_rated_movies:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  guest_session_rated_tv:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  guest_session_rated_tv_episodes:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  keyword_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  keyword_movies:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  list_check_item_status:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  list_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_top_rated_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_upcoming_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_account_states:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_alternative_titles:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_changes:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_credits:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_external_ids:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_images:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_keywords:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_latest_id:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_lists:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_recommendations:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_release_dates:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_reviews:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_similar:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_translations:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_videos:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  movie_watch_providers:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  network_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  details_copy:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  alternative_names_copy:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_popular_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_changes:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_combined_credits:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_external_ids:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_images:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_latest_id:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_movie_credits:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_tv_credits:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  person_tagged_images:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  translations:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  review_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  search_collection:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  search_company:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  search_keyword:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  search_multi:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  search_person:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  search_tv:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  trending_all:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  trending_movies:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  trending_people:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  trending_tv:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_airing_today_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_on_the_air_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_popular_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_top_rated_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_account_states:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_aggregate_credits:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_alternative_titles:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_changes:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_content_ratings:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_credits:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_episode_groups:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_external_ids:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_images:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_keywords:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_latest_id:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  lists_copy:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_recommendations:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_reviews:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_screened_theatrically:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_similar:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_translations:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_videos:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_series_watch_providers:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_account_states:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_aggregate_credits:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_changes_by_id:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_credits:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_external_ids:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_images:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_translations:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_videos:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_season_watch_providers:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_episode_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_episode_account_states:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_episode_changes_by_id:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_episode_credits:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_episode_external_ids:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_episode_images:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_episode_translations:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_episode_videos:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  tv_episode_group_details:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  watch_providers_available_regions:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  watch_providers_movie_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)
  watch_provider_tv_list:
    primary key: id
    fields: cast(array), crew(array), first_air_date(string), id(integer), media_type(string), name(string), overview(string), popularity(number), release_date(string), results(object), title(string), vote_average(number)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

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
