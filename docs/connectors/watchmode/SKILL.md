---
name: pm-watchmode
description: Watchmode connector knowledge and safe action guide.
---

# pm-watchmode

## Purpose

Reads Watchmode title search results, streaming sources, regions, networks, genres, list-titles, releases, per-title details/sources/seasons/episodes/cast-crew, and person details. Read-only.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- person_ids
- regions
- search_val
- start_date
- title_ids
- types
- api_key (secret) (required)

## ETL Streams

- search:
  - primary key: id
  - fields: id(integer), name(string), type(string), year(integer)
- sources:
  - primary key: id
  - fields: id(integer), name(string), region(string), type(string)
- regions:
  - primary key: country
  - fields: country(string), data_tier(integer), flag(string), name(string), plan_enabled(boolean)
- networks:
  - primary key: id
  - fields: id(integer), name(string), origin_country(string), tmdb_id(integer)
- genres:
  - primary key: id
  - fields: id(integer), name(string), tmdb_id(integer)
- titles:
  - primary key: id
  - fields: id(integer), imdb_id(string), title(string), tmdb_id(integer), tmdb_type(string), type(string), year(integer)
- releases:
  - primary key: id, source_id, source_release_date
  - fields: id(integer), imdb_id(string), is_original(integer), poster_url(string), season_number(integer), source_id(integer), source_name(string), source_release_date(string), title(string), tmdb_id(integer), tmdb_type(string), type(string)
- title_details:
  - primary key: id
  - fields: backdrop(string), critic_score(number), end_year(integer), genre_names(array), id(integer), imdb_id(string), original_title(string), plot_overview(string), poster(string), release_date(string), runtime_minutes(integer), title(string), tmdb_id(integer), tmdb_type(string), type(string), us_rating(string), user_rating(number), watchmode_title_id(string), year(integer)
- title_sources:
  - primary key: watchmode_title_id, source_id, region, type
  - fields: episodes(integer), format(string), name(string), price(number), region(string), seasons(integer), source_id(integer), type(string), watchmode_title_id(string), web_url(string)
- title_seasons:
  - primary key: id
  - fields: air_date(string), episode_count(integer), id(integer), name(string), number(integer), overview(string), poster_url(string), watchmode_title_id(string)
- title_episodes:
  - primary key: id
  - fields: episode_number(integer), id(integer), imdb_id(string), name(string), overview(string), release_date(string), runtime_minutes(integer), season_id(integer), season_number(integer), sources(array), tmdb_id(integer), watchmode_title_id(string)
- title_cast_crew:
  - primary key: watchmode_title_id, person_id, type, role
  - fields: episode_count(integer), full_name(string), order(integer), person_id(integer), role(string), type(string), watchmode_title_id(string)
- person_details:
  - primary key: id
  - fields: date_of_birth(string), date_of_death(string), first_name(string), full_name(string), gender(string), id(integer), imdb_id(string), known_for(array), last_name(string), main_profession(string), place_of_birth(string), relevance_percentile(number), tmdb_id(integer), watchmode_person_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Watchmode API read of public title/streaming-source/person media metadata
- approval: none; read-only public media metadata connector
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect watchmode
```

### Inspect as structured JSON

```bash
pm connectors inspect watchmode --json
```

## Agent Rules

- Run pm connectors inspect watchmode before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
