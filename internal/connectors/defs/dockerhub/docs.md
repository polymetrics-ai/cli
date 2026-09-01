# Overview

Reads public Docker Hub repositories and image tags for a configured username or organization via
the Docker Hub registry API.

Readable streams: `repositories`, `tags`, `repository_detail`, `tag_detail`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.docker.com/docker-hub/api/latest/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://hub.docker.com/v2`; format `uri`; Docker Hub
  registry API base URL override for tests or self-hosted proxies.
- `docker_username` (required, string); Docker Hub username or organization namespace whose
  repositories and tags to read. Lowercase alphanumerics, underscores, and hyphens only.
- `page_size` (optional, integer); default `100`; Page size (1-100) for the initial request of each
  paginated stream (Docker Hub's page_size query param); subsequent pages follow the API's own
  absolute next URL verbatim.
- `repository` (optional, string); Repository name (without the namespace prefix) the
  'tags'/'repository_detail'/'tag_detail' streams are scoped to. Required only when reading one of
  those streams.
- `tag` (optional, string); Tag name the 'tag_detail' stream reads a single tag record for (e.g.
  'latest'). Required only when reading the 'tag_detail' stream.

Default configuration values: `base_url=https://hub.docker.com/v2`, `page_size=100`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/namespaces/{{ config.docker_username }}/repositories`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next`; next URLs stay
on the configured API host.

Pagination by stream: next_url: `repositories`, `tags`; none: `repository_detail`, `tag_detail`.

- `repositories`: GET `/namespaces/{{ config.docker_username }}/repositories` - records path
  `results`; query `page`=`1`; `page_size`=`{{ config.page_size }}`; follows a next-page URL from
  the response body; URL path `next`; next URLs stay on the configured API host.
- `tags`: GET `/namespaces/{{ config.docker_username }}/repositories/{{ config.repository }}/tags`
  - records path `results`; query `page`=`1`; `page_size`=`{{ config.page_size }}`; follows a
  next-page URL from the response body; URL path `next`; next URLs stay on the configured API host.
- `repository_detail`: GET `/namespaces/{{ config.docker_username }}/repositories/{{ config.repository }}` - single-object response; records at response root.
- `tag_detail`: GET `/namespaces/{{ config.docker_username }}/repositories/{{ config.repository }}/tags/{{ config.tag }}` - single-object response; records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external Docker Hub API read of public repository and
tag data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other cited artifact endpoints are explicitly classified in `execution bundle`; no undocumented
  legacy endpoint is exposed by this bundle.
