# Overview

Mention is a social listening and media monitoring tool. This bundle reads Mention app metadata,
accounts, alerts (monitored queries), mentions matched by an alert, alert tags, alert shares, alert
preferences, and alert tasks through the Mention REST API (`https://api.mention.net/api`). It is
migrated from `internal/connectors/mention` (read-only; `Capabilities.Write` is false there and
here).

## Auth setup

Provide a Mention API key via the `api_key` secret. It is sent as the RAW `Authorization` header
value (no `Bearer` prefix) via an `api_key_header` auth spec with an empty `prefix` — matching
legacy's `connsdk.APIKeyHeader("Authorization", secret, "")`. Never logged.

## Streams notes

- `app_data` (`GET app/data`) is an unpaginated singleton configuration object with languages,
  tones, sources, countries, folders, actions, integrations, and day names.
- `account_me` (`GET accounts/me`) and `account` (`GET accounts/{account_id}`) are unpaginated,
  single-object reads; primary key `["id"]`.
- `alert` (`GET accounts/{account_id}/alerts`), primary key `["id"]`, is paginated: Mention's list
  responses carry the next cursor at `_links.more.params.cursor` (absent on the final page), sent
  back as the `cursor` query param on the next request — expressed here as `pagination.type:
  cursor` with `token_path: "_links.more.params.cursor"` and `cursor_param: "cursor"` (no
  `stop_path` needed; legacy's own loop stops purely on an empty/repeated cursor token, which is
  the paginator's default behavior when `stop_path` is omitted). `limit` is sent as the static
  literal `100`, matching legacy's `mentionDefaultPageSize`. `spec.json`'s `page_size` property is
  informational only (documents the same default value and legacy's accepted 1-1000 range) and is
  not wired into a template — the same pattern stripe's golden bundle uses for its own `limit`
  query param (see `docs/migration/conventions.md`'s parity-deviation ledger item 3): a
  config-templated page-size query param cannot be safely used because conformance's dynamic
  checks populate EVERY declared config property (including `page_size`) with a synthetic
  non-numeric string, so any bundle that templates `{{ config.page_size }}` directly into a query
  param would send that literal synthetic string to the replay server during conformance instead of
  a real page size — a static literal sidesteps this entirely and exactly matches legacy's runtime
  behavior for a caller that never overrides `page_size` (the common case).
- `mention` (`GET accounts/{account_id}/alerts/{alert_id}/mentions`) uses the identical cursor
  pagination shape; primary key `["id"]`.
- `alert_tag` (`GET accounts/{account_id}/alerts/{alert_id}/tags`) is unpaginated, matching
  legacy's `paginated: false` for this endpoint; primary key `["id"]`.
- `alert_share`, `alert_preferences`, and `alert_task` read the documented alert-scoped shares,
  preferences singleton, and alert task list using the same required `account_id`/`alert_id`
  path inputs as `mention` and `alert_tag`.
- None of Mention's streams expose an incremental/updated-since filter in the API surface legacy
  itself reads, so every stream here is full-refresh only, matching legacy exactly (legacy never
  declared `CursorFields` either).

## Write actions & risks

Not applicable — Mention is read-only (`capabilities.write: false`, no `writes.json`), matching
legacy's `Write` returning `connectors.ErrUnsupportedOperation` unconditionally. Mention's
documented create/update/delete endpoints mutate alerts, accounts, shares, tags, tasks, mention
curation/read state, and alert preferences, so they remain excluded in `api_surface.json`.

## Known limits

- **Documented scope narrowing (account_id auto-discovery dropped):** legacy's `resolveAccountID`
  auto-discovers the account id from a pre-flight `GET accounts/me` call when `config.account_id` is
  unset. The engine's path-templating dialect has no mechanism to derive a path segment from a
  PRIOR response within the same read (no substream/pre-flight-lookup primitive exists in
  `streams.json`'s dialect) — expressing this would require a `StreamHook` (Tier-2/forbidden this
  wave). This bundle instead **requires** `account_id` in `spec.json` (`required: ["account_id"]`).
  This never changes emitted record data for any caller that already supplies `account_id`
  explicitly (the common case for any automated/scheduled sync); it only removes the
  interactive-convenience fallback for a caller that omitted it and expected auto-discovery. A
  caller relying on the auto-discovery convenience must now resolve and supply `account_id`
  themselves (e.g. via a one-time `GET accounts/me` call before configuring the connector).
- **`alert_id` required for `mention`/`alert_tag` streams**: matches legacy's own hard error
  (`"mention config alert_id is required for the mention and alert_tag streams"`) when unset for
  those two streams; declared as an optional `spec.json` property (not globally required) since
  `account_me`/`account`/`alert` don't need it, exactly mirroring legacy's per-stream requirement.
- **`max_pages` config dropped**: legacy accepts a runtime `max_pages` config value (0/`all`/
  `unlimited` for unbounded, else a positive integer hard cap). The engine's
  `PaginationSpec.MaxPages` is a fixed integer baked into the bundle, not a per-request templated
  value, so there is no mechanism to wire a runtime config value into it. Since legacy's own default
  is unbounded (`max_pages` unset), and declaring a fixed cap here would silently CHANGE accepted
  behavior for any caller relying on an unbounded sync, `max_pages` is left undeclared entirely (no
  cap enforced by this bundle, matching legacy's default) rather than kept as dead, unwireable
  config (F6, `docs/migration/conventions.md`).
- Fixtures use Mention's real wire shape for every field. The 2-page `alert`/`mention` fixtures set
  `_links.more.params.cursor` on page 1 and an empty `_links` object on page 2 (no `more` key),
  matching the paginator's stop-on-absent-token behavior exactly.
- Statistics require repeated query-array filters such as `alerts[]`, and mention children/authors
  require extra mention IDs or time/filter selectors. They are documented as excluded API surface
  rather than adding dead or globally required config that legacy did not expose for every stream.
