---
coverage:
  - id: D1
    description: Only typed verified-invalid authentication fences an opaque cohort.
    verification:
      - kind: unit
        ref: internal/coordination/auth_cohort_test.go:TestAuthCohortCoordinator_OnlyVerifiedInvalidAuthenticationFences
        status: pass
    human_judgment: false
  - id: D2
    description: A fence cancels siblings and allows zero further send admissions.
    verification:
      - kind: unit
        ref: internal/coordination/auth_cohort_test.go:TestAuthCohortCoordinator_VerifiedFailureCancelsSiblingsAndRejectsNewAdmissions
        status: pass
      - kind: unit
        ref: internal/coordination/auth_cohort_test.go:TestAuthCohortCoordinator_RestartAndRaceNeverAdmitAFencedCohort
        status: pass
    human_judgment: false
  - id: D3
    description: Repair starts a healthy epoch, preserves fence audit evidence, and refuses stale members.
    verification:
      - kind: unit
        ref: internal/coordination/auth_cohort_test.go:TestAuthCohortCoordinator_IsolatesCohortsAndRepairCreatesHealthyEpoch
        status: pass
    human_judgment: false
---

# SUMMARY — issue #3865 verified-auth cohort fencing

Implemented a connector-neutral cohort health coordinator using only the pre-existing opaque `connectors.AuthCohortKey`.

- The closed `AuthenticationOutcome` vocabulary fences only `verified_invalid`; unverified invalid responses, timeouts, transport failures, provider failures, unknown values, and verified-healthy results do not fence.
- Fencing persists the current epoch, refuses new admissions, and cancels all active same-cohort contexts after the transition commits.
- `Repair` requires `verified_healthy`, starts a new healthy epoch, cancels old members, preserves `LastFencedEpoch` audit evidence, and leaves stale members refused.
- The injected health-store seam preserves opaque state across a coordinator restart; the in-memory implementation is race-safe and has a deterministic 64-attempt zero-send proof.

The GSD `execute-phase`, `verify-work`, and `code-review` prompts were executed inline under the documented single-worker fallback. See `TDD-LEDGER.md`, `VERIFICATION.md`, and `REVIEW.md` for RED/GREEN, automated verification, and review evidence.
