# Overview

Oura is a wave2 fan-out declarative-HTTP migration. It reads Oura member profile info and daily
readiness, sleep, and activity summaries through the Oura API v2 user-collection endpoints
(`GET https://api.ouraring.com/v2/usercollection/...`). This bundle migrates
`internal/connectors/oura` (the hand-written connector); the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide an Oura personal access token via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(key)`
(`oura.go:152`). `base_url` defaults to `https://api.ouraring.com/v2/usercollection` and may be
overridden for tests/proxies.

## Streams notes

`personal_info` is a single-object endpoint (`GET /personal_info`); `records.path: "."` reads the
JSON body root as one record, matching legacy's `recordsPath: "."` short-circuit (`oura.go:118`,
`244`).

`daily_sleep`, `daily_activity`, and `daily_readiness` share an identical shape: records live under
the `data` array, and pagination follows Oura's own `next_token` convention
(`pagination.type: cursor`, `cursor_param: next_token`, `token_path: next_token`) — the engine
requests the next page with `?next_token=<value>` and stops when the response omits `next_token`
(or it resolves empty), matching legacy's `harvest` loop exactly (`oura.go:93-124`).

Each of these three streams optionally accepts `start_date`/`end_date` config values, sent as plain
`start_date`/`end_date` query params via the optional-query dialect (`omit_when_absent: true`) —
present only when configured, matching legacy's `dateQuery` building an empty `url.Values{}` when
neither is set (`oura.go:184-193`).

`day` is declared as this bundle's `x-cursor-field` for the three daily streams, matching legacy's
own `CursorFields: []string{"day"}` (`oura.go:254-256`); no `incremental` block is declared because
Oura's real cursor-driven sync behavior in legacy is client-side date-range filtering only (the
`start_date`/`end_date` config knobs above), not a server-side incremental cursor read loop keyed
off a persisted state cursor — legacy never wires `CursorFields` into any state-cursor-driven
request param either.

## Write actions & risks

None. Oura's user-collection endpoints have no reverse-ETL writes; `capabilities.write` is `false`
and this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **`start_date`/`end_date` must be pre-formatted date-only strings (`YYYY-MM-DD`).** Legacy's
  `dateOnly` helper (`oura.go:195-207`) accepts either an RFC3339 timestamp (truncating it to its
  date portion) or an already-date-only string and passes either through unchanged otherwise. The
  engine dialect has no RFC3339-to-date-only conversion filter for an arbitrary `config.*`
  reference (the `date` `param_format` conversion in `conventions.md` §3 applies only to the
  incremental lower bound, not a plain config-driven query template) — so this bundle requires the
  caller to supply `start_date`/`end_date` already truncated to `YYYY-MM-DD`. This is a strictly
  *stricter* config-acceptance surface than legacy (an RFC3339 value that legacy would silently
  truncate is rejected/mismatched here rather than accepted), never a looser one, and never changes
  emitted record DATA for any accepted config — ACCEPTABLE per conventions.md §5's meta-rule.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size`/`max_pages`
  as config-driven overrides (`pageSize`/`maxPages`, `oura.go:209-231`). The engine's `cursor`
  (`token_path`) paginator never reads `PaginationSpec.PageSize`/`MaxPages` from a per-request
  config value — pagination is driven purely by the `next_token` presence/absence Oura's own API
  returns, matching bitly's identical documented precedent (`docs/migration/conventions.md`,
  bitly's `docs.md`). Neither key is declared in `spec.json` (F6: a declared-but-unwireable config
  key is worse than an absent one).
- **Legacy's fixture-mode-only `previous_cursor` echo field is not modeled.** Legacy's
  `readFixture` path (only reached when `config.mode == "fixture"`, a credential-free
  conformance-harness affordance) stamps `previous_cursor` onto every fixture-mode record when a
  prior state cursor happens to be set (`oura.go:133-135`). This is not part of the LIVE record
  shape; this bundle's schemas target the live path only, per the same precedent bitly documents.
  The engine's own conformance/fixture-replay harness provides the credential-free test affordance
  this bundle needs.
