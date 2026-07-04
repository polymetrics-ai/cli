# Overview

Zoom is a wave2 fan-out declarative-HTTP migration of `internal/connectors/zoom` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads Zoom users, meetings, and webinars through the
Zoom REST API (`GET https://api.zoom.us/v2/...`). Read-only.

## Auth setup

Provide a Zoom OAuth access token via the `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`zoom.go:186`). `base_url` defaults to `https://api.zoom.us/v2` and may
be overridden for tests/proxies.

## Streams notes

`users` (`GET /users`, records at `users`) needs no additional config. `meetings` (`GET
/users/{{ config.user_id }}/meetings`, records at `meetings`) and `webinars` (`GET
/users/{{ config.user_id }}/webinars`, records at `webinars`) are scoped to one Zoom user: the
path template substitutes the `user_id` config value (urlencoded by `InterpolatePath`'s
per-segment default, matching legacy's own `url.PathEscape(userID)` in `resolvePath`); an absent
`user_id` hard-errors on both sides when either of those two streams is read (legacy: `"zoom
stream requires config user_id"`; engine: an unresolved `config.user_id` path-template key — same
failure classification, different literal text, per conventions.md §5's precedent for
config-validation parity). `user_id` is not declared in `spec.json`'s `required[]` since `users`
does not need it, matching legacy's own per-stream (not global) requirement.

Pagination follows Zoom's `next_page_token` cursor convention (`pagination.type: cursor`,
`cursor_param: next_page_token`, `token_path: next_page_token`, no `stop_path`): the next page's
token is read from the response body's own `next_page_token` field, and pagination stops when
that field is empty — identical to legacy's own `harvest` loop (`zoom.go:143-149`), which never
consulted any other stop signal. Page size is configurable via `page_size` (default `100`,
matching legacy's `defaultPageSize` fallback and validated 1-300 range), sent as the `page_size`
query param on every request (`stream.query.page_size`, referencing `{{ config.page_size }}`,
materialized from the spec default when unset) — because Zoom's pagination type is `cursor`
(not `page_number`), `page_size` here is an ordinary per-stream query template, not a paginator's
static field, so it stays genuinely config-overridable exactly like legacy.

All three streams use `projection: "passthrough"` because legacy's `mapRecord` copied every raw
wire field before filling missing derived fields. The legacy alternate-key chains are modeled with
`coalesce`: users fill `name` from `email`/`first_name`/`display_name` and `updated_at` from
`created_at`; meetings and webinars fill `id` from `uuid`, `name` from `topic`, and `updated_at`
from `start_time` only when the primary field is absent or null.

## Write actions & risks

None. Legacy `zoom` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` override (`0`/`all`/
  `unlimited` meaning unbounded, or a positive integer hard cap, `zoom.go:272-282`). The engine's
  `PaginationSpec.MaxPages` is a static bundle-level integer, not config-templated, so there is no
  mechanism to make it runtime-configurable. This bundle omits `max_pages` entirely, which is
  unbounded — legacy's own default when unset/`0`/`all`/`unlimited` (the common case) — so every
  input legacy itself defaults to behaves identically; only an operator who explicitly set a
  positive `max_pages` override loses that cap here.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`, a credential-free conformance-harness affordance)
  stamps an additional `previous_cursor` field (echoing `req.State["cursor"]`) onto emitted
  fixture-mode records. This is not part of the live record shape; this bundle's schemas and
  fixtures target the live path only. The engine's own conformance/fixture-replay harness
  provides the credential-free test affordance this bundle needs.
- The connector's own documented `docs_url` page renders dynamically and could not be fetched by
  automated tooling during this migration; legacy Go source (`internal/connectors/zoom/zoom.go`)
  is the ground truth this bundle was built from, per conventions.md's "legacy is ground truth
  over any doc" rule — this did not block the migration.
