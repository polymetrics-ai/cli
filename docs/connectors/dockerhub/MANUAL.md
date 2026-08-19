# pm connectors inspect dockerhub

```text
NAME
  pm connectors inspect dockerhub - Docker Hub connector manual

SYNOPSIS
  pm connectors inspect dockerhub
  pm connectors inspect dockerhub --json
  pm credentials add <name> --connector dockerhub [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads public Docker Hub repositories and image tags for a configured username or organization via the Docker Hub registry API.

ICON
  id: dockerhub
  asset: icons/dockerhub.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.docker.com/docker-hub/api/latest/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  base_url
  docker_username (required)
  page_size
  repository
  tag

ETL STREAMS
  repositories:
    primary key: name
    cursor: last_updated
    fields: date_registered(string), description(string), is_private(boolean), last_modified(string), last_updated(string), name(string), namespace(string), pull_count(integer), repository_type(string), star_count(integer), status(integer), status_description(string), storage_size(integer)
  tags:
    primary key: id
    cursor: last_updated
    fields: content_type(string), digest(string), full_size(integer), id(integer), last_pushed(string), last_updated(string), last_updater_username(string), media_type(string), name(string), repository(integer), tag_status(string)
  repository_detail:
    primary key: name
    fields: collaborator_count(integer), date_registered(string), description(string), full_description(string), has_starred(boolean), hub_user(string), is_automated(boolean), is_private(boolean), last_updated(string), name(string), namespace(string), pull_count(integer), repository_type(string), star_count(integer), status(integer), status_description(string), storage_size(integer)
  tag_detail:
    primary key: id
    fields: creator(integer), full_size(integer), id(integer), last_updated(string), last_updater(integer), last_updater_username(string), name(string), repository(integer), status(string), tag_last_pulled(string), tag_last_pushed(string), v2(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Docker Hub API read of public repository and tag data
  approval: none; read-only, no obviously-safe reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Docker Hub's declared streams and reverse-ETL actions.
  Usage: pm dockerhub <command> [flags]
  Read streams
  Other Commands
    repositories list - Run the repositories ETL stream [intent=etl availability=implemented stream=repositories]
    repository detail list - Run the repository detail ETL stream [intent=etl availability=implemented stream=repository_detail]
    tag detail list - Run the tag detail ETL stream [intent=etl availability=implemented stream=tag_detail]
    tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect dockerhub

  # Inspect as structured JSON
  pm connectors inspect dockerhub --json

AGENT WORKFLOW
  - Run pm connectors inspect dockerhub before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
