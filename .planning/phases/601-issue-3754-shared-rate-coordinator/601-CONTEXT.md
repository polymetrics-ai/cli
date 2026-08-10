# Phase 601: issue-3754-shared-rate-coordinator - Context

**Gathered:** 2026-08-11
**Status:** Ready for planning

<domain>
## Phase Boundary

Complete the still-open shared half of issue #3754: an optional, same-host,
run-owned Unix-domain-socket rate-budget coordinator. It must preserve the
ordinary dependency-free process-local registry while letting an internal
caller explicitly require shared coordination and fail closed before a send if
it is unavailable. This is a coordination foundation only; it does not add
provider policies, GraphQL behavior, operator event rendering, CLI flags, or
external services.

</domain>

<decisions>
## Implementation Decisions

### Backend boundary and injection

- **D-01:** Keep the existing in-process `RateLimitRegistry` as the default
  ordinary-CLI backend. Preserve its current public behavior; layer atomic
  batch decisions behind a new injected `BudgetCoordinator` decision/finish
  seam rather than replacing it with a global shared service.
- **D-02:** Use a closed backend mode: `process_local` is the explicit normal
  mode and `require_shared` needs an injected, ready shared client. Missing,
  incompatible, expired, or lost shared ownership is a typed refusal with no
  process-local fallback and no requester transport call.
- **D-03:** Thread the selection through the existing internal runtime/request
  admission seam only. It is not a new end-user command or configuration flag,
  so CLI help/manual/website changes are not applicable to this slice.

### Coordinator protocol and privacy

- **D-04:** Implement a run-owned UDS owner/client under
  `internal/coordination` using only the Go standard library. The owner creates
  a short mode-0700 run directory, creates/chmods a mode-0600 socket, performs
  a versioned readiness handshake, and removes the socket/run directory on
  normal close. Tests must prove zero temporary residue.
- **D-05:** Use a small closed, length-bounded protocol with only readiness,
  decision, and finish messages. It transfers policy fingerprints, opaque
  `RateLimitScopeKey` projections, typed costs/budget declarations, lease IDs,
  and typed completion observations. It never transfers credentials,
  credential revisions or hashes, binding preimages, URLs, request bodies,
  headers, variables, raw account/installation values, socket paths, or run
  epochs in operator-facing evidence.
- **D-06:** Policy identity is fingerprint-first: the owner records a stable
  fingerprint on first use for each opaque key and rejects a mismatched later
  declaration. The policy fingerprint is a deterministic canonical
  representation of the already-declared budget contract, not a new provider
  policy source.

### Atomic admission and completion

- **D-07:** `Decide` evaluates every matching consumptive policy and one
  owner-managed in-flight lease as one transaction. A refusal returns typed
  `not_before`/reason information and consumes neither any policy capacity nor
  a lease. A grant returns an opaque lease ID.
- **D-08:** `Finish` is idempotent by opaque lease ID. It always releases the
  shared concurrency lease and can only tighten the owner state from typed
  response observations. A short TTL reclaims concurrency after a crashed
  client, but an indeterminate consumptive reservation remains charged.
- **D-09:** Owner time is authoritative for budget and lease expiry. Older
  reset-window observations cannot widen a newer window; absent/ambiguous
  observations never widen a declared ceiling.

### Failure, deadline, and restart semantics

- **D-10:** The shared client has bounded dial/read/write operations and
  honors caller context. If a returned `not_before` cannot fit before the
  caller deadline, it returns a typed deadline-too-short refusal before a
  request is sent. No backend silently falls back from `require_shared`.
- **D-11:** Owner loss is terminal for the run: future admissions fail closed;
  granted work stays charged/indeterminate. A fresh owner has a fresh epoch;
  an old client is refused rather than joining it. Seamless same-run restart,
  durable budget recovery, cross-host coordination, SQLite, DuckDB locking,
  Redis/Dragonfly, and generic RPC are excluded.

### TDD and evidence

- **D-12:** The first executable RED tests are exactly
  `TestRateBudgetReserveBatchAllOrNothing`,
  `TestUnixRateBudgetCoordinatorMultiProcessTinyBudget`,
  `TestRequireSharedRefusesWithoutCoordinatorBeforeSend`,
  `TestSharedRateBudgetScopesRemainIndependent`,
  `TestSharedRateBudgetDeadlineTooShortDoesNotSend`, and
  `TestSharedRateBudgetOwnerCrashFailsClosed`.
- **D-13:** The multiprocess acceptance test uses eight real helper processes
  held behind one local barrier. With a shared budget of three it must produce
  exactly three grants and five typed unattempted blocks; the process-local
  control must admit all eight. The test runs locally on macOS and verifies
  socket/run-directory cleanup.
- **D-14:** This worker executes generated GSD prompts inline because the
  canonical delivery contract forbids replacement role spawning in this
  runtime. Required skills loaded: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, and
  `golang-concurrency`.

### the agent's Discretion

- Choose the smallest internal type layout and protocol encoding that preserves
  the closed, bounded, opaque contract above.
- Choose deterministic fake-clock and helper-process test mechanics that make
  the macOS mult-process acceptance test reliable without weakening it.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Delivery contract and issue topology

- `AGENTS.md` — mandatory issue-first lifecycle, stacked PR topology, safety,
  test, review, and cleanup requirements.
- `.agents/agentic-delivery/canonical/delivery-contract.json` — canonical
  single-worker state machine and exact child no-mistakes topology.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — issue-first
  GSD/TDD, commit, verification, and PR evidence contract.
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md` —
  parent/child ownership and no-delegation contract.
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md` —
  non-default-base sub-PR requirements.
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md` and
  `.agents/agentic-delivery/workflows/claude-review-loop.md` — required review
  coverage/disposition route.

### Required technical and GSD guidance

- `.agents/agentic-delivery/references/required-skills-routing.md` — loaded
  Go skills for this concurrency/security/test foundation.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — prompt generation,
  inline fallback, and required GSD command sequence.
- `.planning/phases/issue-3752-rate-limit-admission-r1/CONTEXT.md` — original
  `RateLimitAdmission` seam and #3754 ownership boundary.
- `.planning/phases/issue-3753-rate-limit-enforcement-r1/CONTEXT.md` and
  `.planning/phases/issue-3753-rate-limit-enforcement-r1/TDD-LEDGER.md` —
  completed process-local behavior that must remain compatible.
- `.planning/phases/issue-3863-secret-free-coordination-identity-r1/CONTEXT.md`
  — #3863 opaque identity contract; this phase consumes it unchanged.

### Existing implementation

- `internal/coordination/rate_limits.go` — current process-local clocks,
  models, registry, and monotonic observation behavior.
- `internal/coordination/rate_limits_test.go` — current deterministic local
  rate-budget coverage.
- `internal/connectors/connsdk/rate_limit_requester.go` and
  `internal/connectors/connsdk/http.go` — pre-send admission / response
  observation boundary that must prevent a transport send on refusal.
- `internal/connectors/engine/rate_limit_runtime.go` — selector resolution and
  current sequential policy admission to replace with atomic batches.
- `internal/connectors/coordination_identity.go` — opaque
  `RateLimitScopeKey` derivation; never duplicate or widen it.

### Authoritative external research

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-github-rate-budget-research-r1/report.md`
  — complete Sol research; #3754 shared-half scope, non-goals, exact RED/GREEN
  tests, and UDS rationale.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-connector-release-certification-r1/implementation-brief-template.md`
  — child lifecycle, correction-loop cap, no-mistakes, Shepherd, PR, and
  completion-report contract.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `coordination.RateLimitRegistry` / `RateLimiter`: deterministic clock,
  provider-safe budget algorithms, and strict observation behavior.
- `connectors.CoordinationIdentity.RateScopeKey`: the only valid source of an
  opaque shared rate scope.
- `connsdk.Requester` admission hooks: exact pre-send transport boundary that
  already covers JSON, form, multipart, stream, retries, and redirects.

### Established Patterns

- Context is passed explicitly through requester and rate-limit seams; waits
  must observe cancellation/deadlines.
- Requester observations are typed scalars, not raw headers/bodies.
- Test clocks are injected and current local limiters are process-global only
  through an engine test helper.

### Integration Points

- `engine.rateLimitResolver` is the selector/batch construction boundary.
- `RuntimeConfig` is the internal caller configuration boundary; it must not
  create a package/import cycle or expose a user-facing mode in this slice.
- `connsdk.Requester` owns releasing a decision lease for both response and
  indeterminate send paths.

</code_context>

<specifics>
## Specific Ideas

The shared coordinator is intentionally an ephemeral same-host run resource,
not durable provider truth. The core correctness property is all-or-nothing
batch reservation: a policy block must leave every other policy and the
concurrency lease untouched.

</specifics>

<deferred>
## Deferred Ideas

None — GraphQL costs/policies, provider-specific declarations, operator events,
public CLI/docs surfaces, bounded 665-case execution, cross-host services, and
same-run restart are owned by their existing issues (#3755, #3990, #3993,
#3758) or future work.

</deferred>

---

*Phase: 601-issue-3754-shared-rate-coordinator*
*Context gathered: 2026-08-11*
