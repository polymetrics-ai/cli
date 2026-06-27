# ADR — Action Step (Phase 1)

## ADR-001: Inline backoff vs. reusing connsdk.Requester

**Status**: Accepted

**Context**: The action step needs 429-aware exponential backoff + jitter.
`connsdk.Requester` already implements this for HTTP requests. However, Requester
is tied to the HTTP transport layer; the action step dispatches through the connector
`Write` interface, not through Requester directly.

**Decision**: Implement an inline `retryWithBackoff` function in `internal/flow/action.go`
that mirrors the connsdk backoff parameters (base=500ms, max=30s, jitter ±25%). The
injectable `Sleep` field (same pattern as connsdk) enables test-speed execution.
Do NOT import connsdk in the flow package to avoid coupling.

**Consequences**: Minor duplication of backoff logic (~30 LOC). Acceptable because:
- Zero new dependencies.
- Test isolation is cleaner (no HTTP transport needed for action unit tests).
- connsdk backoff logic is stable and the pattern is well-understood.

---

## ADR-002: File-backed identity map vs. in-memory

**Status**: Accepted

**Context**: Identity mapping (pm_id → external_id) needs to survive process restarts
(idempotency across runs, not just within a run).

**Decision**: JSON file at `.polymetrics/state/identity_map.json`, loaded at ActionRunner
init and saved after each successful write. Reuses the `JSONStore` pattern from
`internal/state/store.go`.

**Consequences**: File I/O per successful record write. Acceptable for expected record
counts (<100k per run). Runtime-backed alternative (DragonflyDB) deferred to Phase 3.

---

## ADR-003: Flow-level token vs. per-step token

**Status**: Accepted

**Context**: The existing `PlanReverseETL` / `RunReverseETL` issues one token per plan.
For a multi-step flow with multiple action steps, the user could get one token per action
step, or one token for the entire flow.

**Decision**: Default to one flow-level token covering all action steps. Opt-in `--per-action`
flag issues a token per action step. This reduces friction for common single-action flows while
allowing fine-grained control.

**Consequences**: A single token compromise authorizes all action steps in the flow. Mitigated
by 24h expiry and single-use semantics.

---

## ADR-004: DLQ format (NDJSON per run)

**Status**: Accepted

**Context**: Failed records need to be quarantined and queryable.

**Decision**: One NDJSON file per run at `.polymetrics/dlq/<flow>/<step>/<run-id>.ndjson`.
Values redacted. No database.

**Consequences**: Easy to inspect with `cat | jq`. No DLQ replay command in this phase
(deferred). Operator must manually re-run the flow (idempotency ensures safety).
