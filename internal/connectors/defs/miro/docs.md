# Miro

## Overview

This bundle covers the Miro Developer Platform REST API from the official API reference and linked OpenAPI spec. It keeps the legacy board, board member, item, tag, and connector streams in schema-projection mode, and adds passthrough streams for the broader documented Platform, Enterprise, SCIM, user-group, project, and experimental API surfaces.

## Auth setup

Create a Miro app or token with the scopes required for the streams and write actions you intend to use. Store the token in the `api_key` secret; the connector sends it as bearer auth. `base_url` defaults to `https://api.miro.com`. Board-, organization-, team-, project-, case-, group-, member-, item-, and app-scoped streams require the corresponding ID config values named in `spec.json`.

## Streams notes

The bundle declares 84 streams. Legacy streams `boards`, `board_users`, `board_items`, `board_tags`, and `board_connectors` preserve the exact field projection emitted by the legacy Go connector. Newly added streams use passthrough projection against generated schemas derived from each documented response record shape.

Miro offset-list endpoints use `limit=50` and `offset` pagination where documented. Cursor-list endpoints use `limit=50` as a stream query and follow the response `cursor` token. This bundle currently has 6 offset-paginated streams and 16 cursor-paginated streams.

## Write actions & risks

The bundle declares 98 JSON/SCIM/bodyless write actions for documented POST, PUT, PATCH, and DELETE operations that the declarative dialect can express. Write actions are verb-prefixed and path-based, for example `create_boards`, `update_boards_board_id`, and `delete_boards_board_id`. Every write requires approval; DELETE actions are marked destructive and treat a documented-missing 404 as idempotent success.

## Known limits

OAuth token introspection and revocation endpoints are excluded as authentication lifecycle operations, not data sync resources. Deprecated v1 token revocation is excluded as deprecated. Multipart document/image/file upload operations are excluded because they require binary payload support. Mutations whose required inputs are query parameters are excluded because declarative writes currently model path fields and bodies, not write query strings. The bulk item creation endpoint with a top-level JSON array body is excluded because declarative write bodies are object-shaped records. The documented group-delete path with a literal trailing `?` is excluded because it cannot be represented unambiguously as an HTTP path in the declarative dialect. SCIM list streams are exposed as single-page reads because SCIM `startIndex`/`count` pagination does not map cleanly to the current page-number or offset paginator semantics.
