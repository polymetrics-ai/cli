# PRD — Action Step (Phase 1)

## Goal
Add a first-class `action` step kind to the flow engine that is a thin adapter over
`internal/app` `PlanReverseETL → RunReverseETL`. Lift single-step approval to a
**flow-level token** with an optional `--per-action` flag. Implement all seven must-have
safety features test-first, with an `httptest.Server` as the fake destination.

## Non-negotiables (from prompt)
- Go stdlib only for new logic; reuse `connsdk` retry that is already in the tree.
- No new go.mod dependency — HUMAN GATE if one is needed.
- All behaviour-adding code requires a failing test first (red evidence in TDD-LEDGER.md).
- Real network writes remain behind `plan → preview → approval-token → execute`.
- `make verify` green at phase end.

## Must-have features (all test-first)
1. **Idempotent writes** — keyed by deterministic record id; re-run never sends the same record twice.
2. **Identity mapping** — warehouse PK ↔ external system id, persisted in `internal/state` JSON store.
3. **Dedupe/merge rules** — on email / domain / external-id before any write attempt.
4. **Rate-limit handling** — 429-aware with exponential backoff + jitter; reuse connsdk Requester retry machinery pattern (inline it for the action layer since connsdk is an http layer; implement equivalent backoff inline).
5. **Dead-letter queue + bounded retries** — failed records written to `.polymetrics/dlq/<flow>/<step>/<run>.ndjson`; never silently dropped; capped at `MaxRetries` attempts.
6. **Schema-drift detection** — compare live schema snapshot vs stored snapshot at step start; on breaking change PAUSE before any write and return `ErrSchemaDrift`.
7. **Receipts/audit** — every action step writes a receipt record (redacted) to `internal/ledger`.

## Out of scope for this phase
- RLM, scheduling, agent mode.
- Postgres / Dragonfly-backed state (runtime-backed mode stays optional).
