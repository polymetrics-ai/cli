# Review — issue-3753-rate-limit-enforcement-r1

Manual code review under the GSD fallback found no blocking issue.

- Requester wiring was searched for direct `Requester.Do` sends. Engine paths now resolve through `Runtime.RequesterFor`; the only direct runtime requester test exercises whole-connector hook policy attachment.
- The registry accepts only `connectors.RateLimitScopeKey`, never a binding, secret, subject, or credential revision. The resolver passes the subject directly to `CoordinationIdentity` and retains only its projection in registry state.
- Admission waits use the injected clock and its wall-clock implementation selects on `ctx.Done()`.
- `rate_limits.json` remains test-only; `internal/connectors/defs/defs.go` was not touched.
- Strict direct writes still set `DisableRetries` after policy attachment, preserving the no-replay behavior.

Automated PR review is intentionally deferred to the captain-gated firstmate flow; no PR has been opened by this worker.
