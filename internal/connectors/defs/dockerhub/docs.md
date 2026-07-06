# Overview

Reads public Docker Hub repositories, image tags, and namespace profiles for a configured username
or organization via the Docker Hub registry API.

Readable streams: `repositories`, `tags`, `namespace`, `repository_detail`, `tag_detail`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.docker.com/docker-hub/api/latest/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://hub.docker.com/v2`; format `uri`; Docker Hub
  registry API base URL override for tests or self-hosted proxies.
- `docker_username` (required, string); Docker Hub username or organization namespace to read
  repositories, tags, and the namespace profile for. Lowercase alphanumerics, underscores, and
  hyphens only.
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

Connection checks call GET `/repositories/{{ config.docker_username }}/`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `next`; next URLs stay
on the configured API host.

Pagination by stream: next_url: `repositories`, `tags`; none: `namespace`, `repository_detail`,
`tag_detail`.

- `repositories`: GET `/repositories/{{ config.docker_username }}/` - records path `results`; query
  `page`=`1`; `page_size`=`{{ config.page_size }}`; follows a next-page URL from the response body;
  URL path `next`; next URLs stay on the configured API host.
- `tags`: GET `/repositories/{{ config.docker_username }}/{{ config.repository }}/tags` - records
  path `results`; query `page`=`1`; `page_size`=`{{ config.page_size }}`; follows a next-page URL
  from the response body; URL path `next`; next URLs stay on the configured API host.
- `namespace`: GET `/users/{{ config.docker_username }}/` - single-object response; records at
  response root.
- `repository_detail`: GET `/repositories/{{ config.docker_username }}/{{ config.repository }}/` -
  single-object response; records at response root.
- `tag_detail`: GET `/repositories/{{ config.docker_username }}/{{ config.repository }}/tags/{{
  config.tag }}/` - single-object response; records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external Docker Hub API read of public repository, tag,
and namespace data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=7, duplicate_of=2, non_data_endpoint=3, requires_elevated_scope=41.
