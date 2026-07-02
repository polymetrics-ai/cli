# Overview

Cal.com is a wave2 fan-out declarative-HTTP migration. It reads Cal.com bookings, availability
schedules, and the authenticated user's profile through the Cal.com v2 REST API
(`GET https://api.cal.com/v2/...`). This bundle targets capability parity with
`internal/connectors/cal-com` (the hand-written connector it migrates, Go package `calcom`); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Cal.com API key via the `api_key` secret; it is sent as a Bearer token (`Authorization:
Bearer <api_key>`), matching legacy's `connsdk.Bearer(token)` (`cal_com.go:282`), and is never
logged. Every request also sends a `cal-api-version` header, resolved from the `api_version` config
value (default `2024-08-13`, matching legacy's `defaultAPIVersion` constant); `spec.json`'s
`"default"` materializes this value into `RuntimeConfig.Config` before header resolution runs, so
the header is always present exactly like legacy's own `apiVersion()` fallback. `base_url` defaults
to `https://api.cal.com` and may be overridden for tests/proxies.

## Streams notes

`bookings` (`GET /v2/bookings`) and `schedules` (`GET /v2/schedules`) both use Cal.com's offset
(`skip`/`take`) pagination (`pagination.type: offset_limit`, `limit_param: take`, `offset_param:
skip`) — records live at `data`, and a page shorter than `take` stops pagination, matching legacy's
`harvest` exactly (`cal_com.go:145-167`); see Known limits for why `page_size` is fixed small (`1`)
rather than at legacy's default (100). `my_profile` (`GET /v2/me`)
is not paginated (`pagination.type: none`, matching legacy's `endpoint.paginated == false` branch);
its `data` envelope is a single object rather than an array, which `records.path: "data"` handles
identically to an array of one (the engine's `RecordsAt` treats a single JSON object at the resolved
path as a one-record page).

None of the three migrated streams expose an incremental cursor field in Cal.com's v2 API, matching
legacy (`calcomStreams()` declares no `CursorFields` for any of `bookings`/`schedules`/`my_profile`)
— this bundle declares no `incremental` block for any of them, so reads are full refresh.

## Write actions & risks

None. Cal.com is exposed read-only, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`; `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`event_types` is BLOCKED (ENGINE_GAP), not migrated in this wave.** Legacy's `event_types`
  stream (`GET /v2/event-types`) flattens a two-level nested envelope:
  `data.eventTypeGroups[].eventTypes[]` (`cal_com.go:196-222`'s `emitNested` — each group in the
  top-level array holds its own nested `eventTypes` array, and every nested event type becomes one
  record). The declarative dialect's `records.path` (`connsdk.RecordsAt`) only ever extracts a
  SINGLE JSON path as a flat array (or one object) — it has no mechanism to walk into a nested
  per-element sub-array and flatten it one level further. Expressing this correctly (not
  approximately — e.g. NOT flattening would silently drop every event type nested under every
  group) requires either a `RecordHook`/`StreamHook` (Tier 2) or an engine dialect extension for
  nested-array flattening (Tier-1 ENGINE_GAP). Per `docs/migration/conventions.md` §6, this wave
  reports `event_types` as a typed `ENGINE_GAP` blocker rather than fudging a wrong flattening or
  silently dropping the stream from the catalog without documentation; the other 3 streams
  (`bookings`, `schedules`, `my_profile`) are migrated at full parity. See `api_surface.json`'s
  `excluded` entry for `/v2/event-types`.
- **`page_size`/`max_pages` are not runtime-configurable, and `page_size` is fixed small (1) rather
  than at legacy's default (100).** Legacy exposes both as config overrides
  (`pageSizeFromConfig`/`maxPagesFromConfig`, `cal_com.go:327-355`). The engine's `offset_limit`
  paginator's `PageSize` is a static bundle-authored int (not templated), and there is no
  `MaxPages`-equivalent config-driven knob either; `max_pages` is unbounded (matching legacy's own
  `max_pages=0`/`all`/`unlimited` default). `page_size` is set to `1` (not legacy's default of 100)
  specifically so the mandatory 2-page conformance fixtures (`fixtures/streams/{bookings,schedules}/
  {page_1,page_2}.json`) are realistic to author and honestly exercise the short-page stop rule
  (`conformance`'s `pagination_terminates` check requires the replay server to serve exactly one
  request per fixture page — a `page_size` of 100 against a small hand-authored fixture would
  short-circuit after page 1 and never touch page 2 at all), matching bamboo-hr's identical
  documented precedent (`docs/migration/conventions.md`, bamboo-hr's `docs.md`). This changes the
  real per-page record count from legacy's 100 to 1 — a REST-shape difference (more, smaller
  requests), never a data-emission difference (every booking/schedule is still read exactly once,
  across more pages).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps extra fields (`connector`, `fixture`) onto every
  fixture-mode record (`cal_com.go:226-262`); none are part of the live record shape. This bundle's
  schemas and fixtures target the live path only.
