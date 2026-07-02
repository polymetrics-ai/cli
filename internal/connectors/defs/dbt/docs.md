# Overview

dbt Cloud is a workflow orchestration product for dbt (data build tool). This bundle reads dbt
Cloud projects, runs, repositories, users, and environments through the dbt Cloud Administrative
API v2, account-scoped under `/accounts/{account_id}/`. Read-only, full-refresh only (the
Administrative API v2 exposes no incremental filter parameter on any of these list endpoints).
This bundle migrates `internal/connectors/dbt` (the hand-written connector); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a dbt Cloud service token via the `api_key_2` secret; it is sent as the `Authorization`
header in the form `Token <api_key_2>` (`auth: [{"mode": "api_key_header", "header":
"Authorization", "prefix": "Token "}]`), matching legacy's `connsdk.APIKeyHeader("Authorization",
secret, "Token ")`. A required `account_id` config value scopes every list endpoint under
`/accounts/{account_id}/`.

## Streams notes

All 5 streams (`projects`, `runs`, `repositories`, `users`, `environments`) share the identical
shape: `GET` against the account-scoped resource, records at `data`, primary key `["id"]`. dbt
Cloud's list envelope is `{"data": [...], "extra": {"pagination": {"count", "total_count"},
"filters": {"limit", "offset"}}}`; pagination uses `pagination.type: offset_limit` with
`limit_param: limit`, `offset_param: offset`, `page_size: 100` (matches legacy's
`dbtDefaultPageSize`/`dbtMaxPageSize`, both 100) — the engine's offset/limit paginator stops on a
short page (fewer records than `page_size`), the same primary stop condition legacy's `harvest`
uses.

No stream declares an `incremental` block: the dbt Cloud Administrative API v2 list endpoints
legacy reads have no server-side updated-since filter, so every sync here is full-refresh, exactly
matching legacy (which never declared cursor fields for any dbt stream either).

## Write actions & risks

None. dbt Cloud is read-only here (`capabilities.write: false`); the dbt Cloud Administrative API's
write surface (triggering job runs, project/job/environment CRUD) is operational, not reverse-ETL
data, and legacy itself never implemented `Write` (it returns `ErrUnsupportedOperation`
unconditionally).

## Known limits

- Legacy's `harvest` loop has a defensive extra stop condition this bundle does not reproduce:
  when the response's `extra.pagination.total_count` is present and the running emitted count has
  already reached it, legacy stops even if the just-fetched page was exactly `page_size` long (a
  page that happens to end precisely on a `page_size` boundary). The engine's `offset_limit`
  paginator only recognizes the short-page stop signal (`recordCount < page_size`). For any real
  dbt Cloud account this is a one-extra-page-request difference at most (the very next page would
  come back empty or short and stop there), never a difference in which records are emitted —
  documented parity deviation, not a fixable gap (`offset_limit`'s stop rule is fixed engine
  behavior, not spec-declarable). See `docs/migration/conventions.md` §5's meta-rule: this never
  changes emitted record data for any input legacy would accept.
- `page_size`/`max_pages` config keys legacy exposed are not declared in `spec.json`: the engine's
  `offset_limit` paginator reads its page size and `MaxPages` only from `streams.json`'s
  statically-declared `pagination` block, with no mechanism to source either from
  `RuntimeConfig.Config` at read time (same limitation documented for searxng's `page_size`/
  `max_pages`, `docs/migration/conventions.md`'s Tier-1 read-only variant section). Declaring a
  `spec.json` property no template ever consumes is dead config (F6, REVIEW.md), so this bundle
  fixes `page_size: 100` (legacy's own default and max) and leaves `max_pages` unbounded (`0`,
  legacy's own "unlimited" default) rather than declaring unwireable config.
- The full dbt Cloud Administrative API v2 surface (triggering runs, job/environment CRUD,
  service tokens) is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
- `account_id` is required in `spec.json` even though the dbt Cloud API technically supports a
  "default account" lookup for some UI flows; legacy always requires it explicitly
  (`dbtAccountID`), so this bundle matches that stricter, unambiguous behavior.
