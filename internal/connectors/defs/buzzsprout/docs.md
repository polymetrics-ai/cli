# Overview

Buzzsprout is a wave2 fan-out declarative-HTTP migration, expanded to full documented API-surface
coverage in Pass B. It reads Buzzsprout podcasts and episodes (titles, publish dates, durations, play
counts) and creates/updates episodes through the Buzzsprout REST API
(`https://www.buzzsprout.com/api/...`). This bundle targets capability parity with
`internal/connectors/buzzsprout` (the hand-written connector it migrates, which was read-only); the
legacy package stays registered and unchanged until wave6's registry flip. Buzzsprout's entire
published API (github.com/Buzzsprout/buzzsprout-api, `sections/episodes.md` +
`sections/podcasts.md`) is exactly 5 endpoints across 2 resources — all 5 are covered by this bundle
(2 streams, 2 writes, 1 documented exclusion); see `api_surface.json`.

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

This bundle adds write support beyond legacy (which was read-only, `Write` returning
`connectors.ErrUnsupportedOperation`); `capabilities.write` is now `true` and `writes.json` declares
2 actions, both scoped to the configured `podcast_id` exactly like the `episodes` read stream:

- **`create_episode`** (`POST /api/{{ config.podcast_id }}/episodes.json`, `body_type: json`,
  `minProperties: 1`): creates a new episode. Per Buzzsprout's own docs, this can trigger audio
  processing and an email notification (`email_user_after_audio_processed`, defaults `true` on
  Buzzsprout's side when omitted) once the file finishes processing, and — depending on
  `published_at`/`private` — publish the episode live to the podcast's public feed. **Risk: external
  mutation with real-world publication side effects; approval required.**
- **`update_episode`** (`PUT /api/{{ config.podcast_id }}/episodes/{{ record.id }}.json`,
  `path_fields: ["id"]`, `body_type: json`): updates an existing episode's metadata (title,
  description, `private`, `explicit`, schedule fields, etc.) in place. **Risk: external mutation;
  can change public visibility (`private`) or publish scheduling (`published_at`) of a live episode;
  approval required.**

Buzzsprout documents no delete endpoint for episodes or podcasts at all (verified against the
current published API reference) — there is nothing to exclude as `destructive_admin` here, unlike
the previous (incorrect) draft of this bundle's `api_surface.json`, which listed a DELETE endpoint
that does not actually exist in Buzzsprout's API.

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
- **Single-episode detail fetch (`GET /api/{podcast_id}/episodes/{episode_id}.json`) is excluded as
  `duplicate_of`**: it returns the identical episode shape the `episodes` list stream already emits,
  and Buzzsprout has no per-episode incremental/webhook signal that would make a targeted
  single-record read materially useful over a full/incremental list sync. See `api_surface.json`.
- `create_episode`'s `audio_file`/`artwork_file` multipart-attachment upload path (Buzzsprout's
  alternative to the `audio_url`/`artwork_url` string fields) is not modeled — the declarative
  `body_type` dialect supports `json`/`form`/`none` bodies, not multipart file attachments; only the
  URL-reference variant of episode creation is exposed. A multipart body type is a Tier-2/engine
  extension, not attempted here since `audio_url`/`artwork_url` cover the common (already
  publicly-hosted-file) case.
