# Overview

JobNimbus is a wave2 fan-out declarative-HTTP migration. It reads JobNimbus CRM contacts, jobs,
tasks, activities, and file attachment metadata through the JobNimbus REST API
(`GET https://app.jobnimbus.com/api1/...`). This bundle migrates
`internal/connectors/jobnimbus` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip. Read-only: JobNimbus has no reverse-ETL write path.

## Auth setup

Provide a JobNimbus `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(secret)`
(`jobnimbus.go:225`). Never logged. `base_url` defaults to `https://app.jobnimbus.com/api1`
(legacy's `jobnimbusDefaultBaseURL`) and may be overridden for tests/proxies.

## Streams notes

All 5 streams share JobNimbus's offset pagination shape (an offset `from` param plus a `size` page
size, stopping on a short page); `pagination.type: offset_limit` with `offset_param: from`,
`limit_param: size` reproduces this exactly (`connsdk.OffsetPaginator`'s
`recordCount < PageSize` stop matches legacy's own short-page check). `streams.json`'s
`pagination.page_size: 2` (vs. legacy's real default of 1000) exists purely to keep the required
2-page `contacts` fixture small, per the identical precedent documented on auth0's and
aviationstack's goldens (`docs/migration/conventions.md`) — `PaginationSpec.PageSize` is a fixed
value with no config-driven override on either side, so `page_size`/`max_pages` are not declared
in `spec.json` (F6: a declared-but-unwireable key is worse than an absent one).

JobNimbus's list envelope key is inconsistent per stream, exactly mirrored by each stream's
`records.path`: `contacts`/`jobs`/`tasks` nest under `results`, `activities` nests under
`activity`, and `files` nests under `files`. Every object exposes a flat field set with no
nested-object lifting needed, so no `computed_fields` renames are required — plain schema
projection copies each field by exact key match.

Every stream declares `x-cursor-field: date_updated` (JobNimbus objects all carry a numeric
`date_updated` epoch-seconds timestamp; legacy's own `jobnimbusStreams()` declares the identical
`CursorFields: []string{"date_updated"}`), but **no stream declares an `incremental` block**:
legacy's `Read`/`harvest` never wires `date_updated` into an actual server-side filter request
parameter — it always performs a plain full page-walk regardless of any prior sync state
(`req.State` is read only in `readFixture`, the fixture-mode-only path, never in the live `Read`).
`date_updated` is catalog metadata only on both sides, not a wired incremental filter, so this
bundle matches that behavior exactly (no incremental fetching either).

## Write actions & risks

None. JobNimbus is exposed read-only (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` are not exposed as config.** Legacy exposes both as config-driven
  overrides (`jobnimbusPageSize`/`jobnimbusMaxPages`, `jobnimbus.go:258-286`); the engine's
  `offset_limit` paginator reads `PaginationSpec.PageSize`/`MaxPages` as fixed values resolved once
  at bundle load, with no template/config-driven override mechanism. This bundle sends a fixed
  page size (`2`, chosen only to keep the fixture small — see Streams notes) and does not cap
  `max_pages` (unbounded, matching legacy's own default of 0/unlimited).
- **No incremental fetch behavior on any stream** — see Streams notes. `date_updated` is declared
  as `x-cursor-field` for downstream dedup/ordering purposes only; it was never a wired
  server-side filter in legacy either, so this is capability parity, not a narrowing.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a
  synthetic, shared record shape across all 5 streams plus an extra `previous_cursor` field
  (echoing `req.State["cursor"]`) that is never part of the live wire shape
  (`jobnimbus.go:162-208`). This bundle's schemas and fixtures target the live path only. The
  engine's own conformance/fixture-replay harness provides the credential-free test affordance
  this bundle needs, so no fixture-mode equivalent is needed here.
