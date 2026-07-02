# Overview

Nylas is a wave2 fan-out declarative-HTTP migration. It reads Nylas calendars, contacts,
messages, and events for a connected grant through the Nylas v3 REST API
(`GET https://api.us.nylas.com/v3/grants/{grant_id}/<resource>`). This bundle targets capability
parity with `internal/connectors/nylas` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Nylas v3 API key via the `api_key` secret; it is sent only as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(secret)`.
`grant_id` defaults to `"me"` (the connected grant), matching legacy's `nylasDefaultGrantID`.

## Streams notes

All 4 streams (`calendars`, `contacts`, `messages`, `events`) are grant-scoped list endpoints
sharing the identical shape: `GET /v3/grants/{{ config.grant_id }}/<resource>`, records at `data`,
primary key `["id"]`. Pagination is Nylas's `next_cursor`/`page_token` body-cursor convention
(`pagination.type: cursor` with `token_path: next_cursor`, `cursor_param: page_token`) — no
`stop_path` is declared because legacy stops purely on an absent/empty `next_cursor`, with no
separate boolean stop signal to check. `limit` is sent via `{{ config.page_size }}` (default `50`,
matching legacy's `nylasDefaultPageSize`). The `events` stream additionally requires a
`calendar_id` config value, sent as a query param; matching legacy, an unset `calendar_id` hard-errors
that stream's read (the plain-string query template's absent-key behavior), while the other 3
streams are unaffected since they never reference `calendar_id`.

`messages` and `events` carry cursor fields (`date`, `updated_at` respectively) for catalog
purposes, matching legacy's declared `CursorFields` — but, matching legacy exactly, no
`incremental` block is declared for any stream: legacy's `Read` never applies a server-side date
filter (there is no `request_param`/`param_format` wiring in `nylas.go`), so adding one here would
be new, behavior-changing filtering legacy never performed. Every sync is a full read of every
Nylas v3 page, exactly as legacy behaves today.

## Write actions & risks

None. Nylas is a read-only source connector (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`api_server` (`eu`/default) config key dropped; `base_url` is now the sole region selector.**
  Legacy derives the API host from an `api_server` config value (`"eu"` selects
  `https://api.eu.nylas.com`, anything else defaults to `https://api.us.nylas.com`,
  `nylasBaseURL`, `nylas.go:267-288`). The engine's spec-default materialization (gap-loop cycle-1
  item 6/C3) only fills in a literal per-key default — it cannot express "derive `base_url` from
  `api_server`", the same cross-key-derivation class sentry's `hostname` and chargebee's `site` hit
  (see `docs/migration/conventions.md`'s derived-default guidance). `base_url` defaults to the US
  host (`https://api.us.nylas.com`, matching legacy's own default when `api_server` is unset) and an
  EU-region operator must now set `base_url` to `https://api.eu.nylas.com` directly instead of
  setting `api_server: eu`. This is a documented config-surface narrowing (every legacy-accepted
  `api_server` value has an operator-reachable `base_url` equivalent; no request/data change once
  configured), not a data-shape regression.
- Full Nylas v3 surface (threads, drafts, folders, webhooks, scheduling, notetaker, send) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 4 legacy-parity read streams are implemented.
- `max_pages` (`spec.json`, default `"0"`/unlimited) is wired into `pagination.max_pages`'s
  behavior via the engine's generic `MaxPages` hard-cap enforcement in `read.go`'s
  `readDeclarative` loop when set to a positive integer; matching legacy's `nylasMaxPages` fully
  (0/all/unlimited all mean unbounded on both sides).
