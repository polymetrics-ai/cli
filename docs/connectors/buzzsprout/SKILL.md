---
name: pm-buzzsprout
description: Buzzsprout connector knowledge and safe action guide.
---

# pm-buzzsprout

## Purpose

Reads Buzzsprout podcasts and episodes (titles, publish dates, durations, play counts) and creates/updates episodes through the Buzzsprout REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- podcast_id
- api_key (secret)

## ETL Streams

- episodes:
  - primary key: id
  - cursor: published_at
  - fields: artist(), artwork_url(), audio_url(), description(), duration(), episode_number(), explicit(), guid(), hq(), id(), inactive_at(), magic_mastering(), private(), published_at(), season_number(), summary(), tags(), title(), total_plays()
- podcasts:
  - primary key: id
  - fields: artwork_url(), author(), contact_email(), description(), explicit(), id(), keywords(), language(), main_category(), sub_category(), timezone(), title(), website_address()

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
