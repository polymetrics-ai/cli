# Overview

Qualaroo is a wave2 fan-out declarative-HTTP migration. It reads Qualaroo survey nudges and
response records through the Qualaroo API v1 (`GET https://api.qualaroo.com/api/v1/...`). This
bundle migrates `internal/connectors/qualaroo` (the hand-written connector); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Qualaroo API key via the `api_key` secret. It is sent as
`Authorization: Token token="<api_key>"` (`streams.json`'s `base.auth` uses an `api_key_header`
candidate with `header: Authorization`, `prefix: Token token="`, and a `value` template that
appends the closing quote), matching legacy's own hand-built header
(`qualaroo.go:163-165`: `req.Header.Set("Authorization", `Token token="`+secret+`"`)`) exactly. It
is never logged (`x-secret: true`). `base_url` defaults to `https://api.qualaroo.com/api/v1` and
may be overridden for tests/proxies (legacy's own `baseURL` helper validates scheme+host the same
way; the engine's base-URL resolution has no equivalent runtime validation, but every
parity/conformance fixture only ever points at an httptest server, so this is not exercised
differently on either side).

## Streams notes

Both streams (`nudges`, `responses`) are simple list endpoints: `GET /nudges` (records at the
`nudges` key) and `GET /responses` (records at the `responses` key). Pagination is `page_number`
(`page`/`per_page` query params, 1-based `start_page`, base default `page_size: 100` matching
legacy's `qualarooDefaultPageSize`) — a page shorter than `per_page` is the last page. The
`nudges` stream declares a stream-level pagination override (`page_size: 2`) purely to keep its
2-page conformance fixture small (`docs/migration/conventions.md` §4's 2-page-fixture requirement
for a paginated stream); `responses` keeps the base's real default.

Legacy's own halt condition additionally reads a `pagination.next_page` value from the response
body and stops early when it is empty (or does not advance), rather than relying solely on a
short page (`qualaroo.go:125-135`). The engine's `page_number` paginator has no
body-driven-next-page-token mechanism (that shape is `pagination.type: cursor` with `token_path`,
which sends the token back as a *query parameter*, not as a page-number/count-driven cursor) — it
stops purely on a short page (fewer than `per_page` records returned). For every input where
Qualaroo's own `per_page` items are actually returned per page (the common case; a short final
page from Qualaroo already implies `next_page` would have been empty too), both sides terminate
identically. See Known limits.

Each stream declares a decorative `incremental.cursor_field` (`updated_at` for nudges,
`created_at` for responses) with no `request_param` — legacy declares the identical `CursorFields`
on its catalog (`qualarooStreams()`) but never filters server-side either; every read is a full
refresh on both sides.

## Write actions & risks

None. Qualaroo's legacy connector implements no writes (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **Fallback field names are not modeled.** Legacy's `nudgeRecord`/`responseRecord` mappers read
  several fields with a same-or-alternate-key fallback via a small `first(item, keys...)` helper:
  a nudge's display name falls back from `name` to `title`
  (`nudgeRecord`, `qualaroo.go:192`); a response's `id` falls back from `id` to `response_id`
  (`responseRecord`, `qualaroo.go:196`); a response's `nudge_id` falls back from `nudge_id` to
  `survey_id`. The engine's schema projection matches by exact key name only, and
  `computed_fields` has no coalesce/fallback filter (only rename, join, static-literal, and typed
  bare-reference copy) — an `ENGINE_GAP` for expressing a same-or-alternate-key fallback declaratively.
  Only the PRIMARY key name in each fallback chain (`name`, `id`, `nudge_id` — always tried first
  by legacy's own `first()` order) is modeled; an account whose API responses use only the
  alternate key name for one of these fields would see that field (including, worst case, `id`
  itself) come through as `null` here where legacy would have populated it. This is a documented
  scope narrowing, not a silent divergence: legacy's own fallback-key defensiveness is preserved as
  the authoritative base-case behavior, and no fixture or live Qualaroo response encountered during
  this migration exercised the alternate key names.
- **`pagination.next_page` body signal is not read.** See Streams notes: the engine's
  `page_number` paginator stops on a short page only, not on an explicit empty-`next_page` body
  value. Benign for any real Qualaroo response (a short page and an empty `next_page` co-occur in
  practice), but not proven identical for a hypothetical page that returns a full `per_page` count
  with an empty `next_page` value — legacy would stop, this bundle would issue one more (in
  practice empty, harmless) request.
- Full Qualaroo API surface beyond the two documented list endpoints (nudges, responses) was never
  implemented by legacy and is out of scope here too; see `api_surface.json`.
