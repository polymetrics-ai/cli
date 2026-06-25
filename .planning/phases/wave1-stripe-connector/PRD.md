# PRD — Wave 1 batch 1: Stripe connector (declarative-HTTP reference)

## Problem
Wave 0 delivered the per-system architecture + connsdk + GitHub (a *custom* connector) + DuckDB.
The ~500 declarative-HTTP SaaS connectors need a proven **template built on connsdk**. GitHub kept
its bespoke HTTP code; we now need one connector that exercises connsdk end-to-end (Bearer auth,
cursor pagination, record extraction, incremental cursor, reverse-ETL write) so the rest copy it.

## Goal
Implement `stripe` as `internal/connectors/stripe/` — the reference declarative-HTTP per-system
connector — on `connsdk`. Read (ETL) a core set of streams + reverse-ETL write actions, self-register
via `registryset`, full TDD with `httptest`. Flip `source-stripe` catalog entry to enabled with
`pm_connector_name=stripe`.

## Scope (batch 1)
- **Read streams** (incremental by Stripe `created` where supported): customers, charges, invoices,
  subscriptions, products.
- **Auth**: `Authorization: Bearer <client_secret>` (the Stripe secret API key).
- **Pagination**: Stripe list pages (`limit`, `starting_after=<last id>`, stop on `has_more=false`),
  records under `data[]`.
- **Reverse-ETL write**: a small allow-listed action set (e.g. `create_customer`) with
  `ValidateWrite` + `DryRunWrite` (plan→preview→approve→execute preserved).
- **Query**: not a Stripe-native capability; rely on the DuckDB warehouse for analysis (no Querier).

## Non-Goals
- All Stripe streams (batch 1 is a representative core; more streams later).
- Stripe Connect multi-account beyond passing `account_id` through.
- Other Wave 1 connectors (separate batches).

## Success Metrics
- `stripe` registered; `pm connectors inspect stripe --json` → kind Connector, read+write.
- httptest-backed read returns mapped records with incremental cursor advance; pagination followed.
- `ValidateWrite` accepts allow-listed actions, rejects others; `make verify` green.

## Constraints
- Secrets never logged/printed; reverse-ETL stays plan→preview→approve→execute.
- Built on connsdk (Requester/Bearer/Extractor/cursor state); no per-connector HTTP reinvention.
- `make verify` green at every gate.
