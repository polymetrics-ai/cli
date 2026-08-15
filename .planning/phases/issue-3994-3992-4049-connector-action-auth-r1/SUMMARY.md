---
phase: issue-3994-3992-4049-connector-action-auth-r1
status: complete
coverage:
  - id: D1
    description: Prepared connector actions carry payload-bound execution identity through the fresh binary path.
    verification:
      - kind: e2e
        ref: internal/cli/connector_action_authorization_binary_test.go TestPMBinaryExecutesInstalledApprovedJobFlow
        status: pass
      - kind: integration
        ref: internal/app/flow_action_test.go TestAuthorizedFlowActionPreparedIdentityBindsPayloadAndReachesReceipt
        status: pass
    human_judgment: false
  - id: D2
    description: Flows and schedules inherit standing job approval without a crontab token and revalidate every firing.
    verification:
      - kind: integration
        ref: internal/cli/schedule_fire_test.go TestInstalledScheduleFireRunsAuthorizedRoundTripAndRestoresBackend
        status: pass
      - kind: unit
        ref: internal/cli/flow_cli_test.go flow job-reference refusal tests
        status: pass
      - kind: unit
        ref: internal/cli/schedule_test.go schedule flow-reference refusal tests
        status: pass
      - kind: e2e
        ref: internal/cli/certify_cli_test.go TestCertifyCLISingleConnectorPassExitsZero
        status: pass
    human_judgment: false
  - id: D3
    description: Required shared rate coordination refuses with the stable SDK type and code before transport.
    verification:
      - kind: e2e
        ref: internal/cli/connector_action_authorization_binary_test.go TestPMBinaryRefusesRequiredSharedRateBudgetBeforeSend
        status: pass
      - kind: integration
        ref: internal/connectors/engine/rate_limit_coordination_test.go require-shared refusal tests
        status: pass
    human_judgment: false
---

# Connector action authorization residuals summary

Flows now positively resolve stored ETL connections and already-approved reverse plans before an
atomic manifest write. Schedules positively resolve a stored, still-valid flow before schedule or
backend writes. The installed command remains exactly `pm --root <root> flow run <name> --json`;
each firing reloads job authorization and parks on drift, revocation, expiry, cancellation,
ambiguous delivery, or cleanup failure.

Prepared connector writes carry a deterministic `pex_` identity bound to the resolved manifest,
standing authorization scope, firing, mapped payload, preview, destination, and credential
revision. A durable exclusive marker refuses concurrent/completed/ambiguous replay. Required shared
rate coordination now exposes `*connsdk.RateBudgetRefusalError` with code
`shared_coordinator_unavailable` before any provider request.

The repo's single-worker contract prohibited role spawning, so discuss, TDD planning, execution,
verification, and review used the documented inline/manual GSD fallback.

After PR #4168 initially failed `certify-timing`, the branch was rebased onto moved base
`6b9cfe492`. The branch-owned certification harness was corrected to remove its obsolete schedule
authorization reference and to assert the no-carrier contract directly. The exact CI test and the
timing gate now pass locally.

The next hosted full-suite pass confirmed that timing fix, then exposed an in-memory compatibility
regression in branch-owned live reverse-plan reload. Store-backed production Apps still reload
durable state; deliberately store-less policy fixtures now retain the prior lookup semantics and
their typed replay refusal. Final focused lint also closes the prepared-execution directory handle
explicitly.
