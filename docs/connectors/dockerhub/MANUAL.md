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
  docker_username
  page_size
  repository
  tag

ETL STREAMS
  repositories:
    primary key: name
    cursor: last_updated
    fields: date_registered(), description(), is_private(), last_modified(), last_updated(), name(), namespace(), pull_count(), repository_type(), star_count(), status(), status_description(), storage_size()
  tags:
    primary key: id
    cursor: last_updated
    fields: content_type(), digest(), full_size(), id(), last_pushed(), last_updated(), last_updater_username(), media_type(), name(), repository(), tag_status()
  repository_detail:
    primary key: name
    fields: collaborator_count(), date_registered(), description(), full_description(), has_starred(), hub_user(), is_automated(), is_private(), last_updated(), name(), namespace(), pull_count(), repository_type(), star_count(), status(), status_description(), storage_size()
  tag_detail:
    primary key: id
    fields: creator(), full_size(), id(), last_updated(), last_updater(), last_updater_username(), name(), repository(), status(), tag_last_pulled(), tag_last_pushed(), v2()

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
