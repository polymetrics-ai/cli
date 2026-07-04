# Overview

Brex is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full practical
read/write surface. It reads Brex card transactions, users, card expenses, vendors, budgets,
departments, locations, titles, legal entities, cards, card/cash accounts, card statements, linked
bank accounts, transfers, and webhooks through the Brex Platform REST API
(`https://platform.brexapis.com/...`), and writes vendor/department/location/title/user/card/
expense/webhook lifecycle mutations. This bundle originally migrated `internal/connectors/brex`
(the hand-written connector, read-only); the legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide a Brex API user access token via the `user_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <user_token>`, matching legacy's `connsdk.Bearer(secret)`) and is never
logged. `base_url` defaults to `https://platform.brexapis.com` — Brex's pre-unification API host,
still live and documented by third-party integrations (Retool, Fivetran) alongside the
newer-branded `https://api.brex.com`; both serve the identical `/v1`/`/v2` path surface this
bundle targets, so the existing default is kept unchanged (parity-preserving: switching it would
be an accepted-input behavior change to every already-migrated stream, not a Pass B addition) —
and may be overridden for tests/proxies.

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

**Pass B additions.** `departments`, `locations`, `titles`, `legal_entities`, `cards`,
`accounts_cash`, `card_statements`, `linked_accounts`, `transfers`, and `webhooks` all follow the
identical `cursor`/`next_cursor` pagination convention as the 5 original streams (`query: {"limit":
"100"}`, `records.path: "items"`), none have an incremental cursor (Brex's Team/Payments/Webhooks
API list endpoints accept no server-side modified-since filter). `accounts_card`
(`GET /v2/accounts/card`) is the one exception: it returns a bare top-level JSON array with NO
`items`/`next_cursor` pagination envelope at all (`records.path: ""`, no `pagination` block,
matching the engine's `none` default) — Brex's own OpenAPI spec declares this endpoint's response
as `type: array` directly, unlike every other list endpoint in this bundle.

## Write actions & risks

Fourteen write actions, none present in legacy (legacy shipped `capabilities.write: false`):

- **`update_vendor`** / **`delete_vendor`** — vendor lifecycle (no `create_vendor`; see Known
  limits' Idempotency-Key gap below).
- **`create_department`** / **`create_location`** / **`create_title`** — org-directory entry
  creation (no update/delete endpoints exist in the API for these three resources).
- **`create_user`** / **`update_user`** — user directory lifecycle. `create_user` sends a real
  invitation email; `update_user`'s `status` field can revoke account access.
- **`update_card`** / **`lock_card`** / **`unlock_card`** / **`terminate_card`** — card lifecycle
  (no `create_card`; see Known limits). `lock_card`/`unlock_card` take effect on the physical/
  virtual card immediately; `terminate_card` is irreversible.
- **`update_expense`** — card-expense memo update (the only mutable field Brex exposes on an
  already-posted expense).
- **`update_webhook`** / **`delete_webhook`** — webhook subscription lifecycle (no
  `create_webhook`; see Known limits). Updating a webhook's `url`/`event_types`/`status` redirects
  live event delivery immediately; `delete_webhook` is irreversible.

Every action's per-record `risk` string in `writes.json` is the authoritative, reviewable summary;
`metadata.json`'s `risk.write`/`risk.approval` roll these up for the connector as a whole.

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
- **`ENGINE_GAP`: no write action can attach a request header, so every Brex write endpoint whose
  OpenAPI definition marks `Idempotency-Key` `required: true` is unreachable as a declarative write
  action.** `engine.WriteAction` (bundle.go) has no `headers` field, and `write.go`'s
  `executeWriteRecord` calls `rt.Requester.Do`/`DoForm` with an unconditional `nil` headers
  argument — unlike reads (`read.go`'s `resolveHeaders` reads `b.HTTP.Headers`), a write request in
  this engine can NEVER carry any header at all, dynamic or static. Brex requires this header
  specifically (not merely accepts it) on `create_vendor`, `create_transfer`,
  `create_incoming_transfer`, `create_card`, `create_budget`/`update_budget` (both v1 and v2),
  `create_spend_limit`/`update_spend_limit`, and `create_webhook`/`create_webhook_group` — sending
  any of these without the header is rejected by Brex's own API (`developer.brex.com/guides/
  idempotency.md`: "For endpoints where erroneous duplicate processing would be especially bad...
  we require an idempotency key"), so there is no honest way to implement these as plain
  declarative writes: a fixed/static header value would violate Brex's own collision-avoidance
  contract (the same key reused across genuinely-different requests), and there is no dialect
  mechanism to mint a fresh value (e.g. a UUID) per write call. All of these are excluded in
  `api_surface.json` as `requires_elevated_scope` rather than approximated. The REST of each
  affected resource's lifecycle (read, update, delete, lock/unlock/terminate) is fully modeled
  where Brex's own spec marks the header optional on that specific operation — only the
  create-shaped, highest-consequence half of the vendor/transfer/card/budget/webhook surface is
  affected. Closing this gap requires an engine dialect addition (a `headers` field on
  `WriteAction`, or a `uuid4`-shaped interpolation filter analogous to `const:<value>`) — out of
  scope for a Pass B connector-only expansion per this task's own instructions (no engine changes,
  no new hook packages).
