# Overview

Looker is a read-only declarative-HTTP migration. It reads Looker users, groups, folders, looks,
and dashboards through the Looker API 4.0 (`GET <base_url>/...`). This bundle targets capability
parity with `internal/connectors/looker` (the hand-written connector it migrates); the legacy
package stays registered and unchanged until wave6's registry flip.

## Auth setup

Two credential shapes are accepted, matching legacy's own precedence exactly:

1. **`access_token`** (a pre-obtained Looker API access token) — when set, used directly as a
   Bearer token (`Authorization: Bearer <access_token>`). Takes priority over client_id/
   client_secret when both are configured, matching legacy's `requester`'s `accessToken(cfg)`
   trim-and-check-first order.
2. **`client_id` + `client_secret`** (Looker API3 credentials) — when `access_token` is unset, the
   connector logs in via `POST <token_url>` (form-encoded `client_id`/`client_secret`, plus the
   engine's own `grant_type=client_credentials`/optional `scope` fields that Looker's login
   endpoint ignores) and uses the returned `access_token` as a cached, auto-refreshing Bearer
   token — matching legacy's `loginAuth` exactly (a 60-second-before-expiry refresh margin,
   `expires_in` defaulting to 1 hour when absent/non-positive).

`token_url` defaults to `<base_url>/login` (Looker's own convention: the login endpoint is always
a sibling of the versioned API base), matching legacy's `tokenURL` fallback exactly. An explicit
`token_url` config override (for tests or proxies pointing at a different login path) takes
priority when set — modeled as a `when`-gated auth candidate ahead of the derived-default
candidate, since the engine's `spec.json` `default` mechanism only materializes a fixed literal,
not one derived from another config value (`base_url`).

None of `access_token`/`client_id`/`client_secret` is required by `spec.json` (legacy accepts
either shape), so `requireCredentials`'s "one of the two shapes must be present" rule is
approximated by the auth candidate list itself: if neither shape resolves, no candidate's `when`
matches except the final unconditional `oauth2_client_credentials` candidate, which then hard-fails
at request time on the unresolved `client_id`/`client_secret` — an unauthenticated request is never
sent (see the parity-deviation ledger entry below).

## Streams notes

All 5 streams (`users`, `groups`, `folders`, `looks`, `dashboards`) are `GET` list endpoints
returning a bare JSON array (`records.path: ""`), matching legacy's `streamEndpoints` table
exactly. Pagination is `offset_limit` (`limit`/`offset` query params, `page_size: 100` — Looker's
API-documented max and legacy's `defaultPageSize`), stopping on a short page exactly like legacy's
`harvest` (`len(records) < pageSize`).

`looks` and `dashboards` share legacy's identical record shape (`lookRecord`/`dashboardRecord` are
the same function in legacy) — this bundle keeps them as two schemas with identical fields rather
than collapsing them, since the engine dialect has no "alias one stream's schema to another."

None of the 5 streams is incremental — legacy's `streamEndpoints` declares no incremental filter
for any of them (the Looker list endpoints used here have no server-side updated-since filter), so
no stream declares an `incremental` block, matching legacy's actual behavior (a full sync every
time), not narrowed further.

## Write actions & risks

None. Looker is a read-only source connector (`capabilities.write: false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Runtime-configurable `page_size`/`max_pages` are not modeled.** Legacy exposes `page_size`
  (1-500, default 100) and `max_pages` (0/all/unlimited = unbounded) as per-sync config overrides
  read fresh on every `Read` call. The engine's `offset_limit` paginator's page size
  (`PaginationSpec.PageSize`) and the read loop's `MaxPages` cap are both plain JSON literals in
  `streams.json`, not template-resolvable from `config.*` at read time — there is no mechanism in
  this dialect to make either genuinely config-driven (the same structural gap searxng's `docs.md`
  documents for its own `page_size`/`max_pages`). `page_size` is fixed to 100 (Looker's own
  documented page-size ceiling and legacy's default), `max_pages` is left unset (unbounded,
  matching legacy's default `max_pages=0`/`all`/`unlimited` behavior). Since neither is genuinely
  wireable, neither is declared in `spec.json` (F6: a declared-but-unwireable key is worse than an
  absent one).
- **Field-name fallback aliases are not modeled.** Legacy's `userRecord`/`lookRecord`/
  `dashboardRecord` defensively try alternate camelCase field names
  (`display_name`/`displayName`/`name` for users; `title`/`name` and `folder_id`/`folderId` for
  looks/dashboards) via a `first(item, keys...)` helper, in case a differently-shaped API response
  ever used the alternate casing. Looker API 4.0's real, documented wire shape for these resources
  uses `display_name`/`title`/`folder_id` exclusively (confirmed by legacy's own test fixtures,
  which never exercise the alternate-casing branch) — this bundle's schemas and `records.path: ""`
  passthrough project those fields directly, with no computed_fields rename needed. The
  alternate-casing defensive fallback is not modeled: the dialect has no coalesce/first-match
  filter, and this never diverges for any real Looker API response, only for a hypothetical
  differently-cased shape legacy's own tests never observed. Documented scope narrowing, not a
  behavior change for any real input.
- **`requireCredentials`'s explicit pre-flight message is not reproduced verbatim.** Legacy fails
  fast with a specific error ("looker connector requires secret access_token or client_id and
  client_secret") before ever issuing a request. This bundle instead lets the unconditional
  fallback `oauth2_client_credentials` auth candidate attempt a token exchange with empty
  `client_id`/`client_secret` values, which Looker's real login endpoint rejects with 401 (or the
  fixture/replay harness's own request-shape mismatch) — no request is ever authenticated
  successfully with missing credentials in either case, but the error surfaced differs from
  legacy's upfront validation message. Acceptable: never changes accepted-input behavior (a fully
  unconfigured connector still fails, just with a different message), matching the meta-rule (§5).
