# Overview

You Need A Budget (YNAB) is a read-only declarative bundle migrated from
`internal/connectors/you-need-a-budget-ynab` (the hand-written legacy connector, which stays
registered and unchanged until wave6's registry flip). It reads YNAB budgets, and per-budget
accounts and transactions, through the YNAB REST API v1.

## Auth setup

Provide a YNAB personal access token via the `api_key` secret; it is used only for Bearer auth
(`Authorization: Bearer <api_key>`) and is never logged. `budget_id` (config, optional) scopes the
`accounts`/`transactions` streams to a specific budget and defaults to YNAB's own `last-used`
budget alias when unset (matching legacy's `budgetPath` fallback exactly) via `spec.json`'s
`"default": "last-used"` materialization.

## Streams notes

`budgets` (`GET /budgets`, records at `data.budgets`) has no path parameters. `accounts` (`GET
/budgets/{budget_id}/accounts`, records at `data.accounts`) and `transactions` (`GET
/budgets/{budget_id}/transactions`, records at `data.transactions`) are scoped to `budget_id`,
substituted into the path (urlencoded by default). All 3 streams project in `"passthrough"` mode
(every native YNAB field survives, matching legacy's `mapRecord`, which copies every raw key into
the emitted record) plus per-stream `computed_fields` that alias legacy's derived `id`/`name`/
`updated_at` convenience fields on top: `budgets`/`accounts` alias `updated_at` from the native
`last_modified_on` field (single bare `{{ record.last_modified_on }}` reference — typed extraction
preserves the native string type); `transactions` aliases `name` from `payee_name` and
`updated_at` from `date` (YNAB transactions have no native `name`/`updated_at` field at all).
Primary key is `["id"]` (always native on every YNAB object; legacy's `id`-fallback branch is
therefore always a no-op for real API traffic and is not modeled). No pagination is declared —
legacy issues exactly one request per stream with no pager at all. Optional `since_date`
(YYYY-MM-DD) and `limit` config values are sent as query params on every stream request via the
opt-in optional-query dialect (`omit_when_absent: true`), matching legacy's `baseQuery`, which
attaches both params unconditionally (including on the `budgets` request, where YNAB's API simply
ignores them — legacy's own behavior, reproduced verbatim, not a new no-op param invented here).

## Write actions & risks

None — this connector is read-only (`capabilities.write: false`), matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation` unconditionally.

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
- No pagination or `max_pages`/`page_size` config is modeled — legacy itself has none (a single
  unpaged request per stream).
- Full YNAB API surface (categories, payees, months, scheduled transactions, write mutations) is
  out of scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
