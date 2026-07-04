# Overview

You Need A Budget (YNAB) is a declarative bundle migrated from
`internal/connectors/you-need-a-budget-ynab` (the hand-written legacy connector, which stays
registered and unchanged until wave6's registry flip). It reads YNAB budgets, accounts,
categories, payees, months, transactions, and scheduled transactions, and writes
transaction/account/category/payee/scheduled-transaction mutations, through the YNAB REST API v1.

## Auth setup

Provide a YNAB personal access token via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `budget_id` (config, optional) scopes the
`accounts`/`categories`/`payees`/`months`/`transactions`/`scheduled_transactions` streams and every
write action to a specific budget, and defaults to YNAB's own `last-used` budget alias when unset
(matching legacy's `budgetPath` fallback exactly) via `spec.json`'s `"default": "last-used"`
materialization.

## Streams notes

All 7 streams share YNAB's `{"data": {...}}` response envelope. `budgets` (`GET /budgets`, records
at `data.budgets`) has no path parameters. The other 6 streams are scoped to `budget_id`,
substituted into the path (urlencoded by default): `accounts` (`data.accounts`), `categories`
(`data.category_groups` — YNAB's categories endpoint returns category GROUPS, each with a nested
`categories[]` array; this bundle publishes one record per group, matching the API's own top-level
list shape, not one record per leaf category), `payees` (`data.payees`), `months`
(`data.months` — a `MonthSummary` list; the richer nested per-month `categories[]` breakdown is
only on the single-month detail endpoint, excluded as `duplicate_of` per `api_surface.json`),
`transactions` (`data.transactions`), and `scheduled_transactions`
(`data.scheduled_transactions`).

**`/budgets` vs `/plans` path naming**: YNAB renamed its primary documented resource path from
`/budgets/{budget_id}` to `/plans/{plan_id}` in API v1.79.0 (2026-03-25), including renaming
response keys (`budgets`→`plans`, `budget`→`plan`). The `/budgets/{budget_id}` paths remain fully
functional as an undocumented backward-compatible alias returning the original response key names.
This bundle deliberately keeps every path/response-key on the `/budgets` shape — matching legacy's
own paths exactly (zero accepted-input behavior change) — rather than migrating to `/plans`, which
would be a new-terminology adoption decision, not a mechanical parity port. Revisit if YNAB ever
deprecates the alias outright.

All 7 streams project in `"passthrough"` mode (every native YNAB field survives, matching legacy's
`mapRecord`, which copies every raw key into the emitted record) plus per-stream `computed_fields`
that alias legacy's derived convenience fields on top: `budgets`/`accounts`/`categories`/`payees`
alias `updated_at` from each object's own most-recent-change-bearing field (`last_modified_on` for
budgets/accounts; `name` is used as `categories`/`payees`' `updated_at` surrogate since neither
`CategoryGroup` nor `Payee` publishes any modification timestamp at all — YNAB's API has no
`updated_at`-shaped field on either resource; documented as a known limit below, unchanged
scope-narrowing carried over from the pre-Pass-B bundle's identical treatment of `accounts`).
`months` synthesizes both `id` and `updated_at` from the native `month` field (`MonthSummary` has
no native `id` at all — `month`, e.g. `"2026-01-01"`, is its own natural key). `transactions`
aliases `name` from `payee_name` and `updated_at` from `date`; `scheduled_transactions` aliases
`name` from `payee_name` and `updated_at` from `date_next` (its nearest analogous
recently-changed-shaped field, since scheduled transactions have no `last_modified_on` either).
Primary key is `["id"]` for every stream except `months`, whose primary key is the computed `id`
alias of `month` (no native `id` field exists on `MonthSummary`). No pagination is declared on any
stream — legacy issues exactly one request per stream with no pager at all, matching the real
YNAB API's own behavior (none of these list endpoints paginate; each returns its full collection in
one response, bounded by `server_knowledge`-based delta sync which this bundle does not use).

Optional `since_date` (YYYY-MM-DD) and `limit` config values are sent as query params on
`budgets`/`accounts`/`transactions` reads via the opt-in optional-query dialect
(`omit_when_absent: true`), matching legacy's `baseQuery`, which attaches both params
unconditionally (including on the `budgets` request, where YNAB's API simply ignores them —
legacy's own behavior, reproduced verbatim). `categories`/`payees`/`months`/
`scheduled_transactions` are new Pass-B streams with no legacy analog, so they do not carry
`since_date`/`limit` (YNAB's real API does not document either param for these 4 endpoints at all
— `since_date` is `accounts`/`transactions`-specific server-side incremental filtering, and
`categories`/`payees`/`months`/`scheduled_transactions` have no equivalent).

## Write actions & risks

10 write actions, all requiring approval (`capabilities.write: true`):

- `create_transaction` / `update_transaction` / `delete_transaction` (`POST`/`PUT`/`DELETE
  /budgets/{budget_id}/transactions[/{id}]`): creates, mutates, or deletes a single budget
  transaction. `delete_transaction`'s `missing_ok_status: [404]` treats an already-deleted
  transaction id as a successful (idempotent) delete.
- `create_account` (`POST /budgets/{budget_id}/accounts`): creates a new budget account with an
  opening balance. YNAB's API has no delete-account endpoint at all, so this write has no
  companion delete action — irreversible via the API.
- `create_category` / `update_category` (`POST`/`PATCH .../categories[/{id}]`): creates a category
  within an existing category group, or renames/re-notes/re-goals an existing one.
- `update_month_category` (`PATCH .../months/{month}/categories/{category_id}`): reassigns
  (budgets) an amount to a category for one specific month — YNAB's actual "move money" primitive.
- `create_payee` / `update_payee` (`POST`/`PATCH .../payees[/{id}]`): creates or renames a payee;
  renaming also renames every transaction and the shared cross-app payee-matching history.
- `create_scheduled_transaction` / `delete_scheduled_transaction` (`POST`/`DELETE
  .../scheduled_transactions[/{id}]`): creates or deletes a recurring scheduled transaction that
  auto-posts future budget transactions on its own cadence.

All 10 actions use YNAB's own top-level wrapper-key request-body convention (`{"transaction":
{...}}`, `{"account": {...}}`, `{"category": {...}}`, `{"payee": {...}}`,
`{"scheduled_transaction": {...}}`) — the engine's write dialect has no nested-wrapper body
construction primitive, so each action's `record_schema` declares the wrapper key itself as a
required nested-object field, and the caller-supplied record already carries that shape
(`body_type: json`'s default body-from-record-fields construction then emits it byte-for-byte; see
teamwork's `create_project`/bitly's `create_qr_code.destination` for the identical sanctioned
pattern, documented in `docs/migration/conventions.md`).

## Known limits

- Documented parity deviation (narrower fallback chain, single computed_fields reference only):
  legacy's `transactions` stream derives its convenience `name` field from
  `firstValue(in, ["payee_name", "memo"])` — a 2-key fallback chain trying `payee_name` first,
  falling back to `memo` only when `payee_name` is absent/nil. The engine's `computed_fields`
  dialect has no coalesce/first-of filter (only a single bare `{{ record.<path> }}` reference or a
  filter chain over ONE reference), so this bundle's `name` computed field only aliases
  `payee_name` — a transaction with a null `payee_name` (e.g. a split transaction with no assigned
  payee) emits `name: null` here where legacy would have back-filled the transaction's `memo` text
  instead. This is a genuine, documented narrowing (not cosmetic): it changes emitted `name` data
  for that specific edge-case input. Scope-narrowed rather than blocked because it affects one
  convenience alias field on one stream, not the record's real YNAB data (every native field,
  including `memo` itself, still survives via `passthrough` projection) — see
  `docs/migration/conventions.md` §5's meta-rule and searxng's analogous documented-narrowing
  precedent (ledger item 7).
- Similarly, legacy's `accounts` stream tries `firstValue(in, ["last_modified_on", "updated_at"])`
  for its `updated_at` alias; this bundle aliases only `last_modified_on`. YNAB's real Account
  object has neither field natively (only `last_reconciled_at`), so both legacy and this bundle
  resolve `updated_at` to absent/null for every real account record — this narrowing has no actual
  effect on any real API response and is recorded for completeness only.
- No pagination or `max_pages`/`page_size` config is modeled on any stream — none of the 7
  documented list endpoints this bundle covers actually paginate; YNAB's own change-tracking
  mechanism is `server_knowledge`-based delta sync (`last_knowledge_of_server` query param /
  `server_knowledge` response field), not offset/cursor pagination. This bundle does not implement
  delta sync (an `ENGINE_GAP`-adjacent full-vs-incremental design question, not a pagination gap);
  every read is a full snapshot of the current collection.
- `create_transaction`'s real API also accepts a bulk multi-transaction array body
  (`{"transactions": [...]}` alongside the modeled single-`{"transaction": {...}}"` shape) and a
  dedicated `/transactions/import` bank-resync trigger; neither is modeled (`api_surface.json`:
  `duplicate_of`/`out_of_scope`) — the single-record write dialect emits exactly one record's body
  per write call, so a genuinely bulk array-body endpoint has no natural per-record mapping here.
- Category-group create/update (`POST`/`PATCH /budgets/{budget_id}/category_groups[/{id}]`) is not
  modeled: a category group has no dedicated read stream of its own in this bundle (it is read
  only as the `categories` stream's group-level record, alongside its nested nominal `categories[]`
  array) — see `api_surface.json`'s `out_of_scope` entries.
- `payee_locations` (GPS coordinates auto-captured from mobile-app payee entry) is out of scope —
  convenience geolocation metadata, not core budget business data.
- `money_movements`/`money_movement_groups` (a newer aggregate view over existing
  transactions/transfers) are out of scope as derived, non-source data.
