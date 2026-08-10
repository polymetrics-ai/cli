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
    fields: id(integer), name(string), type(string), year(integer)
  sources:
    primary key: id
    fields: id(integer), name(string), region(string), type(string)
  regions:
    primary key: country
    fields: country(string), data_tier(integer), flag(string), name(string), plan_enabled(boolean)
  networks:
    primary key: id
    fields: id(integer), name(string), origin_country(string), tmdb_id(integer)
  genres:
    primary key: id
    fields: id(integer), name(string), tmdb_id(integer)
  titles:
    primary key: id
    fields: id(integer), imdb_id(string), title(string), tmdb_id(integer), tmdb_type(string), type(string), year(integer)
  releases:
    primary key: id, source_id, source_release_date
    fields: id(integer), imdb_id(string), is_original(integer), poster_url(string), season_number(integer), source_id(integer), source_name(string), source_release_date(string), title(string), tmdb_id(integer), tmdb_type(string), type(string)
  title_details:
    primary key: id
    fields: backdrop(string), critic_score(number), end_year(integer), genre_names(array), id(integer), imdb_id(string), original_title(string), plot_overview(string), poster(string), release_date(string), runtime_minutes(integer), title(string), tmdb_id(integer), tmdb_type(string), type(string), us_rating(string), user_rating(number), watchmode_title_id(string), year(integer)
  title_sources:
    primary key: watchmode_title_id, source_id, region, type
    fields: episodes(integer), format(string), name(string), price(number), region(string), seasons(integer), source_id(integer), type(string), watchmode_title_id(string), web_url(string)
  title_seasons:
    primary key: id
    fields: air_date(string), episode_count(integer), id(integer), name(string), number(integer), overview(string), poster_url(string), watchmode_title_id(string)
  title_episodes:
    primary key: id
    fields: episode_number(integer), id(integer), imdb_id(string), name(string), overview(string), release_date(string), runtime_minutes(integer), season_id(integer), season_number(integer), sources(array), tmdb_id(integer), watchmode_title_id(string)
  title_cast_crew:
    primary key: watchmode_title_id, person_id, type, role
    fields: episode_count(integer), full_name(string), order(integer), person_id(integer), role(string), type(string), watchmode_title_id(string)
  person_details:
    primary key: id
    fields: date_of_birth(string), date_of_death(string), first_name(string), full_name(string), gender(string), id(integer), imdb_id(string), known_for(array), last_name(string), main_profession(string), place_of_birth(string), relevance_percentile(number), tmdb_id(integer), watchmode_person_id(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Watchmode API read of public title/streaming-source/person media metadata
  approval: none; read-only public media metadata connector
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Watchmode's declared streams and reverse-ETL actions.
  Usage: pm watchmode <command> [flags]
  Read streams
  Other Commands
    api get autocomplete-search - Documented GET /autocomplete-search (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.autocomplete-search]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get changes new-people - Documented GET /changes/new_people (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.changes-new-people]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get changes new-titles - Documented GET /changes/new_titles (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.changes-new-titles]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get changes titles-details-changed - Documented GET /changes/titles_details_changed (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.changes-titles-details-changed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get changes titles-episodes-changed - Documented GET /changes/titles_episodes_changed (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.changes-titles-episodes-changed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get changes titles-sources-changed - Documented GET /changes/titles_sources_changed (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.changes-titles-sources-changed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get genres - Documented GET /genres (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.genres]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get list-titles - Documented GET /list-titles (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.list-titles]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get networks - Documented GET /networks (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.networks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get person person-id - Documented GET /person/{person_id} (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.person-person-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get regions - Documented GET /regions (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.regions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get releases - Documented GET /releases (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.releases]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get search - Documented GET /search (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.search]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get sources - Documented GET /sources (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.sources]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get status - Documented GET /status (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get title title-id cast-crew - Documented GET /title/{title_id}/cast-crew (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.title-title-id-cast-crew]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get title title-id details - Documented GET /title/{title_id}/details (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.title-title-id-details]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get title title-id episodes - Documented GET /title/{title_id}/episodes (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.title-title-id-episodes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get title title-id episodes episode-id - Documented GET /title/{title_id}/episodes/{episode_id} (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.title-title-id-episodes-episode-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get title title-id incorrect-data - Documented GET /title/{title_id}/incorrect-data (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.title-title-id-incorrect-data]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get title title-id seasons - Documented GET /title/{title_id}/seasons (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.title-title-id-seasons]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get title title-id sources - Documented GET /title/{title_id}/sources (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.title-title-id-sources]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get title-release-dates - Documented GET /title-release-dates (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.title-release-dates]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 autocomplete-search - Documented GET /v1/autocomplete-search/ (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.v1-autocomplete-search]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 changes new-people - Documented GET /v1/changes/new_people/ (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.v1-changes-new-people]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 changes new-titles - Documented GET /v1/changes/new_titles/ (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.v1-changes-new-titles]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 changes titles-details-changed - Documented GET /v1/changes/titles_details_changed/ (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.v1-changes-titles-details-changed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 changes titles-episodes-changed - Documented GET /v1/changes/titles_episodes_changed/ (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.v1-changes-titles-episodes-changed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 changes titles-sources-changed - Documented GET /v1/changes/titles_sources_changed/ (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.v1-changes-titles-sources-changed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 status - Documented GET /v1/status/ (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.v1-status]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 title title-id incorrect-data - Documented GET /v1/title/{title_id}/incorrect-data/ (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.v1-title-title-id-incorrect-data]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 title-release-dates - Documented GET /v1/title-release-dates/ (not implemented) [intent=direct_read availability=not_implemented operation=watchmode.get.v1-title-release-dates]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    genres list - Run the genres ETL stream [intent=etl availability=implemented stream=genres]; notes: discrepancy=present-in-surface-absent-from-artifact
    networks list - Run the networks ETL stream [intent=etl availability=implemented stream=networks]; notes: discrepancy=present-in-surface-absent-from-artifact
    person details list - Run the person details ETL stream [intent=etl availability=implemented stream=person_details]; notes: discrepancy=present-in-surface-absent-from-artifact
    regions list - Run the regions ETL stream [intent=etl availability=implemented stream=regions]; notes: discrepancy=present-in-surface-absent-from-artifact
    releases list - Run the releases ETL stream [intent=etl availability=implemented stream=releases]; notes: discrepancy=present-in-surface-absent-from-artifact
    search list - Run the search ETL stream [intent=etl availability=implemented stream=search]; notes: discrepancy=present-in-surface-absent-from-artifact
    sources list - Run the sources ETL stream [intent=etl availability=implemented stream=sources]; notes: discrepancy=present-in-surface-absent-from-artifact
    title cast crew list - Run the title cast crew ETL stream [intent=etl availability=implemented stream=title_cast_crew]; notes: discrepancy=present-in-surface-absent-from-artifact
    title details list - Run the title details ETL stream [intent=etl availability=implemented stream=title_details]; notes: discrepancy=present-in-surface-absent-from-artifact
    title episodes list - Run the title episodes ETL stream [intent=etl availability=implemented stream=title_episodes]; notes: discrepancy=present-in-surface-absent-from-artifact
    title seasons list - Run the title seasons ETL stream [intent=etl availability=implemented stream=title_seasons]; notes: discrepancy=present-in-surface-absent-from-artifact
    title sources list - Run the title sources ETL stream [intent=etl availability=implemented stream=title_sources]; notes: discrepancy=present-in-surface-absent-from-artifact
    titles list - Run the titles ETL stream [intent=etl availability=implemented stream=titles]; notes: discrepancy=present-in-surface-absent-from-artifact

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
