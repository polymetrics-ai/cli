---
coverage:
  - id: D1
    description: Redirect admission selects the destination route's policy.
    verification:
      - kind: unit
        ref: internal/connectors/engine/rate_limit_coordination_test.go:TestEndpointSharedRateLimitAdmissionUsesRedirectDestination
        status: pass
    human_judgment: false
  - id: D2
    description: Shared windows cannot overflow duration or coordinator TTL.
    verification:
      - kind: unit
        ref: internal/coordination/shared_rate_limits_test.go:TestSharedRateLimitWindowBoundary
        status: pass
    human_judgment: false
  - id: D3
    description: A provider 401 is a safe typed credential failure through a fresh binary.
    verification:
      - kind: e2e
        ref: internal/cli/errors_test.go:TestFreshBinaryProvider401IsCredentialErrorWithoutWritesOrCheckpointAdvance
        status: pass
    human_judgment: false
---

# Summary: live-path defects r1

The phase delivers three isolated commits on the required integration-based
branch. #4119's regression probes established that the dispatch base already
admitted each redirect destination at the requester send boundary; the commit
closes the proof gap without inventing a duplicate transport change. #4125 now
rejects invalid windows before coordinator I/O. #4169 preserves a safe provider
401 identity through response formatting and exposes it as a credential error.

The corrected legacy fresh-flow assertion remains in place and passes after
the branch absorbed merged PR #4174 from the integration base.

During CI triage, the existing durable-parking process test also showed a
load-sensitive base flake (six observed passes and one failure at
`ef3c71caf`). Its assertion is retained; this PR adds only child-process
failure diagnostics, with the underlying race explicitly left to a follow-up.
