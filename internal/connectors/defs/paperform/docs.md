# Overview

Paperform reads Paperform forms and form submissions through the Paperform REST API. This bundle
migrates `internal/connectors/paperform` (the hand-written legacy connector) to a declarative
Tier-1 defs bundle; the legacy package stays registered and unchanged until the wave6 registry
flip.

## Auth setup

Provide a Paperform API key via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

Two streams: `forms` (`GET /forms`) and `submissions` (`GET /forms/{form_id}/submissions`, scoped
to the `form_id` config value — required for that stream, matching legacy's
`paperformResource`/`form_id`-substitution error). Both share the same shape: records at
`results`, primary key `["id"]`, incremental cursor field `created_at`. Pagination is `page_number`
(`page`/`limit` query params, `start_page: 1`) matching legacy's `harvestPages`/`limit`+`page` loop;
the engine's `page_number` paginator stops purely on a short/empty page (`recordCount < page_size`),
which is functionally equivalent to legacy's own `hasMore != "true" || len(records) == 0` stop rule
for every real response shape (a `has_more: false` final page is also Paperform's final, shorter-or-
equal page in every observed response).

Legacy's `limit`/`max_pages` runtime config overrides are not exposed as `spec.json` properties:
`page_size`/`max_pages` on a `pagination` block are static integers, not `{{ }}`-templatable in
this engine version, so a declared-but-unwireable `spec.json` property would be dead config (F6,
`docs/migration/conventions.md` §3) — this bundle relies on the engine's `page_number` short-page
stop and leaves `max_pages` unbounded (matching legacy's `0`/unlimited default) rather than a
config-driven override. `streams.json`'s `page_size: 100` matches legacy's real default
(`paperformDefaultLimit`) exactly, per campaign-monitor's identical precedent (`page_size: 100` in
its own `base.pagination` block) — `page_size` is purely a client-side chunk/stop-detection size
with no bearing on which records are returned (only how many round trips fetch them), so this does
not change any emitted record data.

## Write actions & risks

None. This connector is read-only, matching legacy (`Capabilities.Write: false`,
`Write` returns `ErrUnsupportedOperation`).

## Known limits

- Only `forms` and `submissions` are implemented, matching legacy's exact stream set. Per-form
  metadata (individual form field definitions), single-submission lookups, and submission deletion
  are out of scope for this wave; see `api_surface.json`'s `excluded` entries.
- `submissions` requires the `form_id` config value; there is no cross-form submissions listing
  endpoint in the Paperform API, matching legacy's own `form_id`-required behavior exactly (legacy
  errors with the same "requires config form_id" condition when unset).
- **`page_size` (100, matching legacy's default) is not runtime-configurable.** Legacy exposes
  `limit`/`max_pages` as config overrides (`pageSize`/`maxPages` in `paperform.go`); the engine's
  `page_number` paginator's `page_size` is a static bundle-authored int, not templated, so it
  cannot vary per connection the way legacy's config-driven `limit` did — `page_size: 100` is fixed
  at legacy's own default rather than a config-driven override, following campaign-monitor's
  identical documented scope-narrowing precedent. `fixtures/streams/forms/page_1.json` ships a
  genuinely full 100-record page (`has_more: true`) and `page_2.json` a short, 1-record final page
  (`has_more: false`) — the page-1-full-page fixture pattern required to exercise a real second
  request under a `page_number`/`offset_limit` paginator's short-page stop rule, rather than
  shrinking the live `page_size` to fit a hand-authored fixture.
