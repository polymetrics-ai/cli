# REVIEW — issue #3614 webhook receiver exposure modes

Verdict: pass, with the inherited #3810 ETL-suite failure tracked in
`VERIFICATION.md` as outside this diff.

## Scope reviewed

- Closed exposure modes and their listener/recovery differences.
- Safe state and CLI serialization paths for callback URLs, credentials,
  signing material, event IDs, and raw bodies.
- Raw-body verifier ordering, durable-receipt acknowledgement ordering,
  duplicate/out-of-order handling, and bounded ingress behavior.
- Tailscale integration boundary: no executable invocation, no dependency,
  no provider registration, no polling executor, and no generic escape hatch.
- Generated CLI/web documentation and website data parity.

## Disposition

The review found and fixed an empty-receipt-map persistence edge case and a
receipt-capacity ordering issue that could have written an encrypted payload
before rejecting an over-capacity receipt. The durable queue now checks
duplicate/capacity admission before encrypting a body, and tests assert a
rejected receipt creates no vault payload.

No remaining finding permits callback URL or signing-material output, permits a
non-loopback tunnel listener, acknowledges before durable receipt persistence,
assumes ordering, starts a tunnel/provider process, or adds a module
dependency.

Provider-specific signature algorithms, subscription registration, polling,
and public Funnel setup remain intentionally owned by their designated lanes or
the operator.
