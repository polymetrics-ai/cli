---
phase: 601
command: gsd-code-review 601
mode: inline_manual_fallback
depth: deep
status: passed_after_fix
correction_rounds: 4
---

# Phase 601 code review — #3754 shared rate-budget coordinator

The generated `code-review` workflow was completed inline because this
repository lane forbids spawning the upstream reviewer/fixer roles. Review
covered the cross-file flow from runtime backend selection through requester
admission/finish, local coordinator state, and the run-owned UDS protocol.

## Resolved finding

| ID | Severity | Finding | Disposition |
| --- | --- | --- | --- |
| R1 | Warning | The first GREEN version retained legacy `Admission`/`Observer` compatibility hooks beside `LeaseAdmission`. Requester gave the lease path precedence, but resolving the unused hooks also populated a second local registry key containing connector/policy identifiers. That contradicted the coordinator identity boundary and wasted state. | Fixed in this review pass: engine now attaches only `LeaseAdmission`; the existing `RateLimitRegistry`/`RateLimiter` direct API remains unchanged and its full suite passes. Requester again delivers a separately configured observer normally, and it rejects an empty lease before any transport send. |
| R2 | Warning | A stricter new-changes-only lint pass found unchecked close results in the new UDS owner/client and helper process test, plus a staticcheck conversion simplification. | Fixed by explicitly discarding cleanup-only close results and using the direct options conversion. `golangci-lint --new-from-rev=origin/docs/4015-connector-release-certification` passes for coordination, connsdk, and engine. |
| R3 | Error/Warning | A two-TTL record retirement could drop a normal late stricter observation; a UDS exchange could wait for its five-second deadline after the caller canceled. | Fixed under #4035: expired leases became inactive but finishable for late typed observations, and client cancellation now interrupts socket I/O and wins a response race. |
| R4 | Warning | The R3 lost-Finish path retained inactive owner records forever, making future admission scans grow with abandoned leases. | Fixed under #4035: one owner-normalized two-minute completion-observation horizon removes a record before lookup at the exact boundary while preserving charged consumption; injected-clock boundary, cleanup, and bounded-scan coverage passes. |

## Review result

No unresolved Critical, Warning, or Info finding remains. The UDS protocol is
closed to versioned ready/decide/finish frames, frame and policy counts are
bounded, endpoint/epoch data remains private and absent from errors/evidence,
and owner loss normalizes to a fail-closed pre-send refusal under
`require_shared`. Batch state retains only a policy fingerprint, opaque scope,
typed budgets, opaque lease, and typed completion observation.

The review-triggered source corrections are **4 / 5**. #4025 is not included: it
remains the separately owned planning-tool traceability issue and caused no
coordinator source change in this review. Rounds 3 and 4 are the #4035
late-completion correction sequence; the focused coordinator command passes
with the existing UDS cancellation cases and the eight-helper 3-grant/5-refusal
control.

The whole-package stricter lint command also reports pre-existing findings in
unrelated connsdk files. Those files were not changed for #3754; the project's
configured `make lint` and the new-changes-only lint both pass, so no unrelated
cleanup was folded into this child.
