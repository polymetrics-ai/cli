# TDD Ledger: live-path defects r1

## Planned test contract

| Issue | Class | Named evidence and observable |
| --- | --- | --- |
| #4119 | Happy | Redirect to a declared `require_shared` destination is admitted/charged by that destination's policy, with exactly the expected provider sends. |
| #4119 | Bad | Canonical destination that cannot be admitted returns `*coordination.SharedRateLimitUnavailableError` naming the destination and sends zero requests. |
| #4119 | Edge | Local-only redirect, base-prefixed route, and unchanged direct route prove canonicalization neither blocks local traffic nor charges the original route. |
| #4125 | Happy | The maximum accepted window produces the exact declared duration/TTL contract. |
| #4125 | Bad | Negative and one-past-maximum windows return the specific typed validation error before cache/coordinator I/O. |
| #4125 | Edge | Zero and a duration-overflow-sized positive window receive named typed outcomes without panic or silent clamp. |
| #4169 | Happy | A provider 401 through the production CLI construction path is a typed credential/authentication outcome whose message omits credential contents. |
| #4169 | Bad | A true internal failure remains the typed `internal_error` outcome and is not collapsed into a credential error. |
| #4169 | Edge | Rejected credential leaves provider writes at zero and checkpoint state unchanged; adjacent statuses retain their defined categories. |

## Actual evidence

### 2026-08-16 — planning checkpoint

- Red: pending. No production paths have been edited.
- Green: pending.
- Manual GSD fallback: prompts resolved and executed inline because no
  compatible isolated GSD worker is available and the task forbids role
  spawning.

### 2026-08-16 — #4119 destination-route regression probe

- Red attempt: the newly added redirect-to-`require_shared` destination probe
  was run before any production edit and passed on `ef3c71caf`.
- Evidence: `Requester.admitRequesterSend` already canonicalizes each physical
  request URL at the send boundary, and its redirect `CheckRedirect` invokes
  the same admission before the destination can be sent. The new tests prove
  the previously unexercised different-budget route: a local `/start` may send
  once, while `/repos/widget` is refused as
  `SharedRateLimitCoordinatorNotConfigured` and receives zero sends.
- Happy class: `TestEndpointLocalRateLimitAdmissionAllowsRedirectDestination`
  asserts the response body and one destination request for local-only traffic.
- Bad class: `TestEndpointSharedRateLimitAdmissionUsesRedirectDestination`
  asserts the destination shared typed refusal and zero destination requests.
- Edge class:
  `TestEndpointSharedRateLimitAdmissionCanonicalizesBasePrefixedRedirectDestination`
  asserts the same typed destination refusal for a base-prefixed redirect.
- Green: the focused engine command passed without a production edit. This is
  an already-satisfied dispatch condition, recorded explicitly rather than
  inventing an unnecessary transport change.

### 2026-08-16 — #4125 shared-window bounds

- Red: `go test -timeout 20m ./internal/coordination -run
  '^TestSharedRateLimitWindowBoundary'` failed before production edits. The
  typed window outcome and bounded TTL contract did not exist.
- Green: `SharedRateLimitWindowError` now distinguishes `non_positive` from
  `too_large`; validation precedes both duration/TTL conversion and the shared
  coordinator availability check. The same defensive ordering applies to
  response observation.
- Happy class: `TestSharedRateLimitWindowBoundaryAcceptsMaximum` asserts exact
  milliseconds and TTL at the largest value that still leaves TTL slack.
- Bad class:
  `TestSharedRateLimitWindowBoundaryRejectsBadInputBeforeCoordinatorIO` asserts
  typed negative/one-past-maximum refusals and zero coordinator connections.
- Edge class:
  `TestSharedRateLimitWindowBoundaryRejectsZeroAndDurationOverflow` names zero
  and an `int` maximum that would overflow `time.Duration`; both are typed
  refusals without a panic or silent clamp.
- Green command: `go test -timeout 20m ./internal/coordination` passed.

### 2026-08-16 — #4169 provider authentication classification

- Red: focused CLI tests failed before production edits: a 401 classified as
  `internal/internal_error` and returned exit 1. The pre-existing fresh binary
  flow harness was also run red-first but currently fails earlier at a valid
  flow control step (exit 3), before it reaches the authentication assertion;
  its assertion was retained rather than weakened.
- Green: a safe `connsdk.CredentialRejectedError` preserves only the provider
  401 identity through response formatting, never its URL or body. CLI maps
  that typed identity (and an unformatted `HTTPError` fallback) to
  `auth/credential_error` with the user-legible `provider rejected the
  credential` message. Generic internal failures remain
  `internal/internal_error`.
- Happy class:
  `TestFreshBinaryProvider401IsCredentialErrorWithoutWritesOrCheckpointAdvance`
  drives a freshly built binary through credential storage and the declared
  GitHub direct-read command; it asserts one provider read, zero writes, the
  typed auth envelope, and unchanged durable checkpoint state.
- Bad class: `TestClassifyErrorInternalFailureRemainsInternal` proves a real
  internal error is not misclassified as a credential problem.
- Edge class: `TestWriteErrorProvider401RedactsCredential` proves JSON and
  stderr retain user guidance but no credential value; the existing fresh-flow
  assertion also retains the full flow's zero provider mutations/checkpoint
  guarantee once its independent setup failure is repaired.
- Green commands: focused engine formatter, focused CLI classification, and
  the new fresh-binary 401 proof passed.
