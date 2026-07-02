# Overview

GitBook is a Tier-1 declarative-HTTP wave2 fan-out migration. It reads GitBook users,
organizations, organization members, and space content (pages) through the GitBook REST API v1.
This bundle migrates `internal/connectors/gitbook` (the hand-written connector); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a GitBook API access token via the `access_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <access_token>`) and is never logged. `base_url` defaults to
`https://api.gitbook.com/v1` and can be overridden for tests or proxies.

## Streams notes

- `users` — `GET /user`, a single-object (non-paginated) endpoint returning the authenticated
  GitBook user. `pagination: {"type": "none"}` overrides the base cursor pagination for this
  stream. `display_name`/`photo_url` are renamed from the raw API's camelCase `displayName`/
  `photoURL` via `computed_fields`.
- `organizations` — `GET /orgs`, a paginated list of organizations the authenticated user belongs
  to. Records live at `items`; `created_at` is renamed from the raw `createdAt` field via
  `computed_fields`. `url` is extracted from the raw API's nested `urls.location` field (GitBook
  returns `urls` as an object, e.g. `{"location": "https://app.gitbook.com/o/<id>"}`, not a bare
  string) so the output stays a schema-conformant `["string", "null"]` value rather than passing
  the whole nested object through.
- `org_members` — `GET /orgs/{{ config.organization_id }}/members`, members of the configured
  organization. `organization_id` is a required-at-read-time config value (interpolated into the
  path; an unset value hard-errors, matching legacy's explicit "config organization_id is required"
  check). Real GitBook member payloads nest the user under a `user` sub-object (`user.id`,
  `user.displayName`, `user.email`); `computed_fields` reaches into that nesting the same way
  legacy's `mapRecord` does. `role` is a flat top-level field on the member object and needs no
  rename (schema projection copies it verbatim).
- `content` — `GET /spaces/{{ config.space_id }}/content/pages`, the page tree of the configured
  space. `space_id` is a required-at-read-time config value. Records live at `pages`.

All 4 streams share the base's cursor pagination (`type: cursor`, `cursor_param: page`,
`token_path: next.page`, matching GitBook's `{"items":[...],"next":{"page":"<cursor>"}}` convention
and its `?page=<cursor>&limit=<n>` request shape); `users` overrides pagination to `none` since
`/user` returns a single object, not a list. GitBook exposes no incremental cursor for any of these
resources (legacy declares no `CursorFields`), so every stream here is full-refresh only, matching
legacy exactly.

## Write actions & risks

None. GitBook is a read-only source in this connector (legacy `Capabilities.Write` is `false`); no
`writes.json` file is present.

## Known limits

- Full GitBook API surface (space content editing/creation, integrations, webhooks, change
  requests) is out of scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope,
  reason: "Pass B capability expansion"}` entries. Only the 4 legacy-parity read streams are
  implemented.
- `organization_id`/`space_id` are required config values only when reading `org_members`/`content`
  respectively; the engine's path-interpolation hard-errors if either is absent when that stream is
  read, matching legacy's explicit `resolveResource` validation.
- GitBook's official developer documentation site (`https://developer.gitbook.com/`) renders its
  API reference client-side; concrete wire-shape details (e.g. the `org_members` nested `user`
  object) were sourced from the legacy connector's own tolerant mapping code and its inline comments
  ("real payloads" nest under `user`), which is authoritative ground truth per migration
  convention — not a documentation gap requiring a blocker.
