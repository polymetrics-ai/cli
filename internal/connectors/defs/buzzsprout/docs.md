# Overview

Buzzsprout is a wave2 fan-out declarative-HTTP migration. It reads Buzzsprout podcasts and episodes
(titles, publish dates, durations, play counts) through the Buzzsprout REST API
(`GET https://www.buzzsprout.com/api/...`). This bundle targets capability parity with
`internal/connectors/buzzsprout` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Buzzsprout API key via the `api_key` secret; it is sent as `Authorization: Token
token=<api_key>` (`auth.mode: api_key_header`, `header: Authorization`, `prefix: "Token token="`),
matching legacy's `connsdk.APIKeyHeader("Authorization", secret, "Token token=")`
(`buzzsprout.go:237`) exactly, and is never logged. `base_url` defaults to
`https://www.buzzsprout.com` and may be overridden for tests/proxies.

## Streams notes

`episodes` (`GET /api/{{ config.podcast_id }}/episodes.json`) requires the `podcast_id` config value,
urlencoded into the path by `InterpolatePath`'s per-segment default (matching legacy's own
`endpointPath`/`isSafePathSegment` scoping); an absent `podcast_id` hard-errors on both sides (legacy:
`"buzzsprout connector requires config podcast_id for this stream"`; engine: an unresolved
`config.podcast_id` path-template key — same failure classification, different literal text, per
conventions.md §5's precedent for config-validation parity). `podcasts` (`GET /api/podcasts.json`) is
account-level and never needs `podcast_id`.

Both streams return a bare top-level JSON array (`records.path: ""`) with no server-side pagination
metadata; legacy's own `harvest` requests `page=1,2,...` and stops on a short/empty page
(`buzzsproutDefaultPageSize = 100`). This is exactly `pagination.type: page_number` with
`size_param: ""` (Buzzsprout never accepts a page-size query parameter — legacy never sends one
either) and `page_size: 100` as the short-page stop threshold, matching legacy's own default exactly
and matching stripe/searxng's precedent for a page-number API with no size param.

`episodes` declares `incremental.cursor_field: published_at` with no `request_param` and no
`client_filtered` — this matches legacy's real (lack of) filtering behavior exactly: legacy's
`InitialState` returns an empty cursor and `harvest` never reads or applies it to any request or
in-process filter, so every read re-emits the full episode set regardless of a prior cursor. The
cursor field exists on the schema (`x-cursor-field`) purely to publish the field a downstream
consumer could sort/dedupe on, matching legacy's own `CursorFields: []string{"published_at"}`
declaration that is likewise never consumed by `harvest`.

## Write actions & risks

None. Buzzsprout is exposed read-only, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`; `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`page_size` is not runtime-configurable.** Legacy exposes `page_size` as a config override
  (`buzzsproutPageSize`, `buzzsprout.go:289-302`, defaulting to 100, capped at 1000) used purely as
  the client-side short-page stop threshold (Buzzsprout's API accepts no page-size query parameter
  on either side). The engine's `page_number` paginator's `PageSize` field is a static bundle-authored
  int, not a templated value, so it cannot read a runtime `config.page_size` override; this bundle
  fixes it at a single JSON value in `streams.json`'s base pagination block, set to `100` to match
  legacy's own default exactly (the conformance fixture for `episodes` is a single page of 2 records —
  a short page relative to `page_size: 100` — so `pagination_terminates` observes exactly one request,
  matching the real one-request-in-production behavior; `podcasts` is likewise a single fixture page).
  This is a REST-shape parity fix, not a data-emission difference: every episode/podcast is still read
  exactly once.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps extra fields (`connector`, `fixture`, `previous_cursor`)
  onto every fixture-mode record (`buzzsprout.go:187-220`); none are part of the live record shape.
  This bundle's schemas and fixtures target the live path only — the engine's own conformance
  fixture-replay harness supersedes legacy's fixture-mode affordance.
- Full Buzzsprout API surface (single-episode fetch, episode create/update/delete) is out of scope;
  legacy itself never implemented writes or single-object reads. See `api_surface.json`.
