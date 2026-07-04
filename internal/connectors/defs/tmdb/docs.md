# Overview

TMDb reads catalog, search, trending, account-state, and reference metadata from The Movie Database
API v3. The four legacy streams (`popular_movies`, `now_playing_movies`, `search_movies`, and
`movie_details`) keep the hand-written connector's passthrough record behavior.

Pass B reviewed TMDb's official ReadMe OpenAPI 3.1 document
(`https://developer.themoviedb.org/openapi/tmdb-api.json`) on 2026-07-04. The document contains
152 operations: 135 GET endpoints, 12 POST endpoints, and 5 DELETE endpoints. This bundle covers
the GET data surface with streams, except authentication/token probes and a scalar reference-list
endpoint. POST/DELETE endpoints remain excluded because they mutate user session, account, list, or
rating state and the connector remains read-only.

## Auth setup

Provide a TMDb v3 API key via the `api_key` secret; it is sent as the `api_key` query parameter,
matching legacy's `baseQuery`. Optional `language` is sent where supported.

Search streams require `query`. `find_by_id` requires `external_source`. Trending streams use
TMDb's documented `day` time window. Single-ID resources use comma-separated fan-out config keys
such as `movie_ids`, `series_ids`, `person_ids`, `collection_ids`, `company_ids`, `keyword_ids`,
`network_ids`, and `review_ids`. Multi-part TV season/episode endpoints use the documented path
config keys directly, such as `series_id`, `season_number`, and `episode_number`.

## Streams notes

The legacy streams stay first and preserve their existing request shapes and fixtures:
`popular_movies`, `now_playing_movies`, and `search_movies` read `results`; `movie_details` reads a
single object at the response root.

Generated Pass B streams use `projection: "passthrough"` to avoid truncating TMDb response fields.
Endpoints whose OpenAPI response schema has a `results` array emit one record per result. Detail
and subresource endpoints that return compound objects emit the response object as a single record,
preserving arrays such as credits, images, translations, providers, cast, and crew together instead
of guessing one preferred child array.

Config-key fan-out streams emit nothing when their ID config is empty. When IDs are supplied, the
engine runs the endpoint once per ID and stamps the parent ID onto each emitted record. These
fan-out streams carry conformance skip markers because the fixture replay harness does not inject
the comma-separated ID config.

## Write actions & risks

None. TMDb's documented mutations create/delete sessions, add/remove list items, mark favorites or
watchlist entries, and add/delete ratings. They require user session/account state and change remote
user data, so this bundle keeps `capabilities.write: false` and ships no `writes.json`.

## Known limits

- **Authentication/session GET endpoints are excluded.** API-key validation and guest/request token
  creation are operational auth endpoints, not syncable data streams.
- **One scalar reference endpoint is excluded.** `/3/configuration/primary_translations` returns
  translation-code scalars rather than objects; the records dialect does not fan out scalar arrays.
- **Trending streams model the `day` window.** TMDb also documents `week`, but the current bundle
  exposes one stable stream per trending endpoint rather than separate time-window fan-out.
- **Generated subresource streams keep compound responses intact.** For endpoints like credits,
  images, translations, and watch providers, a single response object may contain multiple arrays.
  Emitting the whole object avoids silently dropping sibling arrays.
- **Runtime `page_size`/`max_pages` overrides are not modeled.** Legacy accepts those config keys,
  but the declarative paginator uses bundle-authored values. TMDb's list endpoints use page-number
  pagination with fixed server-side page size.
- **Legacy fixture-only fields are not modeled.** Legacy's fixture path stamps `fixture: true`; the
  declarative fixture replay harness replaces that test-only behavior.
