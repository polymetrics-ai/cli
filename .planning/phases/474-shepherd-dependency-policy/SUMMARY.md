# Issue #474 Summary

Status: implementation, exact-head correction, refactor, and parent-declared verification complete.

Ready stacked PR: https://github.com/polymetrics-ai/cli/pull/483

Delivered a three-module, I/O-free policy core covering:

- the complete Shepherd parent lifecycle, guarded transition evidence, correction, durable human
  wait, and terminal states;
- closed-world dependency DAG validation and fail-closed canonical write scopes;
- bounded maximum-cardinality ready selection with running-scope arbitration and read-only
  coexistence;
- transient retry/correction budgets distinct from hard human gates; and
- deterministic reconciliation with one typed repository blocker per `no_spawn` decision.

The correction loop adds authenticated resumable human decisions with terminal abort, a bounded
component scheduler, locale-independent ordering, Darwin/Git conservative scope aliases, coherent
dependency statuses, selected-only mutator isolation, exact hostile-safe DTO validation, correction
advancement guards, and explicit evidence/blocker precedence.

The second correction closes capital sharp-S alias gaps with an explicitly bounded conservative
alias set, cross-validates lifecycle claims against the authoritative queue, snapshots and freezes
caller DTOs before graph traversal, and gives legacy BLOCKED a terminal reconciliation result.

Latest TDD evidence: review-2 RED 35/40 pass with 5 expected failures, final focused 40/40, and full
Shepherd 177/177. Strict no-emit TypeScript passes over all production Shepherd modules against
installed Pi 0.80.6 types; offline RPC discovers `pm-shepherd`.

The parent orchestrator superseded the child-lane full-repository gate while final verification was
running. The attempted `make verify` was intentionally terminated and is recorded as
`cancelled_by_parent_policy`; the phase-equivalent child gate passed completely. Supplemental root
Go vet/test/build also passed. The unsupported `pi --list-extensions` spelling is recorded alongside
the passing Pi 0.80.6 offline RPC discovery substitute.
