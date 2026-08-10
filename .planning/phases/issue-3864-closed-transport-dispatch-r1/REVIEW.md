# Code review — #3864 closed transport dispatch

## Method

Manual inline review is the documented fallback because the canonical contract forbids spawning a
reviewer role for this non-numbered issue phase. Reviewed the descriptor/projection code, the
transport registry/orchestrator, bounded app integration, CLI/help/docs changes, and all focused
tests against the #3864 boundaries.

## Findings and disposition

No unresolved Critical, Warning, or Info findings remain.

- **#4021 / correction loop 1/5:** an empty authored descriptor could appear absent because app
  dispatch inspected individual roles. A dedicated RED test reproduced the legacy fallback; app
  dispatch now detects any authored descriptor and passes it to closed preflight. The correction
  stays limited to `internal/app` routing.
- **#4023 / correction loop 2/5:** `generic-http` did not receive the generic-reference rejection
  applied to `generic_http`. A dedicated RED test reproduced it; identifier validation now
  normalizes hyphens before the closed generic check. The correction stays limited to
  `internal/connectors` validation.

## Review checks

- The registry resolves exact descriptor references and validates integration-family, mode, stream,
  acknowledgement, strategy, registration, and external verification before `ReadTransport`.
- The orchestrator has no API/database-pair branch and keeps warehouse mediation typed; it has no
  generic SQL, HTTP, shell, URL, query, or arbitrary-action input.
- The destination receives only the descriptor-resolved strategy. Its acknowledgement is checked
  through `synccontract.CommitAfterDownstreamAcknowledgement` before the app accepts a pending
  checkpoint state.
- Provider record maps are defensively copied for the warehouse workset; cancellation after stage
  reaches neither destination apply nor checkpoint commit.
- The public projections remain metadata-only and explicitly avoid a certification claim. The
  production verifier remains unavailable by default, so a descriptor cannot self-admit.
- No file under `internal/synccontract`, provider-specific connector implementation, database
  protocol, or live harness changed.
