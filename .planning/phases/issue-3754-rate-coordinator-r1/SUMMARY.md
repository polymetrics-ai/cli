---
coverage:
  - id: D1
    description: Process-local rate-limit protection is real and visibly bounded to one pm process.
    requirement: R1
    verification:
      - kind: unit
        ref: internal/coordination/shared_rate_limits_test.go:TestRateLimitRegistryStatusIsExplicitlyProcessLocal
        status: pass
      - kind: automated_ui
        ref: internal/cli/rate_limit_coordination_cli_test.go:TestConnectorsInspectLabelsProcessLocalRateLimitProtection
        status: pass
    human_judgment: false
  - id: D2
    description: Explicit require_shared policies fail closed when the coordinator is missing and use it when available.
    requirement: R2,R4
    verification:
      - kind: unit
        ref: internal/connectors/engine/rate_limit_coordination_test.go:TestRequireSharedRateLimitPolicyRefusesWithoutCoordinator
        status: pass
      - kind: integration
        ref: internal/connectors/engine/rate_limit_coordination_integration_test.go:TestRequireSharedRateLimitPolicyUsesAvailableCoordinator
        status: pass
    human_judgment: false
  - id: D3
    description: Separate processes share one opaque, server-time-coordinated budget.
    requirement: R3,R5,R6
    verification:
      - kind: integration
        ref: internal/coordination/shared_rate_limits_integration_test.go:TestSharedRateLimitCoordinatesSeparateProcesses
        status: pass
      - kind: integration
        ref: internal/coordination/shared_rate_limits_integration_test.go:TestSharedRateLimitHonorsEveryDeclaredBudgetModel
        status: pass
    human_judgment: false
---

# Summary — issue #3754 rate-limit coordinator

Implemented process-local rate-limit scope protection and the per-policy optional shared path.
`coordination: require_shared` is a closed declaration opt-in; it fails closed with a typed,
non-sensitive reason before any requester send when the coordinator is unavailable. Unspecified
policies stay process-local even if a runtime coordinator address exists.

The shared registry uses one opaque `CoordinationIdentity.RateScopeKey`-derived Redis-compatible
key per connector/policy/scope. Its admission script obtains server time, applies an expiry, and
atomically reserves fixed-window, sliding-window, token-bucket, and leaky-bucket capacity. It
contains no durable truth, credential, raw binding, or raw subject.

The CLI now exposes only safe provenance in `pm connectors inspect`: process-local protection is
explicitly labelled as limited to the current process; `require_shared` explains its fail-closed
contract. Help, generated CLI docs, conventions, and the website reference are in parity.

Live evidence was posted to #3754. It uses the real local Dragonfly service, two child processes,
and an engine requester test. A credentialed provider call is intentionally deferred to #3990,
which owns provider-specific rate-limit declarations.
