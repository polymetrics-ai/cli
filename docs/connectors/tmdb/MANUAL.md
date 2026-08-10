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
  api_key (secret)
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
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external TMDb API read of public catalog, search, account-state, and reference metadata
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run TMDb's declared streams and reverse-ETL actions.
  Usage: pm tmdb <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    account details list - Run the account details ETL stream [intent=etl availability=implemented stream=account_details]
    account favorite tv list - Run the account favorite tv ETL stream [intent=etl availability=implemented stream=account_favorite_tv]
    account get favorites list - Run the account get favorites ETL stream [intent=etl availability=implemented stream=account_get_favorites]
    account lists list - Run the account lists ETL stream [intent=etl availability=implemented stream=account_lists]
    account rated movies list - Run the account rated movies ETL stream [intent=etl availability=implemented stream=account_rated_movies]
    account rated tv episodes list - Run the account rated tv episodes ETL stream [intent=etl availability=implemented stream=account_rated_tv_episodes]
    account rated tv list - Run the account rated tv ETL stream [intent=etl availability=implemented stream=account_rated_tv]
    account watchlist movies list - Run the account watchlist movies ETL stream [intent=etl availability=implemented stream=account_watchlist_movies]
    account watchlist tv list - Run the account watchlist tv ETL stream [intent=etl availability=implemented stream=account_watchlist_tv]
    alternative names copy list - Run the alternative names copy ETL stream [intent=etl availability=implemented stream=alternative_names_copy]
    api delete 3 authentication session - Documented DELETE /3/authentication/session (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.delete.3-authentication-session]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete 3 list list-id - Documented DELETE /3/list/{list_id} (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.delete.3-list-list-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete 3 movie movie-id rating - Documented DELETE /3/movie/{movie_id}/rating (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.delete.3-movie-movie-id-rating]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete 3 tv series-id rating - Documented DELETE /3/tv/{series_id}/rating (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.delete.3-tv-series-id-rating]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete 3 tv series-id season season-number episode episode-number rating - Documented DELETE /3/tv/{series_id}/season/{season_number}/episode/{episode_number}/rating (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.delete.3-tv-series-id-season-season-number-episode-episode-number-rating]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get 3 authentication - Documented GET /3/authentication (not implemented) [intent=direct_read availability=not_implemented operation=tmdb.get.3-authentication]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 3 authentication guest-session new - Documented GET /3/authentication/guest_session/new (not implemented) [intent=direct_read availability=not_implemented operation=tmdb.get.3-authentication-guest-session-new]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 3 authentication token new - Documented GET /3/authentication/token/new (not implemented) [intent=direct_read availability=not_implemented operation=tmdb.get.3-authentication-token-new]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get 3 configuration primary-translations - Documented GET /3/configuration/primary_translations (not implemented) [intent=direct_read availability=not_implemented operation=tmdb.get.3-configuration-primary-translations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post 3 account account-id favorite - Documented POST /3/account/{account_id}/favorite (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-account-account-id-favorite]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 account account-id watchlist - Documented POST /3/account/{account_id}/watchlist (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-account-account-id-watchlist]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 authentication session convert 4 - Documented POST /3/authentication/session/convert/4 (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-authentication-session-convert-4]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 authentication session new - Documented POST /3/authentication/session/new (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-authentication-session-new]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 authentication token validate-with-login - Documented POST /3/authentication/token/validate_with_login (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-authentication-token-validate-with-login]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 list - Documented POST /3/list (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-list]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 list list-id add-item - Documented POST /3/list/{list_id}/add_item (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-list-list-id-add-item]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 list list-id clear - Documented POST /3/list/{list_id}/clear (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-list-list-id-clear]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 list list-id remove-item - Documented POST /3/list/{list_id}/remove_item (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-list-list-id-remove-item]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 movie movie-id rating - Documented POST /3/movie/{movie_id}/rating (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-movie-movie-id-rating]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 tv series-id rating - Documented POST /3/tv/{series_id}/rating (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-tv-series-id-rating]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post 3 tv series-id season season-number episode episode-number rating - Documented POST /3/tv/{series_id}/season/{season_number}/episode/{episode_number}/rating (not implemented) [intent=direct_write availability=not_implemented operation=tmdb.post.3-tv-series-id-season-season-number-episode-episode-number-rating]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    certification movie list list - Run the certification movie list ETL stream [intent=etl availability=implemented stream=certification_movie_list]
    certifications tv list list - Run the certifications tv list ETL stream [intent=etl availability=implemented stream=certifications_tv_list]
    changes movie list list - Run the changes movie list ETL stream [intent=etl availability=implemented stream=changes_movie_list]
    changes people list list - Run the changes people list ETL stream [intent=etl availability=implemented stream=changes_people_list]
    changes tv list list - Run the changes tv list ETL stream [intent=etl availability=implemented stream=changes_tv_list]
    collection details list - Run the collection details ETL stream [intent=etl availability=implemented stream=collection_details]
    collection images list - Run the collection images ETL stream [intent=etl availability=implemented stream=collection_images]
    collection translations list - Run the collection translations ETL stream [intent=etl availability=implemented stream=collection_translations]
    company alternative names list - Run the company alternative names ETL stream [intent=etl availability=implemented stream=company_alternative_names]
    company details list - Run the company details ETL stream [intent=etl availability=implemented stream=company_details]
    company images list - Run the company images ETL stream [intent=etl availability=implemented stream=company_images]
    configuration countries list - Run the configuration countries ETL stream [intent=etl availability=implemented stream=configuration_countries]
    configuration details list - Run the configuration details ETL stream [intent=etl availability=implemented stream=configuration_details]
    configuration jobs list - Run the configuration jobs ETL stream [intent=etl availability=implemented stream=configuration_jobs]
    configuration languages list - Run the configuration languages ETL stream [intent=etl availability=implemented stream=configuration_languages]
    configuration timezones list - Run the configuration timezones ETL stream [intent=etl availability=implemented stream=configuration_timezones]
    credit details list - Run the credit details ETL stream [intent=etl availability=implemented stream=credit_details]
    details copy list - Run the details copy ETL stream [intent=etl availability=implemented stream=details_copy]
    discover movie list - Run the discover movie ETL stream [intent=etl availability=implemented stream=discover_movie]
    discover tv list - Run the discover tv ETL stream [intent=etl availability=implemented stream=discover_tv]
    find by id list - Run the find by id ETL stream [intent=etl availability=implemented stream=find_by_id]
    genre movie list list - Run the genre movie list ETL stream [intent=etl availability=implemented stream=genre_movie_list]
    genre tv list list - Run the genre tv list ETL stream [intent=etl availability=implemented stream=genre_tv_list]
    guest session rated movies list - Run the guest session rated movies ETL stream [intent=etl availability=implemented stream=guest_session_rated_movies]
    guest session rated tv episodes list - Run the guest session rated tv episodes ETL stream [intent=etl availability=implemented stream=guest_session_rated_tv_episodes]
    guest session rated tv list - Run the guest session rated tv ETL stream [intent=etl availability=implemented stream=guest_session_rated_tv]
    keyword details list - Run the keyword details ETL stream [intent=etl availability=implemented stream=keyword_details]
    keyword movies list - Run the keyword movies ETL stream [intent=etl availability=implemented stream=keyword_movies]
    list check item status list - Run the list check item status ETL stream [intent=etl availability=implemented stream=list_check_item_status]
    list details list - Run the list details ETL stream [intent=etl availability=implemented stream=list_details]
    lists copy list - Run the lists copy ETL stream [intent=etl availability=implemented stream=lists_copy]
    movie account states list - Run the movie account states ETL stream [intent=etl availability=implemented stream=movie_account_states]
    movie alternative titles list - Run the movie alternative titles ETL stream [intent=etl availability=implemented stream=movie_alternative_titles]
    movie changes list - Run the movie changes ETL stream [intent=etl availability=implemented stream=movie_changes]
    movie credits list - Run the movie credits ETL stream [intent=etl availability=implemented stream=movie_credits]
    movie details list - Run the movie details ETL stream [intent=etl availability=implemented stream=movie_details]
    movie external ids list - Run the movie external ids ETL stream [intent=etl availability=implemented stream=movie_external_ids]
    movie images list - Run the movie images ETL stream [intent=etl availability=implemented stream=movie_images]
    movie keywords list - Run the movie keywords ETL stream [intent=etl availability=implemented stream=movie_keywords]
    movie latest id list - Run the movie latest id ETL stream [intent=etl availability=implemented stream=movie_latest_id]
    movie lists list - Run the movie lists ETL stream [intent=etl availability=implemented stream=movie_lists]
    movie recommendations list - Run the movie recommendations ETL stream [intent=etl availability=implemented stream=movie_recommendations]
    movie release dates list - Run the movie release dates ETL stream [intent=etl availability=implemented stream=movie_release_dates]
    movie reviews list - Run the movie reviews ETL stream [intent=etl availability=implemented stream=movie_reviews]
    movie similar list - Run the movie similar ETL stream [intent=etl availability=implemented stream=movie_similar]
    movie top rated list list - Run the movie top rated list ETL stream [intent=etl availability=implemented stream=movie_top_rated_list]
    movie translations list - Run the movie translations ETL stream [intent=etl availability=implemented stream=movie_translations]
    movie upcoming list list - Run the movie upcoming list ETL stream [intent=etl availability=implemented stream=movie_upcoming_list]
    movie videos list - Run the movie videos ETL stream [intent=etl availability=implemented stream=movie_videos]
    movie watch providers list - Run the movie watch providers ETL stream [intent=etl availability=implemented stream=movie_watch_providers]
    network details list - Run the network details ETL stream [intent=etl availability=implemented stream=network_details]
    now playing movies list - Run the now playing movies ETL stream [intent=etl availability=implemented stream=now_playing_movies]
    person changes list - Run the person changes ETL stream [intent=etl availability=implemented stream=person_changes]
    person combined credits list - Run the person combined credits ETL stream [intent=etl availability=implemented stream=person_combined_credits]
    person details list - Run the person details ETL stream [intent=etl availability=implemented stream=person_details]
    person external ids list - Run the person external ids ETL stream [intent=etl availability=implemented stream=person_external_ids]
    person images list - Run the person images ETL stream [intent=etl availability=implemented stream=person_images]
    person latest id list - Run the person latest id ETL stream [intent=etl availability=implemented stream=person_latest_id]
    person movie credits list - Run the person movie credits ETL stream [intent=etl availability=implemented stream=person_movie_credits]
    person popular list list - Run the person popular list ETL stream [intent=etl availability=implemented stream=person_popular_list]
    person tagged images list - Run the person tagged images ETL stream [intent=etl availability=implemented stream=person_tagged_images]
    person tv credits list - Run the person tv credits ETL stream [intent=etl availability=implemented stream=person_tv_credits]
    popular movies list - Run the popular movies ETL stream [intent=etl availability=implemented stream=popular_movies]
    review details list - Run the review details ETL stream [intent=etl availability=implemented stream=review_details]
    search collection list - Run the search collection ETL stream [intent=etl availability=implemented stream=search_collection]
    search company list - Run the search company ETL stream [intent=etl availability=implemented stream=search_company]
    search keyword list - Run the search keyword ETL stream [intent=etl availability=implemented stream=search_keyword]
    search movies list - Run the search movies ETL stream [intent=etl availability=implemented stream=search_movies]
    search multi list - Run the search multi ETL stream [intent=etl availability=implemented stream=search_multi]
    search person list - Run the search person ETL stream [intent=etl availability=implemented stream=search_person]
    search tv list - Run the search tv ETL stream [intent=etl availability=implemented stream=search_tv]
    translations list - Run the translations ETL stream [intent=etl availability=implemented stream=translations]
    trending all list - Run the trending all ETL stream [intent=etl availability=implemented stream=trending_all]
    trending movies list - Run the trending movies ETL stream [intent=etl availability=implemented stream=trending_movies]
    trending people list - Run the trending people ETL stream [intent=etl availability=implemented stream=trending_people]
    trending tv list - Run the trending tv ETL stream [intent=etl availability=implemented stream=trending_tv]
    tv episode account states list - Run the tv episode account states ETL stream [intent=etl availability=implemented stream=tv_episode_account_states]
    tv episode changes by id list - Run the tv episode changes by id ETL stream [intent=etl availability=implemented stream=tv_episode_changes_by_id]
    tv episode credits list - Run the tv episode credits ETL stream [intent=etl availability=implemented stream=tv_episode_credits]
    tv episode details list - Run the tv episode details ETL stream [intent=etl availability=implemented stream=tv_episode_details]
    tv episode external ids list - Run the tv episode external ids ETL stream [intent=etl availability=implemented stream=tv_episode_external_ids]
    tv episode group details list - Run the tv episode group details ETL stream [intent=etl availability=implemented stream=tv_episode_group_details]
    tv episode images list - Run the tv episode images ETL stream [intent=etl availability=implemented stream=tv_episode_images]
    tv episode translations list - Run the tv episode translations ETL stream [intent=etl availability=implemented stream=tv_episode_translations]
    tv episode videos list - Run the tv episode videos ETL stream [intent=etl availability=implemented stream=tv_episode_videos]
    tv season account states list - Run the tv season account states ETL stream [intent=etl availability=implemented stream=tv_season_account_states]
    tv season aggregate credits list - Run the tv season aggregate credits ETL stream [intent=etl availability=implemented stream=tv_season_aggregate_credits]
    tv season changes by id list - Run the tv season changes by id ETL stream [intent=etl availability=implemented stream=tv_season_changes_by_id]
    tv season credits list - Run the tv season credits ETL stream [intent=etl availability=implemented stream=tv_season_credits]
    tv season details list - Run the tv season details ETL stream [intent=etl availability=implemented stream=tv_season_details]
    tv season external ids list - Run the tv season external ids ETL stream [intent=etl availability=implemented stream=tv_season_external_ids]
    tv season images list - Run the tv season images ETL stream [intent=etl availability=implemented stream=tv_season_images]
    tv season translations list - Run the tv season translations ETL stream [intent=etl availability=implemented stream=tv_season_translations]
    tv season videos list - Run the tv season videos ETL stream [intent=etl availability=implemented stream=tv_season_videos]
    tv season watch providers list - Run the tv season watch providers ETL stream [intent=etl availability=implemented stream=tv_season_watch_providers]
    tv series account states list - Run the tv series account states ETL stream [intent=etl availability=implemented stream=tv_series_account_states]
    tv series aggregate credits list - Run the tv series aggregate credits ETL stream [intent=etl availability=implemented stream=tv_series_aggregate_credits]
    tv series airing today list list - Run the tv series airing today list ETL stream [intent=etl availability=implemented stream=tv_series_airing_today_list]
    tv series alternative titles list - Run the tv series alternative titles ETL stream [intent=etl availability=implemented stream=tv_series_alternative_titles]
    tv series changes list - Run the tv series changes ETL stream [intent=etl availability=implemented stream=tv_series_changes]
    tv series content ratings list - Run the tv series content ratings ETL stream [intent=etl availability=implemented stream=tv_series_content_ratings]
    tv series credits list - Run the tv series credits ETL stream [intent=etl availability=implemented stream=tv_series_credits]
    tv series details list - Run the tv series details ETL stream [intent=etl availability=implemented stream=tv_series_details]
    tv series episode groups list - Run the tv series episode groups ETL stream [intent=etl availability=implemented stream=tv_series_episode_groups]
    tv series external ids list - Run the tv series external ids ETL stream [intent=etl availability=implemented stream=tv_series_external_ids]
    tv series images list - Run the tv series images ETL stream [intent=etl availability=implemented stream=tv_series_images]
    tv series keywords list - Run the tv series keywords ETL stream [intent=etl availability=implemented stream=tv_series_keywords]
    tv series latest id list - Run the tv series latest id ETL stream [intent=etl availability=implemented stream=tv_series_latest_id]
    tv series on the air list list - Run the tv series on the air list ETL stream [intent=etl availability=implemented stream=tv_series_on_the_air_list]
    tv series popular list list - Run the tv series popular list ETL stream [intent=etl availability=implemented stream=tv_series_popular_list]
    tv series recommendations list - Run the tv series recommendations ETL stream [intent=etl availability=implemented stream=tv_series_recommendations]
    tv series reviews list - Run the tv series reviews ETL stream [intent=etl availability=implemented stream=tv_series_reviews]
    tv series screened theatrically list - Run the tv series screened theatrically ETL stream [intent=etl availability=implemented stream=tv_series_screened_theatrically]
    tv series similar list - Run the tv series similar ETL stream [intent=etl availability=implemented stream=tv_series_similar]
    tv series top rated list list - Run the tv series top rated list ETL stream [intent=etl availability=implemented stream=tv_series_top_rated_list]
    tv series translations list - Run the tv series translations ETL stream [intent=etl availability=implemented stream=tv_series_translations]
    tv series videos list - Run the tv series videos ETL stream [intent=etl availability=implemented stream=tv_series_videos]
    tv series watch providers list - Run the tv series watch providers ETL stream [intent=etl availability=implemented stream=tv_series_watch_providers]
    watch provider tv list list - Run the watch provider tv list ETL stream [intent=etl availability=implemented stream=watch_provider_tv_list]
    watch providers available regions list - Run the watch providers available regions ETL stream [intent=etl availability=implemented stream=watch_providers_available_regions]
    watch providers movie list list - Run the watch providers movie list ETL stream [intent=etl availability=implemented stream=watch_providers_movie_list]

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
