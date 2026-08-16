# Issue #4072 residual context: token budget lifecycle

**Issue:** #4072 — `fix(engine): gate GitHub App token minting through shared rate admission`

**Base:** `integration/4015-mvp-flat-r1` at `ff6a8710199c10f209d9d47cce87e5c8f7c429e6`.

## Locked scope

- PR #4134 already delivers declaration-owned token routing, required-shared
  pre-I/O refusal, no retry, the shared/process-local request path, and the
  Dragonfly proof. This residual must not rebuild or alter those behaviors.
- The missing literal acceptance is a production consumer of
  `connsdk.BudgetCoordinator` at the declared token-send boundary, with one
  `Decide` and one `Finish` for a granted physical mint.
- A refused `Decide` intentionally has no lease to finish. Its specified
  outcome is exactly one `Decide`, zero `Finish`, and zero token sends; the
  test must assert all three rather than leaving completion implicit.
- The engine owns the adapter. Hooks retain the narrow `DeclaredRouteRequest`
  capability; no raw coordinator, HTTP transport, request body, token, JWT,
  or private key enters that contract, coordinator batch, error, or test log.
- The adapter derives its reservation only from the already declaration-
  resolved policy identity, opaque scope key, and declared budgets, then calls
  `Finish` once after the attempted send. Existing requester admission and
  observation remain unchanged.

## Delivery fallback

`scripts/gsd prompt` resolved each required lifecycle command, but the GSD
phase registry accepts roadmap phases rather than this named issue residual.
The task contract also forbids role spawning. Discussion, TDD planning,
execution, verification, and code review therefore run inline and are
recorded in this directory. The launch brief explicitly authorizes autonomous
defaults (`--auto`) and supplies the complete acceptance decision.

## Canonical references

- `internal/connectors/connsdk/rate_budget_coordinator.go` — lifecycle seam.
- `internal/connectors/engine/hooks.go` — declared auth route boundary.
- `internal/connectors/engine/rate_limit_runtime.go` — declaration policy
  resolution and secret-free scopes.
- `internal/connectors/engine/read.go` — auth-runtime construction.
- `internal/connectors/hooks/github/hooks.go` — installation token exchange.
- `internal/connectors/hooks/github/hooks_test.go` — existing pre-I/O and
  no-retry regressions plus the new counter proof.
- `data/production-parity-shared-context.md` (external dispatch context) —
  verification and delivery rules.
- `data/cli-mvp-remaining-delivery-verify-r1/report.md` (external dispatch
  context), section `#4072 — PARTIAL` — precise residual evidence.
