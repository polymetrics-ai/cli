# Overview

Clockify is a wave2 fan-out declarative-HTTP migration. It reads Clockify workspaces, clients,
projects, tags, and users through the Clockify REST API v1
(`GET https://api.clockify.me/api/v1/...`). This bundle targets capability parity with
`internal/connectors/clockify` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Clockify API key via the `api_key` secret; it is sent as the `X-Api-Key` header
(`api_key_header` auth mode), matching legacy's `connsdk.APIKeyHeader("X-Api-Key", secret, "")`
(`clockify.go:202`). It is never logged. `base_url` defaults to `https://api.clockify.me/api` and
may be overridden for tests/proxies (legacy's own `clockifyBaseURL` validates scheme+host the same
way; the engine's base-URL resolution has no equivalent runtime validation, but every
conformance fixture only ever points at an httptest server, so this is not exercised differently
on either side).

## Streams notes

`workspaces` reads the top-level, unscoped `/v1/workspaces` endpoint. `clients`, `projects`,
`tags`, and `users` are scoped under `/v1/workspaces/{{ config.workspace_id }}/<resource>` — an
absent `workspace_id` hard-errors on both sides (legacy: `"clockify connector requires config
workspace_id for this stream"`; engine: an unresolved `config.workspace_id` path-template key —
same failure classification, different literal text, per conventions.md §5's precedent for
config-validation parity).

All five list endpoints return a bare top-level JSON array (no envelope), so every stream declares
`records.path: "."` (the dotted-path root selector). Pagination is 1-indexed page-number
(`pagination.type: page_number`, `page_param: page`, `size_param: page-size`) with `page_size: 50`
matching legacy's `clockifyDefaultPageSize`; a page returning fewer than 50 records is the last
page, matching `connsdk.PageNumberPaginator`'s exact stop rule legacy itself uses
(`clockify.go:136-141`).

None of Clockify's five list endpoints expose an incremental cursor field (legacy's own
`clockifyStreams` comment: "Clockify list endpoints do not expose an updated-at cursor field, so
these streams are full-refresh (no cursor)") — this bundle declares no `incremental` block for any
stream, matching legacy exactly.

## Write actions & risks

None. Clockify is a read-only source connector (legacy's `Write` always returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`clockifyPageSize`/`clockifyMaxPages`, `clockify.go:252-280`). The engine's
  `page_number` paginator's `PageSize`/`MaxPages` fields are plain JSON values in `streams.json`,
  not templated against `config.*` — there is no mechanism in this dialect to wire a runtime
  config value into either field. This bundle ships legacy's own default (`page_size: 50`,
  `max_pages` unbounded) as a static value; an operator can no longer override the page size or
  cap request count per sync. This mirrors the identical, already-accepted limitation documented
  for `bitly`'s `next_url` paginator and other wave1 goldens (conventions.md's fixture-rules
  section references this pattern).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance,
  `clockify.go:154-185`) stamps a broader, cross-stream synthetic record shape (e.g. every fixture
  record carries `workspaceId`, `clientId`, `duration`, etc. regardless of stream) that does not
  match any single stream's real live-API record shape. This bundle's schemas and fixtures target
  the live per-stream record shape only; the engine's own conformance/fixture-replay harness
  provides the credential-free test affordance this bundle needs, so no fixture-mode equivalent is
  needed here.
- **`api_url` alternate config-key name is not modeled.** Legacy's `clockifyBaseURL` also accepts
  `config.api_url` as a fallback name for the base URL override (`clockify.go:233-235`); this
  bundle declares only `base_url` (spec.json's single, canonical property name) since the engine's
  `spec.json` "default" materialization mechanism has no analogous "try this key, then that key"
  fallback chain. An operator relying on the `api_url` config key name specifically would need to
  rename it to `base_url`.
