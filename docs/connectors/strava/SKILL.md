---
name: pm-strava
description: Strava connector knowledge and safe action guide.
---

# pm-strava

## Purpose

Reads the authenticated Strava athlete's profile, activities, lifetime stats, and clubs through the Strava v3 REST API.

## Icon

- id: strava
- asset: icons/strava.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.strava.com/docs/reference/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- athlete_id
- base_url
- client_id (required)
- token_url
- client_secret (secret) (required)
- refresh_token (secret) (required)

## ETL Streams

- activities:
  - primary key: id
  - cursor: start_date
  - fields: achievement_count(integer), average_speed(number), distance(number), elapsed_time(integer), id(integer), kudos_count(integer), max_speed(number), moving_time(integer), name(string), sport_type(string), start_date(string), start_date_local(string), timezone(string), total_elevation_gain(number), type(string)
- athlete:
  - primary key: id
  - fields: city(string), country(string), created_at(string), firstname(string), id(integer), lastname(string), sex(string), state(string), updated_at(string), username(string), weight(number)
- athlete_stats:
  - primary key: id
  - fields: all_ride_totals(object), all_run_totals(object), all_swim_totals(object), biggest_climb_elevation_gain(number), biggest_ride_distance(number), id(integer), recent_ride_totals(object), recent_run_totals(object), recent_swim_totals(object), ytd_ride_totals(object), ytd_run_totals(object), ytd_swim_totals(object)
- clubs:
  - primary key: id
  - fields: city(string), country(string), id(integer), member_count(integer), membership(string), name(string), private(boolean), sport_type(string), state(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Strava API read of the authenticated athlete's profile, activity, stats, and club data
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect strava
```

### Inspect as structured JSON

```bash
pm connectors inspect strava --json
```

## Agent Rules

- Run pm connectors inspect strava before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
