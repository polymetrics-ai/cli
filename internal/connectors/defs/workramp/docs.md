# Overview

WorkRamp reads and writes users and groups, and reads guides, resources, and SCORM courses,
through the real WorkRamp Employee Learning Cloud (ELC) API
(`https://app.workramp.com/api/v1/...`, documented at
[developers.workramp.com](https://developers.workramp.com/) — a private API requiring WorkRamp
support to enable). This bundle was originally migrated from `internal/connectors/workramp` (the
hand-written connector); this Pass B pass researched the real, currently-documented ELC API surface
(via `developers.workramp.com/llms.txt`'s full endpoint index) and corrected the bundle's host/
paths/response-envelope shapes to match it, while keeping the same `users`/`groups`/`courses`
legacy-parity resource names (courses now reads WorkRamp's real "Guides" resource) and adding two
new streams (`resources`, `scorm_courses`) plus writes. `capabilities.write` is now `true`.

## Auth setup

Provide a WorkRamp API key via the `api_key` secret (generated on the Integrations -> API page,
tied to a specific admin user — actions taken via the key are attributed to that user); it is sent
as a Bearer token on every request (`mode: bearer`). `base_url` defaults to
`https://app.workramp.com` — **this corrects legacy's `api.workramp.com`, which does not match the
real documented host** (developers.workramp.com's Getting Started guide names `app.workramp.com`,
with `app.eu.workramp.com` for EU customers).

## Streams notes

`users` (`GET /api/v1/users`) returns a **bare JSON array** at the response root (`records.path:
""`), not a `{"data": [...]}` envelope as legacy assumed — corrected here. Pagination is
`page_number` with `page`/`limit` params, and the real API's `page` parameter is **0-indexed**
(`start_page: 0`, per the S4 engine mini-wave's `PaginationSpec.StartPage *int` 0-vs-unset
distinction), matching the API's own docs: "page (zero indexed) of results to return." Real fields:
`id` (integer primary key), `name`, `email`, `isAdmin`, `isDeleted`, `isPermanentlyDeleted`,
`createdAt`/`updatedAt` (Unix-millisecond numbers, declared `"number"` typed, not stringified).
`updatedAt` is declared `x-cursor-field` for manifest parity (no `incremental` block — see Known
limits).

`groups` (`GET /api/v1/groups`) is **not paginated at all** in the real API (no `page`/`limit`
query parameters are documented) and returns a bare array; this stream declares a stream-level
`"pagination": {"type": "none"}` override. Real fields: `id` (integer), `enterpriseId`, `name`,
`description`, `createdAt`/`updatedAt`.

`courses` now reads the real **Guides** resource (`GET /api/v1/guides`) instead of legacy's
fictional `/v1/courses` path (WorkRamp's core "course" primitive is called a Guide). The real
response is a nested envelope (`{"page", "per_page", "has_more", "data": {"guides": [...]}}`,
`records.path: "data.guides"`), paginated with 1-indexed `page`/`per_page` (default page size 20,
per the endpoint's own `per_page` default). A `has_more: false` response correlates with the
paginator's own short-page stop signal in every observed shape, so no additional `stop_path` wiring
was needed. Real fields: `id` (string UUID), `title`, `num_total_tasks`, `num_total_test_questions`,
`created_at`/`updated_at` (Unix-millisecond numbers), `tags` (array of strings). Note: this
endpoint also supports an optional `legacy_mode=true` query flag (defaulting to `true` for accounts
provisioned before 2025-11-11) that bypasses pagination and returns a flat array instead of the
nested envelope; this bundle targets the standard paginated response shape, matching the API's own
recommendation to migrate off `legacy_mode`.

`resources` (`GET /api/v1/resources`, new stream) is unpaginated, with a distinct flat envelope
(`{"status_code", "resources": [...]}`, `records.path: "resources"`). Real fields: `id` (string
UUID), `name`, `description`, `createdAt`/`updatedAt`.

`scorm_courses` (`GET /api/v1/scorms`, new stream) shares `courses`' nested-envelope/pagination
shape (`records.path: "data.scorms"`, 1-indexed `page`/`per_page`). Real fields: `id` (string
UUID), `title`, `created_at` (Unix-millisecond number), `time_estimate` (integer minutes).

All 5 streams declare `"projection": "passthrough"`, matching legacy's verbatim-emit behavior (no
`mapRecord` field-building anywhere in `workramp.go`) and the post-wave2 §8 rule 1 — every real
WorkRamp field beyond each stream's schema-declared subset survives to the emitted record.

## Write actions & risks

`capabilities.write` is now `true` (previously `false`). Five actions, all Bearer-authenticated:

- `create_user` (`POST /api/v1/users`, requires `email`) / `update_user` (`POST
  /api/v1/users/{user_id}`) / `delete_user` (`DELETE /api/v1/users/{user_id}`) — creates, updates,
  or permanently deletes a WorkRamp user account. Note the real API uses `POST`, not `PUT` or
  `PATCH`, for updates (`update-user-1`'s documented `operationId`/method).
- `create_group` (`POST /api/v1/groups`, requires `name`) / `update_group` (`POST
  /api/v1/groups/{group_id}`) — creates or updates a WorkRamp group. There is no documented
  `DELETE /api/v1/groups/{group_id}` endpoint in the real API (per-endpoint research found no
  group-delete operation), so no `delete_group` action is declared.

All five risk-annotated as external mutations requiring approval (`delete_user` further flagged as
a permanent-deletion risk). `delete_user` uses `body_type: "none"` (pure path-parameterized
DELETE); the rest use `body_type: "json"`.

## Known limits

- **Legacy's host and stream shapes did not match the real API.** Legacy's `api.workramp.com` base
  URL, `{"data": [...]}` envelope assumption for `users`/`groups`, and fictional `/v1/courses` path
  do not correspond to the real, currently-documented ELC API (real host `app.workramp.com`, bare
  arrays for `users`/`groups`, and the real resource is named Guides at `/api/v1/guides`). This
  Pass B pass corrects the bundle to the real, working surface; this is a genuine bug-fix, not a
  parity-narrowing deviation, since legacy's original assumptions would fail against the real live
  API.
- **No `incremental` block is modeled for any stream.** `updatedAt`/`updated_at` fields are present
  in raw responses and declared as `x-cursor-field` for manifest-surface parity with legacy's own
  `cursorFields`, but no documented endpoint in this bundle's scope exposes a server-side
  updated-since/date-range query filter (unlike WorkflowMax's `updatedSince`) — every read is a
  full stream read, matching legacy's own always-full-read behavior exactly.
- **`page_size`/`max_pages` config-driven per-request overrides are not modeled** — the engine's
  `page_number` paginator reads its page size from the static `streams.json` block only.
- **No `delete_group` write action** — the real API documents no group-delete endpoint (see Write
  actions & risks above).
- The remaining real ELC surface (custom user attributes, per-user/per-group assignment endpoints,
  paths/Training-Series, item-folders, libraries, challenges, live-training events, webhook
  subscriptions, white-label-customer enumeration) and the entirely separate Academy
  (customer/partner-facing) product surface are out of scope for this migration; see
  `api_surface.json`'s `excluded` entries for the full per-endpoint accounting.
