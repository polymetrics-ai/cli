---
name: pm-strava
description: Strava connector knowledge and safe action guide.
---

# pm-strava

## Purpose

Reads the authenticated Strava athlete's profile, activities, lifetime stats, and clubs through the Strava v3 REST API.

## Icon

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
- client_id
- token_url
- client_secret (secret)
- refresh_token (secret)

## ETL Streams

- activities:
  - primary key: id
  - cursor: start_date
  - fields: achievement_count(), average_speed(), distance(), elapsed_time(), id(), kudos_count(), max_speed(), moving_time(), name(), sport_type(), start_date(), start_date_local(), timezone(), total_elevation_gain(), type()
- athlete:
  - primary key: id
  - fields: city(), country(), created_at(), firstname(), id(), lastname(), sex(), state(), updated_at(), username(), weight()
- athlete_stats:
  - primary key: id
  - fields: all_ride_totals(), all_run_totals(), all_swim_totals(), biggest_climb_elevation_gain(), biggest_ride_distance(), id(), recent_ride_totals(), recent_run_totals(), recent_swim_totals(), ytd_ride_totals(), ytd_run_totals(), ytd_swim_totals()
- clubs:
  - primary key: id
  - fields: city(), country(), id(), member_count(), membership(), name(), private(), sport_type(), state(), url()

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
