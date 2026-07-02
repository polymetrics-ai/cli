# Overview

Jotform is a wave2 fan-out declarative-HTTP migration. It reads Jotform forms, submissions,
reports, folders, and the authenticated account profile through the Jotform REST API
(`GET https://api.jotform.com/...`). This bundle migrates `internal/connectors/jotform` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip. Read-only: Jotform's writes (creating/deleting forms, submissions) are not modeled by legacy
either — legacy's own package doc calls them "not safe reverse-ETL targets".

## Auth setup

Provide a Jotform `api_key` secret; it is sent as the `APIKEY` request header
(`api_key_header` auth mode), matching legacy's `connsdk.APIKeyHeader("APIKEY", secret, "")`
(`jotform.go:247`). Never logged. `base_url` defaults to `https://api.jotform.com`.

## Streams notes

`forms` and `submissions` use Jotform's `resultSet` offset/limit pagination
(`pagination.type: offset_limit`, `offset_param: offset`, `limit_param: limit`), stopping on a
short/empty page exactly like legacy's own `!endpoint.paginated || len(records) == 0 ||
len(records) < pageSize` check. `reports`, `folders`, and `user` are non-paginated
single-response reads (`pagination.type: none` overrides the base spec per-stream), matching
legacy's `endpoint.paginated == false` branch (no `limit`/`offset` query params sent at all).
`streams.json`'s `pagination.page_size: 2` (vs. legacy's real default of 100) exists purely to
keep the required 2-page `forms` fixture small, per the identical precedent documented on auth0's
and aviationstack's goldens (`docs/migration/conventions.md`) — `PaginationSpec.PageSize` is a
fixed value with no config-driven override on either side, so `page_size`/`max_pages` are not
declared in `spec.json` (F6: a declared-but-unwireable key is worse than an absent one).

Every stream's records live at the top-level `content` key regardless of pagination shape — this
is Jotform's uniform envelope (`{responseCode, content: [...] | {...}, resultSet?: {...}}`);
`records.path: "content"` is declared identically on all 5 streams. `user` is the one stream whose
`content` is a single object rather than an array — the engine's `records` extraction (mirroring
legacy's `connsdk.RecordsAt`) wraps a single JSON object at the records path as one record, so no
special-casing is needed. Every object exposes a flat field set already matching the schema
1:1 (legacy's own `mapRecord` functions do no renames or nested lifting), so no `computed_fields`
are declared on any stream.

`forms`/`submissions`/`reports` declare `x-cursor-field: created_at` (matching legacy's
`jotformStreams()` `CursorFields: []string{"created_at"}`), but no `incremental` block: legacy
never wires `created_at` into an actual request filter — its `Read`/`harvest` always performs a
plain full read regardless of prior sync state. `folders` and `user` declare no cursor field
either, matching legacy exactly.

## Write actions & risks

None. Jotform is a read-only source connector (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`url_prefix`-based regional base URL selection is not modeled.** Legacy derives the EU
  (`https://eu-api.jotform.com`) or HIPAA (`https://hipaa-api.jotform.com`) host from a
  `url_prefix` config value (`eu`/`hipaa`), falling back to the default host otherwise
  (`jotform.go:263-274`). The engine's `spec.json` `"default"` materialization mechanism only fills
  in a fixed literal default and has no way to express a derivation from another config key's
  value (conventions.md §3), so this bundle requires setting `base_url` to the desired regional
  host directly instead of declaring `url_prefix` — the same narrowing documented on datadog's and
  jira's region/site-derived `base_url` in this wave. An operator who previously set
  `url_prefix=eu` now sets `base_url=https://eu-api.jotform.com` instead — same reachable value
  space, different config key name.
- **`page_size`/`max_pages` are not exposed as config.** Legacy exposes both as config-driven
  overrides (`jotformPageSize`/`jotformMaxPages`, `jotform.go:288-316`); the engine's
  `offset_limit` paginator reads `PaginationSpec.PageSize`/`MaxPages` as fixed values resolved once
  at bundle load, with no template/config-driven override mechanism. This bundle sends a fixed
  page size (`2`, chosen only to keep the fixture small — see Streams notes) and does not cap
  `max_pages` (unbounded, matching legacy's own default of 0/unlimited).
- **No incremental fetch behavior on any stream** — see Streams notes. `created_at` is declared as
  `x-cursor-field` on `forms`/`submissions`/`reports` for downstream dedup/ordering purposes only;
  it was never a wired server-side filter in legacy either, so this is capability parity, not a
  narrowing.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (reached only
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a
  synthetic, shared record shape across all 5 streams plus an extra `previous_cursor` field
  (echoing `req.State["cursor"]`) that is never part of the live wire shape
  (`jotform.go:187-230`). This bundle's schemas and fixtures target the live path only. The
  engine's own conformance/fixture-replay harness provides the credential-free test affordance
  this bundle needs, so no fixture-mode equivalent is needed here.
