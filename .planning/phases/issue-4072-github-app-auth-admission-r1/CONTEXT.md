# Issue #4072: GitHub App auth admission - Context

**Gathered:** 2026-08-12
**Status:** Ready for execution
**Lineage:** #3754 child, fresh correction ledger **0/5**

<domain>
## Phase Boundary

Make GitHub App installation-token minting participate in the same declared
shared/process-local rate-admission path as every other GitHub REST request.
The runtime must fail closed before the token POST when `require_shared` has no
ready coordinator or loses one. This phase does not change rate policies,
credential selection, provider behavior, CLI surface, write hooks, GraphQL
exclusion, or any sibling #3754 recovery child.
</domain>

<decisions>
## Implementation Decisions

### Admission boundary

- **D-01:** Construct the rate-limit resolver and a base requester before a
  network-capable custom authenticator is selected.
- **D-02:** Supply only a narrow engine-owned declared-route request capability
  to an auth hook that needs a physical request. Do not expose the raw
  coordinator, a `Runtime`, or a user-facing generic HTTP writer.
- **D-03:** The GitHub App token exchange declares `POST
  /app/installations/{installation_id}/access_tokens` for policy selection and
  sends the separately interpolated escaped installation path.
- **D-04:** One physical token request has exactly one reservation and one
  finish observation. Missing/lost shared coordination refuses before
  transport.

### Privacy and compatibility

- **D-05:** JWTs, private-key material, and minted installation tokens stay out
  of coordination identity, reservation batches, errors, and logs.
- **D-06:** Preserve bearer authentication, ordinary declared requests,
  process-local admission, existing GitHub write-hook admission, the current
  GitHub policy declaration, and its `POST /graphql` exclusion.

### Delivery controls

- **D-07:** This is a manual inline GSD fallback because the project adapter
  only recognizes numeric roadmap phases and the canonical contract forbids
  role spawning in this lane. The lifecycle evidence remains mandatory.
- **D-08:** The canonical private finish-plan snapshot is
  `939f14f61defd993f8ad0335a5eb617d97083c9f73a6a75259d0e312ae8f408`.
  Its live-topology update does not alter this child's requirements or DAG.
</decisions>

<specifics>
## Specific Ideas

- Use only local fake coordinator and recording transport evidence; no GitHub
  credential or provider request is permitted.
- The accepted recovery base is exactly
  `da8a8ff07aaf00e5c7965cd4d1d3c7252017d785` from
  `refs/remotes/no-mistakes/feat/3754-shared-rate-coordinator`.
</specifics>

<canonical_refs>
## Canonical References

### Issue and recovery canon

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-rate-3754-exhaustion-audit-r1/report.md` — finding, required scope, acceptance criteria, and RED/GREEN contract.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-connector-release-certification-r1/FINISH-AND-PARALLELIZATION-PLAN.md` — recovery topology and shared validation gate.
- `.agents/agentic-delivery/canonical/delivery-contract.json` — issue-first delivery contract.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — manual GSD fallback constraints.

### Engine and GitHub implementation

- `internal/connectors/engine/read.go` — current auth/resolver construction order.
- `internal/connectors/engine/auth.go` and `internal/connectors/engine/hooks.go` — custom-auth hook boundary.
- `internal/connectors/engine/rate_limit_runtime.go` — declaration-aware requester resolution.
- `internal/connectors/connsdk/http.go` and `internal/connectors/connsdk/rate_budget_coordinator.go` — admission/finish lifecycle and opaque coordinator seam.
- `internal/connectors/hooks/github/hooks.go` — GitHub App token exchange.
- `internal/connectors/defs/github/rate_limits.json` and `api_surface.json` — existing GitHub REST policy and declared token route.
</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable assets

- `Runtime.RequesterFor(method, path)` resolves policy from declaration-level
  route data and retains no raw runtime subject.
- `connsdk.Requester` owns the single admission/finish lifecycle for each
  physical logical send.
- `RateBudgetRefusalError` exposes the typed
  `shared_coordinator_unavailable` refusal without provider or credential data.

### Required integration point

`newRuntime` currently calls `selectAuth` before creating the resolver, and
the GitHub hook calls `http.DefaultClient` directly. The repair must reverse
that ordering and route the token exchange through the already admission-aware
requester capability.
</code_context>

<deferred>
## Deferred Ideas

- #3990 consumes the shared foundation for GitHub rate-budget certification.
- The second UDS availability finding remains serialized after this child.
- PostgreSQL, warehouse, connector certification, Shepherd, and polling/apply
  work are outside this phase.
</deferred>

---

*Issue phase: issue-4072-github-app-auth-admission-r1*
*Context gathered: 2026-08-12*
