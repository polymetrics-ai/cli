# Overview

PagerDuty is a wave2 fan-out declarative-HTTP migration. It reads PagerDuty incidents, users,
services, and teams through the REST API (`GET https://api.pagerduty.com/...`). This bundle
targets capability parity with `internal/connectors/pagerduty` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.
Read-only: legacy's `Write` always returns `connectors.ErrUnsupportedOperation`, and this bundle
declares `capabilities.write: false` with no `writes.json` to match.

## Auth setup

Provide a PagerDuty REST API key via the `api_key` secret. It is sent on the `Authorization` header
with the literal prefix `Token token=` (`Authorization: Token token=<api_key>`), matching legacy's
`connsdk.APIKeyHeader("Authorization", apiKey, "Token token=")` (`pagerduty.go:196`) exactly via the
engine's `api_key_header` auth mode (`header`/`value`/`prefix` fields map 1:1 onto
`connsdk.APIKeyHeader`'s three constructor arguments). A static `Accept:
application/vnd.pagerduty+json` header is declared on `base.headers`, matching legacy's
`Requester.Accept` field (`pagerduty.go:196`) — PagerDuty's v2 API requires this Accept header on
every request. `base_url` defaults to `https://api.pagerduty.com` and may be overridden for
tests/proxies (legacy's own `baseURL` helper validates scheme+host the same way; the engine's
base-URL resolution has no equivalent runtime validation, but every fixture/conformance run only
ever points at an httptest server, so this is not exercised differently on either side).

## Streams notes

Four streams, all primary-keyed on `id`: `incidents`, `users`, `services`, `teams`. Each is a flat
list endpoint whose records live at a top-level key matching the stream name
(`connsdk.RecordsAt(resp.Body, endpoint.recordsPath)`, `pagerduty.go:145`, where
`recordsPath == resource` for every stream in legacy). Pagination is offset+limit
(`pagination.type: offset_limit`, `limit_param: limit`, `offset_param: offset`, `page_size: 100`
matching legacy's `pagerDutyDefaultLimit`) — the engine's `OffsetPaginator` stops on a short page
(fewer records returned than the page size), which coincides with PagerDuty's own documented
termination behavior (a short final page and `more: false` occur together in practice). Legacy's
own stop condition additionally checks the response body's `more` boolean field directly
(`more != "true" || len(records) == 0`, `pagerduty.go:158`); the engine's `offset_limit` paginator
has no equivalent body-driven stop-signal hook (unlike the `cursor` paginator's `stop_path`), so
this bundle relies on the short-page signal alone — see Known limits.

None of the four streams expose a server-side incremental filter parameter in legacy (`Read` never
sends a created-after query param — `harvestOffset` only ever sends `limit`/`offset`), so this
bundle declares no `incremental` block for any stream, matching legacy exactly. Each schema still
declares `x-cursor-field: created_at`, mirroring legacy's own `Catalog` `CursorFields` declaration
(`pagerduty.go:50-69`) for informational/dedup-mode purposes only — per `docs/migration/
conventions.md` §2, `incremental_append` sync modes are gated on the presence of an `incremental`
block, not on `x-cursor-field` alone, so declaring the field without the block adds no incremental
capability legacy itself doesn't have (see zendesk-support's identical precedent).
`incidents.incident_number` is a real PagerDuty API JSON integer; plain schema projection (no
`computed_fields` needed, since legacy's own `mapRecord` copies raw field values by exact key match,
not via string templating) preserves its native type, matching legacy's own `map[string]any` copy.

## Write actions & risks

None. Legacy's own package doc states it is "a read-only native PagerDuty HTTP API connector";
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`more`-boolean stop signal is not independently checked.** Legacy's `harvestOffset` stops when
  EITHER the response's `more` field is not `"true"` OR the page returned zero records
  (`pagerduty.go:158`); the engine's `offset_limit` paginator (`connsdk.OffsetPaginator`) only
  implements the short-page stop rule (fewer records than the declared page size) and has no
  `stop_path`-equivalent hook to also read `more` directly (that mechanism exists only on the
  `cursor` paginator variant, per `docs/migration/conventions.md` §3's pagination table). In every
  real PagerDuty response, a short/empty final page and `more: false` occur together (a full page
  with `more: false` would be a genuinely unusual/malformed response for this API), so this is not
  expected to diverge in practice; it is documented as an `ENGINE_GAP`-adjacent limitation rather
  than silently assumed identical, since a hypothetical full-size final page with `more: false`
  would cause the engine to request one extra (empty) page that legacy would not.
- **`limit` is not runtime-configurable.** Legacy exposes a config-driven `limit` override
  (`pagerDutyDefaultLimit`/`pagerDutyMaxLimit`, `pagerduty.go:236-249`). The engine's `offset_limit`
  paginator's `PageSize` is a bundle-declared constant (`streams.json`'s `base.pagination.page_size:
  100`), with no per-request config-driven override mechanism — a `{{ config.limit }}` template on
  `page_size` is not expressible (`PaginationSpec.PageSize` is a plain int field, not a template).
  This bundle therefore fixes PagerDuty's own default (`limit=100`) and does not declare `limit` in
  `spec.json` at all (a declared-but-unwireable config key is worse than an absent one, per the
  bitly/searxng F6 precedent).
- **`max_pages` is not modeled.** Legacy's hard request-count cap override (`pagerduty.go:251-264`)
  has no engine-side equivalent wired to a config value for `offset_limit`; pagination is bounded
  only by the short-page stop signal, matching PagerDuty's own real termination behavior.
- **Fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached when
  `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps an extra
  `previous_cursor` field (echoing `req.State["cursor"]` when a prior cursor happens to be set) onto
  every fixture-mode record (`pagerduty.go:173-175`). This is not part of the LIVE record shape;
  this bundle's schemas target the live path only. The engine's own conformance/fixture-replay
  harness provides the credential-free test affordance this bundle needs, so no fixture-mode
  equivalent is needed here.
