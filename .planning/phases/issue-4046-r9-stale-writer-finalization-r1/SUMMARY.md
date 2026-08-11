---
coverage:
  - id: D1
    description: A stale transport checkpoint writer returns the matching non-zero failed run and leaves it durably failed after reopen.
    verification:
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestRunETLTransportStaleWriterFinalizesLosingRun
        status: pass
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestRunETLTransportStaleWriterFailureSurvivesReopen
        status: pass
    human_judgment: false
  - id: D2
    description: Only the typed transport stream-state conflict can finalize from current locked state; ordinary revision conflicts and terminal targets remain protected.
    verification:
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestFailRunTransportConflictPreservesLatestConcurrentState
        status: pass
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestFailRunRetainsRevisionGuardWithoutTransportConflict
        status: pass
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestFailRunTransportConflictRequiresRunningTarget
        status: pass
    human_judgment: false
  - id: D3
    description: Winner stream state and unrelated stream/checkpoint/run data remain unchanged through the conflict finalization path.
    verification:
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestRunETLTransportStaleWriterFinalizesLosingRun
        status: pass
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestFailRunTransportConflictPreservesLatestConcurrentState x20
        status: pass
    human_judgment: false
  - id: D4
    description: The behavior remains correct after acknowledgement-time cancellation and across all seven canonical modes without weakening R7/R8 CAS or source identity.
    verification:
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestRunETLTransportStaleWriterFinalizesAfterCancellation x10
        status: pass
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: TestRunETLTransportStaleWriterFinalizesLosingRunForAllModes
        status: pass
      - kind: unit
        ref: internal/app/transport_dispatch_test.go: R7/R8 nine-test regression suite
        status: pass
    human_judgment: false
---

# Summary — #4046 R9 stale-writer run finalization

## Delivered

- `App.failRun` now recognizes only `errTransportStreamStateConflict` through `errors.Is` and, for that typed error alone, finalizes the matching current `running` run under the JSON-store lock.
- The terminalization changes only that run to `failed`, retains redaction and a completion timestamp, and returns it only after a successful or may-have-committed state outcome; a definite pre-rename persistence failure returns `Run{}` while preserving the typed error chain.
- The ordinary whole-state revision guard remains in place for every non-conflict failure. A conflict cannot replace a missing or already terminal target run.
- No checkpoint retry, checkpoint overwrite, stale-state assignment, `transport_dispatch.go` behavior, source identity, or R7/R8 per-stream CAS behavior changed.

## TDD and commits

- `b2e5a7d25` — planning/TDD evidence before production edits.
- `a51734990` — behavioral RED: two real `App` instances showed the zero returned loser and reopened durable `running` leak while winner/unrelated state held.
- `874a3f400` — minimal typed-conflict-only `failRun` transition; the matching witness turned green.
- `ceb5ea00d` — deterministic restart, intervening writer, cancellation, all-seven-mode, ordinary-error guard, and terminal-target coverage.

## Manual-GSD fallback

The issue phase is not addressable in the archived numeric roadmap, and the canonical delivery contract forbids role spawning. The required `discuss-phase → plan-phase --tdd → execute-phase → verify-work → code-review` lifecycle therefore ran inline. `CONTEXT.md`, `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `UAT.md`, and `REVIEW.md` record its decisions, RED/GREEN evidence, automated acceptance, and review disposition.

## Honest boundary

This records only deterministic package-local fake-backed behavior and local repository gates. It does not claim a live provider, credential, network, warehouse, container, external service, CI result, automated-review result, or certification. The fresh child-local no-mistakes and stacked-PR steps remain pending.
