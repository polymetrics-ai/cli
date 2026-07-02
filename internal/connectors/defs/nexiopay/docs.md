# Overview

Nexio Pay is a read-only connector for the Nexio payments platform. This bundle migrates
`internal/connectors/nexiopay` (the hand-written legacy connector) to a declarative Tier-1 bundle at
full capability parity: it reads the same six streams (`card_tokens`, `recipients`, `spendbacks`,
`payment_types`, `terminal_list`, `user`) through the same Nexio v3 REST endpoints, using HTTP Basic
auth and Nexio's offset/limit pagination. The legacy package stays registered and unchanged until
the wave6 registry flip.

## Auth setup

Provide `username` and `api_key` (both `x-secret`); they are sent as HTTP Basic auth
(`username:api_key`), matching legacy's `connsdk.Basic(username, apiKey)`. Both are never logged.

Legacy resolves `username` from `cfg.Secrets["username"]` first, falling back to
`cfg.Config["username"]` for compatibility. This bundle models `username` as a secret-only field
(the declarative `auth` dialect references exactly one path per field) — the common/documented
credential shape (both values supplied as secrets) is unaffected; a caller relying on the
config-fallback path must supply `username` as a secret instead. See Known limits.

## Streams notes

`card_tokens`, `recipients`, and `spendbacks` share an identical shape: `GET` against the matching
Nexio endpoint, records at `rows`, offset/limit pagination (`limit`/`offset` query params, page size
10, matching legacy's `nexioDefaultPageSize`) declared once on `base.pagination` and inherited by
all three streams. `payment_types`, `terminal_list`, and `user` are unpaginated single-request
reads: `payment_types`/`terminal_list` return a bare top-level JSON array (`records.path: ""`
selects the response root), `user` (`whoAmI`) returns a single JSON object (also `records.path:
""`, which the engine treats as one record) — all three override `pagination: {"type": "none"}` at
the stream level, matching legacy's `endpoint.paginated == false` branch exactly.

## Write actions & risks

None. Nexio's reporting/management endpoints have no safe reverse-ETL write surface;
`capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`ErrUnsupportedOperation` Write stub.

## Known limits

- **`base_url` is now a required-with-default config value, not a subdomain-derived one.** Legacy
  derives the host from a `subdomain` config value (default `"nexiopay"`) as
  `https://api.<subdomain>.com` when no explicit `base_url` override is set. The engine's
  `spec.json` `"default"` materialization mechanism only supports a fixed literal default, not one
  derived from another config field's value (`docs/migration/conventions.md` §3's `spec.json`
  default-materialization note). This bundle declares `base_url`'s default as the literal
  `https://api.nexiopay.com` (the exact URL legacy's own default subdomain produces), so the common
  no-override case is unaffected; a caller that previously set only `subdomain` (never `base_url`)
  to select a non-default Nexio deployment must now set `base_url` directly instead. `subdomain` is
  not declared in `spec.json` (a declared-but-unwireable key is worse than an absent one, F6,
  `docs/migration/conventions.md`).
- **`username`'s config-fallback resolution path is not modeled.** See Auth setup: legacy tries
  `cfg.Secrets["username"]` then `cfg.Config["username"]`; this bundle only models the secrets path.
  Documented scope narrowing, not a silent behavior change for the common (secrets-supplied)
  credential shape.
- **`page_size`/`max_pages` config is not declared.** The `offset_limit` paginator's page size is a
  fixed JSON integer on `base.pagination.page_size` (`PaginationSpec.PageSize`/`MaxPages` are plain
  `int` fields — the dialect has no per-request templating mechanism for either), so neither can be
  overridden per-request the way legacy's `nexioPageSize`/`nexioMaxPages` config parsing allowed. A
  declared-but-unwireable `spec.json` property is worse than an absent one (F6,
  `docs/migration/conventions.md`), so this bundle declares neither; `limit=10` (legacy's own
  default) is always sent, and pagination is unbounded (matching legacy's own `max_pages` default of
  `0`/unlimited when unset), for every read. This mirrors searxng's identical `page_size`/`max_pages`
  scope-narrowing precedent.
