# Overview

PostHog is a wave2 declarative-HTTP migration. It reads PostHog events and persons for a project
through the modern PostHog REST API (`GET /api/projects/{project_id}/events/` and `/persons/`).
This bundle is a Tier-1 pure declarative port of `internal/connectors/posthog` (the hand-written
connector it migrates, itself a connsdk-HTTP-based read-only connector — plain Bearer auth,
`next`-URL pagination, no signature auth or async polling — so no Go escape hatch is needed); the
legacy package stays registered and unchanged until wave6's registry flip. The catalog's
`runtime_kind`/`type` metadata for this connector (`docs/migration/inventory.json`,
`internal/connectors/catalog_data.json`) already correctly labels it `source`/`api` — legacy's own
package doc confirms it is a plain read-only REST source, not a native or destination connector.

## Auth setup

Provide a PostHog personal API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(token)`
(`posthog.go:131`), and is never logged. `project_id` (required) scopes every request to
`/api/projects/{project_id}/...`. `base_url` defaults to `https://app.posthog.com` (PostHog Cloud)
and may be overridden for self-hosted instances or tests, matching legacy's own
`defaultBaseURL`/`baseURL` validation.

## Streams notes

- `events` — `GET /api/projects/{project_id}/events/`, records at `results`. `limit` (default
  `100`, matching legacy's own `intConfig(cfg, "page_size", 100)` default) and an optional `after`
  filter (sent only when `start_date` is set — legacy's own `if stream == "events" &&
  start_date != "" { q.Set("after", start_date) }`, `posthog.go:82-84`) are both templated via
  `stream.Query`'s optional-query dialect.
- `persons` — `GET /api/projects/{project_id}/persons/`, records at `results`. Same `limit`
  default; no `after` filter (legacy only applies it to `events`).

Pagination follows PostHog's own `next` absolute-URL convention (`pagination.type: next_url`,
`next_url_path: "next"`) — PostHog's real wire shape always emits a fully-qualified absolute URL
in `next` (confirmed against legacy's own `posthog_test.go:25`'s fixture, which serves
`srvURL(r) + "/api/projects/42/events/?page=2"`, and legacy's own `Read` loop, which follows `next`
verbatim with no query re-application once past the first page — `posthog.go:110-111` resets
`q = nil` after following `next`). The engine's `next_url` paginator's same-host SSRF guard
(THREAT-MODEL §3) passes cleanly since PostHog's `next` URL is always same-origin as `base_url` in
production.

Neither stream has a `incremental.cursor_field`/`request_param` block: legacy's own catalog
publishes `CursorFields: []string{"timestamp"}` for `events` (preserved here as `schemas/
events.json`'s `x-cursor-field`) but legacy's `Read` never actually applies an incremental
lower-bound filter beyond the one-time `start_date`-from-config `after` param on a fresh sync —
there is no persisted-cursor-driven repeat-sync filter in legacy's read path at all (`posthog.go`'s
`Read` reads `req.Config.Config["start_date"]` only, never `req.State`). Declaring an `incremental`
block here would be new, behavior-changing filtering legacy never performed; `x-cursor-field` is
kept in the schema for downstream sync-mode derivation (design §B.6: `incremental_append[_deduped]`
requires an `incremental` block to apply, so its absence here already yields the correct
"full-refresh-only" sync-mode set matching legacy).

## Write actions & risks

None. Legacy's `Write` returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`events_time_step`** (a config field the wider catalog records for this connector,
  `internal/connectors/catalog_data.json`) is not modeled: it belongs to the original Python/
  low-code source connector's long-running-sync chunking strategy, not this repo's legacy Go port
  (`internal/connectors/posthog/posthog.go`), which never reads it. Nothing in the migrated-from
  Go connector exercises this key, so nothing here needs to either — it was never wired
  end-to-end in the system this bundle actually migrates.
- Per `docs/migration/conventions.md`'s next_url fixture rule: `fixtures/streams/{events,persons}/`
  ship a single-page fixture with `next: null` (the replay server's own address is unknown until
  runtime, so a static fixture cannot embed a correct absolute second-page URL) — this still
  satisfies `pagination_terminates` (which only requires hits == fixture page count) and
  `read_fixture_nonempty`. Two-page `next_url` correctness is proven in this bundle's parity suite
  instead, which drives a real `httptest.Server` and asserts the second page is requested with the
  expected query.
- Full PostHog surface (insights, cohorts, feature flags, dashboards, actions, annotations, event
  capture) is out of scope for this wave; see `api_surface.json`'s
  `api_surface.json` concrete exclusion entries.
