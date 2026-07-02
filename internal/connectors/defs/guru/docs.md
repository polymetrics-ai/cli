# Overview

Guru is a wave2 fan-out declarative-HTTP migration. It reads Guru collections, groups, members,
and teams through the Guru REST API (`GET https://api.getguru.com/api/v1/...`). This bundle
targets capability parity with `internal/connectors/guru` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Guru account email via the `username` config value and a Guru API token via the
`password` secret; both are required. They are sent as HTTP Basic auth
(`Authorization: Basic base64(username:password)`), matching legacy's
`connsdk.Basic(username, secret)` (`guru.go:189-194`) exactly. `base_url` defaults to
`https://api.getguru.com/api/v1` and may be overridden for tests/proxies.

## Streams notes

All 4 streams (`collections`, `groups`, `members`, `teams`) are simple `GET` list endpoints whose
responses are top-level JSON arrays (`records.path: ""`), matching legacy's `guruStreamEndpoints`
table exactly. Pagination follows Guru's RFC 5988 `Link: <url>; rel="next"` header convention
(`pagination.type: link_header`), matching legacy's `connsdk.LinkHeaderPaginator`. `pageSize` is
sent via `{{ config.page_size }}` on every request (default `50`, matching legacy's
`guruDefaultPageSize`).

`members` nests each member's identity (`id`/`email`/`firstName`/`lastName`) under a `user` object
in the live API; `computed_fields` reaches into `record.user.<field>` to promote those fields to
the top level, matching legacy's `guruMemberRecord`'s `pick()` helper. Legacy's `pick()` falls back
to the top-level field when `user` is absent (used only by legacy's fixture-mode records, not the
live API); this bundle reproduces that fallback for free: a `computed_fields` template whose source
path (`record.user.id`, etc.) is absent is silently skipped rather than erroring
(`engine/read.go`'s `applyComputedFields`), so schema projection's own direct top-level key match
(`item["id"]`) still populates the field when there's no nested `user` object — identical net
result to legacy's fallback, with no extra Go.

## Write actions & risks

None. Guru is a read-only source connector (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **The fixture-replay harness cannot exercise `link_header`'s real 2-page continuation.**
  `fixtures/streams/**` (`conformance/replay.go`'s `fixtureResponse` shape) has no field for
  declaring HTTP response headers — only `status` and `body` — so a fixture page can never carry
  the `Link: <url>; rel="next"` header Guru's real API sends, and `pagination_terminates`/
  `records_match_schema` can only ever observe the paginator's natural single-page stop (no Link
  header present = no next page, exactly like a real last page). This is a structural
  fixture-format limitation affecting any `link_header` bundle in this repo (see buildkite's
  identical documented limit), not a guru-specific shortcut. Every stream fixture here is a single,
  representative page; the 2-page Link-header-following codepath itself
  (`internal/connectors/engine/paginate.go`'s `linkHeaderPaginator`) is exercised by the shared
  engine's own `paginate_test.go` coverage, not by this bundle's fixtures. Per hard-rule scope for
  this migration wave, no Go (hooks/paritytest) was authored to work around this.
- **`max_pages` is not runtime-enforced beyond the engine's generic hard cap.** Legacy exposes
  `max_pages` as a config-driven request-count override read fresh on every `Read` call
  (`guruMaxPages`). The engine's declarative read path enforces an equivalent hard cap
  (`PaginationSpec.MaxPages`, `read.go`'s `readDeclarative` loop) when `max_pages` is a *positive*
  integer, but this bundle's `spec.json` types `max_pages` as a free-form string (`"0"` default,
  matching stripe's precedent) with no template wiring it into `pagination.max_pages` (the
  `link_header` paginator type takes no per-request page-count field at all in `PaginationSpec`
  outside the generic `MaxPages` hard cap read from the bundle's own declared pagination block, not
  from a config value) — a config-driven override is therefore not wired for guru's `link_header`
  streams, matching the same gap class buildkite/bitly document for their own non-`page_number`
  paginators. `page_size` (`{{ config.page_size }}`) IS wired and config-driven, unlike
  `max_pages`.
- **Legacy's fixture-mode-only behavior is not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`) emits deterministic synthetic records with a fixed shape
  that diverges slightly from the live API (e.g. always including every field regardless of
  stream). This bundle's schemas and fixtures target the live record shape only; the engine's own
  conformance/fixture-replay harness (`internal/connectors/conformance`) provides the
  credential-free test affordance legacy's fixture mode was built for, so no equivalent is needed
  here.
