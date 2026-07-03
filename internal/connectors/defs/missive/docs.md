# Overview

Missive is a wave2 fan-out declarative-HTTP migration. It reads Missive contacts, contact groups,
users, teams, and shared labels through the Missive REST API (`GET
https://public.missiveapp.com/v1/...`). This bundle targets capability parity with
`internal/connectors/missive` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Missive API token via the `api_key` secret; it is sent as a Bearer token (`Authorization:
Bearer <api_key>`, `auth.mode: bearer`), matching legacy's `connsdk.Bearer(secret)`
(`missive.go:212`); the secret is never logged. `base_url` defaults to
`https://public.missiveapp.com/v1` and may be overridden for tests/proxies.

## Streams notes

All 5 streams (`contacts`, `contact_groups`, `users`, `teams`, `shared_labels`) share Missive's
offset pagination (`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: offset`) —
a page shorter than the declared page size stops pagination, matching legacy's own short-page stop
rule (`missive.go:162-165`) exactly. Every list endpoint wraps its records under a top-level key
matching the resource name (`{"contacts":[...]}`, `{"users":[...]}`, etc.), matching legacy's
`recordsPath == resource` convention (`streams.go:9-16`).

`contact_groups` optionally filters by `kind` (`'group'` or `'organization'`) via the `kind` config
value, sent as a query param only when set (`stream.Query`'s `omit_when_absent` dialect,
conventions.md §3) — matching legacy's own conditional `base.Set("kind", kind)` (`missive.go:123-127`,
only applied when `req.Config.Config["kind"]` is non-empty). `contacts`' `modified_at` field is a
Unix-seconds integer on the wire; the schema declares it `"integer"` (typed passthrough via plain
schema projection, no `computed_fields` needed since the raw key already matches the schema name).
None of the five streams expose an incremental cursor field (legacy's own `missiveStreams()`
declares no `CursorFields` for any stream), so every stream is full-refresh only, matching legacy
exactly.

## Write actions & risks

None. Missive's source is read-only (legacy's own package doc: "Missive's source is read-only
(full-refresh)"); `capabilities.write` is `false` and this bundle ships no `writes.json`, matching
legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`limit`/`page_size` config overrides are not modeled.** Legacy exposes a config-driven
  page-size override (`missivePageSize`, `missive.go:247-263`, reads `limit` first, falls back to
  `page_size`, default 50, max 200). The engine's `offset_limit` paginator's `PageSize` is a static
  bundle-authored int (not templated), so there is no way to expose it as a config override;
  neither `limit` nor `page_size` is declared in `spec.json` (F6, REVIEW.md: a declared-but-unwireable
  config key is worse than an absent one). This bundle declares a fixed `page_size: 50` (legacy's
  own default) rather than a fixture-convenience value — a fixture-convenience page size is never
  leaked into the live pagination config (callrail's/bamboo-hr's documented precedent,
  `docs/migration/conventions.md`). The mandatory 2-page conformance fixture
  (`fixtures/streams/contacts/{page_1,page_2}.json`) is sized to match live behavior instead: page 1
  returns a full 50-record page (so the paginator continues to page 2) and page 2 returns the
  3-record remainder, honestly exercising the short-page stop rule without shrinking the live
  `page_size`.
- **`max_pages` is not runtime-configurable.** Legacy exposes `max_pages` as a config-driven hard
  request-count cap override (`missiveMaxPages`, `missive.go:265-278`, default 0/unbounded). The
  engine's `PaginationSpec.MaxPages` is a static bundle-authored int (not templated) — there is no
  config-driven knob to wire it to, so it is left unset (unbounded), matching legacy's own default
  behavior.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) synthesizes
  deterministic records directly in Go rather than exercising `harvest`/`mapRecord` at all
  (`missive.go:174-196`). This bundle's schemas and fixtures target the LIVE `harvest`/`mapRecord`
  path only; the engine's own conformance/fixture-replay harness provides the credential-free test
  affordance legacy's fixture mode existed for, so no fixture-mode equivalent is needed here.
