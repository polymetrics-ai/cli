---
name: pm-buzzsprout
description: Buzzsprout connector knowledge and safe action guide.
---

# pm-buzzsprout

## Purpose

Reads Buzzsprout podcasts and episodes (titles, publish dates, durations, play counts) and creates/updates episodes through the Buzzsprout REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- podcast_id
- api_key (secret) (required)

## ETL Streams

- episodes:
  - primary key: id
  - cursor: published_at
  - fields: artist(string), artwork_url(string), audio_url(string), description(string), duration(integer), episode_number(integer), explicit(boolean), guid(string), hq(boolean), id(integer), inactive_at(string), magic_mastering(boolean), private(boolean), published_at(string), season_number(integer), summary(string), tags(string), title(string), total_plays(integer)
- podcasts:
  - primary key: id
  - fields: artwork_url(string), author(string), contact_email(string), description(string), explicit(boolean), id(integer), keywords(string), language(string), main_category(string), sub_category(string), timezone(string), title(string), website_address(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_episode:
  - endpoint: POST /api/{{ config.podcast_id }}/episodes.json
  - risk: external mutation; creates a new episode (and can trigger audio processing/publication) on the configured podcast; approval required
- update_episode:
  - endpoint: PUT /api/{{ config.podcast_id }}/episodes/{{ record.id }}.json
  - required fields: id
  - risk: external mutation; overwrites episode metadata on the configured podcast; approval required

## Security

- read risk: external Buzzsprout API read of podcast and episode data
- write risk: external mutation of episode metadata/audio (create_episode, update_episode) on the configured podcast; can trigger audio processing and publish/unpublish an episode
- approval: required for create_episode/update_episode
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect buzzsprout
```

### Inspect as structured JSON

```bash
pm connectors inspect buzzsprout --json
```

## Agent Rules

- Run pm connectors inspect buzzsprout before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
