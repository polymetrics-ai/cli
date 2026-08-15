# Issues #3990 and #4091 — live GitHub proof context

**Gathered:** 2026-08-15
**Status:** Ready for TDD planning
**Mode:** Inline/manual GSD fallback

<domain>
## Phase boundary

Run the missing credentialed proofs for the already-landed GitHub certification budget admission
(#3990) and durable per-connection issue-label authorization (#4091) through the production `pm`
binary. Commit only sanitized, reproducible evidence. Fix only defects the live run exposes in files
this branch must touch; route unrelated findings as `needs-decision:`.

</domain>

<decisions>
## Locked decisions

- D-01: Real GitHub provider calls are mandatory; an in-process fake cannot close either live gap.
- D-02: Every exercised path begins at the built `pm` entry point and follows production
  composition. The PR names entry point → registration → dispatch → component explicitly.
- D-03: Credential values, approval tokens, raw provider bodies, and rendered rate scopes never
  enter terminal evidence, files, commits, issue comments, or the PR body. Credentials are admitted
  only from environment variables or stdin into a disposable `pm` project.
- D-04: Reverse ETL follows plan → preview → approval-token-on-stdin → execute. A token is never
  placed in argv or durable evidence.
- D-05: Observable assertions are mandatory. Success asserts exact returned/written/read-back state;
  refusal asserts the typed error plus unchanged provider state, zero sends where observable, and no
  checkpoint advance.
- D-06: The captain-required boundary matrix covers cancellation, process death, empty/single/large
  inputs, duplicate/out-of-order delivery, schema drift, permission/authentication refusal,
  concurrent same-target races, resume, and acknowledged-item replay. A genuinely impossible live
  observation is named individually in the PR instead of being omitted.
- D-07: `scripts/github-live-*.mjs` and the immutable
  `GITHUB-LIVE-LAB-BOUNDARY.json` are reused before adding any new harness. Historical reports are
  context only and are never promoted as current evidence.
- D-08: #4125 and #4158 remain untouched. A finding outside this proof lane is reported to
  firstmate rather than fixed here.
- D-09: Derived catalog, website, transcript, skill, and manual artifacts are regenerated once after
  the final source state, followed by the same local drift gates CI uses.

</decisions>

<canonical_refs>
## Canonical references

- `AGENTS.md` — repository delivery, credential safety, GSD/TDD, CLI parity, and verification rules.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — issue-to-PR lifecycle and PR body.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — required lifecycle command path.
- `.agents/agentic-delivery/references/required-skills-routing.md` — connector/CLI/security skills.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` — parity checklist if behavior changes.
- `.planning/phases/issue-3990-github-budget-admission-r1/VERIFICATION.md` — landed local admission evidence.
- `.planning/phases/issue-4091-github-destination-modes-r1/VERIFICATION.md` — explicit missing live proof.
- `.planning/phases/issue-3993-github-live-certification/VERIFICATION.md` — historical harness limits and current-SHA rules.
- `scripts/github-live-lab.mjs` — boundary, PM-only execution, planned-write, read-back, and cleanup helpers.
- `scripts/github-live-proof-sweep.mjs` — current whole-surface artifact and execution-model guards.
- `.planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json` — immutable allowed target.
- `internal/cli/github_transport_binary_test.go` — production-binary issue-label lifecycle reference.
- `internal/cli/cli.go` — shared coordinator construction and CLI dispatch.
- `cmd/pm/main.go` — production binary entry point and certify CLI registration.

</canonical_refs>

<code_context>
## Existing production paths

- #3990: `cmd/pm/main.go` → `cli.Run` → `coordination.OpenSharedRateLimitRegistry` /
  `engine.ConfigureSharedRateLimitRegistry` → connector/certification CLI dispatch → definition-owned
  GitHub rate policies → admission/observation hooks → shared coordinator.
- #4091: `cmd/pm/main.go` → `cli.Run` → `app.Open` → bundle/transport composition → `etl transport
  github-issue-label` plan/preview or `etl run` dispatch → issue-label transport approval gate →
  declarative GitHub writer → independent GitHub read-back/checkpoint.
- The committed Node proof sweep intentionally refuses credentialed-live claims from its
  external-per-operation execution model. It can supply boundary/artifact guards, but the current
  #3990 proof must use the production in-process certification route unless a new production-safe
  runner is justified by a failing test.

</code_context>

<deferred>
## Explicit exclusions

- `internal/coordination/shared_rate_limits.go` window overflow (#4125).
- `TestPostgresManagedTargetDriverLiveControlAssertions` (#4158).
- New connector capabilities, generic HTTP writes, and any provider target outside the run-owned
  boundary.

</deferred>

---

*Manual discuss fallback: `scripts/gsd prompt discuss-phase issue-3990-4091-github-live-proof-club-r1 --auto` was generated, but `gsd-sdk query init.phase-op` returned `phase_found: false` because this captain-authorized combined issue task is not a ROADMAP phase. The explicit launch brief supplied all decisions, so no human question was required.*
