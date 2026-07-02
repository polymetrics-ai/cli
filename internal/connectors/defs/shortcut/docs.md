# Overview

Shortcut is a wave2 fan-out declarative-HTTP migration. It reads Shortcut stories, epics,
projects, and iterations through the Shortcut REST API v3
(`GET https://api.app.shortcut.com/api/v3/...`). This bundle migrates
`internal/connectors/shortcut` (the hand-written legacy connector) to a declarative defs bundle at
capability parity; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Shortcut API token via the `api_token` secret; it is sent as the `Shortcut-Token` header
with no prefix (`api_key_header` mode, matching legacy's `connsdk.APIKeyHeader("Shortcut-Token",
token, "")` at `shortcut.go:136`) and is never logged. `base_url` defaults to
`https://api.app.shortcut.com` (legacy's `shortcutDefaultBaseURL`) and may be overridden for
tests/proxies.

## Streams notes

All four streams (`stories`, `epics`, `projects`, `iterations`) share an identical shape: `GET
/api/v3/<stream>`, records at the top-level `data` key, and Shortcut's own `next`-cursor
pagination convention — `pagination.type: cursor` with `cursor_param: next` and `token_path: next`
(the response's own `next` field is echoed back as the next request's `next` query param, matching
legacy's `harvest` loop at `shortcut.go:93-125` exactly: no `stop_path` is declared, since legacy
stops purely on an absent/empty `next` value with no separate boolean stop signal). `limit` is a
config-driven per-page-size override (`{{ config.page_size }}`, defaulting to legacy's own default
of 100 via `spec.json`'s `page_size` default and the query param's own `default: "100"`), matching
legacy's `pageSize` config resolution (`shortcut.go:218-231`) — legacy's `max_pages` config
override has no equivalent in this dialect (see Known limits). None of the four streams expose a
real server-side incremental filter in legacy (no date-range/updated-since query parameter is ever
sent); `x-cursor-field: updated_at` is declared on every schema purely as catalog/sort-key metadata
matching legacy's own `CursorFields` declaration (`shortcut.go:173`), and no `incremental` block is
declared on any stream, matching legacy's full-refresh-only read behavior exactly.

## Write actions & risks

None. Shortcut's legacy connector is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **Fallback field names are not modeled.** Legacy's `mapRecord` reads `state` with a fallback to
  `workflow_state_id`, and `updated_at` with a fallback to `updated_at_override`
  (`shortcut.go:152-154`, `first(item, "state", "workflow_state_id")` /
  `first(item, "updated_at", "updated_at_override")`). This bundle's plain schema projection copies
  `state`/`updated_at` by exact key match only — there is no coalesce/first-non-null filter in this
  dialect's `computed_fields` templating (a bare `{{ record.<path> }}` reference either resolves or
  is silently skipped for that record; it cannot try a second path). Legacy's own test suite
  (`shortcut_test.go`) never exercises the fallback branch for either field (both fixture and live
  test payloads always populate the primary key), and the real Shortcut API's documented story/epic/
  project/iteration shape always includes `state`/`updated_at` directly — this is judged an
  ACCEPTABLE, documented scope-narrowing rather than an `ENGINE_GAP`, per the `encharge` bundle's
  identical precedent for an unexercised defensive fallback.
- **`max_pages` is not runtime-configurable.** Legacy exposes a `max_pages` config override
  (`0`/`all`/`unlimited` for unbounded, or a positive integer hard cap, `shortcut.go:232-242`). The
  engine's `PaginationSpec.MaxPages` is a fixed bundle-authored literal, not config-driven; this
  bundle leaves it unset (unbounded), matching legacy's own default (`max_pages` absent/`0` also
  means unbounded on the legacy side) but losing the ability for an operator to cap page count via
  config.
- **Legacy's fixture-mode-only stamped fields are not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`) stamps a `fixture: true` marker onto every emitted
  record (`shortcut.go:161`); this is a credential-free conformance-harness affordance, not part of
  the live record shape, and is intentionally not reproduced — the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the equivalent test affordance.
