# Overview

Reads TMDb movie, TV, person, collection, company, keyword, review, account, search, trending, and
reference metadata from The Movie Database API.

Readable streams: `popular_movies`, `now_playing_movies`, `search_movies`, `movie_details`,
`account_details`, `account_get_favorites`, `account_favorite_tv`, `account_lists`,
`account_rated_movies`, `account_rated_tv`, `account_rated_tv_episodes`, `account_watchlist_movies`,
`account_watchlist_tv`, `certification_movie_list`, `certifications_tv_list`, `changes_movie_list`,
`changes_people_list`, `changes_tv_list`, `collection_details`, `collection_images`,
`collection_translations`, `company_details`, `company_alternative_names`, `company_images`,
`configuration_details`, `configuration_countries`, `configuration_jobs`, `configuration_languages`,
`configuration_timezones`, `credit_details`, `discover_movie`, `discover_tv`, `find_by_id`,
`genre_movie_list`, `genre_tv_list`, `guest_session_rated_movies`, `guest_session_rated_tv`,
`guest_session_rated_tv_episodes`, `keyword_details`, `keyword_movies`, `list_check_item_status`,
`list_details`, `movie_top_rated_list`, `movie_upcoming_list`, `movie_account_states`,
`movie_alternative_titles`, `movie_changes`, `movie_credits`, `movie_external_ids`, `movie_images`,
`movie_keywords`, `movie_latest_id`, `movie_lists`, `movie_recommendations`, `movie_release_dates`,
`movie_reviews`, `movie_similar`, `movie_translations`, `movie_videos`, `movie_watch_providers`,
`network_details`, `details_copy`, `alternative_names_copy`, `person_popular_list`,
`person_details`, `person_changes`, `person_combined_credits`, `person_external_ids`,
`person_images`, `person_latest_id`, `person_movie_credits`, `person_tv_credits`,
`person_tagged_images`, `translations`, `review_details`, `search_collection`, `search_company`,
`search_keyword`, `search_multi`, `search_person`, `search_tv`, `trending_all`, `trending_movies`,
`trending_people`, `trending_tv`, `tv_series_airing_today_list`, `tv_series_on_the_air_list`,
`tv_series_popular_list`, `tv_series_top_rated_list`, `tv_series_details`,
`tv_series_account_states`, `tv_series_aggregate_credits`, `tv_series_alternative_titles`,
`tv_series_changes`, `tv_series_content_ratings`, `tv_series_credits`, `tv_series_episode_groups`,
`tv_series_external_ids`, `tv_series_images`, `tv_series_keywords`, `tv_series_latest_id`,
`lists_copy`, `tv_series_recommendations`, `tv_series_reviews`, `tv_series_screened_theatrically`,
`tv_series_similar`, `tv_series_translations`, `tv_series_videos`, `tv_series_watch_providers`,
`tv_season_details`, `tv_season_account_states`, `tv_season_aggregate_credits`,
`tv_season_changes_by_id`, `tv_season_credits`, `tv_season_external_ids`, `tv_season_images`,
`tv_season_translations`, `tv_season_videos`, `tv_season_watch_providers`, `tv_episode_details`,
`tv_episode_account_states`, `tv_episode_changes_by_id`, `tv_episode_credits`,
`tv_episode_external_ids`, `tv_episode_images`, `tv_episode_translations`, `tv_episode_videos`,
`tv_episode_group_details`, `watch_providers_available_regions`, `watch_providers_movie_list`,
`watch_provider_tv_list`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.themoviedb.org/reference/intro/getting-started.

## Auth setup

Connection fields:

- `account_ids` (optional, string); Comma-separated IDs for account_watchlist_tv fan-out.
- `api_key` (required, secret, string); TMDb API key (v3 auth), sent as the api_key query parameter.
  Never logged.
- `base_url` (optional, string); default `https://api.themoviedb.org/3`; format `uri`; TMDb API base
  URL override for tests or proxies.
- `collection_ids` (optional, string); Comma-separated IDs for collection_translations fan-out.
- `company_ids` (optional, string); Comma-separated IDs for company_images fan-out.
- `credit_ids` (optional, string); Comma-separated IDs for credit_details fan-out.
- `episode_ids` (optional, string); Comma-separated IDs for tv_episode_changes_by_id fan-out.
- `episode_number` (optional, string); Required config value for tv_episode_details.
- `external_ids` (optional, string); Comma-separated IDs for find_by_id fan-out.
- `external_source` (optional, string); External ID source required by find_by_id (for example
  imdb_id).
- `first_air_date_year` (optional, string); Optional TMDb first air date year filter.
- `guest_session_id` (optional, secret, string); Optional TMDb guest session id for guest-session
  rated endpoints.
- `guest_session_ids` (optional, string); Comma-separated IDs for guest_session_rated_tv_episodes
  fan-out.
- `include_adult` (optional, string); Optional TMDb include_adult filter.
- `keyword_ids` (optional, string); Comma-separated IDs for keyword_movies fan-out.
- `language` (optional, string); Optional ISO 639-1 language code applied to requests that support
  it.
- `list_ids` (optional, string); Comma-separated IDs for list_details fan-out.
- `movie_id` (optional, string).
- `network_ids` (optional, string); Comma-separated IDs for alternative_names_copy fan-out.
- `person_ids` (optional, string); Comma-separated IDs for translations fan-out.
- `primary_release_year` (optional, string); Optional TMDb primary release year filter.
- `query` (optional, string); Search query text required by search streams.
- `region` (optional, string); Optional TMDb region filter.
- `review_ids` (optional, string); Comma-separated IDs for review_details fan-out.
- `season_ids` (optional, string); Comma-separated IDs for tv_season_changes_by_id fan-out.
- `season_number` (optional, string); Required config value for tv_season_details.
- `series_id` (optional, string); Required config value for tv_season_details.
- `series_ids` (optional, string); Comma-separated IDs for tv_series_watch_providers fan-out.
- `session_id` (optional, secret, string); Optional TMDb user session id for account-state
  endpoints.
- `sort_by` (optional, string); Optional TMDb sort field.
- `tv_episode_group_ids` (optional, string); Comma-separated IDs for tv_episode_group_details
  fan-out.
- `watch_region` (optional, string); Optional TMDb watch region filter.
- `with_watch_providers` (optional, string); Optional TMDb watch provider filter.
- `year` (optional, string); Optional TMDb year filter.

Secret fields are redacted in logs and write previews: `api_key`, `guest_session_id`, `session_id`.

Default configuration values: `base_url=https://api.themoviedb.org/3`.

Authentication behavior:

- API key authentication in query parameter `api_key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/movie/popular`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; no page-size parameter; starts at
1; page size 20.

Pagination by stream: none: `movie_details`, `account_details`, `certification_movie_list`,
`certifications_tv_list`, `collection_details`, `collection_images`, `collection_translations`,
`company_details`, `company_images`, `configuration_details`, `configuration_countries`,
`configuration_jobs`, `configuration_languages`, `configuration_timezones`, `credit_details`,
`find_by_id`, `genre_movie_list`, `genre_tv_list`, `keyword_details`, `list_check_item_status`,
`list_details`, `movie_account_states`, `movie_alternative_titles`, `movie_changes`,
`movie_credits`, `movie_external_ids`, `movie_images`, `movie_keywords`, `movie_latest_id`,
`movie_recommendations`, `movie_translations`, `network_details`, `alternative_names_copy`,
`person_details`, `person_changes`, `person_combined_credits`, `person_external_ids`,
`person_images`, `person_latest_id`, `person_movie_credits`, `person_tv_credits`, `translations`,
`review_details`, `tv_series_details`, `tv_series_account_states`, `tv_series_aggregate_credits`,
`tv_series_changes`, `tv_series_credits`, `tv_series_external_ids`, `tv_series_images`,
`tv_series_latest_id`, `tv_series_translations`, `tv_season_details`, `tv_season_aggregate_credits`,
`tv_season_changes_by_id`, `tv_season_credits`, `tv_season_external_ids`, `tv_season_images`,
`tv_season_translations`, `tv_episode_details`, and 7 more; page_number: `popular_movies`,
`now_playing_movies`, `search_movies`, `account_get_favorites`, `account_favorite_tv`,
`account_lists`, `account_rated_movies`, `account_rated_tv`, `account_rated_tv_episodes`,
`account_watchlist_movies`, `account_watchlist_tv`, `changes_movie_list`, `changes_people_list`,
`changes_tv_list`, `company_alternative_names`, `discover_movie`, `discover_tv`,
`guest_session_rated_movies`, `guest_session_rated_tv`, `guest_session_rated_tv_episodes`,
`keyword_movies`, `movie_top_rated_list`, `movie_upcoming_list`, `movie_lists`,
`movie_release_dates`, `movie_reviews`, `movie_similar`, `movie_videos`, `movie_watch_providers`,
`details_copy`, `person_popular_list`, `person_tagged_images`, `search_collection`,
`search_company`, `search_keyword`, `search_multi`, `search_person`, `search_tv`, `trending_all`,
`trending_movies`, `trending_people`, `trending_tv`, `tv_series_airing_today_list`,
`tv_series_on_the_air_list`, `tv_series_popular_list`, `tv_series_top_rated_list`,
`tv_series_alternative_titles`, `tv_series_content_ratings`, `tv_series_episode_groups`,
`tv_series_keywords`, `lists_copy`, `tv_series_recommendations`, `tv_series_reviews`,
`tv_series_screened_theatrically`, `tv_series_similar`, `tv_series_videos`,
`tv_series_watch_providers`, `tv_season_account_states`, `tv_season_videos`,
`tv_season_watch_providers`, and 4 more.

- `popular_movies`: GET `/movie/popular` - records path `results`; query `language` from template
  `{{ config.language }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; emits passthrough records.
- `now_playing_movies`: GET `/movie/now_playing` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `search_movies`: GET `/search/movie` - records path `results`; query `language` from template `{{
  config.language }}`, omitted when absent; `query`=`{{ config.query }}`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 20; emits passthrough
  records.
- `movie_details`: GET `/movie/{{ config.movie_id }}` - records at response root; query `language`
  from template `{{ config.language }}`, omitted when absent; emits passthrough records.
- `account_details`: GET `/account/{{ fanout.id }}` - single-object response; records at response
  root; query `session_id` from template `{{ config.session_id }}`, omitted when absent; fan-out;
  ids from config field `account_ids`; id inserted into the request path; stamps `account_id`; emits
  passthrough records.
- `account_get_favorites`: GET `/account/{{ fanout.id }}/favorite/movies` - records path `results`;
  query `language` from template `{{ config.language }}`, omitted when absent; `session_id` from
  template `{{ config.session_id }}`, omitted when absent; `sort_by` from template `{{
  config.sort_by }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; fan-out; ids from config field `account_ids`; id
  inserted into the request path; stamps `account_id`; emits passthrough records.
- `account_favorite_tv`: GET `/account/{{ fanout.id }}/favorite/tv` - records path `results`; query
  `language` from template `{{ config.language }}`, omitted when absent; `session_id` from template
  `{{ config.session_id }}`, omitted when absent; `sort_by` from template `{{ config.sort_by }}`,
  omitted when absent; page-number pagination; page parameter `page`; no page-size parameter; starts
  at 1; page size 20; fan-out; ids from config field `account_ids`; id inserted into the request
  path; stamps `account_id`; emits passthrough records.
- `account_lists`: GET `/account/{{ fanout.id }}/lists` - records path `results`; query `session_id`
  from template `{{ config.session_id }}`, omitted when absent; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids from config
  field `account_ids`; id inserted into the request path; stamps `account_id`; emits passthrough
  records.
- `account_rated_movies`: GET `/account/{{ fanout.id }}/rated/movies` - records path `results`;
  query `language` from template `{{ config.language }}`, omitted when absent; `session_id` from
  template `{{ config.session_id }}`, omitted when absent; `sort_by` from template `{{
  config.sort_by }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; fan-out; ids from config field `account_ids`; id
  inserted into the request path; stamps `account_id`; emits passthrough records.
- `account_rated_tv`: GET `/account/{{ fanout.id }}/rated/tv` - records path `results`; query
  `language` from template `{{ config.language }}`, omitted when absent; `session_id` from template
  `{{ config.session_id }}`, omitted when absent; `sort_by` from template `{{ config.sort_by }}`,
  omitted when absent; page-number pagination; page parameter `page`; no page-size parameter; starts
  at 1; page size 20; fan-out; ids from config field `account_ids`; id inserted into the request
  path; stamps `account_id`; emits passthrough records.
- `account_rated_tv_episodes`: GET `/account/{{ fanout.id }}/rated/tv/episodes` - records path
  `results`; query `language` from template `{{ config.language }}`, omitted when absent;
  `session_id` from template `{{ config.session_id }}`, omitted when absent; `sort_by` from template
  `{{ config.sort_by }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; fan-out; ids from config field `account_ids`; id
  inserted into the request path; stamps `account_id`; emits passthrough records.
- `account_watchlist_movies`: GET `/account/{{ fanout.id }}/watchlist/movies` - records path
  `results`; query `language` from template `{{ config.language }}`, omitted when absent;
  `session_id` from template `{{ config.session_id }}`, omitted when absent; `sort_by` from template
  `{{ config.sort_by }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; fan-out; ids from config field `account_ids`; id
  inserted into the request path; stamps `account_id`; emits passthrough records.
- `account_watchlist_tv`: GET `/account/{{ fanout.id }}/watchlist/tv` - records path `results`;
  query `language` from template `{{ config.language }}`, omitted when absent; `session_id` from
  template `{{ config.session_id }}`, omitted when absent; `sort_by` from template `{{
  config.sort_by }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; fan-out; ids from config field `account_ids`; id
  inserted into the request path; stamps `account_id`; emits passthrough records.
- `certification_movie_list`: GET `/certification/movie/list` - single-object response; records at
  response root; emits passthrough records.
- `certifications_tv_list`: GET `/certification/tv/list` - single-object response; records at
  response root; emits passthrough records.
- `changes_movie_list`: GET `/movie/changes` - records path `results`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `changes_people_list`: GET `/person/changes` - records path `results`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 20; emits passthrough
  records.
- `changes_tv_list`: GET `/tv/changes` - records path `results`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `collection_details`: GET `/collection/{{ fanout.id }}` - single-object response; records at
  response root; query `language` from template `{{ config.language }}`, omitted when absent;
  fan-out; ids from config field `collection_ids`; id inserted into the request path; stamps
  `collection_id`; emits passthrough records.
- `collection_images`: GET `/collection/{{ fanout.id }}/images` - single-object response; records at
  response root; query `language` from template `{{ config.language }}`, omitted when absent;
  fan-out; ids from config field `collection_ids`; id inserted into the request path; stamps
  `collection_id`; emits passthrough records.
- `collection_translations`: GET `/collection/{{ fanout.id }}/translations` - single-object
  response; records at response root; fan-out; ids from config field `collection_ids`; id inserted
  into the request path; stamps `collection_id`; emits passthrough records.
- `company_details`: GET `/company/{{ fanout.id }}` - single-object response; records at response
  root; fan-out; ids from config field `company_ids`; id inserted into the request path; stamps
  `company_id`; emits passthrough records.
- `company_alternative_names`: GET `/company/{{ fanout.id }}/alternative_names` - records path
  `results`; page-number pagination; page parameter `page`; no page-size parameter; starts at 1;
  page size 20; fan-out; ids from config field `company_ids`; id inserted into the request path;
  stamps `company_id`; emits passthrough records.
- `company_images`: GET `/company/{{ fanout.id }}/images` - single-object response; records at
  response root; fan-out; ids from config field `company_ids`; id inserted into the request path;
  stamps `company_id`; emits passthrough records.
- `configuration_details`: GET `/configuration` - single-object response; records at response root;
  emits passthrough records.
- `configuration_countries`: GET `/configuration/countries` - records at response root; query
  `language` from template `{{ config.language }}`, omitted when absent; emits passthrough records.
- `configuration_jobs`: GET `/configuration/jobs` - records at response root; emits passthrough
  records.
- `configuration_languages`: GET `/configuration/languages` - records at response root; emits
  passthrough records.
- `configuration_timezones`: GET `/configuration/timezones` - records at response root; emits
  passthrough records.
- `credit_details`: GET `/credit/{{ fanout.id }}` - single-object response; records at response
  root; query `language` from template `{{ config.language }}`, omitted when absent; fan-out; ids
  from config field `credit_ids`; id inserted into the request path; stamps `credit_id`; emits
  passthrough records.
- `discover_movie`: GET `/discover/movie` - records path `results`; query `include_adult` from
  template `{{ config.include_adult }}`, omitted when absent; `language` from template `{{
  config.language }}`, omitted when absent; `primary_release_year` from template `{{
  config.primary_release_year }}`, omitted when absent; `region` from template `{{ config.region
  }}`, omitted when absent; `sort_by` from template `{{ config.sort_by }}`, omitted when absent;
  `watch_region` from template `{{ config.watch_region }}`, omitted when absent;
  `with_watch_providers` from template `{{ config.with_watch_providers }}`, omitted when absent;
  `year` from template `{{ config.year }}`, omitted when absent; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `discover_tv`: GET `/discover/tv` - records path `results`; query `first_air_date_year` from
  template `{{ config.first_air_date_year }}`, omitted when absent; `include_adult` from template
  `{{ config.include_adult }}`, omitted when absent; `language` from template `{{ config.language
  }}`, omitted when absent; `sort_by` from template `{{ config.sort_by }}`, omitted when absent;
  `watch_region` from template `{{ config.watch_region }}`, omitted when absent;
  `with_watch_providers` from template `{{ config.with_watch_providers }}`, omitted when absent;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20;
  emits passthrough records.
- `find_by_id`: GET `/find/{{ fanout.id }}` - single-object response; records at response root;
  query `external_source`=`{{ config.external_source }}`; `language` from template `{{
  config.language }}`, omitted when absent; fan-out; ids from config field `external_ids`; id
  inserted into the request path; stamps `external_id`; emits passthrough records.
- `genre_movie_list`: GET `/genre/movie/list` - single-object response; records at response root;
  query `language` from template `{{ config.language }}`, omitted when absent; emits passthrough
  records.
- `genre_tv_list`: GET `/genre/tv/list` - single-object response; records at response root; query
  `language` from template `{{ config.language }}`, omitted when absent; emits passthrough records.
- `guest_session_rated_movies`: GET `/guest_session/{{ fanout.id }}/rated/movies` - records path
  `results`; query `language` from template `{{ config.language }}`, omitted when absent; `sort_by`
  from template `{{ config.sort_by }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids from config field
  `guest_session_ids`; id inserted into the request path; stamps `guest_session_id`; emits
  passthrough records.
- `guest_session_rated_tv`: GET `/guest_session/{{ fanout.id }}/rated/tv` - records path `results`;
  query `language` from template `{{ config.language }}`, omitted when absent; `sort_by` from
  template `{{ config.sort_by }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids from config field
  `guest_session_ids`; id inserted into the request path; stamps `guest_session_id`; emits
  passthrough records.
- `guest_session_rated_tv_episodes`: GET `/guest_session/{{ fanout.id }}/rated/tv/episodes` -
  records path `results`; query `language` from template `{{ config.language }}`, omitted when
  absent; `sort_by` from template `{{ config.sort_by }}`, omitted when absent; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids
  from config field `guest_session_ids`; id inserted into the request path; stamps
  `guest_session_id`; emits passthrough records.
- `keyword_details`: GET `/keyword/{{ fanout.id }}` - single-object response; records at response
  root; fan-out; ids from config field `keyword_ids`; id inserted into the request path; stamps
  `keyword_id`; emits passthrough records.
- `keyword_movies`: GET `/keyword/{{ fanout.id }}/movies` - records path `results`; query
  `include_adult` from template `{{ config.include_adult }}`, omitted when absent; `language` from
  template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids from config field
  `keyword_ids`; id inserted into the request path; stamps `keyword_id`; emits passthrough records.
- `list_check_item_status`: GET `/list/{{ fanout.id }}/item_status` - single-object response;
  records at response root; query `language` from template `{{ config.language }}`, omitted when
  absent; fan-out; ids from config field `list_ids`; id inserted into the request path; stamps
  `list_id`; emits passthrough records.
- `list_details`: GET `/list/{{ fanout.id }}` - single-object response; records at response root;
  query `language` from template `{{ config.language }}`, omitted when absent; fan-out; ids from
  config field `list_ids`; id inserted into the request path; stamps `list_id`; emits passthrough
  records.
- `movie_top_rated_list`: GET `/movie/top_rated` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; `region` from template `{{ config.region
  }}`, omitted when absent; page-number pagination; page parameter `page`; no page-size parameter;
  starts at 1; page size 20; emits passthrough records.
- `movie_upcoming_list`: GET `/movie/upcoming` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; `region` from template `{{ config.region
  }}`, omitted when absent; page-number pagination; page parameter `page`; no page-size parameter;
  starts at 1; page size 20; emits passthrough records.
- `movie_account_states`: GET `/movie/{{ config.movie_id }}/account_states` - single-object
  response; records at response root; query `guest_session_id` from template `{{
  config.guest_session_id }}`, omitted when absent; `session_id` from template `{{ config.session_id
  }}`, omitted when absent; emits passthrough records.
- `movie_alternative_titles`: GET `/movie/{{ config.movie_id }}/alternative_titles` - single-object
  response; records at response root; emits passthrough records.
- `movie_changes`: GET `/movie/{{ config.movie_id }}/changes` - single-object response; records at
  response root; emits passthrough records.
- `movie_credits`: GET `/movie/{{ config.movie_id }}/credits` - single-object response; records at
  response root; query `language` from template `{{ config.language }}`, omitted when absent; emits
  passthrough records.
- `movie_external_ids`: GET `/movie/{{ config.movie_id }}/external_ids` - single-object response;
  records at response root; emits passthrough records.
- `movie_images`: GET `/movie/{{ config.movie_id }}/images` - single-object response; records at
  response root; query `language` from template `{{ config.language }}`, omitted when absent; emits
  passthrough records.
- `movie_keywords`: GET `/movie/{{ config.movie_id }}/keywords` - single-object response; records at
  response root; emits passthrough records.
- `movie_latest_id`: GET `/movie/latest` - single-object response; records at response root; emits
  passthrough records.
- `movie_lists`: GET `/movie/{{ config.movie_id }}/lists` - records path `results`; query `language`
  from template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `movie_recommendations`: GET `/movie/{{ config.movie_id }}/recommendations` - single-object
  response; records at response root; query `language` from template `{{ config.language }}`,
  omitted when absent; emits passthrough records.
- `movie_release_dates`: GET `/movie/{{ config.movie_id }}/release_dates` - records path `results`;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20;
  emits passthrough records.
- `movie_reviews`: GET `/movie/{{ config.movie_id }}/reviews` - records path `results`; query
  `language` from template `{{ config.language }}`, omitted when absent; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 20; emits passthrough
  records.
- `movie_similar`: GET `/movie/{{ config.movie_id }}/similar` - records path `results`; query
  `language` from template `{{ config.language }}`, omitted when absent; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 20; emits passthrough
  records.
- `movie_translations`: GET `/movie/{{ config.movie_id }}/translations` - single-object response;
  records at response root; emits passthrough records.
- `movie_videos`: GET `/movie/{{ config.movie_id }}/videos` - records path `results`; query
  `language` from template `{{ config.language }}`, omitted when absent; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 20; emits passthrough
  records.
- `movie_watch_providers`: GET `/movie/{{ config.movie_id }}/watch/providers` - records path
  `results`; page-number pagination; page parameter `page`; no page-size parameter; starts at 1;
  page size 20; emits passthrough records.
- `network_details`: GET `/network/{{ fanout.id }}` - single-object response; records at response
  root; fan-out; ids from config field `network_ids`; id inserted into the request path; stamps
  `network_id`; emits passthrough records.
- `details_copy`: GET `/network/{{ fanout.id }}/alternative_names` - records path `results`;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20;
  fan-out; ids from config field `network_ids`; id inserted into the request path; stamps
  `network_id`; emits passthrough records.
- `alternative_names_copy`: GET `/network/{{ fanout.id }}/images` - single-object response; records
  at response root; fan-out; ids from config field `network_ids`; id inserted into the request path;
  stamps `network_id`; emits passthrough records.
- `person_popular_list`: GET `/person/popular` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `person_details`: GET `/person/{{ fanout.id }}` - single-object response; records at response
  root; query `language` from template `{{ config.language }}`, omitted when absent; fan-out; ids
  from config field `person_ids`; id inserted into the request path; stamps `person_id`; emits
  passthrough records.
- `person_changes`: GET `/person/{{ fanout.id }}/changes` - single-object response; records at
  response root; fan-out; ids from config field `person_ids`; id inserted into the request path;
  stamps `person_id`; emits passthrough records.
- `person_combined_credits`: GET `/person/{{ fanout.id }}/combined_credits` - single-object
  response; records at response root; query `language` from template `{{ config.language }}`,
  omitted when absent; fan-out; ids from config field `person_ids`; id inserted into the request
  path; stamps `person_id`; emits passthrough records.
- `person_external_ids`: GET `/person/{{ fanout.id }}/external_ids` - single-object response;
  records at response root; fan-out; ids from config field `person_ids`; id inserted into the
  request path; stamps `person_id`; emits passthrough records.
- `person_images`: GET `/person/{{ fanout.id }}/images` - single-object response; records at
  response root; fan-out; ids from config field `person_ids`; id inserted into the request path;
  stamps `person_id`; emits passthrough records.
- `person_latest_id`: GET `/person/latest` - single-object response; records at response root; emits
  passthrough records.
- `person_movie_credits`: GET `/person/{{ fanout.id }}/movie_credits` - single-object response;
  records at response root; query `language` from template `{{ config.language }}`, omitted when
  absent; fan-out; ids from config field `person_ids`; id inserted into the request path; stamps
  `person_id`; emits passthrough records.
- `person_tv_credits`: GET `/person/{{ fanout.id }}/tv_credits` - single-object response; records at
  response root; query `language` from template `{{ config.language }}`, omitted when absent;
  fan-out; ids from config field `person_ids`; id inserted into the request path; stamps
  `person_id`; emits passthrough records.
- `person_tagged_images`: GET `/person/{{ fanout.id }}/tagged_images` - records path `results`;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20;
  fan-out; ids from config field `person_ids`; id inserted into the request path; stamps
  `person_id`; emits passthrough records.
- `translations`: GET `/person/{{ fanout.id }}/translations` - single-object response; records at
  response root; fan-out; ids from config field `person_ids`; id inserted into the request path;
  stamps `person_id`; emits passthrough records.
- `review_details`: GET `/review/{{ fanout.id }}` - single-object response; records at response
  root; fan-out; ids from config field `review_ids`; id inserted into the request path; stamps
  `review_id`; emits passthrough records.
- `search_collection`: GET `/search/collection` - records path `results`; query `include_adult` from
  template `{{ config.include_adult }}`, omitted when absent; `language` from template `{{
  config.language }}`, omitted when absent; `query`=`{{ config.query }}`; `region` from template `{{
  config.region }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; emits passthrough records.
- `search_company`: GET `/search/company` - records path `results`; query `query`=`{{ config.query
  }}`; page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size
  20; emits passthrough records.
- `search_keyword`: GET `/search/keyword` - records path `results`; query `query`=`{{ config.query
  }}`; page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size
  20; emits passthrough records.
- `search_multi`: GET `/search/multi` - records path `results`; query `include_adult` from template
  `{{ config.include_adult }}`, omitted when absent; `language` from template `{{ config.language
  }}`, omitted when absent; `query`=`{{ config.query }}`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `search_person`: GET `/search/person` - records path `results`; query `include_adult` from
  template `{{ config.include_adult }}`, omitted when absent; `language` from template `{{
  config.language }}`, omitted when absent; `query`=`{{ config.query }}`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 20; emits passthrough
  records.
- `search_tv`: GET `/search/tv` - records path `results`; query `first_air_date_year` from template
  `{{ config.first_air_date_year }}`, omitted when absent; `include_adult` from template `{{
  config.include_adult }}`, omitted when absent; `language` from template `{{ config.language }}`,
  omitted when absent; `query`=`{{ config.query }}`; `year` from template `{{ config.year }}`,
  omitted when absent; page-number pagination; page parameter `page`; no page-size parameter; starts
  at 1; page size 20; emits passthrough records.
- `trending_all`: GET `/trending/all/day` - records path `results`; query `language` from template
  `{{ config.language }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; emits passthrough records.
- `trending_movies`: GET `/trending/movie/day` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `trending_people`: GET `/trending/person/day` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `trending_tv`: GET `/trending/tv/day` - records path `results`; query `language` from template `{{
  config.language }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; emits passthrough records.
- `tv_series_airing_today_list`: GET `/tv/airing_today` - records path `results`; query `language`
  from template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `tv_series_on_the_air_list`: GET `/tv/on_the_air` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `tv_series_popular_list`: GET `/tv/popular` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `tv_series_top_rated_list`: GET `/tv/top_rated` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `tv_series_details`: GET `/tv/{{ fanout.id }}` - single-object response; records at response root;
  query `language` from template `{{ config.language }}`, omitted when absent; fan-out; ids from
  config field `series_ids`; id inserted into the request path; stamps `series_id`; emits
  passthrough records.
- `tv_series_account_states`: GET `/tv/{{ fanout.id }}/account_states` - single-object response;
  records at response root; query `guest_session_id` from template `{{ config.guest_session_id }}`,
  omitted when absent; `session_id` from template `{{ config.session_id }}`, omitted when absent;
  fan-out; ids from config field `series_ids`; id inserted into the request path; stamps
  `series_id`; emits passthrough records.
- `tv_series_aggregate_credits`: GET `/tv/{{ fanout.id }}/aggregate_credits` - single-object
  response; records at response root; query `language` from template `{{ config.language }}`,
  omitted when absent; fan-out; ids from config field `series_ids`; id inserted into the request
  path; stamps `series_id`; emits passthrough records.
- `tv_series_alternative_titles`: GET `/tv/{{ fanout.id }}/alternative_titles` - records path
  `results`; page-number pagination; page parameter `page`; no page-size parameter; starts at 1;
  page size 20; fan-out; ids from config field `series_ids`; id inserted into the request path;
  stamps `series_id`; emits passthrough records.
- `tv_series_changes`: GET `/tv/{{ fanout.id }}/changes` - single-object response; records at
  response root; fan-out; ids from config field `series_ids`; id inserted into the request path;
  stamps `series_id`; emits passthrough records.
- `tv_series_content_ratings`: GET `/tv/{{ fanout.id }}/content_ratings` - records path `results`;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20;
  fan-out; ids from config field `series_ids`; id inserted into the request path; stamps
  `series_id`; emits passthrough records.
- `tv_series_credits`: GET `/tv/{{ fanout.id }}/credits` - single-object response; records at
  response root; query `language` from template `{{ config.language }}`, omitted when absent;
  fan-out; ids from config field `series_ids`; id inserted into the request path; stamps
  `series_id`; emits passthrough records.
- `tv_series_episode_groups`: GET `/tv/{{ fanout.id }}/episode_groups` - records path `results`;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20;
  fan-out; ids from config field `series_ids`; id inserted into the request path; stamps
  `series_id`; emits passthrough records.
- `tv_series_external_ids`: GET `/tv/{{ fanout.id }}/external_ids` - single-object response; records
  at response root; fan-out; ids from config field `series_ids`; id inserted into the request path;
  stamps `series_id`; emits passthrough records.
- `tv_series_images`: GET `/tv/{{ fanout.id }}/images` - single-object response; records at response
  root; query `language` from template `{{ config.language }}`, omitted when absent; fan-out; ids
  from config field `series_ids`; id inserted into the request path; stamps `series_id`; emits
  passthrough records.
- `tv_series_keywords`: GET `/tv/{{ fanout.id }}/keywords` - records path `results`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids
  from config field `series_ids`; id inserted into the request path; stamps `series_id`; emits
  passthrough records.
- `tv_series_latest_id`: GET `/tv/latest` - single-object response; records at response root; emits
  passthrough records.
- `lists_copy`: GET `/tv/{{ fanout.id }}/lists` - records path `results`; query `language` from
  template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids from config field
  `series_ids`; id inserted into the request path; stamps `series_id`; emits passthrough records.
- `tv_series_recommendations`: GET `/tv/{{ fanout.id }}/recommendations` - records path `results`;
  query `language` from template `{{ config.language }}`, omitted when absent; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids
  from config field `series_ids`; id inserted into the request path; stamps `series_id`; emits
  passthrough records.
- `tv_series_reviews`: GET `/tv/{{ fanout.id }}/reviews` - records path `results`; query `language`
  from template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids from config field
  `series_ids`; id inserted into the request path; stamps `series_id`; emits passthrough records.
- `tv_series_screened_theatrically`: GET `/tv/{{ fanout.id }}/screened_theatrically` - records path
  `results`; page-number pagination; page parameter `page`; no page-size parameter; starts at 1;
  page size 20; fan-out; ids from config field `series_ids`; id inserted into the request path;
  stamps `series_id`; emits passthrough records.
- `tv_series_similar`: GET `/tv/{{ fanout.id }}/similar` - records path `results`; query `language`
  from template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids from config field
  `series_ids`; id inserted into the request path; stamps `series_id`; emits passthrough records.
- `tv_series_translations`: GET `/tv/{{ fanout.id }}/translations` - single-object response; records
  at response root; fan-out; ids from config field `series_ids`; id inserted into the request path;
  stamps `series_id`; emits passthrough records.
- `tv_series_videos`: GET `/tv/{{ fanout.id }}/videos` - records path `results`; query `language`
  from template `{{ config.language }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; fan-out; ids from config field
  `series_ids`; id inserted into the request path; stamps `series_id`; emits passthrough records.
- `tv_series_watch_providers`: GET `/tv/{{ fanout.id }}/watch/providers` - records path `results`;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20;
  fan-out; ids from config field `series_ids`; id inserted into the request path; stamps
  `series_id`; emits passthrough records.
- `tv_season_details`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number }}` -
  single-object response; records at response root; query `language` from template `{{
  config.language }}`, omitted when absent; emits passthrough records.
- `tv_season_account_states`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number
  }}/account_states` - records path `results`; query `guest_session_id` from template `{{
  config.guest_session_id }}`, omitted when absent; `session_id` from template `{{ config.session_id
  }}`, omitted when absent; page-number pagination; page parameter `page`; no page-size parameter;
  starts at 1; page size 20; emits passthrough records.
- `tv_season_aggregate_credits`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number
  }}/aggregate_credits` - single-object response; records at response root; query `language` from
  template `{{ config.language }}`, omitted when absent; emits passthrough records.
- `tv_season_changes_by_id`: GET `/tv/season/{{ fanout.id }}/changes` - single-object response;
  records at response root; fan-out; ids from config field `season_ids`; id inserted into the
  request path; stamps `season_id`; emits passthrough records.
- `tv_season_credits`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number }}/credits` -
  single-object response; records at response root; query `language` from template `{{
  config.language }}`, omitted when absent; emits passthrough records.
- `tv_season_external_ids`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number
  }}/external_ids` - single-object response; records at response root; emits passthrough records.
- `tv_season_images`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number }}/images` -
  single-object response; records at response root; query `language` from template `{{
  config.language }}`, omitted when absent; emits passthrough records.
- `tv_season_translations`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number
  }}/translations` - single-object response; records at response root; emits passthrough records.
- `tv_season_videos`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number }}/videos` -
  records path `results`; query `language` from template `{{ config.language }}`, omitted when
  absent; page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page
  size 20; emits passthrough records.
- `tv_season_watch_providers`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number
  }}/watch/providers` - records path `results`; query `language` from template `{{ config.language
  }}`, omitted when absent; page-number pagination; page parameter `page`; no page-size parameter;
  starts at 1; page size 20; emits passthrough records.
- `tv_episode_details`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number }}/episode/{{
  config.episode_number }}` - single-object response; records at response root; query `language`
  from template `{{ config.language }}`, omitted when absent; emits passthrough records.
- `tv_episode_account_states`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number
  }}/episode/{{ config.episode_number }}/account_states` - single-object response; records at
  response root; query `guest_session_id` from template `{{ config.guest_session_id }}`, omitted
  when absent; `session_id` from template `{{ config.session_id }}`, omitted when absent; emits
  passthrough records.
- `tv_episode_changes_by_id`: GET `/tv/episode/{{ fanout.id }}/changes` - single-object response;
  records at response root; fan-out; ids from config field `episode_ids`; id inserted into the
  request path; stamps `episode_id`; emits passthrough records.
- `tv_episode_credits`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number }}/episode/{{
  config.episode_number }}/credits` - single-object response; records at response root; query
  `language` from template `{{ config.language }}`, omitted when absent; emits passthrough records.
- `tv_episode_external_ids`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number
  }}/episode/{{ config.episode_number }}/external_ids` - single-object response; records at response
  root; emits passthrough records.
- `tv_episode_images`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number }}/episode/{{
  config.episode_number }}/images` - single-object response; records at response root; query
  `language` from template `{{ config.language }}`, omitted when absent; emits passthrough records.
- `tv_episode_translations`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number
  }}/episode/{{ config.episode_number }}/translations` - single-object response; records at response
  root; emits passthrough records.
- `tv_episode_videos`: GET `/tv/{{ config.series_id }}/season/{{ config.season_number }}/episode/{{
  config.episode_number }}/videos` - records path `results`; query `language` from template `{{
  config.language }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; emits passthrough records.
- `tv_episode_group_details`: GET `/tv/episode_group/{{ fanout.id }}` - single-object response;
  records at response root; fan-out; ids from config field `tv_episode_group_ids`; id inserted into
  the request path; stamps `tv_episode_group_id`; emits passthrough records.
- `watch_providers_available_regions`: GET `/watch/providers/regions` - records path `results`;
  query `language` from template `{{ config.language }}`, omitted when absent; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 20; emits
  passthrough records.
- `watch_providers_movie_list`: GET `/watch/providers/movie` - records path `results`; query
  `language` from template `{{ config.language }}`, omitted when absent; `watch_region` from
  template `{{ config.watch_region }}`, omitted when absent; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 1; page size 20; emits passthrough records.
- `watch_provider_tv_list`: GET `/watch/providers/tv` - records path `results`; query `language`
  from template `{{ config.language }}`, omitted when absent; `watch_region` from template `{{
  config.watch_region }}`, omitted when absent; page-number pagination; page parameter `page`; no
  page-size parameter; starts at 1; page size 20; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external TMDb API read of public catalog, search,
account-state, and reference metadata.

## Known limits

- Batch defaults: read_page_size=20.
- API coverage includes 131 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=4, non_data_endpoint=7, requires_elevated_scope=10.
