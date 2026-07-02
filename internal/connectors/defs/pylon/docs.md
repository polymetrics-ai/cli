# Overview

Pylon is a wave2 fan-out declarative-HTTP migration. It reads Pylon issues, accounts, contacts,
users, and messages through the Pylon REST API (`GET https://api.usepylon.com/<resource>`). This
bundle targets capability parity with `internal/connectors/pylon` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Pylon API token via the `api_token` secret; it is sent as `Authorization: Bearer
<api_token>` (matching legacy's `connsdk.Bearer(token)`), and is never logged. `base_url` defaults to
`https://api.usepylon.com`.

## Streams notes

All five streams share the identical Pylon envelope: `GET /<resource>` returns
`{"data":[...],"pagination":{"next_cursor":...}}`, records live at the top-level `data` array.
Pagination is `cursor` (`cursor_param: cursor`, `token_path: pagination.next_cursor`) — the next
page's `cursor` value is read from the response body's `pagination.next_cursor`; pagination stops
when that value is empty/null, matching legacy's own `strings.TrimSpace(nextCursor) == ""` stop rule
exactly (no `stop_path` is declared since legacy has no separate boolean stop signal beyond the
cursor itself, unlike Zendesk's `has_more`).

Every request sends `limit=100` via each stream's static `query` (matches legacy's
`defaultPageSize`), not via `pagination.size_param`/`page_size` (the `cursor`+`token_path` paginator
constructor never reads those fields — mirrors stripe's `limit=100`-via-static-query pattern).

`start_date`, when configured, is sent as the `updated_after` query parameter on every request for
every stream (matching legacy's `if start := ...; start != "" { query.Set("updated_after", start) }`
— a plain, unconditional-per-request config passthrough, not a resolved/formatted incremental lower
bound). This is declared via the opt-in optional-query object dialect
(`{"template": "{{ config.start_date }}", "omit_when_absent": true}`), which omits the parameter
entirely when `start_date` is unset, matching legacy's own absent-check.

`issues`, `accounts`, `contacts`, and `messages` publish `updated_at` as `x-cursor-field` (matching
legacy's own `CursorFields: []string{"updated_at"}` on those four streams); `users` has no
`x-cursor-field` (legacy declares no `CursorFields` for `users` either). No stream declares an
`incremental.request_param`, matching legacy's own behavior: `updated_after` is a static config
passthrough, never a computed incremental lower bound recorded from prior sync state.

## Write actions & risks

None. Legacy `pylon` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- **`name`/`state` fallback fields are not modeled.** Legacy's `mapRecord` sets
  `name: first(item, "name", "full_name", "email")` and `state: first(item, "state", "status")` —
  if a record's own `name`/`state` key is absent, legacy falls back to `full_name`/`email` or
  `status` respectively. This bundle's schema projection copies only the exact-named `name`/`state`
  keys (no coalesce/fallback filter exists in the current dialect); a record whose only naming field
  is `full_name`/`email`, or whose only status field is `status`, would emit a null `name`/`state`
  here where legacy would populate it. Documented scope narrowing.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-200,
  default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides. The engine's `cursor` paginator has no config-driven page-size or request-count-cap knob;
  neither is declared in `spec.json`, and this bundle sends Pylon's own default (`limit=100`) as a
  static per-stream query value.
- **Legacy's `raw` passthrough field is not modeled.** Legacy's `mapRecord` stashes the entire raw
  item under a `raw` key on every emitted record; this bundle's schema projection keeps only the
  declared parity fields (`id`, `title`, `name`, `state`, `email`, `created_at`, `updated_at`).
