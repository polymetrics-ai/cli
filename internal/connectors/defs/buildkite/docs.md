# Overview

Buildkite is a wave2 fan-out declarative-HTTP migration. It reads Buildkite organizations,
pipelines, builds, and agents through the Buildkite REST API v2
(`GET https://api.buildkite.com/v2/...`). This bundle migrates `internal/connectors/buildkite` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Buildkite API access token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`, matching legacy's `connsdk.Bearer(secret)`) and is never
logged. `base_url` defaults to `https://api.buildkite.com/v2` and may be overridden for
tests/proxies.

## Streams notes

`organizations` (`GET /organizations`) is top-level and needs no organization slug, matching
legacy's `scopeTopLevel`. `pipelines`, `builds`, and `agents` are organization-scoped
(`/organizations/{{ config.organization }}/...`, matching legacy's `scopeOrganization` +
`organizationSlug` validation) and require the `organization` config value; it is declared in
`spec.json` but intentionally NOT in `required[]` since the `organizations` stream never
references it — an absent `organization` hard-errors only when an org-scoped stream's path is
resolved, identical to legacy's own per-stream (not global) validation.

Every stream's list response is a top-level JSON array (`records.path: ""`, root), matching
legacy's "Buildkite list endpoints return a top-level JSON array" comment. Pagination follows
Buildkite's own RFC 5988 `Link: <url>; rel="next"` header convention (`pagination.type:
link_header`) — the byte-accurate parity choice, since legacy's own `connsdk.LinkHeaderPaginator`
IS Link-header following (unlike, e.g., this repo's `github` bundle, which deliberately chose
`page_number` because ITS legacy implementation used `page`/`per_page` query params instead of
Link headers despite GitHub's API also supporting them). Every request sends `per_page=100`
(matches legacy's `buildkiteDefaultPageSize`) via each stream's static `query: {"per_page": "100"}`.

`builds` supports Buildkite's `created_from` incremental lower bound (legacy: `if stream ==
"builds" { base.Set("created_from", createdGTE) }`) — expressed via the opt-in optional-query
dialect referencing `{{ incremental.lower_bound }}` with `omit_when_absent: true`, so
`created_from` is sent ONLY when the incremental lower bound resolves (persisted cursor, or the
RFC3339 `start_date` config value on a fresh sync), exactly matching legacy's own conditional
branch. `organizations`, `pipelines`, and `agents` declare no `incremental` block, matching legacy
(the `created_from` param is attached to `builds` only; legacy's own comment: "for other streams
the param is ignored harmlessly by the API but we only attach it to builds").

## Write actions & risks

None. Buildkite is read-only in this connector (legacy's own `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **The fixture-replay harness cannot exercise `link_header`'s real 2-page continuation.**
  `fixtures/streams/**` (this repo's fixture-replay JSON shape, `conformance/replay.go`'s
  `fixtureResponse`) has no field for declaring HTTP RESPONSE headers — only `status` and `body` —
  so a fixture page can never carry the `Link: <url>; rel="next"` header the real Buildkite API
  sends, and `pagination_terminates`/`records_match_schema` can only ever observe the paginator's
  natural single-page stop (no Link header present = no next page, exactly like a real
  last-page response). This is a structural fixture-format limitation affecting ANY `link_header`
  bundle in this repo, not a buildkite-specific shortcut — the same gap conventions.md §4
  documents for `next_url`'s single-page exception, but `link_header` has no declared harness
  exception of its own (yet) since no fixture-response-header field exists for either pagination
  type to exploit even if one wanted to. Every stream fixture here is therefore a single,
  representative page; the 2-page Link-header-following codepath itself
  (`internal/connectors/engine/paginate.go`'s `linkHeaderPaginator`) is exercised by the shared
  engine's own `paginate_test.go` coverage, not by this bundle's fixtures. Per hard-rule scope for
  this migration wave, no Go (hooks/paritytest) was authored to work around this — a future wave
  extending the fixture-replay format to carry response headers, or reusing github's `page_number`
  substitution IF Buildkite's real API accepted it (it is a legitimate Link-header-only API in
  production for pagination continuation, so that substitution would NOT be byte-accurate parity
  here), would close this gap properly.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (1-100,
  default 100) and `max_pages` as config-driven overrides read fresh on every `Read` call. The
  engine's `link_header` paginator has no page-size knob at all (Buildkite's own next-page URL,
  read from the Link header, already carries whatever `per_page` was requested on the first page);
  `per_page=100` is a static per-stream query literal, matching stripe's `limit=100` precedent.
  `max_pages` has no config-driven override either; pagination is bounded only by the absence of a
  `Link: rel="next"` header, matching Buildkite's own real termination behavior. `spec.json` still
  declares `page_size`/`max_pages` for documentation continuity with legacy's config surface, but
  neither is wired into any template.
- Legacy's fixture-mode-only fields (`connector`, `fixture`, `previous_cursor` static/echoed
  markers stamped only under `config.mode == "fixture"`) are not modeled; this bundle's schemas and
  parity target the live wire shape only, matching this repo's established convention for a legacy
  in-code fixture path now superseded by the engine's own conformance/fixture-replay harness.
