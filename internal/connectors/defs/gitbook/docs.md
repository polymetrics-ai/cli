# Overview

GitBook is a Tier-1 declarative HTTP connector for GitBook API v1. This Pass B bundle was expanded against GitBook's official OpenAPI spec at https://api.gitbook.com/openapi.json and API reference at https://gitbook.com/docs/developers/gitbook-api/api-reference. It covers 185 read streams and 170 write actions while keeping the four legacy stream names for compatibility: `users`, `organizations`, `org_members`, and `content`.

## Auth setup

Provide a GitBook personal access token via the `access_token` secret. The engine sends it as `Authorization: Bearer <token>`; the value is marked `x-secret` in `spec.json` and is not logged or copied into fixtures. `base_url` defaults to `https://api.gitbook.com/v1`.

Most Pass B streams are scoped by GitBook entity identifiers such as `organization_id`, `space_id`, `site_id`, `page_id`, `collection_id`, and related IDs. These identifiers are optional globally but required at read time for the stream whose path interpolates them.

## Streams notes

The legacy streams retain their original schema projection and record shaping: `users` flattens `displayName`/`photoURL`, `organizations` flattens `createdAt` and `urls.location`, `org_members` flattens nested `user` fields, and `content` reads page records from `pages`.

New Pass B streams use the OpenAPI operation ID as the stream name in snake_case and `projection: passthrough` with a minimal permissive schema. List responses read from the documented array envelope such as `items` or `pages`; detail responses read the JSON object root. Cursor pagination follows GitBook's `next.page` convention where present and terminates on fixtures without a next token. GitBook does not document stable incremental cursors for these REST resources, so all streams are full-refresh.

## Write actions & risks

This bundle declares 170 write actions for documented JSON/no-body GitBook mutations. Action names are the OpenAPI operation IDs converted to snake_case. Path identifiers are supplied as record fields, JSON request bodies are built from non-path record fields, and DELETE actions are marked destructive with idempotent `404` handling.

Writes can create, update, publish, archive, delete, import, export, invite, change permissions, merge change requests, manage integrations/sites/spaces/content, and trigger other GitBook workflows. Reverse ETL callers must use plan, preview, approval, execute; dry-run previews resolve request method and path without exposing secrets.

## Known limits

- Three documented GET endpoints return `text/event-stream` and are excluded because the declarative reader extracts JSON records only.
- `PUT /orgs/{organizationId}/sites/{siteId}/context-records` is excluded because its request body is a root JSON array, while the current write engine builds object bodies from record fields.
- New Pass B schemas are intentionally permissive passthrough schemas until dedicated per-resource projections are reviewed. The four legacy streams keep strict legacy-compatible projections.
- Generated write schemas type path identifiers and require documented required fields where known, but leave nested JSON body fields permissive. The declarative engine forwards non-path record fields as the request body; stricter nested request-body validation can be added later without changing request construction.
- Optional query filters are not all modeled; required GET query parameters are represented in `spec.json`, while optional filters can be added later without changing endpoint coverage.
