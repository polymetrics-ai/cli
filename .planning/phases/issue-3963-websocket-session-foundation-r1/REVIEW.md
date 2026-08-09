# Code Review — closed WebSocket session operation foundation, R1

## Delivery method

`scripts/gsd sources code-review` resolved the project-local adapter sources on 2026-08-10. The
official phase runtime does not register this issue-specific phase, and the canonical parent
delivery contract prohibits reviewer-role spawning. This is therefore the required inline/manual
`code-review` fallback, performed after the final `verify-work` re-gate. The reviewer also read the
installed GSD code-review workflow before this pass.

## Scope reviewed

The review covered the foundation delta from rebased parent `3212be755` through the current
foundation head, including:

- the closed `websocket_session` schema, declaration validation, authenticated HTTP 101 upgrade,
  bounded frame codec, cancellation ownership, and redacted event/error paths;
- commandrunner preflight/input confinement and the CLI's preserved reserved-control selection;
- `surface-sync`, surface reconciliation, static conformance, and generated-command validation;
- loopback-only tests, source-sensitive error paths, and the post-`f96a47e80` generated-artifact
  re-gate.

The review specifically checked for caller-controlled origins/protocols/headers/frame controls,
credential or token disclosure, redirect/auth leakage, file traversal/symlink escape, uncapped
memory or frame handling, context/connection leaks, mismatch between generated metadata and runtime
preflight, and regressions to plural-write/direct-write surface coverage.

## Findings and disposition

| Severity | Finding | Disposition |
| --- | --- | --- |
| P1 | Byte bounds did not bound wall-clock execution. `cli.Run` uses a background context, so a server that never sent its terminal close could retain a completed session indefinitely. | **Fixed.** Red commit `1324c271c` captured the closed-schema rejection of the new `max_session_seconds` declaration. The follow-up implementation makes that field required, positive, and capped at one hour; the executor derives a child deadline before upgrade and closes its connection on expiry. A loopback regression deliberately withholds close and proves return at the declared deadline. |
| P2 | None remaining. | Reviewed and closed. Fixed connector-relative path/subprotocol/headers, request authentication, no-redirect policy, frame/output caps, redaction, root-confined PCM16 input, and command metadata all remain declaration-owned. |

## Result

No unresolved correctness, security, or maintainability finding remains in the reviewed foundation
scope. The final re-gate in `TDD-LEDGER.md` and `VERIFICATION.md` is green: scoped tests, full
`internal/cli`, vet, build, bundle validation, surface-sync, and clean docs/site regeneration.

The remaining work is consumer-owned: #3935 must declare Zoom Live Scribe, regenerate artifacts,
and prove the resulting command through the rebuilt binary before the parent can claim Zoom parity.
