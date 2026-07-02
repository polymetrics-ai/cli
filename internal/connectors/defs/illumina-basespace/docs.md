# Overview

Illumina BaseSpace is a wave2 fan-out declarative-HTTP migration. It reads BaseSpace projects,
runs, samples, app sessions, and datasets through the BaseSpace v1pre3 REST API (`GET
https://<domain>/v1pre3/users/<user>/...`). This bundle is a capability-parity port of the
hand-written connector at `internal/connectors/illumina-basespace` (`basespace.go`/`streams.go`),
which stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a BaseSpace access token via the `access_token` secret; it is sent as the
`x-access-token` header (`streams.json`'s `base.auth`, `mode: api_key_header`), matching legacy's
`connsdk.APIKeyHeader("x-access-token", token, "")` (`basespace.go:245`). BaseSpace is
domain-scoped with no single fixed default endpoint (US, EU, and other regional domains all
exist, e.g. `https://api.basespace.illumina.com` or `https://euw2.sh.basespace.illumina.com`), so
`base_url` is **required** here (legacy accepts either a `base_url` or a bare `domain` config
value, promoting a bare host to an `https://` URL when no scheme is present, `basespaceBaseURL`,
`basespace.go:274-299` — see Known limits for the narrowing this required-`base_url`-only bundle
makes). `user` defaults to `"current"` (the authenticated user), matching legacy's
`basespaceDefaultUser` (`basespace.go:33`).

## Streams notes

All five streams are user-scoped list endpoints under `/v1pre3/users/{{ config.user }}/<resource>`
(`projects`, `runs`, `samples`, `appsessions`, `datasets`), matching legacy's per-stream routing
table (`basespaceStreamEndpoints`, `streams.go:23-29`) and its `path := fmt.Sprintf("%s/users/%s/%s",
basespaceAPIPrefix, url.PathEscape(user), endpoint.resource)` construction (`basespace.go:143-144`)
— path segments are urlencoded by `InterpolatePath`'s per-segment default, matching legacy's
explicit `url.PathEscape(user)`. Records live at the nested `Response.Items` path on every
response, matching legacy's `connsdk.RecordsAt(resp.Body, "Response.Items")`
(`basespace.go:168`). Pagination is BaseSpace's Offset/Limit convention
(`pagination.type: offset_limit`, `limit_param: Limit`, `offset_param: Offset`, `page_size: 100`,
matching legacy's exact query param casing `query.Set("Limit", ...)`/`query.Set("Offset", ...)`
and its `basespaceDefaultPageSz` default, `basespace.go:161-163,33`), stopping on a short/empty
final page — the identical stop rule as legacy's `harvest` loop (`basespace.go:180-183`).

Every BaseSpace object exposes a string `Id` and a `DateCreated` timestamp
(`x-cursor-field: date_created` is published on every schema, matching legacy's comment that
cursor fields are "published for downstream incremental use but reads do not require them",
`streams.go:32-35`); no `incremental` block is declared for any stream since the upstream v1pre3
API is full-refresh only and legacy never sends an incremental filter — matching legacy exactly.

## Write actions & risks

None. BaseSpace is read-only in this connector (legacy's own package doc: "there is no safe
reverse-ETL write surface, so Capabilities.Write is false"); `capabilities.write` is `false` and
this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **The `domain`-config fallback is not modeled; `base_url` is required instead.** Legacy accepts
  EITHER a `base_url` OR a bare `domain` config value (promoted to an `https://` URL when no
  scheme is present) and only errors when BOTH are absent (`basespaceBaseURL`,
  `basespace.go:274-287`). The engine's `spec.json` `"default"` materialization mechanism handles
  only a FIXED literal default, not a derived/conditional value computed from a second config key
  at request time (conventions.md §3's "derived default" note names this exact shape — sentry's
  `hostname`-based URL, chargebee's `site`-based URL — as needing either a required `base_url` with
  the derivation dropped, or a future `computed_fields`-style base-URL template the dialect does
  not yet have). This bundle takes the documented "require `base_url`, drop the derivation" option:
  `base_url` is a required `spec.json` property with no `domain` fallback. A caller that previously
  relied on passing only a bare `domain` value must now pass a fully-qualified `base_url` instead
  (a config-surface narrowing, not a silent behavior change — an absent `base_url` hard-errors
  identically to legacy's absent-both case).
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (1-1000) and
  `max_pages` (integer, `all`, or `unlimited`) as config-driven overrides
  (`basespacePageSize`/`basespaceMaxPages`, `basespace.go:301-329`). The engine's `offset_limit`
  paginator reads `PaginationSpec.PageSize` as a fixed bundle-level value (there is no
  `{{ config.page_size }}` template point wired into the pagination block), so `page_size` is
  declared in `spec.json` with a `default: 100` for documentation parity with legacy's own default,
  but is not actually consumable as a runtime override in this bundle — matching the bitly
  `page_size` precedent (`docs/migration/conventions.md`). `max_pages` is likewise not modeled;
  pagination is bounded only by the short/empty-page stop signal, matching BaseSpace's own real
  termination behavior (legacy's `max_pages` unset/`0`/`all`/`unlimited` case is the same
  "unbounded" outcome).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps `connector`, `fixture`, and `previous_cursor` marker
  fields onto every fixture-mode record (`basespace.go:218-222`). This bundle's schemas and
  fixtures target the live `harvest` record shape only; the engine's own conformance/fixture-replay
  harness supplies the credential-free test affordance legacy's fixture mode existed for.
