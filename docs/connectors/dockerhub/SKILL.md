---
name: pm-dockerhub
description: Docker Hub connector knowledge and safe action guide.
---

# pm-dockerhub

## Purpose

Reads public Docker Hub repositories and image tags for a configured username or organization via the Docker Hub registry API.

## Icon

- id: dockerhub
- asset: icons/dockerhub.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.docker.com/docker-hub/api/latest/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- base_url
- docker_username (required)
- page_size
- repository
- tag

## ETL Streams

- repositories:
  - primary key: name
  - cursor: last_updated
  - fields: date_registered(string), description(string), is_private(boolean), last_modified(string), last_updated(string), name(string), namespace(string), pull_count(integer), repository_type(string), star_count(integer), status(integer), status_description(string), storage_size(integer)
- tags:
  - primary key: id
  - cursor: last_updated
  - fields: content_type(string), digest(string), full_size(integer), id(integer), last_pushed(string), last_updated(string), last_updater_username(string), media_type(string), name(string), repository(integer), tag_status(string)
- repository_detail:
  - primary key: name
  - fields: collaborator_count(integer), date_registered(string), description(string), full_description(string), has_starred(boolean), hub_user(string), is_automated(boolean), is_private(boolean), last_updated(string), name(string), namespace(string), pull_count(integer), repository_type(string), star_count(integer), status(integer), status_description(string), storage_size(integer)
- tag_detail:
  - primary key: id
  - fields: creator(integer), full_size(integer), id(integer), last_updated(string), last_updater(integer), last_updater_username(string), name(string), repository(integer), status(string), tag_last_pulled(string), tag_last_pushed(string), v2(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Docker Hub API read of public repository and tag data
- approval: none; read-only, no obviously-safe reverse-ETL writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Docker Hub's declared streams and reverse-ETL actions.
- Usage: pm dockerhub <command> [flags]
- Read streams
- Other Commands
  - repositories list - Run the repositories ETL stream [intent=etl availability=implemented stream=repositories]
  - repository detail list - Run the repository detail ETL stream [intent=etl availability=implemented stream=repository_detail]
  - tag detail list - Run the tag detail ETL stream [intent=etl availability=implemented stream=tag_detail]
  - tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]

## Commands

### Inspect as a manual

```bash
pm connectors inspect dockerhub
```

### Inspect as structured JSON

```bash
pm connectors inspect dockerhub --json
```

## Agent Rules

- Run pm connectors inspect dockerhub before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
