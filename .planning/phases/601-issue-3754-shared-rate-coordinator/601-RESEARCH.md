# Phase 601 research — optional shared rate-budget coordinator

**Issue:** #3754
**Research mode:** Inline/manual fallback. The canonical single-worker contract
prohibits the GSD researcher and pattern-mapper role spawns in this runtime.
**Sources read in full:** Sol report at
`/Users/karthiksivadas/karthik-agent-workspace/data/cli-github-rate-budget-research-r1/report.md`,
issues #3754/#3750/#3755/#3863/#3995/#4015, merged PRs #3875 and #3877,
the current parent branch, and the current local implementation.

## Reconciliation

- PR #3877 delivered the declared-policy and **process-local** registry path.
  It deliberately has no shared owner, run epoch, shared lease, or
  `require_shared` backend; the shared half of #3754 remains open.
- PR #3875 delivered #3863's opaque `CoordinationIdentity.RateScopeKey`.
  This phase consumes that projection unchanged and must not add an identity,
  token hash, credential revision, or binding preimage.
- #3755 owns operator events, public wait/refusal rendering, help, and docs.
  This phase exposes only internal typed decisions/errors; it must not create
  an operator event schema or a public command flag.
- #3995 is an open issue rather than an available automated Shepherd gate. The
  closeout must record bounded equivalent supervisor evidence instead of
  claiming Shepherd automation ran.

## Implementation findings

1. `coordination.RateLimitRegistry` already has deterministic clocked budget
   models and monotonic response tightening, but its `RateLimiter.Admit`
   reserves one policy at a time. The engine's
   `resolvedRateLimitAdmission.Admit` sequentially invokes those limiters, so
   a later policy can block after an earlier policy has consumed capacity.
2. `connsdk.Requester` is the correct send boundary: it calls admission before
   every logical send, retry, and permitted redirect. A lease-aware internal
   extension can preserve the existing `RateLimitAdmission` interface while
   giving a `BudgetCoordinator` a completion callback for transport results.
3. `RuntimeConfig` is the narrow injection point. A closed internal mode can
   choose the existing process-local registry or an injected shared client
   without adding a CLI flag, provider declaration, GraphQL path, or external
   dependency.
4. A UDS owner is the smallest same-host implementation. It needs a short
   0700 temporary directory, a 0600 socket, a versioned ready/decide/finish
   protocol with a maximum frame size, owner-side time, a random epoch, and
   cleanup on normal close. Network and filesystem errors must be converted to
   bounded typed errors so the endpoint is never rendered.

## Chosen internal shape

- Put the injected `BudgetCoordinator`, reservation batch, decision, opaque
  lease, completion observation, and closed backend-mode types in `connsdk`.
  That keeps `RuntimeConfig` dependency-safe and keeps the socket package from
  depending on the engine.
- Extend the existing registry with an atomic `Decide`/`Finish` implementation
  while preserving the existing `Limiter` API and its ordinary process-local
  behavior. A batch locks/evaluates all matching sets before it consumes any;
  an in-flight lease is recorded only when every policy can grant.
- Resolve a deterministic fingerprint from only connector/policy identifiers
  and typed budget declarations. The owner receives the fingerprint and the
  #3863 opaque scope projection, never a source URL, selector URL, raw subject,
  credential material, or request content.
- Use an optional lease-aware requester admission path. It finishes an opaque
  lease on a response or transport outcome and supplies only the existing
  typed observation. Existing non-leased admissions retain their behavior.
- In `require_shared`, a missing or dead injected coordinator returns a typed
  unattempted error at the pre-send boundary. The process-local registry is
  never selected as a fallback.

## Validation strategy

The initial RED suite contains all six user-mandated test names. It covers
atomic rollback, eight real barrier-released helper processes (three shared
grants/five typed blocks versus eight process-local grants), scope isolation,
no-send failure/deadline paths, owner loss, epoch refusal, and temporary
directory cleanup. No provider request or credential is required.

## Non-goals

No GitHub policy, GraphQL/provider code, operator event surface, CLI/help/docs
surface, Redis/Dragonfly, SQLite, DuckDB lock manager, durable truth,
cross-host protocol, or same-run owner restart belongs to this phase.
