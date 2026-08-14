# Issue #3867: rate-limit parking and automatic resumption - Context

**Gathered:** 2026-08-15  
**Status:** Ready for planning

<domain>
## Issue Boundary

Persist a connector-neutral `parked_rate_limit` outcome with the provider's
authoritative reset evidence, block only the matching opaque rate scope, and
resume the parked work at or after reset from its already committed checkpoint.
The outcome must survive coordinator/scheduler reconstruction and emit
truthful, secret-free park/resume events.

</domain>

<decisions>
## Implementation Decisions

### Durable parking contract

- **D-01:** Parking is keyed by the existing opaque `connectors.RateLimitScopeKey`
  and a caller-supplied run identity; credentials, bindings, provider URLs,
  response bodies, and raw headers are neither persisted nor rendered.
- **D-02:** A run is parked only from a typed `*connsdk.RateLimitError` that
  contains a non-zero parsed reset instant. The persisted value is the exact
  UTC reset instant plus the typed reason/source; missing reset evidence stays
  an ordinary error and must not invent a resume time.
- **D-03:** The parking record owns a defensive clone of the already committed
  `synccontract.CheckpointEnvelope`. A resume callback receives that committed
  checkpoint and cannot receive an uncommitted candidate or re-run a prior
  acknowledged destination apply.

### Admission and scheduling

- **D-04:** The parking coordinator is an injected, race-safe opaque-state
  store plus deterministic clock/scheduler seam. It denies same-scope
  admission while a record is parked; unrelated scopes remain independently
  admissible.
- **D-05:** Restoring a coordinator from the same store re-arms each pending
  parking record. Resumption is attempted once at or after its stored reset,
  never before; duplicate observations for one run remain one scheduled
  resumption.
- **D-06:** Cancellation removes a parked run before dispatch. A resumed run is
  removed only after its callback returns success, preserving the record for a
  later retry/restart on callback failure.

### Operator events

- **D-07:** Park and resume events are a closed, secret-free vocabulary. A
  parking event reports the actual typed reason/source and exact persisted
  reset time, never a generic failure message. A resume event identifies the
  corresponding transition without exposing a scope key or request data.

### the agent's Discretion

- Keep the coordinator/application boundary connector-neutral and reuse the
  existing rate-limit requester/event seams where that preserves the decisions
  above. Do not add a CLI surface or a dependency.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Issue and delivery contract

- `.agents/agentic-delivery/references/required-skills-routing.md` — required
  Go skills and GSD lifecycle.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — generated-prompt
  and inline fallback rules.
- `AGENTS.md` — mandatory issue-first/TDD delivery, scope exclusions, and
  local verification commands.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-mvp-verify-coordination-r1/report.md`
  §#3867 — verified missing work and landed prerequisite boundaries.

### Existing contracts

- `internal/coordination/shared_rate_limits.go` — existing shared opaque
  rate-scope admission and observation behavior; do not alter #4125's separate
  `window_seconds` duration-overflow defect.
- `internal/coordination/auth_cohort.go` — opaque-key persistence, restart,
  cancellation, epoch, and race-safe coordinator precedent from #3865.
- `internal/connectors/connsdk/rate_limit_requester.go` — typed
  `RateLimitError`, typed reset facts, and safe request event vocabulary.
- `internal/connectors/connsdk/http.go` — event emission and the no-send
  admission boundary.
- `internal/connectors/engine/rate_limit_runtime.go` — project-scoped engine
  rate limiter/event-sink assembly.
- `internal/synccontract/state.go` — clone/validation rules for durable
  checkpoints.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `coordination.AuthCohortHealthStore` and its memory implementation establish
  the opaque-store/restart seam for a connector-neutral coordinator.
- `coordination.RateLimitKey` and `SharedRateLimitRegistry` already enforce the
  required rate-scope isolation and contain no credentials.
- `connsdk.RateLimitError` retains only parsed typed reset facts, and
  `RateLimitEventSink` is synchronous, bounded reporting.
- `synccontract.CheckpointEnvelope.Clone` keeps an admitted resume value from
  aliasing callers or an uncommitted candidate.

### Established Patterns

- Coordinator state transitions occur under a mutex; persistence commits before
  cancellation/callback work is released.
- Existing tests use deterministic injected clocks/fakes and assert no-send
  counters rather than merely checking a returned error.

### Integration Points

- Engine runtime resolves declared rate policies and attaches their typed event
  sink to requesters.
- The transport/app layer already resumes sources from a committed checkpoint;
  this issue must hand that checkpoint back unchanged rather than replaying an
  acknowledged destination apply.

</code_context>

<specifics>
## Specific Ideas

The requested proof is state-observable: restart reads the stored parked run,
pre-reset scheduling yields exactly zero sends, and the emitted parking event
contains the real typed reason and reset time.

</specifics>

<deferred>
## Deferred Ideas

None. Explicitly excluded: #4125 duration overflow, #4136 certification
validation, and #4090.

</deferred>

---

*Issue: #3867 rate-limit parking and automatic resumption*  
*Context gathered: 2026-08-15*
