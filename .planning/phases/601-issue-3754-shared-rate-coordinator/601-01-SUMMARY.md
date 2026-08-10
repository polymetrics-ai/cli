---
phase: 601
plan: 01
coverage:
  - id: D1
    description: Matching rate-limit policies reserve atomically through the dependency-free process-local coordinator, while a blocked batch charges neither an earlier policy nor the shared lease.
    verification:
      - kind: unit
        ref: internal/coordination/rate_budget_coordinator_test.go:TestRateBudgetReserveBatchAllOrNothing; TestRateBudgetFinishIsIdempotentAndReleasesLease; TestRateBudgetRejectsMismatchedRegisteredFingerprint
        status: pass
      - kind: unit
        ref: internal/coordination/rate_limits_test.go
        status: pass
    human_judgment: false
  - id: D2
    description: A same-host UDS owner/client has closed readiness, decision, and finish operations, correct permissions, per-run epoch rejection, scope independence, and normal-close residue cleanup.
    verification:
      - kind: integration
        ref: internal/coordination/unix_rate_budget_coordinator_test.go:TestUnixRateBudgetCoordinatorMultiProcessTinyBudget; TestSharedRateBudgetScopesRemainIndependent; TestSharedRateBudgetOwnerCrashFailsClosed
        status: pass
      - kind: unit
        ref: internal/coordination/rate_budget_coordinator_test.go:TestRateBudgetLeaseTTLFreesConcurrencyWithoutDroppingLateObservation
        status: pass
    human_judgment: false
  - id: D3
    description: require_shared fails closed before transport when no ready coordinator exists or a known shared wait cannot meet the caller deadline.
    verification:
      - kind: integration
        ref: internal/connectors/engine/rate_limit_runtime_test.go:TestRequireSharedRefusesWithoutCoordinatorBeforeSend; TestSharedRateBudgetDeadlineTooShortDoesNotSend
        status: pass
    human_judgment: false
  - id: D4
    description: Requester leased admission finishes completed and indeterminate logical sends with typed observations, preserving existing non-leased admission behavior.
    verification:
      - kind: unit
        ref: internal/connectors/connsdk/http_test.go:TestRequesterLeaseAdmissionFinishesCompletedSend; TestRequesterLeaseAdmissionFinishesIndeterminateTransportSend
        status: pass
      - kind: unit
        ref: internal/connectors/connsdk and internal/connectors/engine focused suites
        status: pass
    human_judgment: false
  - id: D5
    description: The child remains bounded to generic coordination/runtime seams: no provider declarations, GraphQL/provider code, CLI/docs surface, dependency, external daemon, or credentialed live check is introduced.
    verification:
      - kind: other
        ref: changed-path audit, git diff --check, make connector-boundary, make connectorgen-validate, make connectorgen-surface-sync, make docs-check
        status: pass
    human_judgment: false
---

# Summary — #3754 optional same-host shared rate-budget coordinator

The residual shared half of #3754 is now implemented without replacing the
ordinary dependency-free process-local registry. A closed internal
`BudgetCoordinator` reserves every matching policy in one batch and returns
one opaque lease. `Finish` is idempotent, releases only concurrency, retains
indeterminate consumption, and applies the existing typed tighter response
observations.

The optional same-host backend is a run-owned Unix-domain-socket owner/client:
it uses a private short-lived directory and socket, versioned readiness, a
bounded ready/decide/finish protocol, opaque policy fingerprint plus #3863
scope identity, and fail-closed owner-loss/epoch behavior. It adds no external
service, durable provider-budget truth, cross-host protocol, provider policy,
GraphQL/provider path, command surface, help/manual/website change, or new
dependency.

The mandatory test proof is real local multiprocess execution: eight helper
test binaries released together consume a shared budget of three as exactly
three grants and five typed unattempted blocks; the process-local control
admits all eight. The same test proves permission modes and normal-close
cleanup without putting endpoint or epoch data in evidence.

GSD discuss, plan, execute, and verify are performed inline under the
repository-required no-spawn fallback. The actual RED compile boundary and
GREEN/local-gate evidence are in `TDD-LEDGER.md`; #4025 remains open for its
separate GSD aggregate-state defect owner and is neither implementation scope
nor a coordinator correction round.
