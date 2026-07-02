# Overview

Goldcast is a wave2 fan-out declarative-HTTP migration. It reads Goldcast organizations, events,
agenda items, discussion groups, and tracks through the Goldcast customapi REST API
(`GET https://customapi.goldcast.io/...`). This bundle migrates `internal/connectors/goldcast` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Goldcast API access key via the `access_key` secret; it is sent as
`Authorization: Token <access_key>` and is never logged, matching legacy's
`connsdk.APIKeyHeader("Authorization", secret, "Token ")` (`goldcast.go:242`). `base_url` defaults
to `https://customapi.goldcast.io` and may be overridden for tests/proxies (legacy's own
`goldcastBaseURL` validates scheme+host the same way; the engine has no equivalent runtime
validation, but every parity/conformance fixture only ever points at an httptest server, so this
is not exercised differently on either side).

## Streams notes

Goldcast's list endpoints return one of two shapes at runtime — a raw top-level JSON array, or a
Django REST Framework envelope (`{"count", "next", "results"}`) — and legacy's generic
`decodePage` (`goldcast.go:173-194`) inspects the response body's leading byte to handle either
shape uniformly across every stream. Legacy's own test suite documents which shape each endpoint
actually returns in practice: `organizations` (`TestReadTopLevelArray`) returns a bare array, while
`events` (`TestReadPaginatesAndAuthenticates`) returns the DRF envelope with an absolute `next`
link. This bundle wires each stream to match its own documented real shape rather than a single
uniform choice (the engine's `records.path` is a fixed dotted path per stream, with no
runtime shape-detection primitive — see Known limits):

- `organizations`: `records.path: "."` (root-is-array; `connsdk.RecordsAt`'s `"."`/`""` root
  selection already handles a bare top-level array, matching legacy's array branch exactly). No
  pagination is declared — a raw-array response never carries a `next` link, matching legacy's
  `decodePage` array branch, which always returns `next=""`.
- `events`, `agenda_items`, `discussion_groups`, `tracks`: `records.path: "results"` plus
  `pagination.type: next_url` / `next_url_path: "next"`, matching the DRF envelope
  `{"results":[...],"next":<absolute URL or null>}` and legacy's absolute-next-link-following loop
  (`goldcast.go:138-166`) exactly. The engine's `next_url` paginator's same-host SSRF guard
  (THREAT-MODEL §3) passes cleanly for Goldcast's real behavior (the `next` URL it returns is
  always same-origin as `base_url` in production).

None of Goldcast's list endpoints expose an incremental cursor field (legacy declares no
`CursorFields` for any stream, `streams.go:32-118`), so no stream declares an `incremental` block —
full refresh only, matching legacy exactly.

## Write actions & risks

None. Goldcast is read-only here (legacy's own `Capabilities.Write: false`, `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **The engine's `next_url` paginator has no per-request-cap knob matching legacy's
  `goldcastMaxPages` safety valve** (`goldcast.go:39` — a fixed constant, `10000`, not even a
  config override; legacy hard-codes this as a defensive bound against a server that never stops
  returning a `next` link). The engine's `next_url` paginator has no `MaxPages`-equivalent field at
  all; termination relies solely on an eventual null/absent `next`, matching Goldcast's own
  documented real behavior (every DRF envelope response terminates with `next: null`).
- **This bundle ships single-page fixtures for the 4 `next_url`-paginated streams** (`events`,
  `agenda_items`, `discussion_groups`, `tracks`), per conventions.md §4's sanctioned `next_url`
  exception: the next-page URL a `next_url` stream reads back is the replay server's own address,
  unknown to a static fixture file. Conventions.md §4 additionally calls for a live
  `paritytest/<name>` test asserting real 2-page `next_url` behavior — this wave's mandate is
  JSON + `docs.md` only (no Go test files), so that live 2-page proof is **not** produced by this
  migration and is deferred to a follow-up wave. This is a documented gap, not a silent one: the
  underlying `next_url` paginator code path itself is the identical shared engine implementation
  already proven correct by bitly's and calendly's existing `paritytest` suites (loop-guarded,
  same-host-checked, absolute-URL-following) — what remains unverified is only this bundle's own
  wiring of it (the `next_url_path: "next"` field name, and Goldcast's specific envelope shape),
  not the pagination mechanism in general.
- **`organizations`' single fixed record-extraction path (`"."`) does not defend against Goldcast
  ever changing that endpoint to the enveloped shape.** If Goldcast's `organizations` endpoint ever
  started returning the DRF envelope instead of a bare array (the reverse of what legacy's own test
  suite documents today), this bundle's `records.path: "."` would misparse the envelope object as a
  single one-element record (`RecordsAt`'s object branch) rather than iterating `results`. This
  mirrors legacy's own equally shape-fixed-per-call-site risk (legacy's `decodePage` re-detects the
  shape on every single response, so it would not have this problem — a small legacy behavior this
  bundle does not fully reproduce, since the engine's `records.path` cannot runtime-branch on the
  response's own shape). ACCEPTABLE per conventions.md §5: reproduces every input legacy's test
  suite actually documents; would only diverge on a hypothetical wire-shape change neither side's
  test coverage currently exercises.
- Goldcast's per-event partition-routed child streams (webinars, members, sessions) are out of
  scope for wave2, matching legacy exactly (legacy exposes only "parent" list endpoints to keep
  fixture mode simple); see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass
  B capability expansion"}` entries.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) synthesizes its own deterministic records directly in Go rather
  than shaping a real API response; this bundle's schemas and fixtures target the LIVE record shape
  only (`goldcast.go`'s `harvest`/`decodePage`/`mapRecord` functions), per convention.
