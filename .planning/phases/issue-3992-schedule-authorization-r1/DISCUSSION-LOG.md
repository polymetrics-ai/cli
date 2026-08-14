# Discussion log — Issue #3992

Inline `discuss-phase --auto` fallback: the adapter cannot resolve issue #3992
as a roadmap-numbered phase, and the task brief locks the relevant decisions.
The only durable authority is `internal/app/authorization.go`'s content-free,
revocable scope record. A schedule persists and renders only its opaque
authorization reference; it must never contain an approval token, credential,
secret, payload, or secret-derived preimage. Each firing follows the completed
connector-backed flow-action route, so it re-derives and validates scope before
provider dispatch. Scope drift, expiry, and revocation are typed pre-dispatch
refusals. An overlap, crashed/incomplete firing, rate limit, or ambiguous
post-write result halts/parks rather than replaying a mutation.

The certification proof is hermetic: an isolated connector fixture observes a
real installed crontab payload firing through `pm schedule fire`, the flow
engine, typed connector write, read-back, and app receipt persistence. No live
credential is available in this worktree and none will be requested or stored.
