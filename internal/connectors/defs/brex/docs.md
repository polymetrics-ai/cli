# Overview

Brex is a wave2 fan-out declarative-HTTP migration. It reads Brex card transactions, users, card
expenses, vendors, and budgets through the Brex Platform REST API
(`GET https://platform.brexapis.com/...`). This bundle migrates `internal/connectors/brex` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a Brex API user access token via the `user_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <user_token>`, matching legacy's `connsdk.Bearer(secret)`) and is never
logged. `base_url` defaults to `https://platform.brexapis.com` and may be overridden for
tests/proxies.

## Streams notes

All 5 streams (`transactions`, `users`, `expenses`, `vendors`, `budgets`) share the same list-page
shape: records live at `items`, and pagination follows Brex's own `cursor`/`next_cursor` convention
(`pagination.type: cursor` with `token_path: next_cursor`, `cursor_param: cursor`) — the next
page's `cursor` query param is the previous response's `next_cursor` value, and pagination stops
when `next_cursor` is null/absent, matching legacy's `harvest` loop exactly. Every request sends
`limit=100` (matches legacy's `brexDefaultPageSize`) via each stream's static `query: {"limit":
"100"}`.

`transactions` (`/v2/transactions/card/primary`) and `expenses` (`/v1/expenses/card`) support an
incremental datetime lower bound: `posted_at_start`/`purchased_at_start` respectively
(`incremental.request_param`, `param_format: rfc3339`), computed from the persisted cursor or, on a
fresh sync, the RFC3339 `start_date` config value — matching legacy's `incrementalStart`. `users`,
`vendors`, and `budgets` have no incremental cursor (legacy's own `incrementalStart` returns `("",
"")` for these streams; this bundle declares no `incremental` block for them).

Budgets use `budget_id` (not `id`) as the primary key, matching Brex's own budget object shape and
legacy's `PrimaryKey: []string{"budget_id"}`.

## Write actions & risks

None. Brex is a read-only financial source in this connector (legacy's own package doc: "no safe
reverse-ETL writes"); `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **The fresh-sync `start_date` value is sent in full RFC3339 form (with trailing `Z`/offset)
  rather than legacy's stripped `2006-01-02T15:04:05` form.** Legacy's `incrementalStart`
  reformats an RFC3339 `start_date` config value before sending it as `posted_at_start`/
  `purchased_at_start`, but passes the PERSISTED CURSOR value through unchanged on every
  subsequent (repeat) sync — meaning Brex's real API already accepts the raw RFC3339-with-offset
  form on the much more common repeat-sync path (the cursor value itself is echoed from a prior
  response's own timestamp field, in full RFC3339 form). This bundle's `param_format: rfc3339`
  (the engine's default, sent verbatim) therefore only differs from legacy on the FIRST full sync
  of a stream when `start_date` is set and no cursor exists yet, and even then only differs in
  representation (full offset vs local-time-stripped), not in the moment in time being requested.
  The engine's `param_format` dialect has exactly 4 named formats (`rfc3339`, `unix_seconds`,
  `date`, `github_date_range`); none matches legacy's bespoke stripped-datetime string, and adding
  a 5th format for one connector's cosmetic wire-format preference is out of scope for this
  migration. Documented per `docs/migration/conventions.md` §5's parity-deviation ledger:
  ACCEPTABLE (never changes which records are included, only the on-the-wire string
  representation of the lower bound on the first sync).
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (1-100,
  default 100) and `max_pages` as config-driven overrides read fresh on every `Read` call. The
  engine's `cursor`/`token_path` paginator has no page-size knob at all (the `limit=100` value is a
  static per-stream query literal, matching stripe's `limit=100` precedent), and no
  `MaxPages`-equivalent config-driven override either (pagination is bounded only by the
  `next_cursor` stop signal, matching Brex's own real termination behavior). `spec.json` still
  declares `page_size`/`max_pages` for documentation continuity with legacy's config surface, but
  neither is wired into any template.
- Legacy's fixture-mode-only fields (`connector`, `fixture` static markers stamped only under
  `config.mode == "fixture"`) are not modeled; this bundle's schemas and parity target the live
  wire shape only, matching this repo's established convention for a legacy in-code fixture path
  now superseded by the engine's own conformance/fixture-replay harness.
