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
- **#4029 / correction loop 3/5:** the metadata-only eligibility projection accidentally reused
  the runtime durable-acknowledgement admission condition, hiding a structurally valid
  `acknowledgement: none` destination. The RED test reproduced the omission; inspection now
  projects every structurally valid destination and retains its declared acknowledgement. The
  existing Registry.Preflight rejection of non-durable acknowledgement remains unchanged.
- **#4046 / correction loop 4/5:** the commit callback retained an acknowledged checkpoint only
  in process memory. It now updates only the active stream state with checkpoint, generation, and
  acknowledgement time before returning; it leaves successful-run identity, metrics, and completion
  to source exhaustion. The focused suite covers all seven modes, source failure, cancellation after
  acknowledgement, definite final-state failure, and indeterminate state persistence.
- **#4029 and #4045 / correction loop 4/5:** eligibility had to distinguish independently valid
  metadata roles, while execution had to reject a connector whose other role was malformed. Role
  projection now validates roles independently; preflight full-validates each selected complete
  descriptor before role extraction, planning, or source read.
- **#4048 / correction loop 4/5:** a typed-nil conformance verifier passed the interface-nil check
  and could panic. Registry construction and preflight use the existing typed-nil guard and return
  the unavailable-verifier error without invocation.
- **#4047 / correction loop 4/5:** nested composite cloning still left byte backing storage shared.
  The clone boundary now copies `[]byte` values and the regression covers scalar, map, and list
  positions without changing other record value semantics.
- **#4029 / correction loop 4/5:** the generated CLI manual had policy wording that the canonical
  `connectorsHelp` source lacked. The source, regenerated manual, and golden transcript now agree.
- **#4046 / correction loop 5/5:** an acknowledgement-stamped checkpoint could bypass the active
  resume identity, and a stale app instance could overwrite a newer target-stream checkpoint.
  The callback now validates the stamped envelope before any state mutation, then performs a
  target-entry compare-and-swap under the JSON-store lock. It retains prior successful-run
  metadata through interim pages, distinguishes an absent entry from a present zero entry, and
  compares every opaque checkpoint field as raw bytes. Deterministic tests cover identity and
  generation rebootstrap, binary multi-page advancement, unrelated updates, stale two-app
  writes, cancellation, state-save outcomes, and all seven modes.

## Review checks

- The registry validates each selected complete descriptor before extracting exact role references,
  then validates integration-family, mode, stream, acknowledgement, strategy, registration, and
  external verification before `ReadTransport`.
- The orchestrator has no API/database-pair branch and keeps warehouse mediation typed; it has no
  generic SQL, HTTP, shell, URL, query, or arbitrary-action input.
- The destination receives only the descriptor-resolved strategy. Its acknowledgement is checked
  through `synccontract.CommitAfterDownstreamAcknowledgement`; the app persists an interim active
  stream checkpoint before that callback returns and before a post-acknowledgement cancellation.
  The interim envelope must match the active source resume expectation, and its state update is a
  target-stream compare-and-swap rather than a whole-project revision guard or last-writer-wins
  assignment.
- Provider record maps and byte values are defensively copied for the warehouse workset. Cancellation
  after stage reaches neither destination apply nor checkpoint commit; cancellation after a durable
  acknowledgement commits the acknowledged page before returning cancellation.
- The public projections remain metadata-only and explicitly avoid a certification claim. The
  production verifier remains unavailable by default, so a descriptor cannot self-admit. A valid
  `acknowledgement: none` is visible in inspection but cannot execute.
- No file under `internal/synccontract`, provider-specific connector implementation, database
  protocol, or live harness changed.
